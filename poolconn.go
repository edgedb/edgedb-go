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
	"net"
)

// PoolConn is a pooled connection.
type PoolConn struct {
	pool *Pool
	err  error
	*baseConn
}

// Release the connection back to its pool. Panics if called more than once.
// PoolConn is not useable after Release has been called.
func (c *PoolConn) Release() error {
	if c.pool == nil {
		return ErrReleasedTwice
	}

	err := c.pool.release(c.baseConn, c.err)
	c.pool = nil
	c.baseConn = nil
	c.err = nil

	return err
}

func (c *PoolConn) checkErr(err error) {
	e, ok := err.(*net.OpError)
	if ok && !e.Temporary() {
		c.err = e
	}
}

// Execute an EdgeQL command (or commands).
func (c *PoolConn) Execute(ctx context.Context, cmd string) error {
	err := c.baseConn.Execute(ctx, cmd)
	c.checkErr(err)
	return err
}

// Query runs a query and returns the results.
func (c *PoolConn) Query(
	ctx context.Context,
	cmd string,
	out interface{},
	args ...interface{},
) error {
	err := c.baseConn.Query(ctx, cmd, out, args)
	c.checkErr(err)
	return err
}

// QueryOne runs a singleton-returning query and returns its element.
// If the query executes successfully but doesn't return a result
// ErrorZeroResults is returned.
func (c *PoolConn) QueryOne(
	ctx context.Context,
	cmd string,
	out interface{},
	args ...interface{},
) error {
	err := c.baseConn.QueryOne(ctx, cmd, out, args)
	c.checkErr(err)
	return err
}

// QueryJSON runs a query and return the results as JSON.
func (c *PoolConn) QueryJSON(
	ctx context.Context,
	cmd string,
	out *[]byte,
	args ...interface{},
) error {
	err := c.baseConn.QueryJSON(ctx, cmd, out, args)
	c.checkErr(err)
	return err
}

// QueryOneJSON runs a singleton-returning query
// and return its element in JSON.
// If the query executes successfully but doesn't return a result
// []byte{}, ErrorZeroResults is returned.
func (c *PoolConn) QueryOneJSON(
	ctx context.Context,
	cmd string,
	out *[]byte,
	args ...interface{},
) error {
	err := c.baseConn.QueryOneJSON(ctx, cmd, out, args)
	c.checkErr(err)
	return err
}
