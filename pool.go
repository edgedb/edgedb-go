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
	"runtime"
	"strconv"
	"sync"

	"github.com/edgedb/edgedb-go/internal/cache"
)

var (
	defaultMaxConns = max(4, runtime.NumCPU())
)

func max(a, b int) int {
	if a > b {
		return a
	}

	return b
}

// Pool is a connection pool and is safe for concurrent use.
type Pool struct {
	isClosed *bool
	mu       *sync.RWMutex // locks isClosed

	// A buffered channel of connections ready for use.
	freeConns chan transactableConn

	// A buffered channel of structs representing unconnected capacity.
	potentialConns chan struct{}

	maxConns int

	txOpts    TxOptions
	retryOpts RetryOptions

	cfg *connConfig
	cacheCollection
}

// Connect a pool of connections to a server.
func Connect(ctx context.Context, opts Options) (*Pool, error) { // nolint:gocritic,lll
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
func ConnectDSN(ctx context.Context, dsn string, opts Options) (*Pool, error) { // nolint:gocritic,lll
	cfg, err := parseConnectDSNAndArgs(dsn, &opts)
	if err != nil {
		return nil, err
	}

	False := false
	p := &Pool{
		isClosed:  &False,
		mu:        &sync.RWMutex{},
		cfg:       cfg,
		txOpts:    NewTxOptions(),
		freeConns: make(chan transactableConn, 1),
		retryOpts: RetryOptions{
			txConflict: RetryRule{attempts: 3, backoff: defaultBackoff},
			network:    RetryRule{attempts: 3, backoff: defaultBackoff},
		},
		cacheCollection: cacheCollection{
			serverSettings:    cfg.serverSettings,
			typeIDCache:       cache.New(1_000),
			inCodecCache:      cache.New(1_000),
			outCodecCache:     cache.New(1_000),
			capabilitiesCache: cache.New(1_000),
		},
	}

	conn, err := p.newConn(ctx)
	if err != nil {
		return nil, err
	}

	suggested, err := strconv.Atoi(
		conn.cfg.serverSettings["suggested_pool_concurrency"])
	switch {
	case opts.MaxConns == 0 && err == nil:
		p.maxConns = suggested
	case opts.MaxConns == 0:
		p.maxConns = defaultMaxConns
	default:
		p.maxConns = int(opts.MaxConns)
	}

	p.freeConns <- conn
	p.potentialConns = make(chan struct{}, p.maxConns)
	for i := 0; i < p.maxConns-1; i++ {
		p.potentialConns <- struct{}{}
	}

	return p, nil
}

func (p *Pool) newConn(ctx context.Context) (transactableConn, error) {
	conn := transactableConn{
		txOpts:    p.txOpts,
		retryOpts: p.retryOpts,
		reconnectingConn: &reconnectingConn{
			cfg:             p.cfg,
			cacheCollection: p.cacheCollection,
		},
	}

	if err := conn.reconnect(ctx, false); err != nil {
		return transactableConn{}, err
	}

	return conn, nil
}

func (p *Pool) acquire(ctx context.Context) (transactableConn, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if *p.isClosed {
		return transactableConn{}, &interfaceError{msg: "pool closed"}
	}

	// force do nothing if context is expired
	select {
	case <-ctx.Done():
		return transactableConn{}, fmt.Errorf("edgedb: %w", ctx.Err())
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
			return transactableConn{}, err
		}
		return conn, nil
	case <-ctx.Done():
		return transactableConn{}, fmt.Errorf("edgedb: %w", ctx.Err())
	}
}

// Acquire returns a connection from the pool
// blocking until a connection is available.
// Acquired connections must be released to the pool when no longer needed.
//
// Deprecated: use the query methods on Pool instead
func (p *Pool) Acquire(ctx context.Context) (*PoolConn, error) {
	conn, err := p.acquire(ctx)
	if err != nil {
		return nil, err
	}

	False := false
	return &PoolConn{
		transactableConn: conn,
		pool:             p,
		isClosed:         &False,
	}, nil
}

func (p *Pool) release(conn *transactableConn, err error) error {
	if isClientConnectionError(err) {
		p.potentialConns <- struct{}{}
		return conn.Close()
	}

	select {
	case p.freeConns <- *conn:
	default:
		// we have MinConns idle so no need to keep this connection.
		p.potentialConns <- struct{}{}
		return conn.Close()
	}

	return nil
}

// Close closes all connections in the pool.
// Calling close blocks until all acquired connections have been released,
// and returns an error if called more than once.
func (p *Pool) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if *p.isClosed {
		return &interfaceError{msg: "pool closed"}
	}
	*p.isClosed = true

	wg := sync.WaitGroup{}
	errs := make([]error, p.maxConns)
	for i := 0; i < p.maxConns; i++ {
		select {
		case conn := <-p.freeConns:
			wg.Add(1)
			go func(i int) {
				errs[i] = conn.Close()
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

	q := sfQuery{
		cmd:     cmd,
		headers: conn.headers(),
	}

	err = conn.scriptFlow(ctx, q)
	return firstError(err, p.release(&conn, err))
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

	err = runQuery(ctx, &conn, "Query", cmd, out, args)
	return firstError(err, p.release(&conn, err))
}

// QueryOne runs a singleton-returning query and returns its element.
// If the query executes successfully but doesn't return a result
// a NoDataError is returned.
//
// Deprecated: use QuerySingle()
func (p *Pool) QueryOne(
	ctx context.Context,
	cmd string,
	out interface{},
	args ...interface{},
) error {
	return p.QuerySingle(ctx, cmd, out, args...)
}

// QuerySingle runs a singleton-returning query and returns its element.
// If the query executes successfully but doesn't return a result
// a NoDataError is returned.
func (p *Pool) QuerySingle(
	ctx context.Context,
	cmd string,
	out interface{},
	args ...interface{},
) error {
	conn, err := p.acquire(ctx)
	if err != nil {
		return err
	}

	err = runQuery(ctx, &conn, "QuerySingle", cmd, out, args)
	return firstError(err, p.release(&conn, err))
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

	err = runQuery(ctx, &conn, "QueryJSON", cmd, out, args)
	return firstError(err, p.release(&conn, err))
}

// QueryOneJSON runs a singleton-returning query.
// If the query executes successfully but doesn't have a result
// a NoDataError is returned.
//
// Deprecated: use QuerySingleJSON()
func (p *Pool) QueryOneJSON(
	ctx context.Context,
	cmd string,
	out *[]byte,
	args ...interface{},
) error {
	return p.QuerySingleJSON(ctx, cmd, out, args...)
}

// QuerySingleJSON runs a singleton-returning query.
// If the query executes successfully but doesn't have a result
// a NoDataError is returned.
func (p *Pool) QuerySingleJSON(
	ctx context.Context,
	cmd string,
	out *[]byte,
	args ...interface{},
) error {
	conn, err := p.acquire(ctx)
	if err != nil {
		return err
	}

	err = runQuery(ctx, &conn, "QuerySingleJSON", cmd, out, args)
	return firstError(err, p.release(&conn, err))
}

// RawTx runs an action in a transaction.
// If the action returns an error the transaction is rolled back,
// otherwise it is committed.
func (p *Pool) RawTx(ctx context.Context, action TxBlock) error {
	conn, err := p.acquire(ctx)
	if err != nil {
		return err
	}

	err = conn.RawTx(ctx, action)
	return firstError(err, p.release(&conn, err))
}

// RetryingTx does the same as RawTx but retries failed actions
// if they might succeed on a subsequent attempt.
//
// Retries are governed by retry rules.
// The default rule can be set with WithRetryRule().
// For more fine grained control a retry rule can be set
// for each defined RetryCondition using WithRetryCondition().
// When a transaction fails but is retryable
// the rule for the failure condition is used to determine if the transaction
// should be tried again based on RetryRule.Attempts and the amount of time
// to wait before retrying is determined by RetryRule.Backoff.
// If either field is unset (see RetryRule) then the default rule is used.
// If the object's default is unset the fall back is 3 attempts
// and exponential backoff.
func (p *Pool) RetryingTx(ctx context.Context, action TxBlock) error {
	conn, err := p.acquire(ctx)
	if err != nil {
		return err
	}

	err = conn.RetryingTx(ctx, action)
	return firstError(err, p.release(&conn, err))
}
