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
			buf := buff.NewMessage(desc.([]byte))
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
			buf := buff.NewMessage(desc.([]byte))
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
	ids, err := prepare(ctx, conn, q)
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
	cdcs.in, err = codecs.BuildCodec(buff.NewMessage(descs.in))
	if err != nil {
		return err
	}

	if q.fmt == format.JSON {
		cdcs.out = codecs.JSONBytes
	} else {
		cdcs.out, err = codecs.BuildTypedCodec(buff.NewMessage(descs.out), tp)
		if err != nil {
			return err
		}
	}

	c.inCodecCache.Put(ids.in, cdcs.in)
	c.outCodecCache.Put(codecKey{ID: ids.out, Type: tp}, cdcs.out)
	return c.execute(ctx, conn, out, q, tp, cdcs)
}

func prepare(
	ctx context.Context,
	conn net.Conn,
	q query,
) (ids idPair, err error) {
	buf := buff.NewWriter(nil)
	buf.BeginMessage(message.Prepare)
	buf.PushUint16(0) // no headers
	buf.PushUint8(q.fmt)
	buf.PushUint8(q.expCard)
	buf.PushBytes([]byte{}) // no statement name
	buf.PushString(q.cmd)
	buf.EndMessage()

	buf.BeginMessage(message.Sync)
	buf.EndMessage()

	err = writeAndRead(ctx, conn, buf.Unwrap())
	if err != nil {
		return ids, err
	}

	for buf.Next() {
		msg := buf.PopMessage()

		switch msg.Type {
		case message.PrepareComplete:
			msg.PopUint16() // number of headers, assume 0

			// todo assert cardinality matches query
			msg.PopUint8() // cardianlity

			ids = idPair{
				in:  msg.PopUUID(),
				out: msg.PopUUID(),
			}
		case message.ReadyForCommand:
			msg.PopUint16() // header count (assume 0)
			msg.PopUint8()  // transaction state
		case message.ErrorResponse:
			return ids, decodeError(msg)
		default:
			return ids, fmt.Errorf("unexpected message type: 0x%x", msg.Type)
		}

		msg.Finish()
	}

	return ids, nil
}

func (c *Client) describe(
	ctx context.Context,
	conn net.Conn,
) (descs descPair, err error) {
	buf := buff.NewWriter(nil)
	buf.BeginMessage(message.DescribeStatement)
	buf.PushUint16(0) // no headers
	buf.PushUint8(aspect.DataDescription)
	buf.PushUint32(0) // no statement name
	buf.EndMessage()

	buf.BeginMessage(message.Sync)
	buf.EndMessage()

	err = writeAndRead(ctx, conn, buf.Unwrap())
	if err != nil {
		return descs, err
	}

	for buf.Next() {
		msg := buf.PopMessage()

		switch msg.Type {
		case message.CommandDataDescription:
			msg.PopUint16() // num headers is always 0
			msg.PopUint8()  // cardianlity

			// input descriptor
			msg.PopUUID()
			descs.in = msg.PopBytes()

			// output descriptor
			msg.PopUUID()
			descs.out = msg.PopBytes()
		case message.ReadyForCommand:
			msg.PopUint16() // header count (assume 0)
			msg.PopUint8()  // transaction state
		case message.ErrorResponse:
			return descs, decodeError(msg)
		default:
			return descs, fmt.Errorf("unexpected message type: 0x%x", msg.Type)
		}

		msg.Finish()
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
	buf := buff.NewWriter(nil)
	buf.BeginMessage(message.Execute)
	buf.PushUint16(0)       // no headers
	buf.PushBytes([]byte{}) // no statement name
	cdcs.in.Encode(buf, q.args)
	buf.EndMessage()

	buf.BeginMessage(message.Sync)
	buf.EndMessage()

	err := writeAndRead(ctx, conn, buf.Unwrap())
	if err != nil {
		return err
	}

	tmp := out
	err = ErrorZeroResults
	for buf.Next() {
		msg := buf.PopMessage()

		switch msg.Type {
		case message.Data:
			msg.PopUint16() // number of data elements (always 1)

			if !q.flat() {
				val := reflect.New(tp).Elem()
				cdcs.out.Decode(msg, unsafe.Pointer(val.UnsafeAddr()))
				tmp = reflect.Append(tmp, val)
			} else {
				cdcs.out.Decode(msg, unsafe.Pointer(out.UnsafeAddr()))
			}

			err = nil
		case message.CommandComplete:
			msg.PopUint16() // header count (assume 0)
			msg.PopBytes()  // command status
		case message.ReadyForCommand:
			msg.PopUint16() // header count (assume 0)
			msg.PopUint8()  // transaction state
		case message.ErrorResponse:
			return decodeError(msg)
		default:
			e := c.fallThrough(msg)
			if e != nil {
				return e
			}
		}
		msg.Finish()
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
	buf := buff.NewWriter(c.buffer[:0])
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

	err := writeAndRead(ctx, conn, buf.Unwrap())
	if err != nil {
		return err
	}

	tmp := out
	err = ErrorZeroResults
	for buf.Next() {
		msg := buf.PopMessage()

		switch msg.Type {
		case message.Data:
			msg.PopUint16() // number of data elements (always 1)

			if !q.flat() {
				val := reflect.New(tp).Elem()
				cdcs.out.Decode(msg, unsafe.Pointer(val.UnsafeAddr()))
				tmp = reflect.Append(tmp, val)
			} else {
				cdcs.out.Decode(msg, unsafe.Pointer(out.UnsafeAddr()))
			}
			err = nil
		case message.CommandComplete:
			msg.PopUint16() // header count (assume 0)
			msg.PopBytes()  // command status
		case message.ReadyForCommand:
			msg.PopUint16() // header count (assume 0)
			msg.PopUint8()  // transaction state
		case message.ErrorResponse:
			return decodeError(msg)
		default:
			e := c.fallThrough(msg)
			if e != nil {
				return e
			}
		}
		msg.Finish()
	}

	if !q.flat() {
		out.Set(tmp)
	}

	return err
}
