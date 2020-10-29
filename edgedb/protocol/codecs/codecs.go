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

import (
	"fmt"

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
	Decode(*[]byte) interface{}
	Encode(*[]byte, interface{})
	ID() types.UUID
}

type idField struct {
	id types.UUID
}

func (i *idField) ID() types.UUID {
	return i.id
}

// Cache ...
type Cache map[types.UUID]Codec

// NewCache returns a cache with common types preallocated.
func NewCache() Cache {
	// todo add null tuple?
	return Cache{
		uuidID:      &UUID{idField{uuidID}},
		stringID:    &String{idField{stringID}},
		bytesID:     &Bytes{idField{bytesID}},
		int16ID:     &Int16{idField{int16ID}},
		int32ID:     &Int32{idField{int32ID}},
		int64ID:     &Int64{idField{int64ID}},
		float32ID:   &Float32{idField{float32ID}},
		float64ID:   &Float64{idField{float64ID}},
		decimalID:   nil, // not implemented
		boolID:      &Bool{idField{boolID}},
		dateTimeID:  &DateTime{idField{dateTimeID}},
		localDTID:   nil, // not implemented
		localDateID: nil, // not implemented
		localTimeID: nil, // not implemented
		durationID:  &Duration{idField{durationID}},
		jsonID:      &JSON{idField{jsonID}},
		bigIntID:    nil, // not implemented
	}
}

// UpdateCache a decoder
func UpdateCache(lookup Cache, bts *[]byte) {
	codecs := []Codec{}

	for len(*bts) > 0 {
		descriptorType := protocol.PopUint8(bts)
		id := protocol.PopUUID(bts)

		switch descriptorType {
		case setType:
			lookup[id] = popSetCodec(bts, id, codecs)
		case objectType:
			lookup[id] = popObjectCodec(bts, id, codecs)
		case baseScalarType:
			// base scalar types are preallocated
		case scalarType:
			panic("scalar type descriptor not implemented") // todo
		case tupleType:
			lookup[id] = popTupleCodec(bts, id, codecs)
		case namedTupleType:
			lookup[id] = popNamedTupleCodec(bts, id, codecs)
		case arrayType:
			lookup[id] = popArrayCodec(bts, id, codecs)
		case enumType:
			panic("enum type descriptor not implemented") // todo
		default:
			panic(fmt.Sprintf(
				"unknown descriptor type 0x%x:\n% x\n",
				descriptorType,
				bts,
			))
		}
		codecs = append(codecs, lookup[id])
	}
}
