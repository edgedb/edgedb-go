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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReleasePoolConn(t *testing.T) {
	p := &pool{freeConns: make(chan *reconnectingConn, 1)}
	conn := &reconnectingConn{}
	pConn := &poolConn{pool: p, conn: conn}

	err := pConn.Release()
	require.Nil(t, err)

	result := <-p.freeConns
	assert.Equal(t, conn, result)

	err = pConn.Release()
	assert.EqualError(
		t,
		err,
		"edgedb.InterfaceError: connection released more than once",
	)

	var edbErr Error
	require.True(t, errors.As(err, &edbErr))
	assert.True(t, edbErr.Category(InterfaceError))
}

func TestPoolConnectionRejectsTransaction(t *testing.T) {
	ctx := context.Background()
	p, err := Connect(ctx, opts)
	require.Nil(t, err)
	defer p.Close() // nolint:errcheck

	con, err := p.Acquire(ctx)
	require.Nil(t, err)
	defer con.Release() // nolint:errcheck

	expected := "edgedb.DisabledCapabilityError: " +
		"cannot execute transaction control commands"

	err = con.Execute(ctx, "START TRANSACTION")
	assert.EqualError(t, err, expected)

	var result []byte
	err = con.Query(ctx, "START TRANSACTION", &result)
	assert.EqualError(t, err, expected)

	err = con.QueryJSON(ctx, "START TRANSACTION", &result)
	assert.EqualError(t, err, expected)

	err = con.QueryOne(ctx, "START TRANSACTION", &result)
	assert.EqualError(t, err, expected)

	err = con.QueryOneJSON(ctx, "START TRANSACTION", &result)
	assert.EqualError(t, err, expected)
}
