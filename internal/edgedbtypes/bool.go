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
	"encoding/json"
)

// OptionalBool is an optional bool. Optional types must be used for out
// parameters when a shape field is not required.
type OptionalBool struct {
	val   bool
	isSet bool
}

// Get returns the value and a boolean indicating if the value is present.
func (o *OptionalBool) Get() (bool, bool) { return o.val, o.isSet }

// Set sets the value.
func (o *OptionalBool) Set(val bool) {
	o.val = val
	o.isSet = true
}

// Unset marks the value as missing.
func (o *OptionalBool) Unset() {
	o.val = false
	o.isSet = false
}

func (o OptionalBool) MarshalJSON() ([]byte, error) {
	if o.isSet {
		return json.Marshal(o.val)
	}
	return json.Marshal(nil)
}

func (o *OptionalBool) UnmarshalJSON(bytes []byte) error {
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
