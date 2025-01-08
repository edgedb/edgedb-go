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
	"unsafe"

	"github.com/edgedb/edgedb-go/internal/buff"
	"github.com/edgedb/edgedb-go/internal/descriptor"
	types "github.com/edgedb/edgedb-go/internal/geltypes"
)

var (
	// NoOpDecoder is a noOpDecoder
	NoOpDecoder = noOpDecoder{}

	// NoOpEncoder is a noOpEncoder
	NoOpEncoder = noOpEncoder{}
)

// noOpDecoder decodes empty blocks i.e. does nothing.
//
//	There is one special type with type id of zero:
//	00000000-0000-0000-0000-000000000000.
//	The describe result of this type contains zero blocks.
//	It’s used when a statement returns no meaningful results,
//	e.g. the CREATE DATABASE example statement.
//
// https://www.edgedb.com/docs/internals/protocol/typedesc#type-descriptors
type noOpDecoder struct{}

func (c noOpDecoder) DescriptorID() types.UUID { return descriptor.IDZero }

func (c noOpDecoder) Decode(_ *buff.Reader, _ unsafe.Pointer) error {
	return nil
}

type noOpEncoder struct{}

func (c noOpEncoder) DescriptorID() types.UUID { return descriptor.IDZero }

func (c noOpEncoder) Encode(
	w *buff.Writer,
	_ interface{},
	_ Path,
	_ bool,
) error {
	w.PushUint32(0)
	return nil
}
