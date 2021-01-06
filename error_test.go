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
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewErrorFromCodeAs(t *testing.T) {
	msg := "example error message"
	err := &duplicateCastDefinitionError{msg: msg}
	require.NotNil(t, err)

	msg = "edgedb.DuplicateCastDefinitionError: " + msg

	var cast DuplicateCastDefinitionError
	assert.True(t, errors.As(err, &cast))
	assert.Equal(t, msg, cast.Error())

	var def DuplicateDefinitionError
	assert.True(t, errors.As(err, &def))
	assert.Equal(t, msg, def.Error())

	var schem SchemaDefinitionError
	assert.True(t, errors.As(err, &schem))
	assert.Equal(t, msg, schem.Error())

	var query QueryError
	assert.True(t, errors.As(err, &query))
	assert.Equal(t, msg, query.Error())

	var base Error
	assert.True(t, errors.As(err, &base))
	assert.Equal(t, msg, query.Error())
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

	var bin BinaryProtocolError
	require.True(t, errors.As(err, &bin), "errors.As failed")
	assert.Equal(t, "edgedb.BinaryProtocolError: bad bits!", bin.Error())

	var proto ProtocolError
	require.True(t, errors.As(err, &proto))
	assert.Equal(t, "edgedb.BinaryProtocolError: bad bits!", proto.Error())

	var val InvalidValueError
	require.True(t, errors.As(err, &val))
	assert.Equal(t, "edgedb.InvalidValueError: guess again...", val.Error())

	var exe ExecutionError
	require.True(t, errors.As(err, &exe))
	assert.Equal(t, "edgedb.InvalidValueError: guess again...", exe.Error())
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
