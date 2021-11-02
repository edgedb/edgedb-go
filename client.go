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
	defaultConcurrency = max(4, runtime.NumCPU())
)

func max(a, b int) int {
	if a > b {
		return a
	}

	return b
}

// Client is a connection pool and is safe for concurrent use.
type Client struct {
	isClosed      *bool
	isClosedMutex *sync.RWMutex // locks isClosed

	// A buffered channel of connections ready for use.
	freeConns chan transactableConn

	// A buffered channel of structs representing unconnected capacity.
	// This field remains nil until the first connection is acquired.
	potentialConns       chan struct{}
	potentialConnsMutext *sync.Mutex

	concurrency int

	txOpts    TxOptions
	retryOpts RetryOptions

	cfg *connConfig
	cacheCollection
}

// CreateClient returns a new client. The client connects lazily. Call
// Client.EnsureConnected() to force a connection.
func CreateClient(ctx context.Context, opts Options) (*Client, error) { // nolint:gocritic,lll
	return CreateClientDSN(ctx, "", opts)
}

// CreateClientDSN returns a new client. See also CreateClient.
//
// dsn is either an instance name
// https://www.edgedb.com/docs/clients/00_python/instances/#edgedb-instances
// or it specifies a single string in the following format:
//
//     edgedb://user:password@host:port/database?option=value.
//
// The following options are recognized: host, port, user, database, password.
func CreateClientDSN(ctx context.Context, dsn string, opts Options) (*Client, error) { // nolint:gocritic,lll
	cfg, err := parseConnectDSNAndArgs(dsn, &opts, newCfgPaths())
	if err != nil {
		return nil, err
	}

	False := false
	p := &Client{
		isClosed:             &False,
		isClosedMutex:        &sync.RWMutex{},
		cfg:                  cfg,
		txOpts:               NewTxOptions(),
		concurrency:          int(opts.Concurrency),
		freeConns:            make(chan transactableConn, 1),
		potentialConnsMutext: &sync.Mutex{},
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

	return p, nil
}

func (p *Client) newConn(ctx context.Context) (transactableConn, error) {
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

func (p *Client) acquire(ctx context.Context) (transactableConn, error) {
	p.isClosedMutex.RLock()
	defer p.isClosedMutex.RUnlock()

	if *p.isClosed {
		return transactableConn{}, &interfaceError{msg: "client closed"}
	}

	p.potentialConnsMutext.Lock()
	if p.potentialConns == nil {
		conn, err := p.newConn(ctx)
		if err != nil {
			p.potentialConnsMutext.Unlock()
			return transactableConn{}, err
		}

		if p.concurrency == 0 {
			// The user did not set Concurrency in provided Options.
			// See if the server sends a suggested max size.
			suggested, err := strconv.Atoi(
				string(conn.cfg.serverSettings["suggested_pool_concurrency"]))
			if err == nil {
				p.concurrency = suggested
			} else {
				p.concurrency = defaultConcurrency
			}
		}

		p.potentialConns = make(chan struct{}, p.concurrency)
		for i := 0; i < p.concurrency-1; i++ {
			p.potentialConns <- struct{}{}
		}

		p.potentialConnsMutext.Unlock()
		return conn, nil
	}
	p.potentialConnsMutext.Unlock()

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

func (p *Client) release(conn *transactableConn, err error) error {
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

// EnsureConnected forces the client to connect if it hasn't already.
func (p *Client) EnsureConnected(ctx context.Context) error {
	conn, err := p.acquire(ctx)
	if err != nil {
		return err
	}

	return p.release(&conn, nil)
}

// Close closes all connections in the pool.
// Calling close blocks until all acquired connections have been released,
// and returns an error if called more than once.
func (p *Client) Close() error {
	p.isClosedMutex.Lock()
	defer p.isClosedMutex.Unlock()

	if *p.isClosed {
		return &interfaceError{msg: "client closed"}
	}
	*p.isClosed = true

	p.potentialConnsMutext.Lock()
	if p.potentialConns == nil {
		// The client never made any connections.
		p.potentialConnsMutext.Unlock()
		return nil
	}
	p.potentialConnsMutext.Unlock()

	wg := sync.WaitGroup{}
	errs := make([]error, p.concurrency)
	for i := 0; i < p.concurrency; i++ {
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
func (p *Client) Execute(ctx context.Context, cmd string) error {
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
func (p *Client) Query(
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

// QuerySingle runs a singleton-returning query and returns its element.
// If the query executes successfully but doesn't return a result
// a NoDataError is returned.
func (p *Client) QuerySingle(
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
func (p *Client) QueryJSON(
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

// QuerySingleJSON runs a singleton-returning query.
// If the query executes successfully but doesn't have a result
// a NoDataError is returned.
func (p *Client) QuerySingleJSON(
	ctx context.Context,
	cmd string,
	out interface{},
	args ...interface{},
) error {
	conn, err := p.acquire(ctx)
	if err != nil {
		return err
	}

	err = runQuery(ctx, &conn, "QuerySingleJSON", cmd, out, args)
	return firstError(err, p.release(&conn, err))
}

// Tx runs an action in a transaction retrying failed actions
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
func (p *Client) Tx(ctx context.Context, action TxBlock) error {
	conn, err := p.acquire(ctx)
	if err != nil {
		return err
	}

	err = conn.Tx(ctx, action)
	return firstError(err, p.release(&conn, err))
}
