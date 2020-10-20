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

// todo add context.Context

import (
	"context"
	"errors"
	"net"

	"github.com/fatih/pool"

	"github.com/edgedb/edgedb-go/edgedb/marshal"
	"github.com/edgedb/edgedb-go/edgedb/protocol/codecs"
	"github.com/edgedb/edgedb-go/edgedb/protocol/format"
)

// todo add examples

var (
	// todo should this be returned from Query() and QueryJSON()? :thinking:

	// ErrorZeroResults is returned when a query has no results.
	ErrorZeroResults = errors.New("zero results")
)

type queryCodecs struct {
	in  codecs.DecodeEncoder
	out codecs.DecodeEncoder
}

type queryCacheKey struct {
	query  string
	format int
}

// todo rename Conn to Client

// Client client
type Client struct {
	pool   pool.Pool
	secret []byte

	// todo caches are not thread safe
	codecCache codecs.CodecLookup
	queryCache map[queryCacheKey]queryCodecs
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

	return scriptFlow(ctx, conn, query)
}

// QueryOne runs a singleton-returning query and return its element.
// If the query executes successfully but doesn't return a result
// ErrorZeroResults is returned.
func (c *Client) QueryOne(
	ctx context.Context,
	query string,
	out interface{},
	args ...interface{},
) (err error) {
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

	// todo assert cardinality
	result, err := c.granularFlow(ctx, conn, query, format.Binary, args)
	if err != nil {
		return err
	}

	if len(result) == 0 {
		return ErrorZeroResults
	}

	marshal.Marshal(&out, result[0])
	return nil
}

// Query runs a query and returns the results.
func (c *Client) Query(
	ctx context.Context,
	query string,
	out interface{},
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

	// todo assert that out is a pointer to a slice
	result, err := c.granularFlow(ctx, conn, query, format.Binary, args)
	if err != nil {
		return err
	}

	marshal.Marshal(&out, result)
	return nil
}

// QueryJSON runs a query and return the results as JSON.
func (c *Client) QueryJSON(
	ctx context.Context,
	query string,
	args ...interface{},
) ([]byte, error) {
	conn, err := c.pool.Get()
	if err != nil {
		return nil, err
	}

	defer func() {
		e := conn.Close()
		if e != nil && err == nil {
			err = e
		}
	}()

	result, err := c.granularFlow(ctx, conn, query, format.JSON, args)
	if err != nil {
		return nil, err
	}

	return []byte(result[0].(string)), nil
}

// QueryOneJSON runs a singleton-returning query
// and return its element in JSON.
// If the query executes successfully but doesn't return a result
// []byte{}, ErrorZeroResults is returned.
func (c *Client) QueryOneJSON(
	ctx context.Context,
	query string,
	args ...interface{},
) ([]byte, error) {
	conn, err := c.pool.Get()
	if err != nil {
		return nil, err
	}

	defer func() {
		e := conn.Close()
		if e != nil && err == nil {
			err = e
		}
	}()

	// todo assert cardinally
	result, err := c.granularFlow(ctx, conn, query, format.JSON, args)
	if err != nil {
		return nil, err
	}

	jsonStr := result[0].(string)
	if jsonStr == "[]" {
		return nil, ErrorZeroResults
	}

	return []byte(jsonStr[1 : len(jsonStr)-1]), nil
}

// Connect establishes a connection to an EdgeDB server.
func Connect(ctx context.Context, opts Options) (client *Client, err error) {
	// todo making the pool bigger slows down the tests,
	// and uses way more memory :thinking:
	p, err := pool.NewChannelPool(1, 1, func() (conn net.Conn, e error) {
		var d net.Dialer
		// todo closing over the context is the wrong thing to do.
		conn, e = d.DialContext(ctx, opts.socType(), opts.dialHost())
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
		p,
		nil,
		codecs.CodecLookup{},
		map[queryCacheKey]queryCodecs{},
	}

	return client, nil
}

func writeAndRead(
	ctx context.Context,
	conn net.Conn,
	bts []byte,
) (rcv []byte, err error) {
	defer func() {
		// todo don't mark unusable on timeout
		if err != nil {
			conn.(*pool.PoolConn).MarkUnusable()
		}
	}()

	deadline, _ := ctx.Deadline()
	err = conn.SetDeadline(deadline)
	if err != nil {
		return nil, err
	}

	_, err = conn.Write(bts)
	if err != nil {
		return nil, err
	}

	rcv = []byte{}
	n := 1024 // todo evaluate buffer size
	for n == 1024 {
		tmp := make([]byte, 1024)
		n, err = conn.Read(tmp)
		rcv = append(rcv, tmp[:n]...)
	}

	return rcv, err
}
