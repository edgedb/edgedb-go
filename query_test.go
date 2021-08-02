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

func TestObjectWithoutID(t *testing.T) {
	if conn.conn.protocolVersion.lt(version{0, 10}) {
		t.Skip("not expected to pass on protocol versions below 0.10")
	}

	ctx := context.Background()

	type Database struct {
		Name string `edgedb:"name"`
	}

	var result Database
	err := conn.QuerySingle(
		ctx, `
		SELECT sys::Database{ name }
		FILTER .name = 'edgedb'
		LIMIT 1`,
		&result,
	)
	assert.NoError(t, err)
	assert.Equal(t, "edgedb", result.Name)
}

func TestWrongNumberOfArguments(t *testing.T) {
	var result string
	ctx := context.Background()
	err := conn.QuerySingle(ctx, `SELECT <str>$0`, &result)
	assert.EqualError(t, err,
		"edgedb.InvalidArgumentError: expected 1 arguments got 0")
}

func TestConnRejectsTransactions(t *testing.T) {
	expected := "edgedb.DisabledCapabilityError: " +
		"cannot execute transaction control commands"

	ctx := context.Background()
	err := conn.Execute(ctx, "START TRANSACTION")
	assert.EqualError(t, err, expected)

	var result []byte
	err = conn.Query(ctx, "START TRANSACTION", &result)
	assert.EqualError(t, err, expected)

	err = conn.QueryJSON(ctx, "START TRANSACTION", &result)
	assert.EqualError(t, err, expected)

	err = conn.QuerySingle(ctx, "START TRANSACTION", &result)
	assert.EqualError(t, err, expected)

	err = conn.QuerySingleJSON(ctx, "START TRANSACTION", &result)
	assert.EqualError(t, err, expected)
}

func TestMissmatchedCardinality(t *testing.T) {
	ctx := context.Background()

	var result []int64
	err := conn.QuerySingle(ctx, "SELECT {1, 2, 3}", &result)

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
	err := conn.QuerySingle(ctx, "SELECT (x := (y := (z := 7)))", &result)

	expected := "edgedb.UnsupportedFeatureError: " +
		"the \"out\" argument does not match query schema: " +
		"expected edgedb.A.x.y.z to be int64 got int"
	assert.EqualError(t, err, expected)
}

// The client should read all messages through ReadyForCommand
// before returning from a QueryX()
func TestParseAllMessagesAfterError(t *testing.T) {
	ctx := context.Background()

	// cause error during prepare
	var number float64
	err := conn.QuerySingle(ctx, "SELECT 1 / $0", &number, int64(5))
	expected := `edgedb.QueryError: missing a type cast before the parameter
query:1:12

SELECT 1 / $0
           ^ error`
	assert.EqualError(t, err, expected)

	// cause erroy during execute
	err = conn.QuerySingle(ctx, "SELECT 1 / 0", &number)
	assert.EqualError(t, err, "edgedb.DivisionByZeroError: division by zero")

	// cache query so that it is run optimistically next time
	err = conn.QuerySingle(ctx, "SELECT 1 / <int64>$0", &number, int64(3))
	assert.NoError(t, err)

	// cause error during optimistic execute
	err = conn.QuerySingle(ctx, "SELECT 1 / <int64>$0", &number, int64(0))
	assert.EqualError(t, err, "edgedb.DivisionByZeroError: division by zero")
}

func TestArgumentTypeMissmatch(t *testing.T) {
	type Tuple struct {
		first int16 `edgedb:"0"` // nolint:unused,structcheck
	}

	var res Tuple
	ctx := context.Background()
	err := conn.QuerySingle(ctx,
		"SELECT (<int16>$0 + <int16>$1,)", &res, 1, 1111)

	require.NotNil(t, err)
	assert.EqualError(t, err,
		"edgedb.InvalidArgumentError: expected args[0] to be int16 got int")
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

	require.NoError(t, err)
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

	assert.NoError(t, err)
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

	require.NoError(t, err)
	assert.Equal(
		t,
		"[{\"a\" : 0, \"b\" : 1}, {\"a\" : 42, \"b\" : 2}]",
		actual,
	)
}

func TestQuerySingleJSON(t *testing.T) {
	ctx := context.Background()
	var result []byte
	err := conn.QuerySingleJSON(
		ctx,
		"SELECT (a := 0, b := <int64>$0)",
		&result,
		int64(42),
	)

	// casting to string makes error messages more helpful
	// when this test fails
	actual := string(result)

	assert.NoError(t, err)
	assert.Equal(t, "{\"a\" : 0, \"b\" : 42}", actual)
}

func TestQuerySingleJSONZeroResults(t *testing.T) {
	ctx := context.Background()
	var result []byte
	err := conn.QuerySingleJSON(ctx, "SELECT <int64>{}", &result)

	require.Equal(t, err, errZeroResults)
	assert.Equal(t, []byte(nil), result)
}

func TestQuerySingle(t *testing.T) {
	ctx := context.Background()
	var result int64
	err := conn.QuerySingle(ctx, "SELECT 42", &result)

	assert.NoError(t, err)
	assert.Equal(t, int64(42), result)
}

func TestQuerySingleZeroResults(t *testing.T) {
	ctx := context.Background()
	var result int64
	err := conn.QuerySingle(ctx, "SELECT <int64>{}", &result)

	assert.Equal(t, errZeroResults, err)
}

func TestError(t *testing.T) {
	ctx := context.Background()
	err := conn.Execute(ctx, "malformed query;")
	assert.EqualError(
		t,
		err,
		`edgedb.EdgeQLSyntaxError: Unexpected 'malformed'
query:1:1

malformed query;
^ error`,
	)

	var expected Error
	assert.True(t, errors.As(err, &expected))
}

func TestQueryTimesOut(t *testing.T) {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now())
	defer cancel()

	var r int64
	err := conn.QuerySingle(ctx, "SELECT 1;", &r)
	require.True(
		t,
		errors.Is(err, context.DeadlineExceeded) ||
			errors.Is(err, os.ErrDeadlineExceeded),
		err,
	)
	require.Equal(t, int64(0), r)

	err = conn.QuerySingle(context.Background(), "SELECT 2;", &r)
	require.NoError(t, err)
	assert.Equal(t, int64(2), r)
}

func TestNilResultValue(t *testing.T) {
	ctx := context.Background()
	err := conn.Query(ctx, "SELECT 1", nil)
	assert.EqualError(t, err,
		"the \"out\" argument must be a pointer, got untyped nil")
}
