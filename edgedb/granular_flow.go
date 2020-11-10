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

	"github.com/edgedb/edgedb-go/edgedb/protocol"
	"github.com/edgedb/edgedb-go/edgedb/protocol/aspect"
	"github.com/edgedb/edgedb-go/edgedb/protocol/codecs"
	"github.com/edgedb/edgedb-go/edgedb/protocol/format"
	"github.com/edgedb/edgedb-go/edgedb/protocol/message"
)

func (c *Conn) granularFlow(
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
			d := desc.([]byte)
			in, err = codecs.BuildCodec(&d)
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
			d := desc.([]byte)
			cOut, err = codecs.BuildTypedCodec(&d, tp)
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

func (c *Conn) pesimistic(
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
	cdcs.in, err = codecs.BuildCodec(&descs.in)
	if err != nil {
		return err
	}

	if q.fmt == format.JSON {
		cdcs.out = codecs.JSONBytes
	} else {
		cdcs.out, err = codecs.BuildTypedCodec(&descs.out, tp)
		if err != nil {
			return err
		}
	}

	c.inCodecCache.Put(ids.in, cdcs.in)
	c.outCodecCache.Put(codecKey{ID: ids.out, Type: tp}, cdcs.out)
	return c.execute(ctx, out, q, tp, cdcs)
}

func (c *Conn) prepare(ctx context.Context, q query) (ids idPair, err error) {
	buf := []byte{message.Prepare, 0, 0, 0, 0}
	protocol.PushUint16(&buf, 0) // no headers
	protocol.PushUint8(&buf, q.fmt)
	protocol.PushUint8(&buf, q.expCard)
	protocol.PushBytes(&buf, []byte{}) // no statement name
	protocol.PushString(&buf, q.cmd)
	protocol.PutMsgLength(buf)

	buf = append(buf, message.Sync, 0, 0, 0, 4)

	err = c.writeAndRead(ctx, &buf)
	if err != nil {
		return ids, err
	}

	for len(buf) > 4 {
		msg := protocol.PopMessage(&buf)
		mType := protocol.PopUint8(&msg)

		switch mType {
		case message.PrepareComplete:
			protocol.PopUint32(&msg) // message length
			protocol.PopUint16(&msg) // number of headers, assume 0

			// todo assert cardinality matches query
			protocol.PopUint8(&msg) // cardianlity

			ids = idPair{
				in:  protocol.PopUUID(&msg),
				out: protocol.PopUUID(&msg),
			}
		case message.ReadyForCommand:
		case message.ErrorResponse:
			return ids, decodeError(&msg)
		default:
			panic(fmt.Sprintf("unexpected message type: 0x%x", mType))
		}
	}

	return ids, nil
}

func (c *Conn) describe(ctx context.Context) (descs descPair, err error) {
	buf := []byte{message.DescribeStatement, 0, 0, 0, 0}
	protocol.PushUint16(&buf, 0) // no headers
	protocol.PushUint8(&buf, aspect.DataDescription)
	protocol.PushUint32(&buf, 0) // no statement name
	protocol.PutMsgLength(buf)

	buf = append(buf, message.Sync, 0, 0, 0, 4)

	err = c.writeAndRead(ctx, &buf)
	if err != nil {
		return descs, err
	}

	for len(buf) > 4 {
		msg := protocol.PopMessage(&buf)
		mType := protocol.PopUint8(&msg)

		switch mType {
		case message.CommandDataDescription:
			protocol.PopUint32(&msg) // message length
			protocol.PopUint16(&msg) // num headers is always 0
			protocol.PopUint8(&msg)  // cardianlity

			// input descriptor
			protocol.PopUUID(&msg)
			descs.in = append(descs.in, protocol.PopBytes(&msg)...)

			// output descriptor
			protocol.PopUUID(&msg)
			descs.out = append(descs.out, protocol.PopBytes(&msg)...)
		case message.ReadyForCommand:
		case message.ErrorResponse:
			return descs, decodeError(&msg)
		default:
			panic(fmt.Sprintf("unexpected message type: 0x%x", mType))
		}
	}

	return descs, nil
}

func (c *Conn) execute(
	ctx context.Context,
	out reflect.Value,
	q query,
	tp reflect.Type,
	cdcs codecPair,
) error {
	buf := []byte{message.Execute, 0, 0, 0, 0}
	protocol.PushUint16(&buf, 0)       // no headers
	protocol.PushBytes(&buf, []byte{}) // no statement name
	cdcs.in.Encode(&buf, q.args)
	protocol.PutMsgLength(buf)

	buf = append(buf, message.Sync, 0, 0, 0, 4)

	err := c.writeAndRead(ctx, &buf)
	if err != nil {
		return err
	}

	o := out
	if !q.flat() {
		out.SetLen(0)
	}

	err = ErrorZeroResults
	for len(buf) > 0 {
		msg := protocol.PopMessage(&buf)
		mType := protocol.PopUint8(&msg)

		switch mType {
		case message.Data:
			protocol.PopUint32(&msg) // message length
			protocol.PopUint16(&msg) // number of data elements (always 1)

			if !q.flat() {
				val := reflect.New(tp).Elem()
				cdcs.out.Decode(&msg, val)
				o = reflect.Append(o, val)
			} else {
				cdcs.out.Decode(&msg, out)
			}
			err = nil
		case message.CommandComplete:
		case message.ReadyForCommand:
		case message.ErrorResponse:
			return decodeError(&msg)
		default:
			panic(fmt.Sprintf("unexpected message type: 0x%x", mType))
		}
	}

	if !q.flat() {
		out.Set(o)
	}

	return err
}

func (c *Conn) optimistic(
	ctx context.Context,
	out reflect.Value,
	q query,
	tp reflect.Type,
	cdcs codecPair,
) error {
	inID := cdcs.in.ID()
	outID := cdcs.out.ID()

	buf := c.buffer[:0]
	buf = append(buf,
		message.OptimisticExecute,
		0, 0, 0, 0, // message length slot, to be filled in later
		0, 0, // no headers
		q.fmt,
		q.expCard,
	)

	protocol.PushString(&buf, q.cmd)
	buf = append(buf, inID[:]...)
	buf = append(buf, outID[:]...)
	cdcs.in.Encode(&buf, q.args)
	protocol.PutMsgLength(buf)

	buf = append(buf, message.Sync, 0, 0, 0, 4)

	err := c.writeAndRead(ctx, &buf)
	if err != nil {
		return err
	}

	o := out
	if !q.flat() {
		out.SetLen(0)
	}

	err = ErrorZeroResults
	for len(buf) > 0 {
		msg := protocol.PopMessage(&buf)
		mType := protocol.PopUint8(&msg)

		switch mType {
		case message.Data:
			// skip the following fields
			// message length
			// number of data elements (always 1)
			msg = msg[6:]

			if !q.flat() {
				val := reflect.New(tp).Elem()
				cdcs.out.Decode(&msg, val)
				o = reflect.Append(o, val)
			} else {
				cdcs.out.Decode(&msg, out)
			}
			err = nil
		case message.CommandComplete:
		case message.ReadyForCommand:
		case message.ErrorResponse:
			return decodeError(&msg)
		default:
			panic(fmt.Sprintf("unexpected message type: 0x%x", mType))
		}
	}

	if !q.flat() {
		out.Set(o)
	}

	return err
}
