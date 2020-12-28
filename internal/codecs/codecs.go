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
	"unsafe"

	"github.com/edgedb/edgedb-go/internal/buff"
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
	Decode(*buff.Reader, unsafe.Pointer)
	Encode(*buff.Writer, interface{}) error
	ID() types.UUID
	Type() reflect.Type
	setType(reflect.Type) error
}

// BuildCodec a decoder
func BuildCodec(r *buff.Reader) (Codec, error) {
	codecs := []Codec{}

	for len(r.Buf) > 0 {
		dType := r.PopUint8()
		id := r.PopUUID()
		var codec Codec

		switch dType {
		case setType:
			codec = popSetCodec(r, id, codecs)
		case objectType:
			codec = popObjectCodec(r, id, codecs)
		case baseScalarType:
			var err error
			codec, err = baseScalarCodec(id)
			if err != nil {
				return nil, err
			}
		case scalarType:
			return nil, errors.New("scalar type descriptor not implemented")
		case tupleType:
			codec = popTupleCodec(r, id, codecs)
		case namedTupleType:
			codec = popNamedTupleCodec(r, id, codecs)
		case arrayType:
			codec = popArrayCodec(r, id, codecs)
		case enumType:
			return nil, errors.New("enum type descriptor not implemented")
		default:
			if 0x80 <= dType && dType <= 0xff {
				// ignore unknown type annotations
				r.PopBytes()
				break
			}

			return nil, fmt.Errorf("unknown descriptor type 0x%x", dType)
		}

		codecs = append(codecs, codec)
	}

	return codecs[len(codecs)-1], nil
}

// BuildTypedCodec builds a codec for decoding into a specific type.
func BuildTypedCodec(r *buff.Reader, t reflect.Type) (Codec, error) {
	codec, err := BuildCodec(r)
	if err != nil {
		return nil, err
	}

	if err := codec.setType(t); err != nil {
		return nil, fmt.Errorf(
			"the \"out\" argument does not match query schema: %v", err,
		)
	}

	return codec, nil
}

func pAdd(p unsafe.Pointer, i uintptr) unsafe.Pointer {
	return unsafe.Pointer(uintptr(p) + i)
}

func calcStep(tp reflect.Type) int {
	step := int(tp.Size())
	a := tp.Align()

	if step%a > 0 {
		step = step/a + a
	}

	return step
}

// sliceHeader represent the memory layout for a slice.
type sliceHeader struct {
	Data unsafe.Pointer
	Len  int
	Cap  int
}
