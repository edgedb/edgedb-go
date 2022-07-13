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
	"github.com/edgedb/edgedb-go/internal/header"
	"github.com/edgedb/edgedb-go/internal/message"
)

func ignoreHeaders(r *buff.Reader) {
	n := int(r.PopUint16())

	for i := 0; i < n; i++ {
		r.PopUint16()
		r.PopBytes()
	}
}

func decodeHeaders(r *buff.Reader) header.Header {
	n := int(r.PopUint16())

	headers := make(header.Header, n)
	for i := 0; i < n; i++ {
		key := r.PopUint16()
		val := r.PopBytes()
		headers[key] = make([]byte, len(val))
		copy(headers[key], val)
	}

	return headers
}

func discardHeaders(r *buff.Reader) {
	n := int(r.PopUint16())

	for i := 0; i < n; i++ {
		r.PopUint16()
		r.PopBytes()
	}
}

func writeHeaders(w *buff.Writer, headers header.Header) {
	w.PushUint16(uint16(len(headers)))

	for key, val := range headers {
		w.PushUint16(key)
		w.PushUint32(uint32(len(val)))
		w.PushBytes(val)
	}
}

func (c *protocolConnection) execScriptFlow(r *buff.Reader, q *query) error {
	if len(q.state) != 0 {
		return errStateNotSupported
	}

	w := buff.NewWriter(c.writeMemory[:0])
	w.BeginMessage(message.ExecuteScript)
	writeHeaders(w, q.headers0pX())
	w.PushString(q.cmd)
	w.EndMessage()

	if e := c.soc.WriteAll(w.Unwrap()); e != nil {
		return e
	}

	var err error
	done := buff.NewSignal()

	for r.Next(done.Chan) {
		switch r.MsgType {
		case message.CommandComplete:
			decodeCommandCompleteMsg0pX(r)
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
