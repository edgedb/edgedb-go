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

import "encoding/json"

type emptyRangeJSON struct {
	Empty bool `json:"empty"`
}

// NewRangeInt32 creates a new RangeInt32 value.
func NewRangeInt32(
	lower, upper OptionalInt32,
	incLower, incUpper bool,
) RangeInt32 {
	if lower.isSet && !incLower {
		lower.val++
		incLower = true
	} else if !lower.isSet {
		incLower = false
	}

	if upper.isSet && incUpper {
		upper.val++
		incUpper = false
	} else if !upper.isSet {
		incUpper = false
	}

	if lower.isSet && upper.isSet && lower.val == upper.val {
		return RangeInt32{empty: true}
	}

	return RangeInt32{
		lower:    lower,
		upper:    upper,
		incLower: incLower,
		incUpper: incUpper,
	}
}

// RangeInt32 is an interval of int32 values.
type RangeInt32 struct {
	lower    OptionalInt32 `gel:"lower"`
	upper    OptionalInt32 `gel:"upper"`
	incLower bool          `gel:"inc_lower"`
	incUpper bool          `gel:"inc_upper"`
	empty    bool          `gel:"empty"`
}

// Lower returns the lower bound.
func (r RangeInt32) Lower() OptionalInt32 { return r.lower }

// Upper returns the upper bound.
func (r RangeInt32) Upper() OptionalInt32 { return r.upper }

// IncLower returns true if the lower bound is inclusive.
func (r RangeInt32) IncLower() bool { return r.incLower }

// IncUpper returns true if the upper bound is inclusive.
func (r RangeInt32) IncUpper() bool { return r.incUpper }

// Empty returns true if the range is empty.
func (r RangeInt32) Empty() bool { return r.empty }

type rangeInt32JSON struct {
	Lower    OptionalInt32 `json:"lower"`
	Upper    OptionalInt32 `json:"upper"`
	IncLower bool          `json:"inc_lower"`
	IncUpper bool          `json:"inc_upper"`
}

// MarshalJSON returns r marshaled as json.
func (r RangeInt32) MarshalJSON() ([]byte, error) {
	if r.empty {
		return []byte(`{"empty":true}`), nil
	}

	return json.Marshal(rangeInt32JSON{
		Lower:    r.lower,
		Upper:    r.upper,
		IncLower: r.incLower,
		IncUpper: r.incUpper,
	})
}

// UnmarshalJSON unmarshals bytes into *r.
func (r *RangeInt32) UnmarshalJSON(data []byte) error {
	var empty emptyRangeJSON
	err := json.Unmarshal(data, &empty)
	if err != nil {
		return err
	}

	if empty.Empty {
		r.empty = true
		return nil
	}

	var decoded rangeInt32JSON
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		return err
	}

	r.lower = decoded.Lower
	r.upper = decoded.Upper
	r.incLower = decoded.IncLower
	r.incUpper = decoded.IncUpper
	return nil
}

// NewOptionalRangeInt32 is a convenience function for creating an
// OptionalRangeInt32 with its value set to v.
func NewOptionalRangeInt32(v RangeInt32) OptionalRangeInt32 {
	o := OptionalRangeInt32{}
	o.Set(v)
	return o
}

// OptionalRangeInt32 is an optional RangeInt32. Optional types must be used
// for out parameters when a shape field is not required.
type OptionalRangeInt32 struct {
	val   RangeInt32
	isSet bool
}

// Get returns the value and a boolean indicating if the value is present.
func (o OptionalRangeInt32) Get() (RangeInt32, bool) {
	return o.val, o.isSet
}

// Set sets the value.
func (o *OptionalRangeInt32) Set(val RangeInt32) {
	o.val = val
	o.isSet = true
}

// Unset marks the value as missing.
func (o *OptionalRangeInt32) Unset() {
	o.val = RangeInt32{}
	o.isSet = false
}

// MarshalJSON returns o marshaled as json.
func (o OptionalRangeInt32) MarshalJSON() ([]byte, error) {
	if o.isSet {
		return json.Marshal(o.val)
	}
	return json.Marshal(nil)
}

// UnmarshalJSON unmarshals bytes into *o.
func (o *OptionalRangeInt32) UnmarshalJSON(bytes []byte) error {
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

// NewRangeInt64 creates a new RangeInt64 value.
func NewRangeInt64(
	lower, upper OptionalInt64,
	incLower, incUpper bool,
) RangeInt64 {
	if lower.isSet && !incLower {
		lower.val++
		incLower = true
	} else if !lower.isSet {
		incLower = false
	}

	if upper.isSet && incUpper {
		upper.val++
		incUpper = false
	} else if !upper.isSet {
		incUpper = false
	}

	if lower.isSet && upper.isSet && lower.val == upper.val {
		return RangeInt64{empty: true}
	}

	return RangeInt64{
		lower:    lower,
		upper:    upper,
		incLower: incLower,
		incUpper: incUpper,
	}
}

// RangeInt64 is an interval of int64 values.
type RangeInt64 struct {
	lower    OptionalInt64 `gel:"lower"`
	upper    OptionalInt64 `gel:"upper"`
	incLower bool          `gel:"inc_lower"`
	incUpper bool          `gel:"inc_upper"`
	empty    bool          `gel:"empty"`
}

// Lower returns the lower bound.
func (r RangeInt64) Lower() OptionalInt64 { return r.lower }

// Upper returns the upper bound.
func (r RangeInt64) Upper() OptionalInt64 { return r.upper }

// IncLower returns true if the lower bound is inclusive.
func (r RangeInt64) IncLower() bool { return r.incLower }

// IncUpper returns true if the upper bound is inclusive.
func (r RangeInt64) IncUpper() bool { return r.incUpper }

// Empty returns true if the range is empty.
func (r RangeInt64) Empty() bool { return r.empty }

type rangeInt64JSON struct {
	Lower    OptionalInt64 `json:"lower"`
	Upper    OptionalInt64 `json:"upper"`
	IncLower bool          `json:"inc_lower"`
	IncUpper bool          `json:"inc_upper"`
}

// MarshalJSON returns r marshaled as json.
func (r RangeInt64) MarshalJSON() ([]byte, error) {
	if r.empty {
		return []byte(`{"empty":true}`), nil
	}

	return json.Marshal(rangeInt64JSON{
		Lower:    r.lower,
		Upper:    r.upper,
		IncLower: r.incLower,
		IncUpper: r.incUpper,
	})
}

// UnmarshalJSON unmarshals bytes into *r.
func (r *RangeInt64) UnmarshalJSON(data []byte) error {
	var empty emptyRangeJSON
	err := json.Unmarshal(data, &empty)
	if err != nil {
		return err
	}

	if empty.Empty {
		r.empty = true
		return nil
	}

	var decoded rangeInt64JSON
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		return err
	}

	r.lower = decoded.Lower
	r.upper = decoded.Upper
	r.incLower = decoded.IncLower
	r.incUpper = decoded.IncUpper
	return nil
}

// NewOptionalRangeInt64 is a convenience function for creating an
// OptionalRangeInt64 with its value set to v.
func NewOptionalRangeInt64(v RangeInt64) OptionalRangeInt64 {
	o := OptionalRangeInt64{}
	o.Set(v)
	return o
}

// OptionalRangeInt64 is an optional RangeInt64. Optional
// types must be used for out parameters when a shape field is not required.
type OptionalRangeInt64 struct {
	val   RangeInt64
	isSet bool
}

// Get returns the value and a boolean indicating if the value is present.
func (o OptionalRangeInt64) Get() (RangeInt64, bool) {
	return o.val, o.isSet
}

// Set sets the value.
func (o *OptionalRangeInt64) Set(val RangeInt64) {
	o.val = val
	o.isSet = true
}

// Unset marks the value as missing.
func (o *OptionalRangeInt64) Unset() {
	o.val = RangeInt64{}
	o.isSet = false
}

// MarshalJSON returns o marshaled as json.
func (o OptionalRangeInt64) MarshalJSON() ([]byte, error) {
	if o.isSet {
		return json.Marshal(o.val)
	}
	return json.Marshal(nil)
}

// UnmarshalJSON unmarshals bytes into *o.
func (o *OptionalRangeInt64) UnmarshalJSON(bytes []byte) error {
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

// NewRangeFloat32 creates a new RangeFloat32 value.
func NewRangeFloat32(
	lower, upper OptionalFloat32,
	incLower, incUpper bool,
) RangeFloat32 {
	if !lower.isSet {
		incLower = false
	}

	if !upper.isSet {
		incUpper = false
	}

	if lower.isSet &&
		upper.isSet &&
		lower.val == upper.val &&
		(!incLower || !incUpper) {
		return RangeFloat32{empty: true}
	}

	return RangeFloat32{
		lower:    lower,
		upper:    upper,
		incLower: incLower,
		incUpper: incUpper,
	}
}

// RangeFloat32 is an interval of float32 values.
type RangeFloat32 struct {
	lower    OptionalFloat32 `gel:"lower"`
	upper    OptionalFloat32 `gel:"upper"`
	incLower bool            `gel:"inc_lower"`
	incUpper bool            `gel:"inc_upper"`
	empty    bool            `gel:"empty"`
}

// Lower returns the lower bound.
func (r RangeFloat32) Lower() OptionalFloat32 { return r.lower }

// Upper returns the upper bound.
func (r RangeFloat32) Upper() OptionalFloat32 { return r.upper }

// IncLower returns true if the lower bound is inclusive.
func (r RangeFloat32) IncLower() bool { return r.incLower }

// IncUpper returns true if the upper bound is inclusive.
func (r RangeFloat32) IncUpper() bool { return r.incUpper }

// Empty returns true if the range is empty.
func (r RangeFloat32) Empty() bool { return r.empty }

type rangeFloat32JSON struct {
	Lower    OptionalFloat32 `json:"lower"`
	Upper    OptionalFloat32 `json:"upper"`
	IncLower bool            `json:"inc_lower"`
	IncUpper bool            `json:"inc_upper"`
}

// MarshalJSON returns r marshaled as json.
func (r RangeFloat32) MarshalJSON() ([]byte, error) {
	if r.empty {
		return []byte(`{"empty":true}`), nil
	}

	return json.Marshal(rangeFloat32JSON{
		Lower:    r.lower,
		Upper:    r.upper,
		IncLower: r.incLower,
		IncUpper: r.incUpper,
	})
}

// UnmarshalJSON unmarshals bytes into *r.
func (r *RangeFloat32) UnmarshalJSON(data []byte) error {
	var empty emptyRangeJSON
	err := json.Unmarshal(data, &empty)
	if err != nil {
		return err
	}

	if empty.Empty {
		r.empty = true
		return nil
	}

	var decoded rangeFloat32JSON
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		return err
	}

	r.lower = decoded.Lower
	r.upper = decoded.Upper
	r.incLower = decoded.IncLower
	r.incUpper = decoded.IncUpper
	return nil
}

// OptionalRangeFloat32 is an optional RangeFloat32. Optional
// types must be used for out parameters when a shape field is not required.
type OptionalRangeFloat32 struct {
	val   RangeFloat32
	isSet bool
}

// NewOptionalRangeFloat32 is a convenience function for creating an
// OptionalRangeFloat32 with its value set to v.
func NewOptionalRangeFloat32(v RangeFloat32) OptionalRangeFloat32 {
	o := OptionalRangeFloat32{}
	o.Set(v)
	return o
}

// Get returns the value and a boolean indicating if the value is present.
func (o OptionalRangeFloat32) Get() (RangeFloat32, bool) {
	return o.val, o.isSet
}

// Set sets the value.
func (o *OptionalRangeFloat32) Set(val RangeFloat32) {
	o.val = val
	o.isSet = true
}

// Unset marks the value as missing.
func (o *OptionalRangeFloat32) Unset() {
	o.val = RangeFloat32{}
	o.isSet = false
}

// MarshalJSON returns o marshaled as json.
func (o OptionalRangeFloat32) MarshalJSON() ([]byte, error) {
	if o.isSet {
		return json.Marshal(o.val)
	}
	return json.Marshal(nil)
}

// UnmarshalJSON unmarshals bytes into *o.
func (o *OptionalRangeFloat32) UnmarshalJSON(bytes []byte) error {
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

// NewRangeFloat64 creates a new RangeFloat64 value.
func NewRangeFloat64(
	lower, upper OptionalFloat64,
	incLower, incUpper bool,
) RangeFloat64 {
	if !lower.isSet {
		incLower = false
	}

	if !upper.isSet {
		incUpper = false
	}

	if lower.isSet &&
		upper.isSet &&
		lower.val == upper.val &&
		(!incLower || !incUpper) {
		return RangeFloat64{empty: true}
	}

	return RangeFloat64{
		lower:    lower,
		upper:    upper,
		incLower: incLower,
		incUpper: incUpper,
	}
}

// RangeFloat64 is an interval of float64 values.
type RangeFloat64 struct {
	lower    OptionalFloat64 `gel:"lower"`
	upper    OptionalFloat64 `gel:"upper"`
	incLower bool            `gel:"inc_lower"`
	incUpper bool            `gel:"inc_upper"`
	empty    bool            `gel:"empty"`
}

// Lower returns the lower bound.
func (r RangeFloat64) Lower() OptionalFloat64 { return r.lower }

// Upper returns the upper bound.
func (r RangeFloat64) Upper() OptionalFloat64 { return r.upper }

// IncLower returns true if the lower bound is inclusive.
func (r RangeFloat64) IncLower() bool { return r.incLower }

// IncUpper returns true if the upper bound is inclusive.
func (r RangeFloat64) IncUpper() bool { return r.incUpper }

// Empty returns true if the range is empty.
func (r RangeFloat64) Empty() bool { return r.empty }

type rangeFloat64JSON struct {
	Lower    OptionalFloat64 `json:"lower"`
	Upper    OptionalFloat64 `json:"upper"`
	IncLower bool            `json:"inc_lower"`
	IncUpper bool            `json:"inc_upper"`
}

// MarshalJSON returns r marshaled as json.
func (r RangeFloat64) MarshalJSON() ([]byte, error) {
	if r.empty {
		return []byte(`{"empty":true}`), nil
	}

	return json.Marshal(rangeFloat64JSON{
		Lower:    r.lower,
		Upper:    r.upper,
		IncLower: r.incLower,
		IncUpper: r.incUpper,
	})
}

// UnmarshalJSON unmarshals bytes into *r.
func (r *RangeFloat64) UnmarshalJSON(data []byte) error {
	var empty emptyRangeJSON
	err := json.Unmarshal(data, &empty)
	if err != nil {
		return err
	}

	if empty.Empty {
		r.empty = true
		return nil
	}

	var decoded rangeFloat64JSON
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		return err
	}

	r.lower = decoded.Lower
	r.upper = decoded.Upper
	r.incLower = decoded.IncLower
	r.incUpper = decoded.IncUpper
	return nil
}

// NewOptionalRangeFloat64 is a convenience function for creating an
// OptionalRangeFloat64 with its value set to v.
func NewOptionalRangeFloat64(v RangeFloat64) OptionalRangeFloat64 {
	o := OptionalRangeFloat64{}
	o.Set(v)
	return o
}

// OptionalRangeFloat64 is an optional RangeFloat64. Optional
// types must be used for out parameters when a shape field is not required.
type OptionalRangeFloat64 struct {
	val   RangeFloat64
	isSet bool
}

// Get returns the value and a boolean indicating if the value is present.
func (o OptionalRangeFloat64) Get() (RangeFloat64, bool) {
	return o.val, o.isSet
}

// Set sets the value.
func (o *OptionalRangeFloat64) Set(val RangeFloat64) {
	o.val = val
	o.isSet = true
}

// Unset marks the value as missing.
func (o *OptionalRangeFloat64) Unset() {
	o.val = RangeFloat64{}
	o.isSet = false
}

// MarshalJSON returns o marshaled as json.
func (o OptionalRangeFloat64) MarshalJSON() ([]byte, error) {
	if o.isSet {
		return json.Marshal(o.val)
	}
	return json.Marshal(nil)
}

// UnmarshalJSON unmarshals bytes into *o.
func (o *OptionalRangeFloat64) UnmarshalJSON(bytes []byte) error {
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

// NewRangeDateTime creates a new RangeDateTime value.
func NewRangeDateTime(
	lower, upper OptionalDateTime,
	incLower, incUpper bool,
) RangeDateTime {
	if !lower.isSet {
		incLower = false
	}

	if !upper.isSet {
		incUpper = false
	}

	if lower.isSet &&
		upper.isSet &&
		lower.val == upper.val &&
		(!incLower || !incUpper) {
		return RangeDateTime{empty: true}
	}

	return RangeDateTime{
		lower:    lower,
		upper:    upper,
		incLower: incLower,
		incUpper: incUpper,
	}
}

// RangeDateTime is an interval of time.Time values.
type RangeDateTime struct {
	lower    OptionalDateTime `gel:"lower"`
	upper    OptionalDateTime `gel:"upper"`
	incLower bool             `gel:"inc_lower"`
	incUpper bool             `gel:"inc_upper"`
	empty    bool             `gel:"empty"`
}

// Lower returns the lower bound.
func (r RangeDateTime) Lower() OptionalDateTime { return r.lower }

// Upper returns the upper bound.
func (r RangeDateTime) Upper() OptionalDateTime { return r.upper }

// IncLower returns true if the lower bound is inclusive.
func (r RangeDateTime) IncLower() bool { return r.incLower }

// IncUpper returns true if the upper bound is inclusive.
func (r RangeDateTime) IncUpper() bool { return r.incUpper }

// Empty returns true if the range is empty.
func (r RangeDateTime) Empty() bool { return r.empty }

type rangeDateTimeJSON struct {
	Lower    OptionalDateTime `json:"lower"`
	Upper    OptionalDateTime `json:"upper"`
	IncLower bool             `json:"inc_lower"`
	IncUpper bool             `json:"inc_upper"`
}

// MarshalJSON returns r marshaled as json.
func (r RangeDateTime) MarshalJSON() ([]byte, error) {
	if r.empty {
		return []byte(`{"empty":true}`), nil
	}

	return json.Marshal(rangeDateTimeJSON{
		Lower:    r.lower,
		Upper:    r.upper,
		IncLower: r.incLower,
		IncUpper: r.incUpper,
	})
}

// UnmarshalJSON unmarshals bytes into *r.
func (r *RangeDateTime) UnmarshalJSON(data []byte) error {
	var empty emptyRangeJSON
	err := json.Unmarshal(data, &empty)
	if err != nil {
		return err
	}

	if empty.Empty {
		r.empty = true
		return nil
	}

	var decoded rangeDateTimeJSON
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		return err
	}

	r.lower = decoded.Lower
	r.upper = decoded.Upper
	r.incLower = decoded.IncLower
	r.incUpper = decoded.IncUpper
	return nil
}

// NewOptionalRangeDateTime is a convenience function for creating an
// OptionalRangeDateTime with its value set to v.
func NewOptionalRangeDateTime(v RangeDateTime) OptionalRangeDateTime {
	o := OptionalRangeDateTime{}
	o.Set(v)
	return o
}

// OptionalRangeDateTime is an optional RangeDateTime. Optional
// types must be used for out parameters when a shape field is not required.
type OptionalRangeDateTime struct {
	val   RangeDateTime
	isSet bool
}

// Get returns the value and a boolean indicating if the value is present.
func (o *OptionalRangeDateTime) Get() (RangeDateTime, bool) {
	return o.val, o.isSet
}

// Set sets the value.
func (o *OptionalRangeDateTime) Set(val RangeDateTime) {
	o.val = val
	o.isSet = true
}

// Unset marks the value as missing.
func (o *OptionalRangeDateTime) Unset() {
	o.val = RangeDateTime{}
	o.isSet = false
}

// MarshalJSON returns o marshaled as json.
func (o *OptionalRangeDateTime) MarshalJSON() ([]byte, error) {
	if o.isSet {
		return json.Marshal(o.val)
	}
	return json.Marshal(nil)
}

// UnmarshalJSON unmarshals bytes into *o.
func (o *OptionalRangeDateTime) UnmarshalJSON(bytes []byte) error {
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

// NewRangeLocalDateTime creates a new RangeLocalDateTime value.
func NewRangeLocalDateTime(
	lower, upper OptionalLocalDateTime,
	incLower, incUpper bool,
) RangeLocalDateTime {
	if !lower.isSet {
		incLower = false
	}

	if !upper.isSet {
		incUpper = false
	}

	if lower.isSet &&
		upper.isSet &&
		lower.val == upper.val &&
		(!incLower || !incUpper) {
		return RangeLocalDateTime{empty: true}
	}

	return RangeLocalDateTime{
		lower:    lower,
		upper:    upper,
		incLower: incLower,
		incUpper: incUpper,
	}
}

// RangeLocalDateTime is an interval of LocalDateTime values.
type RangeLocalDateTime struct {
	lower    OptionalLocalDateTime `gel:"lower"`
	upper    OptionalLocalDateTime `gel:"upper"`
	incLower bool                  `gel:"inc_lower"`
	incUpper bool                  `gel:"inc_upper"`
	empty    bool                  `gel:"empty"`
}

// Lower returns the lower bound.
func (r RangeLocalDateTime) Lower() OptionalLocalDateTime { return r.lower }

// Upper returns the upper bound.
func (r RangeLocalDateTime) Upper() OptionalLocalDateTime { return r.upper }

// IncLower returns true if the lower bound is inclusive.
func (r RangeLocalDateTime) IncLower() bool { return r.incLower }

// IncUpper returns true if the upper bound is inclusive.
func (r RangeLocalDateTime) IncUpper() bool { return r.incUpper }

// Empty returns true if the range is empty.
func (r RangeLocalDateTime) Empty() bool { return r.empty }

type rangeLocalDateTimeJSON struct {
	Lower    OptionalLocalDateTime `json:"lower"`
	Upper    OptionalLocalDateTime `json:"upper"`
	IncLower bool                  `json:"inc_lower"`
	IncUpper bool                  `json:"inc_upper"`
}

// MarshalJSON returns r marshaled as json.
func (r RangeLocalDateTime) MarshalJSON() ([]byte, error) {
	if r.empty {
		return []byte(`{"empty":true}`), nil
	}

	return json.Marshal(rangeLocalDateTimeJSON{
		Lower:    r.lower,
		Upper:    r.upper,
		IncLower: r.incLower,
		IncUpper: r.incUpper,
	})
}

// UnmarshalJSON unmarshals bytes into *r.
func (r *RangeLocalDateTime) UnmarshalJSON(data []byte) error {
	var empty emptyRangeJSON
	err := json.Unmarshal(data, &empty)
	if err != nil {
		return err
	}

	if empty.Empty {
		r.empty = true
		return nil
	}

	var decoded rangeLocalDateTimeJSON
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		return err
	}

	r.lower = decoded.Lower
	r.upper = decoded.Upper
	r.incLower = decoded.IncLower
	r.incUpper = decoded.IncUpper
	return nil
}

// NewOptionalRangeLocalDateTime is a convenience function for creating an
// OptionalRangeLocalDateTime with its value set to v.
func NewOptionalRangeLocalDateTime(
	v RangeLocalDateTime,
) OptionalRangeLocalDateTime {
	o := OptionalRangeLocalDateTime{}
	o.Set(v)
	return o
}

// OptionalRangeLocalDateTime is an optional RangeLocalDateTime. Optional
// types must be used for out parameters when a shape field is not required.
type OptionalRangeLocalDateTime struct {
	val   RangeLocalDateTime
	isSet bool
}

// Get returns the value and a boolean indicating if the value is present.
func (o OptionalRangeLocalDateTime) Get() (RangeLocalDateTime, bool) {
	return o.val, o.isSet
}

// Set sets the value.
func (o *OptionalRangeLocalDateTime) Set(val RangeLocalDateTime) {
	o.val = val
	o.isSet = true
}

// Unset marks the value as missing.
func (o *OptionalRangeLocalDateTime) Unset() {
	o.val = RangeLocalDateTime{}
	o.isSet = false
}

// MarshalJSON returns o marshaled as json.
func (o OptionalRangeLocalDateTime) MarshalJSON() ([]byte, error) {
	if o.isSet {
		return json.Marshal(o.val)
	}
	return json.Marshal(nil)
}

// UnmarshalJSON unmarshals bytes into *o.
func (o *OptionalRangeLocalDateTime) UnmarshalJSON(bytes []byte) error {
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

// NewRangeLocalDate creates a new RangeLocalDate value.
func NewRangeLocalDate(
	lower, upper OptionalLocalDate,
	incLower, incUpper bool,
) RangeLocalDate {
	if lower.isSet && !incLower {
		lower.val.days++
		incLower = true
	} else if !lower.isSet {
		incLower = false
	}

	if upper.isSet && incUpper {
		upper.val.days++
		incUpper = false
	} else if !upper.isSet {
		incUpper = false
	}

	if lower.isSet && upper.isSet && lower.val == upper.val {
		return RangeLocalDate{empty: true}
	}

	return RangeLocalDate{
		lower:    lower,
		upper:    upper,
		incLower: incLower,
		incUpper: incUpper,
	}
}

// RangeLocalDate is an interval of LocalDate values.
type RangeLocalDate struct {
	lower    OptionalLocalDate `gel:"lower"`
	upper    OptionalLocalDate `gel:"upper"`
	incLower bool              `gel:"inc_lower"`
	incUpper bool              `gel:"inc_upper"`
	empty    bool              `gel:"empty"`
}

// Lower returns the lower bound.
func (r RangeLocalDate) Lower() OptionalLocalDate { return r.lower }

// Upper returns the upper bound.
func (r RangeLocalDate) Upper() OptionalLocalDate { return r.upper }

// IncLower returns true if the lower bound is inclusive.
func (r RangeLocalDate) IncLower() bool { return r.incLower }

// IncUpper returns true if the upper bound is inclusive.
func (r RangeLocalDate) IncUpper() bool { return r.incUpper }

// Empty returns true if the range is empty.
func (r RangeLocalDate) Empty() bool { return r.empty }

type rangeLocalDateJSON struct {
	Lower    OptionalLocalDate `json:"lower"`
	Upper    OptionalLocalDate `json:"upper"`
	IncLower bool              `json:"inc_lower"`
	IncUpper bool              `json:"inc_upper"`
}

// MarshalJSON returns r marshaled as json.
func (r RangeLocalDate) MarshalJSON() ([]byte, error) {
	if r.empty {
		return []byte(`{"empty":true}`), nil
	}

	return json.Marshal(rangeLocalDateJSON{
		Lower:    r.lower,
		Upper:    r.upper,
		IncLower: r.incLower,
		IncUpper: r.incUpper,
	})
}

// UnmarshalJSON unmarshals bytes into *r.
func (r *RangeLocalDate) UnmarshalJSON(data []byte) error {
	var empty emptyRangeJSON
	err := json.Unmarshal(data, &empty)
	if err != nil {
		return err
	}

	if empty.Empty {
		r.empty = true
		return nil
	}

	var decoded rangeLocalDateJSON
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		return err
	}

	r.lower = decoded.Lower
	r.upper = decoded.Upper
	r.incLower = decoded.IncLower
	r.incUpper = decoded.IncUpper
	return nil
}

// NewOptionalRangeLocalDate is a convenience function for creating an
// OptionalRangeLocalDate with its value set to v.
func NewOptionalRangeLocalDate(v RangeLocalDate) OptionalRangeLocalDate {
	o := OptionalRangeLocalDate{}
	o.Set(v)
	return o
}

// OptionalRangeLocalDate is an optional RangeLocalDate. Optional types must be
// used for out parameters when a shape field is not required.
type OptionalRangeLocalDate struct {
	val   RangeLocalDate
	isSet bool
}

// Get returns the value and a boolean indicating if the value is present.
func (o OptionalRangeLocalDate) Get() (RangeLocalDate, bool) {
	return o.val, o.isSet
}

// Set sets the value.
func (o *OptionalRangeLocalDate) Set(val RangeLocalDate) {
	o.val = val
	o.isSet = true
}

// Unset marks the value as missing.
func (o *OptionalRangeLocalDate) Unset() {
	o.val = RangeLocalDate{}
	o.isSet = false
}

// MarshalJSON returns o marshaled as json.
func (o OptionalRangeLocalDate) MarshalJSON() ([]byte, error) {
	if o.isSet {
		return json.Marshal(o.val)
	}
	return json.Marshal(nil)
}

// UnmarshalJSON unmarshals bytes into *o.
func (o *OptionalRangeLocalDate) UnmarshalJSON(bytes []byte) error {
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
