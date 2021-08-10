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

	"github.com/edgedb/edgedb-go/internal/header"
)

var (
	noTxCapabilities = header.NewAllowCapabilitiesWithout(
		header.AllowCapabilitieTransaction,
	)
)

type reconnectingConn struct {
	borrowableConn

	// isClosed is true when the connection has been closed by a user.
	isClosed bool
}

// reconnect establishes a new connection with the server
// retrying the connection on failure.
// An error is returned if the `baseConn` was closed.
func (c *reconnectingConn) reconnect(ctx context.Context) (err error) {
	if c.isClosed {
		return &interfaceError{msg: "Connection is closed"}
	}

	maxTime := time.Now().Add(c.cfg.waitUntilAvailable)
	if deadline, ok := ctx.Deadline(); ok && deadline.Before(maxTime) {
		maxTime = deadline
	}

	var edbErr Error

	for i := 1; true; i++ {
		for _, addr := range c.cfg.addrs {
			err = connectWithTimeout(ctx, c.borrowableConn.baseConn, addr)
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
func (c *reconnectingConn) ensureConnection(ctx context.Context) error {
	if c.netConn != nil && !c.isClosed {
		return nil
	}

	return c.reconnect(ctx)
}

func (c *reconnectingConn) scriptFlow(ctx context.Context, q sfQuery) error {
	if e := c.ensureConnection(ctx); e != nil {
		return e
	}

	return c.borrowableConn.scriptFlow(ctx, q)
}

func (c *reconnectingConn) granularFlow(
	ctx context.Context,
	q *gfQuery,
) error {
	if e := c.ensureConnection(ctx); e != nil {
		return e
	}

	return c.borrowableConn.granularFlow(ctx, q)
}

func (c *reconnectingConn) close() error {
	c.isClosed = true
	return c.borrowableConn.close()
}
