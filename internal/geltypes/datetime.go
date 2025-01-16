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
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
)

// nolint:lll
var (
	zeroRelativeDuration = RelativeDuration{}
	zeroDateDuration     = DateDuration{}

	isoSecondsRegex       = regexp.MustCompile(`(-?\d+|-?\d+\.\d*|-?\d*\.\d+)S`)
	isoMinutesRegex       = regexp.MustCompile(`(-?\d+|-?\d+\.\d*|-?\d*\.\d+)M`)
	isoHoursRegex         = regexp.MustCompile(`(-?\d+|-?\d+\.\d*|-?\d*\.\d+)H`)
	isoUnitlessHoursRegex = regexp.MustCompile(`^(-?\d+|-?\d+\.\d*|-?\d*\.\d+)$`)
	isoDaysRegex          = regexp.MustCompile(`(-?\d+|-?\d+\.\d*|-?\d*\.\d+)D`)
	isoWeeksRegex         = regexp.MustCompile(`(-?\d+|-?\d+\.\d*|-?\d*\.\d+)W`)
	isoMonthsRegex        = regexp.MustCompile(`(-?\d+|-?\d+\.\d*|-?\d*\.\d+)M`)
	isoYearsRegex         = regexp.MustCompile(`(-?\d+|-?\d+\.\d*|-?\d*\.\d+)Y`)

	humanDurationMillenniumsRegex = regexp.MustCompile(`((?:(?:\s|^)-\s*)?\d*\.?\d*)\s*(?i:mil(\s|\d|\.|$)|millenni(?:um|a)(\s|$))`)
	humanDurationCenturiesRegex   = regexp.MustCompile(`((?:(?:\s|^)-\s*)?\d*\.?\d*)\s*(?i:c(\s|\d|\.|$)|centur(?:y|ies)(\s|$))`)
	humanDurationDecadesRegex     = regexp.MustCompile(`((?:(?:\s|^)-\s*)?\d*\.?\d*)\s*(?i:dec(\s|\d|\.|$)|decades?(\s|$))`)
	humanDurationYearsRegex       = regexp.MustCompile(`((?:(?:\s|^)-\s*)?\d*\.?\d*)\s*(?i:y(\s|\d|\.|$)|years?(\s|$))`)
	humanDurationMonthsRegex      = regexp.MustCompile(`((?:(?:\s|^)-\s*)?\d*\.?\d*)\s*(?i:mon(\s|\d|\.|$)|months?(\s|$))`)
	humanDurationWeeksRegex       = regexp.MustCompile(`((?:(?:\s|^)-\s*)?\d*\.?\d*)\s*(?i:w(\s|\d|\.|$)|weeks?(\s|$))`)
	humanDurationDaysRegex        = regexp.MustCompile(`((?:(?:\s|^)-\s*)?\d*\.?\d*)\s*(?i:d(\s|\d|\.|$)|days?(\s|$))`)
	humanDurationHoursRegex       = regexp.MustCompile(`((?:(?:\s|^)-\s*)?\d*\.?\d*)\s*(?i:h(\s|\d|\.|$)|hours?(\s|$))`)
	humanDurationMinutesRegex     = regexp.MustCompile(`((?:(?:\s|^)-\s*)?\d*\.?\d*)\s*(?i:m(\s|\d|\.|$)|minutes?(\s|$))`)
	humanDurationSecondsRegex     = regexp.MustCompile(`((?:(?:\s|^)-\s*)?\d*\.?\d*)\s*(?i:s(\s|\d|\.|$)|seconds?(\s|$))`)
	humanDurationMSRegex          = regexp.MustCompile(`((?:(?:\s|^)-\s*)?\d*\.?\d*)\s*(?i:ms(\s|\d|\.|$)|milliseconds?(\s|$))`)
	humanDurationUSRegex          = regexp.MustCompile(`((?:(?:\s|^)-\s*)?\d*\.?\d*)\s*(?i:us(\s|\d|\.|$)|microseconds?(\s|$))`)
)

const (
	monthsPerYear  int32 = 12
	usecsPerHour   int64 = 3_600_000_000
	usecsPerMinute int64 = 60_000_000
	usecsPerSecond int64 = 1_000_000

	// timeShift is the number of seconds
	// between 0001-01-01T00:00 and 1970-01-01T00:00
	timeShift = 62135596800
)

// NewOptionalDateTime is a convenience function for creating an
// OptionalDateTime with its value set to v.
func NewOptionalDateTime(v time.Time) OptionalDateTime {
	o := OptionalDateTime{}
	o.Set(v)
	return o
}

// OptionalDateTime is an optional time.Time.  Optional types must be used for
// out parameters when a shape field is not required.
type OptionalDateTime struct {
	val   time.Time
	isSet bool
}

// Get returns the value and a boolean indicating if the value is present.
func (o OptionalDateTime) Get() (time.Time, bool) {
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

// MarshalJSON returns o marshaled as json.
func (o OptionalDateTime) MarshalJSON() ([]byte, error) {
	if o.isSet {
		return json.Marshal(o.val)
	}
	return json.Marshal(nil)
}

// UnmarshalJSON unmarshals bytes into *o.
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
// https://www.edgedb.com/docs/stdlib/datetime#type::cal::local_datetime
type LocalDateTime struct {
	usec int64
}

func (dt LocalDateTime) String() string {
	sec := dt.usec/1_000_000 - timeShift
	nsec := (dt.usec % 1_000_000) * 1_000
	return time.Unix(sec, nsec).UTC().Format("2006-01-02T15:04:05.999999")
}

// MarshalText returns dt marshaled as text.
func (dt LocalDateTime) MarshalText() ([]byte, error) {
	return []byte(dt.String()), nil
}

// UnmarshalText unmarshals bytes into *dt.
func (dt *LocalDateTime) UnmarshalText(b []byte) error {
	t, err := time.Parse("2006-01-02T15:04:05.999999", string(b))
	if err != nil {
		return err
	}
	dt.usec = (t.Unix()+timeShift)*1_000_000 + int64(t.Nanosecond()/1000)

	return nil
}

// NewOptionalLocalDateTime is a convenience function for creating an
// OptionalLocalDateTime with its value set to v.
func NewOptionalLocalDateTime(v LocalDateTime) OptionalLocalDateTime {
	o := OptionalLocalDateTime{}
	o.Set(v)
	return o
}

// OptionalLocalDateTime is an optional LocalDateTime. Optional types must be
// used for out parameters when a shape field is not required.
type OptionalLocalDateTime struct {
	val   LocalDateTime
	isSet bool
}

// Get returns the value and a boolean indicating if the value is present.
func (o OptionalLocalDateTime) Get() (LocalDateTime, bool) {
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

// MarshalJSON returns o marshaled as json.
func (o OptionalLocalDateTime) MarshalJSON() ([]byte, error) {
	if o.isSet {
		return json.Marshal(o.val)
	}
	return json.Marshal(nil)
}

// UnmarshalJSON unmarshals bytes into *o.
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
// https://www.edgedb.com/docs/stdlib/datetime#type::cal::local_date
type LocalDate struct {
	days int32
}

func (d LocalDate) String() string {
	return time.Unix(
		int64(d.days)*86400-timeShift,
		0,
	).UTC().Format("2006-01-02")
}

// MarshalText returns d marshaled as text.
func (d LocalDate) MarshalText() ([]byte, error) {
	return []byte(d.String()), nil
}

// UnmarshalText unmarshals bytes into *d.
func (d *LocalDate) UnmarshalText(b []byte) error {
	t, err := time.Parse("2006-01-02", string(b))
	if err != nil {
		return err
	}
	d.days = int32((t.Unix() + timeShift) / 86400)

	return nil
}

// NewOptionalLocalDate is a convenience function for creating an
// OptionalLocalDate with its value set to v.
func NewOptionalLocalDate(v LocalDate) OptionalLocalDate {
	o := OptionalLocalDate{}
	o.Set(v)
	return o
}

// OptionalLocalDate is an optional LocalDate. Optional types must be used for
// out parameters when a shape field is not required.
type OptionalLocalDate struct {
	val   LocalDate
	isSet bool
}

// Get returns the value and a boolean indicating if the value is present.
func (o OptionalLocalDate) Get() (LocalDate, bool) { return o.val, o.isSet }

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

// MarshalJSON returns o marshaled as json.
func (o OptionalLocalDate) MarshalJSON() ([]byte, error) {
	if o.isSet {
		return json.Marshal(o.val)
	}
	return json.Marshal(nil)
}

// UnmarshalJSON unmarshals bytes into *o.
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
// https://www.edgedb.com/docs/stdlib/datetime#type::cal::local_time
type LocalTime struct {
	usec int64
}

func (t LocalTime) String() string {
	return time.Unix(
		t.usec/1_000_000,
		(t.usec%1_000_000)*1_000,
	).UTC().Format("15:04:05.999999")
}

// MarshalText returns t marshaled as text.
func (t LocalTime) MarshalText() ([]byte, error) {
	return []byte(t.String()), nil
}

// UnmarshalText unmarshals bytes into *t.
func (t *LocalTime) UnmarshalText(b []byte) error {
	pt, err := time.Parse("15:04:05.999999", string(b))
	if err != nil {
		return err
	}
	t.usec = pt.Unix()*1_000_000 + int64(pt.Nanosecond()/1000) +
		// microseconds between 0000-01-01T00:00 and 1970-01-01T00:00
		62_167_219_200_000_000

	return nil
}

// NewOptionalLocalTime is a convenience function for creating an
// OptionalLocalTime with its value set to v.
func NewOptionalLocalTime(v LocalTime) OptionalLocalTime {
	o := OptionalLocalTime{}
	o.Set(v)
	return o
}

// OptionalLocalTime is an optional LocalTime. Optional types must be used for
// out parameters when a shape field is not required.
type OptionalLocalTime struct {
	val   LocalTime
	isSet bool
}

// Get returns the value and a boolean indicating if the value is present.
func (o OptionalLocalTime) Get() (LocalTime, bool) { return o.val, o.isSet }

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

// MarshalJSON returns o marshaled as json.
func (o OptionalLocalTime) MarshalJSON() ([]byte, error) {
	if o.isSet {
		return json.Marshal(o.val)
	}
	return json.Marshal(nil)
}

// UnmarshalJSON unmarshals bytes into *o.
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

func popISOUnit(re *regexp.Regexp, str string) (float64, string, error) {
	matches := re.FindAllStringSubmatch(str, -1)

	var total float64
	s := str
	for _, match := range matches {
		if match[1] == "." || match[1] == "-." {
			return 0, "", fmt.Errorf("%q is not a valid number", match[1])
		}

		val, err := strconv.ParseFloat(match[1], 64)
		if err != nil {
			return 0, "", err
		}

		s = strings.Replace(s, match[0], "", 1)
		total += val
	}

	return total, s, nil
}

func parseDurationISO(str string) (Duration, error) {
	if !strings.HasPrefix(str, "PT") {
		return 0, fmt.Errorf("could not parse gel.Duration from %q", str)
	}

	time := str[2:]
	match := isoUnitlessHoursRegex.FindString(time)
	if match != "" {
		hours, err := strconv.ParseFloat(match, 64)
		if err != nil {
			return 0, fmt.Errorf(
				"could not parse gel.Duration from %q: %w",
				str, err)
		}

		return Duration(math.Round(3_600_000_000 * hours)), nil
	}

	hours, time, err := popISOUnit(isoHoursRegex, time)
	if err != nil {
		return 0, fmt.Errorf(
			"could not parse gel.Duration from %q: %w",
			str, err)
	}

	minutes, time, err := popISOUnit(isoMinutesRegex, time)
	if err != nil {
		return 0, fmt.Errorf(
			"could not parse gel.Duration from %q: %w",
			str, err)
	}

	seconds, time, err := popISOUnit(isoSecondsRegex, time)
	if err != nil {
		return 0, fmt.Errorf(
			"could not parse gel.Duration from %q: %w",
			str, err)
	}

	if time != "" {
		return 0, fmt.Errorf(
			"could not parse gel.Duration from %q: %w",
			str, err)
	}

	return Duration(
		math.Round(3_600_000_000*hours) +
			math.Round(60_000_000*minutes) +
			math.Round(1_000_000*seconds),
	), nil
}

func removeWhitespace(str string) string {
	var b strings.Builder
	b.Grow(len(str))
	for _, ch := range str {
		if !unicode.IsSpace(ch) {
			b.WriteRune(ch)
		}
	}
	return b.String()
}

func popHumanDurationUnit(
	re *regexp.Regexp,
	s string,
) (float64, bool, string, error) {
	match := re.FindStringSubmatch(s)
	if len(match) == 0 {
		return 0, false, s, nil
	}

	var number float64
	if match[1] != "" {
		literal := removeWhitespace(match[1])
		if strings.HasSuffix(literal, ".") {
			return 0, false, s, errors.New("no digits after decimal")
		}

		if strings.HasPrefix(literal, "-.") {
			return 0, false, s, fmt.Errorf(
				"no digits between minus sign and decimal")
		}

		var err error
		number, err = strconv.ParseFloat(literal, 64)
		if err != nil {
			return 0, false, s, err
		}

		s = strings.Replace(s, match[0], match[2], 1)
	}

	return number, true, s, nil
}

func parseDurationHuman(str string) (Duration, error) {
	var found bool

	hour, f, s, err := popHumanDurationUnit(humanDurationHoursRegex, str)
	if err != nil {
		return 0, fmt.Errorf(
			"could not parse gel.Duration from %q: %w",
			str, err)
	}
	found = found || f

	minute, f, s, err := popHumanDurationUnit(humanDurationMinutesRegex, s)
	if err != nil {
		return 0, fmt.Errorf(
			"could not parse gel.Duration from %q: %w",
			str, err)
	}
	found = found || f

	second, f, s, err := popHumanDurationUnit(humanDurationSecondsRegex, s)
	if err != nil {
		return 0, fmt.Errorf(
			"could not parse gel.Duration from %q: %w",
			str, err)
	}
	found = found || f

	ms, f, s, err := popHumanDurationUnit(humanDurationMSRegex, s)
	if err != nil {
		return 0, fmt.Errorf(
			"could not parse gel.Duration from %q: %w",
			str, err)
	}
	found = found || f

	us, f, s, err := popHumanDurationUnit(humanDurationUSRegex, s)
	if err != nil {
		return 0, fmt.Errorf(
			"could not parse gel.Duration from %q: %w",
			str, err)
	}
	found = found || f

	if !found {
		return 0, fmt.Errorf(
			"could not parse gel.Duration from %q: no duration found",
			str)
	}

	if strings.TrimSpace(s) != "" {
		return 0, fmt.Errorf(
			"could not parse gel.Duration from %q: extra characters %q",
			str,
			strings.TrimSpace(s),
		)
	}

	return Duration(
		math.Round(3_600_000_000*hour) +
			math.Round(60_000_000*minute) +
			math.Round(1_000_000*second) +
			math.Round(1_000*ms) +
			math.Round(us),
	), nil
}

// ParseDuration parses an Gel duration string.
func ParseDuration(s string) (Duration, error) {
	if strings.HasPrefix(s, "PT") {
		return parseDurationISO(s)
	}

	return parseDurationHuman(s)
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

// AsNanoseconds returns [time.Duration] represented as nanoseconds,
// after transforming from Duration microsecond representation.
// Returns an error if the Duration is too long and would cause an overflow of
// the internal int64 representation.
func (d Duration) AsNanoseconds() (time.Duration, error) {
	if int64(d) > math.MaxInt64/int64(time.Microsecond) ||
		int64(d) < math.MinInt64/int64(time.Microsecond) {
		return time.Duration(0), fmt.Errorf(
			"Duration is too large to be represented as nanoseconds",
		)
	}
	return time.Duration(d) * time.Microsecond, nil
}

// DurationFromNanoseconds creates a Duration represented as microseconds
// from a [time.Duration] represented as nanoseconds.
func DurationFromNanoseconds(d time.Duration) Duration {
	return Duration(math.RoundToEven(float64(d) / 1e3))
}

// NewOptionalDuration is a convenience function for creating an
// OptionalDuration with its value set to v.
func NewOptionalDuration(v Duration) OptionalDuration {
	o := OptionalDuration{}
	o.Set(v)
	return o
}

// OptionalDuration is an optional Duration. Optional types must be used for
// out parameters when a shape field is not required.
type OptionalDuration struct {
	val   Duration
	isSet bool
}

// Get returns the value and a boolean indicating if the value is present.
func (o OptionalDuration) Get() (Duration, bool) { return o.val, o.isSet }

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

// MarshalJSON returns o marshaled as json.
func (o OptionalDuration) MarshalJSON() ([]byte, error) {
	if o.isSet {
		return json.Marshal(o.val)
	}
	return json.Marshal(nil)
}

// UnmarshalJSON unmarshals bytes into *o.
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

// MarshalText returns rd marshaled as text.
func (rd RelativeDuration) MarshalText() ([]byte, error) {
	return []byte(rd.String()), nil
}

// UnmarshalText unmarshals bytes into *rd.
func (rd *RelativeDuration) UnmarshalText(b []byte) error {
	str := string(b)
	if !strings.HasPrefix(str, "P") {
		var found bool

		millennium, f, s, err := popHumanDurationUnit(
			humanDurationMillenniumsRegex,
			str,
		)
		if err != nil {
			return fmt.Errorf(
				"could not parse gel.RelativeDuration from %q: %w",
				str, err)
		}
		found = found || f

		century, f, s, err := popHumanDurationUnit(
			humanDurationCenturiesRegex,
			s,
		)
		if err != nil {
			return fmt.Errorf(
				"could not parse gel.RelativeDuration from %q: %w",
				str, err)
		}
		found = found || f

		decade, f, s, err := popHumanDurationUnit(humanDurationDecadesRegex, s)
		if err != nil {
			return fmt.Errorf(
				"could not parse gel.RelativeDuration from %q: %w",
				str, err)
		}
		found = found || f

		year, f, s, err := popHumanDurationUnit(humanDurationYearsRegex, s)
		if err != nil {
			return fmt.Errorf(
				"could not parse gel.RelativeDuration from %q: %w",
				str, err)
		}
		found = found || f

		month, f, s, err := popHumanDurationUnit(humanDurationMonthsRegex, s)
		if err != nil {
			return fmt.Errorf(
				"could not parse gel.RelativeDuration from %q: %w",
				str, err)
		}
		found = found || f

		week, f, s, err := popHumanDurationUnit(humanDurationWeeksRegex, s)
		if err != nil {
			return fmt.Errorf(
				"could not parse gel.RelativeDuration from %q: %w",
				str, err)
		}
		found = found || f

		day, f, s, err := popHumanDurationUnit(humanDurationDaysRegex, s)
		if err != nil {
			return fmt.Errorf(
				"could not parse gel.RelativeDuration from %q: %w",
				str, err)
		}
		found = found || f

		hour, f, s, err := popHumanDurationUnit(humanDurationHoursRegex, s)
		if err != nil {
			return fmt.Errorf(
				"could not parse gel.RelativeDuration from %q: %w",
				str, err)
		}
		found = found || f

		minute, f, s, err := popHumanDurationUnit(humanDurationMinutesRegex, s)
		if err != nil {
			return fmt.Errorf(
				"could not parse gel.RelativeDuration from %q: %w",
				str, err)
		}
		found = found || f

		second, f, s, err := popHumanDurationUnit(humanDurationSecondsRegex, s)
		if err != nil {
			return fmt.Errorf(
				"could not parse gel.RelativeDuration from %q: %w",
				str, err)
		}
		found = found || f

		ms, f, s, err := popHumanDurationUnit(humanDurationMSRegex, s)
		if err != nil {
			return fmt.Errorf(
				"could not parse gel.RelativeDuration from %q: %w",
				str, err)
		}
		found = found || f

		us, f, s, err := popHumanDurationUnit(humanDurationUSRegex, s)
		if err != nil {
			return fmt.Errorf(
				"could not parse gel.RelativeDuration from %q: %w",
				str, err)
		}
		found = found || f

		if !found {
			return fmt.Errorf(
				"could not parse gel.RelativeDuration from %q: "+
					"no duration found",
				str)
		}

		if strings.TrimSpace(s) != "" {
			return fmt.Errorf(
				"could not parse gel.RelativeDuration from %q: "+
					"extra characters %q",
				str, strings.TrimSpace(s))
		}

		months, monthsFraction := math.Modf(month)
		rd.months = int32(
			months +
				math.Round(12*(year+10*decade+100*century+1_000*millennium)),
		)

		days, daysFraction := math.Modf(day + 7*week + 30*monthsFraction)
		rd.days = int32(days)

		rd.microseconds = int64(
			math.Round(86_400_000_000*daysFraction) +
				math.Round(3_600_000_000*hour) +
				math.Round(60_000_000*minute) +
				math.Round(1_000_000*second) +
				math.Round(1_000*ms) +
				math.Round(us),
		)

		return nil
	}

	strs := strings.SplitN(str[1:], "T", 2)
	date := strs[0]
	var time string
	if len(strs) == 2 {
		time = strs[1]
	}

	years, date, err := popISOUnit(isoYearsRegex, date)
	if err != nil {
		return fmt.Errorf(
			"could not parse gel.RelativeDuration from %q: %w",
			str, err)
	}

	months, date, err := popISOUnit(isoMonthsRegex, date)
	if err != nil {
		return fmt.Errorf(
			"could not parse gel.RelativeDuration from %q: %w",
			str, err)
	}

	weeks, date, err := popISOUnit(isoWeeksRegex, date)
	if err != nil {
		return fmt.Errorf(
			"could not parse gel.RelativeDuration from %q: %w",
			str, err)
	}

	days, date, err := popISOUnit(isoDaysRegex, date)
	if err != nil {
		return fmt.Errorf(
			"could not parse gel.RelativeDuration from %q: %w",
			str, err)
	}

	if date != "" {
		return fmt.Errorf(
			"could not parse gel.RelativeDuration from %q: "+
				"extra characters in date",
			str)
	}

	match := isoUnitlessHoursRegex.FindString(time)
	if match != "" {
		var hours float64
		hours, err = strconv.ParseFloat(match, 64)
		if err != nil {
			return fmt.Errorf(
				"could not parse gel.RelativeDuration from %q: %w",
				str, err)
		}

		var monthsFraction float64
		months, monthsFraction = math.Modf(months)
		rd.months = int32(months + math.Round(12*years))

		var daysFraction float64
		days, daysFraction = math.Modf(days + 7*weeks + 30*monthsFraction)
		rd.days = int32(days)

		rd.microseconds = int64(
			math.Round(3_600_000_000*hours) +
				math.Round(86_400_000_000*daysFraction),
		)

		return nil
	}

	hours, time, err := popISOUnit(isoHoursRegex, time)
	if err != nil {
		return fmt.Errorf(
			"could not parse gel.RelativeDuration from %q: %w",
			str, err)
	}

	minutes, time, err := popISOUnit(isoMinutesRegex, time)
	if err != nil {
		return fmt.Errorf(
			"could not parse gel.relativeduration from %q: %w",
			str, err)
	}

	seconds, time, err := popISOUnit(isoSecondsRegex, time)
	if err != nil {
		return fmt.Errorf(
			"could not parse gel.relativeduration from %q: %w",
			str, err)
	}

	if time != "" {
		return fmt.Errorf(
			"could not parse gel.RelativeDuration from %q: "+
				"extra characters in time",
			str)
	}

	months, monthsFraction := math.Modf(months)
	rd.months = int32(months + math.Round(12*years))

	days, daysFraction := math.Modf(days + 7*weeks + 30*monthsFraction)
	rd.days = int32(days)

	rd.microseconds = int64(
		math.Round(1_000_000*seconds) +
			math.Round(60_000_000*minutes) +
			math.Round(3_600_000_000*hours) +
			math.Round(86_400_000_000*daysFraction),
	)

	return nil
}

// NewOptionalRelativeDuration is a convenience function for creating an
// OptionalRelativeDuration with its value set to v.
func NewOptionalRelativeDuration(v RelativeDuration) OptionalRelativeDuration {
	o := OptionalRelativeDuration{}
	o.Set(v)
	return o
}

// OptionalRelativeDuration is an optional RelativeDuration. Optional types
// must be used for out parameters when a shape field is not required.
type OptionalRelativeDuration struct {
	val   RelativeDuration
	isSet bool
}

// Get returns the value and a boolean indicating if the value is present.
func (o OptionalRelativeDuration) Get() (RelativeDuration, bool) {
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

// MarshalJSON returns o marshaled as json.
func (o OptionalRelativeDuration) MarshalJSON() ([]byte, error) {
	if o.isSet {
		return json.Marshal(o.val)
	}
	return json.Marshal(nil)
}

// UnmarshalJSON unmarshals bytes into *o.
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

// NewDateDuration returns a new DateDuration
func NewDateDuration(months int32, days int32) DateDuration {
	return DateDuration{days, months}
}

// DateDuration represents the elapsed time between two dates in a fuzzy human
// way.
type DateDuration struct {
	days   int32
	months int32
}

func (dd DateDuration) String() string {
	if dd == zeroDateDuration {
		return "P0D"
	}

	buf := []string{"P"}

	if dd.months != 0 {
		years := dd.months / monthsPerYear
		months := dd.months % monthsPerYear

		if years != 0 {
			buf = append(buf, strconv.FormatInt(int64(years), 10), "Y")
		}

		if months != 0 {
			buf = append(buf, strconv.FormatInt(int64(months), 10), "M")
		}
	}

	if dd.days != 0 {
		buf = append(buf, strconv.FormatInt(int64(dd.days), 10), "D")
	}

	return strings.Join(buf, "")
}

// MarshalText returns dd marshaled as text.
func (dd DateDuration) MarshalText() ([]byte, error) {
	return []byte(dd.String()), nil
}

// UnmarshalText unmarshals bytes into *dd.
func (dd *DateDuration) UnmarshalText(b []byte) error {
	str := string(b)
	if !strings.HasPrefix(str, "P") {
		var found bool

		millennium, f, s, err := popHumanDurationUnit(
			humanDurationMillenniumsRegex,
			str,
		)
		if err != nil {
			return fmt.Errorf(
				"could not parse gel.DateDuration from %q: %w",
				str, err)
		}
		found = found || f

		century, f, s, err := popHumanDurationUnit(
			humanDurationCenturiesRegex,
			s,
		)
		if err != nil {
			return fmt.Errorf(
				"could not parse gel.DateDuration from %q: %w",
				str, err)
		}
		found = found || f

		decade, f, s, err := popHumanDurationUnit(humanDurationDecadesRegex, s)
		if err != nil {
			return fmt.Errorf(
				"could not parse gel.DateDuration from %q: %w",
				str, err)
		}
		found = found || f

		year, f, s, err := popHumanDurationUnit(humanDurationYearsRegex, s)
		if err != nil {
			return fmt.Errorf(
				"could not parse gel.DateDuration from %q: %w",
				str, err)
		}
		found = found || f

		month, f, s, err := popHumanDurationUnit(humanDurationMonthsRegex, s)
		if err != nil {
			return fmt.Errorf(
				"could not parse gel.DateDuration from %q: %w",
				str, err)
		}
		found = found || f

		week, f, s, err := popHumanDurationUnit(humanDurationWeeksRegex, s)
		if err != nil {
			return fmt.Errorf(
				"could not parse gel.DateDuration from %q: %w",
				str, err)
		}
		found = found || f

		day, f, s, err := popHumanDurationUnit(humanDurationDaysRegex, s)
		if err != nil {
			return fmt.Errorf(
				"could not parse gel.DateDuration from %q: %w",
				str, err)
		}
		found = found || f

		if !found {
			return fmt.Errorf(
				"could not parse gel.DateDuration from %q: "+
					"no duration found",
				str)
		}

		if strings.TrimSpace(s) != "" {
			return fmt.Errorf(
				"could not parse gel.DateDuration from %q: "+
					"extra characters %q",
				str, strings.TrimSpace(s))
		}

		months, monthsFraction := math.Modf(month)
		dd.months = int32(
			months +
				math.Round(12*(year+10*decade+100*century+1_000*millennium)),
		)

		days, daysFraction := math.Modf(day + 7*week + 30*monthsFraction)
		dd.days = int32(days)

		if daysFraction != 0 {
			return fmt.Errorf(
				"could not parse gel.DateDuration from %q: "+
					"units smaller than days cannot be used",
				str)
		}

		return nil
	}

	strs := strings.SplitN(str[1:], "T", 2)
	date := strs[0]
	if len(strs) == 2 {
		time := strs[1]
		float, s, err := popISOUnit(isoUnitlessHoursRegex, time)
		if err != nil {
			return fmt.Errorf(
				"could not parse gel.DateDuration from %q: %w",
				str, err)
		}

		if float != 0 {
			return fmt.Errorf(
				"could not parse gel.DateDuration from %q: "+
					"units smaller than days cannot be used",
				str)
		}

		if s != "" {
			return fmt.Errorf(
				"could not parse gel.DateDuration from %q", str)
		}
	}

	years, date, err := popISOUnit(isoYearsRegex, date)
	if err != nil {
		return fmt.Errorf(
			"could not parse gel.DateDuration from %q: %w", str, err)
	}

	months, date, err := popISOUnit(isoMonthsRegex, date)
	if err != nil {
		return fmt.Errorf(
			"could not parse gel.DateDuration from %q: %w", str, err)
	}

	weeks, date, err := popISOUnit(isoWeeksRegex, date)
	if err != nil {
		return fmt.Errorf(
			"could not parse gel.DateDuration from %q: %w", str, err)
	}

	days, date, err := popISOUnit(isoDaysRegex, date)
	if err != nil {
		return fmt.Errorf(
			"could not parse gel.DateDuration from %q: %w", str, err)
	}

	if date != "" {
		return fmt.Errorf("could not parse gel.DateDuration from %q", str)
	}

	months, monthsFraction := math.Modf(months)
	days, daysFraction := math.Modf(days + 7*weeks + 30*monthsFraction)
	if daysFraction != 0 {
		return fmt.Errorf("could not parse gel.DateDuration from %q", str)
	}

	dd.months = int32(months + math.Round(12*years))
	dd.days = int32(days)
	return nil
}

// NewOptionalDateDuration is a convenience function for creating an
// OptionalDateDuration with its value set to v.
func NewOptionalDateDuration(v DateDuration) OptionalDateDuration {
	o := OptionalDateDuration{}
	o.Set(v)
	return o
}

// OptionalDateDuration is an optional DateDuration. Optional types
// must be used for out parameters when a shape field is not required.
type OptionalDateDuration struct {
	val   DateDuration
	isSet bool
}

// Get returns the value and a boolean indicating if the value is present.
func (o *OptionalDateDuration) Get() (DateDuration, bool) {
	return o.val, o.isSet
}

// Set sets the value.
func (o *OptionalDateDuration) Set(val DateDuration) {
	o.val = val
	o.isSet = true
}

// Unset marks the value as missing.
func (o *OptionalDateDuration) Unset() {
	o.val = DateDuration{}
	o.isSet = false
}

// MarshalJSON returns o marshaled as json.
func (o OptionalDateDuration) MarshalJSON() ([]byte, error) {
	if o.isSet {
		return json.Marshal(o.val)
	}
	return json.Marshal(nil)
}

// UnmarshalJSON unmarshals bytes into *o.
func (o *OptionalDateDuration) UnmarshalJSON(bytes []byte) error {
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
