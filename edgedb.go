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
	"net"
	"reflect"

	"github.com/fatih/pool"

	"github.com/edgedb/edgedb-go/cache"
	"github.com/edgedb/edgedb-go/marshal"
	"github.com/edgedb/edgedb-go/protocol/buff"
	"github.com/edgedb/edgedb-go/protocol/cardinality"
	"github.com/edgedb/edgedb-go/protocol/format"
)

// todo add examples

var (
	// todo should this be returned from Query() and QueryJSON()? :thinking:

	// ErrorZeroResults is returned when a query has no results.
	ErrorZeroResults = errors.New("zero results")
)

// Client client
type Client struct {
	pool           pool.Pool
	buf            *buff.Buff
	typeIDCache    *cache.Cache
	inCodecCache   *cache.Cache
	outCodecCache  *cache.Cache
	serverSettings map[string]string
}

// Close the db connection
func (c *Client) Close() (err error) {
	// todo send Terminate on each connection that needs to be closed.
	defer c.pool.Close()
	return nil
}

// Execute an EdgeQL command (or commands).
func (c *Client) Execute(ctx context.Context, query string) (err error) {
	conn, err := c.pool.Get()
	if err != nil {
		return err
	}

	defer func() {
		e := conn.Close()
		if err == nil {
			err = e
		}
	}()

	return c.scriptFlow(ctx, conn, query)
}

// QueryOne runs a singleton-returning query and returns its element.
// If the query executes successfully but doesn't return a result
// ErrorZeroResults is returned.
func (c *Client) QueryOne(
	ctx context.Context,
	cmd string,
	out interface{},
	args ...interface{},
) (err error) {
	val, err := marshal.ValueOf(out)
	if err != nil {
		return err
	}

	conn, err := c.pool.Get()
	if err != nil {
		return err
	}

	defer func() {
		e := conn.Close()
		if e != nil && err == nil {
			err = e
		}
	}()

	q := query{
		cmd:     cmd,
		fmt:     format.Binary,
		expCard: cardinality.One,
		args:    args,
	}

	err = c.granularFlow(ctx, conn, val, q)
	if err != nil {
		return err
	}

	return nil
}

// Query runs a query and returns the results.
func (c *Client) Query(
	ctx context.Context,
	cmd string,
	out interface{},
	args ...interface{},
) error {
	val, err := marshal.ValueOfSlice(out)
	if err != nil {
		return err
	}

	conn, err := c.pool.Get()
	if err != nil {
		return err
	}

	defer func() {
		e := conn.Close()
		if e != nil && err == nil {
			err = e
		}
	}()

	q := query{
		cmd:     cmd,
		fmt:     format.Binary,
		expCard: cardinality.Many,
		args:    args,
	}

	err = c.granularFlow(ctx, conn, val, q)
	if err != nil {
		return err
	}

	return nil
}

// QueryJSON runs a query and return the results as JSON.
func (c *Client) QueryJSON(
	ctx context.Context,
	cmd string,
	out *[]byte,
	args ...interface{},
) error {
	conn, err := c.pool.Get()
	if err != nil {
		return err
	}

	defer func() {
		e := conn.Close()
		if e != nil && err == nil {
			err = e
		}
	}()

	q := query{
		cmd:     cmd,
		fmt:     format.JSON,
		expCard: cardinality.Many,
		args:    args,
	}

	val := reflect.ValueOf(out).Elem()
	err = c.granularFlow(ctx, conn, val, q)
	if err != nil {
		return err
	}

	return nil
}

// QueryOneJSON runs a singleton-returning query
// and return its element in JSON.
// If the query executes successfully but doesn't return a result
// []byte{}, ErrorZeroResults is returned.
func (c *Client) QueryOneJSON(
	ctx context.Context,
	cmd string,
	out *[]byte,
	args ...interface{},
) error {
	conn, err := c.pool.Get()
	if err != nil {
		return err
	}

	defer func() {
		e := conn.Close()
		if e != nil && err == nil {
			err = e
		}
	}()

	q := query{
		cmd:     cmd,
		fmt:     format.JSON,
		expCard: cardinality.One,
		args:    args,
	}

	val := reflect.ValueOf(out).Elem()
	err = c.granularFlow(ctx, conn, val, q)
	if err != nil {
		return err
	}

	if len(*out) == 0 {
		return ErrorZeroResults
	}

	return nil
}

// Connect establishes a connection to an EdgeDB server.
func Connect(ctx context.Context, opts Options) (client *Client, err error) {
	// todo making the pool bigger slows down the tests,
	// and uses way more memory :thinking:
	p, err := pool.NewChannelPool(1, 1, func() (conn net.Conn, e error) {
		var d net.Dialer
		// todo closing over the context is the wrong thing to do.
		conn, e = d.DialContext(ctx, opts.network(), opts.address())
		if e != nil {
			return nil, e
		}

		e = connect(ctx, conn, &opts)
		return conn, e
	})

	if err != nil {
		return nil, err
	}

	client = &Client{
		pool:          p,
		buf:           buff.New(make([]byte, 8192)),
		typeIDCache:   cache.New(1_000),
		inCodecCache:  cache.New(1_000),
		outCodecCache: cache.New(1_000),
	}
	return client, nil
}

func writeAndRead(
	ctx context.Context,
	conn net.Conn,
	buf *[]byte,
) (err error) {
	defer func() {
		// todo don't mark unusable on timeout
		if err != nil {
			conn.(*pool.PoolConn).MarkUnusable()
		}
	}()

	deadline, _ := ctx.Deadline()
	err = conn.SetDeadline(deadline)
	if err != nil {
		return err
	}

	_, err = conn.Write(*buf)
	if err != nil {
		return err
	}

	// expand slice length to full capacity
	*buf = (*buf)[:cap(*buf)]

	n, err := conn.Read(*buf)
	*buf = (*buf)[:n]

	if n < cap(*buf) {
		return err
	}

	n = 1024 // todo evaluate temporary buffer size
	tmp := make([]byte, n)
	for n == 1024 {
		n, err = conn.Read(tmp)
		*buf = append(*buf, tmp[:n]...)
	}

	return err
}
