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
	"fmt"
	"math"
	"math/rand"
	"net"
	"syscall"
	"time"

	"github.com/edgedb/edgedb-go/internal/buff"
	"github.com/edgedb/edgedb-go/internal/cache"
	"github.com/edgedb/edgedb-go/internal/cardinality"
	"github.com/edgedb/edgedb-go/internal/format"
	"github.com/edgedb/edgedb-go/internal/marshal"
	"github.com/edgedb/edgedb-go/internal/soc"
)

const defaultMaxTxRetries = 3

var rnd = rand.New(rand.NewSource(time.Now().UnixNano()))

func defaultBackoff(attempt int) time.Duration {
	backoff := math.Pow(2.0, float64(attempt)) * 100.0
	jitter := rnd.Float64() * 100.0
	return time.Duration(backoff+jitter) * time.Millisecond
}

// Action is work to be done in a transaction.
type Action func(context.Context, Tx) error

// Trier allows trying actions in a transaction.
type Trier interface {
	// TryTx runs an action in a transaction.
	// If the action returns an error the transaction is rolled back,
	// otherwise it is committed.
	TryTx(context.Context, Action) error

	// Retry does the same as TryTx
	// but retries the action if it might succeed on a subsequent attempt.
	Retry(context.Context, Action) error
}

// Executor allows querying the database.
type Executor interface {
	// Execute an EdgeQL command (or commands).
	Execute(context.Context, string) error

	// Query runs a query and returns the results.
	Query(context.Context, string, interface{}, ...interface{}) error

	// QueryOne runs a singleton-returning query and returns its element.
	// If the query executes successfully but doesn't return a result
	// a NoDataError is returned.
	QueryOne(context.Context, string, interface{}, ...interface{}) error

	// QueryJSON runs a query and return the results as JSON.
	QueryJSON(context.Context, string, *[]byte, ...interface{}) error

	// QueryOneJSON runs a singleton-returning query.
	// If the query executes successfully but doesn't have a result
	// a NoDataError is returned.
	QueryOneJSON(context.Context, string, *[]byte, ...interface{}) error
}

type baseConn struct {
	conn net.Conn

	// writeMemory is preallocated memory for payloads to be sent to the server
	writeMemory [1024]byte

	acquireReaderSignal chan struct{}
	readerChan          chan *buff.Reader

	typeIDCache   *cache.Cache
	inCodecCache  *cache.Cache
	outCodecCache *cache.Cache

	serverSettings map[string]string

	// isClosed is true when the connection has been closed by a user.
	isClosed bool

	cfg *connConfig
}

// connectWithTimeout makes a single attempt to connect to `addr`.
func connectWithTimeout(
	ctx context.Context,
	conn *baseConn,
	addr *dialArgs,
) error {
	var (
		cancel context.CancelFunc
		d      net.Dialer
		err    error
	)

	if conn.cfg.connectTimeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, conn.cfg.connectTimeout)
		defer cancel()
	}

	toBeDeserialized := make(chan *soc.Data, 2)
	r := buff.NewReader(toBeDeserialized)

	conn.conn, err = d.DialContext(ctx, addr.network, addr.address)
	if err != nil {
		goto handleError
	}

	conn.acquireReaderSignal = make(chan struct{}, 1)
	conn.readerChan = make(chan *buff.Reader, 1)
	go soc.Read(conn.conn, soc.NewMemPool(4, 256*1024), toBeDeserialized)

	err = conn.setDeadline(ctx)
	if err != nil {
		_ = conn.conn.Close()
		goto handleError
	}

	err = conn.connect(r, conn.cfg)
	if err != nil {
		_ = conn.conn.Close()
		goto handleError
	}

	err = conn.setDeadline(context.Background())
	if err != nil {
		_ = conn.conn.Close()
		goto handleError
	}

	if conn.releaseReader(r, nil) != nil {
		goto handleError
	}

	return nil

handleError:
	conn.conn = nil

	var errEDB Error
	var errNetOp *net.OpError
	var errDSN *net.DNSError

	switch {
	case errors.As(err, &errNetOp) && errNetOp.Timeout():
		return &clientConnectionTimeoutError{err: errNetOp}

	case errors.As(err, &errEDB):
		return err

	case errors.Is(err, syscall.ECONNREFUSED):
		fallthrough
	case errors.Is(err, syscall.ECONNABORTED):
		fallthrough
	case errors.Is(err, syscall.ECONNRESET):
		fallthrough
	case errors.As(err, &errDSN):
		fallthrough
	case errors.Is(err, syscall.ENOENT):
		return &clientConnectionFailedTemporarilyError{err: err}

	default:
		return &clientConnectionFailedError{err: err}
	}
}

// reconnect establishes a new connection with the server
// retrying the connection on failure.
// An error is returned if the `baseConn` was closed.
func (c *baseConn) reconnect(ctx context.Context) (err error) {
	if c.isClosed {
		return &interfaceError{msg: "Connection is closed"}
	}

	maxTime := time.Now().Add(c.cfg.waitUntilAvailable)
	if deadline, ok := ctx.Deadline(); ok && deadline.Before(maxTime) {
		maxTime = deadline
	}

	var connErr ClientConnectionError

	for i := 1; true; i++ {
		for _, addr := range c.cfg.addrs {
			err = connectWithTimeout(ctx, c, addr)
			if err == nil ||
				errors.Is(err, context.Canceled) ||
				errors.Is(err, context.DeadlineExceeded) ||
				!errors.As(err, &connErr) ||
				!connErr.HasTag("SHOULD_RECONNECT") ||
				(i > 1 && time.Now().After(maxTime)) {
				return err
			}
		}

		time.Sleep(time.Duration(10+rnd.Intn(200)) * time.Millisecond)
	}

	// this is unreachable, but the compiler can't tell...
	return err
}

func (c *baseConn) setDeadline(ctx context.Context) error {
	deadline, _ := ctx.Deadline()
	err := c.conn.SetDeadline(deadline)
	if err != nil {
		return &clientConnectionError{err: err}
	}

	return nil
}

func (c *baseConn) acquireReader(ctx context.Context) (*buff.Reader, error) {
	c.acquireReaderSignal <- struct{}{}

	select {
	case r := <-c.readerChan:
		if r.Err != nil {
			return nil, &clientConnectionError{err: r.Err}
		}

		return r, nil
	case <-ctx.Done():
		return nil, fmt.Errorf("edgedb: %w", ctx.Err())
	}
}

func (c *baseConn) releaseReader(r *buff.Reader, err error) error {
	if soc.IsPermanentNetErr(err) {
		_ = c.conn.Close()
		c.conn = nil
		return err
	}

	if e := c.setDeadline(context.Background()); e != nil {
		_ = c.conn.Close()
		c.conn = nil
		return e
	}

	go func() {
		for r.Next(c.acquireReaderSignal) {
			if e := c.fallThrough(r); e != nil {
				panic(e)
			}
		}

		c.readerChan <- r
	}()

	return err
}

// Close the db connection
func (c *baseConn) close() error {
	_, err := c.acquireReader(context.Background())
	if err != nil {
		_ = c.conn.Close()
		c.conn = nil
		return err
	}

	err = c.terminate()
	if err != nil {
		_ = c.conn.Close()
		c.conn = nil
		return err
	}

	err = c.conn.Close()
	c.conn = nil
	if err != nil {
		return &clientConnectionError{err: err}
	}

	c.isClosed = true
	return nil
}

// ensureConnection reconnects to the server if not connected.
func (c *baseConn) ensureConnection(ctx context.Context) error {
	if c.conn != nil && !c.isClosed {
		return nil
	}

	return c.reconnect(ctx)
}

func (c *baseConn) Execute(ctx context.Context, cmd string) error {
	if e := c.ensureConnection(ctx); e != nil {
		return e
	}

	r, err := c.acquireReader(ctx)
	if err != nil {
		return err
	}

	if e := c.setDeadline(ctx); e != nil {
		return e
	}

	return c.releaseReader(r, c.scriptFlow(r, cmd))
}

func (c *baseConn) Query(
	ctx context.Context,
	cmd string,
	out interface{},
	args ...interface{},
) error {
	if e := c.ensureConnection(ctx); e != nil {
		return e
	}

	val, err := marshal.ValueOfSlice(out)
	if err != nil {
		return &invalidArgumentError{msg: err.Error()}
	}

	q := query{
		cmd:     cmd,
		fmt:     format.Binary,
		expCard: cardinality.Many,
		args:    args,
	}

	r, err := c.acquireReader(ctx)
	if err != nil {
		return err
	}

	if e := c.setDeadline(ctx); e != nil {
		return e
	}

	return c.releaseReader(r, c.granularFlow(r, val, q))
}

func (c *baseConn) QueryOne(
	ctx context.Context,
	cmd string,
	out interface{},
	args ...interface{},
) (err error) {
	if e := c.ensureConnection(ctx); e != nil {
		return e
	}

	val, err := marshal.ValueOf(out)
	if err != nil {
		return &invalidArgumentError{msg: err.Error()}
	}

	q := query{
		cmd:     cmd,
		fmt:     format.Binary,
		expCard: cardinality.One,
		args:    args,
	}

	r, err := c.acquireReader(ctx)
	if err != nil {
		return err
	}

	if e := c.setDeadline(ctx); e != nil {
		return e
	}

	return c.releaseReader(r, c.granularFlow(r, val, q))
}

func (c *baseConn) QueryJSON(
	ctx context.Context,
	cmd string,
	out *[]byte,
	args ...interface{},
) error {
	if e := c.ensureConnection(ctx); e != nil {
		return e
	}

	val, err := marshal.ValueOf(out)
	if err != nil {
		return &invalidArgumentError{msg: err.Error()}
	}

	q := query{
		cmd:     cmd,
		fmt:     format.JSON,
		expCard: cardinality.Many,
		args:    args,
	}

	r, err := c.acquireReader(ctx)
	if err != nil {
		return err
	}

	if e := c.setDeadline(ctx); e != nil {
		return e
	}

	return c.releaseReader(r, c.granularFlow(r, val, q))
}

func (c *baseConn) QueryOneJSON(
	ctx context.Context,
	cmd string,
	out *[]byte,
	args ...interface{},
) error {
	if e := c.ensureConnection(ctx); e != nil {
		return e
	}

	val, err := marshal.ValueOf(out)
	if err != nil {
		return &invalidArgumentError{msg: err.Error()}
	}

	q := query{
		cmd:     cmd,
		fmt:     format.JSON,
		expCard: cardinality.One,
		args:    args,
	}

	r, err := c.acquireReader(ctx)
	if err != nil {
		return err
	}

	if e := c.setDeadline(ctx); e != nil {
		return e
	}

	return c.releaseReader(r, c.granularFlow(r, val, q))
}

func (c *baseConn) TryTx(ctx context.Context, action Action) error {
	tx := &transaction{conn: c, isolation: repeatableRead}
	if e := tx.start(ctx); e != nil {
		return e
	}

	if e := action(ctx, tx); e != nil {
		return firstError(e, tx.rollback(ctx))
	}

	return tx.commit(ctx)
}

func (c *baseConn) Retry(ctx context.Context, action Action) error {
	var edbErr Error

	for i := 0; i < defaultMaxTxRetries; i++ {
		tx := &transaction{conn: c, isolation: repeatableRead}

		if e := tx.start(ctx); e != nil {
			return e
		}

		err := action(ctx, tx)

		if err == nil {
			return tx.commit(ctx)
		}

		if e := tx.rollback(ctx); e != nil && !errors.As(e, &edbErr) {
			return e
		}

		if errors.As(err, &edbErr) &&
			edbErr.HasTag(ShouldRetry) &&
			(i+1 < defaultMaxTxRetries) {
			time.Sleep(defaultBackoff(i))
			continue
		}

		return err
	}

	panic("unreachable")
}
