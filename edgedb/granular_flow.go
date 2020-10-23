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

	"github.com/edgedb/edgedb-go/edgedb/protocol"
	"github.com/edgedb/edgedb-go/edgedb/protocol/aspect"
	"github.com/edgedb/edgedb-go/edgedb/protocol/cardinality"
	"github.com/edgedb/edgedb-go/edgedb/protocol/codecs"
	"github.com/edgedb/edgedb-go/edgedb/protocol/message"
	"github.com/edgedb/edgedb-go/edgedb/types"
)

func (c *Client) queryCodecs(query string, ioFmt int) (queryCodecs, bool) {
	// todo this isn't thread safe
	key := queryCacheKey{query, ioFmt}
	codecs, ok := c.queryCache[key]
	return codecs, ok
}

func (c *Client) cacheQuery(
	query string,
	ioFmt int,
	in,
	out codecs.DecodeEncoder,
) {
	// todo this isn't thread safe
	key := queryCacheKey{query: query, format: ioFmt}
	c.queryCache[key] = queryCodecs{in: in, out: out}
}

func (c *Client) granularFlow(
	ctx context.Context,
	conn net.Conn,
	query string,
	ioFmt int,
	args []interface{},
) ([]interface{}, error) {
	codecs, ok := c.queryCodecs(query, ioFmt)
	if ok {
		return c.optimistic(ctx, conn, codecs, query, ioFmt, args)
	}

	return c.pesimistic(ctx, conn, query, ioFmt, args)
}

func (c *Client) pesimistic(
	ctx context.Context,
	conn net.Conn,
	query string,
	ioFmt int,
	args []interface{},
) ([]interface{}, error) {
	inID, outID, err := prepare(ctx, conn, query, ioFmt)
	if err != nil {
		return nil, err
	}

	in, inOK := c.codecCache[inID]
	out, outOK := c.codecCache[outID]
	if !inOK || !outOK {
		err := c.describe(ctx, conn)
		if err != nil {
			return nil, err
		}

		in = c.codecCache[inID]
		out = c.codecCache[outID]
		c.cacheQuery(query, ioFmt, in, out)
	}

	return execute(ctx, conn, in, out, args)
}

func prepare(
	ctx context.Context,
	conn net.Conn,
	query string,
	ioFmt int,
) (in types.UUID, out types.UUID, err error) {
	msg := []byte{message.Prepare, 0, 0, 0, 0}
	protocol.PushUint16(&msg, 0) // no headers
	protocol.PushUint8(&msg, uint8(ioFmt))
	protocol.PushUint8(&msg, cardinality.Many) // todo is this correct?
	protocol.PushBytes(&msg, []byte{})         // no statement name
	protocol.PushString(&msg, query)
	protocol.PutMsgLength(msg)

	pyld := msg
	pyld = append(pyld, message.Sync, 0, 0, 0, 4)

	rcv, err := writeAndRead(ctx, conn, pyld)
	if err != nil {
		return in, out, err
	}

	for len(rcv) > 4 {
		bts := protocol.PopMessage(&rcv)
		mType := protocol.PopUint8(&bts)

		switch mType {
		case message.PrepareComplete:
			protocol.PopUint32(&bts)     // message length
			protocol.PopUint16(&bts)     // number of headers, assume 0
			protocol.PopUint8(&bts)      // cardianlity
			in = protocol.PopUUID(&bts)  // input type id
			out = protocol.PopUUID(&bts) // output type id
		case message.ReadyForCommand:
		case message.ErrorResponse:
			return in, out, decodeError(&bts)
		default:
			panic(fmt.Sprintf("unexpected message type: 0x%x", mType))
		}
	}

	return in, out, nil
}

func (c *Client) describe(ctx context.Context, conn net.Conn) error {
	msg := []byte{message.DescribeStatement, 0, 0, 0, 0}
	protocol.PushUint16(&msg, 0) // no headers
	protocol.PushUint8(&msg, aspect.DataDescription)
	protocol.PushUint32(&msg, 0) // no statement name
	protocol.PutMsgLength(msg)

	pyld := msg
	pyld = append(pyld, message.Sync, 0, 0, 0, 4)

	rcv, err := writeAndRead(ctx, conn, pyld)
	if err != nil {
		return err
	}

	for len(rcv) > 4 {
		bts := protocol.PopMessage(&rcv)
		mType := protocol.PopUint8(&bts)

		switch mType {
		case message.CommandDataDescription:
			protocol.PopUint32(&bts)              // message length
			protocol.PopUint16(&bts)              // num headers is always 0
			protocol.PopUint8(&bts)               // cardianlity
			protocol.PopUUID(&bts)                // input descriptor ID
			descriptor := protocol.PopBytes(&bts) // input descriptor
			for k, v := range codecs.Pop(&descriptor) {
				c.codecCache[k] = v
			}

			protocol.PopUUID(&bts)               // output descriptor ID
			descriptor = protocol.PopBytes(&bts) // input descriptor
			for k, v := range codecs.Pop(&descriptor) {
				c.codecCache[k] = v
			}
		case message.ReadyForCommand:
		case message.ErrorResponse:
			return decodeError(&bts)
		default:
			panic(fmt.Sprintf("unexpected message type: 0x%x", mType))
		}
	}

	return nil
}

func execute(
	ctx context.Context,
	conn net.Conn,
	in codecs.DecodeEncoder,
	out codecs.DecodeEncoder,
	args []interface{},
) ([]interface{}, error) {
	msg := []byte{message.Execute, 0, 0, 0, 0}
	protocol.PushUint16(&msg, 0)       // no headers
	protocol.PushBytes(&msg, []byte{}) // no statement name
	in.Encode(&msg, args)
	protocol.PutMsgLength(msg)

	pyld := msg
	pyld = append(pyld, message.Sync, 0, 0, 0, 4)

	rcv, err := writeAndRead(ctx, conn, pyld)
	if err != nil {
		return nil, err
	}

	result := make(types.Set, 0)

	for len(rcv) > 0 {
		bts := protocol.PopMessage(&rcv)
		mType := protocol.PopUint8(&bts)

		switch mType {
		case message.Data:
			protocol.PopUint32(&bts) // message length
			protocol.PopUint16(&bts) // number of data elements (always 1)
			result = append(result, out.Decode(&bts))
		case message.CommandComplete:
		case message.ReadyForCommand:
		case message.ErrorResponse:
			return nil, decodeError(&bts)
		default:
			panic(fmt.Sprintf("unexpected message type: 0x%x", mType))
		}
	}

	return result, nil
}

func (c *Client) optimistic(
	ctx context.Context,
	conn net.Conn,
	codecs queryCodecs,
	query string,
	ioFmt int,
	args []interface{},
) ([]interface{}, error) {
	inID := codecs.in.ID()
	outID := codecs.out.ID()

	msg := []byte{message.OptimisticExecute, 0, 0, 0, 0}
	protocol.PushUint16(&msg, 0) // no headers
	protocol.PushUint8(&msg, uint8(ioFmt))
	protocol.PushUint8(&msg, cardinality.Many) // todo is this correct?
	protocol.PushString(&msg, query)
	msg = append(msg, inID[:]...)
	msg = append(msg, outID[:]...)
	codecs.in.Encode(&msg, args)
	protocol.PutMsgLength(msg)

	msg = append(msg, message.Sync, 0, 0, 0, 4)

	rcv, err := writeAndRead(ctx, conn, msg)
	if err != nil {
		return nil, err
	}

	result := make(types.Set, 0)
	for len(rcv) > 0 {
		bts := protocol.PopMessage(&rcv)
		mType := protocol.PopUint8(&bts)

		switch mType {
		case message.Data:
			protocol.PopUint32(&bts) // message length
			protocol.PopUint16(&bts) // number of data elements (always 1)
			result = append(result, codecs.out.Decode(&bts))
		case message.CommandComplete:
		case message.ReadyForCommand:
		case message.ErrorResponse:
			return nil, decodeError(&bts)
		default:
			panic(fmt.Sprintf("unexpected message type: 0x%x", mType))
		}
	}

	return result, nil
}
