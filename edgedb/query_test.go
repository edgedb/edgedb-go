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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNamedQueryArguments(t *testing.T) {
	result := [][]int64{}
	err := client.Query(
		"SELECT [<int64>$first, <int64>$second]",
		&result,
		map[string]interface{}{
			"first":  int64(5),
			"second": int64(8),
		},
	)

	assert.Nil(t, err)
	assert.Equal(t, [][]int64{{5, 8}}, result)
}

func TestNumberedQueryArguments(t *testing.T) {
	result := [][]int64{}
	err := client.Query(
		"SELECT [<int64>$0, <int64>$1]",
		&result,
		int64(5),
		int64(8),
	)

	assert.Nil(t, err)
	assert.Equal(t, [][]int64{{5, 8}}, result)
}

func TestQueryJSON(t *testing.T) {
	result, err := client.QueryJSON(
		"SELECT {(a := 0, b := <int64>$0), (a := 42, b := <int64>$1)}",
		int64(1),
		int64(2),
	)

	// casting to string makes error message more helpful
	// when this test fails
	actual := string(result)

	assert.Nil(t, err)
	assert.Equal(
		t,
		"[{\"a\" : 0, \"b\" : 1}, {\"a\" : 42, \"b\" : 2}]",
		actual,
	)
}

func TestQueryOneJSON(t *testing.T) {
	result, err := client.QueryOneJSON(
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
	result, err := client.QueryOneJSON("SELECT <int64>{}")

	assert.Equal(t, err, ErrorZeroResults)
	assert.Equal(t, []byte(nil), result)
}

func TestQueryOne(t *testing.T) {
	var result int64
	err := client.QueryOne("SELECT 42", &result)

	assert.Nil(t, err)
	assert.Equal(t, int64(42), result)
}

func TestQueryOneZeroResults(t *testing.T) {
	result := (*int64)(nil)
	err := client.QueryOne("SELECT <int64>{}", result)

	assert.Equal(t, ErrorZeroResults, err)
	assert.Nil(t, result)
}

func TestError(t *testing.T) {
	err := client.Execute("malformed query;")
	expected := &Error{
		Severity: 120,
		Code:     67174656,
		Message:  "Unexpected 'malformed'",
	}
	assert.Equal(t, expected, err)
}

func TestConcurrentQueries(t *testing.T) {
	eChan := make(chan error, 10)
	rChan := make(chan []byte, 10)

	for i := 0; i < 10; i++ {
		go func() {
			result, err := client.QueryOneJSON("SELECT 1;")
			eChan <- err
			rChan <- result
		}()
	}

	for i := 0; i < 20; i++ {
		select {
		case e := <-eChan:
			require.Nil(t, e)
		case r := <-rChan:
			assert.Equal(t, "1", string(r))
		}
	}

	close(eChan)
	close(rChan)
}
