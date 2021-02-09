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

func ignoreHeaders(r *buff.Reader) {
	n := int(r.PopUint16())

	for i := 0; i < n; i++ {
		r.PopUint16()
		r.PopBytes()
	}
}

func writeHeaders(w *buff.Writer, headers msgHeaders) {
	w.PushUint16(uint16(len(headers)))

	for key, val := range headers {
		w.PushUint16(key)
		w.PushUint32(uint32(len(val)))
		w.PushBytes(val)
	}
}

func (c *baseConn) scriptFlow(r *buff.Reader, q sfQuery) error {
	w := buff.NewWriter(c.writeMemory[:0])
	w.BeginMessage(message.ExecuteScript)
	writeHeaders(w, q.headers)
	w.PushString(q.cmd)
	w.EndMessage()

	if e := w.Send(c.conn); e != nil {
		return &clientConnectionError{err: e}
	}

	var err error
	done := buff.NewSignal()

	for r.Next(done.Chan) {
		switch r.MsgType {
		case message.CommandComplete:
			ignoreHeaders(r)
			r.PopBytes() // command status
		case message.ReadyForCommand:
			ignoreHeaders(r)
			r.Discard(1) // transaction state
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
		return &clientConnectionError{err: r.Err}
	}

	return err
}
