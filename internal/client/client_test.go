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
	"sync"
	"testing"
	"time"
	"unsafe"

	"github.com/edgedb/edgedb-go/internal/edgedbtypes"
	types "github.com/edgedb/edgedb-go/internal/edgedbtypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConnectClient(t *testing.T) {
	ctx := context.Background()
	p, err := CreateClient(ctx, opts)
	require.NoError(t, err)

	var result string
	err = p.QuerySingle(ctx, "SELECT 'hello';", &result)
	assert.NoError(t, err)
	assert.Equal(t, "hello", result)

	p2 := p.WithTxOptions(NewTxOptions())

	err = p.Close()
	assert.NoError(t, err)

	// Client should not be closeable a second time.
	err = p.Close()
	assert.EqualError(t, err, "edgedb.InterfaceError: client closed")

	// Copied clients should be closed if a different copy is closed.
	err = p2.Close()
	assert.EqualError(t, err, "edgedb.InterfaceError: client closed")
}

func TestClientRejectsTransaction(t *testing.T) {
	ctx := context.Background()
	p, err := CreateClient(ctx, opts)
	require.NoError(t, err)

	expected := "edgedb.DisabledCapabilityError: " +
		"cannot execute transaction control commands"

	err = p.Execute(ctx, "START TRANSACTION")
	assert.EqualError(t, err, expected)

	var result []byte
	err = p.Query(ctx, "START TRANSACTION", &result)
	assert.EqualError(t, err, expected)

	err = p.QueryJSON(ctx, "START TRANSACTION", &result)
	assert.EqualError(t, err, expected)

	err = p.QuerySingle(ctx, "START TRANSACTION", &result)
	assert.EqualError(t, err, expected)

	err = p.QuerySingleJSON(ctx, "START TRANSACTION", &result)
	assert.EqualError(t, err, expected)

	err = p.Close()
	assert.NoError(t, err)
}

func TestConnectClientZeroConcurrency(t *testing.T) {
	o := opts
	o.Concurrency = 0

	ctx := context.Background()
	p, err := CreateClient(ctx, o)
	require.NoError(t, err)
	require.NoError(t, p.EnsureConnected(ctx))

	expected := client.cfg.serverSettings.
		Get("suggested_pool_concurrency").(int)
	if err != nil {
		expected = defaultConcurrency
	}
	require.Equal(t, expected, p.concurrency)

	var result string
	err = p.QuerySingle(ctx, "SELECT 'hello';", &result)
	assert.NoError(t, err)
	assert.Equal(t, "hello", result)

	err = p.Close()
	assert.NoError(t, err)
}

func TestCloseClientConcurently(t *testing.T) {
	ctx := context.Background()
	p, err := CreateClient(ctx, opts)
	require.NoError(t, err)

	errs := make(chan error)
	go func() { errs <- p.Close() }()
	go func() { errs <- p.Close() }()

	assert.NoError(t, <-errs)
	var edbErr Error
	require.True(t, errors.As(<-errs, &edbErr), "wrong error: %v", err)
	assert.True(t, edbErr.Category(InterfaceError), "wrong error: %v", err)
}

func TestClientTx(t *testing.T) {
	ctx := context.Background()

	p, err := CreateClient(ctx, opts)
	require.NoError(t, err)
	defer p.Close() // nolint:errcheck

	var result int64
	err = p.Tx(ctx, func(ctx context.Context, tx *Tx) error {
		return tx.QuerySingle(ctx, "SELECT 33*21", &result)
	})

	require.NoError(t, err)
	require.Equal(t, int64(693), result, "Client.Tx() failed")
}

func TestQuerySingleMissingResult(t *testing.T) {
	ctx := context.Background()

	var result string
	err := client.QuerySingle(ctx, "SELECT <str>{}", &result)
	assert.EqualError(t, err, "edgedb.NoDataError: zero results")

	optionalResult := types.NewOptionalStr("this should be set to missing")
	err = client.QuerySingle(ctx, "SELECT <str>{}", &optionalResult)
	assert.NoError(t, err)
	assert.Equal(t, edgedbtypes.OptionalStr{}, optionalResult)

	var objectResult struct {
		Name string `edgedb:"name"`
	}
	err = client.QuerySingle(ctx,
		"SELECT sys::Database { name } FILTER .name = 'does not exist'",
		&objectResult,
	)
	assert.EqualError(t, err, "edgedb.NoDataError: zero results")

	var optionalObjectResult struct {
		edgedbtypes.Optional
		Name string `edgedb:"name"`
	}
	optionalObjectResult.SetMissing(false)
	err = client.QuerySingle(ctx,
		"SELECT sys::Database { name } FILTER .name = 'does not exist'",
		&optionalObjectResult,
	)
	assert.NoError(t, err)
	assert.Equal(t, "", optionalObjectResult.Name)
	assert.True(t, optionalObjectResult.Missing())
}

func TestQuerySingleJSONMissingResult(t *testing.T) {
	ctx := context.Background()

	var result []byte
	err := client.QuerySingleJSON(ctx, "SELECT <str>{}", &result)
	assert.EqualError(t, err, "edgedb.NoDataError: zero results")

	optionalResult := types.NewOptionalBytes(
		[]byte("this should be set to missing"),
	)
	err = client.QuerySingleJSON(ctx, "SELECT <str>{}", &optionalResult)
	assert.NoError(t, err)
	assert.Equal(t, edgedbtypes.OptionalBytes{}, optionalResult)

	var wrongType string
	err = client.QuerySingleJSON(ctx, "SELECT <str>{}", &wrongType)
	assert.EqualError(t, err, "edgedb.InterfaceError: "+
		"the \"out\" argument must be *[]byte or *OptionalBytes, got *string")
}

func TestSessionIdleTimeout(t *testing.T) {
	ctx := context.Background()
	p, err := CreateClient(ctx, opts)
	require.NoError(t, err)

	var result types.Duration
	err = p.QuerySingle(ctx,
		"SELECT assert_single(cfg::Config.session_idle_timeout)", &result)
	require.NoError(t, err)
	require.Equal(t, types.Duration(1_000_000), result)

	// The client keeps one connection in the pool.
	// Get a reference to that connection.
	con1, err := p.acquire(ctx)
	require.NoError(t, err)
	require.NotNil(t, con1)

	err = p.release(con1, nil)
	require.NoError(t, err)

	// After releasing we should get the same connection back again on acquire.
	con2, err := p.acquire(ctx)
	require.NoError(t, err)
	require.NotNil(t, con2)
	assert.Equal(t, con1, con2)

	err = p.release(con2, nil)
	require.NoError(t, err)

	// If the pooled connection is not used for longer than the
	// session_idle_timeout then the next acquired connection should be a new
	// connection.
	time.Sleep(1_200 * time.Millisecond)

	con3, err := p.acquire(ctx)
	require.NoError(t, err)
	require.NotNil(t, con3)
	assert.NotEqual(t, unsafe.Pointer(con1), unsafe.Pointer(con3))

	err = p.release(con3, nil)
	assert.NoError(t, err)
}

// Try to trigger race conditions
func TestConcurentClientUsage(t *testing.T) {
	ctx := context.Background()
	var done sync.WaitGroup

	for i := 0; i < 2; i++ {
		done.Add(1)
		go func() {
			var result int64
			for j := 0; j < 10; j++ {
				_ = client.QuerySingle(ctx, "SELECT 1", &result)
			}
			done.Done()
		}()
	}

	done.Wait()
}
