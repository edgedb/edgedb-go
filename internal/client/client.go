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
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/geldata/gel-go/internal/cache"
	types "github.com/geldata/gel-go/internal/geltypes"
)

const defaultIdleConnectionTimeout = 30 * time.Second

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
	freeConns chan func() *transactableConn

	// A buffered channel of structs representing unconnected capacity.
	// This field remains nil until the first connection is acquired.
	potentialConns       chan struct{}
	potentialConnsMutext *sync.Mutex

	concurrency int

	txOpts    TxOptions
	retryOpts RetryOptions

	cfg *connConfig
	cacheCollection
	state map[string]interface{}

	warningHandler WarningHandler
}

// CreateClient returns a new client. The client connects lazily. Call
// Client.EnsureConnected() to force a connection.
func CreateClient(ctx context.Context, opts Options) (*Client, error) { // nolint:gocritic,lll
	return CreateClientDSN(ctx, "", opts)
}

// CreateClientDSN returns a new client. See also CreateClient.
//
// dsn is either an instance name
// https://www.edgedb.com/docs/clients/connection
// or it specifies a single string in the following format:
//
//	gel://user:password@host:port/database?option=value.
//
// The following options are recognized: host, port, user, database, password.
func CreateClientDSN(_ context.Context, dsn string, opts Options) (*Client, error) { // nolint:gocritic,lll
	cfg, err := parseConnectDSNAndArgs(dsn, &opts, newCfgPaths())
	if err != nil {
		return nil, err
	}

	warningHandler := LogWarnings
	if opts.WarningHandler != nil {
		warningHandler = opts.WarningHandler
	}

	False := false
	p := &Client{
		isClosed:             &False,
		isClosedMutex:        &sync.RWMutex{},
		cfg:                  cfg,
		txOpts:               NewTxOptions(),
		concurrency:          int(opts.Concurrency),
		freeConns:            make(chan func() *transactableConn, 1),
		potentialConnsMutext: &sync.Mutex{},
		retryOpts:            NewRetryOptions(),
		cacheCollection: cacheCollection{
			serverSettings:    cfg.serverSettings,
			typeIDCache:       cache.New(1_000),
			inCodecCache:      cache.New(1_000),
			outCodecCache:     cache.New(1_000),
			capabilitiesCache: cache.New(1_000),
		},
		state:          make(map[string]interface{}),
		warningHandler: warningHandler,
	}

	return p, nil
}

func (p *Client) newConn(ctx context.Context) (*transactableConn, error) {
	conn := transactableConn{
		txOpts:    p.txOpts,
		retryOpts: p.retryOpts,
		reconnectingConn: &reconnectingConn{
			cfg:             p.cfg,
			cacheCollection: p.cacheCollection,
		},
	}

	if err := conn.reconnect(ctx, false); err != nil {
		return nil, err
	}

	return &conn, nil
}

func (p *Client) acquire(ctx context.Context) (*transactableConn, error) {
	p.isClosedMutex.RLock()
	defer p.isClosedMutex.RUnlock()

	if *p.isClosed {
		return nil, &interfaceError{msg: "client closed"}
	}

	p.potentialConnsMutext.Lock()
	if p.potentialConns == nil {
		conn, err := p.newConn(ctx)
		if err != nil {
			p.potentialConnsMutext.Unlock()
			return nil, err
		}

		if p.concurrency == 0 {
			// The user did not set Concurrency in provided Options.
			// See if the server sends a suggested max size.
			suggested, ok := conn.cfg.serverSettings.
				GetOk("suggested_pool_concurrency")
			if ok {
				p.concurrency = suggested.(int)
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
		return nil, fmt.Errorf("gel: %w", ctx.Err())
	default:
	}

	// force using an existing connection over connecting a new socket.
	select {
	case acquireIfNotTimedout := <-p.freeConns:
		conn := acquireIfNotTimedout()
		if conn != nil {
			return conn, nil
		}
	default:
	}

	for {
		select {
		case acquireIfNotTimedout := <-p.freeConns:
			conn := acquireIfNotTimedout()
			if conn != nil {
				return conn, nil
			}
			continue
		case <-p.potentialConns:
			conn, err := p.newConn(ctx)
			if err != nil {
				p.potentialConns <- struct{}{}
				return nil, err
			}
			return conn, nil
		case <-ctx.Done():
			return nil, fmt.Errorf("gel: %w", ctx.Err())
		}
	}
}

type systemConfig struct {
	ID                 types.OptionalUUID     `gel:"id"`
	SessionIdleTimeout types.OptionalDuration `gel:"session_idle_timeout"`
}

func (p *Client) release(conn *transactableConn, err error) error {
	if isClientConnectionError(err) {
		p.potentialConns <- struct{}{}
		return conn.Close()
	}

	timeout := defaultIdleConnectionTimeout
	if t, ok := conn.conn.systemConfig.SessionIdleTimeout.Get(); ok {
		timeout = time.Duration(1_000 * t)
	}

	// 0 or less disables the idle timeout
	if timeout <= 0 {
		select {
		case p.freeConns <- func() *transactableConn { return conn }:
			return nil
		default:
			// we have MinConns idle so no need to keep this connection.
			p.potentialConns <- struct{}{}
			return conn.Close()
		}
	}

	cancel := make(chan struct{}, 1)
	connChan := make(chan *transactableConn, 1)

	acquireIfNotTimedout := func() *transactableConn {
		cancel <- struct{}{}
		return <-connChan
	}

	select {
	case p.freeConns <- acquireIfNotTimedout:
		go func() {
			select {
			case <-cancel:
				connChan <- conn
			case <-time.After(timeout):
				connChan <- nil
				p.potentialConns <- struct{}{}
				if e := conn.Close(); e != nil {
					log.Println("error while closing idle connection:", e)
				}
			}
		}()
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

	return p.release(conn, nil)
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
		case acquireIfNotTimedout := <-p.freeConns:
			wg.Add(1)
			go func(i int) {
				conn := acquireIfNotTimedout()
				if conn != nil {
					errs[i] = conn.Close()
				}
				wg.Done()
			}(i)
		case <-p.potentialConns:
		}
	}

	wg.Wait()
	return wrapAll(errs...)
}

// Execute an EdgeQL command (or commands).
func (p *Client) Execute(
	ctx context.Context,
	cmd string,
	args ...interface{},
) error {
	conn, err := p.acquire(ctx)
	if err != nil {
		return err
	}

	q, err := newQuery(
		"Execute",
		cmd,
		args,
		conn.capabilities1pX(),
		copyState(p.state),
		nil,
		true,
		p.warningHandler,
	)
	if err != nil {
		return err
	}

	err = conn.scriptFlow(ctx, q)
	return firstError(err, p.release(conn, err))
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

	err = runQuery(
		ctx, conn, "Query", cmd, out, args, p.state, p.warningHandler)
	return firstError(err, p.release(conn, err))
}

// QuerySingle runs a singleton-returning query and returns its element.
// If the query executes successfully but doesn't return a result
// a NoDataError is returned. If the out argument is an optional type the out
// argument will be set to missing instead of returning a NoDataError.
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

	err = runQuery(
		ctx,
		conn,
		"QuerySingle",
		cmd,
		out,
		args,
		p.state,
		p.warningHandler,
	)
	return firstError(err, p.release(conn, err))
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

	err = runQuery(
		ctx,
		conn,
		"QueryJSON",
		cmd,
		out,
		args,
		p.state,
		p.warningHandler,
	)
	return firstError(err, p.release(conn, err))
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

	err = runQuery(
		ctx,
		conn,
		"QuerySingleJSON",
		cmd,
		out,
		args,
		p.state,
		p.warningHandler,
	)
	return firstError(err, p.release(conn, err))
}

// QuerySQL runs a SQL query and returns the results.
func (p *Client) QuerySQL(
	ctx context.Context,
	cmd string,
	out interface{},
	args ...interface{},
) error {
	conn, err := p.acquire(ctx)
	if err != nil {
		return err
	}

	err = runQuery(
		ctx, conn, "QuerySQL", cmd, out, args, p.state, p.warningHandler)
	return firstError(err, p.release(conn, err))
}

// ExecuteSQL executes a SQL command (or commands).
func (p *Client) ExecuteSQL(
	ctx context.Context,
	cmd string,
	args ...interface{},
) error {
	conn, err := p.acquire(ctx)
	if err != nil {
		return err
	}

	q, err := newQuery(
		"ExecuteSQL",
		cmd,
		args,
		conn.capabilities1pX(),
		copyState(p.state),
		nil,
		true,
		p.warningHandler,
	)
	if err != nil {
		return err
	}

	err = conn.scriptFlow(ctx, q)
	return firstError(err, p.release(conn, err))
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

	err = conn.tx(ctx, action, p.state, p.warningHandler)
	return firstError(err, p.release(conn, err))
}
