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
	"fmt"
	"net"
	"sync"

	"github.com/edgedb/edgedb-go/internal/cache"
)

// Pool is a pool of connections.
type Pool struct {
	isClosed bool
	mu       sync.RWMutex // locks isClosed

	// A buffered channel of connections ready for use.
	freeConns chan *baseConn

	// A buffered channel of structs representing unconnected capacity.
	potentialConns chan struct{}

	maxConns int
	minConns int

	cfg *connConfig

	typeIDCache   *cache.Cache
	inCodecCache  *cache.Cache
	outCodecCache *cache.Cache
}

// todo check connect tests in other clients

// todo add connectDSN funcs

// Connect a pool of connections to a server.
func Connect(ctx context.Context, opts Options) (*Pool, error) { // nolint:gocritic,lll
	return ConnectDSN(ctx, "", opts)
}

// ConnectDSN a pool of connections to a server.
func ConnectDSN(ctx context.Context, dsn string, opts Options) (*Pool, error) { // nolint:gocritic,lll
	if opts.MinConns < 1 {
		return nil, &configurationError{msg: fmt.Sprintf(
			"MinConns may not be less than 1, got: %v",
			opts.MinConns,
		)}
	}

	if opts.MaxConns < opts.MinConns {
		return nil, &configurationError{msg: fmt.Sprintf(
			"MaxConns (%v) may not be less than MinConns (%v)",
			opts.MaxConns, opts.MinConns,
		)}
	}

	cfg, err := parseConnectDSNAndArgs(dsn, &opts)
	if err != nil {
		return nil, err
	}

	pool := &Pool{
		maxConns: opts.MaxConns,
		minConns: opts.MinConns,
		cfg:      cfg,

		freeConns:      make(chan *baseConn, opts.MinConns),
		potentialConns: make(chan struct{}, opts.MaxConns),

		typeIDCache:   cache.New(1_000),
		inCodecCache:  cache.New(1_000),
		outCodecCache: cache.New(1_000),
	}

	for i := 0; i < opts.MaxConns-opts.MinConns; i++ {
		pool.potentialConns <- struct{}{}
	}

	wg := sync.WaitGroup{}
	errs := make([]error, opts.MinConns)
	for i := 0; i < opts.MinConns; i++ {
		wg.Add(1)
		go func(i int) {
			conn, err := pool.newConn(ctx)
			if err == nil {
				pool.freeConns <- conn
				return
			}
			errs[i] = err
			pool.potentialConns <- struct{}{}
		}(i)
	}

	wg.Done()
	if err := wrapAll(errs...); err != nil {
		_ = pool.Close()
		return nil, err
	}

	return pool, nil
}

func (p *Pool) newConn(ctx context.Context) (*baseConn, error) {
	conn := &baseConn{
		typeIDCache:   p.typeIDCache,
		inCodecCache:  p.inCodecCache,
		outCodecCache: p.outCodecCache,
	}

	if err := connectOne(ctx, p.cfg, conn); err != nil {
		return nil, err
	}

	return conn, nil
}

func (p *Pool) acquire(ctx context.Context) (*baseConn, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.isClosed {
		return nil, &interfaceError{msg: "pool closed"}
	}

	// force do nothing if context is expired
	select {
	case <-ctx.Done():
		err := ctx.Err()
		return nil, &baseError{msg: "edgedb: " + err.Error(), err: err}
	default:
	}

	// force using an existing connection over connecting a new socket.
	select {
	case conn := <-p.freeConns:
		return conn, nil
	default:
	}

	select {
	case conn := <-p.freeConns:
		return conn, nil
	case <-p.potentialConns:
		conn, err := p.newConn(ctx)
		if err != nil {
			p.potentialConns <- struct{}{}
			return nil, err
		}
		return conn, nil
	case <-ctx.Done():
		err := ctx.Err()
		return nil, &baseError{msg: "edgedb: " + err.Error(), err: err}
	}
}

// Acquire gets a connection out of the pool
// blocking until a connection is available.
// Acquired connections must be released to the pool when no longer needed.
func (p *Pool) Acquire(ctx context.Context) (*PoolConn, error) {
	conn, err := p.acquire(ctx)
	if err != nil {
		return nil, err
	}

	return &PoolConn{pool: p, baseConn: conn}, nil
}

func unrecoverable(err error) bool {
	if err == nil {
		return false
	}

	e, ok := err.(*net.OpError)
	if ok && e.Temporary() {
		return false
	}

	return true
}

func (p *Pool) release(conn *baseConn, err error) error {
	if unrecoverable(err) {
		p.potentialConns <- struct{}{}
		return conn.close()
	}

	select {
	case p.freeConns <- conn:
	default:
		// we have MinConns idle so no need to keep this connection.
		p.potentialConns <- struct{}{}
		return conn.close()
	}

	return nil
}

// Close closes all connections in the pool.
// Calling close blocks until all acquired connections have been released,
// and returns an error if called more than once.
func (p *Pool) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.isClosed {
		return &interfaceError{msg: "pool closed"}
	}
	p.isClosed = true

	wg := sync.WaitGroup{}
	errs := make([]error, p.maxConns)
	for i := 0; i < p.maxConns; i++ {
		select {
		case conn := <-p.freeConns:
			wg.Add(1)
			go func(i int) {
				errs[i] = conn.close()
				wg.Done()
			}(i)
		case <-p.potentialConns:
		}
	}

	wg.Wait()
	return wrapAll(errs...)
}

// Execute an EdgeQL command (or commands).
func (p *Pool) Execute(ctx context.Context, cmd string) error {
	conn, err := p.acquire(ctx)
	if err != nil {
		return err
	}

	err = conn.Execute(ctx, cmd)
	return firstError(err, p.release(conn, err))
}

// Query runs a query and returns the results.
func (p *Pool) Query(
	ctx context.Context,
	cmd string,
	out interface{},
	args ...interface{},
) error {
	conn, err := p.acquire(ctx)
	if err != nil {
		return err
	}

	err = conn.Query(ctx, cmd, out, args...)
	return firstError(err, p.release(conn, err))
}

// QueryOne runs a singleton-returning query and returns its element.
// If the query executes successfully but doesn't return a result
// ErrorZeroResults is returned.
func (p *Pool) QueryOne(
	ctx context.Context,
	cmd string,
	out interface{},
	args ...interface{},
) error {
	conn, err := p.acquire(ctx)
	if err != nil {
		return err
	}

	err = conn.QueryOne(ctx, cmd, out, args...)
	return firstError(err, p.release(conn, err))
}

// QueryJSON runs a query and return the results as JSON.
func (p *Pool) QueryJSON(
	ctx context.Context,
	cmd string,
	out *[]byte,
	args ...interface{},
) error {
	conn, err := p.acquire(ctx)
	if err != nil {
		return err
	}

	err = conn.QueryJSON(ctx, cmd, out, args...)
	return firstError(err, p.release(conn, err))
}

// QueryOneJSON runs a singleton-returning query
// and return its element in JSON.
// If the query executes successfully but doesn't return a result
// []byte{}, ErrorZeroResults is returned.
func (p *Pool) QueryOneJSON(
	ctx context.Context,
	cmd string,
	out *[]byte,
	args ...interface{},
) error {
	conn, err := p.acquire(ctx)
	if err != nil {
		return err
	}

	err = conn.QueryOneJSON(ctx, cmd, out, args...)
	return firstError(err, p.release(conn, err))
}
