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

package gel

import gel "github.com/geldata/gel-go/internal/client"

const (
	InternalServerError                    = gel.InternalServerError
	UnsupportedFeatureError                = gel.UnsupportedFeatureError
	ProtocolError                          = gel.ProtocolError
	BinaryProtocolError                    = gel.BinaryProtocolError
	UnsupportedProtocolVersionError        = gel.UnsupportedProtocolVersionError
	TypeSpecNotFoundError                  = gel.TypeSpecNotFoundError
	UnexpectedMessageError                 = gel.UnexpectedMessageError
	InputDataError                         = gel.InputDataError
	ParameterTypeMismatchError             = gel.ParameterTypeMismatchError
	StateMismatchError                     = gel.StateMismatchError
	ResultCardinalityMismatchError         = gel.ResultCardinalityMismatchError
	CapabilityError                        = gel.CapabilityError
	UnsupportedCapabilityError             = gel.UnsupportedCapabilityError
	DisabledCapabilityError                = gel.DisabledCapabilityError
	QueryError                             = gel.QueryError
	InvalidSyntaxError                     = gel.InvalidSyntaxError
	EdgeQLSyntaxError                      = gel.EdgeQLSyntaxError
	SchemaSyntaxError                      = gel.SchemaSyntaxError
	GraphQLSyntaxError                     = gel.GraphQLSyntaxError
	InvalidTypeError                       = gel.InvalidTypeError
	InvalidTargetError                     = gel.InvalidTargetError
	InvalidLinkTargetError                 = gel.InvalidLinkTargetError
	InvalidPropertyTargetError             = gel.InvalidPropertyTargetError
	InvalidReferenceError                  = gel.InvalidReferenceError
	UnknownModuleError                     = gel.UnknownModuleError
	UnknownLinkError                       = gel.UnknownLinkError
	UnknownPropertyError                   = gel.UnknownPropertyError
	UnknownUserError                       = gel.UnknownUserError
	UnknownDatabaseError                   = gel.UnknownDatabaseError
	UnknownParameterError                  = gel.UnknownParameterError
	DeprecatedScopingError                 = gel.DeprecatedScopingError
	SchemaError                            = gel.SchemaError
	SchemaDefinitionError                  = gel.SchemaDefinitionError
	InvalidDefinitionError                 = gel.InvalidDefinitionError
	InvalidModuleDefinitionError           = gel.InvalidModuleDefinitionError
	InvalidLinkDefinitionError             = gel.InvalidLinkDefinitionError
	InvalidPropertyDefinitionError         = gel.InvalidPropertyDefinitionError
	InvalidUserDefinitionError             = gel.InvalidUserDefinitionError
	InvalidDatabaseDefinitionError         = gel.InvalidDatabaseDefinitionError
	InvalidOperatorDefinitionError         = gel.InvalidOperatorDefinitionError
	InvalidAliasDefinitionError            = gel.InvalidAliasDefinitionError
	InvalidFunctionDefinitionError         = gel.InvalidFunctionDefinitionError
	InvalidConstraintDefinitionError       = gel.InvalidConstraintDefinitionError
	InvalidCastDefinitionError             = gel.InvalidCastDefinitionError
	DuplicateDefinitionError               = gel.DuplicateDefinitionError
	DuplicateModuleDefinitionError         = gel.DuplicateModuleDefinitionError
	DuplicateLinkDefinitionError           = gel.DuplicateLinkDefinitionError
	DuplicatePropertyDefinitionError       = gel.DuplicatePropertyDefinitionError
	DuplicateUserDefinitionError           = gel.DuplicateUserDefinitionError
	DuplicateDatabaseDefinitionError       = gel.DuplicateDatabaseDefinitionError
	DuplicateOperatorDefinitionError       = gel.DuplicateOperatorDefinitionError
	DuplicateViewDefinitionError           = gel.DuplicateViewDefinitionError
	DuplicateFunctionDefinitionError       = gel.DuplicateFunctionDefinitionError
	DuplicateConstraintDefinitionError     = gel.DuplicateConstraintDefinitionError
	DuplicateCastDefinitionError           = gel.DuplicateCastDefinitionError
	DuplicateMigrationError                = gel.DuplicateMigrationError
	SessionTimeoutError                    = gel.SessionTimeoutError
	IdleSessionTimeoutError                = gel.IdleSessionTimeoutError
	QueryTimeoutError                      = gel.QueryTimeoutError
	TransactionTimeoutError                = gel.TransactionTimeoutError
	IdleTransactionTimeoutError            = gel.IdleTransactionTimeoutError
	ExecutionError                         = gel.ExecutionError
	InvalidValueError                      = gel.InvalidValueError
	DivisionByZeroError                    = gel.DivisionByZeroError
	NumericOutOfRangeError                 = gel.NumericOutOfRangeError
	AccessPolicyError                      = gel.AccessPolicyError
	QueryAssertionError                    = gel.QueryAssertionError
	IntegrityError                         = gel.IntegrityError
	ConstraintViolationError               = gel.ConstraintViolationError
	CardinalityViolationError              = gel.CardinalityViolationError
	MissingRequiredError                   = gel.MissingRequiredError
	TransactionError                       = gel.TransactionError
	TransactionConflictError               = gel.TransactionConflictError
	TransactionSerializationError          = gel.TransactionSerializationError
	TransactionDeadlockError               = gel.TransactionDeadlockError
	WatchError                             = gel.WatchError
	ConfigurationError                     = gel.ConfigurationError
	AccessError                            = gel.AccessError
	AuthenticationError                    = gel.AuthenticationError
	AvailabilityError                      = gel.AvailabilityError
	BackendUnavailableError                = gel.BackendUnavailableError
	ServerOfflineError                     = gel.ServerOfflineError
	UnknownTenantError                     = gel.UnknownTenantError
	ServerBlockedError                     = gel.ServerBlockedError
	BackendError                           = gel.BackendError
	UnsupportedBackendFeatureError         = gel.UnsupportedBackendFeatureError
	ClientError                            = gel.ClientError
	ClientConnectionError                  = gel.ClientConnectionError
	ClientConnectionFailedError            = gel.ClientConnectionFailedError
	ClientConnectionFailedTemporarilyError = gel.ClientConnectionFailedTemporarilyError
	ClientConnectionTimeoutError           = gel.ClientConnectionTimeoutError
	ClientConnectionClosedError            = gel.ClientConnectionClosedError
	InterfaceError                         = gel.InterfaceError
	QueryArgumentError                     = gel.QueryArgumentError
	MissingArgumentError                   = gel.MissingArgumentError
	UnknownArgumentError                   = gel.UnknownArgumentError
	InvalidArgumentError                   = gel.InvalidArgumentError
	NoDataError                            = gel.NoDataError
	InternalClientError                    = gel.InternalClientError
	ShouldRetry                            = gel.ShouldRetry
	ShouldReconnect                        = gel.ShouldReconnect
)
