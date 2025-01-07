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

type reconnectingConn struct {
	borrowableConn
	cacheCollection
	cfg *connConfig

	// isClosed is true when the connection has been closed by a user.
	isClosed bool
}

// reconnect establishes a new connection with the server retrying the
// connection on failure. Calling reconnect() on a closed connection is an
// error.
func (c *reconnectingConn) reconnect(
	ctx context.Context,
	single bool,
) error {
	if c.isClosed {
		return &interfaceError{msg: "Connection is closed"}
	}

	maxTime := time.Now().Add(c.cfg.waitUntilAvailable)
	if deadline, ok := ctx.Deadline(); ok && deadline.Before(maxTime) {
		maxTime = deadline
	}

	var edbErr Error
	for {
		conn, err := connectWithTimeout(ctx, c.cfg, c.cacheCollection)
		if err == nil {
			c.conn = conn
			return nil
		}
		if single ||
			errors.Is(err, context.Canceled) ||
			errors.Is(err, context.DeadlineExceeded) ||
			!errors.As(err, &edbErr) ||
			!edbErr.Category(ClientConnectionError) ||
			!edbErr.HasTag(ShouldReconnect) ||
			time.Now().After(maxTime) {
			return err
		}

		time.Sleep(time.Duration(10+rnd.Intn(200)) * time.Millisecond)
	}
}

// ensureConnection reconnects to the server if not connected.
func (c *reconnectingConn) ensureConnection(ctx context.Context) error {
	if c.conn != nil && !c.conn.isClosed() && !c.isClosed {
		return nil
	}

	return c.reconnect(ctx, false)
}

func (c *reconnectingConn) scriptFlow(ctx context.Context, q *query) error {
	if e := c.ensureConnection(ctx); e != nil {
		return e
	}

	return c.borrowableConn.scriptFlow(ctx, q)
}

func (c *reconnectingConn) granularFlow(
	ctx context.Context,
	q *query,
) error {
	if e := c.ensureConnection(ctx); e != nil {
		return e
	}

	return c.borrowableConn.granularFlow(ctx, q)
}

// Close closes the connection. Connections are not usable after they are
// closed.
func (c *reconnectingConn) Close() (err error) {
	if c.isClosed {
		return &interfaceError{msg: "connection released more than once"}
	}

	c.isClosed = true
	if c.conn != nil && !c.conn.isClosed() {
		err = c.conn.close()
	}

	return err
}
