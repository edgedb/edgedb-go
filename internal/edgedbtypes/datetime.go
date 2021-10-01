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
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var zeroRelativeDuration = RelativeDuration{}

const (
	monthsPerYear  int32 = 12
	usecsPerHour   int64 = 3_600_000_000
	usecsPerMinute int64 = 60_000_000
	usecsPerSecond int64 = 1_000_000

	// timeShift is the number of seconds
	// between 0001-01-01T00:00 and 1970-01-01T00:00
	timeShift = 62135596800
)

// OptionalDateTime is an optional time.Time.  Optional types must be used for
// out parameters when a shape field is not required.
type OptionalDateTime struct {
	val   time.Time
	isSet bool
}

// Get returns the value and a boolean indicating if the value is present.
func (o *OptionalDateTime) Get() (time.Time, bool) {
	return o.val, o.isSet
}

// Set sets the value.
func (o *OptionalDateTime) Set(val time.Time) {
	o.val = val
	o.isSet = true
}

// Unset marks the value as missing.
func (o *OptionalDateTime) Unset() {
	o.val = time.Time{}
	o.isSet = false
}

func (o OptionalDateTime) MarshalJSON() ([]byte, error) {
	if o.isSet {
		return json.Marshal(o.val)
	}
	return json.Marshal(nil)
}

func (o *OptionalDateTime) UnmarshalJSON(bytes []byte) error {
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

// NewLocalDateTime returns a new LocalDateTime
func NewLocalDateTime(
	year int, month time.Month, day, hour, minute, second, microsecond int,
) LocalDateTime {
	t := time.Date(
		year, month, day, hour, minute, second, microsecond*1_000, time.UTC,
	)
	sec := t.Unix() + timeShift
	nsec := int64(t.Sub(time.Unix(t.Unix(), 0)))
	return LocalDateTime{sec*1_000_000 + nsec/1_000}
}

// LocalDateTime is a date and time without timezone.
// https://www.edgedb.com/docs/datamodel/scalars/datetime/
type LocalDateTime struct {
	usec int64
}

func (dt LocalDateTime) String() string {
	sec := dt.usec/1_000_000 - timeShift
	nsec := (dt.usec % 1_000_000) * 1_000
	return time.Unix(sec, nsec).UTC().Format("2006-01-02T15:04:05.999999")
}

func (dt LocalDateTime) MarshalText() ([]byte, error) {
	return []byte(dt.String()), nil
}

func (dt *LocalDateTime) UnmarshalText(b []byte) error {
	t, err := time.Parse("2006-01-02T15:04:05.999999", string(b))
	if err != nil {
		return err
	}
	dt.usec = t.UnixMicro() + timeShift*1_000_000

	return nil
}

// OptionalLocalDateTime is an optional LocalDateTime. Optional types must be
// used for out parameters when a shape field is not required.
type OptionalLocalDateTime struct {
	val   LocalDateTime
	isSet bool
}

// Get returns the value and a boolean indicating if the value is present.
func (o *OptionalLocalDateTime) Get() (LocalDateTime, bool) {
	return o.val, o.isSet
}

// Set sets the value.
func (o *OptionalLocalDateTime) Set(val LocalDateTime) {
	o.val = val
	o.isSet = true
}

// Unset marks the value as missing.
func (o *OptionalLocalDateTime) Unset() {
	o.val = LocalDateTime{}
	o.isSet = false
}

func (o OptionalLocalDateTime) MarshalJSON() ([]byte, error) {
	if o.isSet {
		return json.Marshal(o.val)
	}
	return json.Marshal(nil)
}

func (o *OptionalLocalDateTime) UnmarshalJSON(bytes []byte) error {
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

// NewLocalDate returns a new LocalDate
func NewLocalDate(year int, month time.Month, day int) LocalDate {
	t := time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
	return LocalDate{int32((t.Unix() + timeShift) / 86400)}
}

// LocalDate is a date without a time zone.
// https://www.edgedb.com/docs/datamodel/scalars/datetime/
type LocalDate struct {
	days int32
}

func (d LocalDate) String() string {
	return time.Unix(
		int64(d.days)*86400-timeShift,
		0,
	).UTC().Format("2006-01-02")
}

func (d LocalDate) MarshalText() ([]byte, error) {
	return []byte(d.String()), nil
}

func (d *LocalDate) UnmarshalText(b []byte) error {
	t, err := time.Parse("2006-01-02", string(b))
	if err != nil {
		return err
	}
	d.days = int32((t.Unix() + timeShift) / 86400)

	return nil
}

// OptionalLocalDate is an optional LocalDate. Optional types must be used for
// out parameters when a shape field is not required.
type OptionalLocalDate struct {
	val   LocalDate
	isSet bool
}

// Get returns the value and a boolean indicating if the value is present.
func (o *OptionalLocalDate) Get() (LocalDate, bool) { return o.val, o.isSet }

// Set sets the value.
func (o *OptionalLocalDate) Set(val LocalDate) {
	o.val = val
	o.isSet = true
}

// Unset marks the value as missing.
func (o *OptionalLocalDate) Unset() {
	o.val = LocalDate{}
	o.isSet = false
}

func (o OptionalLocalDate) MarshalJSON() ([]byte, error) {
	if o.isSet {
		return json.Marshal(o.val)
	}
	return json.Marshal(nil)
}

func (o *OptionalLocalDate) UnmarshalJSON(bytes []byte) error {
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

// NewLocalTime returns a new LocalTime
func NewLocalTime(hour, minute, second, microsecond int) LocalTime {
	if hour < 0 || hour > 23 {
		panic("hour out of range 0-23")
	}

	if minute < 0 || minute > 59 {
		panic("minute out of range 0-59")
	}

	if second < 0 || second > 59 {
		panic("second out of range 0-59")
	}

	if microsecond < 0 || microsecond > 999_999 {
		panic("microsecond out of range 0-999_999")
	}

	t := time.Date(
		1970, 1, 1, hour, minute, second, microsecond*1_000, time.UTC,
	)
	return LocalTime{t.UnixNano() / 1_000}
}

// LocalTime is a time without a time zone.
// https://www.edgedb.com/docs/datamodel/scalars/datetime/
type LocalTime struct {
	usec int64
}

func (t LocalTime) String() string {
	return time.Unix(
		t.usec/1_000_000,
		(t.usec%1_000_000)*1_000,
	).UTC().Format("15:04:05.999999")
}

func (t LocalTime) MarshalText() ([]byte, error) {
	return []byte(t.String()), nil
}

func (t *LocalTime) UnmarshalText(b []byte) error {
	pt, err := time.Parse("15:04:05.999999", string(b))
	if err != nil {
		return err
	}
	// microseconds between 0000-01-01T00:00 and 1970-01-01T00:00
	t.usec = pt.UnixMicro() + 62_167_219_200_000_000

	return nil
}

// OptionalLocalTime is an optional LocalTime. Optional types must be used for
// out parameters when a shape field is not required.
type OptionalLocalTime struct {
	val   LocalTime
	isSet bool
}

// Get returns the value and a boolean indicating if the value is present.
func (o *OptionalLocalTime) Get() (LocalTime, bool) { return o.val, o.isSet }

// Set sets the value.
func (o *OptionalLocalTime) Set(val LocalTime) {
	o.val = val
	o.isSet = true
}

// Unset marks the value as missing.
func (o *OptionalLocalTime) Unset() {
	o.val = LocalTime{}
	o.isSet = false
}

func (o OptionalLocalTime) MarshalJSON() ([]byte, error) {
	if o.isSet {
		return json.Marshal(o.val)
	}
	return json.Marshal(nil)
}

func (o *OptionalLocalTime) UnmarshalJSON(bytes []byte) error {
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

// Duration represents the elapsed time between two instants
// as an int64 microsecond count.
type Duration int64

func (d Duration) String() string {
	if d == 0 {
		return "PT0S"
	}

	usecs := int64(d)
	hours := usecs / usecsPerHour
	usecs -= hours * usecsPerHour
	minutes := usecs / usecsPerMinute
	usecs -= minutes * usecsPerMinute
	seconds := usecs / usecsPerSecond
	usecs -= seconds * usecsPerSecond

	buf := []string{"PT"}

	if hours != 0 {
		buf = append(buf, strconv.FormatInt(hours, 10), "H")
	}

	if minutes != 0 {
		buf = append(buf, strconv.FormatInt(minutes, 10), "M")
	}

	if seconds != 0 || usecs != 0 {
		if seconds < 0 || usecs < 0 {
			buf = append(buf, "-")
		}

		if seconds < 0 {
			seconds = -seconds
		}

		buf = append(buf, strconv.FormatInt(seconds, 10))

		if usecs != 0 {
			if usecs < 0 {
				usecs = -usecs
			}

			str := fmt.Sprintf(".%0.6d", usecs)
			str = strings.TrimRight(str, "0")
			buf = append(buf, str)
		}

		buf = append(buf, "S")
	}

	return strings.Join(buf, "")
}

// OptionalDuration is an optional Duration. Optional types must be used for
// out parameters when a shape field is not required.
type OptionalDuration struct {
	val   Duration
	isSet bool
}

// Get returns the value and a boolean indicating if the value is present.
func (o *OptionalDuration) Get() (Duration, bool) { return o.val, o.isSet }

// Set sets the value.
func (o *OptionalDuration) Set(val Duration) {
	o.val = val
	o.isSet = true
}

// Unset marks the value as missing.
func (o *OptionalDuration) Unset() {
	o.val = 0
	o.isSet = false
}

func (o OptionalDuration) MarshalJSON() ([]byte, error) {
	if o.isSet {
		return json.Marshal(o.val)
	}
	return json.Marshal(nil)
}

func (o *OptionalDuration) UnmarshalJSON(bytes []byte) error {
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

// NewRelativeDuration returns a new RelativeDuration
func NewRelativeDuration(
	months, days int32,
	microseconds int64,
) RelativeDuration {
	return RelativeDuration{microseconds, days, months}
}

// RelativeDuration represents the elapsed time between two instants in a fuzzy
// human way.
type RelativeDuration struct {
	microseconds int64
	days         int32
	months       int32
}

func (rd RelativeDuration) String() string {
	if rd == zeroRelativeDuration {
		return "PT0S"
	}

	buf := []string{"P"}

	if rd.months != 0 {
		years := rd.months / monthsPerYear
		months := rd.months % monthsPerYear

		if years != 0 {
			buf = append(buf, strconv.FormatInt(int64(years), 10), "Y")
		}

		if months != 0 {
			buf = append(buf, strconv.FormatInt(int64(months), 10), "M")
		}
	}

	if rd.days != 0 {
		buf = append(buf, strconv.FormatInt(int64(rd.days), 10), "D")
	}

	if rd.microseconds == 0 {
		return strings.Join(buf, "")
	}

	buf = append(buf, "T")

	usecs := rd.microseconds
	hours := usecs / usecsPerHour
	usecs -= hours * usecsPerHour
	minutes := usecs / usecsPerMinute
	usecs -= minutes * usecsPerMinute
	seconds := usecs / usecsPerSecond
	usecs -= seconds * usecsPerSecond

	if hours != 0 {
		buf = append(buf, strconv.FormatInt(hours, 10), "H")
	}

	if minutes != 0 {
		buf = append(buf, strconv.FormatInt(minutes, 10), "M")
	}

	if seconds != 0 || usecs != 0 {
		if seconds < 0 || usecs < 0 {
			buf = append(buf, "-")
		}

		if seconds < 0 {
			seconds = -seconds
		}

		buf = append(buf, strconv.FormatInt(seconds, 10))

		if usecs != 0 {
			if usecs < 0 {
				usecs = -usecs
			}

			str := fmt.Sprintf(".%0.6d", usecs)
			str = strings.TrimRight(str, "0")
			buf = append(buf, str)
		}

		buf = append(buf, "S")
	}

	return strings.Join(buf, "")
}

func (rd RelativeDuration) MarshalText() ([]byte, error) {
	return []byte(rd.String()), nil
}

var errMalformedRelativeDuration = errors.New(
	"malformed edgedb.RelativeDuration")

var relDurationRegex = regexp.MustCompile(
	`P(?:(-?\d+)Y)?(?:(-?\d+)M)?(?:(-?\d+)D)?` +
		`(?:T(?:(-?\d+)H)?(?:(-?\d+)M)?(?:(-?\d+)(?:\.(\d{1,6}))?S)?)?`,
)

func (rd *RelativeDuration) UnmarshalText(b []byte) error {
	str := string(b)
	*rd = RelativeDuration{}

	if str == "PT0S" {
		return nil
	}

	match := relDurationRegex.FindStringSubmatch(str)
	if len(match) == 0 {
		return errMalformedRelativeDuration
	}

	fmt.Println(match[1], match[2], match[3], match[4], match[5], match[6], match[7])
	if match[1] != "" {
		years, err := strconv.ParseInt(match[1], 10, 32)
		if err != nil {
			return err
		}
		rd.months = int32(years) * monthsPerYear
	}
	if match[2] != "" {
		months, err := strconv.ParseInt(match[2], 10, 32)
		if err != nil {
			return err
		}
		rd.months += int32(months)
	}
	if match[3] != "" {
		days, err := strconv.ParseInt(match[3], 10, 32)
		if err != nil {
			return err
		}
		rd.days = int32(days)
	}
	if match[4] != "" {
		hours, err := strconv.ParseInt(match[4], 10, 64)
		if err != nil {
			return err
		}
		rd.microseconds = hours * usecsPerHour
	}
	if match[5] != "" {
		minutes, err := strconv.ParseInt(match[5], 10, 64)
		if err != nil {
			return err
		}
		rd.microseconds += minutes * usecsPerMinute
	}
	if match[6] != "" {
		secs, err := strconv.ParseInt(match[6], 10, 64)
		if err != nil {
			return err
		}
		rd.microseconds += secs * usecsPerSecond
	}
	if match[7] != "" {
		usecs, err := strconv.ParseInt(match[7], 10, 64)
		if err != nil {
			return err
		}
		rd.microseconds += usecs
	}

	return nil
}

// OptionalRelativeDuration is an optional RelativeDuration. Optional types
// must be used for out parameters when a shape field is not required.
type OptionalRelativeDuration struct {
	val   RelativeDuration
	isSet bool
}

// Get returns the value and a boolean indicating if the value is present.
func (o *OptionalRelativeDuration) Get() (RelativeDuration, bool) {
	return o.val, o.isSet
}

// Set sets the value.
func (o *OptionalRelativeDuration) Set(val RelativeDuration) {
	o.val = val
	o.isSet = true
}

// Unset marks the value as missing.
func (o *OptionalRelativeDuration) Unset() {
	o.val = RelativeDuration{}
	o.isSet = false
}

func (o OptionalRelativeDuration) MarshalJSON() ([]byte, error) {
	if o.isSet {
		return json.Marshal(o.val)
	}
	return json.Marshal(nil)
}

func (o *OptionalRelativeDuration) UnmarshalJSON(bytes []byte) error {
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
