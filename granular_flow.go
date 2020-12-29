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
	"reflect"
	"unsafe"

	"github.com/edgedb/edgedb-go/internal/aspect"
	"github.com/edgedb/edgedb-go/internal/buff"
	"github.com/edgedb/edgedb-go/internal/codecs"
	"github.com/edgedb/edgedb-go/internal/format"
	"github.com/edgedb/edgedb-go/internal/message"
)

func (c *baseConn) granularFlow(
	r *buff.Reader,
	out reflect.Value,
	q query,
) (err error) {
	tp := out.Type()
	if !q.flat() {
		tp = tp.Elem()
	}

	ids, ok := c.getTypeIDs(q)
	if !ok {
		return c.pesimistic(r, out, q, tp)
	}

	in, ok := c.inCodecCache.Get(ids.in)
	if !ok {
		if desc, OK := descCache.Get(ids.in); OK {
			in, err = codecs.BuildCodec(buff.SimpleReader(desc.([]byte)))
			if err != nil {
				return err
			}
		} else {
			return c.pesimistic(r, out, q, tp)
		}
	}

	cOut, ok := c.outCodecCache.Get(codecKey{ID: ids.out, Type: tp})
	if !ok {
		if desc, ok := descCache.Get(ids.out); ok {
			d := buff.SimpleReader(desc.([]byte))
			cOut, err = codecs.BuildTypedCodec(d, tp)
			if err != nil {
				return err
			}
		} else {
			return c.pesimistic(r, out, q, tp)
		}
	}

	cdsc := codecPair{in: in.(codecs.Codec), out: cOut.(codecs.Codec)}
	return c.optimistic(r, out, q, tp, cdsc)
}

func (c *baseConn) pesimistic(
	r *buff.Reader,
	out reflect.Value,
	q query,
	tp reflect.Type,
) error {
	ids, err := c.prepare(r, q)
	if err != nil {
		return err
	}
	c.putTypeIDs(q, ids)

	descs, err := c.describe(r)
	if err != nil {
		return err
	}
	descCache.Put(ids.in, descs.in)
	descCache.Put(ids.out, descs.out)

	var cdcs codecPair
	cdcs.in, err = codecs.BuildCodec(buff.SimpleReader(descs.in))
	if err != nil {
		return err
	}

	if q.fmt == format.JSON {
		cdcs.out = codecs.JSONBytes
	} else {
		d := buff.SimpleReader(descs.out)
		cdcs.out, err = codecs.BuildTypedCodec(d, tp)
		if err != nil {
			return err
		}
	}

	c.inCodecCache.Put(ids.in, cdcs.in)
	c.outCodecCache.Put(codecKey{ID: ids.out, Type: tp}, cdcs.out)
	return c.execute(r, out, q, tp, cdcs)
}

func (c *baseConn) prepare(r *buff.Reader, q query) (idPair, error) {
	c.writer.BeginMessage(message.Prepare)
	c.writer.PushUint16(0) // no headers
	c.writer.PushUint8(q.fmt)
	c.writer.PushUint8(q.expCard)
	c.writer.PushBytes([]byte{}) // no statement name
	c.writer.PushString(q.cmd)
	c.writer.EndMessage()

	c.writer.BeginMessage(message.Sync)
	c.writer.EndMessage()

	if e := c.writer.Send(c.conn); e != nil {
		return idPair{}, e
	}

	var (
		err error
		ids idPair
	)

	done := buff.NewSignal()

	for r.Next(done.Chan) {
		switch r.MsgType {
		case message.PrepareComplete:
			r.Discard(2)     // number of headers, assume 0
			_ = r.PopUint8() // cardianlity
			ids = idPair{in: r.PopUUID(), out: r.PopUUID()}
		case message.ReadyForCommand:
			// header count (assume 0)
			// transaction state
			r.Discard(3)

			done.Signal()
		case message.ErrorResponse:
			err = wrapAll(err, decodeError(r))
			done.Signal()
		default:
			if e := c.fallThrough(r); e != nil {
				// the connection will not be usable after this x_x
				return idPair{}, e
			}
		}
	}

	if r.Err != nil {
		return idPair{}, r.Err
	}

	return ids, err
}

func (c *baseConn) describe(r *buff.Reader) (descPair, error) {
	c.writer.BeginMessage(message.DescribeStatement)
	c.writer.PushUint16(0) // no headers
	c.writer.PushUint8(aspect.DataDescription)
	c.writer.PushUint32(0) // no statement name
	c.writer.EndMessage()

	c.writer.BeginMessage(message.Sync)
	c.writer.EndMessage()

	var descs descPair
	if e := c.writer.Send(c.conn); e != nil {
		return descPair{}, e
	}

	var err error
	done := buff.NewSignal()

	for r.Next(done.Chan) {
		switch r.MsgType {
		case message.CommandDataDescription:
			// num headers is always 0
			// cardianlity
			// input descriptor ID
			r.Discard(19)

			// input descriptor
			descs.in = r.PopBytes()

			// output descriptor
			r.Discard(16) // descriptor ID
			descs.out = r.PopBytes()
		case message.ReadyForCommand:
			// header count (assume 0)
			// transaction state
			r.Discard(3)

			done.Signal()
		case message.ErrorResponse:
			err = wrapAll(err, decodeError(r))
			done.Signal()
		default:
			if e := c.fallThrough(r); e != nil {
				// the connection will not be usable after this x_x
				return descPair{}, e
			}
		}
	}

	if r.Err != nil {
		return descPair{}, r.Err
	}

	return descs, err
}

func (c *baseConn) execute(
	r *buff.Reader,
	out reflect.Value,
	q query,
	tp reflect.Type,
	cdcs codecPair,
) error {
	c.writer.BeginMessage(message.Execute)
	c.writer.PushUint16(0)       // no headers
	c.writer.PushBytes([]byte{}) // no statement name
	cdcs.in.Encode(c.writer, q.args)
	c.writer.EndMessage()

	c.writer.BeginMessage(message.Sync)
	c.writer.EndMessage()

	if e := c.writer.Send(c.conn); e != nil {
		return e
	}

	tmp := out
	err := ErrZeroResults
	done := buff.NewSignal()

	for r.Next(done.Chan) {
		switch r.MsgType {
		case message.Data:
			r.Discard(2) // number of data elements (always 1)

			if !q.flat() {
				val := reflect.New(tp).Elem()
				cdcs.out.Decode(r, unsafe.Pointer(val.UnsafeAddr()))
				tmp = reflect.Append(tmp, val)
			} else {
				cdcs.out.Decode(r, unsafe.Pointer(out.UnsafeAddr()))
			}

			if err == ErrZeroResults {
				err = nil
			}
		case message.CommandComplete:
			r.Discard(2) // header count (assume 0)
			r.PopBytes() // command status
		case message.ReadyForCommand:
			// header count (assume 0)
			// transaction state
			r.Discard(3)
			done.Signal()
		case message.ErrorResponse:
			if err == ErrZeroResults {
				err = nil
			}

			err = wrapAll(err, decodeError(r))
			done.Signal()
		default:
			if e := c.fallThrough(r); e != nil {
				// the connection will not be usable after this x_x
				return e
			}
		}
	}

	if r.Err != nil {
		return r.Err
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
	tp reflect.Type,
	cdcs codecPair,
) error {
	c.writer.BeginMessage(message.OptimisticExecute)
	c.writer.PushUint16(0) // no headers
	c.writer.PushUint8(q.fmt)
	c.writer.PushUint8(q.expCard)
	c.writer.PushString(q.cmd)
	c.writer.PushUUID(cdcs.in.ID())
	c.writer.PushUUID(cdcs.out.ID())
	cdcs.in.Encode(c.writer, q.args)
	c.writer.EndMessage()

	c.writer.BeginMessage(message.Sync)
	c.writer.EndMessage()

	if e := c.writer.Send(c.conn); e != nil {
		return e
	}

	tmp := out
	err := ErrZeroResults
	done := buff.NewSignal()

	for r.Next(done.Chan) {
		switch r.MsgType {
		case message.Data:
			r.Discard(2) // number of data elements (always 1)

			if !q.flat() {
				val := reflect.New(tp).Elem()
				cdcs.out.Decode(r, unsafe.Pointer(val.UnsafeAddr()))
				tmp = reflect.Append(tmp, val)
			} else {
				cdcs.out.Decode(r, unsafe.Pointer(out.UnsafeAddr()))
			}

			if err == ErrZeroResults {
				err = nil
			}
		case message.CommandComplete:
			r.Discard(2) // header count (assume 0)
			r.PopBytes() // command status
		case message.ReadyForCommand:
			// header count (assume 0)
			// transaction state
			r.Discard(3)
			done.Signal()
		case message.ErrorResponse:
			if err == ErrZeroResults {
				err = nil
			}

			err = wrapAll(err, decodeError(r))
			done.Signal()
		default:
			if e := c.fallThrough(r); e != nil {
				// the connection will not be usable after this x_x
				return e
			}
		}
	}

	if r.Err != nil {
		return r.Err
	}

	if !q.flat() {
		out.Set(tmp)
	}

	return err
}
