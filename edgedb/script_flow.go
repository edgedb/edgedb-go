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
	"github.com/edgedb/edgedb-go/edgedb/protocol/message"
)

func scriptFlow(ctx context.Context, conn net.Conn, query string) error {
	buf := []byte{message.ExecuteScript, 0, 0, 0, 0}
	protocol.PushUint16(&buf, 0) // no headers
	protocol.PushString(&buf, query)
	protocol.PutMsgLength(buf)

	err := writeAndRead(ctx, conn, &buf)
	if err != nil {
		return err
	}

	for len(buf) > 0 {
		msg := protocol.PopMessage(&buf)
		mType := protocol.PopUint8(&msg)

		switch mType {
		case message.CommandComplete:
		case message.ReadyForCommand:
		case message.ErrorResponse:
			return decodeError(&msg)
		default:
			return fmt.Errorf("unexpected message type: 0x%x", mType)
		}
	}

	return nil
}
