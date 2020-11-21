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

	"github.com/edgedb/edgedb-go/cache"
)

// Pool is a pool of connections.
type Pool struct {
	// mu locks the closeSignal channel in calls to Pool.Close()
	mu sync.Mutex

	// pool.Close() sends on this channel to signal closing the pool
	closeSignal chan chan struct{}

	// A buffered channel of connections ready for use.
	freeConns chan *baseConn

	// A buffered channel of structs representing unconnected capacity.
	potentialConns chan struct{}

	// Connections that did not produce a connection error
	// are sent back to the pool on releasedConns.
	releasedConns chan *baseConn

	opts          *Options
	typeIDCache   *cache.Cache
	inCodecCache  *cache.Cache
	outCodecCache *cache.Cache
}

// Connect a pool of connections to a server.
func Connect(ctx context.Context, opts Options) (*Pool, error) { // nolint
	// todo should 0 be a valid value for MinConns?
	if opts.MinConns < 1 {
		return nil, fmt.Errorf(
			"MinConns may not be less than 1, got: %v%w",
			opts.MinConns,
			ErrorConfiguration,
		)
	}

	if opts.MaxConns < opts.MinConns {
		return nil, fmt.Errorf(
			"MaxConns may not be less than MinConns%w",
			ErrorConfiguration,
		)
	}

	pool := &Pool{
		opts: &opts,

		closeSignal:    make(chan chan struct{}, 1),
		freeConns:      make(chan *baseConn, opts.MinConns),
		potentialConns: make(chan struct{}, opts.MaxConns),
		releasedConns:  make(chan *baseConn, opts.MaxConns),

		typeIDCache:   cache.New(1_000),
		inCodecCache:  cache.New(1_000),
		outCodecCache: cache.New(1_000),
	}

	for i := 0; i < opts.MaxConns-opts.MinConns; i++ {
		pool.potentialConns <- struct{}{}
	}

	errCh := make(chan error, opts.MinConns)

	for i := 0; i < opts.MinConns; i++ {
		go func() {
			conn, e := pool.newConn(ctx)
			pool.releasedConns <- conn
			errCh <- e
		}()
	}

	var err error
	for i := 0; i < opts.MinConns; i++ {
		e := <-errCh
		if e != nil {
			err = e
		}
	}

	if err != nil {
		pool.closeSignal <- make(chan struct{}, 1)
		pool.daemon()
		return nil, err
	}

	go pool.daemon()
	return pool, nil
}

func (p *Pool) daemon() {
	closeIn := p.closeSignal
	connCount := p.opts.MaxConns
	var closeOut chan struct{}

	for {
		select {
		case closeOut = <-closeIn:
			close(p.freeConns)
			close(p.potentialConns)

			for range p.potentialConns {
				connCount--
			}

			for conn := range p.freeConns {
				go conn.close() // nolint:errcheck
				connCount--
			}

			goto shutdown
		case conn := <-p.releasedConns:
			if conn == nil {
				p.potentialConns <- struct{}{}
				break
			}

			select {
			case p.freeConns <- conn:
			default:
				// we have MinConns idle so no need to keep this connection.
				go conn.close() // nolint:errcheck
				p.potentialConns <- struct{}{}
			}
		}
	}

shutdown:
	for connCount > 0 {
		conn := <-p.releasedConns
		if conn != nil {
			go conn.close() // nolint:errcheck
		}
		connCount--
	}

	closeOut <- struct{}{}
}

func (p *Pool) newConn(ctx context.Context) (*baseConn, error) {
	conn := &baseConn{
		typeIDCache:   p.typeIDCache,
		inCodecCache:  p.inCodecCache,
		outCodecCache: p.outCodecCache,
	}

	if err := connectOne(ctx, p.opts, conn); err != nil {
		return nil, err
	}

	return conn, nil
}

func (p *Pool) acquire(ctx context.Context) (*baseConn, error) {
	// force do nothing if context is expired
	select {
	case <-ctx.Done():
		return nil, ErrorContextExpired
	default:
	}

	// force using an existing connection over connecting a new socket.
	select {
	case conn, ok := <-p.freeConns:
		if !ok {
			return nil, ErrorPoolClosed
		}
		return conn, nil
	default:
	}

	select {
	case conn, ok := <-p.freeConns:
		if !ok {
			return nil, ErrorPoolClosed
		}
		return conn, nil
	case _, ok := <-p.potentialConns:
		if !ok {
			return nil, ErrorPoolClosed
		}

		conn, err := p.newConn(ctx)
		if err != nil {
			p.releasedConns <- nil
			return nil, err
		}
		return conn, nil
	case <-ctx.Done():
		return nil, ErrorContextExpired
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

func (p *Pool) release(conn *baseConn, err error) {
	if err == nil {
		p.releasedConns <- conn
		return
	}

	e, ok := err.(*net.OpError)
	if ok && e.Temporary() {
		p.releasedConns <- conn
		return
	}

	go conn.close() // nolint:errcheck
	p.releasedConns <- nil
}

// Close closes all connections in the pool.
// Calling close blocks until all acquired connections have been released.
// Returns an error if called more than once.
func (p *Pool) Close() error {
	p.mu.Lock()

	if p.closeSignal == nil {
		p.mu.Unlock()
		return ErrorPoolClosed
	}

	ch := make(chan struct{})

	p.closeSignal <- ch
	<-ch
	p.closeSignal = nil
	p.mu.Unlock()
	return nil
}

// Execute an EdgeQL command (or commands).
func (p *Pool) Execute(ctx context.Context, cmd string) (err error) {
	conn, err := p.acquire(ctx)
	if err != nil {
		return err
	}

	err = conn.Execute(ctx, cmd)
	p.release(conn, err)
	return err
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
	p.release(conn, err)
	return err
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
	p.release(conn, err)
	return err
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
	p.release(conn, err)
	return err
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
	p.release(conn, err)
	return err
}
