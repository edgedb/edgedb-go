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
	"unsafe"

	"github.com/edgedb/edgedb-go/internal/aspect"
	"github.com/edgedb/edgedb-go/internal/buff"
	"github.com/edgedb/edgedb-go/internal/cardinality"
	"github.com/edgedb/edgedb-go/internal/codecs"
	"github.com/edgedb/edgedb-go/internal/descriptor"
	"github.com/edgedb/edgedb-go/internal/format"
	"github.com/edgedb/edgedb-go/internal/header"
	"github.com/edgedb/edgedb-go/internal/message"
)

func (c *protocolConnection) execGranularFlow(
	r *buff.Reader,
	q *gfQuery,
) error {
	ids, ok := c.getCachedTypeIDs(q)
	if !ok {
		return c.pesimistic(r, q)
	}

	cdcs, err := c.codecsFromIDs(ids, q)
	if err != nil {
		return err
	} else if cdcs == nil {
		return c.pesimistic(r, q)
	}

	// When descriptors are returned the codec ids sent didn't match the
	// server's.  The codecs should be rebuilt with the new descriptors and the
	// execution retried.
	descs, err := c.optimistic(r, q, cdcs)
	if err != nil {
		return err
	} else if descs == nil { // optimistic execute succeeded
		return nil
	}

	cdcs, err = c.codecsFromDescriptors(q, descs)
	if err != nil {
		return err
	}

	return c.execute(r, q, cdcs)
}

func (c *protocolConnection) pesimistic(r *buff.Reader, q *gfQuery) error {
	err := c.prepare(r, q)
	if err != nil {
		return err
	}

	descs, err := c.describe(r, q)
	if err != nil {
		return err
	}

	cdcs, err := c.codecsFromDescriptors(q, descs)
	if err != nil {
		return err
	}

	return c.execute(r, q, cdcs)
}

func (c *protocolConnection) codecsFromIDs(
	ids *idPair,
	q *gfQuery,
) (*codecPair, error) {
	var err error

	in, ok := c.inCodecCache.Get(ids.in)
	if !ok {
		desc, OK := descCache.Get(ids.in)
		if !OK {
			return nil, nil
		}

		in, err = codecs.BuildEncoder(
			desc.(descriptor.Descriptor),
			c.protocolVersion,
		)
		if err != nil {
			return nil, &invalidArgumentError{msg: err.Error()}
		}
	}

	out, ok := c.outCodecCache.Get(codecKey{ID: ids.out, Type: q.outType})
	if !ok {
		desc, OK := descCache.Get(ids.out)
		if !OK {
			return nil, nil
		}

		d := desc.(descriptor.Descriptor)
		path := codecs.Path(q.outType.String())
		out, err = codecs.BuildDecoder(d, q.outType, path)
		if err != nil {
			return nil, &invalidArgumentError{msg: fmt.Sprintf(
				"the \"out\" argument does not match query schema: %v", err)}
		}
	}

	return &codecPair{in: in.(codecs.Encoder), out: out.(codecs.Decoder)}, nil
}

func (c *protocolConnection) codecsFromDescriptors(
	q *gfQuery,
	descs *descPair,
) (*codecPair, error) {
	var cdcs codecPair
	var err error
	cdcs.in, err = codecs.BuildEncoder(descs.in, c.protocolVersion)
	if err != nil {
		return nil, &invalidArgumentError{msg: err.Error()}
	}

	if q.fmt == format.JSON {
		cdcs.out = codecs.JSONBytes
	} else {
		path := codecs.Path(q.outType.String())
		cdcs.out, err = codecs.BuildDecoder(descs.out, q.outType, path)
		if err != nil {
			err = fmt.Errorf(
				"the \"out\" argument does not match query schema: %v",
				err,
			)
			return nil, &invalidArgumentError{msg: err.Error()}
		}
	}

	c.inCodecCache.Put(cdcs.in.DescriptorID(), cdcs.in)
	c.outCodecCache.Put(
		codecKey{ID: cdcs.out.DescriptorID(), Type: q.outType},
		cdcs.out,
	)

	return &cdcs, nil
}

func (c *protocolConnection) prepare(r *buff.Reader, q *gfQuery) error {
	headers := copyHeaders(q.headers)

	if c.protocolVersion.GTE(protocolVersion0p10) {
		headers[header.ExplicitObjectIDs] = []byte("true")
	}

	w := buff.NewWriter(c.writeMemory[:0])
	w.BeginMessage(message.Prepare)
	writeHeaders(w, headers)
	w.PushUint8(q.fmt)
	w.PushUint8(q.expCard)
	w.PushUint32(0) // no statement name
	w.PushString(q.cmd)
	w.EndMessage()

	w.BeginMessage(message.Sync)
	w.EndMessage()

	if e := c.soc.WriteAll(w.Unwrap()); e != nil {
		return &clientConnectionClosedError{err: e}
	}

	var (
		err error
	)

	done := buff.NewSignal()

	for r.Next(done.Chan) {
		switch r.MsgType {
		case message.PrepareComplete:
			c.cacheCapabilities(q, decodeHeaders(r))
			r.Discard(1) // cardinality
			ids := idPair{in: r.PopUUID(), out: r.PopUUID()}
			c.cacheTypeIDs(q, ids)
		case message.ReadyForCommand:
			decodeReadyForCommandMsg(r)
			done.Signal()
		case message.ErrorResponse:
			err = wrapAll(err, decodeErrorResponseMsg(r, q.cmd))
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
	return err
}

func (c *protocolConnection) describe(
	r *buff.Reader,
	q *gfQuery,
) (*descPair, error) {
	w := buff.NewWriter(c.writeMemory[:0])
	w.BeginMessage(message.DescribeStatement)
	w.PushUint16(0) // no headers
	w.PushUint8(aspect.DataDescription)
	w.PushUint32(0) // no statement name
	w.EndMessage()

	w.BeginMessage(message.Sync)
	w.EndMessage()

	if e := c.soc.WriteAll(w.Unwrap()); e != nil {
		return nil, &clientConnectionClosedError{err: e}
	}

	var (
		descs *descPair
		err   error
	)
	done := buff.NewSignal()

	for r.Next(done.Chan) {
		switch r.MsgType {
		case message.CommandDataDescription:
			descs, _, err = c.decodeCommandDataDescriptionMsg(r, q)
		case message.ReadyForCommand:
			decodeReadyForCommandMsg(r)
			done.Signal()
		case message.ErrorResponse:
			err = wrapAll(err, decodeErrorResponseMsg(r, q.cmd))
		default:
			if e := c.fallThrough(r); e != nil {
				// the connection will not be usable after this x_x
				return nil, e
			}
		}
	}

	if r.Err != nil {
		return nil, r.Err
	}

	return descs, err
}

func (c *protocolConnection) execute(
	r *buff.Reader,
	q *gfQuery,
	cdcs *codecPair,
) error {
	w := buff.NewWriter(c.writeMemory[:0])
	w.BeginMessage(message.Execute)
	writeHeaders(w, q.headers)
	w.PushUint32(0) // no statement name
	if e := cdcs.in.Encode(w, q.args, codecs.Path("args"), true); e != nil {
		return &invalidArgumentError{msg: e.Error()}
	}
	w.EndMessage()

	w.BeginMessage(message.Sync)
	w.EndMessage()

	if e := c.soc.WriteAll(w.Unwrap()); e != nil {
		return &clientConnectionClosedError{err: e}
	}

	tmp := q.out
	err := error(nil)
	if q.expCard == cardinality.AtMostOne {
		err = errZeroResults
	}
	done := buff.NewSignal()

	for r.Next(done.Chan) {
		switch r.MsgType {
		case message.Data:
			val, ok, e := decodeDataMsg(r, q, cdcs)
			if e != nil {
				if err == errZeroResults {
					err = e
				} else {
					err = wrapAll(err, e)
				}
			}
			if ok {
				tmp = reflect.Append(tmp, val)
			}

			if err == errZeroResults {
				err = nil
			}
		case message.CommandComplete:
			decodeCommandCompleteMsg(r)
		case message.ReadyForCommand:
			decodeReadyForCommandMsg(r)
			done.Signal()
		case message.ErrorResponse:
			if err == errZeroResults {
				err = nil
			}

			err = wrapAll(err, decodeErrorResponseMsg(r, q.cmd))
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
		q.out.Set(tmp)
	}

	return err
}

func (c *protocolConnection) optimistic(
	r *buff.Reader,
	q *gfQuery,
	cdcs *codecPair,
) (*descPair, error) {
	headers := copyHeaders(q.headers)

	if c.protocolVersion.GTE(protocolVersion0p10) {
		headers[header.ExplicitObjectIDs] = []byte("true")
	}

	w := buff.NewWriter(c.writeMemory[:0])
	w.BeginMessage(message.OptimisticExecute)
	writeHeaders(w, headers)
	w.PushUint8(q.fmt)
	w.PushUint8(q.expCard)
	w.PushString(q.cmd)
	w.PushUUID(cdcs.in.DescriptorID())
	w.PushUUID(cdcs.out.DescriptorID())
	if e := cdcs.in.Encode(w, q.args, codecs.Path("args"), true); e != nil {
		return nil, &invalidArgumentError{msg: e.Error()}
	}
	w.EndMessage()

	w.BeginMessage(message.Sync)
	w.EndMessage()

	if e := c.soc.WriteAll(w.Unwrap()); e != nil {
		return nil, &clientConnectionClosedError{err: e}
	}

	tmp := q.out
	err := error(nil)
	if q.expCard == cardinality.AtMostOne {
		err = errZeroResults
	}
	done := buff.NewSignal()

	var descs *descPair
	for r.Next(done.Chan) {
		switch r.MsgType {
		case message.Data:
			val, ok, e := decodeDataMsg(r, q, cdcs)
			if e != nil {
				if err == errZeroResults {
					err = e
				} else {
					err = wrapAll(err, e)
				}
			}
			if ok {
				tmp = reflect.Append(tmp, val)
			}

			if err == errZeroResults {
				err = nil
			}
		case message.CommandComplete:
			decodeCommandCompleteMsg(r)
		case message.CommandDataDescription:
			var headers msgHeaders
			descs, headers, err = c.decodeCommandDataDescriptionMsg(r, q)
			c.cacheCapabilities(q, headers)
		case message.ReadyForCommand:
			decodeReadyForCommandMsg(r)
			done.Signal()
		case message.ErrorResponse:
			if err == errZeroResults {
				err = nil
			}

			err = wrapAll(err, decodeErrorResponseMsg(r, q.cmd))
		default:
			if e := c.fallThrough(r); e != nil {
				// the connection will not be usable after this x_x
				return nil, e
			}
		}
	}

	if r.Err != nil {
		return nil, r.Err
	}

	if !q.flat() {
		q.out.Set(tmp)
	}

	return descs, err
}

func decodeCommandCompleteMsg(r *buff.Reader) {
	ignoreHeaders(r)
	r.PopBytes() // command status
}

func decodeReadyForCommandMsg(r *buff.Reader) {
	ignoreHeaders(r)
	r.Discard(1) // transaction state
}

func decodeDataMsg(
	r *buff.Reader,
	q *gfQuery,
	cdcs *codecPair,
) (reflect.Value, bool, error) {
	elmCount := r.PopUint16()
	if elmCount != 1 {
		return reflect.Value{}, false, fmt.Errorf(
			"unexpected number of elements: expected 1, got %v", elmCount)
	}
	elmLen := r.PopUint32()

	if !q.flat() {
		val := reflect.New(q.outType).Elem()
		err := cdcs.out.Decode(
			r.PopSlice(elmLen),
			unsafe.Pointer(val.UnsafeAddr()),
		)
		if err != nil {
			return reflect.Value{}, false, err
		}
		return val, true, nil
	}

	err := cdcs.out.Decode(
		r.PopSlice(elmLen),
		unsafe.Pointer(q.out.UnsafeAddr()),
	)
	if err != nil {
		return reflect.Value{}, false, err
	}

	return reflect.Value{}, false, nil
}

func (c *protocolConnection) decodeCommandDataDescriptionMsg(
	r *buff.Reader,
	q *gfQuery,
) (*descPair, msgHeaders, error) {
	headers := decodeHeaders(r)
	card := r.PopUint8()

	var (
		descs descPair
		err   error
	)
	id := r.PopUUID() // in descriptor id
	descs.in, err = descriptor.Pop(
		r.PopSlice(r.PopUint32()),
		c.protocolVersion,
	)
	if err != nil {
		return nil, nil, err
	}
	if descs.in.ID != id {
		return nil, nil, &clientError{msg: fmt.Sprintf(
			"unexpected in descriptor id: %v", descs.in.ID)}
	}

	id = r.PopUUID() // output descriptor ID
	descs.out, err = descriptor.Pop(
		r.PopSlice(r.PopUint32()),
		c.protocolVersion,
	)
	if err != nil {
		return nil, nil, err
	}
	if descs.out.ID != id {
		return nil, nil, &clientError{msg: fmt.Sprintf(
			"unexpected out descriptor id: %v", descs.in.ID)}
	}

	if q.expCard == cardinality.AtMostOne && card == cardinality.Many {
		return nil, nil, &resultCardinalityMismatchError{msg: fmt.Sprintf(
			"the query has cardinality %v "+
				"which does not match the expected cardinality %v",
			cardinality.ToStr[card],
			cardinality.ToStr[q.expCard],
		)}
	}

	descCache.Put(descs.in.ID, descs.in)
	descCache.Put(descs.out.ID, descs.out)
	return &descs, headers, nil
}
