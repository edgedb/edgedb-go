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

	"github.com/stretchr/testify/require"
)

func TestTxRollesBack(t *testing.T) {
	ctx := context.Background()
	err := conn.TryTx(ctx, func(ctx context.Context, tx Tx) error {
		query := "INSERT TxTest {name := 'Test Roll Back'};"
		if e := tx.Execute(ctx, query); e != nil {
			return e
		}

		return tx.Execute(ctx, "SELECT 1 / 0;")
	})

	var edbErr Error
	require.True(t, errors.As(err, &edbErr), "wrong error: %v", err)
	require.True(
		t,
		edbErr.Category(DivisionByZeroError),
		"wrong error: %v",
		err,
	)

	query := `
		SELECT (
			SELECT TxTest {name}
			FILTER .name = 'Test Roll Back'
		).name
		LIMIT 1
	`

	var testNames []string
	err = conn.Query(ctx, query, &testNames)

	require.Nil(t, err, "unexpected error: %v", err)
	require.Equal(t, 0, len(testNames), "The transaction wasn't rolled back")
}

func TestTxRollesBackOnUserError(t *testing.T) {
	ctx := context.Background()
	err := conn.TryTx(ctx, func(ctx context.Context, tx Tx) error {
		query := "INSERT TxTest {name := 'Test Roll Back'};"
		if e := tx.Execute(ctx, query); e != nil {
			return e
		}

		return errors.New("user defined error")
	})

	require.Equal(t, err, errors.New("user defined error"))

	query := `
		SELECT (
			SELECT TxTest {name}
			FILTER .name = 'Test Roll Back'
		).name
		LIMIT 1
	`

	var testNames []string
	err = conn.Query(ctx, query, &testNames)

	require.Nil(t, err, "unexpected error: %v", err)
	require.Equal(t, 0, len(testNames), "The transaction wasn't rolled back")
}

func TestTxCommits(t *testing.T) {
	ctx := context.Background()
	err := conn.TryTx(ctx, func(ctx context.Context, tx Tx) error {
		return tx.Execute(ctx, "INSERT TxTest {name := 'Test Commit'};")
	})
	require.Nil(t, err, err)

	query := `
		SELECT (
			SELECT TxTest {name}
			FILTER .name = 'Test Commit'
		).name
		LIMIT 1
	`

	var testNames []string
	err = conn.Query(ctx, query, &testNames)

	require.Nil(t, err, "unexpected error: %v", err)
	require.Equal(
		t,
		[]string{"Test Commit"},
		testNames,
		"The transaction wasn't commited",
	)
}

func TestTxCanNotUseConn(t *testing.T) {
	ctx := context.Background()
	err := conn.TryTx(ctx, func(ctx context.Context, tx Tx) error {
		var num []int64
		return conn.Query(ctx, "SELECT 7*9;", &num)
	})

	var edbErr Error
	require.True(t, errors.As(err, &edbErr), "wrong error: %v", err)
	require.True(t, edbErr.Category(InterfaceError), "wrong error: %v", err)

	expected := "edgedb.InterfaceError: " +
		"Connection is borrowed for a transaction. " +
		"Use the methods on transaction object instead."
	require.EqualError(t, err, expected)
}
