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

// OptionalStr is an optional string. Optional types must be used for out
// parameters when a shape field is not required.
type OptionalStr struct {
	val   string
	isSet bool
}

// Get returns the value and a boolean indicating if the value is present.
func (o *OptionalStr) Get() (string, bool) { return o.val, o.isSet }

// Set sets the value.
func (o *OptionalStr) Set(val string) {
	o.val = val
	o.isSet = true
}

// Unset marks the value as missing.
func (o *OptionalStr) Unset() {
	o.val = ""
	o.isSet = false
}
