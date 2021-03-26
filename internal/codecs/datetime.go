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
	dateTimeType      = reflect.TypeOf(time.Time{})
	localDateTimeType = reflect.TypeOf(types.LocalDateTime{})
	localDateType     = reflect.TypeOf(types.LocalDate{})
	localTimeType     = reflect.TypeOf(types.LocalTime{})
	durationType      = reflect.TypeOf(types.Duration(0))
)

var (
	dateTimeID  = types.UUID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0xa}
	localDTID   = types.UUID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0xb}
	localDateID = types.UUID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0xc}
	localTimeID = types.UUID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0xd}
	durationID  = types.UUID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0xe}
)

// DateTimeMarshaler is the interface implemented by an object
// that can marshal itself into the datetime wire format.
// https://www.edgedb.com/docs/internals/protocol/dataformats#std-datetime
//
// MarshalEdgeDBDateTime encodes the receiver
// into a binary form and returns the result.
type DateTimeMarshaler interface {
	MarshalEdgeDBDateTime() ([]byte, error)
}

// DateTimeUnmarshaler is the interface implemented by an object
// that can unmarshal the datetime wire format representation of itself.
// https://www.edgedb.com/docs/internals/protocol/dataformats#std-datetime
//
// UnmarshalEdgeDBDateTime must be able to decode the datetime wire format.
// UnmarshalEdgeDBDateTime must copy the data if it wishes to retain the data
// after returning.
type DateTimeUnmarshaler interface {
	UnmarshalEdgeDBDateTime(data []byte) error
}

type dateTimeCodec struct{}

func (c *dateTimeCodec) Type() reflect.Type { return dateTimeType }

func (c *dateTimeCodec) DescriptorID() types.UUID { return dateTimeID }

func (c *dateTimeCodec) Decode(r *buff.Reader, out unsafe.Pointer) {
	val := int64(r.PopUint64())
	seconds := val / 1_000_000
	microseconds := val % 1_000_000
	*(*time.Time)(out) = time.Unix(
		946_684_800+seconds,
		1_000*microseconds,
	).UTC()
}

func (c *dateTimeCodec) Encode(
	w *buff.Writer,
	val interface{},
	path Path,
) error {
	switch date := val.(type) {
	case time.Time:
		seconds := date.Unix() - 946_684_800
		nanoseconds := int64(date.Sub(time.Unix(date.Unix(), 0)))
		microseconds := seconds*1_000_000 + nanoseconds/1_000
		w.PushUint32(8) // data length
		w.PushUint64(uint64(microseconds))
	case DateTimeMarshaler:
		data, err := date.MarshalEdgeDBDateTime()
		if err != nil {
			return err
		}

		w.BeginBytes()
		w.PushBytes(data)
		w.EndBytes()
	default:
		return fmt.Errorf("expected %v to be time.Time got %T", path, val)
	}

	return nil
}

// LocalDateTimeMarshaler is the interface implemented by an object
// that can marshal itself into the local_datetime wire format.
// https://www.edgedb.com/docs/internals/protocol/dataformats
//
// MarshalEdgeDBLocalDateTime encodes the receiver
// into a binary form and returns the result.
type LocalDateTimeMarshaler interface {
	MarshalEdgeDBLocalDateTime() ([]byte, error)
}

// LocalDateTimeUnmarshaler is the interface implemented by an object
// that can unmarshal the local_datetime wire format representation of itself.
// https://www.edgedb.com/docs/internals/protocol/dataformats
//
// UnmarshalEdgeDBLocalDateTime must be able to decode the local_datetime wire
// format. UnmarshalEdgeDBLocalDateTime must copy the data if it wishes to
// retain the data after returning.
type LocalDateTimeUnmarshaler interface {
	UnmarshalEdgeDBLocalDateTime(data []byte) error
}

type localDateTimeCodec struct{}

func (c *localDateTimeCodec) Type() reflect.Type { return localDateTimeType }

func (c *localDateTimeCodec) DescriptorID() types.UUID { return localDTID }

// localDateTimeLayout is the memory layout for edgedbtypes.LocalDateTime
type localDateTimeLayout struct {
	usec uint64
}

func (c *localDateTimeCodec) Encode(
	w *buff.Writer,
	val interface{},
	path Path,
) error {
	switch in := val.(type) {
	case types.LocalDateTime:
		val := (*localDateTimeLayout)(unsafe.Pointer(&in))
		w.PushUint32(8)
		w.PushUint64(val.usec - 63_082_281_600_000_000)
	case LocalDateTimeMarshaler:
		data, err := in.MarshalEdgeDBLocalDateTime()
		if err != nil {
			return err
		}

		w.BeginBytes()
		w.PushBytes(data)
		w.EndBytes()
	default:
		return fmt.Errorf(
			"expected %v to be edgedb.LocalDateTime got %T", path, val,
		)
	}

	return nil
}

func (c *localDateTimeCodec) Decode(r *buff.Reader, out unsafe.Pointer) {
	(*localDateTimeLayout)(out).usec = r.PopUint64() + 63_082_281_600_000_000
}

// LocalDateMarshaler is the interface implemented by an object
// that can marshal itself into the local_date wire format.
// https://www.edgedb.com/docs/internals/protocol/dataformats#std-local-date
//
// MarshalEdgeDBLocalDate encodes the receiver
// into a binary form and returns the result.
type LocalDateMarshaler interface {
	MarshalEdgeDBLocalDate() ([]byte, error)
}

// LocalDateUnmarshaler is the interface implemented by an object
// that can unmarshal the local_date wire format representation of itself.
// https://www.edgedb.com/docs/internals/protocol/dataformats#std-local-date
//
// UnmarshalEdgeDBLocalDate must be able to decode the local_date wire format.
// UnmarshalEdgeDBLocalDate must copy the data if it wishes to retain the data
// after returning.
type LocalDateUnmarshaler interface {
	UnmarshalEdgeDBLocalDate(data []byte) error
}

type localDateCodec struct{}

func (c *localDateCodec) Type() reflect.Type { return localDateType }

func (c *localDateCodec) DescriptorID() types.UUID { return localDateID }

// localDateLayout is the memory layout for edgedbtypes.LocalDate
type localDateLayout struct {
	days uint32
}

func (c *localDateCodec) Encode(
	w *buff.Writer,
	val interface{},
	path Path,
) error {
	switch in := val.(type) {
	case types.LocalDate:
		w.PushUint32(4)
		w.PushUint32((*localDateLayout)(unsafe.Pointer(&in)).days - 730119)
	case LocalDateMarshaler:
		data, err := in.MarshalEdgeDBLocalDate()
		if err != nil {
			return err
		}

		w.BeginBytes()
		w.PushBytes(data)
		w.EndBytes()
	default:
		return fmt.Errorf(
			"expected %v to be edgedb.LocalDate got %T", path, val,
		)
	}

	return nil
}

func (c *localDateCodec) Decode(r *buff.Reader, out unsafe.Pointer) {
	(*localDateLayout)(out).days = r.PopUint32() + 730119
}

// LocalTimeMarshaler is the interface implemented by an object
// that can marshal itself into the local_time wire format.
// https://www.edgedb.com/docs/internals/protocol/dataformats#std-local-time
//
// MarshalEdgeDBLocalTime encodes the receiver
// into a binary form and returns the result.
type LocalTimeMarshaler interface {
	MarshalEdgeDBLocalTime() ([]byte, error)
}

// LocalTimeUnmarshaler is the interface implemented by an object
// that can unmarshal the local_time wire format representation of itself.
// https://www.edgedb.com/docs/internals/protocol/dataformats#std-local-time
//
// UnmarshalEdgeDBLocalTime must be able to decode the local_time wire format.
// UnmarshalEdgeDBLocalTime must copy the data if it wishes to retain the data
// after returning.
type LocalTimeUnmarshaler interface {
	UnmarshalEdgeDBLocalTime(data []byte) error
}

type localTimeCodec struct{}

func (c *localTimeCodec) Type() reflect.Type { return localTimeType }

func (c *localTimeCodec) DescriptorID() types.UUID { return localTimeID }

// localTimeLayout is the memory layout for edgedbtypes.LocalTime
type localTimeLayout struct {
	usec uint64
}

func (c *localTimeCodec) Encode(
	w *buff.Writer,
	val interface{},
	path Path,
) error {
	switch in := val.(type) {
	case types.LocalTime:
		w.PushUint32(8)
		w.PushUint64((*localTimeLayout)(unsafe.Pointer(&in)).usec)
	case LocalTimeMarshaler:
		data, err := in.MarshalEdgeDBLocalTime()
		if err != nil {
			return err
		}

		w.BeginBytes()
		w.PushBytes(data)
		w.EndBytes()
	default:
		return fmt.Errorf(
			"expected %v to be edgedb.LocalTime got %T", path, val,
		)
	}

	return nil
}

func (c *localTimeCodec) Decode(r *buff.Reader, out unsafe.Pointer) {
	(*localTimeLayout)(out).usec = r.PopUint64()
}

// DurationMarshaler is the interface implemented by an object
// that can marshal itself into the duration wire format.
// https://www.edgedb.com/docs/internals/protocol/dataformats#std-duration
//
// MarshalEdgeDBDuration encodes the receiver
// into a binary form and returns the result.
type DurationMarshaler interface {
	MarshalEdgeDBDuration() ([]byte, error)
}

// DurationUnmarshaler is the interface implemented by an object
// that can unmarshal the duration wire format representation of itself.
// https://www.edgedb.com/docs/internals/protocol/dataformats#std-duration
//
// UnmarshalEdgeDBDuration must be able to decode the duration wire format.
// UnmarshalEdgeDBDuration must copy the data if it wishes to retain the data
// after returning.
type DurationUnmarshaler interface {
	UnmarshalEdgeDBDuration(data []byte) error
}

type durationCodec struct{}

func (c *durationCodec) Type() reflect.Type { return durationType }

func (c *durationCodec) DescriptorID() types.UUID { return durationID }

func (c *durationCodec) Decode(r *buff.Reader, out unsafe.Pointer) {
	*(*uint64)(out) = r.PopUint64()
	r.Discard(8) // reserved
}

func (c *durationCodec) Encode(
	w *buff.Writer,
	val interface{},
	path Path,
) error {
	switch in := val.(type) {
	case types.Duration:
		w.PushUint32(16) // data length
		w.PushUint64(uint64(in))
		w.PushUint32(0) // reserved
		w.PushUint32(0) // reserved
	case DurationMarshaler:
		data, err := in.MarshalEdgeDBDuration()
		if err != nil {
			return err
		}

		w.BeginBytes()
		w.PushBytes(data)
		w.EndBytes()
	default:
		return fmt.Errorf(
			"expected %v to be edgedb.Duration got %T", path, val,
		)
	}

	return nil
}
