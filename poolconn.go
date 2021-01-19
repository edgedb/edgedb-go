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

	"github.com/edgedb/edgedb-go/internal/soc"
)

// PoolConn is a pooled connection.
type PoolConn interface {
	Executor
	Trier

	// Release the connection back to its pool.
	// Release returns an error if called more than once.
	// A PoolConn is not usable after Release has been called.
	Release() error
}

type poolConn struct {
	pool *pool
	err  error
	*baseConn
	borrowable
}

func (c *poolConn) Release() error {
	if c.pool == nil {
		msg := "connection released more than once"
		return &interfaceError{msg: msg}
	}

	err := c.pool.release(c.baseConn, c.err)
	c.pool = nil
	c.baseConn = nil
	c.err = nil

	return err
}

func (c *poolConn) checkErr(err error) {
	if soc.IsPermanentNetErr(err) {
		c.err = err
	}
}

func (c *poolConn) Execute(ctx context.Context, cmd string) error {
	if e := c.assertUnborrowed(); e != nil {
		return e
	}

	err := c.baseConn.Execute(ctx, cmd)
	c.checkErr(err)
	return err
}

func (c *poolConn) Query(
	ctx context.Context,
	cmd string,
	out interface{},
	args ...interface{},
) error {
	if e := c.assertUnborrowed(); e != nil {
		return e
	}

	err := c.baseConn.Query(ctx, cmd, out, args...)
	c.checkErr(err)
	return err
}

func (c *poolConn) QueryOne(
	ctx context.Context,
	cmd string,
	out interface{},
	args ...interface{},
) error {
	if e := c.assertUnborrowed(); e != nil {
		return e
	}

	err := c.baseConn.QueryOne(ctx, cmd, out, args...)
	c.checkErr(err)
	return err
}

func (c *poolConn) QueryJSON(
	ctx context.Context,
	cmd string,
	out *[]byte,
	args ...interface{},
) error {
	if e := c.assertUnborrowed(); e != nil {
		return e
	}

	err := c.baseConn.QueryJSON(ctx, cmd, out, args...)
	c.checkErr(err)
	return err
}

func (c *poolConn) QueryOneJSON(
	ctx context.Context,
	cmd string,
	out *[]byte,
	args ...interface{},
) error {
	if e := c.assertUnborrowed(); e != nil {
		return e
	}

	err := c.baseConn.QueryOneJSON(ctx, cmd, out, args...)
	c.checkErr(err)
	return err
}

func (c *poolConn) TryTx(ctx context.Context, action Action) error {
	if e := c.borrow("transaction"); e != nil {
		return e
	}
	defer c.unborrow()

	return c.baseConn.TryTx(ctx, action)
}

func (c *poolConn) Retry(ctx context.Context, action Action) error {
	if e := c.borrow("transaction"); e != nil {
		return e
	}
	defer c.unborrow()

	return c.baseConn.Retry(ctx, action)
}
