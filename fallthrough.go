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
	"fmt"
	"log"

	"github.com/edgedb/edgedb-go/internal/buff"
	"github.com/edgedb/edgedb-go/internal/message"
)

var logMsgSeverityLookup map[uint8]string = map[uint8]string{
	0x14: "DEBUG",
	0x28: "INFO",
	0x3c: "NOTICE",
	0x50: "WARNING",
}

func (c *baseConn) fallThrough(r *buff.Reader) error {
	switch r.MsgType {
	case message.ParameterStatus:
		name := r.PopString()
		value := r.PopString()
		c.serverSettings[name] = value
	case message.LogMessage:
		severity := logMsgSeverityLookup[r.PopUint8()]
		code := r.PopUint32()
		message := r.PopString()
		r.Discard(2) // number of headers, assume 0
		log.Println("SERVER MESSAGE", severity, code, message)
	default:
		msg := fmt.Sprintf("unexpected message type: 0x%x", r.MsgType)
		return newError(msg)
	}

	return nil
}
