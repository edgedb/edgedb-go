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
	"math/big"
	"reflect"
	"unsafe"

	"github.com/edgedb/edgedb-go/internal/buff"
	types "github.com/edgedb/edgedb-go/internal/edgedbtypes"
)

var (
	decimalID = types.UUID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 8}
	bigIntID  = types.UUID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0x10}

	bigIntType = reflect.TypeOf(&big.Int{})

	big10k  = big.NewInt(10_000)
	bigOne  = big.NewInt(1)
	bigZero = big.NewInt(0)
)

// BigIntMarshaler is the interface implemented by an object
// that can marshal itself into the bigint wire format.
// https://www.edgedb.com/docs/internals/protocol/dataformats#std-bigint
//
// MarshalEdgeDBBigInt encodes the receiver
// into a binary form and returns the result.
type BigIntMarshaler interface {
	MarshalEdgeDBBigInt() ([]byte, error)
}

// BigIntUnmarshaler is the interface implemented by an object
// that can unmarshal the bigint wire format representation of itself.
// https://www.edgedb.com/docs/internals/protocol/dataformats#std-bigint
//
// UnmarshalEdgeDBBigInt must be able to decode the bigint wire format.
// UnmarshalEdgeDBBigInt must copy the data if it wishes to retain the data
// after returning.
type BigIntUnmarshaler interface {
	UnmarshalEdgeDBBigInt(data []byte) error
}

type bigIntCodec struct{}

func (c *bigIntCodec) Type() reflect.Type { return bigIntType }

func (c *bigIntCodec) DescriptorID() types.UUID { return bigIntID }

func (c *bigIntCodec) Decode(r *buff.Reader, out unsafe.Pointer) {
	n := int(r.PopUint16())
	weight := big.NewInt(int64(r.PopUint16()))
	sign := r.PopUint16()
	r.Discard(2) // reserved

	result := (**big.Int)(out)
	if *result == nil {
		*result = &big.Int{}
	}

	digit := &big.Int{}
	shift := &big.Int{}

	for i := 0; i < n; i++ {
		shift.Exp(big10k, weight, nil)
		digit.SetBytes(r.Buf[:2])
		digit.Mul(digit, shift)
		(*result).Add(*result, digit)
		weight.Sub(weight, bigOne)
		r.Discard(2)
	}

	if sign == 0x4000 {
		(*result).Neg(*result)
	}
}

func (c *bigIntCodec) Encode(
	w *buff.Writer,
	val interface{},
	path Path,
) error {
	switch in := val.(type) {
	case *big.Int:
		// copy to prevent mutating the user's value
		cpy := &big.Int{}
		cpy.Set(in)

		var sign uint16
		if in.Sign() == -1 {
			sign = 0x4000
			cpy = cpy.Neg(cpy)
		}

		digits := []byte{}
		rem := &big.Int{}

		for cpy.CmpAbs(bigZero) != 0 {
			rem.Mod(cpy, big10k)

			// pad bytes
			bts := rem.Bytes()
			for len(bts) < 2 {
				bts = append([]byte{0}, bts...)
			}

			digits = append(bts, digits...)
			cpy = cpy.Div(cpy, big10k)
		}

		w.BeginBytes()
		w.PushUint16(uint16(len(digits) / 2))
		w.PushUint16(uint16(len(digits)/2 - 1))
		w.PushUint16(sign)
		w.PushUint16(0) // reserved
		w.PushBytes(digits)
		w.EndBytes()
	case BigIntMarshaler:
		data, err := in.MarshalEdgeDBBigInt()
		if err != nil {
			return err
		}

		w.BeginBytes()
		w.PushBytes(data)
		w.EndBytes()
	default:
		return fmt.Errorf("expected %v to be *big.Int got %T", path, val)
	}

	return nil
}

// DecimalMarshaler is the interface implemented by an object
// that can marshal itself into the decimal wire format.
// https://www.edgedb.com/docs/internals/protocol/dataformats#std-decimal
//
// MarshalEdgeDBDecimal encodes the receiver
// into a binary form and returns the result.
type DecimalMarshaler interface {
	MarshalEdgeDBDecimal() ([]byte, error)
}

// DecimalUnmarshaler is the interface implemented by an object
// that can unmarshal the decimal wire format representation of itself.
// https://www.edgedb.com/docs/internals/protocol/dataformats#std-decimal
//
// UnmarshalEdgeDBDecimal must be able to decode the decimal wire format.
// UnmarshalEdgeDBDecimal must copy the data if it wishes to retain the data
// after returning.
type DecimalUnmarshaler interface {
	UnmarshalEdgeDBDecimal(data []byte) error
}

type decimalEncoder struct{}

func (c *decimalEncoder) DescriptorID() types.UUID { return decimalID }

func (c *decimalEncoder) Encode(
	w *buff.Writer,
	val interface{},
	path Path,
) error {
	switch in := val.(type) {
	case DecimalMarshaler:
		data, err := in.MarshalEdgeDBDecimal()
		if err != nil {
			return err
		}

		w.BeginBytes()
		w.PushBytes(data)
		w.EndBytes()
	default:
		return fmt.Errorf("expected %v to be *big.Int got %T", path, val)
	}

	return nil
}
