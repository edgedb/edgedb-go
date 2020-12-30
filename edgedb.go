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

	"github.com/edgedb/edgedb-go/internal/buff"
	"github.com/edgedb/edgedb-go/internal/cache"
	"github.com/edgedb/edgedb-go/internal/cardinality"
	"github.com/edgedb/edgedb-go/internal/format"
	"github.com/edgedb/edgedb-go/internal/marshal"
	"github.com/edgedb/edgedb-go/internal/soc"
)

// todo add examples

type baseConn struct {
	conn   net.Conn
	writer *buff.Writer

	acquireReaderSignal chan struct{}
	readerChan          chan *buff.Reader

	typeIDCache   *cache.Cache
	inCodecCache  *cache.Cache
	outCodecCache *cache.Cache

	serverSettings map[string]string
}

// ConnectOne establishes a connection to an EdgeDB server.
func ConnectOne(ctx context.Context, opts Options) (*Conn, error) { // nolint:gocritic,lll
	return ConnectOneDSN(ctx, "", opts)
}

// ConnectOneDSN establishes a connection to an EdgeDB server.
func ConnectOneDSN(
	ctx context.Context,
	dsn string,
	opts Options, // nolint:gocritic
) (*Conn, error) {
	conn := &baseConn{
		typeIDCache:   cache.New(1_000),
		inCodecCache:  cache.New(1_000),
		outCodecCache: cache.New(1_000),
	}

	config, err := parseConnectDSNAndArgs(dsn, &opts)
	if err != nil {
		return nil, err
	}

	if err := connectOne(ctx, config, conn); err != nil {
		return nil, err
	}

	return &Conn{*conn}, nil
}

// connectOne expectes a singleConn that has a nil net.Conn.
func connectOne(ctx context.Context, cfg *connConfig, conn *baseConn) error {
	var (
		d   net.Dialer
		err error
	)

	conn.writer = buff.NewWriter()

	for _, addr := range cfg.addrs { // nolint:gocritic
		conn.conn, err = d.DialContext(ctx, addr.network, addr.address)
		if err != nil {
			continue
		}

		toBeDeserialized := make(chan *soc.Data, 2)
		r := buff.NewReader(toBeDeserialized)
		go soc.Read(conn.conn, soc.NewMemPool(4, 256*1024), toBeDeserialized)

		err = conn.setDeadline(ctx)
		if err != nil {
			_ = conn.conn.Close()
			continue
		}

		err = conn.connect(r, cfg)
		if err != nil {
			_ = conn.conn.Close()
			continue
		}

		err = conn.setDeadline(context.Background())
		if err != nil {
			_ = conn.conn.Close()
			continue
		}

		conn.acquireReaderSignal = make(chan struct{}, 1)
		conn.readerChan = make(chan *buff.Reader, 1)
		err = conn.releaseReader(r, nil)
		if err != nil {
			continue
		}

		return nil
	}

	conn.conn = nil
	return err
}

func (c *baseConn) setDeadline(ctx context.Context) error {
	deadline, _ := ctx.Deadline()
	return wrapError(c.conn.SetDeadline(deadline))
}

func (c *baseConn) acquireReader(ctx context.Context) (*buff.Reader, error) {
	c.acquireReaderSignal <- struct{}{}

	select {
	case r := <-c.readerChan:
		if r.Err != nil {
			return nil, wrapError(r.Err)
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
		return err
	}

	err = c.terminate()
	if err != nil {
		_ = c.conn.Close()
		return err
	}

	err = c.conn.Close()
	if err != nil {
		return wrapError(err)
	}

	return nil
}

// Execute an EdgeQL command (or commands).
func (c *baseConn) Execute(ctx context.Context, cmd string) error {
	r, err := c.acquireReader(ctx)
	if err != nil {
		return err
	}

	if e := c.setDeadline(ctx); e != nil {
		return e
	}

	return c.releaseReader(r, c.scriptFlow(r, cmd))
}

// QueryOne runs a singleton-returning query and returns its element.
// If the query executes successfully but doesn't return a result
// ErrorZeroResults is returned.
func (c *baseConn) QueryOne(
	ctx context.Context,
	cmd string,
	out interface{},
	args ...interface{},
) (err error) {
	val, err := marshal.ValueOf(out)
	if err != nil {
		return newErrorFromCode(invalidArgumentErrorCode, err.Error())
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

// Query runs a query and returns the results.
func (c *baseConn) Query(
	ctx context.Context,
	cmd string,
	out interface{},
	args ...interface{},
) error {
	val, err := marshal.ValueOfSlice(out)
	if err != nil {
		return newErrorFromCode(invalidArgumentErrorCode, err.Error())
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

// QueryJSON runs a query and return the results as JSON.
func (c *baseConn) QueryJSON(
	ctx context.Context,
	cmd string,
	out *[]byte,
	args ...interface{},
) error {
	val, err := marshal.ValueOf(out)
	if err != nil {
		return newErrorFromCode(invalidArgumentErrorCode, err.Error())
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

// QueryOneJSON runs a singleton-returning query
// and return its element in JSON.
// If the query executes successfully but doesn't return a result
// []byte{}, ErrorZeroResults is returned.
func (c *baseConn) QueryOneJSON(
	ctx context.Context,
	cmd string,
	out *[]byte,
	args ...interface{},
) error {
	val, err := marshal.ValueOf(out)
	if err != nil {
		return newErrorFromCode(invalidArgumentErrorCode, err.Error())
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
