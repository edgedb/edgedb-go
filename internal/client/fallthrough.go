// This source file is part of the EdgeDB open source project.
//
// Copyright EdgeDB Inc. and the EdgeDB authors.
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

package gel

import (
	"fmt"
	"log"
	"reflect"
	"strconv"
	"unsafe"

	"github.com/geldata/gel-go/internal/buff"
	"github.com/geldata/gel-go/internal/codecs"
	"github.com/geldata/gel-go/internal/descriptor"
)

var logMsgSeverityLookup = map[uint8]string{
	0x14: "DEBUG",
	0x28: "INFO",
	0x3c: "NOTICE",
	0x50: "WARNING",
}

func (c *protocolConnection) fallThrough(r *buff.Reader) error {
	if c.protocolVersion.GTE(protocolVersion2p0) {
		return c.fallThrough2pX(r)
	}
	switch Message(r.MsgType) {
	case ParameterStatus:
		name := r.PopString()
		switch name {
		case "pgaddr":
		case "pgdsn":
			r.PopBytes() // discard
		case "suggested_pool_concurrency":
			i, err := strconv.Atoi(r.PopString())
			if err != nil {
				return &binaryProtocolError{err: fmt.Errorf(
					"decoding ParameterStatus suggested_pool_concurrency: %w",
					err)}
			}
			c.serverSettings.Set(name, i)
		case "system_config":
			p := r.PopSlice(r.PopUint32())
			d := p.PopSlice(p.PopUint32())
			id := d.PopUUID()
			desc, err := descriptor.Pop(d, c.protocolVersion)
			if err != nil {
				return &binaryProtocolError{err: fmt.Errorf(
					"decoding ParameterStatus system_config descriptor: %w",
					err)}
			} else if desc.ID != id {
				return &binaryProtocolError{err: fmt.Errorf(
					"system_config descriptor ids don't match: %v != %v",
					id, desc.ID)}
			}

			var cfg systemConfig
			codec, err := codecs.BuildDecoder(
				desc,
				reflect.TypeOf(cfg),
				codecs.Path("system_config"),
			)
			if err != nil {
				return &binaryProtocolError{err: fmt.Errorf(
					"building codec from ParameterStatus "+
						"system_config descriptor: %w", err)}
			}

			err = codec.Decode(p.PopSlice(p.PopUint32()), unsafe.Pointer(&cfg))
			if err != nil {
				return &binaryProtocolError{err: fmt.Errorf(
					"decoding ParameterStatus system_config: %w", err)}
			}

			c.systemConfig = cfg
		default:
			return &unexpectedMessageError{msg: fmt.Sprintf(
				"got ParameterStatus for unknown parameter %q", name)}
		}
	case LogMessage:
		severity := logMsgSeverityLookup[r.PopUint8()]
		code := r.PopUint32()
		message := r.PopString()
		ignoreHeaders(r)
		log.Println("SERVER MESSAGE", severity, code, message)
	default:
		msg := fmt.Sprintf("unexpected message type: 0x%x", r.MsgType)
		return &unexpectedMessageError{msg: msg}
	}

	return nil
}

func (c *protocolConnection) fallThrough2pX(r *buff.Reader) error {
	switch Message(r.MsgType) {
	case ParameterStatus:
		name := r.PopString()
		switch name {
		case "pgaddr":
		case "pgdsn":
			r.PopBytes() // discard
		case "suggested_pool_concurrency":
			i, err := strconv.Atoi(r.PopString())
			if err != nil {
				return &binaryProtocolError{err: fmt.Errorf(
					"decoding ParameterStatus suggested_pool_concurrency: %w",
					err)}
			}
			c.serverSettings.Set(name, i)
		case "system_config":
			p := r.PopSlice(r.PopUint32())
			d := p.PopSlice(p.PopUint32())
			id := d.PopUUID()
			desc, err := descriptor.PopV2(d, c.protocolVersion)
			if err != nil {
				return &binaryProtocolError{err: fmt.Errorf(
					"decoding ParameterStatus system_config descriptor: %w",
					err)}
			} else if desc.ID != id {
				return &binaryProtocolError{err: fmt.Errorf(
					"system_config descriptor ids don't match: %v != %v",
					id, desc.ID)}
			}

			var cfg systemConfig
			codec, err := codecs.BuildDecoderV2(
				&desc,
				reflect.TypeOf(cfg),
				codecs.Path("system_config"),
			)
			if err != nil {
				return &binaryProtocolError{err: fmt.Errorf(
					"building codec from ParameterStatus "+
						"system_config descriptor: %w", err)}
			}

			err = codec.Decode(
				p.PopSlice(p.PopUint32()), unsafe.Pointer(&cfg),
			)
			if err != nil {
				return &binaryProtocolError{err: fmt.Errorf(
					"decoding ParameterStatus system_config: %w", err)}
			}

			c.systemConfig = cfg
		default:
			return &unexpectedMessageError{msg: fmt.Sprintf(
				"got ParameterStatus for unknown parameter %q", name)}
		}
	case LogMessage:
		severity := logMsgSeverityLookup[r.PopUint8()]
		code := r.PopUint32()
		message := r.PopString()
		ignoreHeaders(r)
		log.Println("SERVER MESSAGE", severity, code, message)
	default:
		msg := fmt.Sprintf("unexpected message type: 0x%x", r.MsgType)
		return &unexpectedMessageError{msg: msg}
	}

	return nil
}
