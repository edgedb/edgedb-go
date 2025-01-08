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

package gel

import (
	gel "github.com/edgedb/edgedb-go/internal/client"
	"github.com/edgedb/edgedb-go/internal/geltypes"
)

const (
	// NetworkError indicates that the transaction was interupted
	// by a network error.
	NetworkError = gel.NetworkError

	// Serializable is the only isolation level
	Serializable = gel.Serializable

	// TLSModeDefault makes security mode inferred from other options
	TLSModeDefault = gel.TLSModeDefault

	// TLSModeInsecure results in no certificate verification whatsoever
	TLSModeInsecure = gel.TLSModeInsecure

	// TLSModeNoHostVerification enables certificate verification
	// against CAs, but hostname matching is not performed.
	TLSModeNoHostVerification = gel.TLSModeNoHostVerification

	// TLSModeStrict enables full certificate and hostname verification.
	TLSModeStrict = gel.TLSModeStrict

	// TxConflict indicates that the server could not complete a transaction
	// because it encountered a deadlock or serialization error.
	TxConflict = gel.TxConflict
)

type (
	// Client is a connection pool and is safe for concurrent use.
	Client = gel.Client

	// DateDuration represents the elapsed time between two dates in a fuzzy human
	// way.
	DateDuration = geltypes.DateDuration

	// Duration represents the elapsed time between two instants
	// as an int64 microsecond count.
	Duration = geltypes.Duration

	// Error is the error type returned from gel.
	Error = gel.Error

	// ErrorCategory values represent Gel's error types.
	ErrorCategory = gel.ErrorCategory

	// ErrorTag is the argument type to Error.HasTag().
	ErrorTag = gel.ErrorTag

	// Executor is a common interface between *Client and *Tx,
	// that can run queries on an Gel database.
	Executor = gel.Executor

	// IsolationLevel documentation can be found here
	// https://www.edgedb.com/docs/reference/edgeql/tx_start#parameters
	IsolationLevel = gel.IsolationLevel

	// LocalDate is a date without a time zone.
	// https://www.edgedb.com/docs/stdlib/datetime#type::cal::local_date
	LocalDate = geltypes.LocalDate

	// LocalDateTime is a date and time without timezone.
	// https://www.edgedb.com/docs/stdlib/datetime#type::cal::local_datetime
	LocalDateTime = geltypes.LocalDateTime

	// LocalTime is a time without a time zone.
	// https://www.edgedb.com/docs/stdlib/datetime#type::cal::local_time
	LocalTime = geltypes.LocalTime

	// Memory represents memory in bytes.
	Memory = geltypes.Memory

	// ModuleAlias is an alias name and module name pair.
	ModuleAlias = gel.ModuleAlias

	// Optional represents a shape field that is not required.
	// Optional is embedded in structs to make them optional. For example:
	//
	//	type User struct {
	//	    gel.Optional
	//	    Name string `gel:"name"`
	//	}
	Optional = geltypes.Optional

	// OptionalBigInt is an optional *big.Int. Optional types must be used for out
	// parameters when a shape field is not required.
	OptionalBigInt = geltypes.OptionalBigInt

	// OptionalBool is an optional bool. Optional types must be used for out
	// parameters when a shape field is not required.
	OptionalBool = geltypes.OptionalBool

	// OptionalBytes is an optional []byte. Optional types must be used for out
	// parameters when a shape field is not required.
	OptionalBytes = geltypes.OptionalBytes

	// OptionalDateDuration is an optional DateDuration. Optional types
	// must be used for out parameters when a shape field is not required.
	OptionalDateDuration = geltypes.OptionalDateDuration

	// OptionalDateTime is an optional time.Time.  Optional types must be used for
	// out parameters when a shape field is not required.
	OptionalDateTime = geltypes.OptionalDateTime

	// OptionalDuration is an optional Duration. Optional types must be used for
	// out parameters when a shape field is not required.
	OptionalDuration = geltypes.OptionalDuration

	// OptionalFloat32 is an optional float32. Optional types must be used for out
	// parameters when a shape field is not required.
	OptionalFloat32 = geltypes.OptionalFloat32

	// OptionalFloat64 is an optional float64. Optional types must be used for out
	// parameters when a shape field is not required.
	OptionalFloat64 = geltypes.OptionalFloat64

	// OptionalInt16 is an optional int16. Optional types must be used for out
	// parameters when a shape field is not required.
	OptionalInt16 = geltypes.OptionalInt16

	// OptionalInt32 is an optional int32. Optional types must be used for out
	// parameters when a shape field is not required.
	OptionalInt32 = geltypes.OptionalInt32

	// OptionalInt64 is an optional int64. Optional types must be used for out
	// parameters when a shape field is not required.
	OptionalInt64 = geltypes.OptionalInt64

	// OptionalLocalDate is an optional LocalDate. Optional types must be used for
	// out parameters when a shape field is not required.
	OptionalLocalDate = geltypes.OptionalLocalDate

	// OptionalLocalDateTime is an optional LocalDateTime. Optional types must be
	// used for out parameters when a shape field is not required.
	OptionalLocalDateTime = geltypes.OptionalLocalDateTime

	// OptionalLocalTime is an optional LocalTime. Optional types must be used for
	// out parameters when a shape field is not required.
	OptionalLocalTime = geltypes.OptionalLocalTime

	// OptionalMemory is an optional Memory. Optional types must be used for
	// out parameters when a shape field is not required.
	OptionalMemory = geltypes.OptionalMemory

	// OptionalRangeDateTime is an optional RangeDateTime. Optional
	// types must be used for out parameters when a shape field is not required.
	OptionalRangeDateTime = geltypes.OptionalRangeDateTime

	// OptionalRangeFloat32 is an optional RangeFloat32. Optional
	// types must be used for out parameters when a shape field is not required.
	OptionalRangeFloat32 = geltypes.OptionalRangeFloat32

	// OptionalRangeFloat64 is an optional RangeFloat64. Optional
	// types must be used for out parameters when a shape field is not required.
	OptionalRangeFloat64 = geltypes.OptionalRangeFloat64

	// OptionalRangeInt32 is an optional RangeInt32. Optional types must be used
	// for out parameters when a shape field is not required.
	OptionalRangeInt32 = geltypes.OptionalRangeInt32

	// OptionalRangeInt64 is an optional RangeInt64. Optional
	// types must be used for out parameters when a shape field is not required.
	OptionalRangeInt64 = geltypes.OptionalRangeInt64

	// OptionalRangeLocalDate is an optional RangeLocalDate. Optional types must be
	// used for out parameters when a shape field is not required.
	OptionalRangeLocalDate = geltypes.OptionalRangeLocalDate

	// OptionalRangeLocalDateTime is an optional RangeLocalDateTime. Optional
	// types must be used for out parameters when a shape field is not required.
	OptionalRangeLocalDateTime = geltypes.OptionalRangeLocalDateTime

	// OptionalRelativeDuration is an optional RelativeDuration. Optional types
	// must be used for out parameters when a shape field is not required.
	OptionalRelativeDuration = geltypes.OptionalRelativeDuration

	// OptionalStr is an optional string. Optional types must be used for out
	// parameters when a shape field is not required.
	OptionalStr = geltypes.OptionalStr

	// OptionalUUID is an optional UUID. Optional types must be used for out
	// parameters when a shape field is not required.
	OptionalUUID = geltypes.OptionalUUID

	// Options for connecting to a Gel server
	Options = gel.Options

	// RangeDateTime is an interval of time.Time values.
	RangeDateTime = geltypes.RangeDateTime

	// RangeFloat32 is an interval of float32 values.
	RangeFloat32 = geltypes.RangeFloat32

	// RangeFloat64 is an interval of float64 values.
	RangeFloat64 = geltypes.RangeFloat64

	// RangeInt32 is an interval of int32 values.
	RangeInt32 = geltypes.RangeInt32

	// RangeInt64 is an interval of int64 values.
	RangeInt64 = geltypes.RangeInt64

	// RangeLocalDate is an interval of LocalDate values.
	RangeLocalDate = geltypes.RangeLocalDate

	// RangeLocalDateTime is an interval of LocalDateTime values.
	RangeLocalDateTime = geltypes.RangeLocalDateTime

	// RelativeDuration represents the elapsed time between two instants in a fuzzy
	// human way.
	RelativeDuration = geltypes.RelativeDuration

	// RetryBackoff returns the duration to wait after the nth attempt
	// before making the next attempt when retrying a transaction.
	RetryBackoff = gel.RetryBackoff

	// RetryCondition represents scenarios that can cause a transaction
	// run in Tx() methods to be retried.
	RetryCondition = gel.RetryCondition

	// RetryOptions configures how Tx() retries failed transactions.  Use
	// NewRetryOptions to get a default RetryOptions value instead of creating one
	// yourself.
	RetryOptions = gel.RetryOptions

	// RetryRule determines how transactions should be retried when run in Tx()
	// methods. See Client.Tx() for details.
	RetryRule = gel.RetryRule

	// TLSOptions contains the parameters needed to configure TLS on Gel
	// server connections.
	TLSOptions = gel.TLSOptions

	// TLSSecurityMode specifies how strict TLS validation is.
	TLSSecurityMode = gel.TLSSecurityMode

	// Tx is a transaction. Use Client.Tx() to get a transaction.
	Tx = gel.Tx

	// TxBlock is work to be done in a transaction.
	TxBlock = gel.TxBlock

	// TxOptions configures how transactions behave.
	TxOptions = gel.TxOptions

	// UUID is a universally unique identifier
	// https://www.edgedb.com/docs/stdlib/uuid
	UUID = geltypes.UUID

	// WarningHandler takes a slice of gel.Error that represent warnings and
	// optionally returns an error. This can be used to log warnings, increment
	// metrics, promote warnings to errors by returning them etc.
	WarningHandler = gel.WarningHandler
)

var (
	// CreateClient returns a new client. The client connects lazily. Call
	// Client.EnsureConnected() to force a connection.
	CreateClient = gel.CreateClient

	// CreateClientDSN returns a new client. See also CreateClient.
	//
	// dsn is either an instance name
	// https://www.edgedb.com/docs/clients/connection
	// or it specifies a single string in the following format:
	//
	//	gel://user:password@host:port/database?option=value.
	//
	// The following options are recognized: host, port, user, database, password.
	CreateClientDSN = gel.CreateClientDSN

	// DurationFromNanoseconds creates a Duration represented as microseconds
	// from a [time.Duration] represented as nanoseconds.
	DurationFromNanoseconds = geltypes.DurationFromNanoseconds

	// LogWarnings is an gel.WarningHandler that logs warnings.
	LogWarnings = gel.LogWarnings

	// NewDateDuration returns a new DateDuration
	NewDateDuration = geltypes.NewDateDuration

	// NewLocalDate returns a new LocalDate
	NewLocalDate = geltypes.NewLocalDate

	// NewLocalDateTime returns a new LocalDateTime
	NewLocalDateTime = geltypes.NewLocalDateTime

	// NewLocalTime returns a new LocalTime
	NewLocalTime = geltypes.NewLocalTime

	// NewOptionalBigInt is a convenience function for creating an OptionalBigInt
	// with its value set to v.
	NewOptionalBigInt = geltypes.NewOptionalBigInt

	// NewOptionalBool is a convenience function for creating an OptionalBool with
	// its value set to v.
	NewOptionalBool = geltypes.NewOptionalBool

	// NewOptionalBytes is a convenience function for creating an OptionalBytes
	// with its value set to v.
	NewOptionalBytes = geltypes.NewOptionalBytes

	// NewOptionalDateDuration is a convenience function for creating an
	// OptionalDateDuration with its value set to v.
	NewOptionalDateDuration = geltypes.NewOptionalDateDuration

	// NewOptionalDateTime is a convenience function for creating an
	// OptionalDateTime with its value set to v.
	NewOptionalDateTime = geltypes.NewOptionalDateTime

	// NewOptionalDuration is a convenience function for creating an
	// OptionalDuration with its value set to v.
	NewOptionalDuration = geltypes.NewOptionalDuration

	// NewOptionalFloat32 is a convenience function for creating an OptionalFloat32
	// with its value set to v.
	NewOptionalFloat32 = geltypes.NewOptionalFloat32

	// NewOptionalFloat64 is a convenience function for creating an OptionalFloat64
	// with its value set to v.
	NewOptionalFloat64 = geltypes.NewOptionalFloat64

	// NewOptionalInt16 is a convenience function for creating an OptionalInt16
	// with its value set to v.
	NewOptionalInt16 = geltypes.NewOptionalInt16

	// NewOptionalInt32 is a convenience function for creating an OptionalInt32
	// with its value set to v.
	NewOptionalInt32 = geltypes.NewOptionalInt32

	// NewOptionalInt64 is a convenience function for creating an OptionalInt64
	// with its value set to v.
	NewOptionalInt64 = geltypes.NewOptionalInt64

	// NewOptionalLocalDate is a convenience function for creating an
	// OptionalLocalDate with its value set to v.
	NewOptionalLocalDate = geltypes.NewOptionalLocalDate

	// NewOptionalLocalDateTime is a convenience function for creating an
	// OptionalLocalDateTime with its value set to v.
	NewOptionalLocalDateTime = geltypes.NewOptionalLocalDateTime

	// NewOptionalLocalTime is a convenience function for creating an
	// OptionalLocalTime with its value set to v.
	NewOptionalLocalTime = geltypes.NewOptionalLocalTime

	// NewOptionalMemory is a convenience function for creating an
	// OptionalMemory with its value set to v.
	NewOptionalMemory = geltypes.NewOptionalMemory

	// NewOptionalRangeDateTime is a convenience function for creating an
	// OptionalRangeDateTime with its value set to v.
	NewOptionalRangeDateTime = geltypes.NewOptionalRangeDateTime

	// NewOptionalRangeFloat32 is a convenience function for creating an
	// OptionalRangeFloat32 with its value set to v.
	NewOptionalRangeFloat32 = geltypes.NewOptionalRangeFloat32

	// NewOptionalRangeFloat64 is a convenience function for creating an
	// OptionalRangeFloat64 with its value set to v.
	NewOptionalRangeFloat64 = geltypes.NewOptionalRangeFloat64

	// NewOptionalRangeInt32 is a convenience function for creating an
	// OptionalRangeInt32 with its value set to v.
	NewOptionalRangeInt32 = geltypes.NewOptionalRangeInt32

	// NewOptionalRangeInt64 is a convenience function for creating an
	// OptionalRangeInt64 with its value set to v.
	NewOptionalRangeInt64 = geltypes.NewOptionalRangeInt64

	// NewOptionalRangeLocalDate is a convenience function for creating an
	// OptionalRangeLocalDate with its value set to v.
	NewOptionalRangeLocalDate = geltypes.NewOptionalRangeLocalDate

	// NewOptionalRangeLocalDateTime is a convenience function for creating an
	// OptionalRangeLocalDateTime with its value set to v.
	NewOptionalRangeLocalDateTime = geltypes.NewOptionalRangeLocalDateTime

	// NewOptionalRelativeDuration is a convenience function for creating an
	// OptionalRelativeDuration with its value set to v.
	NewOptionalRelativeDuration = geltypes.NewOptionalRelativeDuration

	// NewOptionalStr is a convenience function for creating an OptionalStr with
	// its value set to v.
	NewOptionalStr = geltypes.NewOptionalStr

	// NewOptionalUUID is a convenience function for creating an OptionalUUID with
	// its value set to v.
	NewOptionalUUID = geltypes.NewOptionalUUID

	// NewRangeDateTime creates a new RangeDateTime value.
	NewRangeDateTime = geltypes.NewRangeDateTime

	// NewRangeFloat32 creates a new RangeFloat32 value.
	NewRangeFloat32 = geltypes.NewRangeFloat32

	// NewRangeFloat64 creates a new RangeFloat64 value.
	NewRangeFloat64 = geltypes.NewRangeFloat64

	// NewRangeInt32 creates a new RangeInt32 value.
	NewRangeInt32 = geltypes.NewRangeInt32

	// NewRangeInt64 creates a new RangeInt64 value.
	NewRangeInt64 = geltypes.NewRangeInt64

	// NewRangeLocalDate creates a new RangeLocalDate value.
	NewRangeLocalDate = geltypes.NewRangeLocalDate

	// NewRangeLocalDateTime creates a new RangeLocalDateTime value.
	NewRangeLocalDateTime = geltypes.NewRangeLocalDateTime

	// NewRelativeDuration returns a new RelativeDuration
	NewRelativeDuration = geltypes.NewRelativeDuration

	// NewRetryOptions returns the default retry options.
	NewRetryOptions = gel.NewRetryOptions

	// NewRetryRule returns the default RetryRule value.
	NewRetryRule = gel.NewRetryRule

	// NewTxOptions returns the default TxOptions value.
	NewTxOptions = gel.NewTxOptions

	// ParseUUID parses s into a UUID or returns an error.
	ParseUUID = geltypes.ParseUUID

	// WarningsAsErrors is an gel.WarningHandler that returns warnings as
	// errors.
	WarningsAsErrors = gel.WarningsAsErrors
)
