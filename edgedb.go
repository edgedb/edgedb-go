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
	"math/rand"
	"net"
	"syscall"
	"time"

	"github.com/edgedb/edgedb-go/internal/buff"
	"github.com/edgedb/edgedb-go/internal/cache"
	"github.com/edgedb/edgedb-go/internal/soc"
)

var rnd = rand.New(rand.NewSource(time.Now().UnixNano()))

// Action is work to be done in a transaction.
type Action func(context.Context, *Tx) error

type baseConn struct {
	conn             net.Conn
	errUnrecoverable error

	// writeMemory is preallocated memory for payloads to be sent to the server
	writeMemory [1024]byte

	acquireReaderSignal chan struct{}
	readerChan          chan *buff.Reader

	typeIDCache   *cache.Cache
	inCodecCache  *cache.Cache
	outCodecCache *cache.Cache

	protocolVersion version

	// indicates whether the protocol version supports
	// the EXPLICIT_OBJECTIDS header.
	explicitIDs bool

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

func (c *baseConn) setDeadline(ctx context.Context) error {
	deadline, _ := ctx.Deadline()
	err := c.conn.SetDeadline(deadline)
	if err != nil {
		return &clientConnectionError{err: err}
	}

	return nil
}

func (c *baseConn) acquireReader(ctx context.Context) (*buff.Reader, error) {
	if c.errUnrecoverable != nil {
		return nil, c.errUnrecoverable
	}

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
				c.errUnrecoverable = e
				_ = c.conn.Close()
				return
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

	return nil
}

func (c *baseConn) ScriptFlow(ctx context.Context, q sfQuery) error {
	r, err := c.acquireReader(ctx)
	if err != nil {
		return err
	}

	if e := c.setDeadline(ctx); e != nil {
		return e
	}

	return c.releaseReader(r, c.scriptFlow(r, q))
}

func (c *baseConn) GranularFlow(ctx context.Context, q *gfQuery) error {
	r, err := c.acquireReader(ctx)
	if err != nil {
		return err
	}

	if e := c.setDeadline(ctx); e != nil {
		return e
	}

	return c.releaseReader(r, c.granularFlow(r, q))
}
