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

package codecs

// todo better error messages for nested data structures

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/edgedb/edgedb-go/protocol/buff"
	"github.com/edgedb/edgedb-go/types"
)

const (
	setType = iota
	objectType
	baseScalarType
	scalarType
	tupleType
	namedTupleType
	arrayType
	enumType
)

// Codec interface
type Codec interface {
	// todo update name
	Decode(*buff.Message, reflect.Value)
	Encode(*buff.Writer, interface{})
	ID() types.UUID
	Type() reflect.Type
	setType(reflect.Type) error
}

// BuildCodec a decoder
func BuildCodec(msg *buff.Message) (Codec, error) {
	codecs := []Codec{}

	for msg.Len() > 0 {
		dType := msg.PopUint8()
		id := msg.PopUUID()
		var codec Codec

		switch dType {
		case setType:
			codec = popSetCodec(msg, id, codecs)
		case objectType:
			codec = popObjectCodec(msg, id, codecs)
		case baseScalarType:
			var err error
			codec, err = baseScalarCodec(id)
			if err != nil {
				return nil, err
			}
		case scalarType:
			// todo implement scalar type descriptor
			return nil, errors.New("scalar type descriptor not implemented")
		case tupleType:
			codec = popTupleCodec(msg, id, codecs)
		case namedTupleType:
			codec = popNamedTupleCodec(msg, id, codecs)
		case arrayType:
			codec = popArrayCodec(msg, id, codecs)
		case enumType:
			// todo implement enum type descriptor
			return nil, errors.New("enum type descriptor not implemented")
		default:
			if 0x80 <= dType && dType <= 0xff {
				// ignore unknown type annotations
				msg.PopBytes()
				break
			}

			return nil, fmt.Errorf("unknown descriptor type 0x%x", dType)
		}

		codecs = append(codecs, codec)
	}

	return codecs[len(codecs)-1], nil
}

// BuildTypedCodec builds a codec for decoding into a specific type.
func BuildTypedCodec(msg *buff.Message, t reflect.Type) (Codec, error) {
	codec, err := BuildCodec(msg)
	if err != nil {
		return nil, err
	}

	if err := codec.setType(t); err != nil {
		return nil, fmt.Errorf(
			"the \"out\" argument does not match query schema: %w",
			err,
		)
	}

	return codec, nil
}
