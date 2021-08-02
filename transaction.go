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

	"github.com/edgedb/edgedb-go/internal/cardinality"
	"github.com/edgedb/edgedb-go/internal/format"
)

type transactionState int

const (
	newTx transactionState = iota
	startedTx
	committedTx
	rolledBackTx
	failedTx
)

// Tx is a transaction. Use RetryingTx() or RawTx() to get a transaction.
type Tx struct {
	conn    *baseConn
	state   transactionState
	options TxOptions
}

func (t *Tx) execute(
	ctx context.Context,
	cmd string,
	sucessState transactionState,
) error {
	err := t.conn.ScriptFlow(ctx, sfQuery{cmd: cmd})

	switch err {
	case nil:
		t.state = sucessState
	default:
		t.state = failedTx
	}

	return err
}

// assertNotDone returns an error if the transaction is in a done state.
func (t *Tx) assertNotDone(opName string) error {
	switch t.state {
	case committedTx:
		return &interfaceError{msg: fmt.Sprintf(
			"cannot %v; the transaction is already committed", opName,
		)}
	case rolledBackTx:
		return &interfaceError{msg: fmt.Sprintf(
			"cannot %v; the transaction is already rolled back", opName,
		)}
	case failedTx:
		return &interfaceError{msg: fmt.Sprintf(
			"cannot %v; the transaction is in error state", opName,
		)}
	default:
		return nil
	}
}

// assertStarted returns an error if the transaction is not in Started state.
func (t *Tx) assertStarted(opName string) error {
	switch t.state {
	case startedTx:
		return nil
	case newTx:
		return &interfaceError{msg: fmt.Sprintf(
			"cannot %v; the transaction is not yet started", opName,
		)}
	default:
		return t.assertNotDone(opName)
	}
}

func (t *Tx) start(ctx context.Context) error {
	if e := t.assertNotDone("start"); e != nil {
		return e
	}

	if t.state == startedTx {
		return &interfaceError{
			msg: "cannot start; the transaction is already started",
		}
	}

	query := t.options.startTxQuery()
	return t.execute(ctx, query, startedTx)
}

func (t *Tx) commit(ctx context.Context) error {
	if e := t.assertStarted("commit"); e != nil {
		return e
	}

	return t.execute(ctx, "COMMIT;", committedTx)
}

func (t *Tx) rollback(ctx context.Context) error {
	if e := t.assertStarted("rollback"); e != nil {
		return e
	}

	return t.execute(ctx, "ROLLBACK;", rolledBackTx)
}

// Execute an EdgeQL command (or commands).
func (t *Tx) Execute(ctx context.Context, cmd string) error {
	if e := t.assertStarted("Execute"); e != nil {
		return e
	}

	return t.conn.ScriptFlow(ctx, sfQuery{cmd: cmd})
}

// Query runs a query and returns the results.
func (t *Tx) Query(
	ctx context.Context,
	cmd string,
	out interface{},
	args ...interface{},
) error {
	if e := t.assertStarted("Query"); e != nil {
		return e
	}

	q, err := newQuery(cmd, format.Binary, cardinality.Many, args, nil, out)
	if err != nil {
		return err
	}

	return t.conn.GranularFlow(ctx, q)
}

// QueryOne runs a singleton-returning query and returns its element.
// If the query executes successfully but doesn't return a result
// a NoDataError is returned.
//
// Deprecated: use QuerySingle()
func (t *Tx) QueryOne(
	ctx context.Context,
	cmd string,
	out interface{},
	args ...interface{},
) error {
	return t.QuerySingle(ctx, cmd, out, args...)
}

// QuerySingle runs a singleton-returning query and returns its element.
// If the query executes successfully but doesn't return a result
// a NoDataError is returned.
func (t *Tx) QuerySingle(
	ctx context.Context,
	cmd string,
	out interface{},
	args ...interface{},
) error {
	if e := t.assertStarted("QuerySingle"); e != nil {
		return e
	}

	q, err := newQuery(cmd, format.Binary, cardinality.Single, args, nil, out)
	if err != nil {
		return err
	}

	return t.conn.GranularFlow(ctx, q)
}

// QueryJSON runs a query and return the results as JSON.
func (t *Tx) QueryJSON(
	ctx context.Context,
	cmd string,
	out *[]byte,
	args ...interface{},
) error {
	if e := t.assertStarted("QueryJSON"); e != nil {
		return e
	}

	q, err := newQuery(cmd, format.JSON, cardinality.Many, args, nil, out)
	if err != nil {
		return err
	}

	return t.conn.GranularFlow(ctx, q)
}

// QueryOneJSON runs a singleton-returning query.
// If the query executes successfully but doesn't have a result
// a NoDataError is returned.
//
// Deprecated: use QuerySingleJSON()
func (t *Tx) QueryOneJSON(
	ctx context.Context,
	cmd string,
	out *[]byte,
	args ...interface{},
) error {
	return t.QuerySingleJSON(ctx, cmd, out, args...)
}

// QuerySingleJSON runs a singleton-returning query.
// If the query executes successfully but doesn't have a result
// a NoDataError is returned.
func (t *Tx) QuerySingleJSON(
	ctx context.Context,
	cmd string,
	out *[]byte,
	args ...interface{},
) error {
	if e := t.assertStarted("QueryJSON"); e != nil {
		return e
	}

	q, err := newQuery(cmd, format.JSON, cardinality.Single, args, nil, out)
	if err != nil {
		return err
	}

	return t.conn.GranularFlow(ctx, q)
}
