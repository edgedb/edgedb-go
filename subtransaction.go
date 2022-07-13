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
)

type subtxable interface {
	borrow(string) (*protocolConnection, error)
	unborrow() error
	txOptions() TxOptions
	txstate() *txState
}

func runSubtx(
	ctx context.Context,
	action SubtxBlock,
	c subtxable,
	state map[string]interface{},
) (err error) {
	conn, err := c.borrow("subtransaction")
	if err != nil {
		return err
	}
	defer func() { err = firstError(err, c.unborrow()) }()

	subtx := &Subtx{
		borrowableConn: borrowableConn{conn: conn},
		txState:        c.txstate(),
		options:        c.txOptions(),
		state:          state,
	}

	if e := subtx.declare(ctx); e != nil {
		return e
	}

	if e := action(ctx, subtx); e != nil {
		return firstError(subtx.rollback(ctx), e)
	}

	return subtx.release(ctx)
}

// SubtxBlock is work to be done in a subtransaction.
type SubtxBlock func(context.Context, *Subtx) error

// Subtx is a subtransaction.
type Subtx struct {
	borrowableConn
	*txState
	options TxOptions
	name    string
	state   map[string]interface{}
}

func (t *Subtx) declare(ctx context.Context) error {
	if e := t.assertStarted("start subtransaction"); e != nil {
		return e
	}

	t.name = t.nextSavepointName()
	cmd := "DECLARE SAVEPOINT " + t.name
	q, err := newQuery("Execute", cmd, nil, txCapabilities, nil, t.state)
	if err != nil {
		return err
	}

	return t.scriptFlow(ctx, q)
}

func (t *Subtx) release(ctx context.Context) error {
	if e := t.assertStarted("release subtransaction"); e != nil {
		return e
	}

	cmd := "RELEASE SAVEPOINT " + t.name
	q, err := newQuery("Execute", cmd, nil, txCapabilities, nil, t.state)
	if err != nil {
		return err
	}

	return t.scriptFlow(ctx, q)
}

func (t *Subtx) rollback(ctx context.Context) error {
	if e := t.assertStarted("rollback subtransaction"); e != nil {
		return e
	}

	cmd := "ROLLBACK TO SAVEPOINT " + t.name
	q, err := newQuery("Execute", cmd, nil, txCapabilities, nil, t.state)
	if err != nil {
		return err
	}

	return t.scriptFlow(ctx, q)
}

func (t *Subtx) txOptions() TxOptions { return t.options }

func (t *Subtx) txstate() *txState { return t.txState }

// Subtx runs an action in a savepoint.
// If the action returns an error the savepoint is rolled back,
// otherwise it is released.
func (t *Subtx) Subtx(ctx context.Context, action SubtxBlock) error {
	return runSubtx(ctx, action, t, t.state)
}

// Execute an EdgeQL command (or commands).
func (t *Subtx) Execute(
	ctx context.Context,
	cmd string,
	args ...interface{},
) error {
	if e := t.assertStarted("Execute"); e != nil {
		return e
	}

	q, err := newQuery("Execute", cmd, args, t.capabilities1pX(), nil, t.state)
	if err != nil {
		return err
	}

	return t.scriptFlow(ctx, q)
}

func (t *Subtx) granularFlow(ctx context.Context, q *query) error {
	if e := t.assertStarted(q.method); e != nil {
		return e
	}

	return t.borrowableConn.granularFlow(ctx, q)
}

// Query runs a query and returns the results.
func (t *Subtx) Query(
	ctx context.Context,
	cmd string,
	out interface{},
	args ...interface{},
) error {
	return runQuery(ctx, t, "Query", cmd, out, args, t.state)
}

// QuerySingle runs a singleton-returning query and returns its element.
// If the query executes successfully but doesn't return a result
// a NoDataError is returned.
func (t *Subtx) QuerySingle(
	ctx context.Context,
	cmd string,
	out interface{},
	args ...interface{},
) error {
	return runQuery(ctx, t, "QuerySingle", cmd, out, args, t.state)
}

// QueryJSON runs a query and return the results as JSON.
func (t *Subtx) QueryJSON(
	ctx context.Context,
	cmd string,
	out *[]byte,
	args ...interface{},
) error {
	return runQuery(ctx, t, "QueryJSON", cmd, out, args, t.state)
}

// QuerySingleJSON runs a singleton-returning query.
// If the query executes successfully but doesn't have a result
// a NoDataError is returned.
func (t *Subtx) QuerySingleJSON(
	ctx context.Context,
	cmd string,
	out interface{},
	args ...interface{},
) error {
	return runQuery(ctx, t, "QuerySingleJSON", cmd, out, args, t.state)
}
