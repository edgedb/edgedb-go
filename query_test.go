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

func TestNamedQueryArguments(t *testing.T) {
	ctx := context.Background()
	var result [][]int64
	err := client.Query(
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
	err := client.Query(
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
	result, err := client.QueryJSON(
		ctx,
		"SELECT {(a := 0, b := <int64>$0), (a := 42, b := <int64>$1)}",
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
	result, err := client.QueryOneJSON(
		ctx,
		"SELECT (a := 0, b := <int64>$0)",
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
	result, err := client.QueryOneJSON(ctx, "SELECT <int64>{}")

	require.Equal(t, err, ErrorZeroResults)
	assert.Equal(t, []byte(nil), result)
}

func TestQueryOne(t *testing.T) {
	ctx := context.Background()
	var result int64
	err := client.QueryOne(ctx, "SELECT 42", &result)

	assert.Nil(t, err)
	assert.Equal(t, int64(42), result)
}

func TestQueryOneZeroResults(t *testing.T) {
	ctx := context.Background()
	var result int64
	err := client.QueryOne(ctx, "SELECT <int64>{}", &result)

	assert.Equal(t, ErrorZeroResults, err)
}

func TestError(t *testing.T) {
	ctx := context.Background()
	err := client.Execute(ctx, "malformed query;")
	expected := errors.New("Unexpected 'malformed'")
	assert.Equal(t, expected, err)
}

func TestQueryTimesOut(t *testing.T) {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now())
	err := client.Execute(ctx, "SELECT 1;")

	assert.True(t, errors.Is(err, os.ErrDeadlineExceeded))
	cancel()
}
