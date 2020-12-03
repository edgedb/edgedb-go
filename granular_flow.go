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
	"context"
	"fmt"
	"reflect"
	"unsafe"

	"github.com/edgedb/edgedb-go/protocol/aspect"
	"github.com/edgedb/edgedb-go/protocol/buff"
	"github.com/edgedb/edgedb-go/protocol/codecs"
	"github.com/edgedb/edgedb-go/protocol/format"
	"github.com/edgedb/edgedb-go/protocol/message"
)

func (c *baseConn) granularFlow(
	ctx context.Context,
	out reflect.Value,
	q query,
) (err error) {
	tp := out.Type()
	if !q.flat() {
		tp = tp.Elem()
	}

	ids, ok := c.getTypeIDs(q)
	if !ok {
		return c.pesimistic(ctx, out, q, tp)
	}

	in, ok := c.inCodecCache.Get(ids.in)
	if !ok {
		if desc, OK := descCache.Get(ids.in); OK {
			buf := buff.NewReader(desc.([]byte))
			in, err = codecs.BuildCodec(buf)
			if err != nil {
				return err
			}
		} else {
			return c.pesimistic(ctx, out, q, tp)
		}
	}

	cOut, ok := c.outCodecCache.Get(codecKey{ID: ids.out, Type: tp})
	if !ok {
		if desc, ok := descCache.Get(ids.out); ok {
			buf := buff.NewReader(desc.([]byte))
			cOut, err = codecs.BuildTypedCodec(buf, tp)
			if err != nil {
				return err
			}
		} else {
			return c.pesimistic(ctx, out, q, tp)
		}
	}

	cdsc := codecPair{in: in.(codecs.Codec), out: cOut.(codecs.Codec)}
	return c.optimistic(ctx, out, q, tp, cdsc)
}

func (c *baseConn) pesimistic(
	ctx context.Context,
	out reflect.Value,
	q query,
	tp reflect.Type,
) error {
	ids, err := c.prepare(ctx, q)
	if err != nil {
		return err
	}
	c.putTypeIDs(q, ids)

	descs, err := c.describe(ctx)
	if err != nil {
		return err
	}
	descCache.Put(ids.in, descs.in)
	descCache.Put(ids.out, descs.out)

	var cdcs codecPair
	cdcs.in, err = codecs.BuildCodec(buff.NewReader(descs.in))
	if err != nil {
		return err
	}

	if q.fmt == format.JSON {
		cdcs.out = codecs.JSONBytes
	} else {
		cdcs.out, err = codecs.BuildTypedCodec(buff.NewReader(descs.out), tp)
		if err != nil {
			return err
		}
	}

	c.inCodecCache.Put(ids.in, cdcs.in)
	c.outCodecCache.Put(codecKey{ID: ids.out, Type: tp}, cdcs.out)
	return c.execute(ctx, out, q, tp, cdcs)
}

func (c *baseConn) prepare(
	ctx context.Context,
	q query,
) (ids idPair, err error) {
	buf := buff.New(nil)
	buf.BeginMessage(message.Prepare)
	buf.PushUint16(0) // no headers
	buf.PushUint8(q.fmt)
	buf.PushUint8(q.expCard)
	buf.PushBytes([]byte{}) // no statement name
	buf.PushString(q.cmd)
	buf.EndMessage()

	buf.BeginMessage(message.Sync)
	buf.EndMessage()

	err = c.writeAndRead(ctx, buf.Unwrap())
	if err != nil {
		return ids, err
	}

	for buf.Next() {
		switch buf.MsgType {
		case message.PrepareComplete:
			buf.PopUint16() // number of headers, assume 0

			// todo assert cardinality matches query
			buf.PopUint8() // cardianlity

			ids = idPair{
				in:  buf.PopUUID(),
				out: buf.PopUUID(),
			}
		case message.ReadyForCommand:
			buf.PopUint16() // header count (assume 0)
			buf.PopUint8()  // transaction state
		case message.ErrorResponse:
			return ids, decodeError(buf)
		default:
			return ids, fmt.Errorf(
				"unexpected message type: 0x%x",
				buf.MsgType,
			)
		}
	}

	return ids, nil
}

func (c *baseConn) describe(ctx context.Context) (descPair, error) {
	buf := buff.New(c.buffer[:0])
	buf.BeginMessage(message.DescribeStatement)
	buf.PushUint16(0) // no headers
	buf.PushUint8(aspect.DataDescription)
	buf.PushUint32(0) // no statement name
	buf.EndMessage()

	buf.BeginMessage(message.Sync)
	buf.EndMessage()

	var descs descPair
	err := c.writeAndRead(ctx, buf.Unwrap())
	if err != nil {
		return descs, err
	}

	for buf.Next() {
		switch buf.MsgType {
		case message.CommandDataDescription:
			buf.PopUint16() // num headers is always 0
			buf.PopUint8()  // cardianlity

			// input descriptor
			buf.PopUUID()
			descs.in = buf.PopBytes()

			// output descriptor
			buf.PopUUID()
			descs.out = buf.PopBytes()
		case message.ReadyForCommand:
			buf.PopUint16() // header count (assume 0)
			buf.PopUint8()  // transaction state
		case message.ErrorResponse:
			return descs, decodeError(buf)
		default:
			return descs, fmt.Errorf(
				"unexpected message type: 0x%x",
				buf.MsgType,
			)
		}
	}

	return descs, nil
}

func (c *baseConn) execute(
	ctx context.Context,
	out reflect.Value,
	q query,
	tp reflect.Type,
	cdcs codecPair,
) error {
	buf := buff.New(c.buffer[:0])
	buf.BeginMessage(message.Execute)
	buf.PushUint16(0)       // no headers
	buf.PushBytes([]byte{}) // no statement name
	cdcs.in.Encode(buf, q.args)
	buf.EndMessage()

	buf.BeginMessage(message.Sync)
	buf.EndMessage()

	err := c.writeAndRead(ctx, buf.Unwrap())
	if err != nil {
		return err
	}

	tmp := out
	err = ErrorZeroResults
	for buf.Next() {
		switch buf.MsgType {
		case message.Data:
			buf.PopUint16() // number of data elements (always 1)

			if !q.flat() {
				val := reflect.New(tp).Elem()
				cdcs.out.Decode(buf, unsafe.Pointer(val.UnsafeAddr()))
				tmp = reflect.Append(tmp, val)
			} else {
				cdcs.out.Decode(buf, unsafe.Pointer(out.UnsafeAddr()))
			}

			err = nil
		case message.CommandComplete:
			buf.PopUint16() // header count (assume 0)
			buf.PopBytes()  // command status
		case message.ReadyForCommand:
			buf.PopUint16() // header count (assume 0)
			buf.PopUint8()  // transaction state
		case message.ErrorResponse:
			return decodeError(buf)
		default:
			e := c.fallThrough(buf)
			if e != nil {
				return e
			}
		}
	}

	if !q.flat() {
		out.Set(tmp)
	}

	return err
}

func (c *baseConn) optimistic(
	ctx context.Context,
	out reflect.Value,
	q query,
	tp reflect.Type,
	cdcs codecPair,
) error {
	buf := buff.New(c.buffer[:0])
	buf.BeginMessage(message.OptimisticExecute)
	buf.PushUint16(0) // no headers
	buf.PushUint8(q.fmt)
	buf.PushUint8(q.expCard)
	buf.PushString(q.cmd)
	buf.PushUUID(cdcs.in.ID())
	buf.PushUUID(cdcs.out.ID())
	cdcs.in.Encode(buf, q.args)
	buf.EndMessage()

	buf.BeginMessage(message.Sync)
	buf.EndMessage()

	err := c.writeAndRead(ctx, buf.Unwrap())
	if err != nil {
		return err
	}

	tmp := out
	err = ErrorZeroResults
	for buf.Next() {
		switch buf.MsgType {
		case message.Data:
			buf.PopUint16() // number of data elements (always 1)

			if !q.flat() {
				val := reflect.New(tp).Elem()
				cdcs.out.Decode(buf, unsafe.Pointer(val.UnsafeAddr()))
				tmp = reflect.Append(tmp, val)
			} else {
				cdcs.out.Decode(buf, unsafe.Pointer(out.UnsafeAddr()))
			}
			err = nil
		case message.CommandComplete:
			buf.PopUint16() // header count (assume 0)
			buf.PopBytes()  // command status
		case message.ReadyForCommand:
			buf.PopUint16() // header count (assume 0)
			buf.PopUint8()  // transaction state
		case message.ErrorResponse:
			return decodeError(buf)
		default:
			e := c.fallThrough(buf)
			if e != nil {
				return e
			}
		}
	}

	if !q.flat() {
		out.Set(tmp)
	}

	return err
}
