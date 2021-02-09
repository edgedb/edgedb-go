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
	"runtime"
	"sync"

	"github.com/edgedb/edgedb-go/internal/cache"
	"github.com/edgedb/edgedb-go/internal/cardinality"
	"github.com/edgedb/edgedb-go/internal/format"
	"github.com/edgedb/edgedb-go/internal/header"
)

var (
	defaultMinConns = 1
	defaultMaxConns = max(4, runtime.NumCPU())
)

func max(a, b int) int {
	if a > b {
		return a
	}

	return b
}

// Pool is a connection pool.
type Pool interface {
	Executor
	Trier

	// Acquire returns a connection from the pool
	// blocking until a connection is available.
	// Acquired connections must be released to the pool when no longer needed.
	Acquire(context.Context) (PoolConn, error)

	// Close closes all connections in the pool.
	// Calling close blocks until all acquired connections have been released,
	// and returns an error if called more than once.
	Close() error
}

type pool struct {
	isClosed bool
	mu       sync.RWMutex // locks isClosed

	// A buffered channel of connections ready for use.
	freeConns chan *reconnectingConn

	// A buffered channel of structs representing unconnected capacity.
	potentialConns chan struct{}

	maxConns int
	minConns int

	cfg *connConfig

	typeIDCache   *cache.Cache
	inCodecCache  *cache.Cache
	outCodecCache *cache.Cache
}

// Connect a pool of connections to a server.
func Connect(ctx context.Context, opts Options) (Pool, error) { // nolint:gocritic,lll
	return ConnectDSN(ctx, "", opts)
}

// ConnectDSN connects a pool to a server.
//
// dsn is either an instance name
// https://www.edgedb.com/docs/clients/00_python/instances/#edgedb-instances
// or it specifies a single string in the following format:
//
//     edgedb://user:password@host:port/database?option=value.
//
// The following options are recognized: host, port, user, database, password.
func ConnectDSN(ctx context.Context, dsn string, opts Options) (Pool, error) { // nolint:gocritic,lll
	minConns := defaultMinConns
	if opts.MinConns > 0 {
		minConns = int(opts.MinConns)
	}

	maxConns := defaultMaxConns
	if opts.MaxConns > 0 {
		maxConns = int(opts.MaxConns)
	}

	if maxConns < minConns {
		return nil, &configurationError{msg: fmt.Sprintf(
			"MaxConns (%v) may not be less than MinConns (%v)",
			maxConns, minConns,
		)}
	}

	cfg, err := parseConnectDSNAndArgs(dsn, &opts)
	if err != nil {
		return nil, err
	}

	p := &pool{
		maxConns: maxConns,
		minConns: minConns,
		cfg:      cfg,

		freeConns:      make(chan *reconnectingConn, minConns),
		potentialConns: make(chan struct{}, maxConns),

		typeIDCache:   cache.New(1_000),
		inCodecCache:  cache.New(1_000),
		outCodecCache: cache.New(1_000),
	}

	for i := 0; i < maxConns-minConns; i++ {
		p.potentialConns <- struct{}{}
	}

	wg := &sync.WaitGroup{}
	errs := make([]error, opts.MinConns)
	for i := 0; i < minConns; i++ {
		wg.Add(1)
		go func(i int, wg *sync.WaitGroup) {
			defer wg.Done()

			conn, err := p.newConn(ctx)
			if err == nil {
				p.freeConns <- conn
				return
			}
			errs[i] = err
			p.potentialConns <- struct{}{}
		}(i, wg)
	}

	wg.Wait()
	if err := wrapAll(errs...); err != nil {
		_ = p.Close()
		return nil, err
	}

	return p, nil
}

func (p *pool) newConn(ctx context.Context) (*reconnectingConn, error) {
	conn := &reconnectingConn{
		conn: &baseConn{
			cfg:           p.cfg,
			typeIDCache:   p.typeIDCache,
			inCodecCache:  p.inCodecCache,
			outCodecCache: p.outCodecCache,
		},
	}

	if err := conn.reconnect(ctx); err != nil {
		return nil, err
	}

	return conn, nil
}

func (p *pool) acquire(ctx context.Context) (*reconnectingConn, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.isClosed {
		return nil, &interfaceError{msg: "pool closed"}
	}

	// force do nothing if context is expired
	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("edgedb: %w", ctx.Err())
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
		return nil, fmt.Errorf("edgedb: %w", ctx.Err())
	}
}

func (p *pool) Acquire(ctx context.Context) (PoolConn, error) {
	conn, err := p.acquire(ctx)
	if err != nil {
		return nil, err
	}

	return &poolConn{pool: p, conn: conn}, nil
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

func (p *pool) release(conn *reconnectingConn, err error) error {
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

func (p *pool) Close() error {
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
func (p *pool) Execute(ctx context.Context, cmd string) error {
	conn, err := p.acquire(ctx)
	if err != nil {
		return err
	}

	hdrs := msgHeaders{header.AllowCapabilities: noTxCapabilities}
	q := sfQuery{cmd: cmd, headers: hdrs}
	err = conn.scriptFlow(ctx, q)
	return firstError(err, p.release(conn, err))
}

func (p *pool) Query(
	ctx context.Context,
	cmd string,
	out interface{},
	args ...interface{},
) error {
	conn, err := p.acquire(ctx)
	if err != nil {
		return err
	}

	hdrs := msgHeaders{header.AllowCapabilities: noTxCapabilities}
	q, err := newQuery(cmd, format.Binary, cardinality.Many, args, hdrs, out)
	if err != nil {
		return err
	}

	err = conn.granularFlow(ctx, q)
	return firstError(err, p.release(conn, err))
}

func (p *pool) QueryOne(
	ctx context.Context,
	cmd string,
	out interface{},
	args ...interface{},
) error {
	conn, err := p.acquire(ctx)
	if err != nil {
		return err
	}

	hdrs := msgHeaders{header.AllowCapabilities: noTxCapabilities}
	q, err := newQuery(cmd, format.Binary, cardinality.One, args, hdrs, out)
	if err != nil {
		return err
	}

	err = conn.granularFlow(ctx, q)
	return firstError(err, p.release(conn, err))
}

func (p *pool) QueryJSON(
	ctx context.Context,
	cmd string,
	out *[]byte,
	args ...interface{},
) error {
	conn, err := p.acquire(ctx)
	if err != nil {
		return err
	}

	hdrs := msgHeaders{header.AllowCapabilities: noTxCapabilities}
	q, err := newQuery(cmd, format.JSON, cardinality.Many, args, hdrs, out)
	if err != nil {
		return err
	}

	err = conn.granularFlow(ctx, q)
	return firstError(err, p.release(conn, err))
}

func (p *pool) QueryOneJSON(
	ctx context.Context,
	cmd string,
	out *[]byte,
	args ...interface{},
) error {
	conn, err := p.acquire(ctx)
	if err != nil {
		return err
	}

	hdrs := msgHeaders{header.AllowCapabilities: noTxCapabilities}
	q, err := newQuery(cmd, format.JSON, cardinality.One, args, hdrs, out)
	if err != nil {
		return err
	}

	err = conn.granularFlow(ctx, q)
	return firstError(err, p.release(conn, err))
}

func (p *pool) TryTx(ctx context.Context, action Action) error {
	conn, err := p.acquire(ctx)
	if err != nil {
		return err
	}

	return firstError(
		conn.TryTx(ctx, action),
		p.release(conn, err),
	)
}

func (p *pool) Retry(ctx context.Context, action Action) error {
	conn, err := p.acquire(ctx)
	if err != nil {
		return err
	}

	return firstError(
		conn.Retry(ctx, action),
		p.release(conn, err),
	)
}
