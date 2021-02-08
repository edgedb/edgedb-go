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
	"errors"
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSendAndReceveUUID(t *testing.T) {
	id := UUID{
		0x75, 0x96, 0x37, 0xd8, 0x66, 0x35, 0x11, 0xe9,
		0xb9, 0xd4, 0x09, 0x80, 0x02, 0xd4, 0x59, 0xd5,
	}

	var result UUID
	ctx := context.Background()
	err := conn.QueryOne(ctx, "SELECT <uuid>$0", &result, id)

	expected := UUID{
		0x75, 0x96, 0x37, 0xd8, 0x66, 0x35, 0x11, 0xe9,
		0xb9, 0xd4, 0x09, 0x80, 0x02, 0xd4, 0x59, 0xd5,
	}

	assert.Nil(t, err, "unexpected error: %v", err)
	assert.Equal(t, expected, result)
	assert.Equal(t, expected, id, "input value was mutated")

	var nested []interface{}
	err = conn.QueryOne(ctx, "SELECT ([<uuid>$0],)", &nested, id)

	assert.Nil(t, err, "unexpected error: %v", err)
	assert.Equal(t, []interface{}{[]UUID{expected}}, nested)
	assert.Equal(t, expected, id, "input value was mutated")
}

func TestMissmatchedCardinality(t *testing.T) {
	ctx := context.Background()

	var result []int64
	err := conn.QueryOne(ctx, "SELECT {1, 2, 3}", &result)

	expected := "edgedb.ResultCardinalityMismatchError: " +
		"the query has cardinality MANY " +
		"which does not match the expected cardinality ONE"
	assert.EqualError(t, err, expected)
}

func TestFetchBigInt(t *testing.T) {
	names := []string{
		"0",
		"1",
		"-1",
		"123",
		"-123",
		"123789",
		"-123789",
		"19876",
		"-19876",
		"19876",
		"-19876",
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

	ctx := context.Background()

	for _, name := range names {
		t.Run(name, func(t *testing.T) {
			arg, ok := (&big.Int{}).SetString(name, 10)
			require.True(t, ok, "invalid big.Int literal: %v", name)
			require.Equal(t, name, arg.String())

			var result *big.Int
			err := conn.QueryOne(ctx, "SELECT <bigint>$0", &result, arg)

			require.Nil(t, err, "unexpected error: %v", err)
			require.Equal(t, name, arg.String(), "argument was mutated")
			assert.Equal(t, arg, result, "unexpected result")
		})
	}
}

func TestMissmatchedResultType(t *testing.T) {
	type C struct { // nolint:unused
		z int // nolint:structcheck
	}

	type B struct { // nolint:unused
		y C // nolint:structcheck
	}

	type A struct {
		x B // nolint:structcheck,unused
	}

	var result A

	ctx := context.Background()
	err := conn.QueryOne(ctx, "SELECT (x := (y := (z := 7)))", &result)

	expected := "edgedb.UnsupportedFeatureError: " +
		"the \"out\" argument does not match query schema: " +
		"expected edgedb.A.x.y.z to be int64 got int"
	assert.EqualError(t, err, expected)
}

func TestSendAndReceveDateTime(t *testing.T) {
	ctx := context.Background()

	samples := []struct {
		str string
		dt  time.Time
	}{
		{
			"2019-05-06T12:00:00+00:00",
			time.Date(2019, 5, 6, 12, 0, 0, 0, time.UTC),
		},
		{
			"1986-04-26T08:23:40.000001+00:00",
			time.Date(
				1986, 4, 26, 1, 23, 40, 1_000,
				time.FixedZone("", -25200),
			),
		},
		{
			"0001-01-01T00:00:00+00:00",
			time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			"9999-09-09T00:09:00+00:00",
			time.Date(9999, 9, 9, 9, 9, 0, 0, time.FixedZone("", 32400)),
		},
	}

	type Result struct {
		Encoded   string        `edgedb:"encoded"`
		Decoded   time.Time     `edgedb:"decoded"`
		RoundTrip time.Time     `edgedb:"round_trip"`
		IsEqual   bool          `edgedb:"is_equal"`
		Nested    []interface{} `edgedb:"nested"`
		String    string        `edgedb:"string"`
	}

	for _, s := range samples {
		t.Run(s.str, func(t *testing.T) {
			query := `
				WITH
					dt := <datetime>$0,
					s := <str>$1
				SELECT (
					encoded := <str>dt,
					decoded := <datetime>s,
					round_trip := dt,
					is_equal := <datetime>s = dt,
					nested := ([dt],),
					string := <str><datetime>s,
				)
			`

			var result Result
			err := conn.QueryOne(ctx, query, &result, s.dt, s.str)
			assert.Nil(t, err, "unexpected error: %v", err)

			assert.True(t, result.IsEqual, "equality check faild")
			assert.Equal(t, s.str, result.Encoded, "encoding failed")
			assert.Equal(t, s.str, result.String)
			assert.True(t,
				s.dt.Equal(result.Decoded),
				"decoding failed: %v != %v", s, result.Decoded,
			)
			assert.True(t,
				s.dt.Equal(result.RoundTrip),
				"round trip failed: %v != %v", s, result.RoundTrip,
			)

			assert.Equal(t, 1, len(result.Nested))
			nested, ok := result.Nested[0].([]time.Time)
			assert.True(t, ok)
			assert.Equal(t, 1, len(nested))
			assert.True(t,
				s.dt.Equal(nested[0]),
				"nested failed: %v != %v", s, nested,
			)
		})
	}
}

func TestSendAndReceveLocalDateTime(t *testing.T) {
	ctx := context.Background()

	samples := []struct {
		str string
		dt  LocalDateTime
	}{
		{
			str: "2019-05-06T12:00:00",
			dt:  NewLocalDateTime(2019, 5, 6, 12, 0, 0, 0),
		},
		{
			str: "2018-05-07T15:01:22.306916",
			dt:  NewLocalDateTime(2018, 5, 7, 15, 1, 22, 306916),
		},
		{
			str: "0001-01-01T01:01:00",
			dt:  NewLocalDateTime(1, 1, 1, 1, 1, 0, 0),
		},
		{
			str: "9999-09-09T09:09:09",
			dt:  NewLocalDateTime(9999, 9, 9, 9, 9, 9, 0),
		},
	}

	type Result struct {
		Encoded   string        `edgedb:"encoded"`
		Decoded   LocalDateTime `edgedb:"decoded"`
		RoundTrip LocalDateTime `edgedb:"round_trip"`
		IsEqual   bool          `edgedb:"is_equal"`
		Nested    []interface{} `edgedb:"nested"`
		String    string        `edgedb:"string"`
	}

	for _, s := range samples {
		t.Run(s.str, func(t *testing.T) {
			query := `
				WITH
					dt := <cal::local_datetime>$0,
					s := <str>$1
				SELECT (
					encoded := <str>dt,
					decoded := <cal::local_datetime>s,
					round_trip := dt,
					is_equal := <cal::local_datetime>s = dt,
					nested := ([dt],),
					string := <str><cal::local_datetime>s,
				)
			`

			var result Result
			err := conn.QueryOne(ctx, query, &result, s.dt, s.str)
			assert.Nil(t, err, "unexpected error: %v", err)

			assert.True(t, result.IsEqual, "equality check faild")
			assert.Equal(t, s.str, result.Encoded, "encoding failed")
			assert.Equal(t, s.dt, result.Decoded)
			assert.Equal(t, s.dt, result.RoundTrip)
			assert.Equal(t, s.str, result.String)

			assert.Equal(t, 1, len(result.Nested))
			nested, ok := result.Nested[0].([]LocalDateTime)
			assert.True(t, ok)
			assert.Equal(t, 1, len(nested))
			assert.Equal(t, s.dt, nested[0])
		})
	}
}

func TestSendAndReceveLocalDate(t *testing.T) {
	ctx := context.Background()

	samples := []struct {
		str  string
		date LocalDate
	}{
		{"0001-01-01", NewLocalDate(1, 1, 1)},
		{"2000-01-01", NewLocalDate(2000, 1, 1)},
		{"2019-05-06", NewLocalDate(2019, 5, 6)},
		{"4444-12-30", NewLocalDate(4444, 12, 30)},
		{"9999-09-09", NewLocalDate(9999, 9, 9)},
	}

	type Result struct {
		Encoded   string        `edgedb:"encoded"`
		Decoded   LocalDate     `edgedb:"decoded"`
		RoundTrip LocalDate     `edgedb:"round_trip"`
		IsEqual   bool          `edgedb:"is_equal"`
		Nested    []interface{} `edgedb:"nested"`
		String    string        `edgedb:"string"`
	}

	for _, s := range samples {
		t.Run(s.str, func(t *testing.T) {
			query := `
				WITH
					d := <cal::local_date>$1,
					s := <str>$0
				SELECT (
					encoded := <str>d,
					decoded := <cal::local_date>s,
					round_trip := d,
					is_equal := <cal::local_date>s = d,
					nested := ([d],),
					string := <str><cal::local_date>s,
				)
			`

			var result Result
			err := conn.QueryOne(ctx, query, &result, s.str, s.date)
			require.Nil(t, err, "unexpected error: %v", err)

			assert.Equal(t, result.RoundTrip, s.date, "round trip failed")
			assert.Equal(t, s.date, result.Decoded, "decode is wrong")
			assert.Equal(t, s.str, result.Encoded, "encode is wrong")
			assert.True(t, result.IsEqual, "equality failed")
			assert.Equal(t, s.str, result.String)

			assert.Equal(t, 1, len(result.Nested))
			nested, ok := result.Nested[0].([]LocalDate)
			assert.True(t, ok)
			assert.Equal(t, 1, len(nested))
			assert.Equal(t, s.date, nested[0])
		})
	}
}

func TestSendAndReceveLocalTime(t *testing.T) {
	ctx := context.Background()

	samples := []struct {
		str  string
		time LocalTime
	}{
		{"00:00:00", NewLocalTime(0, 0, 0, 0)},
		{"00:00:00.000001", NewLocalTime(0, 0, 0, 1)},
		{"00:00:00.00001", NewLocalTime(0, 0, 0, 10)},
		{"00:00:00.0001", NewLocalTime(0, 0, 0, 100)},
		{"00:00:00.001", NewLocalTime(0, 0, 0, 1000)},
		{"00:00:00.01", NewLocalTime(0, 0, 0, 10000)},
		{"00:00:00.1", NewLocalTime(0, 0, 0, 100000)},
		{"00:00:00.123456", NewLocalTime(0, 0, 0, 123456)},
		{"05:04:03", NewLocalTime(5, 4, 3, 0)},
		{"20:39:57", NewLocalTime(20, 39, 57, 0)},
		{"23:59:59.999999", NewLocalTime(23, 59, 59, 999999)},
	}

	type Result struct {
		Encoded   string        `edgedb:"encoded"`
		Decoded   LocalTime     `edgedb:"decoded"`
		RoundTrip LocalTime     `edgedb:"round_trip"`
		IsEqual   bool          `edgedb:"is_equal"`
		Nested    []interface{} `edgedb:"nested"`
		String    string        `edgedb:"string"`
	}

	for _, s := range samples {
		t.Run(s.str, func(t *testing.T) {
			query := `
				WITH
					t := <cal::local_time>$1,
					s := <str>$0
				SELECT (
					encoded := <str>t,
					decoded := <cal::local_time>s,
					round_trip := t,
					is_equal := <cal::local_time>s = t,
					nested := ([t],),
					string := <str><cal::local_time>s
				)
			`

			var result Result
			err := conn.QueryOne(ctx, query, &result, s.str, s.time)
			require.Nil(t, err, "unexpected error: %v", err)

			assert.Equal(t, result.RoundTrip, s.time, "round trip failed")
			assert.Equal(t, s.time, result.Decoded, "decode is wrong")
			assert.Equal(t, s.str, result.Encoded, "encode is wrong")
			assert.True(t, result.IsEqual, "equality failed")
			assert.Equal(t, s.str, result.String)

			assert.Equal(t, 1, len(result.Nested))
			nested, ok := result.Nested[0].([]LocalTime)
			assert.True(t, ok)
			assert.Equal(t, 1, len(nested))
			assert.Equal(t, s.time, nested[0])
		})
	}
}

func TestSendAndReceveDuration(t *testing.T) {
	ctx := context.Background()

	samples := []struct {
		str string
		d   Duration
	}{
		{"00:00:00", Duration(0)},
		{"-00:00:00.000001", Duration(-1)},
		{"24:00:00", Duration(86400000000)},
		{"00:00:01", Duration(1_000_000)},
		{"854015929:20:18.258432", Duration(3074457345618258432)},
	}

	type Result struct {
		Encoded   string        `edgedb:"encoded"`
		Decoded   Duration      `edgedb:"decoded"`
		RoundTrip Duration      `edgedb:"round_trip"`
		IsEqual   bool          `edgedb:"is_equal"`
		Nested    []interface{} `edgedb:"nested"`
		String    string        `edgedb:"string"`
	}

	for _, s := range samples {
		t.Run(s.str, func(t *testing.T) {
			query := `
				WITH
					d := <duration>$0,
				SELECT (
				encoded := <str>d,
				decoded := <duration><str>$1,
				round_trip := d,
				is_equal := <duration><str>$1 = d,
				nested := ([d],),
				string := <str><duration><str>$2,
			)`

			var result Result
			err := conn.QueryOne(ctx, query, &result, s.d, s.str, s.d.String())
			assert.Nil(t, err, "unexpected error: %v", err)

			assert.True(t, result.IsEqual, "equality check faild")
			assert.Equal(t, s.str, result.Encoded, "encoding failed")
			assert.Equal(t, s.d, result.Decoded, "decoding failed")
			assert.Equal(t, s.d, result.RoundTrip)
			assert.Equal(t, s.str, result.String)

			require.Equal(t, 1, len(result.Nested))
			nested, ok := result.Nested[0].([]Duration)
			require.True(t, ok, "%v", result.Nested)
			require.Equal(t, 1, len(nested))
			assert.Equal(t, s.d, nested[0])
		})
	}
}

func TestSendAndReceveJSON(t *testing.T) {
	json := []byte(`"hello"`)

	var result []byte
	ctx := context.Background()
	err := conn.QueryOne(ctx, "SELECT <json>$0", &result, json)

	expected := []byte(`"hello"`)
	assert.Nil(t, err, "unexpected error: %v", err)
	assert.Equal(t, `"hello"`, string(result))
	assert.Equal(t, expected, json, "input value was mutated")

	var nested []interface{}
	query := "SELECT ([<json>$0],)"
	err = conn.QueryOne(ctx, query, &nested, []byte(`"hello"`))

	assert.Nil(t, err, "unexpected error: %v", err)
	assert.Equal(t, []interface{}{[][]byte{expected}}, nested)
	assert.Equal(t, expected, json, "input value was mutated")
}

// The client should read all messages through ReadyForCommand
// before returning from a QueryX()
func TestParseAllMessagesAfterError(t *testing.T) {
	ctx := context.Background()

	// cause error during prepare
	var number float64
	err := conn.QueryOne(ctx, "SELECT 1 / $0", &number, int64(5))
	expected := "edgedb.QueryError: missing a type cast before the parameter"
	assert.EqualError(t, err, expected)

	// cause error during execute
	err = conn.QueryOne(ctx, "SELECT 1 / 0", &number)
	assert.EqualError(t, err, "edgedb.DivisionByZeroError: division by zero")

	// cache query so that it is run optimistically next time
	err = conn.QueryOne(ctx, "SELECT 1 / <int64>$0", &number, int64(3))
	assert.Nil(t, err)

	// cause error during optimistic execute
	err = conn.QueryOne(ctx, "SELECT 1 / <int64>$0", &number, int64(0))
	assert.EqualError(t, err, "edgedb.DivisionByZeroError: division by zero")
}

func TestArgumentTypeMissmatch(t *testing.T) {
	var res []interface{}
	ctx := context.Background()
	err := conn.QueryOne(ctx, "SELECT (<int16>$0 + <int16>$1,)", &res, 1, 1111)

	require.NotNil(t, err)
	assert.Equal(
		t,
		"edgedb.InvalidArgumentError: expected args[0] to be int16 got int",
		err.Error(),
	)
}

func TestDeeplyNestedTuple(t *testing.T) {
	var result []interface{}
	ctx := context.Background()
	query := "SELECT ([(1, 2), (3, 4)], (5, (6, 7)))"
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

func TestNamedQueryArguments(t *testing.T) {
	ctx := context.Background()
	var result [][]int64
	err := conn.Query(
		ctx,
		"SELECT [<int64>$first, <int64>$second]",
		&result,
		map[string]interface{}{
			"first":  int64(5),
			"second": int64(8),
		},
	)

	require.Nil(t, err)
	assert.Equal(t, [][]int64{{5, 8}}, result)
}

func TestNumberedQueryArguments(t *testing.T) {
	ctx := context.Background()
	result := [][]int64{}
	err := conn.Query(
		ctx,
		"SELECT [<int64>$0, <int64>$1]",
		&result,
		int64(5),
		int64(8),
	)

	assert.Nil(t, err)
	assert.Equal(t, [][]int64{{5, 8}}, result)
}

func TestQueryJSON(t *testing.T) {
	ctx := context.Background()
	var result []byte
	err := conn.QueryJSON(
		ctx,
		"SELECT {(a := 0, b := <int64>$0), (a := 42, b := <int64>$1)}",
		&result,
		int64(1),
		int64(2),
	)

	// casting to string makes error message more helpful
	// when this test fails
	actual := string(result)

	require.Nil(t, err)
	assert.Equal(
		t,
		"[{\"a\" : 0, \"b\" : 1}, {\"a\" : 42, \"b\" : 2}]",
		actual,
	)
}

func TestQueryOneJSON(t *testing.T) {
	ctx := context.Background()
	var result []byte
	err := conn.QueryOneJSON(
		ctx,
		"SELECT (a := 0, b := <int64>$0)",
		&result,
		int64(42),
	)

	// casting to string makes error messages more helpful
	// when this test fails
	actual := string(result)

	assert.Nil(t, err)
	assert.Equal(t, "{\"a\" : 0, \"b\" : 42}", actual)
}

func TestQueryOneJSONZeroResults(t *testing.T) {
	ctx := context.Background()
	var result []byte
	err := conn.QueryOneJSON(ctx, "SELECT <int64>{}", &result)

	require.Equal(t, err, errZeroResults)
	assert.Equal(t, []byte(nil), result)
}

func TestQueryOne(t *testing.T) {
	ctx := context.Background()
	var result int64
	err := conn.QueryOne(ctx, "SELECT 42", &result)

	assert.Nil(t, err)
	assert.Equal(t, int64(42), result)
}

func TestQueryOneZeroResults(t *testing.T) {
	ctx := context.Background()
	var result int64
	err := conn.QueryOne(ctx, "SELECT <int64>{}", &result)

	assert.Equal(t, errZeroResults, err)
}

func TestError(t *testing.T) {
	ctx := context.Background()
	err := conn.Execute(ctx, "malformed query;")
	assert.EqualError(
		t,
		err,
		"edgedb.EdgeQLSyntaxError: Unexpected 'malformed'",
	)

	var expected Error
	assert.True(t, errors.As(err, &expected))
}

func TestQueryTimesOut(t *testing.T) {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now())
	defer cancel()

	var r int64
	err := conn.QueryOne(ctx, "SELECT 1;", &r)
	require.True(
		t,
		errors.Is(err, context.DeadlineExceeded) ||
			errors.Is(err, os.ErrDeadlineExceeded),
		err,
	)
	require.Equal(t, int64(0), r)

	err = conn.QueryOne(context.Background(), "SELECT 2;", &r)
	require.Nil(t, err)
	assert.Equal(t, int64(2), r)
}
