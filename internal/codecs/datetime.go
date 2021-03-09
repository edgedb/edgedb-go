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
	localTimeType = reflect.TypeOf(types.LocalTime{})
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
type DateTime struct{}

// ID returns the descriptor id.
func (c *DateTime) ID() types.UUID { return dateTimeID }

func (c *DateTime) setType(typ reflect.Type, path Path) error {
	if typ != c.Type() {
		return fmt.Errorf("expected %v to be %v got %v", path, c.Type(), typ)
	}

	return nil
}

// Type returns the reflect.Type that this codec decodes to.
func (c *DateTime) Type() reflect.Type { return dateTimeType }

// Decode a datetime.
func (c *DateTime) Decode(r *buff.Reader, out unsafe.Pointer) {
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

func (c *LocalDateTime) setType(typ reflect.Type, path Path) error {
	if typ != localDTType {
		return fmt.Errorf(
			"expected %v to be %v got %v", path, localDTType, typ,
		)
	}

	return nil
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
func (c *LocalDateTime) Decode(r *buff.Reader, out unsafe.Pointer) {
	(*localDateTimeLayout)(out).usec = r.PopUint64() + 63_082_281_600_000_000
}

// LocalDate is an EdgeDB cal::local_date codec
type LocalDate struct{}

// ID returns the descriptor id.
func (c *LocalDate) ID() types.UUID { return localDateID }

// Type returns the reflect.Type that this codec decodes to.
func (c *LocalDate) Type() reflect.Type { return localDateType }

func (c *LocalDate) setType(typ reflect.Type, path Path) error {
	if typ != localDateType {
		return fmt.Errorf("expected %v to be %v got %v", path, c.Type(), typ)
	}

	return nil
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
func (c *LocalDate) Decode(r *buff.Reader, out unsafe.Pointer) {
	(*localDateLayout)(out).days = r.PopUint32() + 730119
}

// LocalTime is an EdgeDB cal::local_time codec
type LocalTime struct{}

// ID returns the descriptor id.
func (c *LocalTime) ID() types.UUID { return localTimeID }

// Type returns the reflect.Type that this codec decodes to.
func (c *LocalTime) Type() reflect.Type { return localTimeType }

func (c *LocalTime) setType(typ reflect.Type, path Path) error {
	if typ != localTimeType {
		return fmt.Errorf("expected %v to be %v got %v", path, c.Type(), typ)
	}

	return nil
}

// localTimeLayout is the memory layout for edgedbtypes.LocalTime
type localTimeLayout struct {
	usec uint64
}

// Encode a LocalTime
func (c *LocalTime) Encode(
	w *buff.Writer,
	val interface{},
	path Path,
) error {
	in, ok := val.(types.LocalTime)
	if !ok {
		return fmt.Errorf(
			"expected %v to be edgedb.LocalTime got %T", path, val,
		)
	}

	w.PushUint32(8)
	w.PushUint64((*localTimeLayout)(unsafe.Pointer(&in)).usec)
	return nil
}

// Decode a LocalTime
func (c *LocalTime) Decode(r *buff.Reader, out unsafe.Pointer) {
	(*localTimeLayout)(out).usec = r.PopUint64()
}

// Duration is an EdgeDB duration codec.
type Duration struct{}

// ID returns the descriptor id.
func (c *Duration) ID() types.UUID { return durationID }

func (c *Duration) setType(typ reflect.Type, path Path) error {
	if typ != c.Type() {
		return fmt.Errorf(
			"expected %v to be edgedb.Duration got %v", path, typ,
		)
	}

	return nil
}

// Type returns the reflect.Type that this codec decodes to.
func (c *Duration) Type() reflect.Type { return durationType }

// Decode a duration.
func (c *Duration) Decode(r *buff.Reader, out unsafe.Pointer) {
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
