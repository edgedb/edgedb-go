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
	"fmt"

	"github.com/edgedb/edgedb-go/internal/header"
)

type borrowableConn struct {
	*baseConn
	reason string
}

func (c *borrowableConn) borrow(reason string) (*baseConn, error) {
	switch c.reason {
	case "":
		// this is the expected value
	case "transaction":
		return nil, &interfaceError{
			msg: "The connection is borrowed for a transaction. " +
				"Use the methods on the transaction object instead.",
		}
	case "subtransaction":
		return nil, &interfaceError{
			msg: "The transaction is borrowed for a subtransaction. " +
				"Use the methods on the subtransaction object instead.",
		}
	default:
		panic(fmt.Sprintf("unexpected reason: %q", c.reason))
	}

	switch reason {
	case "transaction", "subtransaction":
		c.reason = reason
		return c.baseConn, nil
	default:
		panic(fmt.Sprintf("unexpected reason: %q", reason))
	}
}

func (c *borrowableConn) unborrow() {
	if c.reason == "" {
		panic("not currently borrowed, can not unborrow")
	}

	c.reason = ""
}

func (c *borrowableConn) assertUnborrowed() error {
	switch c.reason {
	case "":
		return nil
	case "transaction":
		return &interfaceError{
			msg: "The connection is borrowed for a transaction. " +
				"Use the methods on the transaction object instead.",
		}
	case "subtransaction":
		return &interfaceError{
			msg: "The transaction is borrowed for a subtransaction. " +
				"Use the methods on the subtransaction object instead.",
		}
	default:
		panic(fmt.Sprintf("unexpected reason: %q", c.reason))
	}
}

func (c *borrowableConn) headers() msgHeaders {
	return msgHeaders{header.AllowCapabilities: noTxCapabilities}
}

func (c *borrowableConn) scriptFlow(ctx context.Context, q sfQuery) error {
	if e := c.assertUnborrowed(); e != nil {
		return e
	}

	return c.baseConn.scriptFlow(ctx, q)
}

func (c *borrowableConn) granularFlow(ctx context.Context, q *gfQuery) error {
	if e := c.assertUnborrowed(); e != nil {
		return e
	}

	return c.baseConn.granularFlow(ctx, q)
}

// Execute an EdgeQL command (or commands).
func (c *borrowableConn) Execute(ctx context.Context, cmd string) error {
	return c.scriptFlow(ctx, sfQuery{
		cmd:     cmd,
		headers: c.headers(),
	})
}

// Query runs a query and returns the results.
func (c *borrowableConn) Query(
	ctx context.Context,
	cmd string,
	out interface{},
	args ...interface{},
) error {
	return runQuery(ctx, c, "Query", cmd, out, args)
}

// QueryOne runs a singleton-returning query and returns its element.
// If the query executes successfully but doesn't return a result
// a NoDataError is returned.
func (c *borrowableConn) QueryOne(
	ctx context.Context,
	cmd string,
	out interface{},
	args ...interface{},
) error {
	return runQuery(ctx, c, "QueryOne", cmd, out, args)
}

// QueryJSON runs a query and return the results as JSON.
func (c *borrowableConn) QueryJSON(
	ctx context.Context,
	cmd string,
	out *[]byte,
	args ...interface{},
) error {
	return runQuery(ctx, c, "QueryJSON", cmd, out, args)
}

// QueryOneJSON runs a singleton-returning query.
// If the query executes successfully but doesn't have a result
// a NoDataError is returned.
func (c *borrowableConn) QueryOneJSON(
	ctx context.Context,
	cmd string,
	out *[]byte,
	args ...interface{},
) error {
	return runQuery(ctx, c, "QueryOneJSON", cmd, out, args)
}
