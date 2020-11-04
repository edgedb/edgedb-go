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

	"github.com/edgedb/edgedb-go/edgedb/protocol"
	"github.com/edgedb/edgedb-go/edgedb/protocol/aspect"
	"github.com/edgedb/edgedb-go/edgedb/protocol/codecs"
	"github.com/edgedb/edgedb-go/edgedb/protocol/format"
	"github.com/edgedb/edgedb-go/edgedb/protocol/message"
	"github.com/edgedb/edgedb-go/edgedb/types"
)

func (c *Client) queryCodecs(q query, t reflect.Type) (queryCodecs, bool) {
	// todo this isn't thread safe
	key := queryCacheKey{q.cmd, q.fmt, q.expCard, t}
	codecs, ok := c.codecs[key]
	return codecs, ok
}

func (c *Client) cacheQueryCodecs(
	q query,
	t reflect.Type,
	in,
	out codecs.Codec,
) {
	if in == nil {
		panic("in codec is nil")
	}

	if out == nil {
		panic("out codec is nil")
	}

	// todo this isn't thread safe
	key := queryCacheKey{q.cmd, q.fmt, q.expCard, t}
	c.codecs[key] = queryCodecs{in: in, out: out}
}

func (c *Client) granularFlow(
	ctx context.Context,
	conn net.Conn,
	out reflect.Value,
	q query,
) error {
	if _, ok := c.queryCodecs(q, out.Type()); ok {
		return c.optimistic(ctx, conn, out, q)
	}

	return c.pesimistic(ctx, conn, out, q)
}

func (c *Client) pesimistic(
	ctx context.Context,
	conn net.Conn,
	out reflect.Value,
	q query,
) error {
	outType := out.Type()
	if !q.flat() {
		outType = outType.Elem()
	}

	inID, outID, err := prepare(ctx, conn, q)
	if err != nil {
		return err
	}

	dIn, inOK := c.descriptors[inID]
	dOut, outOK := c.descriptors[outID]
	if !inOK || !outOK {
		err = c.describe(ctx, conn)
		if err != nil {
			return err
		}

		dIn = c.descriptors[inID]
		dOut = c.descriptors[outID]
	}

	cIn, err := codecs.BuildCodec(&dIn)
	if err != nil {
		return err
	}

	var cOut codecs.Codec
	if q.fmt == format.JSON {
		cOut = codecs.JSONBytes
	} else {
		cOut, err = codecs.BuildTypedCodec(&dOut, outType)
		if err != nil {
			return err
		}
	}

	c.cacheQueryCodecs(q, outType, cIn, cOut)
	return c.execute(ctx, conn, out, q)
}

func prepare(
	ctx context.Context,
	conn net.Conn,
	q query,
) (in types.UUID, out types.UUID, err error) {
	buf := []byte{message.Prepare, 0, 0, 0, 0}
	protocol.PushUint16(&buf, 0) // no headers
	protocol.PushUint8(&buf, q.fmt)
	protocol.PushUint8(&buf, q.expCard)
	protocol.PushBytes(&buf, []byte{}) // no statement name
	protocol.PushString(&buf, q.cmd)
	protocol.PutMsgLength(buf)

	buf = append(buf, message.Sync, 0, 0, 0, 4)

	err = writeAndRead(ctx, conn, &buf)
	if err != nil {
		return in, out, err
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

			in = protocol.PopUUID(&msg)  // input type id
			out = protocol.PopUUID(&msg) // output type id
		case message.ReadyForCommand:
		case message.ErrorResponse:
			return in, out, decodeError(&msg)
		default:
			panic(fmt.Sprintf("unexpected message type: 0x%x", mType))
		}
	}

	return in, out, nil
}

func (c *Client) describe(ctx context.Context, conn net.Conn) error {
	buf := []byte{message.DescribeStatement, 0, 0, 0, 0}
	protocol.PushUint16(&buf, 0) // no headers
	protocol.PushUint8(&buf, aspect.DataDescription)
	protocol.PushUint32(&buf, 0) // no statement name
	protocol.PutMsgLength(buf)

	buf = append(buf, message.Sync, 0, 0, 0, 4)

	err := writeAndRead(ctx, conn, &buf)
	if err != nil {
		return err
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
			id := protocol.PopUUID(&msg)
			d := protocol.PopBytes(&msg)
			o := make([]byte, len(d))
			copy(o, d)
			c.descriptors[id] = o

			// output descriptor
			id = protocol.PopUUID(&msg)
			d = protocol.PopBytes(&msg)
			o = make([]byte, len(d))
			copy(o, d)
			c.descriptors[id] = o
		case message.ReadyForCommand:
		case message.ErrorResponse:
			return decodeError(&msg)
		default:
			panic(fmt.Sprintf("unexpected message type: 0x%x", mType))
		}
	}

	return nil
}

func (c *Client) execute(
	ctx context.Context,
	conn net.Conn,
	out reflect.Value,
	q query,
) error {
	outType := out.Type()
	if !q.flat() {
		outType = outType.Elem()
	}

	cdcs, _ := c.queryCodecs(q, outType)
	buf := []byte{message.Execute, 0, 0, 0, 0}
	protocol.PushUint16(&buf, 0)       // no headers
	protocol.PushBytes(&buf, []byte{}) // no statement name
	cdcs.in.Encode(&buf, q.args)
	protocol.PutMsgLength(buf)

	buf = append(buf, message.Sync, 0, 0, 0, 4)

	err := writeAndRead(ctx, conn, &buf)
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
				val := reflect.New(outType).Elem()
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

func (c *Client) optimistic(
	ctx context.Context,
	conn net.Conn,
	out reflect.Value,
	q query,
) error {
	outType := out.Type()
	if !q.flat() {
		outType = outType.Elem()
	}

	cdcs, _ := c.queryCodecs(q, out.Type())
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

	err := writeAndRead(ctx, conn, &buf)
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
				val := reflect.New(outType).Elem()
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
