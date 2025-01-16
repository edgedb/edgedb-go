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
	"fmt"
	"strconv"
	"strings"
)

const (
	petabyte = 1_024 * 1_024 * 1_024 * 1_024 * 1_024
	terabyte = 1_024 * 1_024 * 1_024 * 1_024
	gigabyte = 1_024 * 1_024 * 1_024
	megabyte = 1_024 * 1_024
	kilobyte = 1_024
)

// Memory represents memory in bytes.
type Memory int64

func (m Memory) String() string {
	switch {
	case m == 0:
		return "0B"
	case m%petabyte == 0:
		return fmt.Sprintf("%vPiB", int64(m)/petabyte)
	case m%terabyte == 0:
		return fmt.Sprintf("%vTiB", int64(m)/terabyte)
	case m%gigabyte == 0:
		return fmt.Sprintf("%vGiB", int64(m)/gigabyte)
	case m%megabyte == 0:
		return fmt.Sprintf("%vMiB", int64(m)/megabyte)
	case m%kilobyte == 0:
		return fmt.Sprintf("%vKiB", int64(m)/kilobyte)
	default:
		return fmt.Sprintf("%vB", int64(m))
	}
}

// MarshalText returns m marshaled as text.
func (m Memory) MarshalText() ([]byte, error) {
	return []byte(m.String()), nil
}

// UnmarshalText unmarshals bytes into *m.
func (m *Memory) UnmarshalText(b []byte) error {
	s := string(b)
	suffixLen := 3
	var multiplier int64 = 1
	switch {
	case strings.HasSuffix(s, "PiB"):
		multiplier = petabyte
	case strings.HasSuffix(s, "TiB"):
		multiplier = terabyte
	case strings.HasSuffix(s, "GiB"):
		multiplier = gigabyte
	case strings.HasSuffix(s, "MiB"):
		multiplier = megabyte
	case strings.HasSuffix(s, "KiB"):
		multiplier = kilobyte
	case strings.HasSuffix(s, "B"):
		suffixLen = 1
	default:
		return fmt.Errorf("malformed gel.Memory: %q", s)
	}

	i, err := strconv.ParseInt(s[:len(s)-suffixLen], 10, 64)
	if err != nil {
		return fmt.Errorf("malformed gel.Memory: %w", err)
	}

	*m = Memory(i * multiplier)
	return nil
}

// NewOptionalMemory is a convenience function for creating an
// OptionalMemory with its value set to v.
func NewOptionalMemory(v Memory) OptionalMemory {
	o := OptionalMemory{}
	o.Set(v)
	return o
}

// OptionalMemory is an optional Memory. Optional types must be used for
// out parameters when a shape field is not required.
type OptionalMemory struct {
	val   Memory
	isSet bool
}

// Get returns the value and a boolean indicating if the value is present.
func (o OptionalMemory) Get() (Memory, bool) { return o.val, o.isSet }

// Set sets the value.
func (o *OptionalMemory) Set(val Memory) {
	o.val = val
	o.isSet = true
}

// Unset marks the value as missing.
func (o *OptionalMemory) Unset() {
	o.val = 0
	o.isSet = false
}

// MarshalJSON returns o marshaled as json.
func (o OptionalMemory) MarshalJSON() ([]byte, error) {
	if o.isSet {
		return json.Marshal(o.val)
	}
	return json.Marshal(nil)
}

// UnmarshalJSON unmarshals bytes into *o.
func (o *OptionalMemory) UnmarshalJSON(bytes []byte) error {
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
