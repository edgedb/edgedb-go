// This source file is part of the EdgeDB open source project.
//
// Copyright EdgeDB Inc. and the EdgeDB authors.
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

package gel

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

func (c *transactableConn) granularFlow(ctx context.Context, q *query) error {
	var (
		err    error
		edbErr Error
	)

	for i := 1; true; i++ {
		if errors.As(err, &edbErr) && c.conn.soc.Closed() {
			err = c.reconnect(ctx, true)
			if err != nil {
				goto Error
			}
		}

		err = c.reconnectingConn.granularFlow(ctx, q)

	Error:
		// q is a read only query if it has no capabilities
		// i.e. capabilities == 0. Read only queries are always
		// retryable, mutation queries are retryable if the
		// error explicitly indicates a transaction conflict.
		capabilities, ok := c.getCachedCapabilities(q)
		if ok &&
			errors.As(err, &edbErr) &&
			edbErr.HasTag(ShouldRetry) &&
			(capabilities == 0 || edbErr.Category(TransactionConflictError)) {
			rule, e := c.retryOpts.ruleForException(edbErr)
			if e != nil {
				return e
			}

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

func (c *transactableConn) tx(
	ctx context.Context,
	action TxBlock,
	state map[string]interface{},
	warningHandler WarningHandler,
) (err error) {
	conn, err := c.borrow("transaction")
	if err != nil {
		return err
	}
	defer func() { err = firstError(err, c.unborrow()) }()

	var edbErr Error
	for i := 1; true; i++ {
		if errors.As(err, &edbErr) && c.conn.soc.Closed() {
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
				state:          state,
				warningHandler: warningHandler,
			}
			err = tx.start(ctx)
			if err != nil {
				goto Error
			}

			err = action(ctx, tx)
			if err == nil {
				err = tx.commit(ctx)
				if errors.As(err, &edbErr) &&
					edbErr.Category(TransactionError) &&
					edbErr.HasTag(ShouldRetry) {
					goto Error
				}
				return err
			} else if isClientConnectionError(err) {
				goto Error
			}

			if e := tx.rollback(ctx); e != nil && !errors.As(e, &edbErr) {
				return e
			}
		}

	Error:
		if errors.As(err, &edbErr) && edbErr.HasTag(ShouldRetry) {
			rule, e := c.retryOpts.ruleForException(edbErr)
			if e != nil {
				return e
			}

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
