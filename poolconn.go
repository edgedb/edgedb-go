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

	"github.com/edgedb/edgedb-go/internal/soc"
)

// PoolConn is a pooled connection.
//
// Deprecated: use Pool.RetryingTx() or Pool.RawTx()
type PoolConn struct {
	pool      *Pool
	err       *error
	conn      *reconnectingConn
	txOpts    TxOptions
	retryOpts RetryOptions
}

// Release the connection back to its pool.
// Release returns an error if called more than once.
// A PoolConn is not usable after Release has been called.
func (c *PoolConn) Release() error {
	if c.pool == nil {
		msg := "connection released more than once"
		return &interfaceError{msg: msg}
	}

	var err error
	if c.err != nil {
		err = *c.err
	}

	err = c.pool.release(c.conn, err)
	c.pool = nil
	c.conn = nil
	c.err = nil

	return err
}

// checkErr records errors that indicate the connection should be closed
// so that this connection can be recycled when it is released.
func (c *PoolConn) checkErr(err error) {
	if soc.IsPermanentNetErr(err) {
		c.err = &err
		return
	}

	var edbErr Error
	if errors.As(err, &edbErr) && edbErr.Category(UnexpectedMessageError) {
		c.err = &err
	}
}

// Execute an EdgeQL command (or commands).
func (c *PoolConn) Execute(ctx context.Context, cmd string) error {
	err := c.conn.Execute(ctx, cmd)
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
	err := c.conn.Query(ctx, cmd, out, args...)
	c.checkErr(err)
	return err
}

// QueryOne runs a singleton-returning query and returns its element.
// If the query executes successfully but doesn't return a result
// a NoDataError is returned.
//
// Deprecated: use QuerySingle()
func (c *PoolConn) QueryOne(
	ctx context.Context,
	cmd string,
	out interface{},
	args ...interface{},
) error {
	return c.QuerySingle(ctx, cmd, out, args...)
}

// QuerySingle runs a singleton-returning query and returns its element.
// If the query executes successfully but doesn't return a result
// a NoDataError is returned.
func (c *PoolConn) QuerySingle(
	ctx context.Context,
	cmd string,
	out interface{},
	args ...interface{},
) error {
	err := c.conn.QuerySingle(ctx, cmd, out, args...)
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
	err := c.conn.QueryJSON(ctx, cmd, out, args...)
	c.checkErr(err)
	return err
}

// QueryOneJSON runs a singleton-returning query.
// If the query executes successfully but doesn't have a result
// a NoDataError is returned.
//
// Deprecated: use QuerySingleJSON()
func (c *PoolConn) QueryOneJSON(
	ctx context.Context,
	cmd string,
	out *[]byte,
	args ...interface{},
) error {
	return c.QuerySingleJSON(ctx, cmd, out, args...)
}

// QuerySingleJSON runs a singleton-returning query.
// If the query executes successfully but doesn't have a result
// a NoDataError is returned.
func (c *PoolConn) QuerySingleJSON(
	ctx context.Context,
	cmd string,
	out *[]byte,
	args ...interface{},
) error {
	err := c.conn.QuerySingleJSON(ctx, cmd, out, args...)
	c.checkErr(err)
	return err
}

// RawTx runs an action in a transaction.
// If the action returns an error the transaction is rolled back,
// otherwise it is committed.
func (c *PoolConn) RawTx(ctx context.Context, action Action) error {
	err := c.conn.rawTx(ctx, action, c.txOpts)
	c.checkErr(err)
	return err
}

// RetryingTx does the same as RawTx but retries failed actions
// if they might succeed on a subsequent attempt.
func (c *PoolConn) RetryingTx(ctx context.Context, action Action) error {
	err := c.conn.retryingTx(ctx, action, c.txOpts, c.retryOpts)
	c.checkErr(err)
	return err
}
