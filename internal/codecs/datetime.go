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

func (c *dateTimeCodec) Decode(r *buff.Reader, out unsafe.Pointer) error {
	val := int64(r.PopUint64())
	seconds := val / 1_000_000
	microseconds := val % 1_000_000
	*(*time.Time)(out) = time.Unix(
		946_684_800+seconds,
		1_000*microseconds,
	).UTC()
	return nil
}

type optionalDateTimeMarshaler interface {
	marshal.DateTimeMarshaler
	marshal.OptionalMarshaler
}

func (c *dateTimeCodec) Encode(
	w *buff.Writer,
	val interface{},
	path Path,
	required bool,
) error {
	switch in := val.(type) {
	case time.Time:
		return c.encodeData(w, in)
	case types.OptionalDateTime:
		data, ok := in.Get()
		return encodeOptional(w, !ok, required,
			func() error { return c.encodeData(w, data) },
			func() error {
				return missingValueError("edgedb.OptionalDateTime", path)
			})
	case optionalDateTimeMarshaler:
		return encodeOptional(w, in.Missing(), required,
			func() error { return c.encodeMarshaler(w, in, path) },
			func() error { return missingValueError(in, path) })
	case marshal.DateTimeMarshaler:
		return c.encodeMarshaler(w, in, path)
	default:
		return fmt.Errorf("expected %v to be time.Time, "+
			"edgedb.OptionalDateTime or DateTimeMarshaler got %T", path, val)
	}
}

func (c *dateTimeCodec) encodeData(w *buff.Writer, data time.Time) error {
	seconds := data.Unix() - 946_684_800
	nanoseconds := int64(data.Sub(time.Unix(data.Unix(), 0)))
	microseconds := seconds*1_000_000 + nanoseconds/1_000
	w.PushUint32(8) // data length
	w.PushUint64(uint64(microseconds))
	return nil
}

func (c *dateTimeCodec) encodeMarshaler(
	w *buff.Writer,
	val marshal.DateTimeMarshaler,
	path Path,
) error {
	return encodeMarshaler(w, val, val.MarshalEdgeDBDateTime, 8, path)
}

type optionalDateTime struct {
	val time.Time
	set bool
}

type optionalDateTimeDecoder struct {
	id types.UUID
}

func (c *optionalDateTimeDecoder) DescriptorID() types.UUID { return c.id }

func (c *optionalDateTimeDecoder) Decode(
	r *buff.Reader,
	out unsafe.Pointer,
) error {
	op := (*optionalDateTime)(out)
	op.set = true

	val := int64(r.PopUint64())
	seconds := val / 1_000_000
	microseconds := val % 1_000_000
	op.val = time.Unix(
		946_684_800+seconds,
		1_000*microseconds,
	).UTC()
	return nil
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

type optionalLocalDateTimeMarshaler interface {
	marshal.LocalDateTimeMarshaler
	marshal.OptionalMarshaler
}

func (c *localDateTimeCodec) Encode(
	w *buff.Writer,
	val interface{},
	path Path,
	required bool,
) error {
	switch in := val.(type) {
	case types.LocalDateTime:
		return c.encodeData(w, in)
	case types.OptionalLocalDateTime:
		data, ok := in.Get()
		return encodeOptional(w, !ok, required,
			func() error { return c.encodeData(w, data) },
			func() error {
				return missingValueError("edgedb.OptionalLocalDateTime", path)
			})
	case optionalLocalDateTimeMarshaler:
		return encodeOptional(w, in.Missing(), required,
			func() error { return c.encodeMarshaler(w, in, path) },
			func() error { return missingValueError(in, path) })
	case marshal.LocalDateTimeMarshaler:
		return c.encodeMarshaler(w, in, path)
	default:
		return fmt.Errorf("expected %v to be edgedb.LocalDateTime, "+
			"edgedb.OptionalLocalDateTime or LocalDateTimeMarshaler got %T",
			path, val)
	}
}

func (c *localDateTimeCodec) encodeData(
	w *buff.Writer,
	data types.LocalDateTime,
) error {
	v := (*localDateTimeLayout)(unsafe.Pointer(&data))
	w.PushUint32(8)
	w.PushUint64(v.usec - 63_082_281_600_000_000)
	return nil
}

func (c *localDateTimeCodec) encodeMarshaler(
	w *buff.Writer,
	val marshal.LocalDateTimeMarshaler,
	path Path,
) error {
	return encodeMarshaler(w, val, val.MarshalEdgeDBLocalDateTime, 8, path)
}

func (c *localDateTimeCodec) Decode(r *buff.Reader, out unsafe.Pointer) error {
	(*localDateTimeLayout)(out).usec = r.PopUint64() + 63_082_281_600_000_000
	return nil
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
) error {
	op := (*optionalLocalDateTime)(out)
	op.set = true
	op.val.usec = r.PopUint64() + 63_082_281_600_000_000
	return nil
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

type optionalLocalDateMarshaler interface {
	marshal.LocalDateMarshaler
	marshal.OptionalMarshaler
}

func (c *localDateCodec) Encode(
	w *buff.Writer,
	val interface{},
	path Path,
	required bool,
) error {
	switch in := val.(type) {
	case types.LocalDate:
		return c.encodeData(w, in)
	case types.OptionalLocalDate:
		data, ok := in.Get()
		return encodeOptional(w, !ok, required,
			func() error { return c.encodeData(w, data) },
			func() error {
				return missingValueError("edgedb.OptionalLocalDate", path)
			})
	case optionalLocalDateMarshaler:
		return encodeOptional(w, in.Missing(), required,
			func() error { return c.encodeMarshaler(w, in, path) },
			func() error { return missingValueError(in, path) })
	case marshal.LocalDateMarshaler:
		return c.encodeMarshaler(w, in, path)
	default:
		return fmt.Errorf("expected %v to be edgedb.LocalDate, "+
			"edgedb.OptionalLocalDate or LocalDateMarshaler got %T", path, val)
	}
}

func (c *localDateCodec) encodeData(
	w *buff.Writer,
	data types.LocalDate,
) error {
	w.PushUint32(4)
	w.PushUint32((*localDateLayout)(unsafe.Pointer(&data)).days - 730119)
	return nil
}

func (c *localDateCodec) encodeMarshaler(
	w *buff.Writer,
	val marshal.LocalDateMarshaler,
	path Path,
) error {
	return encodeMarshaler(w, val, val.MarshalEdgeDBLocalDate, 4, path)
}

func (c *localDateCodec) Decode(r *buff.Reader, out unsafe.Pointer) error {
	(*localDateLayout)(out).days = r.PopUint32() + 730119
	return nil
}

type optionalLocalDate struct {
	val localDateLayout
	set bool
}

type optionalLocalDateDecoder struct {
	id types.UUID
}

func (c *optionalLocalDateDecoder) DescriptorID() types.UUID { return c.id }

func (c *optionalLocalDateDecoder) Decode(
	r *buff.Reader,
	out unsafe.Pointer,
) error {
	op := (*optionalLocalDate)(out)
	op.set = true
	op.val.days = r.PopUint32() + 730119
	return nil
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

type optionalLocalTimeMarshaler interface {
	marshal.LocalTimeMarshaler
	marshal.OptionalMarshaler
}

func (c *localTimeCodec) Encode(
	w *buff.Writer,
	val interface{},
	path Path,
	required bool,
) error {
	switch in := val.(type) {
	case types.LocalTime:
		return c.encodeData(w, in)
	case types.OptionalLocalTime:
		data, ok := in.Get()
		return encodeOptional(w, !ok, required,
			func() error { return c.encodeData(w, data) },
			func() error {
				return missingValueError("edgedb.OptionalLocalTime", path)
			})
	case optionalLocalTimeMarshaler:
		return encodeOptional(w, in.Missing(), required,
			func() error { return c.encodeMarshaler(w, in, path) },
			func() error { return missingValueError(val, path) })
	case marshal.LocalTimeMarshaler:
		return c.encodeMarshaler(w, in, path)
	default:
		return fmt.Errorf("expected %v to be edgedb.LocalTime, "+
			"edgedb.OptionalLocalTime or LocalTimeMarshaler got %T", path, val)
	}
}

func (c *localTimeCodec) encodeData(
	w *buff.Writer,
	data types.LocalTime,
) error {
	w.PushUint32(8)
	w.PushUint64((*localTimeLayout)(unsafe.Pointer(&data)).usec)
	return nil
}

func (c *localTimeCodec) encodeMarshaler(
	w *buff.Writer,
	val marshal.LocalTimeMarshaler,
	path Path,
) error {
	return encodeMarshaler(w, val, val.MarshalEdgeDBLocalTime, 8, path)
}

func (c *localTimeCodec) Decode(r *buff.Reader, out unsafe.Pointer) error {
	(*localTimeLayout)(out).usec = r.PopUint64()
	return nil
}

type optionalLocalTime struct {
	val localTimeLayout
	set bool
}

type optionalLocalTimeDecoder struct {
	id types.UUID
}

func (c *optionalLocalTimeDecoder) DescriptorID() types.UUID { return c.id }

func (c *optionalLocalTimeDecoder) Decode(
	r *buff.Reader,
	out unsafe.Pointer,
) error {
	op := (*optionalLocalTime)(out)
	op.set = true
	op.val.usec = r.PopUint64()
	return nil
}

func (c *optionalLocalTimeDecoder) DecodeMissing(out unsafe.Pointer) {
	(*types.OptionalLocalTime)(out).Unset()
}

func (c *optionalLocalTimeDecoder) DecodePresent(out unsafe.Pointer) {}

type durationCodec struct{}

func (c *durationCodec) Type() reflect.Type { return durationType }

func (c *durationCodec) DescriptorID() types.UUID { return durationID }

func (c *durationCodec) Decode(r *buff.Reader, out unsafe.Pointer) error {
	*(*uint64)(out) = r.PopUint64()
	r.Discard(8) // reserved
	return nil
}

type optionalDurationMarshaler interface {
	marshal.DurationMarshaler
	marshal.OptionalMarshaler
}

func (c *durationCodec) Encode(
	w *buff.Writer,
	val interface{},
	path Path,
	required bool,
) error {
	switch in := val.(type) {
	case types.Duration:
		return c.encodeData(w, in)
	case types.OptionalDuration:
		data, ok := in.Get()
		return encodeOptional(w, !ok, required,
			func() error { return c.encodeData(w, data) },
			func() error {
				return missingValueError("edgedb.OptionalDuration", path)
			})
	case optionalDurationMarshaler:
		return encodeOptional(w, in.Missing(), required,
			func() error { return c.encodeMarshaler(w, in, path) },
			func() error { return missingValueError(in, path) })
	case marshal.DurationMarshaler:
		return c.encodeMarshaler(w, in, path)
	default:
		return fmt.Errorf("expected %v to be edgedb.Duration, "+
			"edgedb.OptionalDuration or DurationMarshaler got %T", path, val)
	}
}

func (c *durationCodec) encodeData(w *buff.Writer, data types.Duration) error {
	w.PushUint32(16) // data length
	w.PushUint64(uint64(data))
	w.PushUint32(0) // reserved
	w.PushUint32(0) // reserved
	return nil
}

func (c *durationCodec) encodeMarshaler(
	w *buff.Writer,
	val marshal.DurationMarshaler,
	path Path,
) error {
	return encodeMarshaler(w, val, val.MarshalEdgeDBDuration, 16, path)
}

type optionalDuration struct {
	val uint64
	set bool
}

type optionalDurationDecoder struct {
	id types.UUID
}

func (c *optionalDurationDecoder) DecodePresent(out unsafe.Pointer) {}

func (c *optionalDurationDecoder) DescriptorID() types.UUID { return c.id }

func (c *optionalDurationDecoder) Decode(
	r *buff.Reader,
	out unsafe.Pointer,
) error {
	op := (*optionalDuration)(out)
	op.set = true
	op.val = r.PopUint64()
	r.Discard(8) // reserved
	return nil
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
) error {
	rd := (*relativeDurationLayout)(out)
	rd.microseconds = r.PopUint64()
	rd.days = r.PopUint32()
	rd.months = r.PopUint32()
	return nil
}

type optionalRelativeDurationMarshaler interface {
	marshal.RelativeDurationMarshaler
	marshal.OptionalMarshaler
}

func (c *relativeDurationCodec) Encode(
	w *buff.Writer,
	val interface{},
	path Path,
	required bool,
) error {
	switch in := val.(type) {
	case types.RelativeDuration:
		return c.encodeData(w, in)
	case types.OptionalRelativeDuration:
		data, ok := in.Get()
		return encodeOptional(w, !ok, required,
			func() error { return c.encodeData(w, data) },
			func() error {
				return missingValueError(
					"edgedb.OptionalRelativeDuration",
					path,
				)
			})
	case optionalRelativeDurationMarshaler:
		return encodeOptional(w, in.Missing(), required,
			func() error { return c.encodeMarshaler(w, in, path) },
			func() error { return missingValueError(val, path) })
	case marshal.RelativeDurationMarshaler:
		return c.encodeMarshaler(w, in, path)
	default:
		return fmt.Errorf("expected %v to be edgedb.RelativeDuration, "+
			"edgedb.OptionalRelativeDuration or "+
			"RelativeDurationMarshaler got %T", path, val)
	}
}

func (c *relativeDurationCodec) encodeData(
	w *buff.Writer,
	data types.RelativeDuration,
) error {
	d := (*relativeDurationLayout)(unsafe.Pointer(&data))
	w.PushUint32(16) // data length
	w.PushUint64(d.microseconds)
	w.PushUint32(d.days)
	w.PushUint32(d.months)
	return nil
}

func (c *relativeDurationCodec) encodeMarshaler(
	w *buff.Writer,
	val marshal.RelativeDurationMarshaler,
	path Path,
) error {
	return encodeMarshaler(w, val, val.MarshalEdgeDBRelativeDuration, 16, path)
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
) error {
	op := (*optionalRelativeDuration)(out)
	op.set = true
	op.val.microseconds = r.PopUint64()
	op.val.days = r.PopUint32()
	op.val.months = r.PopUint32()
	return nil
}

func (c *optionalRelativeDurationDecoder) DecodeMissing(out unsafe.Pointer) {
	(*types.OptionalRelativeDuration)(out).Unset()
}

func (c *optionalRelativeDurationDecoder) DecodePresent(out unsafe.Pointer) {}
