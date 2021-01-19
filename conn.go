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

	"github.com/edgedb/edgedb-go/internal/cache"
)

// Conn is a single connection to a server.
// Conn implementations are not safe for concurrent use.
type Conn interface {
	Executor
	Trier

	// Close closes the connection.
	// Connections are not usable after they are closed.
	Close() error
}

// connection is the standalone connection implementation.
type connection struct {
	*baseConn
	borrowable
}

func (c *connection) Close() error {
	return c.baseConn.close()
}

func (c *connection) Execute(ctx context.Context, cmd string) error {
	if e := c.assertUnborrowed(); e != nil {
		return e
	}

	return c.baseConn.Execute(ctx, cmd)
}

func (c *connection) Query(
	ctx context.Context,
	cmd string,
	out interface{},
	args ...interface{},
) error {
	if e := c.assertUnborrowed(); e != nil {
		return e
	}

	return c.baseConn.Query(ctx, cmd, out, args...)
}

func (c *connection) QueryOne(
	ctx context.Context,
	cmd string,
	out interface{},
	args ...interface{},
) error {
	if e := c.assertUnborrowed(); e != nil {
		return e
	}

	return c.baseConn.QueryOne(ctx, cmd, out, args...)
}

func (c *connection) QueryJSON(
	ctx context.Context,
	cmd string,
	out *[]byte,
	args ...interface{},
) error {
	if e := c.assertUnborrowed(); e != nil {
		return e
	}

	return c.baseConn.QueryJSON(ctx, cmd, out, args...)
}

func (c *connection) QueryOneJSON(
	ctx context.Context,
	cmd string,
	out *[]byte,
	args ...interface{},
) error {
	if e := c.assertUnborrowed(); e != nil {
		return e
	}

	return c.baseConn.QueryOneJSON(ctx, cmd, out, args...)
}

func (c *connection) TryTx(ctx context.Context, action Action) error {
	if e := c.borrow("transaction"); e != nil {
		return e
	}
	defer c.unborrow()

	return c.baseConn.TryTx(ctx, action)
}

func (c *connection) Retry(ctx context.Context, action Action) error {
	if e := c.borrow("transaction"); e != nil {
		return e
	}
	defer c.unborrow()

	return c.baseConn.Retry(ctx, action)
}

// ConnectOne establishes a connection to an EdgeDB server.
func ConnectOne(ctx context.Context, opts Options) (Conn, error) { // nolint:gocritic,lll
	return ConnectOneDSN(ctx, "", opts)
}

// ConnectOneDSN establishes a connection to an EdgeDB server.
func ConnectOneDSN(
	ctx context.Context,
	dsn string,
	opts Options, // nolint:gocritic
) (Conn, error) {
	config, err := parseConnectDSNAndArgs(dsn, &opts)
	if err != nil {
		return nil, err
	}

	conn := &baseConn{
		typeIDCache:   cache.New(1_000),
		inCodecCache:  cache.New(1_000),
		outCodecCache: cache.New(1_000),
		cfg:           config,
	}

	if err := conn.reconnect(ctx); err != nil {
		return nil, err
	}

	return &connection{baseConn: conn}, nil
}
