// This source file is part of the EdgeDB open source project.
//
// Copyright EdgeDB Inc. and the EdgeDB authors.
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

	"github.com/geldata/gel-go/internal/buff"
	types "github.com/geldata/gel-go/internal/geltypes"
	"github.com/geldata/gel-go/internal/marshal"
)

// DateTimeCodec encodes/decodes time.Time values.
type DateTimeCodec struct{}

// Type returns the type the codec encodes/decodes
func (c *DateTimeCodec) Type() reflect.Type { return dateTimeType }

// DescriptorID returns the codecs descriptor id.
func (c *DateTimeCodec) DescriptorID() types.UUID { return DateTimeID }

// Decode decodes a value
func (c *DateTimeCodec) Decode(r *buff.Reader, out unsafe.Pointer) error {
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

// Encode encodes a value
func (c *DateTimeCodec) Encode(
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
				return missingValueError("gel.OptionalDateTime", path)
			})
	case optionalDateTimeMarshaler:
		return encodeOptional(w, in.Missing(), required,
			func() error { return c.encodeMarshaler(w, in, path) },
			func() error { return missingValueError(in, path) })
	case marshal.DateTimeMarshaler:
		return c.encodeMarshaler(w, in, path)
	default:
		return fmt.Errorf("expected %v to be time.Time, "+
			"gel.OptionalDateTime or DateTimeMarshaler got %T", path, val)
	}
}

func (c *DateTimeCodec) encodeData(w *buff.Writer, data time.Time) error {
	seconds := data.Unix() - 946_684_800
	nanoseconds := int64(data.Sub(time.Unix(data.Unix(), 0)))

	rounded := nanoseconds / 1_000
	remainder := nanoseconds % 1_000
	if remainder == 500 && rounded%2 == 1 || remainder > 500 {
		rounded++
	}

	microseconds := seconds*1_000_000 + rounded
	w.PushUint32(8) // data length
	w.PushUint64(uint64(microseconds))
	return nil
}

func (c *DateTimeCodec) encodeMarshaler(
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

type optionalDateTimeDecoder struct{}

func (c *optionalDateTimeDecoder) DescriptorID() types.UUID {
	return DateTimeID
}

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

func (c *optionalDateTimeDecoder) DecodePresent(_ unsafe.Pointer) {}

// LocalDateTimeCodec encodes/decodes LocalDateTime values.
type LocalDateTimeCodec struct{}

// Type returns the type the codec encodes/decodes
func (c *LocalDateTimeCodec) Type() reflect.Type { return localDateTimeType }

// DescriptorID returns the codecs descriptor id.
func (c *LocalDateTimeCodec) DescriptorID() types.UUID { return LocalDTID }

// localDateTimeLayout is the memory layout for geltypes.LocalDateTime
type localDateTimeLayout struct {
	usec uint64
}

type optionalLocalDateTimeMarshaler interface {
	marshal.LocalDateTimeMarshaler
	marshal.OptionalMarshaler
}

// Encode encodes a value
func (c *LocalDateTimeCodec) Encode(
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
				return missingValueError("gel.OptionalLocalDateTime", path)
			})
	case optionalLocalDateTimeMarshaler:
		return encodeOptional(w, in.Missing(), required,
			func() error { return c.encodeMarshaler(w, in, path) },
			func() error { return missingValueError(in, path) })
	case marshal.LocalDateTimeMarshaler:
		return c.encodeMarshaler(w, in, path)
	default:
		return fmt.Errorf("expected %v to be gel.LocalDateTime, "+
			"gel.OptionalLocalDateTime or LocalDateTimeMarshaler got %T",
			path, val)
	}
}

func (c *LocalDateTimeCodec) encodeData(
	w *buff.Writer,
	data types.LocalDateTime,
) error {
	v := (*localDateTimeLayout)(unsafe.Pointer(&data))
	w.PushUint32(8)
	w.PushUint64(v.usec - 63_082_281_600_000_000)
	return nil
}

func (c *LocalDateTimeCodec) encodeMarshaler(
	w *buff.Writer,
	val marshal.LocalDateTimeMarshaler,
	path Path,
) error {
	return encodeMarshaler(w, val, val.MarshalEdgeDBLocalDateTime, 8, path)
}

// Decode decodes a value
func (c *LocalDateTimeCodec) Decode(r *buff.Reader, out unsafe.Pointer) error {
	(*localDateTimeLayout)(out).usec = r.PopUint64() + 63_082_281_600_000_000
	return nil
}

type optionalLocalDateTime struct {
	val localDateTimeLayout
	set bool
}

type optionalLocalDateTimeDecoder struct{}

func (c *optionalLocalDateTimeDecoder) DescriptorID() types.UUID {
	return LocalDTID
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

func (c *optionalLocalDateTimeDecoder) DecodePresent(_ unsafe.Pointer) {}

// LocalDateCodec encodes/decodes LocalDate values.
type LocalDateCodec struct{}

// Type returns the type the codec encodes/decodes
func (c *LocalDateCodec) Type() reflect.Type { return localDateType }

// DescriptorID returns the codecs descriptor id.
func (c *LocalDateCodec) DescriptorID() types.UUID { return LocalDateID }

// localDateLayout is the memory layout for geltypes.LocalDate
type localDateLayout struct {
	days uint32
}

type optionalLocalDateMarshaler interface {
	marshal.LocalDateMarshaler
	marshal.OptionalMarshaler
}

// Encode encodes a value
func (c *LocalDateCodec) Encode(
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
				return missingValueError("gel.OptionalLocalDate", path)
			})
	case optionalLocalDateMarshaler:
		return encodeOptional(w, in.Missing(), required,
			func() error { return c.encodeMarshaler(w, in, path) },
			func() error { return missingValueError(in, path) })
	case marshal.LocalDateMarshaler:
		return c.encodeMarshaler(w, in, path)
	default:
		return fmt.Errorf("expected %v to be gel.LocalDate, "+
			"gel.OptionalLocalDate or LocalDateMarshaler got %T", path, val)
	}
}

func (c *LocalDateCodec) encodeData(
	w *buff.Writer,
	data types.LocalDate,
) error {
	w.PushUint32(4)
	w.PushUint32((*localDateLayout)(unsafe.Pointer(&data)).days - 730119)
	return nil
}

func (c *LocalDateCodec) encodeMarshaler(
	w *buff.Writer,
	val marshal.LocalDateMarshaler,
	path Path,
) error {
	return encodeMarshaler(w, val, val.MarshalEdgeDBLocalDate, 4, path)
}

// Decode decodes a value
func (c *LocalDateCodec) Decode(r *buff.Reader, out unsafe.Pointer) error {
	(*localDateLayout)(out).days = r.PopUint32() + 730119
	return nil
}

type optionalLocalDate struct {
	val localDateLayout
	set bool
}

type optionalLocalDateDecoder struct{}

func (c *optionalLocalDateDecoder) DescriptorID() types.UUID {
	return LocalDateID
}

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

func (c *optionalLocalDateDecoder) DecodePresent(_ unsafe.Pointer) {}

// LocalTimeCodec encodes/decodes LocalTime values.
type LocalTimeCodec struct{}

// Type returns the type the codec encodes/decodes
func (c *LocalTimeCodec) Type() reflect.Type { return localTimeType }

// DescriptorID returns the codecs descriptor id.
func (c *LocalTimeCodec) DescriptorID() types.UUID { return LocalTimeID }

// localTimeLayout is the memory layout for geltypes.LocalTime
type localTimeLayout struct {
	usec uint64
}

type optionalLocalTimeMarshaler interface {
	marshal.LocalTimeMarshaler
	marshal.OptionalMarshaler
}

// Encode encodes a value
func (c *LocalTimeCodec) Encode(
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
				return missingValueError("gel.OptionalLocalTime", path)
			})
	case optionalLocalTimeMarshaler:
		return encodeOptional(w, in.Missing(), required,
			func() error { return c.encodeMarshaler(w, in, path) },
			func() error { return missingValueError(val, path) })
	case marshal.LocalTimeMarshaler:
		return c.encodeMarshaler(w, in, path)
	default:
		return fmt.Errorf("expected %v to be gel.LocalTime, "+
			"gel.OptionalLocalTime or LocalTimeMarshaler got %T", path, val)
	}
}

func (c *LocalTimeCodec) encodeData(
	w *buff.Writer,
	data types.LocalTime,
) error {
	w.PushUint32(8)
	w.PushUint64((*localTimeLayout)(unsafe.Pointer(&data)).usec)
	return nil
}

func (c *LocalTimeCodec) encodeMarshaler(
	w *buff.Writer,
	val marshal.LocalTimeMarshaler,
	path Path,
) error {
	return encodeMarshaler(w, val, val.MarshalEdgeDBLocalTime, 8, path)
}

// Decode decodes a value
func (c *LocalTimeCodec) Decode(r *buff.Reader, out unsafe.Pointer) error {
	(*localTimeLayout)(out).usec = r.PopUint64()
	return nil
}

type optionalLocalTime struct {
	val localTimeLayout
	set bool
}

type optionalLocalTimeDecoder struct{}

func (c *optionalLocalTimeDecoder) DescriptorID() types.UUID {
	return LocalTimeID
}

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

func (c *optionalLocalTimeDecoder) DecodePresent(_ unsafe.Pointer) {}

// DurationCodec encodes/decodes Duration values.
type DurationCodec struct{}

// Type returns the type the codec encodes/decodes
func (c *DurationCodec) Type() reflect.Type { return durationType }

// DescriptorID returns the codecs descriptor id.
func (c *DurationCodec) DescriptorID() types.UUID { return DurationID }

// Decode decodes a value
func (c *DurationCodec) Decode(r *buff.Reader, out unsafe.Pointer) error {
	*(*uint64)(out) = r.PopUint64()
	r.Discard(8) // reserved
	return nil
}

type optionalDurationMarshaler interface {
	marshal.DurationMarshaler
	marshal.OptionalMarshaler
}

// Encode encodes a value
func (c *DurationCodec) Encode(
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
				return missingValueError("gel.OptionalDuration", path)
			})
	case optionalDurationMarshaler:
		return encodeOptional(w, in.Missing(), required,
			func() error { return c.encodeMarshaler(w, in, path) },
			func() error { return missingValueError(in, path) })
	case marshal.DurationMarshaler:
		return c.encodeMarshaler(w, in, path)
	default:
		return fmt.Errorf("expected %v to be gel.Duration, "+
			"gel.OptionalDuration or DurationMarshaler got %T", path, val)
	}
}

func (c *DurationCodec) encodeData(w *buff.Writer, data types.Duration) error {
	w.PushUint32(16) // data length
	w.PushUint64(uint64(data))
	w.PushUint32(0) // reserved
	w.PushUint32(0) // reserved
	return nil
}

func (c *DurationCodec) encodeMarshaler(
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

type optionalDurationDecoder struct{}

func (c *optionalDurationDecoder) DecodePresent(_ unsafe.Pointer) {}

func (c *optionalDurationDecoder) DescriptorID() types.UUID {
	return DurationID
}

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

// RelativeDurationCodec encodes/decodes RelativeDuration values.
type RelativeDurationCodec struct{}

// Type returns the type the codec encodes/decodes
func (c *RelativeDurationCodec) Type() reflect.Type {
	return relativeDurationType
}

// DescriptorID returns the codecs descriptor id.
func (c *RelativeDurationCodec) DescriptorID() types.UUID {
	return RelativeDurationID
}

type relativeDurationLayout struct {
	microseconds uint64
	days         uint32
	months       uint32
}

// Decode decodes a value
func (c *RelativeDurationCodec) Decode(
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

// Encode encodes a value
func (c *RelativeDurationCodec) Encode(
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
					"gel.OptionalRelativeDuration",
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
		return fmt.Errorf("expected %v to be gel.RelativeDuration, "+
			"gel.OptionalRelativeDuration or "+
			"RelativeDurationMarshaler got %T", path, val)
	}
}

func (c *RelativeDurationCodec) encodeData(
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

func (c *RelativeDurationCodec) encodeMarshaler(
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

type optionalRelativeDurationDecoder struct{}

func (c *optionalRelativeDurationDecoder) DescriptorID() types.UUID {
	return RelativeDurationID
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

func (c *optionalRelativeDurationDecoder) DecodePresent(_ unsafe.Pointer) {}

// DateDurationCodec encodes/decodes DateDuration values.
type DateDurationCodec struct{}

// Type returns the type the codec encodes/decodes
func (c *DateDurationCodec) Type() reflect.Type {
	return dateDurationType
}

// DescriptorID returns the codecs descriptor id.
func (c *DateDurationCodec) DescriptorID() types.UUID {
	return DateDurationID
}

type dateDurationLayout struct {
	days   uint32
	months uint32
}

// Decode decodes a value
func (c *DateDurationCodec) Decode(
	r *buff.Reader,
	out unsafe.Pointer,
) error {
	rd := (*dateDurationLayout)(out)
	r.Discard(8) // microseconds are unused
	rd.days = r.PopUint32()
	rd.months = r.PopUint32()
	return nil
}

type optionalDateDurationMarshaler interface {
	marshal.DateDurationMarshaler
	marshal.OptionalMarshaler
}

// Encode encodes a value
func (c *DateDurationCodec) Encode(
	w *buff.Writer,
	val interface{},
	path Path,
	required bool,
) error {
	switch in := val.(type) {
	case types.DateDuration:
		return c.encodeData(w, in)
	case types.OptionalDateDuration:
		data, ok := in.Get()
		return encodeOptional(w, !ok, required,
			func() error { return c.encodeData(w, data) },
			func() error {
				return missingValueError(
					"gel.OptionalDateDuration",
					path,
				)
			})
	case optionalDateDurationMarshaler:
		return encodeOptional(w, in.Missing(), required,
			func() error { return c.encodeMarshaler(w, in, path) },
			func() error { return missingValueError(val, path) })
	case marshal.DateDurationMarshaler:
		return c.encodeMarshaler(w, in, path)
	default:
		return fmt.Errorf("expected %v to be gel.DateDuration, "+
			"gel.OptionalDateDuration or "+
			"DateDurationMarshaler got %T", path, val)
	}
}

func (c *DateDurationCodec) encodeData(
	w *buff.Writer,
	data types.DateDuration,
) error {
	d := (*dateDurationLayout)(unsafe.Pointer(&data))
	w.PushUint32(16) // data length
	w.PushUint64(0)  // microseconds are unused
	w.PushUint32(d.days)
	w.PushUint32(d.months)
	return nil
}

func (c *DateDurationCodec) encodeMarshaler(
	w *buff.Writer,
	val marshal.DateDurationMarshaler,
	path Path,
) error {
	return encodeMarshaler(w, val, val.MarshalEdgeDBDateDuration, 16, path)
}

type optionalDateDuration struct {
	val dateDurationLayout
	set bool
}

type optionalDateDurationDecoder struct{}

func (c *optionalDateDurationDecoder) DescriptorID() types.UUID {
	return DateDurationID
}

func (c *optionalDateDurationDecoder) Decode(
	r *buff.Reader,
	out unsafe.Pointer,
) error {
	op := (*optionalDateDuration)(out)
	op.set = true
	r.Discard(8)
	op.val.days = r.PopUint32()
	op.val.months = r.PopUint32()
	return nil
}

func (c *optionalDateDurationDecoder) DecodeMissing(out unsafe.Pointer) {
	(*types.OptionalDateDuration)(out).Unset()
}

func (c *optionalDateDurationDecoder) DecodePresent(_ unsafe.Pointer) {}
