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

package edgedb

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSyntaxError(t *testing.T) {
	samples := []struct {
		query string
		err   string
	}{
		{
			"SELECT 1 2 3",
			`edgedb.EdgeQLSyntaxError: Unexpected '2'
query:1:10

SELECT 1 2 3
         ^ error`,
		},
		{
			"SELECT (foo (((1 2) 3)) 4)",
			`edgedb.EdgeQLSyntaxError: Unexpected token: <Token ICONST "2">
query:1:18

SELECT (foo (((1 2) 3)) 4)
                 ^ It appears that a ',' is missing in a tuple before '2'`,
		},
		{
			`SELECT (Foo {
				foo
				bar
			} 2);`,
			`edgedb.EdgeQLSyntaxError: Unexpected token: <Token IDENT "bar">
query:3:5

    bar
    ^ It appears that a ',' is missing in a shape before 'bar'`,
		},
	}

	for _, s := range samples {
		t.Run(s.query, func(t *testing.T) {
			var result int64
			ctx := context.Background()
			err := client.QuerySingle(ctx, s.query, &result)
			assert.EqualError(t, err, s.err)
		})
	}
}

func TestNewErrorFromCodeAs(t *testing.T) {
	msg := "example error message"
	err := &duplicateCastDefinitionError{msg: msg}
	require.NotNil(t, err)

	assert.EqualError(t, err, "edgedb.DuplicateCastDefinitionError: "+msg)

	var edbErr Error
	require.True(t, errors.As(err, &edbErr))

	assert.True(t, edbErr.Category(DuplicateCastDefinitionError))
	assert.True(t, edbErr.Category(DuplicateDefinitionError))
	assert.True(t, edbErr.Category(SchemaDefinitionError))
	assert.True(t, edbErr.Category(QueryError))
}

func TestWrapAllAs(t *testing.T) {
	err1 := &binaryProtocolError{msg: "bad bits!"}
	err2 := &invalidValueError{msg: "guess again..."}
	err := wrapAll(err1, err2)

	require.NotNil(t, err)
	assert.Equal(
		t,
		"edgedb.BinaryProtocolError: bad bits!; "+
			"edgedb.InvalidValueError: guess again...",
		err.Error(),
	)

	var bin *binaryProtocolError
	require.True(t, errors.As(err, &bin), "errors.As failed")
	assert.Equal(t, "edgedb.BinaryProtocolError: bad bits!", bin.Error())

	var val *invalidValueError
	require.True(t, errors.As(err, &val))
	assert.Equal(t, "edgedb.InvalidValueError: guess again...", val.Error())
}

func TestWrapAllIs(t *testing.T) {
	errA := errors.New("error A")
	errB := errors.New("error B")
	err := wrapAll(errA, errB)

	require.NotNil(t, err)
	assert.Equal(t, "error A; error B", err.Error())
	assert.True(t, errors.Is(err, errA))
	assert.True(t, errors.Is(err, errB))
}
