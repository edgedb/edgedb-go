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

package edgedb

import (
	"context"
	"encoding/binary"
	"fmt"
	"math"
	"math/big"
	"math/rand"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSendAndReceveInt64(t *testing.T) {
	ctx := context.Background()

	numbers := []int64{
		-1,
		1,
		0,
		11,
		-11,
		15,
		22,
		113,
		-11111,
		110000,
		-1100000,
		346456723423,
		-346456723423,
		281474976710656,
		2251799813685125,
		9007199254740992,
		-2251799813685125,
		1152921504594725865,
		-1152921504594725865,
	}

	for i := 0; i < 1000; i++ {
		numbers = append(numbers, int64(rand.Uint64()))
	}

	strings := make([]string, len(numbers))
	for i, n := range numbers {
		strings[i] = fmt.Sprint(n)
	}

	type Result struct {
		Encoded   string        `edgedb:"encoded"`
		Decoded   int64         `edgedb:"decoded"`
		RoundTrip int64         `edgedb:"round_trip"`
		IsEqual   bool          `edgedb:"is_equal"`
		Nested    []interface{} `edgedb:"nested"`
		String    string        `edgedb:"string"`
	}

	query := `
		WITH
			x := (
				WITH
					n := enumerate(array_unpack(<array<int64>>$0)),
					s := enumerate(array_unpack(<array<str>>$1)),
				SELECT (
					n := n.1,
					s := s.1,
				)
				FILTER n.0 = s.0
			)
		SELECT (
			encoded := <str>x.n,
			decoded := <int64>x.s,
			round_trip := x.n,
			is_equal := <int64>x.s = x.n,
			nested := ([x.n],),
			string := <str><int64>x.s,
		)
	`

	var results []Result
	err := conn.Query(ctx, query, &results, numbers, strings)
	require.Nil(t, err, "unexpected error: %v", err)
	require.Equal(t, len(numbers), len(results), "unexpected result count")

	for i, s := range strings {
		t.Run(s, func(t *testing.T) {
			n := numbers[i]
			r := results[i]

			assert.True(t, r.IsEqual, "equality check faild")
			assert.Equal(t, s, r.Encoded, "encoding failed")
			assert.Equal(t, n, r.Decoded, "decoding failed")
			assert.Equal(t, n, r.RoundTrip, "round trip failed")
			assert.Equal(t, s, r.String)
			assert.Equal(t, []interface{}{[]int64{n}}, r.Nested)
		})
	}
}

func TestSendAndReceveInt32(t *testing.T) {
	ctx := context.Background()

	numbers := []int32{-1, 0, 1, 10, 2147483647}
	for i := 0; i < 1000; i++ {
		numbers = append(numbers, int32(rand.Uint32()))
	}

	strings := make([]string, len(numbers))
	for i, n := range numbers {
		strings[i] = fmt.Sprint(n)
	}

	type Result struct {
		Encoded   string        `edgedb:"encoded"`
		Decoded   int32         `edgedb:"decoded"`
		RoundTrip int32         `edgedb:"round_trip"`
		IsEqual   bool          `edgedb:"is_equal"`
		Nested    []interface{} `edgedb:"nested"`
		String    string        `edgedb:"string"`
	}

	query := `
		WITH
			x := (
				WITH
					n := enumerate(array_unpack(<array<int32>>$0)),
					s := enumerate(array_unpack(<array<str>>$1)),
				SELECT (
					n := n.1,
					s := s.1,
				)
				FILTER n.0 = s.0
			)
		SELECT (
			encoded := <str>x.n,
			decoded := <int32>x.s,
			round_trip := x.n,
			is_equal := <int32>x.s = x.n,
			nested := ([x.n],),
			string := <str><int32>x.s,
		)
	`

	var results []Result
	err := conn.Query(ctx, query, &results, numbers, strings)
	require.Nil(t, err, "unexpected error: %v", err)
	require.Equal(t, len(numbers), len(results), "wrong number of results")

	for i, s := range strings {
		t.Run(s, func(t *testing.T) {
			n := numbers[i]
			r := results[i]

			assert.True(t, r.IsEqual, "equality check faild")
			assert.Equal(t, s, r.Encoded, "encoding failed")
			assert.Equal(t, n, r.Decoded, "decoding failed")
			assert.Equal(t, n, r.RoundTrip)
			assert.Equal(t, s, r.String)
			assert.Equal(t, []interface{}{[]int32{n}}, r.Nested)
		})
	}
}

func TestSendAndReceveInt16(t *testing.T) {
	ctx := context.Background()

	numbers := []int16{-1, 0, 1, 10, 15, 22, -1111}
	for i := 0; i < 1000; i++ {
		numbers = append(numbers, int16(rand.Uint32()))
	}

	strings := make([]string, len(numbers))
	for i, n := range numbers {
		strings[i] = fmt.Sprint(n)
	}

	type Result struct {
		Encoded   string        `edgedb:"encoded"`
		Decoded   int16         `edgedb:"decoded"`
		RoundTrip int16         `edgedb:"round_trip"`
		IsEqual   bool          `edgedb:"is_equal"`
		Nested    []interface{} `edgedb:"nested"`
		String    string        `edgedb:"string"`
	}

	query := `
		WITH
			x := (
				WITH
					n := enumerate(array_unpack(<array<int16>>$0)),
					s := enumerate(array_unpack(<array<str>>$1)),
				SELECT (
					n := n.1,
					s := s.1,
				)
				FILTER n.0 = s.0
			)
		SELECT (
			encoded := <str>x.n,
			decoded := <int16>x.s,
			round_trip := x.n,
			is_equal := <int16>x.s = x.n,
			nested := ([x.n],),
			string := <str><int16>x.s,
		)
	`

	var results []Result
	err := conn.Query(ctx, query, &results, numbers, strings)
	require.Nil(t, err, "unexpected error: %v", err)
	require.Equal(t, len(numbers), len(results), "wrong number of results")

	for i, s := range strings {
		t.Run(s, func(t *testing.T) {
			n := numbers[i]
			r := results[i]

			assert.True(t, r.IsEqual, "equality check faild")
			assert.Equal(t, s, r.Encoded, "encoding failed")
			assert.Equal(t, n, r.Decoded, "decoding failed")
			assert.Equal(t, n, r.RoundTrip, "round trip failed")
			assert.Equal(t, s, r.String)
			assert.Equal(t, []interface{}{[]int16{n}}, r.Nested)
		})
	}
}

func TestSendAndReceveBool(t *testing.T) {
	ctx := context.Background()

	query := `
		WITH
			i := <bool>$0,
			s := <str>$1,
		SELECT (
			encoded := <str>i,
			decoded := <bool>s,
			round_trip := i,
			is_equal := <bool>s = i,
			nested := ([i],),
			string := <str><bool>s,
		)
	`

	type Result struct {
		Encoded   string        `edgedb:"encoded"`
		Decoded   bool          `edgedb:"decoded"`
		RoundTrip bool          `edgedb:"round_trip"`
		IsEqual   bool          `edgedb:"is_equal"`
		Nested    []interface{} `edgedb:"nested"`
		String    string        `edgedb:"string"`
	}

	samples := []bool{true, false}

	for _, i := range samples {
		s := fmt.Sprint(i)
		t.Run(s, func(t *testing.T) {
			var result Result
			err := conn.QueryOne(ctx, query, &result, i, s)
			assert.Nil(t, err, "unexpected error: %v", err)

			assert.True(t, result.IsEqual, "equality check faild")
			assert.Equal(t, s, result.Encoded, "encoding failed")
			assert.Equal(t, i, result.Decoded, "decoding failed")
			assert.Equal(t, i, result.RoundTrip)
			assert.Equal(t, s, result.String)
			assert.Equal(t, []interface{}{[]bool{i}}, result.Nested)
		})
	}
}

func TestSendAndReceveFloat64(t *testing.T) {
	ctx := context.Background()

	numbers := []float64{0, 1, 123.2, -1.1}
	for i := 0; i < 1000; i++ {
		n := math.Float64frombits(rand.Uint64())

		// NaN is not equal to itself so assertions will fail.
		if !math.IsNaN(n) {
			numbers = append(numbers, n)
		}
	}

	strings := make([]string, len(numbers))
	for i, n := range numbers {
		strings[i] = fmt.Sprint(n)
	}

	type Result struct {
		Encoded   string        `edgedb:"encoded"`
		Decoded   float64       `edgedb:"decoded"`
		RoundTrip float64       `edgedb:"round_trip"`
		IsEqual   bool          `edgedb:"is_equal"`
		Nested    []interface{} `edgedb:"nested"`
	}

	query := `
		WITH
			x := (
				WITH
					n := enumerate(array_unpack(<array<float64>>$0)),
					s := enumerate(array_unpack(<array<str>>$1)),
				SELECT (
					n := n.1,
					s := s.1,
				)
				FILTER n.0 = s.0
			)
		SELECT (
			encoded := <str>x.n,
			decoded := <float64>x.s,
			round_trip := x.n,
			is_equal := <float64>x.s = x.n,
			nested := ([x.n],),
		)
	`

	var results []Result
	err := conn.Query(ctx, query, &results, numbers, strings)
	require.Nil(t, err, "unexpected error: %v", err)
	require.Equal(t, len(numbers), len(results), "wrong number of results")

	for i, s := range strings {
		t.Run(s, func(t *testing.T) {
			n := numbers[i]
			r := results[i]

			encoded, err := strconv.ParseFloat(r.Encoded, 64)
			require.Nil(t, err)

			assert.True(t, r.IsEqual, "equality check faild")
			assert.Equal(t, n, encoded, "encoding failed")
			assert.Equal(t, n, r.Decoded, "decoding failed")
			assert.Equal(t, n, r.RoundTrip, "round trip failed")
			assert.Equal(t, []interface{}{[]float64{n}}, r.Nested)
		})
	}
}

func TestSendAndReceveFloat32(t *testing.T) {
	ctx := context.Background()

	numbers := []float32{0, 1, 123.2, -1.1}
	for i := 0; i < 1000; i++ {
		n := math.Float32frombits(rand.Uint32())

		// NaN is not equal to itself so assertions will fail.
		if !math.IsNaN(float64(n)) {
			numbers = append(numbers, n)
		}
	}

	strings := make([]string, len(numbers))
	for i, n := range numbers {
		strings[i] = fmt.Sprint(n)
	}

	type Result struct {
		Encoded   string        `edgedb:"encoded"`
		Decoded   float32       `edgedb:"decoded"`
		RoundTrip float32       `edgedb:"round_trip"`
		IsEqual   bool          `edgedb:"is_equal"`
		Nested    []interface{} `edgedb:"nested"`
	}

	query := `
		WITH
			x := (
				WITH
					n := enumerate(array_unpack(<array<float32>>$0)),
					s := enumerate(array_unpack(<array<str>>$1)),
				SELECT (
					n := n.1,
					s := s.1,
				)
				FILTER n.0 = s.0
			)
		SELECT (
			encoded := <str><float32>x.n,
			decoded := <float32>x.s,
			round_trip := x.n,
			is_equal := <float32>x.s = x.n,
			nested := ([x.n],),
		)
	`

	var results []Result
	err := conn.Query(ctx, query, &results, numbers, strings)
	require.Nil(t, err, "unexpected error: %v", err)
	require.Equal(t, len(numbers), len(results), "wrong number of results")

	for i, s := range strings {
		t.Run(s, func(t *testing.T) {
			n := numbers[i]
			r := results[i]

			encoded, err := strconv.ParseFloat(r.Encoded, 32)
			require.Nil(t, err)

			assert.True(t, r.IsEqual, "equality check faild")
			assert.Equal(t, n, float32(encoded), "encoding failed")
			assert.Equal(t, n, r.Decoded, "decoding failed")
			assert.Equal(t, n, r.RoundTrip, "round trip failed")
			assert.Equal(t, []interface{}{[]float32{n}}, r.Nested)
		})
	}
}

func TestSendAndReceveBytes(t *testing.T) {
	ctx := context.Background()

	samples := [][]byte{
		[]byte("abcdef"),
	}

	for i := 0; i < 1000; i++ {
		n := rand.Intn(999) + 1
		b := make([]byte, n)

		for i := 0; i < n; i++ {
			b[i] = uint8(rand.Uint32())
		}

		samples = append(samples, b)
	}

	type Result struct {
		RoundTrip []byte        `edgedb:"round_trip"`
		Nested    []interface{} `edgedb:"nested"`
	}

	query := `
		WITH b := array_unpack(<array<bytes>>$0)
		SELECT (
			round_trip := b,
			nested := ([b],),
		)
	`

	var results []Result
	err := conn.Query(ctx, query, &results, samples)
	require.Nil(t, err, "unexpected error: %v", err)
	require.Equal(t, len(samples), len(results), "wrong number of results")

	for i, b := range samples {
		t.Run(string(b), func(t *testing.T) {
			r := results[i]

			assert.Equal(t, b, r.RoundTrip)
			assert.Equal(t, []interface{}{[][]byte{b}}, r.Nested)
		})
	}
}

func TestSendAndReceveString(t *testing.T) {
	ctx := context.Background()

	query := `
		WITH
			s := <str>$0,
		SELECT (
			round_trip := s,
			nested := ([s],),
		)
	`

	type Result struct {
		RoundTrip string        `edgedb:"round_trip"`
		Nested    []interface{} `edgedb:"nested"`
	}

	sample := "abcdef"

	var result Result
	err := conn.QueryOne(ctx, query, &result, sample)
	require.Nil(t, err, "unexpected error: %v", err)

	assert.Equal(t, sample, result.RoundTrip, "round trip failed")
	assert.Equal(t, []interface{}{[]string{sample}}, result.Nested)
}

func TestSendAndReceveJSON(t *testing.T) {
	ctx := context.Background()

	strings := []string{"123", "-3.14", "true", "false", "[1, 2, 3]", "null"}

	samples := make([][]byte, len(strings))
	for i, s := range strings {
		samples[i] = []byte(s)
	}

	type Result struct {
		RoundTrip []byte        `edgedb:"round_trip"`
		Nested    []interface{} `edgedb:"nested"`
	}

	query := `
		WITH j := array_unpack(<array<json>>$0)
		SELECT (
			round_trip := j,
			nested := ([j],),
		)
	`

	var results []Result
	err := conn.Query(ctx, query, &results, samples)
	require.Nil(t, err, "unexpected error: %v", err)
	require.Equal(t, len(samples), len(results), "wrong number of results")

	for i, s := range strings {
		t.Run(s, func(t *testing.T) {
			b := samples[i]
			r := results[i]

			assert.Equal(t, b, r.RoundTrip)
			assert.Equal(t, []interface{}{[][]byte{b}}, r.Nested)
		})
	}
}

func TestSendAndReceveEnum(t *testing.T) {
	ctx := context.Background()

	type Result struct {
		Encoded   string        `edgedb:"encoded"`
		Decoded   string        `edgedb:"decoded"`
		RoundTrip string        `edgedb:"round_trip"`
		IsEqual   bool          `edgedb:"is_equal"`
		Nested    []interface{} `edgedb:"nested"`
		String    string        `edgedb:"string"`
	}

	query := `
		WITH
			e := <ColorEnum>$0,
			s := <str>$1
		SELECT (
			encoded := <str>e,
			decoded := <ColorEnum>s,
			round_trip := e,
			is_equal := <ColorEnum>s = e,
			nested := ([e],),
			string := <str><ColorEnum>s
		)
	`

	var result Result
	color := "Red"
	err := conn.QueryOne(ctx, query, &result, color, color)
	require.Nil(t, err, "unexpected error: %v", err)

	assert.Equal(t, color, result.Encoded, "encoding failed")
	assert.Equal(t, color, result.Decoded, "decoding failed")
	assert.Equal(t, color, result.RoundTrip, "round trip failed")
	assert.True(t, result.IsEqual, "equality failed")
	assert.Equal(t, color, result.String)
	assert.Equal(t, []interface{}{[]string{color}}, result.Nested)

	query = "SELECT (decoded := <ColorEnum><str>$0)"
	err = conn.QueryOne(ctx, query, &result, "invalid")

	expected := "edgedb.InvalidValueError: " +
		"invalid input value for enum 'default::ColorEnum': \"invalid\""
	assert.EqualError(t, err, expected)
}

func TestSendAndReceveDuration(t *testing.T) {
	ctx := context.Background()

	durations := []Duration{
		Duration(0),
		Duration(-1),
		Duration(86400000000),
		Duration(1_000_000),
		Duration(3074457345618258432),
	}

	strings := []string{
		"00:00:00",
		"-00:00:00.000001",
		"24:00:00",
		"00:00:01",
		"854015929:20:18.258432",
	}

	for i := 0; i < 1000; i++ {
		d := Duration(rand.Int63n(86400000000))
		durations = append(durations, d)
		strings = append(strings, d.String())
	}

	type Result struct {
		Decoded   Duration      `edgedb:"decoded"`
		RoundTrip Duration      `edgedb:"round_trip"`
		IsEqual   bool          `edgedb:"is_equal"`
		Nested    []interface{} `edgedb:"nested"`
	}

	query := `
		WITH
			sample := (
				WITH
					d := enumerate(array_unpack(<array<duration>>$0)),
					s := enumerate(array_unpack(<array<str>>$1)),
				SELECT (
					d := d.1,
					str := s.1,
				)
				FILTER d.0 = s.0
			)
		SELECT (
			decoded := <duration>sample.str,
			round_trip := sample.d,
			is_equal := <duration>sample.str = sample.d,
			nested := ([sample.d],),
		)
	`

	var results []Result
	err := conn.Query(ctx, query, &results, durations, strings)
	require.Nil(t, err, "unexpected error: %v", err)
	require.Equal(t, len(durations), len(results), "wrong number of results")

	for i, s := range strings {
		t.Run(s, func(t *testing.T) {
			d := durations[i]
			result := results[i]
			assert.True(t, result.IsEqual, "equality check faild")
			assert.Equal(t, d, result.RoundTrip, "round trip failed")
			assert.Equal(t, d, result.Decoded, "decoding failed")
			assert.Equal(t,
				[]interface{}{[]Duration{d}},
				result.Nested,
				"nested value failed",
			)
		})
	}
}

func TestSendAndReceveLocalTime(t *testing.T) {
	ctx := context.Background()

	times := []LocalTime{
		NewLocalTime(0, 0, 0, 0),
		NewLocalTime(0, 0, 0, 1),
		NewLocalTime(0, 0, 0, 10),
		NewLocalTime(0, 0, 0, 100),
		NewLocalTime(0, 0, 0, 1000),
		NewLocalTime(0, 0, 0, 10000),
		NewLocalTime(0, 0, 0, 100000),
		NewLocalTime(0, 0, 0, 123456),
		NewLocalTime(0, 1, 11, 340000),
		NewLocalTime(5, 4, 3, 0),
		NewLocalTime(11, 12, 13, 0),
		NewLocalTime(20, 39, 57, 0),
		NewLocalTime(23, 59, 59, 999000),
		NewLocalTime(23, 59, 59, 999999),
	}

	for i := 0; i < 1_000; i++ {
		times = append(times, NewLocalTime(
			rand.Intn(24),
			rand.Intn(60),
			rand.Intn(60),
			rand.Intn(1_000_000),
		))
	}

	strings := make([]string, len(times))
	for i, t := range times {
		strings[i] = t.String()
	}

	type Result struct {
		Encoded   string        `edgedb:"encoded"`
		Decoded   LocalTime     `edgedb:"decoded"`
		RoundTrip LocalTime     `edgedb:"round_trip"`
		IsEqual   bool          `edgedb:"is_equal"`
		Nested    []interface{} `edgedb:"nested"`
		String    string        `edgedb:"string"`
	}

	query := `
		WITH
			x := (
				WITH
					t := enumerate(array_unpack(<array<cal::local_time>>$0)),
					s := enumerate(array_unpack(<array<str>>$1)),
				SELECT (
					t := t.1,
					s := s.1,
				)
				FILTER t.0 = s.0
			)
		SELECT (
			encoded := <str>x.t,
			decoded := <cal::local_time>x.s,
			round_trip := x.t,
			is_equal := <cal::local_time>x.s = x.t,
			nested := ([x.t],),
			string := <str><cal::local_time><str>x.s,
		)
	`

	var results []Result
	err := conn.Query(ctx, query, &results, times, strings)
	require.Nil(t, err, "unexpected error: %v", err)

	for i, s := range strings {
		t.Run(s, func(t *testing.T) {
			time := times[i]
			r := results[i]

			assert.Equal(t, time, r.RoundTrip, "round trip failed")
			assert.Equal(t, time, r.Decoded, "decode is wrong")
			assert.Equal(t, s, r.Encoded, "encode is wrong")
			assert.True(t, r.IsEqual, "equality failed")
			assert.Equal(t, s, r.String)
			assert.Equal(t, []interface{}{[]LocalTime{time}}, r.Nested)
		})
	}
}

func TestSendAndReceveLocalDate(t *testing.T) {
	ctx := context.Background()

	dates := []LocalDate{
		NewLocalDate(1, 1, 1),
		NewLocalDate(2000, 1, 1),
		NewLocalDate(2019, 5, 6),
		NewLocalDate(4444, 12, 30),
		NewLocalDate(9999, 9, 9),
	}

	for i := 0; i < 1_000; i++ {
		dates = append(dates, NewLocalDate(
			rand.Intn(9999)+1,
			time.Month(rand.Intn(12)+1),
			rand.Intn(30)+1,
		))
	}

	strings := make([]string, len(dates))
	for i, d := range dates {
		strings[i] = d.String()
	}

	type Result struct {
		Encoded   string        `edgedb:"encoded"`
		Decoded   LocalDate     `edgedb:"decoded"`
		RoundTrip LocalDate     `edgedb:"round_trip"`
		IsEqual   bool          `edgedb:"is_equal"`
		Nested    []interface{} `edgedb:"nested"`
		String    string        `edgedb:"string"`
	}

	query := `
		WITH
			x := (
				WITH
					d := enumerate(array_unpack(<array<cal::local_date>>$0)),
					s := enumerate(array_unpack(<array<str>>$1)),
				SELECT (
					d := d.1,
					s := s.1,
				)
				FILTER d.0 = s.0
			)
		SELECT (
			encoded := <str>x.d,
			decoded := <cal::local_date>x.s,
			round_trip := x.d,
			is_equal := <cal::local_date>x.s = x.d,
			nested := ([x.d],),
			string := <str><cal::local_date>x.s,
		)
	`

	var results []Result
	err := conn.Query(ctx, query, &results, dates, strings)
	require.Nil(t, err, "unexpected error: %v", err)
	require.Equal(t, len(dates), len(results))

	for i, s := range strings {
		t.Run(s, func(t *testing.T) {
			d := dates[i]
			r := results[i]

			assert.Equal(t, d, r.RoundTrip, "round trip failed")
			assert.Equal(t, d, r.Decoded, "decode is wrong")
			assert.Equal(t, s, r.Encoded, "encode is wrong")
			assert.True(t, r.IsEqual, "equality failed")
			assert.Equal(t, s, r.String)
			assert.Equal(t, []interface{}{[]LocalDate{d}}, r.Nested)
		})
	}
}

func TestSendAndReceveLocalDateTime(t *testing.T) {
	ctx := context.Background()

	datetimes := []LocalDateTime{
		NewLocalDateTime(2019, 5, 6, 12, 0, 0, 0),
		NewLocalDateTime(2018, 5, 7, 15, 1, 22, 306916),
		NewLocalDateTime(1, 1, 1, 1, 1, 0, 0),
		NewLocalDateTime(9999, 9, 9, 9, 9, 9, 0),
	}

	for i := 0; i < 1_000; i++ {
		dt := NewLocalDateTime(
			rand.Intn(9999)+1,
			time.Month(rand.Intn(12))+1,
			rand.Intn(30)+1,
			rand.Intn(24),
			rand.Intn(60),
			rand.Intn(60),
			rand.Intn(1_000_000),
		)

		datetimes = append(datetimes, dt)
	}

	strings := make([]string, len(datetimes))
	for i, t := range datetimes {
		strings[i] = t.String()
	}

	type Result struct {
		Encoded   string        `edgedb:"encoded"`
		Decoded   LocalDateTime `edgedb:"decoded"`
		RoundTrip LocalDateTime `edgedb:"round_trip"`
		IsEqual   bool          `edgedb:"is_equal"`
		Nested    []interface{} `edgedb:"nested"`
		String    string        `edgedb:"string"`
	}

	query := `
		WITH
			x := (
				WITH
					dt := enumerate(array_unpack(
						<array<cal::local_datetime>>$0
					)),
					s := enumerate(array_unpack(<array<str>>$1)),
				SELECT (
					dt := dt.1,
					s := s.1,
				)
				FILTER dt.0 = s.0
			)
		SELECT (
			encoded := <str>x.dt,
			decoded := <cal::local_datetime>x.s,
			round_trip := x.dt,
			is_equal := <cal::local_datetime>x.s = x.dt,
			nested := ([x.dt],),
			string := <str><cal::local_datetime>x.s,
		)
	`

	var results []Result
	err := conn.Query(ctx, query, &results, datetimes, strings)
	require.Nil(t, err, "unexpected error: %v", err)
	require.Equal(t, len(datetimes), len(results), "wrong number of results")

	for i, s := range strings {
		t.Run(s, func(t *testing.T) {
			dt := datetimes[i]
			r := results[i]

			assert.True(t, r.IsEqual, "equality check faild")
			assert.Equal(t, s, r.Encoded, "encoding failed")
			assert.Equal(t, dt, r.Decoded)
			assert.Equal(t, dt, r.RoundTrip)
			assert.Equal(t, s, r.String)
			assert.Equal(t, []interface{}{[]LocalDateTime{dt}}, r.Nested)
		})
	}
}

func TestSendAndReceveDateTime(t *testing.T) {
	ctx := context.Background()
	format := "2006-01-02T15:04:05.999999-07:00"

	samples := []time.Time{
		time.Date(2019, 5, 6, 12, 0, 0, 0, time.UTC),
		time.Date(1986, 4, 26, 1, 23, 40, 1_000, time.FixedZone("", -25200)),
		time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(9999, 9, 9, 9, 9, 0, 0, time.FixedZone("", 32400)),
	}

	const maxDate = 253402300799
	const minDate = -62135596800

	for i := 0; i < 1000; i++ {
		samples = append(samples, time.Unix(
			rand.Int63n(maxDate-minDate)+minDate,
			1_000*rand.Int63n(1_000_000),
		))
	}

	strings := make([]string, len(samples))
	for i, t := range samples {
		strings[i] = t.UTC().Format(format)
	}

	type Result struct {
		Encoded   string        `edgedb:"encoded"`
		Decoded   time.Time     `edgedb:"decoded"`
		RoundTrip time.Time     `edgedb:"round_trip"`
		IsEqual   bool          `edgedb:"is_equal"`
		Nested    []interface{} `edgedb:"nested"`
		String    string        `edgedb:"string"`
	}

	query := `
		WITH
			x := (
				WITH
					dt := enumerate(array_unpack(<array<datetime>>$0)),
					s := enumerate(array_unpack(<array<str>>$1)),
				SELECT (
					dt := dt.1,
					s := s.1,
				)
				FILTER dt.0 = s.0
			)
		SELECT (
			encoded := <str>x.dt,
			decoded := <datetime>x.s,
			round_trip := x.dt,
			is_equal := <datetime>x.s = x.dt,
			nested := ([x.dt],),
			string := <str><datetime>x.s,
		)
	`

	var results []Result
	err := conn.Query(ctx, query, &results, samples, strings)
	require.Nil(t, err, "unexpected error: %v", err)
	require.Equal(t, len(samples), len(results), "wrong number of results")

	for i, s := range strings {
		t.Run(s, func(t *testing.T) {
			dt := samples[i].UTC()
			r := results[i]

			assert.True(t, r.IsEqual, "equality check faild: %v", dt.Unix())
			assert.Equal(t, s, r.Encoded, "encoding failed")
			assert.Equal(t, s, r.String, "string failed")
			assert.True(t,
				dt.Equal(r.Decoded),
				"decoding failed: %v != %v", dt, r.Decoded,
			)
			assert.True(t,
				dt.Equal(r.RoundTrip),
				"round trip failed: %v != %v", dt, r.RoundTrip,
			)

			// equivalent time.Time structs are not always ==
			// unpack the data structure to use time.Time.Equal()
			assert.Equal(t, 1, len(r.Nested))
			nested, ok := r.Nested[0].([]time.Time)
			assert.True(t, ok)
			assert.Equal(t, 1, len(nested))
			assert.True(t, dt.Equal(nested[0]))
		})
	}
}

func TestSendAndReceveBigInt(t *testing.T) {
	ctx := context.Background()

	query := `
		WITH
			i := <bigint>$0,
			s := <str>$1
		SELECT (
			encoded := <str>i,
			decoded := <bigint>s,
			round_trip := i,
			is_equal := <bigint>s = i,
			nested := ([i],),
			string := <str><bigint>s,
		)
	`

	type Result struct {
		Encoded   string        `edgedb:"encoded"`
		Decoded   *big.Int      `edgedb:"decoded"`
		RoundTrip *big.Int      `edgedb:"round_trip"`
		IsEqual   bool          `edgedb:"is_equal"`
		Nested    []interface{} `edgedb:"nested"`
		String    string        `edgedb:"string"`
	}

	samples := []string{
		"0",
		"1",
		"-1",
		"11",
		"-11",
		"123",
		"-123",
		"123789",
		"-123789",
		"19876",
		"-19876",
		"19876",
		"-19876",
		"11001200000031231238172638172637981268371628312300000000",
		"-11001231231238172638172637981268371628312300",
		"198761239812739812739801279371289371932",
		"-198761182763908473812974620938742386",
		"98761239812739812739801279371289371932",
		"-98761182763908473812974620938742386",
		"8761239812739812739801279371289371932",
		"-8761182763908473812974620938742386",
		"761239812739812739801279371289371932",
		"-761182763908473812974620938742386",
		"61239812739812739801279371289371932",
		"-61182763908473812974620938742386",
		"1239812739812739801279371289371932",
		"-1182763908473812974620938742386",
		"9812739812739801279371289371932",
		"-3908473812974620938742386",
		"98127373373209",
		"-4620938742386",
		"100000000000",
		"-100000000000",
		"10000000000",
		"-10000000000",
		"1000000000",
		"-1000000000",
		"100000000",
		"-100000000",
		"10000000",
		"-10000000",
		"1000000",
		"-1000000",
		"100000",
		"-100000",
		"10000",
		"-10000",
		"1000",
		"-1000",
		"100",
		"-100",
		"10",
		"-10",
		"100030000010",
		"-100000600004",
		"10000000100",
		"-10030000000",
		"1000040000",
		"-1000000000",
		"1010000001",
		"-1000000001",
		"1001001000",
		"-10000099",
		"99999",
		"9999",
		"999",
		"1011",
		"1009",
		"1709",
	}

	// Generate random bigints
	for i := 0; i < 1000; i++ {
		n := rand.Intn(30) + 1
		num := make([]byte, n)

		for j := 0; j < n; j++ {
			num[j] = "0123456789"[rand.Intn(10)]
		}

		t := strings.TrimLeft(string(num), "0")
		if t == "" {
			continue
		}

		// 33% chance for a negative number
		if rand.Intn(3) == 0 {
			t = "-" + t
		}

		samples = append(samples, t)
	}

	// Generate more random bigints consisting from mostly 0s
	for i := 0; i < 1000; i++ {
		n := rand.Intn(50) + 1
		num := make([]byte, n)

		for j := 0; j < n; j++ {
			k := rand.Intn(10)
			num[j] = "00000000000000000000000000000000000123456789"[k]
		}

		t := strings.TrimLeft(string(num), "0")
		if t == "" {
			continue
		}

		// 33% chance for a negative number
		if rand.Intn(3) == 0 {
			t = "-" + t
		}

		samples = append(samples, t)
	}

	for _, s := range samples {
		t.Run(s, func(t *testing.T) {
			i, ok := (&big.Int{}).SetString(s, 10)
			require.True(t, ok, "invalid big.Int literal: %v", s)
			require.Equal(t, s, i.String())

			var result Result
			err := conn.QueryOne(ctx, query, &result, i, s)
			assert.Nil(t, err, "unexpected error: %v", err)

			assert.True(t, result.IsEqual, "equality check faild")
			assert.Equal(t, s, result.Encoded, "encoding failed")
			assert.Equal(t, i, result.Decoded)
			assert.Equal(t, i, result.RoundTrip)
			assert.Equal(t, s, result.String)
			assert.Equal(t, []interface{}{[]*big.Int{i}}, result.Nested)
			require.Equal(t, s, i.String(), "argument was mutated")
		})
	}
}

func TestSendAndReceveUUID(t *testing.T) {
	ctx := context.Background()

	query := `
		WITH
			id := <uuid>$0,
			s := <str>$1
		SELECT (
			encoded := <str>id,
			decoded := <uuid>s,
			round_trip := id,
			is_equal := <uuid>s = id,
			nested := ([id],),
			string := <str><uuid>s,
		)
	`

	type Result struct {
		Encoded   string        `edgedb:"encoded"`
		Decoded   UUID          `edgedb:"decoded"`
		RoundTrip UUID          `edgedb:"round_trip"`
		IsEqual   bool          `edgedb:"is_equal"`
		Nested    []interface{} `edgedb:"nested"`
		String    string        `edgedb:"string"`
	}

	samples := []string{
		"759637d8-6635-11e9-b9d4-098002d459d5",
		"00000000-0000-0000-0000-000000000000",
		"ffffffff-ffff-ffff-ffff-ffffffffffff",
	}

	for i := 0; i < 1000; i++ {
		var id UUID
		binary.BigEndian.PutUint64(id[:8], rand.Uint64())
		binary.BigEndian.PutUint64(id[8:], rand.Uint64())
		samples = append(samples, id.String())
	}

	for _, s := range samples {
		t.Run(s, func(t *testing.T) {
			var id UUID
			err := id.UnmarshalText([]byte(s))
			require.Nil(t, err)

			var result Result
			err = conn.QueryOne(ctx, query, &result, id, s)
			assert.Nil(t, err, "unexpected error: %v", err)

			assert.True(t, result.IsEqual, "equality check faild")
			assert.Equal(t, s, result.Encoded, "encoding failed")
			assert.Equal(t, id, result.Decoded)
			assert.Equal(t, id, result.RoundTrip)
			assert.Equal(t, s, result.String)
			assert.Equal(t, []interface{}{[]UUID{id}}, result.Nested)
			require.Equal(t, s, id.String(), "argument was mutated")
		})
	}
}

func TestSendAndReceveCustomScalars(t *testing.T) {
	ctx := context.Background()

	query := `
		WITH
			i := <CustomInt64>$0,
			s := <str>$1,
		SELECT (
			encoded := <str>i,
			decoded := <CustomInt64>s,
			round_trip := i,
			is_equal := i = <CustomInt64>s,
			nested := ([i],),
		)
	`

	type Result struct {
		Encoded   string        `edgedb:"encoded"`
		Decoded   int64         `edgedb:"decoded"`
		RoundTrip int64         `edgedb:"round_trip"`
		IsEqual   bool          `edgedb:"is_equal"`
		Nested    []interface{} `edgedb:"nested"`
	}

	samples := []int64{0, 1, 9223372036854775807, -9223372036854775808}

	for i := 0; i < 1000; i++ {
		samples = append(samples, int64(rand.Uint64()))
	}

	for _, i := range samples {
		s := fmt.Sprint(i)
		t.Run(s, func(t *testing.T) {
			var result Result
			err := conn.QueryOne(ctx, query, &result, i, s)

			assert.Nil(t, err, "unexpected error: %v", err)
			assert.Equal(t, s, result.Encoded)
			assert.Equal(t, i, result.Decoded)
			assert.Equal(t, i, result.Decoded)
			assert.True(t, result.IsEqual)
			assert.Equal(t, []interface{}{[]int64{i}}, result.Nested)
		})
	}
}

func TestDecodeDeeplyNestedTuple(t *testing.T) {
	ctx := context.Background()
	query := "SELECT ([(1, 2), (3, 4)], (5, (6, 7)))"

	var result []interface{}
	err := conn.QueryOne(ctx, query, &result)
	require.Nil(t, err, "unexpected error: %v", err)

	expected := []interface{}{
		[][]interface{}{
			{int64(1), int64(2)},
			{int64(3), int64(4)},
		},
		[]interface{}{
			int64(5),
			[]interface{}{int64(6), int64(7)},
		},
	}

	assert.Equal(t, expected, result)
}
