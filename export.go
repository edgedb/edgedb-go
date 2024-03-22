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

// This file is auto generated. Do not edit!
// run 'go generate ./...' to regenerate

package edgedb

import (
	edgedb "github.com/edgedb/edgedb-go/internal/client"
	"github.com/edgedb/edgedb-go/internal/edgedbtypes"
)

const (
	// NetworkError indicates that the transaction was interupted
	// by a network error.
	NetworkError = edgedb.NetworkError

	// Serializable is the only isolation level
	Serializable = edgedb.Serializable

	// TLSModeDefault makes security mode inferred from other options
	TLSModeDefault = edgedb.TLSModeDefault

	// TLSModeInsecure results in no certificate verification whatsoever
	TLSModeInsecure = edgedb.TLSModeInsecure

	// TLSModeNoHostVerification enables certificate verification
	// against CAs, but hostname matching is not performed.
	TLSModeNoHostVerification = edgedb.TLSModeNoHostVerification

	// TLSModeStrict enables full certificate and hostname verification.
	TLSModeStrict = edgedb.TLSModeStrict

	// TxConflict indicates that the server could not complete a transaction
	// because it encountered a deadlock or serialization error.
	TxConflict = edgedb.TxConflict
)

type (
	// Client is a connection pool and is safe for concurrent use.
	Client = edgedb.Client

	// DateDuration represents the elapsed time between two dates in a fuzzy human
	// way.
	DateDuration = edgedbtypes.DateDuration

	// Duration represents the elapsed time between two instants
	// as an int64 microsecond count.
	Duration = edgedbtypes.Duration

	// Error is the error type returned from edgedb.
	Error = edgedb.Error

	// ErrorCategory values represent EdgeDB's error types.
	ErrorCategory = edgedb.ErrorCategory

	// ErrorTag is the argument type to Error.HasTag().
	ErrorTag = edgedb.ErrorTag

	// Executor is a common interface between Client and Tx,
	// that can run queries on an EdgeDB database.
	Executor = edgedb.Executor

	// IsolationLevel documentation can be found here
	// https://www.edgedb.com/docs/reference/edgeql/tx_start#parameters
	IsolationLevel = edgedb.IsolationLevel

	// LocalDate is a date without a time zone.
	// https://www.edgedb.com/docs/stdlib/datetime#type::cal::local_date
	LocalDate = edgedbtypes.LocalDate

	// LocalDateTime is a date and time without timezone.
	// https://www.edgedb.com/docs/stdlib/datetime#type::cal::local_datetime
	LocalDateTime = edgedbtypes.LocalDateTime

	// LocalTime is a time without a time zone.
	// https://www.edgedb.com/docs/stdlib/datetime#type::cal::local_time
	LocalTime = edgedbtypes.LocalTime

	// Memory represents memory in bytes.
	Memory = edgedbtypes.Memory

	// ModuleAlias is an alias name and module name pair.
	ModuleAlias = edgedb.ModuleAlias

	// Optional represents a shape field that is not required.
	// Optional is embedded in structs to make them optional. For example:
	//
	//	type User struct {
	//	    edgedb.Optional
	//	    Name string `edgedb:"name"`
	//	}
	Optional = edgedbtypes.Optional

	// OptionalBigInt is an optional *big.Int. Optional types must be used for out
	// parameters when a shape field is not required.
	OptionalBigInt = edgedbtypes.OptionalBigInt

	// OptionalBool is an optional bool. Optional types must be used for out
	// parameters when a shape field is not required.
	OptionalBool = edgedbtypes.OptionalBool

	// OptionalBytes is an optional []byte. Optional types must be used for out
	// parameters when a shape field is not required.
	OptionalBytes = edgedbtypes.OptionalBytes

	// OptionalDateDuration is an optional DateDuration. Optional types
	// must be used for out parameters when a shape field is not required.
	OptionalDateDuration = edgedbtypes.OptionalDateDuration

	// OptionalDateTime is an optional time.Time.  Optional types must be used for
	// out parameters when a shape field is not required.
	OptionalDateTime = edgedbtypes.OptionalDateTime

	// OptionalDuration is an optional Duration. Optional types must be used for
	// out parameters when a shape field is not required.
	OptionalDuration = edgedbtypes.OptionalDuration

	// OptionalFloat32 is an optional float32. Optional types must be used for out
	// parameters when a shape field is not required.
	OptionalFloat32 = edgedbtypes.OptionalFloat32

	// OptionalFloat64 is an optional float64. Optional types must be used for out
	// parameters when a shape field is not required.
	OptionalFloat64 = edgedbtypes.OptionalFloat64

	// OptionalInt16 is an optional int16. Optional types must be used for out
	// parameters when a shape field is not required.
	OptionalInt16 = edgedbtypes.OptionalInt16

	// OptionalInt32 is an optional int32. Optional types must be used for out
	// parameters when a shape field is not required.
	OptionalInt32 = edgedbtypes.OptionalInt32

	// OptionalInt64 is an optional int64. Optional types must be used for out
	// parameters when a shape field is not required.
	OptionalInt64 = edgedbtypes.OptionalInt64

	// OptionalLocalDate is an optional LocalDate. Optional types must be used for
	// out parameters when a shape field is not required.
	OptionalLocalDate = edgedbtypes.OptionalLocalDate

	// OptionalLocalDateTime is an optional LocalDateTime. Optional types must be
	// used for out parameters when a shape field is not required.
	OptionalLocalDateTime = edgedbtypes.OptionalLocalDateTime

	// OptionalLocalTime is an optional LocalTime. Optional types must be used for
	// out parameters when a shape field is not required.
	OptionalLocalTime = edgedbtypes.OptionalLocalTime

	// OptionalMemory is an optional Memory. Optional types must be used for
	// out parameters when a shape field is not required.
	OptionalMemory = edgedbtypes.OptionalMemory

	// OptionalRangeDateTime is an optional RangeDateTime. Optional
	// types must be used for out parameters when a shape field is not required.
	OptionalRangeDateTime = edgedbtypes.OptionalRangeDateTime

	// OptionalRangeFloat32 is an optional RangeFloat32. Optional
	// types must be used for out parameters when a shape field is not required.
	OptionalRangeFloat32 = edgedbtypes.OptionalRangeFloat32

	// OptionalRangeFloat64 is an optional RangeFloat64. Optional
	// types must be used for out parameters when a shape field is not required.
	OptionalRangeFloat64 = edgedbtypes.OptionalRangeFloat64

	// OptionalRangeInt32 is an optional RangeInt32. Optional types must be used
	// for out parameters when a shape field is not required.
	OptionalRangeInt32 = edgedbtypes.OptionalRangeInt32

	// OptionalRangeInt64 is an optional RangeInt64. Optional
	// types must be used for out parameters when a shape field is not required.
	OptionalRangeInt64 = edgedbtypes.OptionalRangeInt64

	// OptionalRangeLocalDate is an optional RangeLocalDate. Optional types must be
	// used for out parameters when a shape field is not required.
	OptionalRangeLocalDate = edgedbtypes.OptionalRangeLocalDate

	// OptionalRangeLocalDateTime is an optional RangeLocalDateTime. Optional
	// types must be used for out parameters when a shape field is not required.
	OptionalRangeLocalDateTime = edgedbtypes.OptionalRangeLocalDateTime

	// OptionalRelativeDuration is an optional RelativeDuration. Optional types
	// must be used for out parameters when a shape field is not required.
	OptionalRelativeDuration = edgedbtypes.OptionalRelativeDuration

	// OptionalStr is an optional string. Optional types must be used for out
	// parameters when a shape field is not required.
	OptionalStr = edgedbtypes.OptionalStr

	// OptionalUUID is an optional UUID. Optional types must be used for out
	// parameters when a shape field is not required.
	OptionalUUID = edgedbtypes.OptionalUUID

	// Options for connecting to an EdgeDB server
	Options = edgedb.Options

	// RangeDateTime is an interval of time.Time values.
	RangeDateTime = edgedbtypes.RangeDateTime

	// RangeFloat32 is an interval of float32 values.
	RangeFloat32 = edgedbtypes.RangeFloat32

	// RangeFloat64 is an interval of float64 values.
	RangeFloat64 = edgedbtypes.RangeFloat64

	// RangeInt32 is an interval of int32 values.
	RangeInt32 = edgedbtypes.RangeInt32

	// RangeInt64 is an interval of int64 values.
	RangeInt64 = edgedbtypes.RangeInt64

	// RangeLocalDate is an interval of LocalDate values.
	RangeLocalDate = edgedbtypes.RangeLocalDate

	// RangeLocalDateTime is an interval of LocalDateTime values.
	RangeLocalDateTime = edgedbtypes.RangeLocalDateTime

	// RelativeDuration represents the elapsed time between two instants in a fuzzy
	// human way.
	RelativeDuration = edgedbtypes.RelativeDuration

	// RetryBackoff returns the duration to wait after the nth attempt
	// before making the next attempt when retrying a transaction.
	RetryBackoff = edgedb.RetryBackoff

	// RetryCondition represents scenarios that can caused a transaction
	// run in Tx() methods to be retried.
	RetryCondition = edgedb.RetryCondition

	// RetryOptions configures how Tx() retries failed transactions.  Use
	// NewRetryOptions to get a default RetryOptions value instead of creating one
	// yourself.
	RetryOptions = edgedb.RetryOptions

	// RetryRule determines how transactions should be retried when run in Tx()
	// methods. See Client.Tx() for details.
	RetryRule = edgedb.RetryRule

	// TLSOptions contains the parameters needed to configure TLS on EdgeDB
	// server connections.
	TLSOptions = edgedb.TLSOptions

	// TLSSecurityMode specifies how strict TLS validation is.
	TLSSecurityMode = edgedb.TLSSecurityMode

	// Tx is a transaction. Use Client.Tx() to get a transaction.
	Tx = edgedb.Tx

	// TxBlock is work to be done in a transaction.
	TxBlock = edgedb.TxBlock

	// TxOptions configures how transactions behave.
	TxOptions = edgedb.TxOptions

	// UUID is a universally unique identifier
	// https://www.edgedb.com/docs/stdlib/uuid
	UUID = edgedbtypes.UUID
)

var (
	// CreateClient returns a new client. The client connects lazily. Call
	// Client.EnsureConnected() to force a connection.
	CreateClient = edgedb.CreateClient

	// CreateClientDSN returns a new client. See also CreateClient.
	//
	// dsn is either an instance name
	// https://www.edgedb.com/docs/clients/connection
	// or it specifies a single string in the following format:
	//
	//	edgedb://user:password@host:port/database?option=value.
	//
	// The following options are recognized: host, port, user, database, password.
	CreateClientDSN = edgedb.CreateClientDSN

	// DurationFromNanoseconds creates a Duration represented as microseconds
	// from a [time.Duration] represented as nanoseconds.
	DurationFromNanoseconds = edgedbtypes.DurationFromNanoseconds

	// NewDateDuration returns a new DateDuration
	NewDateDuration = edgedbtypes.NewDateDuration

	// NewLocalDate returns a new LocalDate
	NewLocalDate = edgedbtypes.NewLocalDate

	// NewLocalDateTime returns a new LocalDateTime
	NewLocalDateTime = edgedbtypes.NewLocalDateTime

	// NewLocalTime returns a new LocalTime
	NewLocalTime = edgedbtypes.NewLocalTime

	// NewOptionalBigInt is a convenience function for creating an OptionalBigInt
	// with its value set to v.
	NewOptionalBigInt = edgedbtypes.NewOptionalBigInt

	// NewOptionalBool is a convenience function for creating an OptionalBool with
	// its value set to v.
	NewOptionalBool = edgedbtypes.NewOptionalBool

	// NewOptionalBytes is a convenience function for creating an OptionalBytes
	// with its value set to v.
	NewOptionalBytes = edgedbtypes.NewOptionalBytes

	// NewOptionalDateDuration is a convenience function for creating an
	// OptionalDateDuration with its value set to v.
	NewOptionalDateDuration = edgedbtypes.NewOptionalDateDuration

	// NewOptionalDateTime is a convenience function for creating an
	// OptionalDateTime with its value set to v.
	NewOptionalDateTime = edgedbtypes.NewOptionalDateTime

	// NewOptionalDuration is a convenience function for creating an
	// OptionalDuration with its value set to v.
	NewOptionalDuration = edgedbtypes.NewOptionalDuration

	// NewOptionalFloat32 is a convenience function for creating an OptionalFloat32
	// with its value set to v.
	NewOptionalFloat32 = edgedbtypes.NewOptionalFloat32

	// NewOptionalFloat64 is a convenience function for creating an OptionalFloat64
	// with its value set to v.
	NewOptionalFloat64 = edgedbtypes.NewOptionalFloat64

	// NewOptionalInt16 is a convenience function for creating an OptionalInt16
	// with its value set to v.
	NewOptionalInt16 = edgedbtypes.NewOptionalInt16

	// NewOptionalInt32 is a convenience function for creating an OptionalInt32
	// with its value set to v.
	NewOptionalInt32 = edgedbtypes.NewOptionalInt32

	// NewOptionalInt64 is a convenience function for creating an OptionalInt64
	// with its value set to v.
	NewOptionalInt64 = edgedbtypes.NewOptionalInt64

	// NewOptionalLocalDate is a convenience function for creating an
	// OptionalLocalDate with its value set to v.
	NewOptionalLocalDate = edgedbtypes.NewOptionalLocalDate

	// NewOptionalLocalDateTime is a convenience function for creating an
	// OptionalLocalDateTime with its value set to v.
	NewOptionalLocalDateTime = edgedbtypes.NewOptionalLocalDateTime

	// NewOptionalLocalTime is a convenience function for creating an
	// OptionalLocalTime with its value set to v.
	NewOptionalLocalTime = edgedbtypes.NewOptionalLocalTime

	// NewOptionalMemory is a convenience function for creating an
	// OptionalMemory with its value set to v.
	NewOptionalMemory = edgedbtypes.NewOptionalMemory

	// NewOptionalRangeDateTime is a convenience function for creating an
	// OptionalRangeDateTime with its value set to v.
	NewOptionalRangeDateTime = edgedbtypes.NewOptionalRangeDateTime

	// NewOptionalRangeFloat32 is a convenience function for creating an
	// OptionalRangeFloat32 with its value set to v.
	NewOptionalRangeFloat32 = edgedbtypes.NewOptionalRangeFloat32

	// NewOptionalRangeFloat64 is a convenience function for creating an
	// OptionalRangeFloat64 with its value set to v.
	NewOptionalRangeFloat64 = edgedbtypes.NewOptionalRangeFloat64

	// NewOptionalRangeInt32 is a convenience function for creating an
	// OptionalRangeInt32 with its value set to v.
	NewOptionalRangeInt32 = edgedbtypes.NewOptionalRangeInt32

	// NewOptionalRangeInt64 is a convenience function for creating an
	// OptionalRangeInt64 with its value set to v.
	NewOptionalRangeInt64 = edgedbtypes.NewOptionalRangeInt64

	// NewOptionalRangeLocalDate is a convenience function for creating an
	// OptionalRangeLocalDate with its value set to v.
	NewOptionalRangeLocalDate = edgedbtypes.NewOptionalRangeLocalDate

	// NewOptionalRangeLocalDateTime is a convenience function for creating an
	// OptionalRangeLocalDateTime with its value set to v.
	NewOptionalRangeLocalDateTime = edgedbtypes.NewOptionalRangeLocalDateTime

	// NewOptionalRelativeDuration is a convenience function for creating an
	// OptionalRelativeDuration with its value set to v.
	NewOptionalRelativeDuration = edgedbtypes.NewOptionalRelativeDuration

	// NewOptionalStr is a convenience function for creating an OptionalStr with
	// its value set to v.
	NewOptionalStr = edgedbtypes.NewOptionalStr

	// NewOptionalUUID is a convenience function for creating an OptionalUUID with
	// its value set to v.
	NewOptionalUUID = edgedbtypes.NewOptionalUUID

	// NewRangeDateTime creates a new RangeDateTime value.
	NewRangeDateTime = edgedbtypes.NewRangeDateTime

	// NewRangeFloat32 creates a new RangeFloat32 value.
	NewRangeFloat32 = edgedbtypes.NewRangeFloat32

	// NewRangeFloat64 creates a new RangeFloat64 value.
	NewRangeFloat64 = edgedbtypes.NewRangeFloat64

	// NewRangeInt32 creates a new RangeInt32 value.
	NewRangeInt32 = edgedbtypes.NewRangeInt32

	// NewRangeInt64 creates a new RangeInt64 value.
	NewRangeInt64 = edgedbtypes.NewRangeInt64

	// NewRangeLocalDate creates a new RangeLocalDate value.
	NewRangeLocalDate = edgedbtypes.NewRangeLocalDate

	// NewRangeLocalDateTime creates a new RangeLocalDateTime value.
	NewRangeLocalDateTime = edgedbtypes.NewRangeLocalDateTime

	// NewRelativeDuration returns a new RelativeDuration
	NewRelativeDuration = edgedbtypes.NewRelativeDuration

	// NewRetryOptions returns the default RetryOptions value.
	NewRetryOptions = edgedb.NewRetryOptions

	// NewRetryRule returns the default RetryRule value.
	NewRetryRule = edgedb.NewRetryRule

	// NewTxOptions returns the default TxOptions value.
	NewTxOptions = edgedb.NewTxOptions

	// ParseUUID parses s into a UUID or returns an error.
	ParseUUID = edgedbtypes.ParseUUID
)
