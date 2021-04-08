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
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConnectPool(t *testing.T) {
	ctx := context.Background()
	p, err := Connect(ctx, opts)
	require.Nil(t, err)

	var result string
	err = p.QueryOne(ctx, "SELECT 'hello';", &result)
	assert.Nil(t, err)
	assert.Equal(t, "hello", result)

	p2 := p.WithTxOptions(NewTxOptions())

	err = p.Close()
	assert.Nil(t, err)

	// Copied pools should be closed if a different copy is closed.
	err = p2.Close()
	assert.EqualError(t, err, "edgedb.InterfaceError: pool closed")
}

func TestPoolRejectsTransaction(t *testing.T) {
	ctx := context.Background()
	p, err := Connect(ctx, opts)
	require.Nil(t, err)
	defer p.Close() // nolint:errcheck

	expected := "edgedb.DisabledCapabilityError: " +
		"cannot execute transaction control commands"

	err = p.Execute(ctx, "START TRANSACTION")
	assert.EqualError(t, err, expected)

	var result []byte
	err = p.Query(ctx, "START TRANSACTION", &result)
	assert.EqualError(t, err, expected)

	err = p.QueryJSON(ctx, "START TRANSACTION", &result)
	assert.EqualError(t, err, expected)

	err = p.QueryOne(ctx, "START TRANSACTION", &result)
	assert.EqualError(t, err, expected)

	err = p.QueryOneJSON(ctx, "START TRANSACTION", &result)
	assert.EqualError(t, err, expected)
}

func TestConnectPoolZeroMinAndMaxConns(t *testing.T) {
	o := opts
	o.MinConns = 0
	o.MaxConns = 0

	ctx := context.Background()
	p, err := Connect(ctx, o)
	require.Nil(t, err)

	require.Equal(t, defaultMinConns, p.minConns)
	require.Equal(t, defaultMaxConns, p.maxConns)

	var result string
	err = p.QueryOne(ctx, "SELECT 'hello';", &result)
	assert.Nil(t, err)
	assert.Equal(t, "hello", result)

	err = p.Close()
	assert.Nil(t, err)
}

func TestClosePoolConcurently(t *testing.T) {
	ctx := context.Background()
	p, err := Connect(ctx, opts)
	require.Nil(t, err)

	errs := make(chan error)
	go func() { errs <- p.Close() }()
	go func() { errs <- p.Close() }()

	assert.Nil(t, <-errs)
	var edbErr Error
	require.True(t, errors.As(<-errs, &edbErr), "wrong error: %v", err)
	assert.True(t, edbErr.Category(InterfaceError), "wrong error: %v", err)
}

func TestConnectPoolMinConnLteMaxConn(t *testing.T) {
	ctx := context.Background()
	o := Options{MinConns: 5, MaxConns: 1}
	_, err := Connect(ctx, o)
	assert.EqualError(
		t,
		err,
		"edgedb.ConfigurationError: "+
			"MaxConns (1) may not be less than MinConns (5)",
	)

	var expected Error
	require.True(t, errors.As(err, &expected), "wrong error: %v", err)
	assert.True(
		t,
		expected.Category(ConfigurationError),
		"wrong error: %v",
		err,
	)
}

func mockPool(opts Options) *Pool { // nolint:gocritic
	False := false

	return &Pool{
		isClosed:       &False,
		mu:             &sync.RWMutex{},
		maxConns:       int(opts.MaxConns),
		minConns:       int(opts.MinConns),
		freeConns:      make(chan *reconnectingConn, opts.MinConns),
		potentialConns: make(chan struct{}, opts.MaxConns),
	}
}

func TestAcquireFromClosedPool(t *testing.T) {
	p := mockPool(Options{})
	err := p.Close()
	require.Nil(t, err)

	conn, err := p.Acquire(context.TODO())
	var edbErr Error
	require.True(t, errors.As(err, &edbErr), "wrong error: %v", err)
	assert.True(t, edbErr.Category(InterfaceError), "wrong error: %v", err)
	assert.Nil(t, conn)
}

func TestAcquireFreeConnFromPool(t *testing.T) {
	p := mockPool(Options{MinConns: 1})
	conn := &reconnectingConn{}
	p.freeConns <- conn

	pConn, err := p.Acquire(context.Background())
	assert.Nil(t, err)
	assert.Equal(t, conn, pConn.conn)
}

func BenchmarkPoolAcquireRelease(b *testing.B) {
	p := mockPool(Options{MaxConns: 2, MinConns: 2})

	for i := 0; i < p.maxConns; i++ {
		p.freeConns <- &reconnectingConn{}
	}

	var conn *reconnectingConn
	ctx := context.Background()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		conn, _ = p.acquire(ctx)
		_ = p.release(conn, nil)
	}
}

func TestAcquirePotentialConnFromPool(t *testing.T) {
	p, err := Connect(context.Background(), opts)
	require.Nil(t, err)

	// free connection
	a, err := p.Acquire(context.Background())
	require.Nil(t, err)
	require.NotNil(t, a)

	// potential connection
	b, err := p.Acquire(context.Background())
	require.Nil(t, err)
	require.NotNil(t, b)

	require.Nil(t, b.Release())
	require.Nil(t, a.Release())
	require.Nil(t, p.Close())
}

func TestPoolAcquireExpiredContext(t *testing.T) {
	p := mockPool(Options{MaxConns: 1, MinConns: 1})
	p.freeConns <- &reconnectingConn{}
	p.potentialConns <- struct{}{}

	ctx, cancel := context.WithDeadline(context.Background(), time.Now())
	cancel()

	conn, err := p.Acquire(ctx)
	assert.True(t, errors.Is(err, context.DeadlineExceeded))
	assert.Nil(t, conn)
}

func TestPoolAcquireThenContextExpires(t *testing.T) {
	p := mockPool(Options{})

	deadline := time.Now().Add(10 * time.Millisecond)
	ctx, cancel := context.WithDeadline(context.Background(), deadline)
	conn, err := p.Acquire(ctx)
	assert.True(t, errors.Is(err, context.DeadlineExceeded))
	assert.Nil(t, conn)
	cancel()
}

func TestClosePool(t *testing.T) {
	p := mockPool(Options{})

	err := p.Close()
	assert.Nil(t, err)

	err = p.Close()
	var edbErr Error
	require.True(t, errors.As(err, &edbErr), "wrong error: %v", err)
	assert.True(t, edbErr.Category(InterfaceError), "wrong error: %v", err)
}

func TestPoolRetryingTx(t *testing.T) {
	ctx := context.Background()

	p, err := Connect(ctx, opts)
	require.Nil(t, err, "unexpected error: %v", err)
	defer p.Close() // nolint:errcheck

	var result int64
	err = p.RetryingTx(ctx, func(ctx context.Context, tx *Tx) error {
		return tx.QueryOne(ctx, "SELECT 33*21", &result)
	})

	require.Nil(t, err, "unexpected error: %v", err)
	require.Equal(t, int64(693), result, "Pool.RetryingTx() failed")
}
