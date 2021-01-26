// This source file is part of the EdgeDB open source project.
//
// Copyright 2020-present EdgeDB Inc. and the EdgeDB authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package edgedb

import (
	"fmt"
	"reflect"

	"github.com/edgedb/edgedb-go/internal/aspect"
	"github.com/edgedb/edgedb-go/internal/buff"
	"github.com/edgedb/edgedb-go/internal/cardinality"
	"github.com/edgedb/edgedb-go/internal/codecs"
	"github.com/edgedb/edgedb-go/internal/format"
	"github.com/edgedb/edgedb-go/internal/message"
)

func (c *baseConn) granularFlow(
	r *buff.Reader,
	out reflect.Value,
	q query,
) (err error) {
	typ := out.Type()
	if !q.flat() {
		typ = typ.Elem()
	}

	ids, ok := c.getTypeIDs(q)
	if !ok {
		return c.pesimistic(r, out, q, typ)
	}

	in, ok := c.inCodecCache.Get(ids.in)
	if !ok {
		if desc, OK := descCache.Get(ids.in); OK {
			in, err = codecs.BuildCodec(buff.SimpleReader(desc.([]byte)))
			if err != nil {
				return &unsupportedFeatureError{msg: err.Error()}
			}
		} else {
			return c.pesimistic(r, out, q, typ)
		}
	}

	cOut, ok := c.outCodecCache.Get(codecKey{ID: ids.out, Type: typ})
	if !ok {
		if desc, ok := descCache.Get(ids.out); ok {
			d := buff.SimpleReader(desc.([]byte))
			cOut, err = codecs.BuildTypedCodec(d, typ)
			if err != nil {
				return &unsupportedFeatureError{msg: err.Error()}
			}
		} else {
			return c.pesimistic(r, out, q, typ)
		}
	}

	cdsc := codecPair{in: in.(codecs.Codec), out: cOut.(codecs.Codec)}
	return c.optimistic(r, out, q, typ, cdsc)
}

func (c *baseConn) pesimistic(
	r *buff.Reader,
	out reflect.Value,
	q query,
	typ reflect.Type,
) error {
	ids, err := c.prepare(r, q)
	if err != nil {
		return err
	}
	c.putTypeIDs(q, ids)

	descs, err := c.describe(r, q)
	if err != nil {
		return err
	}
	descCache.Put(ids.in, descs.in)
	descCache.Put(ids.out, descs.out)

	var cdcs codecPair
	cdcs.in, err = codecs.BuildCodec(buff.SimpleReader(descs.in))
	if err != nil {
		return &unsupportedFeatureError{msg: err.Error()}
	}

	if q.fmt == format.JSON {
		cdcs.out = codecs.JSONBytes
	} else {
		d := buff.SimpleReader(descs.out)
		cdcs.out, err = codecs.BuildTypedCodec(d, typ)
		if err != nil {
			return &unsupportedFeatureError{msg: err.Error()}
		}
	}

	c.inCodecCache.Put(ids.in, cdcs.in)
	c.outCodecCache.Put(codecKey{ID: ids.out, Type: typ}, cdcs.out)
	return c.execute(r, out, q, typ, cdcs)
}

func (c *baseConn) prepare(r *buff.Reader, q query) (idPair, error) {
	w := buff.NewWriter(c.writeMemory[:0])
	w.BeginMessage(message.Prepare)
	w.PushUint16(0) // no headers
	w.PushUint8(q.fmt)
	w.PushUint8(q.expCard)
	w.PushBytes([]byte{}) // no statement name
	w.PushString(q.cmd)
	w.EndMessage()

	w.BeginMessage(message.Sync)
	w.EndMessage()

	if e := w.Send(c.conn); e != nil {
		return idPair{}, &clientConnectionError{err: e}
	}

	var (
		err error
		ids idPair
	)

	done := buff.NewSignal()

	for r.Next(done.Chan) {
		switch r.MsgType {
		case message.PrepareComplete:
			ignoreHeaders(r)
			r.Discard(1) // cardianlity
			ids = idPair{in: [16]byte(r.PopUUID()), out: [16]byte(r.PopUUID())}
		case message.ReadyForCommand:
			ignoreHeaders(r)
			r.Discard(1) // transaction state
			done.Signal()
		case message.ErrorResponse:
			err = wrapAll(err, decodeError(r))
		default:
			if e := c.fallThrough(r); e != nil {
				// the connection will not be usable after this x_x
				return idPair{}, e
			}
		}
	}

	if r.Err != nil {
		return idPair{}, &clientConnectionError{err: r.Err}
	}

	return ids, err
}

func (c *baseConn) describe(r *buff.Reader, q query) (descPair, error) {
	w := buff.NewWriter(c.writeMemory[:0])
	w.BeginMessage(message.DescribeStatement)
	w.PushUint16(0) // no headers
	w.PushUint8(aspect.DataDescription)
	w.PushUint32(0) // no statement name
	w.EndMessage()

	w.BeginMessage(message.Sync)
	w.EndMessage()

	var descs descPair
	if e := w.Send(c.conn); e != nil {
		return descPair{}, &clientConnectionError{err: e}
	}

	var err error
	done := buff.NewSignal()

	for r.Next(done.Chan) {
		switch r.MsgType {
		case message.CommandDataDescription:
			ignoreHeaders(r)
			card := r.PopUint8()
			// input descriptor ID
			r.Discard(16)

			// input descriptor
			descs.in = r.PopBytes()

			// output descriptor
			r.Discard(16) // descriptor ID
			descs.out = r.PopBytes()

			if q.expCard == cardinality.One && card == cardinality.Many {
				err = &resultCardinalityMismatchError{msg: fmt.Sprintf(
					"the query has cardinality %v "+
						"which does not match the expected cardinality %v",
					cardinality.ToStr[card],
					cardinality.ToStr[q.expCard],
				)}
			}
		case message.ReadyForCommand:
			ignoreHeaders(r)
			r.Discard(1) // transaction state
			done.Signal()
		case message.ErrorResponse:
			err = wrapAll(err, decodeError(r))
		default:
			if e := c.fallThrough(r); e != nil {
				// the connection will not be usable after this x_x
				return descPair{}, e
			}
		}
	}

	if r.Err != nil {
		return descPair{}, &clientConnectionError{err: r.Err}
	}

	return descs, err
}

func (c *baseConn) execute(
	r *buff.Reader,
	out reflect.Value,
	q query,
	typ reflect.Type,
	cdcs codecPair,
) error {
	w := buff.NewWriter(c.writeMemory[:0])
	w.BeginMessage(message.Execute)
	w.PushUint16(0)       // no headers
	w.PushBytes([]byte{}) // no statement name
	if e := cdcs.in.Encode(w, q.args); e != nil {
		return &invalidArgumentError{msg: e.Error()}
	}
	w.EndMessage()

	w.BeginMessage(message.Sync)
	w.EndMessage()

	if e := w.Send(c.conn); e != nil {
		return &clientConnectionError{err: e}
	}

	tmp := out
	err := error(nil)
	if q.expCard == cardinality.One {
		err = errZeroResults
	}
	done := buff.NewSignal()

	for r.Next(done.Chan) {
		switch r.MsgType {
		case message.Data:
			elmCount := r.PopUint16()
			if elmCount != 1 {
				panic(fmt.Sprintf(
					"unexpected number of elements: expected 1, got %v",
					elmCount,
				))
			}
			elmLen := r.PopUint32()

			if !q.flat() {
				val := reflect.New(typ).Elem()
				s := r.PopSlice(elmLen)
				cdcs.out.Decode(s, val)
				tmp = reflect.Append(tmp, val)
			} else {
				cdcs.out.Decode(r.PopSlice(elmLen), out)
			}

			if err == errZeroResults {
				err = nil
			}
		case message.CommandComplete:
			ignoreHeaders(r)
			r.PopBytes() // command status
		case message.ReadyForCommand:
			ignoreHeaders(r)
			r.Discard(1) // transaction state
			done.Signal()
		case message.ErrorResponse:
			if err == errZeroResults {
				err = nil
			}

			err = wrapAll(err, decodeError(r))
		default:
			if e := c.fallThrough(r); e != nil {
				// the connection will not be usable after this x_x
				return e
			}
		}
	}

	if r.Err != nil {
		return &clientConnectionError{err: r.Err}
	}

	if !q.flat() {
		out.Set(tmp)
	}

	return err
}

func (c *baseConn) optimistic(
	r *buff.Reader,
	out reflect.Value,
	q query,
	typ reflect.Type,
	cdcs codecPair,
) error {
	w := buff.NewWriter(c.writeMemory[:0])
	w.BeginMessage(message.OptimisticExecute)
	w.PushUint16(0) // no headers
	w.PushUint8(q.fmt)
	w.PushUint8(q.expCard)
	w.PushString(q.cmd)
	w.PushUUID(cdcs.in.ID())
	w.PushUUID(cdcs.out.ID())
	if e := cdcs.in.Encode(w, q.args); e != nil {
		return &invalidArgumentError{msg: e.Error()}
	}
	w.EndMessage()

	w.BeginMessage(message.Sync)
	w.EndMessage()

	if e := w.Send(c.conn); e != nil {
		return &clientConnectionError{err: e}
	}

	tmp := out
	err := error(nil)
	if q.expCard == cardinality.One {
		err = errZeroResults
	}
	done := buff.NewSignal()

	for r.Next(done.Chan) {
		switch r.MsgType {
		case message.Data:
			elmCount := r.PopUint16()
			if elmCount != 1 {
				panic(fmt.Sprintf(
					"unexpected number of elements: expected 1, got %v",
					elmCount,
				))
			}

			elmLen := r.PopUint32()

			if !q.flat() {
				val := reflect.New(typ).Elem()
				cdcs.out.Decode(r.PopSlice(elmLen), val)
				tmp = reflect.Append(tmp, val)
			} else {
				cdcs.out.Decode(r.PopSlice(elmLen), out)
			}

			if err == errZeroResults {
				err = nil
			}
		case message.CommandComplete:
			ignoreHeaders(r)
			r.PopBytes() // command status
		case message.ReadyForCommand:
			ignoreHeaders(r)
			r.Discard(1) // transaction state
			done.Signal()
		case message.ErrorResponse:
			if err == errZeroResults {
				err = nil
			}

			err = wrapAll(err, decodeError(r))
		default:
			if e := c.fallThrough(r); e != nil {
				// the connection will not be usable after this x_x
				return e
			}
		}
	}

	if r.Err != nil {
		return &clientConnectionError{err: r.Err}
	}

	if !q.flat() {
		out.Set(tmp)
	}

	return err
}
