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

// Optional represents a shape field that is not required.
type Optional struct {
	isSet bool
}

// Missing returns true if the value is missing.
func (o *Optional) Missing() bool { return !o.isSet }

// SetMissing sets the structs missing status. true means missing and false
// means present.
func (o *Optional) SetMissing(missing bool) { o.isSet = !missing }

// OptionalInt16 is an optional int16. Optional types must be used for out
// parameters when a shape field is not required.
type OptionalInt16 struct {
	val   int16
	isSet bool
}

// Get returns the value and a boolean indicating if the value is present.
func (o *OptionalInt16) Get() (int16, bool) { return o.val, o.isSet }

// Set sets the value.
func (o *OptionalInt16) Set(val int16) {
	o.val = val
	o.isSet = true
}

// Unset marks the value as missing.
func (o *OptionalInt16) Unset() {
	o.val = 0
	o.isSet = false
}

// OptionalInt32 is an optional int32. Optional types must be used for out
// parameters when a shape field is not required.
type OptionalInt32 struct {
	val   int32
	isSet bool
}

// Get returns the value and a boolean indicating if the value is present.
func (o *OptionalInt32) Get() (int32, bool) { return o.val, o.isSet }

// Set sets the value.
func (o *OptionalInt32) Set(val int32) {
	o.val = val
	o.isSet = true
}

// Unset marks the value as missing.
func (o *OptionalInt32) Unset() {
	o.val = 0
	o.isSet = false
}

// OptionalInt64 is an optional int64. Optional types must be used for out
// parameters when a shape field is not required.
type OptionalInt64 struct {
	val   int64
	isSet bool
}

// Get returns the value and a boolean indicating if the value is present.
func (o *OptionalInt64) Get() (int64, bool) { return o.val, o.isSet }

// Set sets the value.
func (o *OptionalInt64) Set(val int64) *OptionalInt64 {
	o.val = val
	o.isSet = true
	return o
}

// Unset marks the value as missing.
func (o *OptionalInt64) Unset() *OptionalInt64 {
	o.val = 0
	o.isSet = false
	return o
}

// OptionalFloat32 is an optional float32. Optional types must be used for out
// parameters when a shape field is not required.
type OptionalFloat32 struct {
	val   float32
	isSet bool
}

// Get returns the value and a boolean indicating if the value is present.
func (o *OptionalFloat32) Get() (float32, bool) { return o.val, o.isSet }

// Set sets the value.
func (o *OptionalFloat32) Set(val float32) {
	o.val = val
	o.isSet = true
}

// Unset marks the value as missing.
func (o *OptionalFloat32) Unset() {
	o.val = 0
	o.isSet = false
}

// OptionalFloat64 is an optional float64. Optional types must be used for out
// parameters when a shape field is not required.
type OptionalFloat64 struct {
	val   float64
	isSet bool
}

// Get returns the value and a boolean indicating if the value is present.
func (o *OptionalFloat64) Get() (float64, bool) { return o.val, o.isSet }

// Set sets the value.
func (o *OptionalFloat64) Set(val float64) {
	o.val = val
	o.isSet = true
}

// Unset marks the value as missing.
func (o *OptionalFloat64) Unset() {
	o.val = 0
	o.isSet = false
}
