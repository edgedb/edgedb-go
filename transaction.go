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
)

// TxBlock is work to be done in a transaction.
type TxBlock func(context.Context, *Tx) error

type txStatus int

const (
	newTx txStatus = iota
	startedTx
	committedTx
	rolledBackTx
	failedTx
)

type txState struct {
	txStatus         txStatus
	txSavepointCount int
}

// assertNotDone returns an error if the transaction is in a done state.
func (s *txState) assertNotDone(opName string) error {
	switch s.txStatus {
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
func (s *txState) assertStarted(opName string) error {
	switch s.txStatus {
	case startedTx:
		return nil
	case newTx:
		return &interfaceError{msg: fmt.Sprintf(
			"cannot %v; the transaction is not yet started", opName,
		)}
	default:
		return s.assertNotDone(opName)
	}
}

func (s *txState) nextSavepointName() string {
	s.txSavepointCount++
	return fmt.Sprintf("EdgeDBGoSavepoint%v", s.txSavepointCount)
}

// Tx is a transaction. Use RetryingTx() or RawTx() to get a transaction.
type Tx struct {
	borrowableConn
	*txState
	options TxOptions
}

func (t *Tx) execute(
	ctx context.Context,
	cmd string,
	sucessState txStatus,
) error {
	err := t.borrowableConn.scriptFlow(ctx, sfQuery{cmd: cmd})

	switch err {
	case nil:
		t.txStatus = sucessState
	default:
		t.txStatus = failedTx
	}

	return err
}

func (t *Tx) start(ctx context.Context) error {
	if e := t.assertNotDone("start"); e != nil {
		return e
	}

	if t.txStatus == startedTx {
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

func (t *Tx) txOptions() TxOptions { return t.options }

func (t *Tx) txstate() *txState { return t.txState }

// Subtx runs an action in a savepoint.
// If the action returns an error the savepoint is rolled back,
// otherwise it is released.
func (t *Tx) Subtx(ctx context.Context, action SubtxBlock) error {
	return runSubtx(ctx, action, t)
}

func (t *Tx) scriptFlow(ctx context.Context, q sfQuery) error {
	if e := t.assertStarted("Execute"); e != nil {
		return e
	}

	return t.borrowableConn.scriptFlow(ctx, q)
}

func (t *Tx) granularFlow(ctx context.Context, q *gfQuery) error {
	if e := t.assertStarted(q.method); e != nil {
		return e
	}

	return t.borrowableConn.granularFlow(ctx, q)
}

// Execute an EdgeQL command (or commands).
func (t *Tx) Execute(ctx context.Context, cmd string) error {
	return t.scriptFlow(ctx, sfQuery{
		cmd:     cmd,
		headers: t.headers(),
	})
}

// Query runs a query and returns the results.
func (t *Tx) Query(
	ctx context.Context,
	cmd string,
	out interface{},
	args ...interface{},
) error {
	return runQuery(ctx, t, "Query", cmd, out, args)
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
	return runQuery(ctx, t, "QuerySingle", cmd, out, args)
}

// QueryJSON runs a query and return the results as JSON.
func (t *Tx) QueryJSON(
	ctx context.Context,
	cmd string,
	out *[]byte,
	args ...interface{},
) error {
	return runQuery(ctx, t, "QueryJSON", cmd, out, args)
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
	return runQuery(ctx, t, "QuerySingleJSON", cmd, out, args)
}
