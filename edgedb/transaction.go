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

// Transaction represents a transaction or save point block.
// Transactions are created by calling the Conn.Transaction() method.
// Most callers should use `Conn.RunInTransaction()` instead.
type Transaction struct {
	conn *Conn
}

// Start a transaction or save point.
func (tx Transaction) Start() error {
	// todo handle nested blocks and other options.
	return tx.conn.Execute("START TRANSACTION;")
}

// Commit the transaction or save point preserving changes.
func (tx Transaction) Commit() error {
	// todo handle nested blocks etc.
	return tx.conn.Execute("COMMIT;")
}

// RollBack the transaction or save point block discarding changes.
func (tx Transaction) RollBack() error {
	// todo handle nested blocks etc.
	return tx.conn.Execute("ROLLBACK;")
}

// Transaction creates a new trasaction struct.
func (conn *Conn) Transaction() (Transaction, error) {
	// todo support transaction options
	return Transaction{conn}, nil
}

// RunInTransaction runs a function in a transaction.
// If function returns an error transaction is rolled back,
// otherwise transaction is committed.
func (conn *Conn) RunInTransaction(fn func() error) error {
	// see https://pkg.go.dev/github.com/go-pg/pg/v10#DB.RunInTransaction
	panic("RunInTransaction() not implemented") // todo
}
