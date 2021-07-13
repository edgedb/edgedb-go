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
	"github.com/edgedb/edgedb-go/internal/marshal"
)

var (
	dateTimeType         = reflect.TypeOf(time.Time{})
	localDateTimeType    = reflect.TypeOf(types.LocalDateTime{})
	localDateType        = reflect.TypeOf(types.LocalDate{})
	localTimeType        = reflect.TypeOf(types.LocalTime{})
	durationType         = reflect.TypeOf(types.Duration(0))
	relativeDurationType = reflect.TypeOf(types.RelativeDuration{})

	optionalDateTimeType      = reflect.TypeOf(types.OptionalDateTime{})
	optionalLocalDateTimeType = reflect.TypeOf(
		types.OptionalLocalDateTime{})
	optionalLocalDateType        = reflect.TypeOf(types.OptionalLocalDate{})
	optionalLocalTimeType        = reflect.TypeOf(types.OptionalLocalTime{})
	optionalDurationType         = reflect.TypeOf(types.OptionalDuration{})
	optionalRelativeDurationType = reflect.TypeOf(
		types.OptionalRelativeDuration{})
)

var (
	dateTimeID = types.UUID{
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0x0a}
	localDTID = types.UUID{
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0x0b}
	localDateID = types.UUID{
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0x0c}
	localTimeID = types.UUID{
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0x0d}
	durationID = types.UUID{
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0x0e}
	relativeDurationID = types.UUID{
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0x11}
)

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

func (c *dateTimeCodec) DecodeMissing(out unsafe.Pointer) {
	panic("unreachable")
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
	case types.OptionalDateTime:
		val, ok := date.Get()
		if !ok {
			return fmt.Errorf("cannot encode edgedb.OptionalDateTime at %v "+
				"because its value is missing", path)
		}

		seconds := val.Unix() - 946_684_800
		nanoseconds := int64(val.Sub(time.Unix(val.Unix(), 0)))
		microseconds := seconds*1_000_000 + nanoseconds/1_000
		w.PushUint32(8) // data length
		w.PushUint64(uint64(microseconds))
	case marshal.DateTimeMarshaler:
		data, err := date.MarshalEdgeDBDateTime()
		if err != nil {
			return err
		}

		w.BeginBytes()
		w.PushBytes(data)
		w.EndBytes()
	default:
		return fmt.Errorf("expected %v to be time.Time, "+
			"edgedb.OptionalDateTime or DateTimeMarshaler got %T", path, val)
	}

	return nil
}

type optionalDateTime struct {
	val time.Time
	set bool
}

type optionalDateTimeDecoder struct {
	id types.UUID
}

func (c *optionalDateTimeDecoder) DescriptorID() types.UUID { return c.id }

func (c *optionalDateTimeDecoder) Decode(r *buff.Reader, out unsafe.Pointer) {
	op := (*optionalDateTime)(out)
	op.set = true

	val := int64(r.PopUint64())
	seconds := val / 1_000_000
	microseconds := val % 1_000_000
	op.val = time.Unix(
		946_684_800+seconds,
		1_000*microseconds,
	).UTC()
}

func (c *optionalDateTimeDecoder) DecodeMissing(out unsafe.Pointer) {
	(*types.OptionalDateTime)(out).Unset()
}

func (c *optionalDateTimeDecoder) DecodePresent(out unsafe.Pointer) {}

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
	case types.OptionalLocalDateTime:
		val, ok := in.Get()
		if !ok {
			return fmt.Errorf("cannot encode edgedb.OptionalLocalDateTime "+
				"at %v because its value is missing", path)
		}

		v := (*localDateTimeLayout)(unsafe.Pointer(&val))
		w.PushUint32(8)
		w.PushUint64(v.usec - 63_082_281_600_000_000)
	case marshal.LocalDateTimeMarshaler:
		data, err := in.MarshalEdgeDBLocalDateTime()
		if err != nil {
			return err
		}

		w.BeginBytes()
		w.PushBytes(data)
		w.EndBytes()
	default:
		return fmt.Errorf("expected %v to be edgedb.LocalDateTime, "+
			"edgedb.OptionalLocalDateTime or LocalDateTimeMarshaler got %T",
			path, val)
	}

	return nil
}

func (c *localDateTimeCodec) Decode(r *buff.Reader, out unsafe.Pointer) {
	(*localDateTimeLayout)(out).usec = r.PopUint64() + 63_082_281_600_000_000
}

func (c *localDateTimeCodec) DecodeMissing(out unsafe.Pointer) {
	panic("unreachable")
}

type optionalLocalDateTime struct {
	val localDateTimeLayout
	set bool
}

type optionalLocalDateTimeDecoder struct {
	id types.UUID
}

func (c *optionalLocalDateTimeDecoder) DescriptorID() types.UUID {
	return c.id
}

func (c *optionalLocalDateTimeDecoder) Decode(
	r *buff.Reader,
	out unsafe.Pointer,
) {
	op := (*optionalLocalDateTime)(out)
	op.set = true
	op.val.usec = r.PopUint64() + 63_082_281_600_000_000
}

func (c *optionalLocalDateTimeDecoder) DecodeMissing(out unsafe.Pointer) {
	(*types.OptionalLocalDateTime)(out).Unset()
}

func (c *optionalLocalDateTimeDecoder) DecodePresent(out unsafe.Pointer) {}

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
	case types.OptionalLocalDate:
		val, ok := in.Get()
		if !ok {
			return fmt.Errorf("cannot encode edgedb.OptionalLocalDate at %v "+
				"because its value is missing", path)
		}

		w.PushUint32(4)
		w.PushUint32((*localDateLayout)(unsafe.Pointer(&val)).days - 730119)
	case marshal.LocalDateMarshaler:
		data, err := in.MarshalEdgeDBLocalDate()
		if err != nil {
			return err
		}

		w.BeginBytes()
		w.PushBytes(data)
		w.EndBytes()
	default:
		return fmt.Errorf("expected %v to be edgedb.LocalDate, "+
			"edgedb.OptionalLocalDate or LocalDateMarshaler got %T", path, val)
	}

	return nil
}

func (c *localDateCodec) Decode(r *buff.Reader, out unsafe.Pointer) {
	(*localDateLayout)(out).days = r.PopUint32() + 730119
}

func (c *localDateCodec) DecodeMissing(out unsafe.Pointer) {
	panic("unreachable")
}

type optionalLocalDate struct {
	val localDateLayout
	set bool
}

type optionalLocalDateDecoder struct {
	id types.UUID
}

func (c *optionalLocalDateDecoder) DescriptorID() types.UUID { return c.id }

func (c *optionalLocalDateDecoder) Decode(r *buff.Reader, out unsafe.Pointer) {
	op := (*optionalLocalDate)(out)
	op.set = true
	op.val.days = r.PopUint32() + 730119
}

func (c *optionalLocalDateDecoder) DecodeMissing(out unsafe.Pointer) {
	(*types.OptionalLocalDate)(out).Unset()
}

func (c *optionalLocalDateDecoder) DecodePresent(out unsafe.Pointer) {}

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
	case types.OptionalLocalTime:
		val, ok := in.Get()
		if !ok {
			return fmt.Errorf("cannot encode edgedb.OptionalLocalTime at %v "+
				"because its value is missing", path)
		}

		w.PushUint32(8)
		w.PushUint64((*localTimeLayout)(unsafe.Pointer(&val)).usec)
	case marshal.LocalTimeMarshaler:
		data, err := in.MarshalEdgeDBLocalTime()
		if err != nil {
			return err
		}

		w.BeginBytes()
		w.PushBytes(data)
		w.EndBytes()
	default:
		return fmt.Errorf("expected %v to be edgedb.LocalTime, "+
			"edgedb.OptionalLocalTime or LocalTimeMarshaler got %T", path, val)
	}

	return nil
}

func (c *localTimeCodec) Decode(r *buff.Reader, out unsafe.Pointer) {
	(*localTimeLayout)(out).usec = r.PopUint64()
}

func (c *localTimeCodec) DecodeMissing(out unsafe.Pointer) {
	panic("unreachable")
}

type optionalLocalTime struct {
	val localTimeLayout
	set bool
}

type optionalLocalTimeDecoder struct {
	id types.UUID
}

func (c *optionalLocalTimeDecoder) DescriptorID() types.UUID { return c.id }

func (c *optionalLocalTimeDecoder) Decode(r *buff.Reader, out unsafe.Pointer) {
	op := (*optionalLocalTime)(out)
	op.set = true
	op.val.usec = r.PopUint64()
}

func (c *optionalLocalTimeDecoder) DecodeMissing(out unsafe.Pointer) {
	(*types.OptionalLocalTime)(out).Unset()
}

func (c *optionalLocalTimeDecoder) DecodePresent(out unsafe.Pointer) {}

type durationCodec struct{}

func (c *durationCodec) Type() reflect.Type { return durationType }

func (c *durationCodec) DescriptorID() types.UUID { return durationID }

func (c *durationCodec) Decode(r *buff.Reader, out unsafe.Pointer) {
	*(*uint64)(out) = r.PopUint64()
	r.Discard(8) // reserved
}

func (c *durationCodec) DecodeMissing(out unsafe.Pointer) {
	panic("unreachable")
}

func (c *optionalDurationDecoder) DecodePresent(out unsafe.Pointer) {}

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
	case types.OptionalDuration:
		val, ok := in.Get()
		if !ok {
			return fmt.Errorf("cannot encode edgedb.OptionalDuration at %v "+
				"because its value is missing", path)
		}

		w.PushUint32(16) // data length
		w.PushUint64(uint64(val))
		w.PushUint32(0) // reserved
		w.PushUint32(0) // reserved
	case marshal.DurationMarshaler:
		data, err := in.MarshalEdgeDBDuration()
		if err != nil {
			return err
		}

		w.BeginBytes()
		w.PushBytes(data)
		w.EndBytes()
	default:
		return fmt.Errorf("expected %v to be edgedb.Duration, "+
			"edgedb.OptionalDuration or DurationMarshaler got %T", path, val)
	}

	return nil
}

type optionalDuration struct {
	val uint64
	set bool
}

type optionalDurationDecoder struct {
	id types.UUID
}

func (c *optionalDurationDecoder) DescriptorID() types.UUID { return c.id }

func (c *optionalDurationDecoder) Decode(r *buff.Reader, out unsafe.Pointer) {
	op := (*optionalDuration)(out)
	op.set = true
	op.val = r.PopUint64()
	r.Discard(8) // reserved
}

func (c *optionalDurationDecoder) DecodeMissing(out unsafe.Pointer) {
	(*types.OptionalDuration)(out).Unset()
}

type relativeDurationCodec struct{}

func (c *relativeDurationCodec) Type() reflect.Type {
	return relativeDurationType
}

func (c *relativeDurationCodec) DescriptorID() types.UUID {
	return relativeDurationID
}

type relativeDurationLayout struct {
	microseconds uint64
	days         uint32
	months       uint32
}

func (c *relativeDurationCodec) Decode(
	r *buff.Reader,
	out unsafe.Pointer,
) {
	rd := (*relativeDurationLayout)(out)
	rd.microseconds = r.PopUint64()
	rd.days = r.PopUint32()
	rd.months = r.PopUint32()
}

func (c *relativeDurationCodec) DecodeMissing(out unsafe.Pointer) {
	panic("unreachable")
}

func (c *relativeDurationCodec) Encode(
	w *buff.Writer,
	val interface{},
	path Path,
) error {
	switch in := val.(type) {
	case types.RelativeDuration:
		data := (*relativeDurationLayout)(unsafe.Pointer(&in))
		w.PushUint32(16) // data length
		w.PushUint64(data.microseconds)
		w.PushUint32(data.days)
		w.PushUint32(data.months)
	case types.OptionalRelativeDuration:
		val, ok := in.Get()
		if !ok {
			return fmt.Errorf("cannot encode edgedb.OptionalRelativeDuration "+
				"at %v because its value is missing", path)
		}

		data := (*relativeDurationLayout)(unsafe.Pointer(&val))
		w.PushUint32(16) // data length
		w.PushUint64(data.microseconds)
		w.PushUint32(data.days)
		w.PushUint32(data.months)
	case marshal.RelativeDurationMarshaler:
		data, err := in.MarshalEdgeDBRelativeDuration()
		if err != nil {
			return err
		}

		w.BeginBytes()
		w.PushBytes(data)
		w.EndBytes()
	default:
		return fmt.Errorf("expected %v to be edgedb.RelativeDuration, "+
			"edgedb.OptionalRelativeDuration or "+
			"RelativeDurationMarshaler got %T", path, val)
	}

	return nil
}

type optionalRelativeDuration struct {
	val relativeDurationLayout
	set bool
}

type optionalRelativeDurationDecoder struct {
	id types.UUID
}

func (c *optionalRelativeDurationDecoder) DescriptorID() types.UUID {
	return c.id
}

func (c *optionalRelativeDurationDecoder) Decode(
	r *buff.Reader,
	out unsafe.Pointer,
) {
	op := (*optionalRelativeDuration)(out)
	op.set = true
	op.val.microseconds = r.PopUint64()
	op.val.days = r.PopUint32()
	op.val.months = r.PopUint32()
}

func (c *optionalRelativeDurationDecoder) DecodeMissing(out unsafe.Pointer) {
	(*types.OptionalRelativeDuration)(out).Unset()
}

func (c *optionalRelativeDurationDecoder) DecodePresent(out unsafe.Pointer) {}
