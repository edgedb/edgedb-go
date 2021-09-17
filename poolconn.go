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
)

// PoolConn is a pooled connection.
//
// Deprecated: use the query methods on Pool instead
type PoolConn struct {
	transactableConn
	isClosed *bool
	pool     *Pool
	err      *error
}

// Release the connection back to its pool.
// Release returns an error if called more than once.
// A PoolConn is not usable after Release has been called.
func (c *PoolConn) Release() error {
	if *c.isClosed {
		return &interfaceError{msg: "connection released more than once"}
	}

	var err error
	if c.err != nil {
		err = *c.err
	}

	err = c.pool.release(&c.transactableConn, err)
	c.pool = nil
	c.transactableConn = transactableConn{}
	c.err = nil
	*c.isClosed = true

	return err
}

// checkErr records errors that indicate the connection should be closed
// so that this connection can be recycled when it is released.
func (c *PoolConn) checkErr(err error) {
	var edbErr Error
	if errors.As(err, &edbErr) &&
		(edbErr.Category(UnexpectedMessageError) ||
			edbErr.Category(ClientConnectionError)) {
		c.err = &err
	}
}

// Execute an EdgeQL command (or commands).
func (c *PoolConn) Execute(ctx context.Context, cmd string) error {
	err := c.transactableConn.Execute(ctx, cmd)
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
	err := c.transactableConn.Query(ctx, cmd, out, args...)
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
	err := c.transactableConn.QuerySingle(ctx, cmd, out, args...)
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
	err := c.transactableConn.QueryJSON(ctx, cmd, out, args...)
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
	err := c.transactableConn.QuerySingleJSON(ctx, cmd, out, args...)
	c.checkErr(err)
	return err
}

// RawTx runs an action in a transaction.
// If the action returns an error the transaction is rolled back,
// otherwise it is committed.
func (c *PoolConn) RawTx(ctx context.Context, action TxBlock) error {
	err := c.transactableConn.RawTx(ctx, action)
	c.checkErr(err)
	return err
}

// RetryingTx does the same as RawTx but retries failed actions
// if they might succeed on a subsequent attempt.
func (c *PoolConn) RetryingTx(ctx context.Context, action TxBlock) error {
	err := c.transactableConn.RetryingTx(ctx, action)
	c.checkErr(err)
	return err
}
