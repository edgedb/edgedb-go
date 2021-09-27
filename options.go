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
	"errors"
	"fmt"
	"math"
	"time"
)

// Options for connecting to an EdgeDB server
type Options struct {
	// Hosts is a slice of database host addresses as one of the following
	//
	// - an IP address or domain name
	//
	// - an absolute path to the directory
	//   containing the database server Unix-domain socket
	//   (not supported on Windows)
	//
	// If the slice is empty, the following will be tried, in order:
	//
	// - host address(es) parsed from the dsn argument
	//
	// - the value of the EDGEDB_HOST environment variable
	//
	// - on Unix, common directories used for EdgeDB Unix-domain sockets:
	//   "/run/edgedb" and "/var/run/edgedb"
	//
	// - "localhost"
	Hosts []string

	// Ports is a slice of port numbers to connect to at the server host
	// (or Unix-domain socket file extension).
	//
	// Ports may either be:
	//
	// - the same length ans Hosts
	//
	// - a single port to be used all specified hosts
	//
	// - empty indicating the value parsed from the dsn argument
	//   should be used, or the value of the EDGEDB_PORT environment variable,
	//   or 5656 if neither is specified.
	Ports []int

	// User is the name of the database role used for authentication.
	// If not specified, the value parsed from the dsn argument is used,
	// or the value of the EDGEDB_USER environment variable,
	// or the operating system name of the user running the application.
	User string

	// Database is the name of the database to connect to.
	// If not specified, the value parsed from the dsn argument is used,
	// or the value of the EDGEDB_DATABASE environment variable,
	// or the operating system name of the user running the application.
	Database string

	// Password to be used for authentication,
	// if the server requires one. If not specified,
	// the value parsed from the dsn argument is used,
	// or the value of the EDGEDB_PASSWORD environment variable.
	// Note that the use of the environment variable is discouraged
	// as other users and applications may be able to read it
	// without needing specific privileges.
	Password string

	// ConnectTimeout is used when establishing connections in the background.
	ConnectTimeout time.Duration

	// WaitUntilAvailable determines how long to wait
	// to reestablish a connection.
	WaitUntilAvailable time.Duration

	// Concurrency determines the maximum number of connections.
	// If Concurrency is zero, max(4, runtime.NumCPU()) will be used.
	// Has no effect for single connections.
	Concurrency uint

	// Read the TLS certificate from this file
	TLSCAFile string

	// If false don't verify the server's hostname when using TLS.
	TLSVerifyHostname OptionalBool

	// ServerSettings is currently unused.
	ServerSettings map[string]string
}

// RetryBackoff returns the duration to wait after the nth attempt
// before making the next attempt when retrying a transaction.
type RetryBackoff func(n int) time.Duration

func defaultBackoff(attempt int) time.Duration {
	backoff := math.Pow(2.0, float64(attempt)) * 100.0
	jitter := rnd.Float64() * 100.0
	return time.Duration(backoff+jitter) * time.Millisecond
}

// RetryCondition represents scenarios that can caused a transaction
// run in RetryingTx() methods to be retried.
type RetryCondition int

// The following conditions can be configured with a custom RetryRule.
// See RetryOptions.
const (
	// TxConflict indicates that the server could not complete a transaction
	// because it encountered a deadlock or serialization error.
	TxConflict = iota

	// NetworkError indicates that the transaction was interupted
	// by a network error.
	NetworkError
)

// NewRetryRule returns the default RetryRule value.
func NewRetryRule() RetryRule {
	return RetryRule{
		fromFactory: true,
		attempts:    3,
		backoff:     defaultBackoff,
	}
}

// RetryRule determines how transactions should be retried
// when run in RetryingTx() methods. See Client.RetryingTx() for details.
type RetryRule struct {
	// fromFactory indicates that a RetryOptions value was created using
	// NewRetryOptions() and not created directly. Requiring users to use the
	// factory function allows for nonzero default values.
	fromFactory bool

	// Total number of times to attempt a transaction.
	// attempts <= 0 indicate that a default value should be used.
	attempts int

	// backoff determines how long to wait between transaction attempts.
	// nil indicates that a default function should be used.
	backoff RetryBackoff
}

// WithAttempts sets the rule's attempts. attempts must be greater than zero.
func (r RetryRule) WithAttempts(attempts int) RetryRule {
	if attempts < 1 {
		panic(fmt.Sprintf(
			"RetryRule attempts must be greater than 0, got %v",
			attempts,
		))
	}

	r.attempts = attempts
	return r
}

// WithBackoff returns a copy of the RetryRule with backoff set to fn.
func (r RetryRule) WithBackoff(fn RetryBackoff) RetryRule {
	if fn == nil {
		panic("the backoff function must not be nil")
	}

	r.backoff = fn
	return r
}

// RetryOptions configures how RetryingTx() retries failed transactions.
// Use NewRetryOptions to get a default RetryOptions value
// instead of creating one yourself.
type RetryOptions struct {
	fromFactory bool
	txConflict  RetryRule
	network     RetryRule
}

// WithDefault sets the rule for all conditions to rule.
func (o RetryOptions) WithDefault(rule RetryRule) RetryOptions { // nolint:gocritic,lll
	if !rule.fromFactory {
		panic("RetryRule not created with NewRetryRule() is not valid")
	}

	o.txConflict = rule
	o.network = rule
	return o
}

// WithCondition sets the retry rule for the specified condition.
func (o RetryOptions) WithCondition( // nolint:gocritic
	condition RetryCondition,
	rule RetryRule,
) RetryOptions {
	if !rule.fromFactory {
		panic("RetryRule not created with NewRetryRule() is not valid")
	}

	switch condition {
	case TxConflict:
		o.txConflict = rule
	case NetworkError:
		o.network = rule
	default:
		panic(fmt.Sprintf("unexpected condition: %v", condition))
	}

	return o
}

func (o RetryOptions) ruleForException(err Error) RetryRule { // nolint:gocritic,lll
	var edbErr Error
	if !errors.As(err, &edbErr) {
		panic(fmt.Sprintf("unexpected error type: %T", err))
	}

	switch {
	case edbErr.Category(TransactionConflictError):
		return o.txConflict
	case edbErr.Category(ClientError):
		return o.network
	default:
		panic(fmt.Sprintf("unexpected error type: %T", err))
	}
}

// IsolationLevel documentation can be found here
// https://www.edgedb.com/docs/edgeql/statements/tx_start#parameters
type IsolationLevel string

// The available levels are:
const (
	Serializable   IsolationLevel = "serializable"
	RepeatableRead IsolationLevel = "repeatable_read"
)

// NewTxOptions returns the default TxOptions value.
func NewTxOptions() TxOptions {
	return TxOptions{
		fromFactory: true,
		isolation:   RepeatableRead,
	}
}

// TxOptions configures how transactions behave.
type TxOptions struct {
	// fromFactory indicates that a TxOptions value was created using
	// NewTxOptions() and not created directly with TxOptions{}.
	// Requiring users to use the factory function allows for nonzero
	// default values.
	fromFactory bool

	readOnly   bool
	deferrable bool
	isolation  IsolationLevel
}

// WithIsolation returns a copy of the TxOptions
// with the isolation level set to i.
func (o TxOptions) WithIsolation(i IsolationLevel) TxOptions {
	if i != Serializable && i != RepeatableRead {
		panic(fmt.Sprintf("unknown isolation level: %q", i))
	}

	o.isolation = i
	return o
}

// WithReadOnly returns a shallow copy of the client
// with the transaction read only access mode set to r.
func (o TxOptions) WithReadOnly(r bool) TxOptions {
	o.readOnly = r
	return o
}

// WithDeferrable returns a shallow copy of the client
// with the transaction deferrable mode set to d.
func (o TxOptions) WithDeferrable(d bool) TxOptions {
	o.deferrable = d
	return o
}

func (o TxOptions) startTxQuery() string { // nolint:gocritic
	query := "START TRANSACTION"

	switch o.isolation {
	case RepeatableRead:
		query += " ISOLATION REPEATABLE READ"
	case Serializable:
		query += " ISOLATION SERIALIZABLE"
	default:
		panic(fmt.Sprintf("unknown isolation level: %q", o.isolation))
	}

	if o.readOnly {
		query += ", READ ONLY"
	} else {
		query += ", READ WRITE"
	}

	if o.deferrable {
		query += ", DEFERRABLE"
	} else {
		query += ", NOT DEFERRABLE"
	}

	query += ";"
	return query
}

// WithTxOptions returns a shallow copy of the client
// with the TxOptions set to opts.
func (p Client) WithTxOptions(opts TxOptions) *Client { // nolint:gocritic
	if !opts.fromFactory {
		panic("TxOptions not created with NewTxOptions() are not valid")
	}

	p.txOpts = opts
	return &p
}

// WithRetryOptions returns a shallow copy of the client
// with the RetryOptions set to opts.
func (p Client) WithRetryOptions( // nolint:gocritic
	opts RetryOptions,
) *Client {
	if !opts.fromFactory {
		panic("RetryOptions not created with NewRetryOptions() are not valid")
	}

	p.retryOpts = opts
	return &p
}
