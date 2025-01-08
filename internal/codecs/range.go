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
	"errors"
	"fmt"
	"reflect"
	"unsafe"

	"github.com/edgedb/edgedb-go/internal"
	"github.com/edgedb/edgedb-go/internal/buff"
	"github.com/edgedb/edgedb-go/internal/descriptor"
	types "github.com/edgedb/edgedb-go/internal/geltypes"
	"github.com/edgedb/edgedb-go/internal/introspect"
)

func buildRangeDecoder(
	desc descriptor.Descriptor,
	typ reflect.Type,
	path Path,
) (Decoder, error) {
	if typ == rangeInt32Type ||
		typ == rangeInt64Type ||
		typ == rangeFloat32Type ||
		typ == rangeFloat64Type ||
		typ == rangeDateTimeType ||
		typ == rangeLocalDateTimeType ||
		typ == rangeLocalDateType {
		return buildRequiredRangeDecoder(desc, typ, path)
	}

	if typ == optionalRangeInt32Type ||
		typ == optionalRangeInt64Type ||
		typ == optionalRangeFloat32Type ||
		typ == optionalRangeFloat64Type ||
		typ == optionalRangeDateTimeType ||
		typ == optionalRangeLocalDateTimeType ||
		typ == optionalRangeLocalDateType {
		return buildOptionalRangeDecoder(desc, typ, path)
	}

	return nil, fmt.Errorf(
		"expected %v to be an gel.Range type got %v",
		path, typ)
}

func buildRangeDecoderV2(
	desc *descriptor.V2,
	typ reflect.Type,
	path Path,
) (Decoder, error) {
	if typ == rangeInt32Type ||
		typ == rangeInt64Type ||
		typ == rangeFloat32Type ||
		typ == rangeFloat64Type ||
		typ == rangeDateTimeType ||
		typ == rangeLocalDateTimeType ||
		typ == rangeLocalDateType {
		return buildRequiredRangeDecoderV2(desc, typ, path)
	}

	if typ == optionalRangeInt32Type ||
		typ == optionalRangeInt64Type ||
		typ == optionalRangeFloat32Type ||
		typ == optionalRangeFloat64Type ||
		typ == optionalRangeDateTimeType ||
		typ == optionalRangeLocalDateTimeType ||
		typ == optionalRangeLocalDateType {
		return buildOptionalRangeDecoderV2(desc, typ, path)
	}

	return nil, fmt.Errorf(
		"expected %v to be an gel.Range type got %v",
		path, typ)
}

func buildRequiredRangeDecoder(
	desc descriptor.Descriptor,
	typ reflect.Type,
	path Path,
) (Decoder, error) {
	fieldDesc := desc.Fields[0].Desc
	lower, child, err := buildField(typ, "lower", path, fieldDesc)
	if err != nil {
		return nil, err
	}

	upper, _, err := buildField(typ, "upper", path, fieldDesc)
	if err != nil {
		return nil, err
	}

	incLower, _, err := buildField(typ, "inc_lower", path, fieldDesc)
	if err != nil {
		return nil, err
	}

	incUpper, _, err := buildField(typ, "inc_upper", path, fieldDesc)
	if err != nil {
		return nil, err
	}

	empty, _, err := buildField(typ, "empty", path, fieldDesc)
	if err != nil {
		return nil, err
	}

	return &rangeCodec{
		id:             desc.ID,
		child:          child.(OptionalDecoder),
		lowerOffset:    lower.Offset,
		upperOffset:    upper.Offset,
		incLowerOffset: incLower.Offset,
		incUpperOffset: incUpper.Offset,
		emptyOffset:    empty.Offset,
	}, nil
}

func buildRequiredRangeDecoderV2(
	desc *descriptor.V2,
	typ reflect.Type,
	path Path,
) (Decoder, error) {
	fieldDesc := desc.Fields[0].Desc
	lower, child, err := buildFieldV2(typ, "lower", path, &fieldDesc)
	if err != nil {
		return nil, err
	}

	upper, _, err := buildFieldV2(typ, "upper", path, &fieldDesc)
	if err != nil {
		return nil, err
	}

	incLower, _, err := buildFieldV2(typ, "inc_lower", path, &fieldDesc)
	if err != nil {
		return nil, err
	}

	incUpper, _, err := buildFieldV2(typ, "inc_upper", path, &fieldDesc)
	if err != nil {
		return nil, err
	}

	empty, _, err := buildFieldV2(typ, "empty", path, &fieldDesc)
	if err != nil {
		return nil, err
	}

	return &rangeCodec{
		id:             desc.ID,
		child:          child.(OptionalDecoder),
		lowerOffset:    lower.Offset,
		upperOffset:    upper.Offset,
		incLowerOffset: incLower.Offset,
		incUpperOffset: incUpper.Offset,
		emptyOffset:    empty.Offset,
	}, nil
}

func buildField(
	typ reflect.Type,
	name string,
	path Path,
	desc descriptor.Descriptor,
) (reflect.StructField, Decoder, error) {
	sf, ok := introspect.StructField(typ, name)
	if !ok {
		return reflect.StructField{}, nil, fmt.Errorf(
			"expected %v to have a field named %q", path, name)
	}

	switch name {
	case "inc_lower", "inc_upper", "empty":
		if sf.Type.Kind() != reflect.Bool {
			return reflect.StructField{}, nil, fmt.Errorf(
				"expected field %q to be bool",
				name)
		}

		return sf, nil, nil
	case "upper", "lower":
		child, err := buildScalarDecoder(
			desc,
			sf.Type,
			path.AddField(name),
		)
		if err != nil {
			return reflect.StructField{}, nil, err
		}

		if _, isOptional := child.(OptionalDecoder); !isOptional {
			typeName, ok := optionalTypeNameLookup[reflect.TypeOf(child)]
			if !ok {
				typeName = "OptionalUnmarshaler interface"
			}
			return reflect.StructField{}, nil, fmt.Errorf(
				"expected %v at %v.%v to be %v"+
					"because the field is not required",
				sf.Type, path, name, typeName)
		}

		return sf, child, nil
	default:
		return reflect.StructField{}, nil, errors.New("unreachable 10118")
	}
}

func buildFieldV2(
	typ reflect.Type,
	name string,
	path Path,
	desc *descriptor.V2,
) (reflect.StructField, Decoder, error) {
	sf, ok := introspect.StructField(typ, name)
	if !ok {
		return reflect.StructField{}, nil, fmt.Errorf(
			"expected %v to have a field named %q", path, name)
	}

	switch name {
	case "inc_lower", "inc_upper", "empty":
		if sf.Type.Kind() != reflect.Bool {
			return reflect.StructField{}, nil, fmt.Errorf(
				"expected field %q to be bool",
				name)
		}

		return sf, nil, nil
	case "upper", "lower":
		child, err := buildScalarDecoderV2(
			desc,
			sf.Type,
			path.AddField(name),
		)
		if err != nil {
			return reflect.StructField{}, nil, err
		}

		if _, isOptional := child.(OptionalDecoder); !isOptional {
			typeName, ok := optionalTypeNameLookup[reflect.TypeOf(child)]
			if !ok {
				typeName = "OptionalUnmarshaler interface"
			}
			return reflect.StructField{}, nil, fmt.Errorf(
				"expected %v at %v.%v to be %v"+
					"because the field is not required",
				sf.Type, path, name, typeName)
		}

		return sf, child, nil
	default:
		return reflect.StructField{}, nil, errors.New("unreachable 10118")
	}
}

type rangeCodec struct {
	id             types.UUID
	child          OptionalDecoder
	lowerOffset    uintptr
	upperOffset    uintptr
	incLowerOffset uintptr
	incUpperOffset uintptr
	emptyOffset    uintptr
}

func (c *rangeCodec) DescriptorID() types.UUID { return c.id }

func (c *rangeCodec) Decode(r *buff.Reader, out unsafe.Pointer) error {
	flags := r.PopUint8()
	empty := (flags & rangeEmpty) != 0
	incLower := (flags & rangeLBInc) != 0
	incUpper := (flags & rangeUBInc) != 0
	hasLower := (flags & (rangeEmpty | rangeLBInf)) == 0
	hasUpper := (flags & (rangeEmpty | rangeUBInf)) == 0

	if hasLower {
		p := pAdd(out, c.lowerOffset)
		subLen := r.PopUint32()
		if subLen == 0xffffffff {
			c.child.DecodeMissing(p)
		} else {
			err := c.child.Decode(r.PopSlice(subLen), p)
			if err != nil {
				return err
			}
			// todo check that the  popped slice was consumed
		}
	}

	if hasUpper {
		p := pAdd(out, c.upperOffset)
		subLen := r.PopUint32()
		if subLen == 0xffffffff {
			c.child.DecodeMissing(p)
		} else {
			subBuf := r.PopSlice(subLen)
			err := c.child.Decode(subBuf, p)
			if err != nil {
				return err
			}
			// todo check that the  popped slice was consumed
		}
	}

	*(*bool)(pAdd(out, c.emptyOffset)) = empty
	*(*bool)(pAdd(out, c.incLowerOffset)) = incLower
	*(*bool)(pAdd(out, c.incUpperOffset)) = incUpper
	return nil
}

func buildOptionalRangeDecoder(
	desc descriptor.Descriptor,
	typ reflect.Type,
	path Path,
) (OptionalDecoder, error) {
	val, ok := introspect.StructField(typ, "val")
	if !ok {
		return nil, fmt.Errorf("unreachable 11248: val not found")
	}

	codec, err := buildRequiredRangeDecoder(
		desc,
		val.Type,
		path.AddField("val"),
	)
	if err != nil {
		return nil, err
	}

	isSet, ok := introspect.StructField(typ, "isSet")
	if !ok {
		return nil, fmt.Errorf("unreachable 22467: isSet not found")
	}

	child := codec.(*rangeCodec)
	return &optionalRangeDecoder{
		id:     desc.ID,
		offset: isSet.Offset,
		child:  child,
	}, nil
}

func buildOptionalRangeDecoderV2(
	desc *descriptor.V2,
	typ reflect.Type,
	path Path,
) (OptionalDecoder, error) {
	val, ok := introspect.StructField(typ, "val")
	if !ok {
		return nil, fmt.Errorf("unreachable 11248: val not found")
	}

	codec, err := buildRequiredRangeDecoderV2(
		desc,
		val.Type,
		path.AddField("val"),
	)
	if err != nil {
		return nil, err
	}

	isSet, ok := introspect.StructField(typ, "isSet")
	if !ok {
		return nil, fmt.Errorf("unreachable 22467: isSet not found")
	}

	child := codec.(*rangeCodec)
	return &optionalRangeDecoder{
		id:     desc.ID,
		offset: isSet.Offset,
		child:  child,
	}, nil
}

type optionalRangeDecoder struct {
	id     types.UUID
	child  *rangeCodec
	offset uintptr
}

func (c *optionalRangeDecoder) DescriptorID() types.UUID { return c.id }

func (c *optionalRangeDecoder) Decode(
	r *buff.Reader,
	out unsafe.Pointer,
) error {
	err := c.child.Decode(r, out)
	if err != nil {
		return err
	}

	*(*bool)(pAdd(out, c.offset)) = true
	return nil
}

func (c *optionalRangeDecoder) DecodeMissing(out unsafe.Pointer) {
	switch c.child.child.(type) {
	case *optionalInt32Decoder:
		(*types.OptionalRangeInt32)(out).Unset()
	case *optionalInt64Decoder:
		(*types.OptionalRangeInt64)(out).Unset()
	case *optionalFloat32Decoder:
		(*types.OptionalRangeFloat32)(out).Unset()
	case *optionalFloat64Decoder:
		(*types.OptionalRangeFloat64)(out).Unset()
	case *optionalDateTimeDecoder:
		(*types.OptionalRangeDateTime)(out).Unset()
	case *optionalLocalDateTimeDecoder:
		(*types.OptionalRangeLocalDateTime)(out).Unset()
	case *optionalLocalDateDecoder:
		(*types.OptionalRangeLocalDate)(out).Unset()
	default:
		panic("unreachable 7189")
	}
}

func buildRangeEncoder(
	desc descriptor.Descriptor,
	version internal.ProtocolVersion,
) (Encoder, error) {
	child, err := BuildEncoder(desc.Fields[0].Desc, version)
	if err != nil {
		return nil, err
	}

	return &rangeEncoder{id: desc.ID, child: child}, nil
}

func buildRangeEncoderV2(
	desc *descriptor.V2,
	version internal.ProtocolVersion,
) (Encoder, error) {
	child, err := BuildEncoderV2(&desc.Fields[0].Desc, version)
	if err != nil {
		return nil, err
	}

	return &rangeEncoder{id: desc.ID, child: child}, nil
}

type rangeEncoder struct {
	id    types.UUID
	child Encoder
}

func (c *rangeEncoder) DescriptorID() types.UUID { return c.id }

func (c *rangeEncoder) Encode(
	w *buff.Writer,
	val interface{},
	path Path,
	required bool,
) error {
	switch in := val.(type) {
	case types.OptionalRangeInt32:
		data, ok := in.Get()
		return encodeOptional(w, !ok, required,
			func() error { return c.encode(w, data, path) },
			func() error {
				return missingValueError(
					"gel.OptionalRangeInt32",
					path,
				)
			},
		)
	case types.OptionalRangeInt64:
		data, ok := in.Get()
		return encodeOptional(w, !ok, required,
			func() error { return c.encode(w, data, path) },
			func() error {
				return missingValueError(
					"gel.OptionalRangeInt64",
					path,
				)
			},
		)
	case types.OptionalRangeFloat32:
		data, ok := in.Get()
		return encodeOptional(w, !ok, required,
			func() error { return c.encode(w, data, path) },
			func() error {
				return missingValueError(
					"gel.OptionalRangeFloat32",
					path,
				)
			},
		)
	case types.OptionalRangeFloat64:
		data, ok := in.Get()
		return encodeOptional(w, !ok, required,
			func() error { return c.encode(w, data, path) },
			func() error {
				return missingValueError(
					"gel.OptionalRangeFloat64",
					path,
				)
			},
		)
	case types.OptionalRangeDateTime:
		data, ok := in.Get()
		return encodeOptional(w, !ok, required,
			func() error { return c.encode(w, data, path) },
			func() error {
				return missingValueError(
					"gel.OptionalRangeDateTime",
					path,
				)
			},
		)
	case types.OptionalRangeLocalDateTime:
		data, ok := in.Get()
		return encodeOptional(w, !ok, required,
			func() error { return c.encode(w, data, path) },
			func() error {
				return missingValueError(
					"gel.OptionalRangeLocalDateTime",
					path,
				)
			},
		)
	case types.OptionalRangeLocalDate:
		data, ok := in.Get()
		return encodeOptional(w, !ok, required,
			func() error { return c.encode(w, data, path) },
			func() error {
				return missingValueError(
					"gel.OptionalRangeLocalDate",
					path,
				)
			},
		)
	default:
		return c.encode(w, val, path)
	}
}

func (c *rangeEncoder) encode(
	w *buff.Writer,
	val interface{},
	path Path,
) error {
	var (
		flags    uint8
		lower    interface{}
		upper    interface{}
		hasLower bool
		hasUpper bool
	)

	switch in := val.(type) {
	case types.RangeInt32:
		if in.Empty() {
			flags |= rangeEmpty
		} else {
			_, hasLower = in.Lower().Get()
			if !hasLower {
				flags |= rangeLBInf
			} else if in.IncLower() {
				flags |= rangeLBInc
			}

			_, hasUpper = in.Upper().Get()
			if !hasUpper {
				flags |= rangeUBInf
			} else if in.IncUpper() {
				flags |= rangeUBInc
			}

			lower = in.Lower()
			upper = in.Upper()
		}
	case types.RangeInt64:
		if in.Empty() {
			flags |= rangeEmpty
		} else {
			_, hasLower = in.Lower().Get()
			if !hasLower {
				flags |= rangeLBInf
			} else if in.IncLower() {
				flags |= rangeLBInc
			}

			_, hasUpper = in.Upper().Get()
			if !hasUpper {
				flags |= rangeUBInf
			} else if in.IncUpper() {
				flags |= rangeUBInc
			}

			lower = in.Lower()
			upper = in.Upper()
		}
	case types.RangeFloat32:
		if in.Empty() {
			flags |= rangeEmpty
		} else {
			_, hasLower = in.Lower().Get()
			if !hasLower {
				flags |= rangeLBInf
			} else if in.IncLower() {
				flags |= rangeLBInc
			}

			_, hasUpper = in.Upper().Get()
			if !hasUpper {
				flags |= rangeUBInf
			} else if in.IncUpper() {
				flags |= rangeUBInc
			}

			lower = in.Lower()
			upper = in.Upper()
		}
	case types.RangeFloat64:
		if in.Empty() {
			flags |= rangeEmpty
		} else {
			_, hasLower = in.Lower().Get()
			if !hasLower {
				flags |= rangeLBInf
			} else if in.IncLower() {
				flags |= rangeLBInc
			}

			_, hasUpper = in.Upper().Get()
			if !hasUpper {
				flags |= rangeUBInf
			} else if in.IncUpper() {
				flags |= rangeUBInc
			}

			lower = in.Lower()
			upper = in.Upper()
		}
	case types.RangeDateTime:
		if in.Empty() {
			flags |= rangeEmpty
		} else {
			_, hasLower = in.Lower().Get()
			if !hasLower {
				flags |= rangeLBInf
			} else if in.IncLower() {
				flags |= rangeLBInc
			}

			_, hasUpper = in.Upper().Get()
			if !hasUpper {
				flags |= rangeUBInf
			} else if in.IncUpper() {
				flags |= rangeUBInc
			}

			lower = in.Lower()
			upper = in.Upper()
		}
	case types.RangeLocalDateTime:
		if in.Empty() {
			flags |= rangeEmpty
		} else {
			_, hasLower = in.Lower().Get()
			if !hasLower {
				flags |= rangeLBInf
			} else if in.IncLower() {
				flags |= rangeLBInc
			}

			_, hasUpper = in.Upper().Get()
			if !hasUpper {
				flags |= rangeUBInf
			} else if in.IncUpper() {
				flags |= rangeUBInc
			}

			lower = in.Lower()
			upper = in.Upper()
		}
	case types.RangeLocalDate:
		if in.Empty() {
			flags |= rangeEmpty
		} else {
			_, hasLower = in.Lower().Get()
			if !hasLower {
				flags |= rangeLBInf
			} else if in.IncLower() {
				flags |= rangeLBInc
			}

			_, hasUpper = in.Upper().Get()
			if !hasUpper {
				flags |= rangeUBInf
			} else if in.IncUpper() {
				flags |= rangeUBInc
			}

			lower = in.Lower()
			upper = in.Upper()
		}
	default:
		return fmt.Errorf("invalid range type at %v: %T", path, val)
	}

	w.BeginBytes()
	w.PushUint8(flags)
	if hasLower {
		err := c.child.Encode(w, lower, path.AddField("lower"), false)
		if err != nil {
			return err
		}
	}

	if hasUpper {
		err := c.child.Encode(w, upper, path.AddField("upper"), false)
		if err != nil {
			return err
		}
	}
	w.EndBytes()
	return nil
}
