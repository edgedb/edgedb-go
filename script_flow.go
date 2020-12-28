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
	"github.com/edgedb/edgedb-go/internal/buff"
	"github.com/edgedb/edgedb-go/internal/message"
)

func (c *baseConn) scriptFlow(r *buff.Reader, query string) error {
	c.writer.BeginMessage(message.ExecuteScript)
	c.writer.PushUint16(0) // no headers
	c.writer.PushString(query)
	c.writer.EndMessage()

	if e := c.writer.Send(c.conn); e != nil {
		return wrapError(e)
	}

	var err error
	done := buff.NewSignal()

	for r.Next(done.Chan) {
		switch r.MsgType {
		case message.CommandComplete:
			r.Discard(2) // header count (assume 0)
			r.PopBytes() // command status
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
				return e
			}
		}
	}

	if r.Err != nil {
		return wrapError(r.Err)
	}

	return err
}
