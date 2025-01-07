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

package gel

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	types "github.com/edgedb/edgedb-go/internal/edgedbtypes"
	"github.com/stretchr/testify/assert"
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

func TestWithConfigInTx(t *testing.T) {
	if protocolVersion.LT(protocolVersion1p0) {
		t.Skip()
	}

	ctx := context.Background()

	err := client.Tx(ctx, func(ctx context.Context, tx *Tx) error {
		var id types.UUID
		_, e := rnd.Read(id[:])
		assert.NoError(t, e)

		e = tx.Execute(ctx, `insert User { id := <uuid>$0 }`, id)
		assert.True(t, strings.HasPrefix(
			e.Error(),
			"edgedb.QueryError: cannot assign to property 'id'",
		))

		return errors.New("rollback")
	})
	assert.EqualError(t, err, "rollback")

	c := client.WithConfig(map[string]interface{}{
		"allow_user_specified_id": true,
	})

	var id types.UUID
	_, e := rnd.Read(id[:])
	assert.NoError(t, e)

	// todo: remove this Execute query after
	// https://github.com/edgedb/edgedb/issues/4816
	// is resolved
	e = c.Execute(ctx, `insert User { id := <uuid>$0 }`, id)
	assert.NoError(t, e)

	err = c.Tx(ctx, func(ctx context.Context, tx *Tx) error {
		var id types.UUID
		_, e := rnd.Read(id[:])
		assert.NoError(t, e)

		e = tx.Execute(ctx, `insert User { id := <uuid>$0 }`, id)
		assert.NoError(t, e)

		return errors.New("rollback")
	})
	assert.EqualError(t, err, "rollback")
}
