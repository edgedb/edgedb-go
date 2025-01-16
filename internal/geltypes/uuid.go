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

package geltypes

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// ParseUUID parses s into a UUID or returns an error.
func ParseUUID(s string) (UUID, error) {
	s = strings.ReplaceAll(s, "-", "")
	if len(s) != 32 {
		return UUID{}, errMalformedUUID
	}

	var tmp UUID
	for i := 0; i < 16; i++ {
		val, err := strconv.ParseUint(s[:2], 16, 8)
		if err != nil {
			return UUID{}, errMalformedUUID
		}

		tmp[i] = uint8(val)
		s = s[2:]
	}

	return tmp, nil
}

// UUID is a universally unique identifier
// https://www.edgedb.com/docs/stdlib/uuid
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

var errMalformedUUID = errors.New("malformed gel.UUID")

// UnmarshalText unmarshals the id from a string.
func (id *UUID) UnmarshalText(b []byte) error {
	tmp, err := ParseUUID(string(b))
	if err != nil {
		return err
	}

	*id = tmp
	return nil
}

// NewOptionalUUID is a convenience function for creating an OptionalUUID with
// its value set to v.
func NewOptionalUUID(v UUID) OptionalUUID {
	o := OptionalUUID{}
	o.Set(v)
	return o
}

// OptionalUUID is an optional UUID. Optional types must be used for out
// parameters when a shape field is not required.
type OptionalUUID struct {
	val   UUID
	isSet bool
}

// Get returns the value and a boolean indicating if the value is present.
func (o OptionalUUID) Get() (UUID, bool) { return o.val, o.isSet }

// Set sets the value.
func (o *OptionalUUID) Set(val UUID) {
	o.val = val
	o.isSet = true
}

// Unset marks the value as missing.
func (o *OptionalUUID) Unset() {
	o.val = UUID{}
	o.isSet = false
}

// MarshalJSON returns o marshaled as json.
func (o OptionalUUID) MarshalJSON() ([]byte, error) {
	if o.isSet {
		return json.Marshal(o.val)
	}
	return json.Marshal(nil)
}

// UnmarshalJSON unmarshals bytes into *o
func (o *OptionalUUID) UnmarshalJSON(bytes []byte) error {
	if bytes[0] == 0x6e { // null
		o.Unset()
		return nil
	}

	if err := json.Unmarshal(bytes, &o.val); err != nil {
		return err
	}
	o.isSet = true

	return nil
}
