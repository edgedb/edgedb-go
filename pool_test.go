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
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConnectPool(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	o := opts
	o.MinConns = 1
	o.MaxConns = 2
	pool, err := Connect(ctx, o)
	require.Nil(t, err)

	var result string
	err = pool.QueryOne(ctx, "SELECT 'hello';", &result)
	assert.Nil(t, err)
	assert.Equal(t, "hello", result)

	err = pool.Close()
	assert.Nil(t, err)
}

func TestClosePoolConcurently(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	o := opts
	o.MinConns = 1
	o.MaxConns = 2
	pool, err := Connect(ctx, o)
	require.Nil(t, err)

	errs := make(chan error)
	go func() { errs <- pool.Close() }()
	go func() { errs <- pool.Close() }()

	assert.Nil(t, <-errs)
	assert.Equal(t, ErrPoolClosed, <-errs)
}

func TestConnectPoolMinConnGteZero(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	o := Options{MinConns: 0, MaxConns: 10}
	_, err := Connect(ctx, o)
	assert.EqualError(
		t,
		err,
		"edgedb: MinConns may not be less than 1, got: 0",
	)

	var expected *ConfigurationError
	assert.True(t, errors.As(err, &expected))
}

func TestConnectPoolMinConnLteMaxConn(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	o := Options{MinConns: 5, MaxConns: 1}
	_, err := Connect(ctx, o)
	assert.EqualError(
		t,
		err,
		"edgedb: MaxConns (1) may not be less than MinConns (5)",
	)

	var expected *ConfigurationError
	assert.True(t, errors.As(err, &expected))
}

func TestAcquireFromClosedPool(t *testing.T) {
	pool := &Pool{
		isClosed:       true,
		freeConns:      make(chan *baseConn),
		potentialConns: make(chan struct{}),
	}

	conn, err := pool.Acquire(context.TODO())
	require.Equal(t, err, ErrPoolClosed)
	assert.Nil(t, conn)
}

func TestAcquireFreeConnFromPool(t *testing.T) {
	conn := &baseConn{}
	pool := &Pool{freeConns: make(chan *baseConn, 1)}
	pool.freeConns <- conn

	result, err := pool.Acquire(context.TODO())
	assert.Nil(t, err)
	assert.Equal(t, conn, result.baseConn)
}

func BenchmarkPoolAcquireRelease(b *testing.B) {
	pool := &Pool{
		maxConns:       2,
		minConns:       2,
		freeConns:      make(chan *baseConn, 2),
		potentialConns: make(chan struct{}, 2),
	}

	for i := 0; i < pool.maxConns; i++ {
		pool.freeConns <- &baseConn{}
	}

	var conn *baseConn
	ctx := context.TODO()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		conn, _ = pool.acquire(ctx)
		_ = pool.release(conn, nil)
	}
}

func TestAcquirePotentialConnFromPool(t *testing.T) {
	o := opts
	o.MaxConns = 2
	o.MinConns = 1
	pool, err := Connect(context.TODO(), o)
	require.Nil(t, err)
	defer func() {
		assert.Nil(t, pool.Close())
	}()

	// free connection
	a, err := pool.Acquire(context.TODO())
	require.Nil(t, err)
	require.NotNil(t, a)
	defer func() { assert.Nil(t, a.Release()) }()

	// potential connection
	b, err := pool.Acquire(context.TODO())
	require.Nil(t, err)
	require.NotNil(t, b)
	defer func() { assert.Nil(t, b.Release()) }()
}

func TestPoolAcquireExpiredContext(t *testing.T) {
	pool := &Pool{
		freeConns:      make(chan *baseConn, 1),
		potentialConns: make(chan struct{}, 1),
	}
	pool.freeConns <- &baseConn{}
	pool.potentialConns <- struct{}{}

	ctx, cancel := context.WithDeadline(context.Background(), time.Now())
	cancel()

	conn, err := pool.Acquire(ctx)
	assert.True(t, errors.Is(err, context.DeadlineExceeded))
	assert.Nil(t, conn)
}

func TestPoolAcquireThenContextExpires(t *testing.T) {
	pool := &Pool{}

	deadline := time.Now().Add(10 * time.Millisecond)
	ctx, cancel := context.WithDeadline(context.Background(), deadline)
	conn, err := pool.Acquire(ctx)
	assert.True(t, errors.Is(err, context.DeadlineExceeded))
	assert.Nil(t, conn)
	cancel()
}

func TestClosePool(t *testing.T) {
	pool := &Pool{
		maxConns:       0,
		minConns:       0,
		freeConns:      make(chan *baseConn),
		potentialConns: make(chan struct{}),
	}

	err := pool.Close()
	assert.Nil(t, err)

	err = pool.Close()
	assert.Equal(t, err, ErrPoolClosed)
}
