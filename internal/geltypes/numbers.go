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
)

// Optional represents a shape field that is not required.
// Optional is embedded in structs to make them optional. For example:
//
//	type User struct {
//	    gel.Optional
//	    Name string `gel:"name"`
//	}
type Optional struct {
	isSet bool
}

// Missing returns true if the value is missing.
func (o *Optional) Missing() bool { return !o.isSet }

// SetMissing sets the structs missing status. true means missing and false
// means present.
func (o *Optional) SetMissing(missing bool) { o.isSet = !missing }

// Unset marks the value as missing
func (o *Optional) Unset() { o.isSet = false }

// NewOptionalInt16 is a convenience function for creating an OptionalInt16
// with its value set to v.
func NewOptionalInt16(v int16) OptionalInt16 {
	o := OptionalInt16{}
	o.Set(v)
	return o
}

// OptionalInt16 is an optional int16. Optional types must be used for out
// parameters when a shape field is not required.
type OptionalInt16 struct {
	val   int16
	isSet bool
}

// Get returns the value and a boolean indicating if the value is present.
func (o OptionalInt16) Get() (int16, bool) { return o.val, o.isSet }

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

// MarshalJSON returns o marshaled as json.
func (o OptionalInt16) MarshalJSON() ([]byte, error) {
	if o.isSet {
		return json.Marshal(o.val)
	}
	return json.Marshal(nil)
}

// UnmarshalJSON unmarshals bytes into *o.
func (o *OptionalInt16) UnmarshalJSON(bytes []byte) error {
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

// NewOptionalInt32 is a convenience function for creating an OptionalInt32
// with its value set to v.
func NewOptionalInt32(v int32) OptionalInt32 {
	o := OptionalInt32{}
	o.Set(v)
	return o
}

// OptionalInt32 is an optional int32. Optional types must be used for out
// parameters when a shape field is not required.
type OptionalInt32 struct {
	val   int32
	isSet bool
}

// Get returns the value and a boolean indicating if the value is present.
func (o OptionalInt32) Get() (int32, bool) { return o.val, o.isSet }

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

// MarshalJSON returns o marshaled as json.
func (o OptionalInt32) MarshalJSON() ([]byte, error) {
	if o.isSet {
		return json.Marshal(o.val)
	}
	return json.Marshal(nil)
}

// UnmarshalJSON unmarshals bytes into *o.
func (o *OptionalInt32) UnmarshalJSON(bytes []byte) error {
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

// NewOptionalInt64 is a convenience function for creating an OptionalInt64
// with its value set to v.
func NewOptionalInt64(v int64) OptionalInt64 {
	o := OptionalInt64{}
	o.Set(v)
	return o
}

// OptionalInt64 is an optional int64. Optional types must be used for out
// parameters when a shape field is not required.
type OptionalInt64 struct {
	val   int64
	isSet bool
}

// Get returns the value and a boolean indicating if the value is present.
func (o OptionalInt64) Get() (int64, bool) { return o.val, o.isSet }

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

// MarshalJSON returns o marshaled as json.
func (o OptionalInt64) MarshalJSON() ([]byte, error) {
	if o.isSet {
		return json.Marshal(o.val)
	}
	return json.Marshal(nil)
}

// UnmarshalJSON unmarshals bytes into *o.
func (o *OptionalInt64) UnmarshalJSON(bytes []byte) error {
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

// NewOptionalFloat32 is a convenience function for creating an OptionalFloat32
// with its value set to v.
func NewOptionalFloat32(v float32) OptionalFloat32 {
	o := OptionalFloat32{}
	o.Set(v)
	return o
}

// OptionalFloat32 is an optional float32. Optional types must be used for out
// parameters when a shape field is not required.
type OptionalFloat32 struct {
	val   float32
	isSet bool
}

// Get returns the value and a boolean indicating if the value is present.
func (o OptionalFloat32) Get() (float32, bool) { return o.val, o.isSet }

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

// MarshalJSON returns o marshaled as json.
func (o OptionalFloat32) MarshalJSON() ([]byte, error) {
	if o.isSet {
		return json.Marshal(o.val)
	}
	return json.Marshal(nil)
}

// UnmarshalJSON unmarshals bytes into *o.
func (o *OptionalFloat32) UnmarshalJSON(bytes []byte) error {
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

// NewOptionalFloat64 is a convenience function for creating an OptionalFloat64
// with its value set to v.
func NewOptionalFloat64(v float64) OptionalFloat64 {
	o := OptionalFloat64{}
	o.Set(v)
	return o
}

// OptionalFloat64 is an optional float64. Optional types must be used for out
// parameters when a shape field is not required.
type OptionalFloat64 struct {
	val   float64
	isSet bool
}

// Get returns the value and a boolean indicating if the value is present.
func (o OptionalFloat64) Get() (float64, bool) { return o.val, o.isSet }

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

// MarshalJSON returns o marshaled as json.
func (o OptionalFloat64) MarshalJSON() ([]byte, error) {
	if o.isSet {
		return json.Marshal(o.val)
	}
	return json.Marshal(nil)
}

// UnmarshalJSON unmarshals bytes into *o.
func (o *OptionalFloat64) UnmarshalJSON(bytes []byte) error {
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
