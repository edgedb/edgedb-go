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
	"github.com/xdg/scram"
)

func connect(ctx context.Context, conn net.Conn, opts *Options) (err error) {
	buf := []byte{message.ClientHandshake, 0, 0, 0, 0}
	protocol.PushUint16(&buf, 0) // major version
	protocol.PushUint16(&buf, 8) // minor version
	protocol.PushUint16(&buf, 2) // number of parameters
	protocol.PushString(&buf, "database")
	protocol.PushString(&buf, opts.Database)
	protocol.PushString(&buf, "user")
	protocol.PushString(&buf, opts.User)
	protocol.PushUint16(&buf, 0) // no extensions
	protocol.PutMsgLength(buf)

	err = writeAndRead(ctx, conn, &buf)
	if err != nil {
		return err
	}

	for len(buf) > 0 {
		msg := protocol.PopMessage(&buf)
		mType := protocol.PopUint8(&msg)

		switch mType {
		case message.ServerHandshake:
			// The client _MUST_ close the connection
			// if the protocol version can't be supported.
			// https://edgedb.com/docs/internals/protocol/overview
			protocol.PopUint32(&msg) // message length
			major := protocol.PopUint16(&msg)
			minor := protocol.PopUint16(&msg)

			if major != 0 || minor != 8 {
				err = conn.Close()
				if err != nil {
					return err
				}

				err = fmt.Errorf(
					"unsupported protocol version: %v.%v",
					major,
					minor,
				)
				return err
			}
		case message.ServerKeyData:
		case message.ReadyForCommand:
			return nil
		case message.ErrorResponse:
			return decodeError(&msg)
		case message.Authentication:
			protocol.PopUint32(&msg) // message length
			authStatus := protocol.PopUint32(&msg)

			if authStatus == 0 {
				continue
			}

			err := authenticate(ctx, conn, opts)
			if err != nil {
				return err
			}
		default:
			panic(fmt.Sprintf("unexpected message type: 0x%x", mType))
		}
	}
	return nil
}

func authenticate(
	ctx context.Context,
	conn net.Conn,
	opts *Options,
) (err error) {
	client, err := scram.SHA256.NewClient(opts.User, opts.Password, "")
	if err != nil {
		panic(err)
	}

	conv := client.NewConversation()
	scramMsg, err := conv.Step("")
	if err != nil {
		panic(err)
	}

	buf := []byte{message.AuthenticationSASLInitialResponse, 0, 0, 0, 0}
	protocol.PushString(&buf, "SCRAM-SHA-256")
	protocol.PushString(&buf, scramMsg)
	protocol.PutMsgLength(buf)

	err = writeAndRead(ctx, conn, &buf)
	if err != nil {
		return err
	}

	mType := protocol.PopUint8(&buf)

	switch mType {
	case message.Authentication:
		protocol.PopUint32(&buf) // message length
		authStatus := protocol.PopUint32(&buf)
		if authStatus != 0xb {
			panic(fmt.Sprintf(
				"unexpected authentication status: 0x%x",
				authStatus,
			))
		}

		scramRcv := protocol.PopString(&buf)
		scramMsg, err = conv.Step(scramRcv)
		if err != nil {
			panic(err)
		}
	case message.ErrorResponse:
		return decodeError(&buf)
	default:
		panic(fmt.Sprintf("unexpected message type: 0x%x", mType))
	}

	buf = []byte{message.AuthenticationSASLResponse, 0, 0, 0, 0}
	protocol.PushString(&buf, scramMsg)
	protocol.PutMsgLength(buf)

	err = writeAndRead(ctx, conn, &buf)
	if err != nil {
		return err
	}

	for len(buf) > 0 {
		msg := protocol.PopMessage(&buf)
		mType := protocol.PopUint8(&msg)

		switch mType {
		case message.Authentication:
			protocol.PopUint32(&msg) // message length
			authStatus := protocol.PopUint32(&msg)

			switch authStatus {
			case 0:
				continue
			case 0xc:
				scramRcv := protocol.PopString(&msg)
				_, err = conv.Step(scramRcv)
				if err != nil {
					panic(err)
				}
			default:
				panic(fmt.Sprintf(
					"unexpected authentication status: 0x%x",
					authStatus,
				))
			}
		case message.ServerKeyData:
		case message.ReadyForCommand:
			return nil
		case message.ErrorResponse:
			return decodeError(&msg)
		default:
			panic(fmt.Sprintf("unexpected message type: 0x%x", mType))
		}
	}

	return nil
}
