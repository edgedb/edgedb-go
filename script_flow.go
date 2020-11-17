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
	"net"

	"github.com/edgedb/edgedb-go/protocol/buff"
	"github.com/edgedb/edgedb-go/protocol/message"
)

func (c *Client) scriptFlow(
	ctx context.Context,
	conn net.Conn,
	query string,
) error {
	buf := buff.New(nil)
	buf.BeginMessage(message.ExecuteScript)
	buf.PushUint16(0) // no headers
	buf.PushString(query)
	buf.EndMessage()

	err := writeAndRead(ctx, conn, buf.Unwrap())
	if err != nil {
		return err
	}

	for buf.Next() {
		switch buf.MsgType {
		case message.CommandComplete:
			buf.PopUint16() // header count (assume 0)
			buf.PopBytes()  // command status
		case message.ReadyForCommand:
			buf.PopUint16() // header count (assume 0)
			buf.PopUint8()  // transaction state
		case message.ErrorResponse:
			return decodeError(buf)
		default:
			err = c.fallThrough(buf)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
