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
	"reflect"
	"time"
	"unsafe"

	"github.com/edgedb/edgedb-go/internal/buff"
	types "github.com/edgedb/edgedb-go/internal/edgedbtypes"
)

var (
	dateTimeType  = reflect.TypeOf(time.Time{})
	localDTType   = reflect.TypeOf(types.LocalDateTime{})
	localDateType = reflect.TypeOf(types.LocalDate{})
	durationType  = reflect.TypeOf(types.Duration(0))
)

var (
	dateTimeID  = types.UUID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0xa}
	localDTID   = types.UUID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0xb}
	localDateID = types.UUID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0xc}
	localTimeID = types.UUID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0xd}
	durationID  = types.UUID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0xe}
)

// DateTime is an EdgeDB datetime type codec.
type DateTime struct {
	id  types.UUID
	typ reflect.Type
}

// ID returns the descriptor id.
func (c *DateTime) ID() types.UUID {
	return c.id
}
func (c *DateTime) setDefaultType() {}

func (c *DateTime) setType(typ reflect.Type, path Path) (bool, error) {
	if typ != c.typ {
		return false, fmt.Errorf(
			"expected %v to be %v got %v", path, c.typ, typ,
		)
	}

	return false, nil
}

// Type returns the reflect.Type that this codec decodes to.
func (c *DateTime) Type() reflect.Type {
	return c.typ
}

// Decode a datetime.
func (c *DateTime) Decode(r *buff.Reader, out reflect.Value) {
	c.DecodeReflect(r, out, Path(out.Type().String()))
}

// DecodeReflect decodes a datetime into a reflect.Value.
func (c *DateTime) DecodeReflect(
	r *buff.Reader,
	out reflect.Value,
	path Path,
) {
	if out.Type() != c.typ {
		panic(fmt.Errorf(
			"expected %v to be %v got %v", path, c.typ, out.Type(),
		))
	}

	c.DecodePtr(r, unsafe.Pointer(out.UnsafeAddr()))
}

// DecodePtr decodes a datetime into an unsafe.Pointer.
func (c *DateTime) DecodePtr(r *buff.Reader, out unsafe.Pointer) {
	val := int64(r.PopUint64())
	seconds := val / 1_000_000
	microseconds := val % 1_000_000
	*(*time.Time)(out) = time.Unix(
		946_684_800+seconds,
		1_000*microseconds,
	).UTC()
}

// Encode a datetime.
func (c *DateTime) Encode(w *buff.Writer, val interface{}, path Path) error {
	date, ok := val.(time.Time)
	if !ok {
		return fmt.Errorf("expected %v to be time.Time got %T", path, val)
	}

	seconds := date.Unix() - 946_684_800
	nanoseconds := int64(date.Sub(time.Unix(date.Unix(), 0)))
	microseconds := seconds*1_000_000 + nanoseconds/1_000
	w.PushUint32(8) // data length
	w.PushUint64(uint64(microseconds))
	return nil
}

// LocalDateTime is an EdgeDB cal::local_datetime codec
type LocalDateTime struct{}

// ID returns the descriptor id.
func (c *LocalDateTime) ID() types.UUID { return localDTID }

// Type returns the reflect.Type that this codec decodes to.
func (c *LocalDateTime) Type() reflect.Type { return localDTType }

func (c *LocalDateTime) setDefaultType() {}

func (c *LocalDateTime) setType(typ reflect.Type, path Path) (bool, error) {
	if typ != localDTType {
		return false, fmt.Errorf(
			"expected %v to be %v got %v", path, localDTType, typ,
		)
	}

	return false, nil
}

// localDateTimeLayout is the memory layout for edgedbtypes.LocalDateTime
type localDateTimeLayout struct {
	usec uint64
}

// Encode a LocalDateTime
func (c *LocalDateTime) Encode(
	w *buff.Writer,
	val interface{},
	path Path,
) error {
	in, ok := val.(types.LocalDateTime)
	if !ok {
		return fmt.Errorf(
			"expected %v to be edgedb.LocalDateTime got %T", path, val,
		)
	}

	w.PushUint32(8)
	w.PushUint64((*localDateTimeLayout)(unsafe.Pointer(&in)).usec -
		63_082_281_600_000_000)
	return nil
}

// Decode a LocalDateTime
func (c *LocalDateTime) Decode(r *buff.Reader, out reflect.Value) {
	c.DecodePtr(r, unsafe.Pointer(out.UnsafeAddr()))
}

// DecodeReflect decodes a LocalDateTime using reflection
func (c *LocalDateTime) DecodeReflect(
	r *buff.Reader,
	out reflect.Value,
	path Path,
) {
	if out.Type() != localDTType {
		panic(fmt.Sprintf(
			"expected %v to be edgedb.LocalDateTime got %v", path, out.Type(),
		))
	}

	c.DecodePtr(r, unsafe.Pointer(out.UnsafeAddr()))
}

// DecodePtr decodes a LocalDateTime into an unsafe.Pointer.
func (c *LocalDateTime) DecodePtr(r *buff.Reader, out unsafe.Pointer) {
	(*localDateTimeLayout)(out).usec = r.PopUint64() + 63_082_281_600_000_000
}

// LocalDate is an EdgeDB cal::local_datetime codec
type LocalDate struct{}

// ID returns the descriptor id.
func (c *LocalDate) ID() types.UUID { return localDateID }

// Type returns the reflect.Type that this codec decodes to.
func (c *LocalDate) Type() reflect.Type { return localDateType }

func (c *LocalDate) setDefaultType() {}

func (c *LocalDate) setType(typ reflect.Type, path Path) (bool, error) {
	if typ != localDateType {
		return false, fmt.Errorf(
			"expected %v to be %v got %v", path, localDateType, typ,
		)
	}

	return false, nil
}

// localDateLayout is the memory layout for edgedbtypes.LocalDate
type localDateLayout struct {
	days uint32
}

// Encode a LocalDate
func (c *LocalDate) Encode(
	w *buff.Writer,
	val interface{},
	path Path,
) error {
	in, ok := val.(types.LocalDate)
	if !ok {
		return fmt.Errorf(
			"expected %v to be edgedb.LocalDate got %T", path, val,
		)
	}

	w.PushUint32(4)
	w.PushUint32((*localDateLayout)(unsafe.Pointer(&in)).days - 730119)
	return nil
}

// Decode a LocalDate
func (c *LocalDate) Decode(r *buff.Reader, out reflect.Value) {
	c.DecodePtr(r, unsafe.Pointer(out.UnsafeAddr()))
}

// DecodeReflect decodes a LocalDateTime using reflection
func (c *LocalDate) DecodeReflect(
	r *buff.Reader,
	out reflect.Value,
	path Path,
) {
	if out.Type() != localDateType {
		panic(fmt.Sprintf(
			"expected %v to be edgedb.LocalDate got %v", path, out.Type(),
		))
	}

	c.DecodePtr(r, unsafe.Pointer(out.UnsafeAddr()))
}

// DecodePtr decodes a LocalDate into an unsafe.Pointer.
func (c *LocalDate) DecodePtr(r *buff.Reader, out unsafe.Pointer) {
	(*localDateLayout)(out).days = r.PopUint32() + 730119
}

// Duration is an EdgeDB duration codec.
type Duration struct{}

// ID returns the descriptor id.
func (c *Duration) ID() types.UUID { return durationID }

func (c *Duration) setDefaultType() {}

func (c *Duration) setType(typ reflect.Type, path Path) (bool, error) {
	if typ != durationType {
		return false, fmt.Errorf(
			"expected %v to be edgedb.Duration got %v", path, typ,
		)
	}

	return false, nil
}

// Type returns the reflect.Type that this codec decodes to.
func (c *Duration) Type() reflect.Type { return durationType }

// Decode a duration.
func (c *Duration) Decode(r *buff.Reader, out reflect.Value) {
	c.DecodePtr(r, unsafe.Pointer(out.UnsafeAddr()))
}

// DecodeReflect decodes a duration into a reflect.Value.
func (c *Duration) DecodeReflect(
	r *buff.Reader,
	out reflect.Value,
	path Path,
) {
	if out.Type() != durationType {
		panic(fmt.Errorf(
			"expected %v to be edgedb.Duration got %v", path, out.Type(),
		))
	}

	c.DecodePtr(r, unsafe.Pointer(out.UnsafeAddr()))
}

// DecodePtr decodes a duration into an unsafe.Pointer.
func (c *Duration) DecodePtr(r *buff.Reader, out unsafe.Pointer) {
	*(*uint64)(out) = r.PopUint64()
	r.Discard(8) // reserved
}

// Encode a duration.
func (c *Duration) Encode(w *buff.Writer, val interface{}, path Path) error {
	duration, ok := val.(types.Duration)
	if !ok {
		return fmt.Errorf(
			"expected %v to be edgedb.Duration got %T", path, val,
		)
	}

	w.PushUint32(16) // data length
	w.PushUint64(uint64(duration))
	w.PushUint32(0) // reserved
	w.PushUint32(0) // reserved
	return nil
}
