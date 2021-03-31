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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSubtxRollback(t *testing.T) {
	ctx := context.Background()

	insertName := func(s string) string {
		return fmt.Sprintf("INSERT TxTest {name := 'subtx %v'};", s)
	}

	err := conn.RawTx(ctx, func(ctx context.Context, tx *Tx) error {
		err := tx.Subtx(ctx, func(ctx context.Context, stx *Subtx) error {
			err := stx.Execute(ctx, insertName("rollback 1"))
			assert.Nil(t, err, "unexpected error: %v", err)

			return firstError(err, errors.New("user error 1"))
		})
		assert.EqualError(t, err, "user error 1")
		if err != nil && err.Error() != "user error 1" {
			return err
		}

		err = tx.Subtx(ctx, func(ctx context.Context, stx *Subtx) error {
			err = stx.Subtx(ctx, func(ctx context.Context, stx2 *Subtx) error {
				err = stx2.Execute(ctx, insertName("commit 1"))
				assert.Nil(t, err, "unexpected error: %v", err)
				return err
			})
			assert.Nil(t, err, "unexpected error: %v", err)
			if err != nil {
				return err
			}

			err = stx.Subtx(ctx, func(ctx context.Context, stx2 *Subtx) error {
				err = stx2.Execute(ctx, insertName("rollback 2"))
				assert.Nil(t, err, "unexpected error: %v", err)

				return firstError(err, errors.New("user error 2"))
			})
			assert.EqualError(t, err, "user error 2")
			if err != nil && err.Error() != "user error 2" {
				return err
			}

			err = stx.Execute(ctx, insertName("commit 2"))
			assert.Nil(t, err, "unexpected error: %v", err)

			return err
		})
		assert.Nil(t, err, "unexpected error: %v", err)

		return err
	})
	assert.Nil(t, err, "unexpected error: %v", err)

	var names []string
	err = conn.Query(
		ctx, `
		SELECT names := (SELECT TxTest {name}).name
		FILTER names LIKE 'subtx %'
		ORDER BY names`,
		&names,
	)
	require.Nil(t, err, "unexpected error: %v", err)

	expected := []string{
		"subtx commit 1",
		"subtx commit 2",
	}

	require.Equal(t, expected, names, "subtransaction wasn't rolled back")
}

func TestSubtxBorrowing(t *testing.T) {
	ctx := context.Background()

	noOpSubtx := func(ctx context.Context, stx *Subtx) error { return nil }
	noOpTx := func(ctx context.Context, tx *Tx) error { return nil }

	expected := "edgedb.InterfaceError: " +
		"The connection is borrowed for a transaction. " +
		"Use the methods on the transaction object instead."

	connCopy := conn.WithTxOptions(NewTxOptions())

	err := conn.RawTx(ctx, func(ctx context.Context, tx *Tx) error {
		// the connection should not be borrowable
		err := conn.Execute(ctx, "SELECT 1")
		assert.EqualError(t, err, expected)

		var result []byte
		err = conn.Query(ctx, "SELECT b''", &result)
		assert.EqualError(t, err, expected)

		err = conn.QueryOne(ctx, "SELECT b''", &result)
		assert.EqualError(t, err, expected)

		err = conn.QueryJSON(ctx, "SELECT b''", &result)
		assert.EqualError(t, err, expected)

		err = conn.QueryOneJSON(ctx, "SELECT b''", &result)
		assert.EqualError(t, err, expected)

		err = conn.RawTx(ctx, noOpTx)
		assert.EqualError(t, err, expected)

		err = conn.RetryingTx(ctx, noOpTx)
		assert.EqualError(t, err, expected)

		// copied connections should not be borrowable either
		err = connCopy.Execute(ctx, "SELECT 1")
		assert.EqualError(t, err, expected)

		err = connCopy.Query(ctx, "SELECT b''", &result)
		assert.EqualError(t, err, expected)

		err = connCopy.QueryOne(ctx, "SELECT b''", &result)
		assert.EqualError(t, err, expected)

		err = connCopy.QueryJSON(ctx, "SELECT b''", &result)
		assert.EqualError(t, err, expected)

		err = connCopy.QueryOneJSON(ctx, "SELECT b''", &result)
		assert.EqualError(t, err, expected)

		err = connCopy.RawTx(ctx, noOpTx)
		assert.EqualError(t, err, expected)

		err = connCopy.RetryingTx(ctx, noOpTx)
		assert.EqualError(t, err, expected)

		err = tx.Subtx(ctx, func(ctx context.Context, stx *Subtx) error {
			expected := "edgedb.InterfaceError: " +
				"The transaction is borrowed for a subtransaction. " +
				"Use the methods on the subtransaction object instead."

			err = tx.Execute(ctx, "SELECT 1")
			assert.EqualError(t, err, expected)

			var result []byte
			err = tx.Query(ctx, "SELECT b''", &result)
			assert.EqualError(t, err, expected)

			err = tx.QueryOne(ctx, "SELECT b''", &result)
			assert.EqualError(t, err, expected)

			err = tx.QueryJSON(ctx, "SELECT b''", &result)
			assert.EqualError(t, err, expected)

			err = tx.QueryOneJSON(ctx, "SELECT b''", &result)
			assert.EqualError(t, err, expected)

			err = tx.Subtx(ctx, noOpSubtx)
			assert.EqualError(t, err, expected)

			err = stx.Subtx(ctx, func(ctx context.Context, stx2 *Subtx) error {
				err = stx.Execute(ctx, "SELECT 1")
				assert.EqualError(t, err, expected)

				var result []byte
				err = stx.Query(ctx, "SELECT b''", &result)
				assert.EqualError(t, err, expected)

				err = stx.QuerySingle(ctx, "SELECT b''", &result)
				assert.EqualError(t, err, expected)

				err = stx.QueryJSON(ctx, "SELECT b''", &result)
				assert.EqualError(t, err, expected)

				err = stx.QuerySingleJSON(ctx, "SELECT b''", &result)
				assert.EqualError(t, err, expected)

				err = stx.Subtx(ctx, noOpSubtx)
				assert.EqualError(t, err, expected)

				return nil
			})
			assert.Nil(t, err, "unexpected error: %v", err)

			return nil
		})
		assert.Nil(t, err, "unexpected error: %v", err)

		return nil
	})
	assert.Nil(t, err, "unexpected error: %v", err)
}
