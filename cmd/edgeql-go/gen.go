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

package main

import (
	"bytes"
	"fmt"
	"sync"

	"github.com/edgedb/edgedb-go/internal/codecs"
	"github.com/edgedb/edgedb-go/internal/descriptor"
	types "github.com/edgedb/edgedb-go/internal/edgedbtypes"
)

type lookupKey struct {
	ID       types.UUID
	Required bool
}

// Type is a go type definition
type Type struct {
	definition string
	imports    []string
}

// Generator generates go code from EdgeDB types.
type Generator struct {
	mx             sync.RWMutex
	typeNameLookup map[lookupKey]Type
}

func (g *Generator) getType(
	desc descriptor.Descriptor,
	required bool,
) (Type, error) {
	key := lookupKey{desc.ID, required}

	g.mx.RLock()
	typ, ok := g.typeNameLookup[key]
	g.mx.RUnlock()

	if !ok {
		text, imports, err := g.generateType(desc, required)
		if err != nil {
			return Type{}, err
		}

		typ = Type{string(text), imports}

		g.mx.Lock()
		g.typeNameLookup[key] = typ
		g.mx.Unlock()
	}

	return typ, nil
}

func (g *Generator) generateType(
	desc descriptor.Descriptor,
	required bool,
) ([]byte, []string, error) {
	var (
		buf     bytes.Buffer
		err     error
		imports []string
	)

	switch desc.Type {
	case descriptor.Set, descriptor.Array:
		imports, err = g.generateSlice(&buf, desc)
	case descriptor.Object, descriptor.NamedTuple:
		imports, err = g.generateObject(&buf, desc, required)
	case descriptor.Tuple:
		imports, err = g.generateTuple(&buf, desc, required)
	case descriptor.BaseScalar, descriptor.Scalar, descriptor.Enum:
		imports, err = g.generateBaseScalar(&buf, desc, required)
	case descriptor.Range:
		err = g.generateRange(&buf, desc, required)
	default:
		err = fmt.Errorf(
			"generating type: unknown descriptor type %v",
			desc.Type,
		)
	}

	if err != nil {
		return nil, nil, err
	}

	return buf.Bytes(), imports, nil
}

func (g *Generator) generateRange(
	buf *bytes.Buffer,
	desc descriptor.Descriptor,
	required bool,
) error {
	buf.WriteString("edgedb.")
	if required {
		buf.WriteString("Optional")
	}
	buf.WriteString("Range")

	fieldDesc := desc.Fields[0].Desc
	switch fieldDesc.ID {
	case codecs.Int32ID:
		buf.WriteString("Int32")
	case codecs.Int64ID:
		buf.WriteString("Int64")
	case codecs.Float32ID:
		buf.WriteString("Float32")
	case codecs.Float64ID:
		buf.WriteString("Float64")
	case codecs.DateTimeID:
		buf.WriteString("DateTime")
	case codecs.LocalDTID:
		buf.WriteString("LocalDateTime")
	case codecs.LocalDateID:
		buf.WriteString("LocalDate")
	default:
		return fmt.Errorf(
			"generating range: unknown %v with id %v",
			fieldDesc.Type,
			fieldDesc.ID,
		)
	}
	return nil
}

func (g *Generator) generateSlice(
	buf *bytes.Buffer,
	desc descriptor.Descriptor,
) ([]string, error) {
	typ, err := g.getType(desc.Fields[0].Desc, desc.Fields[0].Required)
	if err != nil {
		return nil, err
	}

	fmt.Fprint(buf, typ.definition)
	return typ.imports, nil
}

func (g *Generator) generateObject(
	buf *bytes.Buffer,
	desc descriptor.Descriptor,
	required bool,
) ([]string, error) {
	var imports []string
	fmt.Fprintln(buf, `struct {`)

	for _, field := range desc.Fields {
		typ, err := g.getType(field.Desc, field.Required)
		if err != nil {
			return nil, err
		}

		fmt.Fprintf(
			buf,
			"%s %s `edgedb:\"%s\"`\n",
			field.Name,
			typ.definition,
			field.Name,
		)
		imports = append(imports, typ.imports...)
	}
	fmt.Fprint(buf, `}`)

	return imports, nil
}

func (g *Generator) generateTuple(
	buf *bytes.Buffer,
	desc descriptor.Descriptor,
	required bool,
) ([]string, error) {
	var imports []string
	fmt.Fprintln(buf, `struct {`)
	fmt.Fprintln(buf, "// descriptor", desc.ID)

	for _, field := range desc.Fields {
		typ, err := g.getType(field.Desc, field.Required)
		if err != nil {
			return nil, err
		}

		fmt.Fprintf(
			buf,
			"Field%s %s `edgedb:\"%s\"`\n",
			field.Name,
			typ.definition,
			field.Name,
		)
		imports = append(imports, typ.imports...)
	}
	fmt.Fprint(buf, `}`)

	return imports, nil
}

func (g *Generator) generateBaseScalar(
	buf *bytes.Buffer,
	desc descriptor.Descriptor,
	required bool,
) ([]string, error) {
	if desc.Type == descriptor.Scalar {
		desc = codecs.GetScalarDescriptor(desc)
	}

	if desc.Type == descriptor.Enum {
		if required {
			buf.WriteString("string")
		} else {
			buf.WriteString("edgedb.OptionalStr")
		}
		return nil, nil
	}

	var imports []string
	switch desc.ID {
	case codecs.UUIDID:
		if required {
			buf.WriteString("edgedb.UUID")
		} else {
			buf.WriteString("edgedb.OptionalUUID")
		}
	case codecs.StrID:
		if required {
			buf.WriteString("string")
		} else {
			buf.WriteString("edgedb.OptionalStr")
		}
	case codecs.BytesID, codecs.JSONID:
		if required {
			buf.WriteString("[]byte")
		} else {
			buf.WriteString("edgedb.OptionalBytes")
		}
	case codecs.Int16ID:
		if required {
			buf.WriteString("int16")
		} else {
			buf.WriteString("edgedb.OptionalInt16")
		}
	case codecs.Int32ID:
		if required {
			buf.WriteString("int32")
		} else {
			buf.WriteString("edgedb.OptionalInt32")
		}
	case codecs.Int64ID:
		if required {
			buf.WriteString("int64")
		} else {
			buf.WriteString("edgedb.OptionalInt64")
		}
	case codecs.Float32ID:
		if required {
			buf.WriteString("float32")
		} else {
			buf.WriteString("edgedb.OptionalFloat32")
		}
	case codecs.Float64ID:
		if required {
			buf.WriteString("float64")
		} else {
			buf.WriteString("edgedb.OptionalFloat64")
		}
	case codecs.BoolID:
		if required {
			buf.WriteString("bool")
		} else {
			buf.WriteString("edgedb.OptionalBool")
		}
	case codecs.DateTimeID:
		if required {
			imports = append(imports, "time")
			buf.WriteString("time.Time")
		} else {
			buf.WriteString("edgedb.OptionalDateTime")
		}
	case codecs.LocalDTID:
		if required {
			buf.WriteString("edgedb.LocalDateTime")
		} else {
			buf.WriteString("edgedb.OptionalLocalDateTime")
		}
	case codecs.LocalDateID:
		if required {
			buf.WriteString("edgedb.LocalDate")
		} else {
			buf.WriteString("edgedb.OptionalLocalDate")
		}
	case codecs.LocalTimeID:
		if required {
			buf.WriteString("edgedb.LocalTime")
		} else {
			buf.WriteString("edgedb.OptionalLocalTime")
		}
	case codecs.DurationID:
		if required {
			buf.WriteString("edgedb.Duration")
		} else {
			buf.WriteString("edgedb.OptionalDuration")
		}
	case codecs.BigIntID:
		if required {
			imports = append(imports, "math/big")
			buf.WriteString("*big.Int")
		} else {
			buf.WriteString("edgedb.OptionalBigInt")
		}
	case codecs.RelativeDurationID:
		if required {
			buf.WriteString("edgedb.RelativeDuration")
		} else {
			buf.WriteString("edgedb.OptionalRelativeDuration")
		}
	case codecs.DateDurationID:
		if required {
			buf.WriteString("edgedb.DateDuration")
		} else {
			buf.WriteString("edgedb.OptionalDateDuration")
		}
	case codecs.MemoryID:
		if required {
			buf.WriteString("edgedb.Memory")
		} else {
			buf.WriteString("edgedb.OptionalMemory")
		}
	}

	return imports, nil
}
