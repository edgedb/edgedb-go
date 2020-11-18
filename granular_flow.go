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
	"net"
	"reflect"
	"unsafe"

	"github.com/edgedb/edgedb-go/protocol/aspect"
	"github.com/edgedb/edgedb-go/protocol/buff"
	"github.com/edgedb/edgedb-go/protocol/codecs"
	"github.com/edgedb/edgedb-go/protocol/format"
	"github.com/edgedb/edgedb-go/protocol/message"
)

func (c *Client) granularFlow(
	ctx context.Context,
	conn net.Conn,
	out reflect.Value,
	q query,
) (err error) {
	tp := out.Type()
	if !q.flat() {
		tp = tp.Elem()
	}

	ids, ok := c.getTypeIDs(q)
	if !ok {
		return c.pesimistic(ctx, conn, out, q, tp)
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
			return c.pesimistic(ctx, conn, out, q, tp)
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
			return c.pesimistic(ctx, conn, out, q, tp)
		}
	}

	cdsc := codecPair{in: in.(codecs.Codec), out: cOut.(codecs.Codec)}
	return c.optimistic(ctx, conn, out, q, tp, cdsc)
}

func (c *Client) pesimistic(
	ctx context.Context,
	conn net.Conn,
	out reflect.Value,
	q query,
	tp reflect.Type,
) error {
	ids, err := c.prepare(ctx, conn, q)
	if err != nil {
		return err
	}
	c.putTypeIDs(q, ids)

	descs, err := c.describe(ctx, conn)
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
	return c.execute(ctx, conn, out, q, tp, cdcs)
}

func (c *Client) prepare(
	ctx context.Context,
	conn net.Conn,
	q query,
) (ids idPair, err error) {
	c.buf.Reset()
	c.buf.BeginMessage(message.Prepare)
	c.buf.PushUint16(0) // no headers
	c.buf.PushUint8(q.fmt)
	c.buf.PushUint8(q.expCard)
	c.buf.PushBytes([]byte{}) // no statement name
	c.buf.PushString(q.cmd)
	c.buf.EndMessage()

	c.buf.BeginMessage(message.Sync)
	c.buf.EndMessage()

	err = writeAndRead(ctx, conn, c.buf.Unwrap())
	if err != nil {
		return ids, err
	}

	for c.buf.Next() {
		switch c.buf.MsgType {
		case message.PrepareComplete:
			c.buf.PopUint16() // number of headers, assume 0

			// todo assert cardinality matches query
			c.buf.PopUint8() // cardianlity

			ids = idPair{
				in:  c.buf.PopUUID(),
				out: c.buf.PopUUID(),
			}
		case message.ReadyForCommand:
			c.buf.PopUint16() // header count (assume 0)
			c.buf.PopUint8()  // transaction state
		case message.ErrorResponse:
			return ids, decodeError(c.buf)
		default:
			return ids, fmt.Errorf(
				"unexpected message type: 0x%x",
				c.buf.MsgType,
			)
		}
	}

	return ids, nil
}

func (c *Client) describe(
	ctx context.Context,
	conn net.Conn,
) (descs descPair, err error) {
	c.buf.Reset()
	c.buf.BeginMessage(message.DescribeStatement)
	c.buf.PushUint16(0) // no headers
	c.buf.PushUint8(aspect.DataDescription)
	c.buf.PushUint32(0) // no statement name
	c.buf.EndMessage()

	c.buf.BeginMessage(message.Sync)
	c.buf.EndMessage()

	err = writeAndRead(ctx, conn, c.buf.Unwrap())
	if err != nil {
		return descs, err
	}

	for c.buf.Next() {
		switch c.buf.MsgType {
		case message.CommandDataDescription:
			c.buf.PopUint16() // num headers is always 0
			c.buf.PopUint8()  // cardianlity

			// input descriptor
			c.buf.PopUUID()
			descs.in = c.buf.PopBytes()

			// output descriptor
			c.buf.PopUUID()
			descs.out = c.buf.PopBytes()
		case message.ReadyForCommand:
			c.buf.PopUint16() // header count (assume 0)
			c.buf.PopUint8()  // transaction state
		case message.ErrorResponse:
			return descs, decodeError(c.buf)
		default:
			return descs, fmt.Errorf(
				"unexpected message type: 0x%x",
				c.buf.MsgType,
			)
		}
	}

	return descs, nil
}

func (c *Client) execute(
	ctx context.Context,
	conn net.Conn,
	out reflect.Value,
	q query,
	tp reflect.Type,
	cdcs codecPair,
) error {
	c.buf.Reset()
	c.buf.BeginMessage(message.Execute)
	c.buf.PushUint16(0)       // no headers
	c.buf.PushBytes([]byte{}) // no statement name
	cdcs.in.Encode(c.buf, q.args)
	c.buf.EndMessage()

	c.buf.BeginMessage(message.Sync)
	c.buf.EndMessage()

	err := writeAndRead(ctx, conn, c.buf.Unwrap())
	if err != nil {
		return err
	}

	tmp := out
	err = ErrorZeroResults
	for c.buf.Next() {
		switch c.buf.MsgType {
		case message.Data:
			c.buf.PopUint16() // number of data elements (always 1)

			if !q.flat() {
				val := reflect.New(tp).Elem()
				cdcs.out.Decode(c.buf, unsafe.Pointer(val.UnsafeAddr()))
				tmp = reflect.Append(tmp, val)
			} else {
				cdcs.out.Decode(c.buf, unsafe.Pointer(out.UnsafeAddr()))
			}

			err = nil
		case message.CommandComplete:
			c.buf.PopUint16() // header count (assume 0)
			c.buf.PopBytes()  // command status
		case message.ReadyForCommand:
			c.buf.PopUint16() // header count (assume 0)
			c.buf.PopUint8()  // transaction state
		case message.ErrorResponse:
			return decodeError(c.buf)
		default:
			e := c.fallThrough(c.buf)
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

func (c *Client) optimistic(
	ctx context.Context,
	conn net.Conn,
	out reflect.Value,
	q query,
	tp reflect.Type,
	cdcs codecPair,
) error {
	c.buf.Reset()
	c.buf.BeginMessage(message.OptimisticExecute)
	c.buf.PushUint16(0) // no headers
	c.buf.PushUint8(q.fmt)
	c.buf.PushUint8(q.expCard)
	c.buf.PushString(q.cmd)
	c.buf.PushUUID(cdcs.in.ID())
	c.buf.PushUUID(cdcs.out.ID())
	cdcs.in.Encode(c.buf, q.args)
	c.buf.EndMessage()

	c.buf.BeginMessage(message.Sync)
	c.buf.EndMessage()

	err := writeAndRead(ctx, conn, c.buf.Unwrap())
	if err != nil {
		return err
	}

	tmp := out
	err = ErrorZeroResults
	for c.buf.Next() {
		switch c.buf.MsgType {
		case message.Data:
			c.buf.PopUint16() // number of data elements (always 1)

			if !q.flat() {
				val := reflect.New(tp).Elem()
				cdcs.out.Decode(c.buf, unsafe.Pointer(val.UnsafeAddr()))
				tmp = reflect.Append(tmp, val)
			} else {
				cdcs.out.Decode(c.buf, unsafe.Pointer(out.UnsafeAddr()))
			}
			err = nil
		case message.CommandComplete:
			c.buf.PopUint16() // header count (assume 0)
			c.buf.PopBytes()  // command status
		case message.ReadyForCommand:
			c.buf.PopUint16() // header count (assume 0)
			c.buf.PopUint8()  // transaction state
		case message.ErrorResponse:
			return decodeError(c.buf)
		default:
			e := c.fallThrough(c.buf)
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
