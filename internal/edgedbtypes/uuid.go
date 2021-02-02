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

package edgedbtypes

import (
	"fmt"
	"strconv"
	"strings"
)

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

// MarshalText returns the id as a byte string.
func (id UUID) MarshalText() ([]byte, error) {
	return []byte(id.String()), nil
}

// UnmarshalText unmarshals the id from a string.
func (id *UUID) UnmarshalText(b []byte) error {
	s := string(b)
	s = strings.Replace(s, "-", "", 4)

	var tmp UUID
	for i := 0; i < 16; i++ {
		val, err := strconv.ParseUint(s[:2], 16, 8)
		if err != nil {
			return err
		}

		tmp[i] = uint8(val)
		s = s[2:]
	}

	*id = tmp
	return nil
}
