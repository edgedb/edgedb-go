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

	"github.com/edgedb/edgedb-go/edgedb/protocol"
	"github.com/edgedb/edgedb-go/edgedb/types"
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
	Decode(*[]byte, reflect.Value)
	Encode(*[]byte, interface{})
	ID() types.UUID
	Type() reflect.Type
	setType(reflect.Type) error
}

// BuildCodec a decoder
func BuildCodec(bts *[]byte) (Codec, error) {
	codecs := []Codec{}

	for len(*bts) > 0 {
		descriptorType := protocol.PopUint8(bts)
		id := protocol.PopUUID(bts)
		var codec Codec

		switch descriptorType {
		case setType:
			codec = popSetCodec(bts, id, codecs)
		case objectType:
			codec = popObjectCodec(bts, id, codecs)
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
			codec = popTupleCodec(bts, id, codecs)
		case namedTupleType:
			codec = popNamedTupleCodec(bts, id, codecs)
		case arrayType:
			codec = popArrayCodec(bts, id, codecs)
		case enumType:
			// todo implement enum type descriptor
			return nil, errors.New("enum type descriptor not implemented")
		default:
			return nil, fmt.Errorf(
				"unknown descriptor type 0x%x",
				descriptorType,
			)
		}

		codecs = append(codecs, codec)
	}

	root := codecs[len(codecs)-1]
	return root, nil
}

// BuildTypedCodec builds a codec for decoding into a specific type.
func BuildTypedCodec(bts *[]byte, t reflect.Type) (Codec, error) {
	codec, err := BuildCodec(bts)
	if err != nil {
		return nil, err
	}

	if err := codec.setType(t); err != nil {
		return nil, err
	}

	return codec, nil
}
