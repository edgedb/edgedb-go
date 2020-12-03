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
	"os"
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
	assert.Equal(t, ErrorPoolClosed, <-errs)
}

func TestConnectPoolMinConnGteZero(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	o := Options{MinConns: 0, MaxConns: 10}
	_, err := Connect(ctx, o)
	assert.EqualError(t, err, "MinConns may not be less than 1, got: 0")
	assert.True(t, errors.Is(err, ErrorConfiguration))
}

func TestConnectPoolMinConnLteMaxConn(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	o := Options{MinConns: 5, MaxConns: 1}
	_, err := Connect(ctx, o)
	assert.EqualError(t, err, "MaxConns may not be less than MinConns")
	assert.True(t, errors.Is(err, ErrorConfiguration))
}

func TestAcquireFromClosedPool(t *testing.T) {
	pool := &Pool{
		isClosed:       true,
		freeConns:      make(chan *baseConn),
		potentialConns: make(chan struct{}),
	}

	conn, err := pool.Acquire(context.TODO())
	require.Equal(t, err, ErrorPoolClosed)
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
	opts := &Options{MinConns: 2, MaxConns: 2}
	pool := &Pool{
		opts:           opts,
		freeConns:      make(chan *baseConn, opts.MaxConns),
		potentialConns: make(chan struct{}, opts.MaxConns),
	}

	for i := 0; i < opts.MaxConns; i++ {
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
	pool := &Pool{
		potentialConns: make(chan struct{}, 1),
		opts:           &opts,
	}
	pool.potentialConns <- struct{}{}

	deadline := time.Now().Add(10 * time.Millisecond)
	ctx, cancel := context.WithDeadline(context.Background(), deadline)
	conn, err := pool.Acquire(ctx)
	assert.True(t, errors.Is(err, os.ErrDeadlineExceeded))
	assert.Nil(t, conn)
	cancel()
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
	assert.Equal(t, err, ErrorContextExpired)
	assert.Nil(t, conn)
}

func TestPoolAcquireThenContextExpires(t *testing.T) {
	pool := &Pool{}

	deadline := time.Now().Add(10 * time.Millisecond)
	ctx, cancel := context.WithDeadline(context.Background(), deadline)
	conn, err := pool.Acquire(ctx)
	assert.Equal(t, err, ErrorContextExpired)
	assert.Nil(t, conn)
	cancel()
}

func TestClosePool(t *testing.T) {
	pool := &Pool{
		freeConns:      make(chan *baseConn),
		potentialConns: make(chan struct{}),
		opts:           &Options{MaxConns: 0, MinConns: 0},
	}
	err := pool.Close()
	assert.Nil(t, err)

	err = pool.Close()
	assert.Equal(t, err, ErrorPoolClosed)
}
