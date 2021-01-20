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
	"fmt"
	"time"
)

type reconnectingConn struct {
	borrowReason string

	// isClosed is true when the connection has been closed by a user.
	isClosed bool
	conn     *baseConn
}

func (b *reconnectingConn) assertUnborrowed() error {
	switch b.borrowReason {
	case "transaction":
		return &interfaceError{
			msg: "Connection is borrowed for a transaction. " +
				"Use the methods on transaction object instead.",
		}
	case "":
		return nil
	default:
		panic(fmt.Sprintf("unexpected reason: %q", b.borrowReason))
	}
}

func (b *reconnectingConn) borrow(reason string) error {
	if b.borrowReason != "" {
		msg := "connection is already borrowed for " + b.borrowReason
		return &interfaceError{msg: msg}
	}

	if reason != "transaction" {
		panic(fmt.Sprintf("unexpected reason: %q", reason))
	}

	b.borrowReason = reason
	return nil
}

func (b *reconnectingConn) unborrow() {
	if b.borrowReason == "" {
		panic("not currently borrowed, can not unborrow")
	}

	b.borrowReason = ""
}

// reconnect establishes a new connection with the server
// retrying the connection on failure.
// An error is returned if the `baseConn` was closed.
func (b *reconnectingConn) reconnect(ctx context.Context) (err error) {
	if b.isClosed {
		return &interfaceError{msg: "Connection is closed"}
	}

	maxTime := time.Now().Add(b.conn.cfg.waitUntilAvailable)
	if deadline, ok := ctx.Deadline(); ok && deadline.Before(maxTime) {
		maxTime = deadline
	}

	var edbErr Error

	for i := 1; true; i++ {
		for _, addr := range b.conn.cfg.addrs {
			err = connectWithTimeout(ctx, b.conn, addr)
			if err == nil ||
				errors.Is(err, context.Canceled) ||
				errors.Is(err, context.DeadlineExceeded) ||
				!errors.As(err, &edbErr) ||
				!edbErr.Category(ClientConnectionError) ||
				!edbErr.HasTag("SHOULD_RECONNECT") ||
				(i > 1 && time.Now().After(maxTime)) {
				return err
			}
		}

		time.Sleep(time.Duration(10+rnd.Intn(200)) * time.Millisecond)
	}

	panic("unreachable")
}

// ensureConnection reconnects to the server if not connected.
func (b *reconnectingConn) ensureConnection(ctx context.Context) error {
	if b.conn != nil && !b.isClosed {
		return nil
	}

	return b.reconnect(ctx)
}

func (b *reconnectingConn) Execute(ctx context.Context, cmd string) error {
	if e := b.assertUnborrowed(); e != nil {
		return e
	}

	if e := b.ensureConnection(ctx); e != nil {
		return e
	}

	return b.conn.Execute(ctx, cmd)
}

func (b *reconnectingConn) Query(
	ctx context.Context,
	cmd string,
	out interface{},
	args ...interface{},
) error {
	if e := b.assertUnborrowed(); e != nil {
		return e
	}

	if e := b.ensureConnection(ctx); e != nil {
		return e
	}

	return b.conn.Query(ctx, cmd, out, args...)
}

func (b *reconnectingConn) QueryOne(
	ctx context.Context,
	cmd string,
	out interface{},
	args ...interface{},
) error {
	if e := b.assertUnborrowed(); e != nil {
		return e
	}

	if e := b.ensureConnection(ctx); e != nil {
		return e
	}

	return b.conn.QueryOne(ctx, cmd, out, args...)
}

func (b *reconnectingConn) QueryJSON(
	ctx context.Context,
	cmd string,
	out *[]byte,
	args ...interface{},
) error {
	if e := b.assertUnborrowed(); e != nil {
		return e
	}

	if e := b.ensureConnection(ctx); e != nil {
		return e
	}

	return b.conn.QueryJSON(ctx, cmd, out, args...)
}

func (b *reconnectingConn) QueryOneJSON(
	ctx context.Context,
	cmd string,
	out *[]byte,
	args ...interface{},
) error {
	if e := b.assertUnborrowed(); e != nil {
		return e
	}

	if e := b.ensureConnection(ctx); e != nil {
		return e
	}

	return b.conn.QueryOneJSON(ctx, cmd, out, args...)
}

func (b *reconnectingConn) TryTx(ctx context.Context, action Action) error {
	if e := b.borrow("transaction"); e != nil {
		return e
	}
	defer b.unborrow()

	if e := b.ensureConnection(ctx); e != nil {
		return e
	}

	tx := &transaction{conn: b.conn, isolation: repeatableRead}
	if e := tx.start(ctx); e != nil {
		return e
	}

	if e := action(ctx, tx); e != nil {
		return firstError(e, tx.rollback(ctx))
	}

	return tx.commit(ctx)
}

func (b *reconnectingConn) Retry(ctx context.Context, action Action) error {
	if e := b.borrow("transaction"); e != nil {
		return e
	}
	defer b.unborrow()

	var edbErr Error

	for i := 0; i < defaultMaxTxRetries; i++ {
		if e := b.ensureConnection(ctx); e != nil {
			return e
		}

		tx := &transaction{conn: b.conn, isolation: repeatableRead}
		if e := tx.start(ctx); e != nil {
			return e
		}

		err := action(ctx, tx)
		if err == nil {
			return tx.commit(ctx)
		}

		if e := tx.rollback(ctx); e != nil && !errors.As(e, &edbErr) {
			return e
		}

		if errors.As(err, &edbErr) &&
			edbErr.HasTag(ShouldRetry) &&
			(i+1 < defaultMaxTxRetries) {
			time.Sleep(defaultBackoff(i))
			continue
		}

		return err
	}

	panic("unreachable")
}

func (b *reconnectingConn) close() error {
	b.isClosed = true
	return b.conn.close()
}
