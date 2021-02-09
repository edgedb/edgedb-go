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

type isolationLevel string

const (
	serializable   isolationLevel = "serializable"
	repeatableRead isolationLevel = "repeatable_read"
)

// Tx is a transaction.
type Tx interface {
	Executor
}

type transaction struct {
	conn       *baseConn
	state      transactionState
	isolation  isolationLevel
	readOnly   bool
	deferrable bool
}

func (t *transaction) execute(
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
func (t *transaction) assertNotDone(opName string) error {
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
func (t *transaction) assertStarted(opName string) error {
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

func (t *transaction) start(ctx context.Context) error {
	if e := t.assertNotDone("start"); e != nil {
		return e
	}

	if t.state == startedTx {
		return &interfaceError{
			msg: "cannot start; the transaction is already started",
		}
	}

	query := "START TRANSACTION"

	switch t.isolation {
	case repeatableRead:
		query += " ISOLATION REPEATABLE READ"
	case serializable:
		query += " ISOLATION SERIALIZABLE"
	default:
		return &configurationError{
			msg: fmt.Sprintf("unknown isolation level: %q", t.isolation),
		}
	}

	if t.readOnly {
		query += ", READ ONLY"
	} else {
		query += ", READ WRITE"
	}

	if t.deferrable {
		query += ", DEFERRABLE"
	} else {
		query += ", NOT DEFERRABLE"
	}

	query += ";"

	return t.execute(ctx, query, startedTx)
}

func (t *transaction) commit(ctx context.Context) error {
	if e := t.assertStarted("commit"); e != nil {
		return e
	}

	return t.execute(ctx, "COMMIT;", committedTx)
}

func (t *transaction) rollback(ctx context.Context) error {
	if e := t.assertStarted("rollback"); e != nil {
		return e
	}

	return t.execute(ctx, "ROLLBACK;", rolledBackTx)
}

func (t *transaction) Execute(ctx context.Context, cmd string) error {
	if e := t.assertStarted("Execute"); e != nil {
		return e
	}

	return t.conn.ScriptFlow(ctx, sfQuery{cmd: cmd})
}

func (t *transaction) Query(
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

func (t *transaction) QueryOne(
	ctx context.Context,
	cmd string,
	out interface{},
	args ...interface{},
) error {
	if e := t.assertStarted("QueryOne"); e != nil {
		return e
	}

	q, err := newQuery(cmd, format.Binary, cardinality.One, args, nil, out)
	if err != nil {
		return err
	}

	return t.conn.GranularFlow(ctx, q)
}

func (t *transaction) QueryJSON(
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

func (t *transaction) QueryOneJSON(
	ctx context.Context,
	cmd string,
	out *[]byte,
	args ...interface{},
) error {
	if e := t.assertStarted("QueryJSON"); e != nil {
		return e
	}

	q, err := newQuery(cmd, format.JSON, cardinality.One, args, nil, out)
	if err != nil {
		return err
	}

	return t.conn.GranularFlow(ctx, q)
}
