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
	"crypto/tls"
	"fmt"

	"github.com/edgedb/edgedb-go/internal"
	"github.com/edgedb/edgedb-go/internal/buff"
	"github.com/edgedb/edgedb-go/internal/message"
	"github.com/xdg/scram"
)

var (
	protocolVersionMin  = internal.ProtocolVersion{Major: 0, Minor: 9}
	protocolVersionMax  = internal.ProtocolVersion{Major: 0, Minor: 11}
	protocolVersion0p10 = internal.ProtocolVersion{Major: 0, Minor: 10}
	protocolVersion0p11 = internal.ProtocolVersion{Major: 0, Minor: 11}
)

func (c *baseConn) connect(r *buff.Reader, cfg *connConfig) error {
	w := buff.NewWriter(c.writeMemory[:0])
	w.BeginMessage(message.ClientHandshake)
	w.PushUint16(protocolVersionMax.Major)
	w.PushUint16(protocolVersionMax.Minor)
	w.PushUint16(2) // number of parameters
	w.PushString("database")
	w.PushString(cfg.database)
	w.PushString("user")
	w.PushString(cfg.user)
	w.PushUint16(0) // no extensions
	w.EndMessage()

	c.protocolVersion = protocolVersionMax

	if err := w.Send(c.conn); err != nil {
		return &clientConnectionError{err: err}
	}

	var err error
	done := buff.NewSignal()

	for r.Next(done.Chan) {
		switch r.MsgType {
		case message.ServerHandshake:
			// The client _MUST_ close the connection
			// if the protocol version can't be supported.
			// https://edgedb.com/docs/internals/protocol/overview
			protocolVersion := internal.ProtocolVersion{
				Major: r.PopUint16(),
				Minor: r.PopUint16(),
			}

			if protocolVersion.LT(protocolVersionMin) ||
				protocolVersion.GT(protocolVersionMax) {
				_ = c.conn.Close()
				msg := fmt.Sprintf(
					"unsupported protocol version: %v.%v",
					protocolVersion.Major,
					protocolVersion.Minor,
				)
				return &unsupportedProtocolVersionError{msg: msg}
			}

			c.protocolVersion = protocolVersion

			n := r.PopUint16()
			for i := uint16(0); i < n; i++ {
				r.PopBytes() // extension name
				ignoreHeaders(r)
			}
		case message.ServerKeyData:
			r.DiscardMessage() // key data
		case message.ReadyForCommand:
			ignoreHeaders(r)
			r.Discard(1) // transaction state
			done.Signal()
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

			done.Signal()
		case message.ErrorResponse:
			err = wrapAll(err, decodeError(r, ""))
			done.Signal()
		default:
			if e := c.fallThrough(r); e != nil {
				// the connection will not be usable after this x_x
				return e
			}
		}
	}

	_, isTLS := c.conn.(*tls.Conn)
	if !isTLS && c.protocolVersion.GTE(protocolVersion0p11) {
		_ = c.close()
		return &clientConnectionError{msg: fmt.Sprintf(
			"server claims to use protocol version %v.%v without using TLS",
			c.protocolVersion.Major, c.protocolVersion.Minor)}
	}
	if c.protocolVersion.GTE(protocolVersion0p10) {
		c.explicitIDs = true
	}

	if r.Err != nil {
		return &clientConnectionError{err: r.Err}
	}

	return err
}

func (c *baseConn) authenticate(r *buff.Reader, cfg *connConfig) error {
	client, err := scram.SHA256.NewClient(cfg.user, cfg.password, "")
	if err != nil {
		return &authenticationError{msg: err.Error()}
	}

	conv := client.NewConversation()
	scramMsg, err := conv.Step("")
	if err != nil {
		return &authenticationError{msg: err.Error()}
	}

	w := buff.NewWriter(c.writeMemory[:0])
	w.BeginMessage(message.AuthenticationSASLInitialResponse)
	w.PushString("SCRAM-SHA-256")
	w.PushString(scramMsg)
	w.EndMessage()

	if e := w.Send(c.conn); e != nil {
		return &clientConnectionError{err: e}
	}

	done := buff.NewSignal()

	for r.Next(done.Chan) {
		switch r.MsgType {
		case message.Authentication:
			authStatus := r.PopUint32()
			if authStatus != 0xb {
				// the connection will not be usable after this x_x
				return &authenticationError{msg: fmt.Sprintf(
					"unexpected authentication status: 0x%x", authStatus,
				)}
			}

			scramRcv := r.PopString()
			scramMsg, err = conv.Step(scramRcv)
			if err != nil {
				// the connection will not be usable after this x_x
				return &authenticationError{msg: err.Error()}
			}

			done.Signal()
		case message.ErrorResponse:
			err = decodeError(r, "")
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

	if err != nil {
		return err
	}

	w = buff.NewWriter(c.writeMemory[:0])
	w.BeginMessage(message.AuthenticationSASLResponse)
	w.PushString(scramMsg)
	w.EndMessage()

	if e := w.Send(c.conn); e != nil {
		return &clientConnectionError{err: e}
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
					return &authenticationError{msg: e.Error()}
				}
			default:
				// the connection will not be usable after this x_x
				return &authenticationError{msg: fmt.Sprintf(
					"unexpected authentication status: 0x%x", authStatus,
				)}
			}
		case message.ServerKeyData:
			r.DiscardMessage() // key data
		case message.ReadyForCommand:
			ignoreHeaders(r)
			r.Discard(1) // transaction state
			done.Signal()
		case message.ErrorResponse:
			err = wrapAll(decodeError(r, ""))
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

func (c *baseConn) terminate() error {
	w := buff.NewWriter(c.writeMemory[:0])
	w.BeginMessage(message.Terminate)
	w.EndMessage()

	if e := w.Send(c.conn); e != nil {
		return &clientConnectionError{err: e}
	}

	return nil
}
