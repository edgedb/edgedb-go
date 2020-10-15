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

package types

import "fmt"

// UUID a universally unique identifier
// https://www.edgedb.com/docs/datamodel/scalars/uuid#type::std::uuid
type UUID [16]byte

func (id UUID) String() string {
	return fmt.Sprintf(
		"%x-%x-%x-%x-%x",
		id[0:4],
		id[4:6],
		id[6:8],
		id[8:10],
		id[10:16],
	)
}

// Set https://www.edgedb.com/docs/edgeql/overview#everything-is-a-set
type Set []interface{}

// Object https://www.edgedb.com/docs/datamodel/objects#type::std::Object
type Object map[string]interface{}

// Array https://www.edgedb.com/docs/datamodel/colltypes#type::std::array
type Array []interface{}

// Tuple https://www.edgedb.com/docs/datamodel/colltypes#type::std::tuple
type Tuple []interface{}

// NamedTuple ?
type NamedTuple map[string]interface{}
