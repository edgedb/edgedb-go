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
	"fmt"
	"log"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTxRollesBack(t *testing.T) {
	ctx := context.Background()
	err := client.Tx(ctx, func(ctx context.Context, tx *Tx) error {
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
	err = client.Query(ctx, query, &testNames)

	require.NoError(t, err)
	require.Equal(t, 0, len(testNames), "The transaction wasn't rolled back")
}

func TestTxRollesBackOnUserError(t *testing.T) {
	ctx := context.Background()
	err := client.Tx(ctx, func(ctx context.Context, tx *Tx) error {
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
	err = client.Query(ctx, query, &testNames)

	require.NoError(t, err)
	require.Equal(t, 0, len(testNames), "The transaction wasn't rolled back")
}

func TestTxCommits(t *testing.T) {
	ctx := context.Background()
	err := client.Tx(ctx, func(ctx context.Context, tx *Tx) error {
		return tx.Execute(ctx, "INSERT TxTest {name := 'Test Commit'};")
	})
	require.NoError(t, err)

	query := `
		SELECT (
			SELECT TxTest {name}
			FILTER .name = 'Test Commit'
		).name
		LIMIT 1
	`

	var testNames []string
	err = client.Query(ctx, query, &testNames)

	require.NoError(t, err)
	require.Equal(
		t,
		[]string{"Test Commit"},
		testNames,
		"The transaction wasn't commited",
	)
}

func newTxOpts(level IsolationLevel, readOnly, deferrable bool) TxOptions {
	return NewTxOptions().
		WithIsolation(level).
		WithReadOnly(readOnly).
		WithDeferrable(deferrable)
}

func TestTxKinds(t *testing.T) {
	ctx := context.Background()

	combinations := []TxOptions{
		newTxOpts(Serializable, true, true),
		newTxOpts(Serializable, true, false),
		newTxOpts(Serializable, false, true),
		newTxOpts(Serializable, false, false),
		NewTxOptions().WithIsolation(Serializable).WithReadOnly(true),
		NewTxOptions().WithIsolation(Serializable).WithReadOnly(false),
		NewTxOptions().WithIsolation(Serializable).WithDeferrable(true),
		NewTxOptions().WithIsolation(Serializable).WithDeferrable(false),
		NewTxOptions().WithReadOnly(true).WithDeferrable(true),
		NewTxOptions().WithReadOnly(true).WithDeferrable(false),
		NewTxOptions().WithReadOnly(false).WithDeferrable(true),
		NewTxOptions().WithReadOnly(false).WithDeferrable(false),
		NewTxOptions().WithIsolation(Serializable),
		NewTxOptions().WithReadOnly(true),
		NewTxOptions().WithReadOnly(false),
		NewTxOptions().WithDeferrable(true),
		NewTxOptions().WithDeferrable(false),
	}

	noOp := func(ctx context.Context, tx *Tx) error { return nil }

	for _, opts := range combinations {
		name := fmt.Sprintf("%#v", opts)

		t.Run(name, func(t *testing.T) {
			p := client.WithTxOptions(opts)
			require.NoError(t, p.Tx(ctx, noOp))
		})
	}
}

// Transactions can be executed using the Tx() method. Note that queries are
// executed on the Tx object. Queries executed on the client in a transaction
// callback will not run in the transaction and will be applied immediately. In
// edgedb-go the callback may be re-run if any of the queries fail in a way
// that might succeed on subsequent attempts. Transaction behavior can be
// configured with TxOptions and the retrying behavior can be configured with
// RetryOptions.
func ExampleTx() {
	ctx := context.Background()
	err := client.Tx(ctx, func(ctx context.Context, tx *Tx) error {
		return tx.Execute(ctx, "INSERT User { name := 'Don' }")
	})
	if err != nil {
		log.Println(err)
	}
}
