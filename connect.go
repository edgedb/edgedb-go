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
	"sync"

	"github.com/edgedb/edgedb-go/internal/buff"
	"github.com/edgedb/edgedb-go/internal/message"
	"github.com/xdg/scram"
)

const (
	protocolVersionMajor uint16 = 0
	protocolVersionMinor uint16 = 8
)

func (c *baseConn) connect(r *buff.Reader, cfg *connConfig) error {
	c.writer.BeginMessage(message.ClientHandshake)
	c.writer.PushUint16(0) // major version
	c.writer.PushUint16(8) // minor version
	c.writer.PushUint16(2) // number of parameters
	c.writer.PushString("database")
	c.writer.PushString(cfg.database)
	c.writer.PushString("user")
	c.writer.PushString(cfg.user)
	c.writer.PushUint16(0) // no extensions
	c.writer.EndMessage()

	if err := c.writer.Send(c.conn); err != nil {
		return wrapError(err)
	}

	var (
		err  error
		once sync.Once
	)

	doneReadingSignal := make(chan struct{}, 1)
	done := func() { doneReadingSignal <- struct{}{} }

	for r.Next(doneReadingSignal) {
		switch r.MsgType {
		case message.ServerHandshake:
			// The client _MUST_ close the connection
			// if the protocol version can't be supported.
			// https://edgedb.com/docs/internals/protocol/overview
			major := r.PopUint16()
			minor := r.PopUint16()

			if major != protocolVersionMajor || minor != protocolVersionMinor {
				_ = c.conn.Close()
				msg := fmt.Sprintf(
					"unsupported protocol version: %v.%v", major, minor)
				return newErrorFromCode(
					unsupportedProtocolVersionErrorCode, msg)
			}
		case message.ServerKeyData:
			r.DiscardMessage() // key data
		case message.ReadyForCommand:
			// header count (assume 0)
			// transaction state
			r.Discard(3)

			once.Do(done)
		case message.Authentication:
			if r.PopUint32() == 0 { // auth status
				continue
			}

			// skip supported SASL methods
			n := int(r.PopUint32()) // method count
			for i := 0; i < n; i++ {
				r.PopBytes()
			}

			if e := c.authenticate(r, cfg); e != nil {
				return e
			}

			once.Do(done)
		case message.ErrorResponse:
			err = wrapAll(err, decodeError(r))
			once.Do(done)
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

func (c *baseConn) authenticate(r *buff.Reader, cfg *connConfig) error {
	client, err := scram.SHA256.NewClient(cfg.user, cfg.password, "")
	if err != nil {
		return newErrorFromCode(authenticationErrorCode, err.Error())
	}

	conv := client.NewConversation()
	scramMsg, err := conv.Step("")
	if err != nil {
		return newErrorFromCode(authenticationErrorCode, err.Error())
	}

	c.writer.BeginMessage(message.AuthenticationSASLInitialResponse)
	c.writer.PushString("SCRAM-SHA-256")
	c.writer.PushString(scramMsg)
	c.writer.EndMessage()

	if e := c.writer.Send(c.conn); e != nil {
		return wrapError(e)
	}

	done := buff.NewSignal()

	for r.Next(done.Chan) {
		switch r.MsgType {
		case message.Authentication:
			authStatus := r.PopUint32()
			if authStatus != 0xb {
				// the connection will not be usable after this x_x
				return newErrorFromCode(authenticationErrorCode, fmt.Sprintf(
					"unexpected authentication status: 0x%x", authStatus,
				))
			}

			scramRcv := r.PopString()
			scramMsg, err = conv.Step(scramRcv)
			if err != nil {
				// the connection will not be usable after this x_x
				return newErrorFromCode(authenticationErrorCode, err.Error())
			}

			done.Signal()
		case message.ErrorResponse:
			err = decodeError(r)
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

	if err != nil {
		return err
	}

	c.writer.BeginMessage(message.AuthenticationSASLResponse)
	c.writer.PushString(scramMsg)
	c.writer.EndMessage()

	if e := c.writer.Send(c.conn); e != nil {
		return wrapError(e)
	}

	done = buff.NewSignal()

	for r.Next(done.Chan) {
		switch r.MsgType {
		case message.Authentication:
			authStatus := r.PopUint32()
			switch authStatus {
			case 0:
			case 0xc:
				scramRcv := r.PopString()
				_, e := conv.Step(scramRcv)
				if e != nil {
					// the connection will not be usable after this x_x
					return newErrorFromCode(authenticationErrorCode, e.Error())
				}
			default:
				// the connection will not be usable after this x_x
				return newErrorFromCode(authenticationErrorCode, fmt.Sprintf(
					"unexpected authentication status: 0x%x", authStatus,
				))
			}
		case message.ServerKeyData:
			r.DiscardMessage() // key data
		case message.ReadyForCommand:
			// header count (assume 0)
			// transaction state
			r.Discard(3)
			done.Signal()
		case message.ErrorResponse:
			err = wrapAll(decodeError(r))
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

func (c *baseConn) terminate() error {
	c.writer.BeginMessage(message.Terminate)
	c.writer.EndMessage()

	if e := c.writer.Send(c.conn); e != nil {
		return wrapError(e)
	}

	return nil
}
