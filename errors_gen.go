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
// run 'make errors' to regenerate

package edgedb

import edgedb "github.com/edgedb/edgedb-go/internal/client"

const (
	InternalServerError                    = edgedb.InternalServerError
	UnsupportedFeatureError                = edgedb.UnsupportedFeatureError
	ProtocolError                          = edgedb.ProtocolError
	BinaryProtocolError                    = edgedb.BinaryProtocolError
	UnsupportedProtocolVersionError        = edgedb.UnsupportedProtocolVersionError
	TypeSpecNotFoundError                  = edgedb.TypeSpecNotFoundError
	UnexpectedMessageError                 = edgedb.UnexpectedMessageError
	InputDataError                         = edgedb.InputDataError
	ParameterTypeMismatchError             = edgedb.ParameterTypeMismatchError
	StateMismatchError                     = edgedb.StateMismatchError
	ResultCardinalityMismatchError         = edgedb.ResultCardinalityMismatchError
	CapabilityError                        = edgedb.CapabilityError
	UnsupportedCapabilityError             = edgedb.UnsupportedCapabilityError
	DisabledCapabilityError                = edgedb.DisabledCapabilityError
	QueryError                             = edgedb.QueryError
	InvalidSyntaxError                     = edgedb.InvalidSyntaxError
	EdgeQLSyntaxError                      = edgedb.EdgeQLSyntaxError
	SchemaSyntaxError                      = edgedb.SchemaSyntaxError
	GraphQLSyntaxError                     = edgedb.GraphQLSyntaxError
	InvalidTypeError                       = edgedb.InvalidTypeError
	InvalidTargetError                     = edgedb.InvalidTargetError
	InvalidLinkTargetError                 = edgedb.InvalidLinkTargetError
	InvalidPropertyTargetError             = edgedb.InvalidPropertyTargetError
	InvalidReferenceError                  = edgedb.InvalidReferenceError
	UnknownModuleError                     = edgedb.UnknownModuleError
	UnknownLinkError                       = edgedb.UnknownLinkError
	UnknownPropertyError                   = edgedb.UnknownPropertyError
	UnknownUserError                       = edgedb.UnknownUserError
	UnknownDatabaseError                   = edgedb.UnknownDatabaseError
	UnknownParameterError                  = edgedb.UnknownParameterError
	SchemaError                            = edgedb.SchemaError
	SchemaDefinitionError                  = edgedb.SchemaDefinitionError
	InvalidDefinitionError                 = edgedb.InvalidDefinitionError
	InvalidModuleDefinitionError           = edgedb.InvalidModuleDefinitionError
	InvalidLinkDefinitionError             = edgedb.InvalidLinkDefinitionError
	InvalidPropertyDefinitionError         = edgedb.InvalidPropertyDefinitionError
	InvalidUserDefinitionError             = edgedb.InvalidUserDefinitionError
	InvalidDatabaseDefinitionError         = edgedb.InvalidDatabaseDefinitionError
	InvalidOperatorDefinitionError         = edgedb.InvalidOperatorDefinitionError
	InvalidAliasDefinitionError            = edgedb.InvalidAliasDefinitionError
	InvalidFunctionDefinitionError         = edgedb.InvalidFunctionDefinitionError
	InvalidConstraintDefinitionError       = edgedb.InvalidConstraintDefinitionError
	InvalidCastDefinitionError             = edgedb.InvalidCastDefinitionError
	DuplicateDefinitionError               = edgedb.DuplicateDefinitionError
	DuplicateModuleDefinitionError         = edgedb.DuplicateModuleDefinitionError
	DuplicateLinkDefinitionError           = edgedb.DuplicateLinkDefinitionError
	DuplicatePropertyDefinitionError       = edgedb.DuplicatePropertyDefinitionError
	DuplicateUserDefinitionError           = edgedb.DuplicateUserDefinitionError
	DuplicateDatabaseDefinitionError       = edgedb.DuplicateDatabaseDefinitionError
	DuplicateOperatorDefinitionError       = edgedb.DuplicateOperatorDefinitionError
	DuplicateViewDefinitionError           = edgedb.DuplicateViewDefinitionError
	DuplicateFunctionDefinitionError       = edgedb.DuplicateFunctionDefinitionError
	DuplicateConstraintDefinitionError     = edgedb.DuplicateConstraintDefinitionError
	DuplicateCastDefinitionError           = edgedb.DuplicateCastDefinitionError
	SessionTimeoutError                    = edgedb.SessionTimeoutError
	IdleSessionTimeoutError                = edgedb.IdleSessionTimeoutError
	QueryTimeoutError                      = edgedb.QueryTimeoutError
	TransactionTimeoutError                = edgedb.TransactionTimeoutError
	IdleTransactionTimeoutError            = edgedb.IdleTransactionTimeoutError
	ExecutionError                         = edgedb.ExecutionError
	InvalidValueError                      = edgedb.InvalidValueError
	DivisionByZeroError                    = edgedb.DivisionByZeroError
	NumericOutOfRangeError                 = edgedb.NumericOutOfRangeError
	AccessPolicyError                      = edgedb.AccessPolicyError
	IntegrityError                         = edgedb.IntegrityError
	ConstraintViolationError               = edgedb.ConstraintViolationError
	CardinalityViolationError              = edgedb.CardinalityViolationError
	MissingRequiredError                   = edgedb.MissingRequiredError
	TransactionError                       = edgedb.TransactionError
	TransactionConflictError               = edgedb.TransactionConflictError
	TransactionSerializationError          = edgedb.TransactionSerializationError
	TransactionDeadlockError               = edgedb.TransactionDeadlockError
	ConfigurationError                     = edgedb.ConfigurationError
	AccessError                            = edgedb.AccessError
	AuthenticationError                    = edgedb.AuthenticationError
	AvailabilityError                      = edgedb.AvailabilityError
	BackendUnavailableError                = edgedb.BackendUnavailableError
	BackendError                           = edgedb.BackendError
	UnsupportedBackendFeatureError         = edgedb.UnsupportedBackendFeatureError
	ClientError                            = edgedb.ClientError
	ClientConnectionError                  = edgedb.ClientConnectionError
	ClientConnectionFailedError            = edgedb.ClientConnectionFailedError
	ClientConnectionFailedTemporarilyError = edgedb.ClientConnectionFailedTemporarilyError
	ClientConnectionTimeoutError           = edgedb.ClientConnectionTimeoutError
	ClientConnectionClosedError            = edgedb.ClientConnectionClosedError
	InterfaceError                         = edgedb.InterfaceError
	QueryArgumentError                     = edgedb.QueryArgumentError
	MissingArgumentError                   = edgedb.MissingArgumentError
	UnknownArgumentError                   = edgedb.UnknownArgumentError
	InvalidArgumentError                   = edgedb.InvalidArgumentError
	NoDataError                            = edgedb.NoDataError
	InternalClientError                    = edgedb.InternalClientError
	ShouldReconnect                        = edgedb.ShouldReconnect
	ShouldRetry                            = edgedb.ShouldRetry
)
