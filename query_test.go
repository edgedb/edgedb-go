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
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMissmatchedCardinality(t *testing.T) {
	ctx := context.Background()

	var result []int64
	err := conn.QueryOne(ctx, "SELECT {1, 2, 3}", &result)

	expected := "edgedb.ResultCardinalityMismatchError: " +
		"the query has cardinality MANY " +
		"which does not match the expected cardinality ONE"
	assert.EqualError(t, err, expected)
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
	}

	for _, s := range samples {
		t.Run(s.str, func(t *testing.T) {
			query := `SELECT (
				encoded := <str><datetime>$0,
				decoded := <datetime><str>$1,
				round_trip := <datetime>$0,
				is_equal := <datetime><str>$1 = <datetime>$0,
				nested := ([<datetime>$0],),
			)`

			var result Result
			err := conn.QueryOne(ctx, query, &result, s.dt, s.str)
			assert.Nil(t, err, "unexpected error: %v", err)

			assert.True(t, result.IsEqual, "equality check faild")
			assert.Equal(t, s.str, result.Encoded, "encoding failed")
			assert.True(t,
				s.dt.Equal(result.Decoded),
				"decoding failed: %v != %v", s, result.Decoded,
			)
			assert.True(t,
				s.dt.Equal(result.RoundTrip),
				"round trip failed: %v != %v", s, result.RoundTrip,
			)

			nested := result.Nested[0].([]time.Time)[0]
			assert.True(t,
				s.dt.Equal(nested),
				"nested failed: %v != %v", s, nested,
			)
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
