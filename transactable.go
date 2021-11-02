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
	"time"
)

type transactableConn struct {
	*reconnectingConn
	txOpts    TxOptions
	retryOpts RetryOptions
}

// Execute an EdgeQL command (or commands).
func (c *transactableConn) Execute(ctx context.Context, cmd string) error {
	return c.scriptFlow(ctx, sfQuery{
		cmd:     cmd,
		headers: c.headers(),
	})
}

// Query runs a query and returns the results.
func (c *transactableConn) Query(
	ctx context.Context,
	cmd string,
	out interface{},
	args ...interface{},
) error {
	return runQuery(ctx, c, "Query", cmd, out, args)
}

// QuerySingle runs a singleton-returning query and returns its element.
// If the query executes successfully but doesn't return a result
// a NoDataError is returned.
func (c *transactableConn) QuerySingle(
	ctx context.Context,
	cmd string,
	out interface{},
	args ...interface{},
) error {
	return runQuery(ctx, c, "QuerySingle", cmd, out, args)
}

// QueryJSON runs a query and return the results as JSON.
func (c *transactableConn) QueryJSON(
	ctx context.Context,
	cmd string,
	out *[]byte,
	args ...interface{},
) error {
	return runQuery(ctx, c, "QueryJSON", cmd, out, args)
}

// QuerySingleJSON runs a singleton-returning query.
// If the query executes successfully but doesn't have a result
// a NoDataError is returned.
func (c *transactableConn) QuerySingleJSON(
	ctx context.Context,
	cmd string,
	out interface{},
	args ...interface{},
) error {
	return runQuery(ctx, c, "QuerySingleJSON", cmd, out, args)
}

func (c *transactableConn) granularFlow(
	ctx context.Context,
	q *gfQuery,
) error {
	var (
		err    error
		edbErr Error
	)

	for i := 1; true; i++ {
		if errors.As(err, &edbErr) && edbErr.HasTag(ShouldReconnect) {
			err = c.reconnect(ctx, true)
			if err != nil {
				goto Error
			}
		}

		err = c.reconnectingConn.granularFlow(ctx, q)

	Error:
		// q is a read only query if it has no capabilities
		// i.e. capabilities == 0. Read only queries are retryable.
		capabilities, ok := c.getCachedCapabilities(q)
		if ok && capabilities == 0 &&
			errors.As(err, &edbErr) &&
			edbErr.HasTag(ShouldRetry) {
			rule := c.retryOpts.ruleForException(edbErr)

			if i >= rule.attempts {
				return err
			}

			time.Sleep(rule.backoff(i))
			continue
		}

		return err
	}

	return &clientError{msg: "unreachable"}
}

// Tx runs an action in a transaction retrying failed actions if they might
// succeed on a subsequent attempt.
func (c *transactableConn) Tx(
	ctx context.Context,
	action TxBlock,
) (err error) {
	conn, err := c.borrow("transaction")
	if err != nil {
		return err
	}
	defer func() { err = firstError(err, c.unborrow()) }()

	var edbErr Error
	for i := 1; true; i++ {
		if errors.As(err, &edbErr) && edbErr.HasTag(ShouldReconnect) {
			err = c.reconnect(ctx, true)
			if err != nil {
				goto Error
			}
			// get the newly connected protocolConnection
			conn = c.conn
		}

		{
			tx := &Tx{
				borrowableConn: borrowableConn{conn: conn},
				txState:        &txState{},
				options:        c.txOpts,
			}
			err = tx.start(ctx)
			if err != nil {
				goto Error
			}

			err = action(ctx, tx)
			if err == nil {
				return tx.commit(ctx)
			} else if isClientConnectionError(err) {
				goto Error
			}

			if e := tx.rollback(ctx); e != nil && !errors.As(e, &edbErr) {
				return e
			}
		}

	Error:
		if errors.As(err, &edbErr) && edbErr.HasTag(ShouldRetry) {
			rule := c.retryOpts.ruleForException(edbErr)

			if i >= rule.attempts {
				return err
			}

			time.Sleep(rule.backoff(i))
			continue
		}

		return err
	}

	return &clientError{msg: "unreachable"}
}
