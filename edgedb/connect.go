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
	"net"

	"github.com/edgedb/edgedb-go/edgedb/protocol"
	"github.com/edgedb/edgedb-go/edgedb/protocol/message"
	"github.com/xdg/scram"
)

func connect(conn net.Conn, opts *Options) (err error) {
	msg := []byte{message.ClientHandshake, 0, 0, 0, 0}
	protocol.PushUint16(&msg, 0) // major version
	protocol.PushUint16(&msg, 8) // minor version
	protocol.PushUint16(&msg, 2) // number of parameters
	protocol.PushString(&msg, "database")
	protocol.PushString(&msg, opts.Database)
	protocol.PushString(&msg, "user")
	protocol.PushString(&msg, opts.User)
	protocol.PushUint16(&msg, 0) // no extensions
	protocol.PutMsgLength(msg)

	rcv, err := writeAndRead(conn, msg)
	if err != nil {
		return err
	}

	for len(rcv) > 0 {
		bts := protocol.PopMessage(&rcv)
		mType := protocol.PopUint8(&bts)

		switch mType {
		case message.ServerHandshake:
			// The client _MUST_ close the connection
			// if the protocol version can't be supported.
			// https://edgedb.com/docs/internals/protocol/overview
			protocol.PopUint32(&bts) // message length
			major := protocol.PopUint16(&bts)
			minor := protocol.PopUint16(&bts)

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
			return decodeError(&bts)
		case message.Authentication:
			protocol.PopUint32(&bts) // message length
			authStatus := protocol.PopUint32(&bts)

			if authStatus == 0 {
				continue
			}

			err := authenticate(conn, opts)
			if err != nil {
				return err
			}
		default:
			panic(fmt.Sprintf("unexpected message type: 0x%x", mType))
		}
	}
	return nil
}

func authenticate(conn net.Conn, opts *Options) (err error) {
	client, err := scram.SHA256.NewClient(opts.User, opts.Password, "")
	if err != nil {
		panic(err)
	}

	conv := client.NewConversation()
	scramMsg, err := conv.Step("")
	if err != nil {
		panic(err)
	}

	msg := []byte{message.AuthenticationSASLInitialResponse, 0, 0, 0, 0}
	protocol.PushString(&msg, "SCRAM-SHA-256")
	protocol.PushString(&msg, scramMsg)
	protocol.PutMsgLength(msg)

	rcv, err := writeAndRead(conn, msg)
	if err != nil {
		return err
	}

	mType := protocol.PopUint8(&rcv)

	switch mType {
	case message.Authentication:
		protocol.PopUint32(&rcv) // message length
		authStatus := protocol.PopUint32(&rcv)
		if authStatus != 0xb {
			panic(fmt.Sprintf(
				"unexpected authentication status: 0x%x",
				authStatus,
			))
		}

		scramRcv := protocol.PopString(&rcv)
		scramMsg, err = conv.Step(scramRcv)
		if err != nil {
			panic(err)
		}
	case message.ErrorResponse:
		return decodeError(&rcv)
	default:
		panic(fmt.Sprintf("unexpected message type: 0x%x", mType))
	}

	msg = []byte{message.AuthenticationSASLResponse, 0, 0, 0, 0}
	protocol.PushString(&msg, scramMsg)
	protocol.PutMsgLength(msg)

	rcv, err = writeAndRead(conn, msg)
	if err != nil {
		return err
	}

	for len(rcv) > 0 {
		bts := protocol.PopMessage(&rcv)
		mType := protocol.PopUint8(&bts)

		switch mType {
		case message.Authentication:
			protocol.PopUint32(&bts) // message length
			authStatus := protocol.PopUint32(&bts)

			switch authStatus {
			case 0:
				continue
			case 0xc:
				scramRcv := protocol.PopString(&bts)
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
			return decodeError(&bts)
		default:
			panic(fmt.Sprintf("unexpected message type: 0x%x", mType))
		}
	}

	return nil
}
