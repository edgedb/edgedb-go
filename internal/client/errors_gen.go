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

// internal/cmd/export should ignore this file
//go:build !export

package gel

import "fmt"

const (
	ShouldRetry     ErrorTag = "SHOULD_RETRY"
	ShouldReconnect ErrorTag = "SHOULD_RECONNECT"
)

const (
	InternalServerError                    ErrorCategory = "errors::InternalServerError"
	UnsupportedFeatureError                ErrorCategory = "errors::UnsupportedFeatureError"
	ProtocolError                          ErrorCategory = "errors::ProtocolError"
	BinaryProtocolError                    ErrorCategory = "errors::BinaryProtocolError"
	UnsupportedProtocolVersionError        ErrorCategory = "errors::UnsupportedProtocolVersionError"
	TypeSpecNotFoundError                  ErrorCategory = "errors::TypeSpecNotFoundError"
	UnexpectedMessageError                 ErrorCategory = "errors::UnexpectedMessageError"
	InputDataError                         ErrorCategory = "errors::InputDataError"
	ParameterTypeMismatchError             ErrorCategory = "errors::ParameterTypeMismatchError"
	StateMismatchError                     ErrorCategory = "errors::StateMismatchError"
	ResultCardinalityMismatchError         ErrorCategory = "errors::ResultCardinalityMismatchError"
	CapabilityError                        ErrorCategory = "errors::CapabilityError"
	UnsupportedCapabilityError             ErrorCategory = "errors::UnsupportedCapabilityError"
	DisabledCapabilityError                ErrorCategory = "errors::DisabledCapabilityError"
	QueryError                             ErrorCategory = "errors::QueryError"
	InvalidSyntaxError                     ErrorCategory = "errors::InvalidSyntaxError"
	EdgeQLSyntaxError                      ErrorCategory = "errors::EdgeQLSyntaxError"
	SchemaSyntaxError                      ErrorCategory = "errors::SchemaSyntaxError"
	GraphQLSyntaxError                     ErrorCategory = "errors::GraphQLSyntaxError"
	InvalidTypeError                       ErrorCategory = "errors::InvalidTypeError"
	InvalidTargetError                     ErrorCategory = "errors::InvalidTargetError"
	InvalidLinkTargetError                 ErrorCategory = "errors::InvalidLinkTargetError"
	InvalidPropertyTargetError             ErrorCategory = "errors::InvalidPropertyTargetError"
	InvalidReferenceError                  ErrorCategory = "errors::InvalidReferenceError"
	UnknownModuleError                     ErrorCategory = "errors::UnknownModuleError"
	UnknownLinkError                       ErrorCategory = "errors::UnknownLinkError"
	UnknownPropertyError                   ErrorCategory = "errors::UnknownPropertyError"
	UnknownUserError                       ErrorCategory = "errors::UnknownUserError"
	UnknownDatabaseError                   ErrorCategory = "errors::UnknownDatabaseError"
	UnknownParameterError                  ErrorCategory = "errors::UnknownParameterError"
	DeprecatedScopingError                 ErrorCategory = "errors::DeprecatedScopingError"
	SchemaError                            ErrorCategory = "errors::SchemaError"
	SchemaDefinitionError                  ErrorCategory = "errors::SchemaDefinitionError"
	InvalidDefinitionError                 ErrorCategory = "errors::InvalidDefinitionError"
	InvalidModuleDefinitionError           ErrorCategory = "errors::InvalidModuleDefinitionError"
	InvalidLinkDefinitionError             ErrorCategory = "errors::InvalidLinkDefinitionError"
	InvalidPropertyDefinitionError         ErrorCategory = "errors::InvalidPropertyDefinitionError"
	InvalidUserDefinitionError             ErrorCategory = "errors::InvalidUserDefinitionError"
	InvalidDatabaseDefinitionError         ErrorCategory = "errors::InvalidDatabaseDefinitionError"
	InvalidOperatorDefinitionError         ErrorCategory = "errors::InvalidOperatorDefinitionError"
	InvalidAliasDefinitionError            ErrorCategory = "errors::InvalidAliasDefinitionError"
	InvalidFunctionDefinitionError         ErrorCategory = "errors::InvalidFunctionDefinitionError"
	InvalidConstraintDefinitionError       ErrorCategory = "errors::InvalidConstraintDefinitionError"
	InvalidCastDefinitionError             ErrorCategory = "errors::InvalidCastDefinitionError"
	DuplicateDefinitionError               ErrorCategory = "errors::DuplicateDefinitionError"
	DuplicateModuleDefinitionError         ErrorCategory = "errors::DuplicateModuleDefinitionError"
	DuplicateLinkDefinitionError           ErrorCategory = "errors::DuplicateLinkDefinitionError"
	DuplicatePropertyDefinitionError       ErrorCategory = "errors::DuplicatePropertyDefinitionError"
	DuplicateUserDefinitionError           ErrorCategory = "errors::DuplicateUserDefinitionError"
	DuplicateDatabaseDefinitionError       ErrorCategory = "errors::DuplicateDatabaseDefinitionError"
	DuplicateOperatorDefinitionError       ErrorCategory = "errors::DuplicateOperatorDefinitionError"
	DuplicateViewDefinitionError           ErrorCategory = "errors::DuplicateViewDefinitionError"
	DuplicateFunctionDefinitionError       ErrorCategory = "errors::DuplicateFunctionDefinitionError"
	DuplicateConstraintDefinitionError     ErrorCategory = "errors::DuplicateConstraintDefinitionError"
	DuplicateCastDefinitionError           ErrorCategory = "errors::DuplicateCastDefinitionError"
	DuplicateMigrationError                ErrorCategory = "errors::DuplicateMigrationError"
	SessionTimeoutError                    ErrorCategory = "errors::SessionTimeoutError"
	IdleSessionTimeoutError                ErrorCategory = "errors::IdleSessionTimeoutError"
	QueryTimeoutError                      ErrorCategory = "errors::QueryTimeoutError"
	TransactionTimeoutError                ErrorCategory = "errors::TransactionTimeoutError"
	IdleTransactionTimeoutError            ErrorCategory = "errors::IdleTransactionTimeoutError"
	ExecutionError                         ErrorCategory = "errors::ExecutionError"
	InvalidValueError                      ErrorCategory = "errors::InvalidValueError"
	DivisionByZeroError                    ErrorCategory = "errors::DivisionByZeroError"
	NumericOutOfRangeError                 ErrorCategory = "errors::NumericOutOfRangeError"
	AccessPolicyError                      ErrorCategory = "errors::AccessPolicyError"
	QueryAssertionError                    ErrorCategory = "errors::QueryAssertionError"
	IntegrityError                         ErrorCategory = "errors::IntegrityError"
	ConstraintViolationError               ErrorCategory = "errors::ConstraintViolationError"
	CardinalityViolationError              ErrorCategory = "errors::CardinalityViolationError"
	MissingRequiredError                   ErrorCategory = "errors::MissingRequiredError"
	TransactionError                       ErrorCategory = "errors::TransactionError"
	TransactionConflictError               ErrorCategory = "errors::TransactionConflictError"
	TransactionSerializationError          ErrorCategory = "errors::TransactionSerializationError"
	TransactionDeadlockError               ErrorCategory = "errors::TransactionDeadlockError"
	WatchError                             ErrorCategory = "errors::WatchError"
	ConfigurationError                     ErrorCategory = "errors::ConfigurationError"
	AccessError                            ErrorCategory = "errors::AccessError"
	AuthenticationError                    ErrorCategory = "errors::AuthenticationError"
	AvailabilityError                      ErrorCategory = "errors::AvailabilityError"
	BackendUnavailableError                ErrorCategory = "errors::BackendUnavailableError"
	ServerOfflineError                     ErrorCategory = "errors::ServerOfflineError"
	UnknownTenantError                     ErrorCategory = "errors::UnknownTenantError"
	ServerBlockedError                     ErrorCategory = "errors::ServerBlockedError"
	BackendError                           ErrorCategory = "errors::BackendError"
	UnsupportedBackendFeatureError         ErrorCategory = "errors::UnsupportedBackendFeatureError"
	ClientError                            ErrorCategory = "errors::ClientError"
	ClientConnectionError                  ErrorCategory = "errors::ClientConnectionError"
	ClientConnectionFailedError            ErrorCategory = "errors::ClientConnectionFailedError"
	ClientConnectionFailedTemporarilyError ErrorCategory = "errors::ClientConnectionFailedTemporarilyError"
	ClientConnectionTimeoutError           ErrorCategory = "errors::ClientConnectionTimeoutError"
	ClientConnectionClosedError            ErrorCategory = "errors::ClientConnectionClosedError"
	InterfaceError                         ErrorCategory = "errors::InterfaceError"
	QueryArgumentError                     ErrorCategory = "errors::QueryArgumentError"
	MissingArgumentError                   ErrorCategory = "errors::MissingArgumentError"
	UnknownArgumentError                   ErrorCategory = "errors::UnknownArgumentError"
	InvalidArgumentError                   ErrorCategory = "errors::InvalidArgumentError"
	NoDataError                            ErrorCategory = "errors::NoDataError"
	InternalClientError                    ErrorCategory = "errors::InternalClientError"
)

type internalServerError struct {
	msg string
	err error
}

func (e *internalServerError) Error() string {
	msg := e.msg
	if e.err != nil {
		msg = e.err.Error()
	}

	return "gel.InternalServerError: " + msg
}

func (e *internalServerError) Unwrap() error { return e.err }

func (e *internalServerError) Category(c ErrorCategory) bool {
	switch c {
	case InternalServerError:
		return true
	default:
		return false
	}
}

func (e *internalServerError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}

type unsupportedFeatureError struct {
	msg string
	err error
}

func (e *unsupportedFeatureError) Error() string {
	msg := e.msg
	if e.err != nil {
		msg = e.err.Error()
	}

	return "gel.UnsupportedFeatureError: " + msg
}

func (e *unsupportedFeatureError) Unwrap() error { return e.err }

func (e *unsupportedFeatureError) Category(c ErrorCategory) bool {
	switch c {
	case UnsupportedFeatureError:
		return true
	default:
		return false
	}
}

func (e *unsupportedFeatureError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}

type protocolError struct {
	msg string
	err error
}

func (e *protocolError) Error() string {
	msg := e.msg
	if e.err != nil {
		msg = e.err.Error()
	}

	return "gel.ProtocolError: " + msg
}

func (e *protocolError) Unwrap() error { return e.err }

func (e *protocolError) Category(c ErrorCategory) bool {
	switch c {
	case ProtocolError:
		return true
	default:
		return false
	}
}

func (e *protocolError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}

type binaryProtocolError struct {
	msg string
	err error
}

func (e *binaryProtocolError) Error() string {
	msg := e.msg
	if e.err != nil {
		msg = e.err.Error()
	}

	return "gel.BinaryProtocolError: " + msg
}

func (e *binaryProtocolError) Unwrap() error { return e.err }

func (e *binaryProtocolError) Category(c ErrorCategory) bool {
	switch c {
	case BinaryProtocolError:
		return true
	case ProtocolError:
		return true
	default:
		return false
	}
}

func (e *binaryProtocolError) isEdgeDBProtocolError() {}

func (e *binaryProtocolError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}

type unsupportedProtocolVersionError struct {
	msg string
	err error
}

func (e *unsupportedProtocolVersionError) Error() string {
	msg := e.msg
	if e.err != nil {
		msg = e.err.Error()
	}

	return "gel.UnsupportedProtocolVersionError: " + msg
}

func (e *unsupportedProtocolVersionError) Unwrap() error { return e.err }

func (e *unsupportedProtocolVersionError) Category(c ErrorCategory) bool {
	switch c {
	case UnsupportedProtocolVersionError:
		return true
	case BinaryProtocolError:
		return true
	case ProtocolError:
		return true
	default:
		return false
	}
}

func (e *unsupportedProtocolVersionError) isEdgeDBBinaryProtocolError() {}

func (e *unsupportedProtocolVersionError) isEdgeDBProtocolError() {}

func (e *unsupportedProtocolVersionError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}

type typeSpecNotFoundError struct {
	msg string
	err error
}

func (e *typeSpecNotFoundError) Error() string {
	msg := e.msg
	if e.err != nil {
		msg = e.err.Error()
	}

	return "gel.TypeSpecNotFoundError: " + msg
}

func (e *typeSpecNotFoundError) Unwrap() error { return e.err }

func (e *typeSpecNotFoundError) Category(c ErrorCategory) bool {
	switch c {
	case TypeSpecNotFoundError:
		return true
	case BinaryProtocolError:
		return true
	case ProtocolError:
		return true
	default:
		return false
	}
}

func (e *typeSpecNotFoundError) isEdgeDBBinaryProtocolError() {}

func (e *typeSpecNotFoundError) isEdgeDBProtocolError() {}

func (e *typeSpecNotFoundError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}

type unexpectedMessageError struct {
	msg string
	err error
}

func (e *unexpectedMessageError) Error() string {
	msg := e.msg
	if e.err != nil {
		msg = e.err.Error()
	}

	return "gel.UnexpectedMessageError: " + msg
}

func (e *unexpectedMessageError) Unwrap() error { return e.err }

func (e *unexpectedMessageError) Category(c ErrorCategory) bool {
	switch c {
	case UnexpectedMessageError:
		return true
	case BinaryProtocolError:
		return true
	case ProtocolError:
		return true
	default:
		return false
	}
}

func (e *unexpectedMessageError) isEdgeDBBinaryProtocolError() {}

func (e *unexpectedMessageError) isEdgeDBProtocolError() {}

func (e *unexpectedMessageError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}

type inputDataError struct {
	msg string
	err error
}

func (e *inputDataError) Error() string {
	msg := e.msg
	if e.err != nil {
		msg = e.err.Error()
	}

	return "gel.InputDataError: " + msg
}

func (e *inputDataError) Unwrap() error { return e.err }

func (e *inputDataError) Category(c ErrorCategory) bool {
	switch c {
	case InputDataError:
		return true
	case ProtocolError:
		return true
	default:
		return false
	}
}

func (e *inputDataError) isEdgeDBProtocolError() {}

func (e *inputDataError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}

type parameterTypeMismatchError struct {
	msg string
	err error
}

func (e *parameterTypeMismatchError) Error() string {
	msg := e.msg
	if e.err != nil {
		msg = e.err.Error()
	}

	return "gel.ParameterTypeMismatchError: " + msg
}

func (e *parameterTypeMismatchError) Unwrap() error { return e.err }

func (e *parameterTypeMismatchError) Category(c ErrorCategory) bool {
	switch c {
	case ParameterTypeMismatchError:
		return true
	case InputDataError:
		return true
	case ProtocolError:
		return true
	default:
		return false
	}
}

func (e *parameterTypeMismatchError) isEdgeDBInputDataError() {}

func (e *parameterTypeMismatchError) isEdgeDBProtocolError() {}

func (e *parameterTypeMismatchError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}

type stateMismatchError struct {
	msg string
	err error
}

func (e *stateMismatchError) Error() string {
	msg := e.msg
	if e.err != nil {
		msg = e.err.Error()
	}

	return "gel.StateMismatchError: " + msg
}

func (e *stateMismatchError) Unwrap() error { return e.err }

func (e *stateMismatchError) Category(c ErrorCategory) bool {
	switch c {
	case StateMismatchError:
		return true
	case InputDataError:
		return true
	case ProtocolError:
		return true
	default:
		return false
	}
}

func (e *stateMismatchError) isEdgeDBInputDataError() {}

func (e *stateMismatchError) isEdgeDBProtocolError() {}

func (e *stateMismatchError) HasTag(tag ErrorTag) bool {
	switch tag {
	case ShouldRetry:
		return true
	default:
		return false
	}
}

type resultCardinalityMismatchError struct {
	msg string
	err error
}

func (e *resultCardinalityMismatchError) Error() string {
	msg := e.msg
	if e.err != nil {
		msg = e.err.Error()
	}

	return "gel.ResultCardinalityMismatchError: " + msg
}

func (e *resultCardinalityMismatchError) Unwrap() error { return e.err }

func (e *resultCardinalityMismatchError) Category(c ErrorCategory) bool {
	switch c {
	case ResultCardinalityMismatchError:
		return true
	case ProtocolError:
		return true
	default:
		return false
	}
}

func (e *resultCardinalityMismatchError) isEdgeDBProtocolError() {}

func (e *resultCardinalityMismatchError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}

type capabilityError struct {
	msg string
	err error
}

func (e *capabilityError) Error() string {
	msg := e.msg
	if e.err != nil {
		msg = e.err.Error()
	}

	return "gel.CapabilityError: " + msg
}

func (e *capabilityError) Unwrap() error { return e.err }

func (e *capabilityError) Category(c ErrorCategory) bool {
	switch c {
	case CapabilityError:
		return true
	case ProtocolError:
		return true
	default:
		return false
	}
}

func (e *capabilityError) isEdgeDBProtocolError() {}

func (e *capabilityError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}

type unsupportedCapabilityError struct {
	msg string
	err error
}

func (e *unsupportedCapabilityError) Error() string {
	msg := e.msg
	if e.err != nil {
		msg = e.err.Error()
	}

	return "gel.UnsupportedCapabilityError: " + msg
}

func (e *unsupportedCapabilityError) Unwrap() error { return e.err }

func (e *unsupportedCapabilityError) Category(c ErrorCategory) bool {
	switch c {
	case UnsupportedCapabilityError:
		return true
	case CapabilityError:
		return true
	case ProtocolError:
		return true
	default:
		return false
	}
}

func (e *unsupportedCapabilityError) isEdgeDBCapabilityError() {}

func (e *unsupportedCapabilityError) isEdgeDBProtocolError() {}

func (e *unsupportedCapabilityError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}

type disabledCapabilityError struct {
	msg string
	err error
}

func (e *disabledCapabilityError) Error() string {
	msg := e.msg
	if e.err != nil {
		msg = e.err.Error()
	}

	return "gel.DisabledCapabilityError: " + msg
}

func (e *disabledCapabilityError) Unwrap() error { return e.err }

func (e *disabledCapabilityError) Category(c ErrorCategory) bool {
	switch c {
	case DisabledCapabilityError:
		return true
	case CapabilityError:
		return true
	case ProtocolError:
		return true
	default:
		return false
	}
}

func (e *disabledCapabilityError) isEdgeDBCapabilityError() {}

func (e *disabledCapabilityError) isEdgeDBProtocolError() {}

func (e *disabledCapabilityError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}

type queryError struct {
	msg string
	err error
}

func (e *queryError) Error() string {
	msg := e.msg
	if e.err != nil {
		msg = e.err.Error()
	}

	return "gel.QueryError: " + msg
}

func (e *queryError) Unwrap() error { return e.err }

func (e *queryError) Category(c ErrorCategory) bool {
	switch c {
	case QueryError:
		return true
	default:
		return false
	}
}

func (e *queryError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}

type invalidSyntaxError struct {
	msg string
	err error
}

func (e *invalidSyntaxError) Error() string {
	msg := e.msg
	if e.err != nil {
		msg = e.err.Error()
	}

	return "gel.InvalidSyntaxError: " + msg
}

func (e *invalidSyntaxError) Unwrap() error { return e.err }

func (e *invalidSyntaxError) Category(c ErrorCategory) bool {
	switch c {
	case InvalidSyntaxError:
		return true
	case QueryError:
		return true
	default:
		return false
	}
}

func (e *invalidSyntaxError) isEdgeDBQueryError() {}

func (e *invalidSyntaxError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}

type edgeQLSyntaxError struct {
	msg string
	err error
}

func (e *edgeQLSyntaxError) Error() string {
	msg := e.msg
	if e.err != nil {
		msg = e.err.Error()
	}

	return "gel.EdgeQLSyntaxError: " + msg
}

func (e *edgeQLSyntaxError) Unwrap() error { return e.err }

func (e *edgeQLSyntaxError) Category(c ErrorCategory) bool {
	switch c {
	case EdgeQLSyntaxError:
		return true
	case InvalidSyntaxError:
		return true
	case QueryError:
		return true
	default:
		return false
	}
}

func (e *edgeQLSyntaxError) isEdgeDBInvalidSyntaxError() {}

func (e *edgeQLSyntaxError) isEdgeDBQueryError() {}

func (e *edgeQLSyntaxError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}

type schemaSyntaxError struct {
	msg string
	err error
}

func (e *schemaSyntaxError) Error() string {
	msg := e.msg
	if e.err != nil {
		msg = e.err.Error()
	}

	return "gel.SchemaSyntaxError: " + msg
}

func (e *schemaSyntaxError) Unwrap() error { return e.err }

func (e *schemaSyntaxError) Category(c ErrorCategory) bool {
	switch c {
	case SchemaSyntaxError:
		return true
	case InvalidSyntaxError:
		return true
	case QueryError:
		return true
	default:
		return false
	}
}

func (e *schemaSyntaxError) isEdgeDBInvalidSyntaxError() {}

func (e *schemaSyntaxError) isEdgeDBQueryError() {}

func (e *schemaSyntaxError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}

type graphQLSyntaxError struct {
	msg string
	err error
}

func (e *graphQLSyntaxError) Error() string {
	msg := e.msg
	if e.err != nil {
		msg = e.err.Error()
	}

	return "gel.GraphQLSyntaxError: " + msg
}

func (e *graphQLSyntaxError) Unwrap() error { return e.err }

func (e *graphQLSyntaxError) Category(c ErrorCategory) bool {
	switch c {
	case GraphQLSyntaxError:
		return true
	case InvalidSyntaxError:
		return true
	case QueryError:
		return true
	default:
		return false
	}
}

func (e *graphQLSyntaxError) isEdgeDBInvalidSyntaxError() {}

func (e *graphQLSyntaxError) isEdgeDBQueryError() {}

func (e *graphQLSyntaxError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}

type invalidTypeError struct {
	msg string
	err error
}

func (e *invalidTypeError) Error() string {
	msg := e.msg
	if e.err != nil {
		msg = e.err.Error()
	}

	return "gel.InvalidTypeError: " + msg
}

func (e *invalidTypeError) Unwrap() error { return e.err }

func (e *invalidTypeError) Category(c ErrorCategory) bool {
	switch c {
	case InvalidTypeError:
		return true
	case QueryError:
		return true
	default:
		return false
	}
}

func (e *invalidTypeError) isEdgeDBQueryError() {}

func (e *invalidTypeError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}

type invalidTargetError struct {
	msg string
	err error
}

func (e *invalidTargetError) Error() string {
	msg := e.msg
	if e.err != nil {
		msg = e.err.Error()
	}

	return "gel.InvalidTargetError: " + msg
}

func (e *invalidTargetError) Unwrap() error { return e.err }

func (e *invalidTargetError) Category(c ErrorCategory) bool {
	switch c {
	case InvalidTargetError:
		return true
	case InvalidTypeError:
		return true
	case QueryError:
		return true
	default:
		return false
	}
}

func (e *invalidTargetError) isEdgeDBInvalidTypeError() {}

func (e *invalidTargetError) isEdgeDBQueryError() {}

func (e *invalidTargetError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}

type invalidLinkTargetError struct {
	msg string
	err error
}

func (e *invalidLinkTargetError) Error() string {
	msg := e.msg
	if e.err != nil {
		msg = e.err.Error()
	}

	return "gel.InvalidLinkTargetError: " + msg
}

func (e *invalidLinkTargetError) Unwrap() error { return e.err }

func (e *invalidLinkTargetError) Category(c ErrorCategory) bool {
	switch c {
	case InvalidLinkTargetError:
		return true
	case InvalidTargetError:
		return true
	case InvalidTypeError:
		return true
	case QueryError:
		return true
	default:
		return false
	}
}

func (e *invalidLinkTargetError) isEdgeDBInvalidTargetError() {}

func (e *invalidLinkTargetError) isEdgeDBInvalidTypeError() {}

func (e *invalidLinkTargetError) isEdgeDBQueryError() {}

func (e *invalidLinkTargetError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}

type invalidPropertyTargetError struct {
	msg string
	err error
}

func (e *invalidPropertyTargetError) Error() string {
	msg := e.msg
	if e.err != nil {
		msg = e.err.Error()
	}

	return "gel.InvalidPropertyTargetError: " + msg
}

func (e *invalidPropertyTargetError) Unwrap() error { return e.err }

func (e *invalidPropertyTargetError) Category(c ErrorCategory) bool {
	switch c {
	case InvalidPropertyTargetError:
		return true
	case InvalidTargetError:
		return true
	case InvalidTypeError:
		return true
	case QueryError:
		return true
	default:
		return false
	}
}

func (e *invalidPropertyTargetError) isEdgeDBInvalidTargetError() {}

func (e *invalidPropertyTargetError) isEdgeDBInvalidTypeError() {}

func (e *invalidPropertyTargetError) isEdgeDBQueryError() {}

func (e *invalidPropertyTargetError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}

type invalidReferenceError struct {
	msg string
	err error
}

func (e *invalidReferenceError) Error() string {
	msg := e.msg
	if e.err != nil {
		msg = e.err.Error()
	}

	return "gel.InvalidReferenceError: " + msg
}

func (e *invalidReferenceError) Unwrap() error { return e.err }

func (e *invalidReferenceError) Category(c ErrorCategory) bool {
	switch c {
	case InvalidReferenceError:
		return true
	case QueryError:
		return true
	default:
		return false
	}
}

func (e *invalidReferenceError) isEdgeDBQueryError() {}

func (e *invalidReferenceError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}

type unknownModuleError struct {
	msg string
	err error
}

func (e *unknownModuleError) Error() string {
	msg := e.msg
	if e.err != nil {
		msg = e.err.Error()
	}

	return "gel.UnknownModuleError: " + msg
}

func (e *unknownModuleError) Unwrap() error { return e.err }

func (e *unknownModuleError) Category(c ErrorCategory) bool {
	switch c {
	case UnknownModuleError:
		return true
	case InvalidReferenceError:
		return true
	case QueryError:
		return true
	default:
		return false
	}
}

func (e *unknownModuleError) isEdgeDBInvalidReferenceError() {}

func (e *unknownModuleError) isEdgeDBQueryError() {}

func (e *unknownModuleError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}

type unknownLinkError struct {
	msg string
	err error
}

func (e *unknownLinkError) Error() string {
	msg := e.msg
	if e.err != nil {
		msg = e.err.Error()
	}

	return "gel.UnknownLinkError: " + msg
}

func (e *unknownLinkError) Unwrap() error { return e.err }

func (e *unknownLinkError) Category(c ErrorCategory) bool {
	switch c {
	case UnknownLinkError:
		return true
	case InvalidReferenceError:
		return true
	case QueryError:
		return true
	default:
		return false
	}
}

func (e *unknownLinkError) isEdgeDBInvalidReferenceError() {}

func (e *unknownLinkError) isEdgeDBQueryError() {}

func (e *unknownLinkError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}

type unknownPropertyError struct {
	msg string
	err error
}

func (e *unknownPropertyError) Error() string {
	msg := e.msg
	if e.err != nil {
		msg = e.err.Error()
	}

	return "gel.UnknownPropertyError: " + msg
}

func (e *unknownPropertyError) Unwrap() error { return e.err }

func (e *unknownPropertyError) Category(c ErrorCategory) bool {
	switch c {
	case UnknownPropertyError:
		return true
	case InvalidReferenceError:
		return true
	case QueryError:
		return true
	default:
		return false
	}
}

func (e *unknownPropertyError) isEdgeDBInvalidReferenceError() {}

func (e *unknownPropertyError) isEdgeDBQueryError() {}

func (e *unknownPropertyError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}

type unknownUserError struct {
	msg string
	err error
}

func (e *unknownUserError) Error() string {
	msg := e.msg
	if e.err != nil {
		msg = e.err.Error()
	}

	return "gel.UnknownUserError: " + msg
}

func (e *unknownUserError) Unwrap() error { return e.err }

func (e *unknownUserError) Category(c ErrorCategory) bool {
	switch c {
	case UnknownUserError:
		return true
	case InvalidReferenceError:
		return true
	case QueryError:
		return true
	default:
		return false
	}
}

func (e *unknownUserError) isEdgeDBInvalidReferenceError() {}

func (e *unknownUserError) isEdgeDBQueryError() {}

func (e *unknownUserError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}

type unknownDatabaseError struct {
	msg string
	err error
}

func (e *unknownDatabaseError) Error() string {
	msg := e.msg
	if e.err != nil {
		msg = e.err.Error()
	}

	return "gel.UnknownDatabaseError: " + msg
}

func (e *unknownDatabaseError) Unwrap() error { return e.err }

func (e *unknownDatabaseError) Category(c ErrorCategory) bool {
	switch c {
	case UnknownDatabaseError:
		return true
	case InvalidReferenceError:
		return true
	case QueryError:
		return true
	default:
		return false
	}
}

func (e *unknownDatabaseError) isEdgeDBInvalidReferenceError() {}

func (e *unknownDatabaseError) isEdgeDBQueryError() {}

func (e *unknownDatabaseError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}

type unknownParameterError struct {
	msg string
	err error
}

func (e *unknownParameterError) Error() string {
	msg := e.msg
	if e.err != nil {
		msg = e.err.Error()
	}

	return "gel.UnknownParameterError: " + msg
}

func (e *unknownParameterError) Unwrap() error { return e.err }

func (e *unknownParameterError) Category(c ErrorCategory) bool {
	switch c {
	case UnknownParameterError:
		return true
	case InvalidReferenceError:
		return true
	case QueryError:
		return true
	default:
		return false
	}
}

func (e *unknownParameterError) isEdgeDBInvalidReferenceError() {}

func (e *unknownParameterError) isEdgeDBQueryError() {}

func (e *unknownParameterError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}

type deprecatedScopingError struct {
	msg string
	err error
}

func (e *deprecatedScopingError) Error() string {
	msg := e.msg
	if e.err != nil {
		msg = e.err.Error()
	}

	return "gel.DeprecatedScopingError: " + msg
}

func (e *deprecatedScopingError) Unwrap() error { return e.err }

func (e *deprecatedScopingError) Category(c ErrorCategory) bool {
	switch c {
	case DeprecatedScopingError:
		return true
	case InvalidReferenceError:
		return true
	case QueryError:
		return true
	default:
		return false
	}
}

func (e *deprecatedScopingError) isEdgeDBInvalidReferenceError() {}

func (e *deprecatedScopingError) isEdgeDBQueryError() {}

func (e *deprecatedScopingError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}

type schemaError struct {
	msg string
	err error
}

func (e *schemaError) Error() string {
	msg := e.msg
	if e.err != nil {
		msg = e.err.Error()
	}

	return "gel.SchemaError: " + msg
}

func (e *schemaError) Unwrap() error { return e.err }

func (e *schemaError) Category(c ErrorCategory) bool {
	switch c {
	case SchemaError:
		return true
	case QueryError:
		return true
	default:
		return false
	}
}

func (e *schemaError) isEdgeDBQueryError() {}

func (e *schemaError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}

type schemaDefinitionError struct {
	msg string
	err error
}

func (e *schemaDefinitionError) Error() string {
	msg := e.msg
	if e.err != nil {
		msg = e.err.Error()
	}

	return "gel.SchemaDefinitionError: " + msg
}

func (e *schemaDefinitionError) Unwrap() error { return e.err }

func (e *schemaDefinitionError) Category(c ErrorCategory) bool {
	switch c {
	case SchemaDefinitionError:
		return true
	case QueryError:
		return true
	default:
		return false
	}
}

func (e *schemaDefinitionError) isEdgeDBQueryError() {}

func (e *schemaDefinitionError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}

type invalidDefinitionError struct {
	msg string
	err error
}

func (e *invalidDefinitionError) Error() string {
	msg := e.msg
	if e.err != nil {
		msg = e.err.Error()
	}

	return "gel.InvalidDefinitionError: " + msg
}

func (e *invalidDefinitionError) Unwrap() error { return e.err }

func (e *invalidDefinitionError) Category(c ErrorCategory) bool {
	switch c {
	case InvalidDefinitionError:
		return true
	case SchemaDefinitionError:
		return true
	case QueryError:
		return true
	default:
		return false
	}
}

func (e *invalidDefinitionError) isEdgeDBSchemaDefinitionError() {}

func (e *invalidDefinitionError) isEdgeDBQueryError() {}

func (e *invalidDefinitionError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}

type invalidModuleDefinitionError struct {
	msg string
	err error
}

func (e *invalidModuleDefinitionError) Error() string {
	msg := e.msg
	if e.err != nil {
		msg = e.err.Error()
	}

	return "gel.InvalidModuleDefinitionError: " + msg
}

func (e *invalidModuleDefinitionError) Unwrap() error { return e.err }

func (e *invalidModuleDefinitionError) Category(c ErrorCategory) bool {
	switch c {
	case InvalidModuleDefinitionError:
		return true
	case InvalidDefinitionError:
		return true
	case SchemaDefinitionError:
		return true
	case QueryError:
		return true
	default:
		return false
	}
}

func (e *invalidModuleDefinitionError) isEdgeDBInvalidDefinitionError() {}

func (e *invalidModuleDefinitionError) isEdgeDBSchemaDefinitionError() {}

func (e *invalidModuleDefinitionError) isEdgeDBQueryError() {}

func (e *invalidModuleDefinitionError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}

type invalidLinkDefinitionError struct {
	msg string
	err error
}

func (e *invalidLinkDefinitionError) Error() string {
	msg := e.msg
	if e.err != nil {
		msg = e.err.Error()
	}

	return "gel.InvalidLinkDefinitionError: " + msg
}

func (e *invalidLinkDefinitionError) Unwrap() error { return e.err }

func (e *invalidLinkDefinitionError) Category(c ErrorCategory) bool {
	switch c {
	case InvalidLinkDefinitionError:
		return true
	case InvalidDefinitionError:
		return true
	case SchemaDefinitionError:
		return true
	case QueryError:
		return true
	default:
		return false
	}
}

func (e *invalidLinkDefinitionError) isEdgeDBInvalidDefinitionError() {}

func (e *invalidLinkDefinitionError) isEdgeDBSchemaDefinitionError() {}

func (e *invalidLinkDefinitionError) isEdgeDBQueryError() {}

func (e *invalidLinkDefinitionError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}

type invalidPropertyDefinitionError struct {
	msg string
	err error
}

func (e *invalidPropertyDefinitionError) Error() string {
	msg := e.msg
	if e.err != nil {
		msg = e.err.Error()
	}

	return "gel.InvalidPropertyDefinitionError: " + msg
}

func (e *invalidPropertyDefinitionError) Unwrap() error { return e.err }

func (e *invalidPropertyDefinitionError) Category(c ErrorCategory) bool {
	switch c {
	case InvalidPropertyDefinitionError:
		return true
	case InvalidDefinitionError:
		return true
	case SchemaDefinitionError:
		return true
	case QueryError:
		return true
	default:
		return false
	}
}

func (e *invalidPropertyDefinitionError) isEdgeDBInvalidDefinitionError() {}

func (e *invalidPropertyDefinitionError) isEdgeDBSchemaDefinitionError() {}

func (e *invalidPropertyDefinitionError) isEdgeDBQueryError() {}

func (e *invalidPropertyDefinitionError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}

type invalidUserDefinitionError struct {
	msg string
	err error
}

func (e *invalidUserDefinitionError) Error() string {
	msg := e.msg
	if e.err != nil {
		msg = e.err.Error()
	}

	return "gel.InvalidUserDefinitionError: " + msg
}

func (e *invalidUserDefinitionError) Unwrap() error { return e.err }

func (e *invalidUserDefinitionError) Category(c ErrorCategory) bool {
	switch c {
	case InvalidUserDefinitionError:
		return true
	case InvalidDefinitionError:
		return true
	case SchemaDefinitionError:
		return true
	case QueryError:
		return true
	default:
		return false
	}
}

func (e *invalidUserDefinitionError) isEdgeDBInvalidDefinitionError() {}

func (e *invalidUserDefinitionError) isEdgeDBSchemaDefinitionError() {}

func (e *invalidUserDefinitionError) isEdgeDBQueryError() {}

func (e *invalidUserDefinitionError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}

type invalidDatabaseDefinitionError struct {
	msg string
	err error
}

func (e *invalidDatabaseDefinitionError) Error() string {
	msg := e.msg
	if e.err != nil {
		msg = e.err.Error()
	}

	return "gel.InvalidDatabaseDefinitionError: " + msg
}

func (e *invalidDatabaseDefinitionError) Unwrap() error { return e.err }

func (e *invalidDatabaseDefinitionError) Category(c ErrorCategory) bool {
	switch c {
	case InvalidDatabaseDefinitionError:
		return true
	case InvalidDefinitionError:
		return true
	case SchemaDefinitionError:
		return true
	case QueryError:
		return true
	default:
		return false
	}
}

func (e *invalidDatabaseDefinitionError) isEdgeDBInvalidDefinitionError() {}

func (e *invalidDatabaseDefinitionError) isEdgeDBSchemaDefinitionError() {}

func (e *invalidDatabaseDefinitionError) isEdgeDBQueryError() {}

func (e *invalidDatabaseDefinitionError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}

type invalidOperatorDefinitionError struct {
	msg string
	err error
}

func (e *invalidOperatorDefinitionError) Error() string {
	msg := e.msg
	if e.err != nil {
		msg = e.err.Error()
	}

	return "gel.InvalidOperatorDefinitionError: " + msg
}

func (e *invalidOperatorDefinitionError) Unwrap() error { return e.err }

func (e *invalidOperatorDefinitionError) Category(c ErrorCategory) bool {
	switch c {
	case InvalidOperatorDefinitionError:
		return true
	case InvalidDefinitionError:
		return true
	case SchemaDefinitionError:
		return true
	case QueryError:
		return true
	default:
		return false
	}
}

func (e *invalidOperatorDefinitionError) isEdgeDBInvalidDefinitionError() {}

func (e *invalidOperatorDefinitionError) isEdgeDBSchemaDefinitionError() {}

func (e *invalidOperatorDefinitionError) isEdgeDBQueryError() {}

func (e *invalidOperatorDefinitionError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}

type invalidAliasDefinitionError struct {
	msg string
	err error
}

func (e *invalidAliasDefinitionError) Error() string {
	msg := e.msg
	if e.err != nil {
		msg = e.err.Error()
	}

	return "gel.InvalidAliasDefinitionError: " + msg
}

func (e *invalidAliasDefinitionError) Unwrap() error { return e.err }

func (e *invalidAliasDefinitionError) Category(c ErrorCategory) bool {
	switch c {
	case InvalidAliasDefinitionError:
		return true
	case InvalidDefinitionError:
		return true
	case SchemaDefinitionError:
		return true
	case QueryError:
		return true
	default:
		return false
	}
}

func (e *invalidAliasDefinitionError) isEdgeDBInvalidDefinitionError() {}

func (e *invalidAliasDefinitionError) isEdgeDBSchemaDefinitionError() {}

func (e *invalidAliasDefinitionError) isEdgeDBQueryError() {}

func (e *invalidAliasDefinitionError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}

type invalidFunctionDefinitionError struct {
	msg string
	err error
}

func (e *invalidFunctionDefinitionError) Error() string {
	msg := e.msg
	if e.err != nil {
		msg = e.err.Error()
	}

	return "gel.InvalidFunctionDefinitionError: " + msg
}

func (e *invalidFunctionDefinitionError) Unwrap() error { return e.err }

func (e *invalidFunctionDefinitionError) Category(c ErrorCategory) bool {
	switch c {
	case InvalidFunctionDefinitionError:
		return true
	case InvalidDefinitionError:
		return true
	case SchemaDefinitionError:
		return true
	case QueryError:
		return true
	default:
		return false
	}
}

func (e *invalidFunctionDefinitionError) isEdgeDBInvalidDefinitionError() {}

func (e *invalidFunctionDefinitionError) isEdgeDBSchemaDefinitionError() {}

func (e *invalidFunctionDefinitionError) isEdgeDBQueryError() {}

func (e *invalidFunctionDefinitionError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}

type invalidConstraintDefinitionError struct {
	msg string
	err error
}

func (e *invalidConstraintDefinitionError) Error() string {
	msg := e.msg
	if e.err != nil {
		msg = e.err.Error()
	}

	return "gel.InvalidConstraintDefinitionError: " + msg
}

func (e *invalidConstraintDefinitionError) Unwrap() error { return e.err }

func (e *invalidConstraintDefinitionError) Category(c ErrorCategory) bool {
	switch c {
	case InvalidConstraintDefinitionError:
		return true
	case InvalidDefinitionError:
		return true
	case SchemaDefinitionError:
		return true
	case QueryError:
		return true
	default:
		return false
	}
}

func (e *invalidConstraintDefinitionError) isEdgeDBInvalidDefinitionError() {}

func (e *invalidConstraintDefinitionError) isEdgeDBSchemaDefinitionError() {}

func (e *invalidConstraintDefinitionError) isEdgeDBQueryError() {}

func (e *invalidConstraintDefinitionError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}

type invalidCastDefinitionError struct {
	msg string
	err error
}

func (e *invalidCastDefinitionError) Error() string {
	msg := e.msg
	if e.err != nil {
		msg = e.err.Error()
	}

	return "gel.InvalidCastDefinitionError: " + msg
}

func (e *invalidCastDefinitionError) Unwrap() error { return e.err }

func (e *invalidCastDefinitionError) Category(c ErrorCategory) bool {
	switch c {
	case InvalidCastDefinitionError:
		return true
	case InvalidDefinitionError:
		return true
	case SchemaDefinitionError:
		return true
	case QueryError:
		return true
	default:
		return false
	}
}

func (e *invalidCastDefinitionError) isEdgeDBInvalidDefinitionError() {}

func (e *invalidCastDefinitionError) isEdgeDBSchemaDefinitionError() {}

func (e *invalidCastDefinitionError) isEdgeDBQueryError() {}

func (e *invalidCastDefinitionError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}

type duplicateDefinitionError struct {
	msg string
	err error
}

func (e *duplicateDefinitionError) Error() string {
	msg := e.msg
	if e.err != nil {
		msg = e.err.Error()
	}

	return "gel.DuplicateDefinitionError: " + msg
}

func (e *duplicateDefinitionError) Unwrap() error { return e.err }

func (e *duplicateDefinitionError) Category(c ErrorCategory) bool {
	switch c {
	case DuplicateDefinitionError:
		return true
	case SchemaDefinitionError:
		return true
	case QueryError:
		return true
	default:
		return false
	}
}

func (e *duplicateDefinitionError) isEdgeDBSchemaDefinitionError() {}

func (e *duplicateDefinitionError) isEdgeDBQueryError() {}

func (e *duplicateDefinitionError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}

type duplicateModuleDefinitionError struct {
	msg string
	err error
}

func (e *duplicateModuleDefinitionError) Error() string {
	msg := e.msg
	if e.err != nil {
		msg = e.err.Error()
	}

	return "gel.DuplicateModuleDefinitionError: " + msg
}

func (e *duplicateModuleDefinitionError) Unwrap() error { return e.err }

func (e *duplicateModuleDefinitionError) Category(c ErrorCategory) bool {
	switch c {
	case DuplicateModuleDefinitionError:
		return true
	case DuplicateDefinitionError:
		return true
	case SchemaDefinitionError:
		return true
	case QueryError:
		return true
	default:
		return false
	}
}

func (e *duplicateModuleDefinitionError) isEdgeDBDuplicateDefinitionError() {}

func (e *duplicateModuleDefinitionError) isEdgeDBSchemaDefinitionError() {}

func (e *duplicateModuleDefinitionError) isEdgeDBQueryError() {}

func (e *duplicateModuleDefinitionError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}

type duplicateLinkDefinitionError struct {
	msg string
	err error
}

func (e *duplicateLinkDefinitionError) Error() string {
	msg := e.msg
	if e.err != nil {
		msg = e.err.Error()
	}

	return "gel.DuplicateLinkDefinitionError: " + msg
}

func (e *duplicateLinkDefinitionError) Unwrap() error { return e.err }

func (e *duplicateLinkDefinitionError) Category(c ErrorCategory) bool {
	switch c {
	case DuplicateLinkDefinitionError:
		return true
	case DuplicateDefinitionError:
		return true
	case SchemaDefinitionError:
		return true
	case QueryError:
		return true
	default:
		return false
	}
}

func (e *duplicateLinkDefinitionError) isEdgeDBDuplicateDefinitionError() {}

func (e *duplicateLinkDefinitionError) isEdgeDBSchemaDefinitionError() {}

func (e *duplicateLinkDefinitionError) isEdgeDBQueryError() {}

func (e *duplicateLinkDefinitionError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}

type duplicatePropertyDefinitionError struct {
	msg string
	err error
}

func (e *duplicatePropertyDefinitionError) Error() string {
	msg := e.msg
	if e.err != nil {
		msg = e.err.Error()
	}

	return "gel.DuplicatePropertyDefinitionError: " + msg
}

func (e *duplicatePropertyDefinitionError) Unwrap() error { return e.err }

func (e *duplicatePropertyDefinitionError) Category(c ErrorCategory) bool {
	switch c {
	case DuplicatePropertyDefinitionError:
		return true
	case DuplicateDefinitionError:
		return true
	case SchemaDefinitionError:
		return true
	case QueryError:
		return true
	default:
		return false
	}
}

func (e *duplicatePropertyDefinitionError) isEdgeDBDuplicateDefinitionError() {}

func (e *duplicatePropertyDefinitionError) isEdgeDBSchemaDefinitionError() {}

func (e *duplicatePropertyDefinitionError) isEdgeDBQueryError() {}

func (e *duplicatePropertyDefinitionError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}

type duplicateUserDefinitionError struct {
	msg string
	err error
}

func (e *duplicateUserDefinitionError) Error() string {
	msg := e.msg
	if e.err != nil {
		msg = e.err.Error()
	}

	return "gel.DuplicateUserDefinitionError: " + msg
}

func (e *duplicateUserDefinitionError) Unwrap() error { return e.err }

func (e *duplicateUserDefinitionError) Category(c ErrorCategory) bool {
	switch c {
	case DuplicateUserDefinitionError:
		return true
	case DuplicateDefinitionError:
		return true
	case SchemaDefinitionError:
		return true
	case QueryError:
		return true
	default:
		return false
	}
}

func (e *duplicateUserDefinitionError) isEdgeDBDuplicateDefinitionError() {}

func (e *duplicateUserDefinitionError) isEdgeDBSchemaDefinitionError() {}

func (e *duplicateUserDefinitionError) isEdgeDBQueryError() {}

func (e *duplicateUserDefinitionError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}

type duplicateDatabaseDefinitionError struct {
	msg string
	err error
}

func (e *duplicateDatabaseDefinitionError) Error() string {
	msg := e.msg
	if e.err != nil {
		msg = e.err.Error()
	}

	return "gel.DuplicateDatabaseDefinitionError: " + msg
}

func (e *duplicateDatabaseDefinitionError) Unwrap() error { return e.err }

func (e *duplicateDatabaseDefinitionError) Category(c ErrorCategory) bool {
	switch c {
	case DuplicateDatabaseDefinitionError:
		return true
	case DuplicateDefinitionError:
		return true
	case SchemaDefinitionError:
		return true
	case QueryError:
		return true
	default:
		return false
	}
}

func (e *duplicateDatabaseDefinitionError) isEdgeDBDuplicateDefinitionError() {}

func (e *duplicateDatabaseDefinitionError) isEdgeDBSchemaDefinitionError() {}

func (e *duplicateDatabaseDefinitionError) isEdgeDBQueryError() {}

func (e *duplicateDatabaseDefinitionError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}

type duplicateOperatorDefinitionError struct {
	msg string
	err error
}

func (e *duplicateOperatorDefinitionError) Error() string {
	msg := e.msg
	if e.err != nil {
		msg = e.err.Error()
	}

	return "gel.DuplicateOperatorDefinitionError: " + msg
}

func (e *duplicateOperatorDefinitionError) Unwrap() error { return e.err }

func (e *duplicateOperatorDefinitionError) Category(c ErrorCategory) bool {
	switch c {
	case DuplicateOperatorDefinitionError:
		return true
	case DuplicateDefinitionError:
		return true
	case SchemaDefinitionError:
		return true
	case QueryError:
		return true
	default:
		return false
	}
}

func (e *duplicateOperatorDefinitionError) isEdgeDBDuplicateDefinitionError() {}

func (e *duplicateOperatorDefinitionError) isEdgeDBSchemaDefinitionError() {}

func (e *duplicateOperatorDefinitionError) isEdgeDBQueryError() {}

func (e *duplicateOperatorDefinitionError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}

type duplicateViewDefinitionError struct {
	msg string
	err error
}

func (e *duplicateViewDefinitionError) Error() string {
	msg := e.msg
	if e.err != nil {
		msg = e.err.Error()
	}

	return "gel.DuplicateViewDefinitionError: " + msg
}

func (e *duplicateViewDefinitionError) Unwrap() error { return e.err }

func (e *duplicateViewDefinitionError) Category(c ErrorCategory) bool {
	switch c {
	case DuplicateViewDefinitionError:
		return true
	case DuplicateDefinitionError:
		return true
	case SchemaDefinitionError:
		return true
	case QueryError:
		return true
	default:
		return false
	}
}

func (e *duplicateViewDefinitionError) isEdgeDBDuplicateDefinitionError() {}

func (e *duplicateViewDefinitionError) isEdgeDBSchemaDefinitionError() {}

func (e *duplicateViewDefinitionError) isEdgeDBQueryError() {}

func (e *duplicateViewDefinitionError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}

type duplicateFunctionDefinitionError struct {
	msg string
	err error
}

func (e *duplicateFunctionDefinitionError) Error() string {
	msg := e.msg
	if e.err != nil {
		msg = e.err.Error()
	}

	return "gel.DuplicateFunctionDefinitionError: " + msg
}

func (e *duplicateFunctionDefinitionError) Unwrap() error { return e.err }

func (e *duplicateFunctionDefinitionError) Category(c ErrorCategory) bool {
	switch c {
	case DuplicateFunctionDefinitionError:
		return true
	case DuplicateDefinitionError:
		return true
	case SchemaDefinitionError:
		return true
	case QueryError:
		return true
	default:
		return false
	}
}

func (e *duplicateFunctionDefinitionError) isEdgeDBDuplicateDefinitionError() {}

func (e *duplicateFunctionDefinitionError) isEdgeDBSchemaDefinitionError() {}

func (e *duplicateFunctionDefinitionError) isEdgeDBQueryError() {}

func (e *duplicateFunctionDefinitionError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}

type duplicateConstraintDefinitionError struct {
	msg string
	err error
}

func (e *duplicateConstraintDefinitionError) Error() string {
	msg := e.msg
	if e.err != nil {
		msg = e.err.Error()
	}

	return "gel.DuplicateConstraintDefinitionError: " + msg
}

func (e *duplicateConstraintDefinitionError) Unwrap() error { return e.err }

func (e *duplicateConstraintDefinitionError) Category(c ErrorCategory) bool {
	switch c {
	case DuplicateConstraintDefinitionError:
		return true
	case DuplicateDefinitionError:
		return true
	case SchemaDefinitionError:
		return true
	case QueryError:
		return true
	default:
		return false
	}
}

func (e *duplicateConstraintDefinitionError) isEdgeDBDuplicateDefinitionError() {}

func (e *duplicateConstraintDefinitionError) isEdgeDBSchemaDefinitionError() {}

func (e *duplicateConstraintDefinitionError) isEdgeDBQueryError() {}

func (e *duplicateConstraintDefinitionError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}

type duplicateCastDefinitionError struct {
	msg string
	err error
}

func (e *duplicateCastDefinitionError) Error() string {
	msg := e.msg
	if e.err != nil {
		msg = e.err.Error()
	}

	return "gel.DuplicateCastDefinitionError: " + msg
}

func (e *duplicateCastDefinitionError) Unwrap() error { return e.err }

func (e *duplicateCastDefinitionError) Category(c ErrorCategory) bool {
	switch c {
	case DuplicateCastDefinitionError:
		return true
	case DuplicateDefinitionError:
		return true
	case SchemaDefinitionError:
		return true
	case QueryError:
		return true
	default:
		return false
	}
}

func (e *duplicateCastDefinitionError) isEdgeDBDuplicateDefinitionError() {}

func (e *duplicateCastDefinitionError) isEdgeDBSchemaDefinitionError() {}

func (e *duplicateCastDefinitionError) isEdgeDBQueryError() {}

func (e *duplicateCastDefinitionError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}

type duplicateMigrationError struct {
	msg string
	err error
}

func (e *duplicateMigrationError) Error() string {
	msg := e.msg
	if e.err != nil {
		msg = e.err.Error()
	}

	return "gel.DuplicateMigrationError: " + msg
}

func (e *duplicateMigrationError) Unwrap() error { return e.err }

func (e *duplicateMigrationError) Category(c ErrorCategory) bool {
	switch c {
	case DuplicateMigrationError:
		return true
	case DuplicateDefinitionError:
		return true
	case SchemaDefinitionError:
		return true
	case QueryError:
		return true
	default:
		return false
	}
}

func (e *duplicateMigrationError) isEdgeDBDuplicateDefinitionError() {}

func (e *duplicateMigrationError) isEdgeDBSchemaDefinitionError() {}

func (e *duplicateMigrationError) isEdgeDBQueryError() {}

func (e *duplicateMigrationError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}

type sessionTimeoutError struct {
	msg string
	err error
}

func (e *sessionTimeoutError) Error() string {
	msg := e.msg
	if e.err != nil {
		msg = e.err.Error()
	}

	return "gel.SessionTimeoutError: " + msg
}

func (e *sessionTimeoutError) Unwrap() error { return e.err }

func (e *sessionTimeoutError) Category(c ErrorCategory) bool {
	switch c {
	case SessionTimeoutError:
		return true
	case QueryError:
		return true
	default:
		return false
	}
}

func (e *sessionTimeoutError) isEdgeDBQueryError() {}

func (e *sessionTimeoutError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}

type idleSessionTimeoutError struct {
	msg string
	err error
}

func (e *idleSessionTimeoutError) Error() string {
	msg := e.msg
	if e.err != nil {
		msg = e.err.Error()
	}

	return "gel.IdleSessionTimeoutError: " + msg
}

func (e *idleSessionTimeoutError) Unwrap() error { return e.err }

func (e *idleSessionTimeoutError) Category(c ErrorCategory) bool {
	switch c {
	case IdleSessionTimeoutError:
		return true
	case SessionTimeoutError:
		return true
	case QueryError:
		return true
	default:
		return false
	}
}

func (e *idleSessionTimeoutError) isEdgeDBSessionTimeoutError() {}

func (e *idleSessionTimeoutError) isEdgeDBQueryError() {}

func (e *idleSessionTimeoutError) HasTag(tag ErrorTag) bool {
	switch tag {
	case ShouldRetry:
		return true
	default:
		return false
	}
}

type queryTimeoutError struct {
	msg string
	err error
}

func (e *queryTimeoutError) Error() string {
	msg := e.msg
	if e.err != nil {
		msg = e.err.Error()
	}

	return "gel.QueryTimeoutError: " + msg
}

func (e *queryTimeoutError) Unwrap() error { return e.err }

func (e *queryTimeoutError) Category(c ErrorCategory) bool {
	switch c {
	case QueryTimeoutError:
		return true
	case SessionTimeoutError:
		return true
	case QueryError:
		return true
	default:
		return false
	}
}

func (e *queryTimeoutError) isEdgeDBSessionTimeoutError() {}

func (e *queryTimeoutError) isEdgeDBQueryError() {}

func (e *queryTimeoutError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}

type transactionTimeoutError struct {
	msg string
	err error
}

func (e *transactionTimeoutError) Error() string {
	msg := e.msg
	if e.err != nil {
		msg = e.err.Error()
	}

	return "gel.TransactionTimeoutError: " + msg
}

func (e *transactionTimeoutError) Unwrap() error { return e.err }

func (e *transactionTimeoutError) Category(c ErrorCategory) bool {
	switch c {
	case TransactionTimeoutError:
		return true
	case SessionTimeoutError:
		return true
	case QueryError:
		return true
	default:
		return false
	}
}

func (e *transactionTimeoutError) isEdgeDBSessionTimeoutError() {}

func (e *transactionTimeoutError) isEdgeDBQueryError() {}

func (e *transactionTimeoutError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}

type idleTransactionTimeoutError struct {
	msg string
	err error
}

func (e *idleTransactionTimeoutError) Error() string {
	msg := e.msg
	if e.err != nil {
		msg = e.err.Error()
	}

	return "gel.IdleTransactionTimeoutError: " + msg
}

func (e *idleTransactionTimeoutError) Unwrap() error { return e.err }

func (e *idleTransactionTimeoutError) Category(c ErrorCategory) bool {
	switch c {
	case IdleTransactionTimeoutError:
		return true
	case TransactionTimeoutError:
		return true
	case SessionTimeoutError:
		return true
	case QueryError:
		return true
	default:
		return false
	}
}

func (e *idleTransactionTimeoutError) isEdgeDBTransactionTimeoutError() {}

func (e *idleTransactionTimeoutError) isEdgeDBSessionTimeoutError() {}

func (e *idleTransactionTimeoutError) isEdgeDBQueryError() {}

func (e *idleTransactionTimeoutError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}

type executionError struct {
	msg string
	err error
}

func (e *executionError) Error() string {
	msg := e.msg
	if e.err != nil {
		msg = e.err.Error()
	}

	return "gel.ExecutionError: " + msg
}

func (e *executionError) Unwrap() error { return e.err }

func (e *executionError) Category(c ErrorCategory) bool {
	switch c {
	case ExecutionError:
		return true
	default:
		return false
	}
}

func (e *executionError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}

type invalidValueError struct {
	msg string
	err error
}

func (e *invalidValueError) Error() string {
	msg := e.msg
	if e.err != nil {
		msg = e.err.Error()
	}

	return "gel.InvalidValueError: " + msg
}

func (e *invalidValueError) Unwrap() error { return e.err }

func (e *invalidValueError) Category(c ErrorCategory) bool {
	switch c {
	case InvalidValueError:
		return true
	case ExecutionError:
		return true
	default:
		return false
	}
}

func (e *invalidValueError) isEdgeDBExecutionError() {}

func (e *invalidValueError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}

type divisionByZeroError struct {
	msg string
	err error
}

func (e *divisionByZeroError) Error() string {
	msg := e.msg
	if e.err != nil {
		msg = e.err.Error()
	}

	return "gel.DivisionByZeroError: " + msg
}

func (e *divisionByZeroError) Unwrap() error { return e.err }

func (e *divisionByZeroError) Category(c ErrorCategory) bool {
	switch c {
	case DivisionByZeroError:
		return true
	case InvalidValueError:
		return true
	case ExecutionError:
		return true
	default:
		return false
	}
}

func (e *divisionByZeroError) isEdgeDBInvalidValueError() {}

func (e *divisionByZeroError) isEdgeDBExecutionError() {}

func (e *divisionByZeroError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}

type numericOutOfRangeError struct {
	msg string
	err error
}

func (e *numericOutOfRangeError) Error() string {
	msg := e.msg
	if e.err != nil {
		msg = e.err.Error()
	}

	return "gel.NumericOutOfRangeError: " + msg
}

func (e *numericOutOfRangeError) Unwrap() error { return e.err }

func (e *numericOutOfRangeError) Category(c ErrorCategory) bool {
	switch c {
	case NumericOutOfRangeError:
		return true
	case InvalidValueError:
		return true
	case ExecutionError:
		return true
	default:
		return false
	}
}

func (e *numericOutOfRangeError) isEdgeDBInvalidValueError() {}

func (e *numericOutOfRangeError) isEdgeDBExecutionError() {}

func (e *numericOutOfRangeError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}

type accessPolicyError struct {
	msg string
	err error
}

func (e *accessPolicyError) Error() string {
	msg := e.msg
	if e.err != nil {
		msg = e.err.Error()
	}

	return "gel.AccessPolicyError: " + msg
}

func (e *accessPolicyError) Unwrap() error { return e.err }

func (e *accessPolicyError) Category(c ErrorCategory) bool {
	switch c {
	case AccessPolicyError:
		return true
	case InvalidValueError:
		return true
	case ExecutionError:
		return true
	default:
		return false
	}
}

func (e *accessPolicyError) isEdgeDBInvalidValueError() {}

func (e *accessPolicyError) isEdgeDBExecutionError() {}

func (e *accessPolicyError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}

type queryAssertionError struct {
	msg string
	err error
}

func (e *queryAssertionError) Error() string {
	msg := e.msg
	if e.err != nil {
		msg = e.err.Error()
	}

	return "gel.QueryAssertionError: " + msg
}

func (e *queryAssertionError) Unwrap() error { return e.err }

func (e *queryAssertionError) Category(c ErrorCategory) bool {
	switch c {
	case QueryAssertionError:
		return true
	case InvalidValueError:
		return true
	case ExecutionError:
		return true
	default:
		return false
	}
}

func (e *queryAssertionError) isEdgeDBInvalidValueError() {}

func (e *queryAssertionError) isEdgeDBExecutionError() {}

func (e *queryAssertionError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}

type integrityError struct {
	msg string
	err error
}

func (e *integrityError) Error() string {
	msg := e.msg
	if e.err != nil {
		msg = e.err.Error()
	}

	return "gel.IntegrityError: " + msg
}

func (e *integrityError) Unwrap() error { return e.err }

func (e *integrityError) Category(c ErrorCategory) bool {
	switch c {
	case IntegrityError:
		return true
	case ExecutionError:
		return true
	default:
		return false
	}
}

func (e *integrityError) isEdgeDBExecutionError() {}

func (e *integrityError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}

type constraintViolationError struct {
	msg string
	err error
}

func (e *constraintViolationError) Error() string {
	msg := e.msg
	if e.err != nil {
		msg = e.err.Error()
	}

	return "gel.ConstraintViolationError: " + msg
}

func (e *constraintViolationError) Unwrap() error { return e.err }

func (e *constraintViolationError) Category(c ErrorCategory) bool {
	switch c {
	case ConstraintViolationError:
		return true
	case IntegrityError:
		return true
	case ExecutionError:
		return true
	default:
		return false
	}
}

func (e *constraintViolationError) isEdgeDBIntegrityError() {}

func (e *constraintViolationError) isEdgeDBExecutionError() {}

func (e *constraintViolationError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}

type cardinalityViolationError struct {
	msg string
	err error
}

func (e *cardinalityViolationError) Error() string {
	msg := e.msg
	if e.err != nil {
		msg = e.err.Error()
	}

	return "gel.CardinalityViolationError: " + msg
}

func (e *cardinalityViolationError) Unwrap() error { return e.err }

func (e *cardinalityViolationError) Category(c ErrorCategory) bool {
	switch c {
	case CardinalityViolationError:
		return true
	case IntegrityError:
		return true
	case ExecutionError:
		return true
	default:
		return false
	}
}

func (e *cardinalityViolationError) isEdgeDBIntegrityError() {}

func (e *cardinalityViolationError) isEdgeDBExecutionError() {}

func (e *cardinalityViolationError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}

type missingRequiredError struct {
	msg string
	err error
}

func (e *missingRequiredError) Error() string {
	msg := e.msg
	if e.err != nil {
		msg = e.err.Error()
	}

	return "gel.MissingRequiredError: " + msg
}

func (e *missingRequiredError) Unwrap() error { return e.err }

func (e *missingRequiredError) Category(c ErrorCategory) bool {
	switch c {
	case MissingRequiredError:
		return true
	case IntegrityError:
		return true
	case ExecutionError:
		return true
	default:
		return false
	}
}

func (e *missingRequiredError) isEdgeDBIntegrityError() {}

func (e *missingRequiredError) isEdgeDBExecutionError() {}

func (e *missingRequiredError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}

type transactionError struct {
	msg string
	err error
}

func (e *transactionError) Error() string {
	msg := e.msg
	if e.err != nil {
		msg = e.err.Error()
	}

	return "gel.TransactionError: " + msg
}

func (e *transactionError) Unwrap() error { return e.err }

func (e *transactionError) Category(c ErrorCategory) bool {
	switch c {
	case TransactionError:
		return true
	case ExecutionError:
		return true
	default:
		return false
	}
}

func (e *transactionError) isEdgeDBExecutionError() {}

func (e *transactionError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}

type transactionConflictError struct {
	msg string
	err error
}

func (e *transactionConflictError) Error() string {
	msg := e.msg
	if e.err != nil {
		msg = e.err.Error()
	}

	return "gel.TransactionConflictError: " + msg
}

func (e *transactionConflictError) Unwrap() error { return e.err }

func (e *transactionConflictError) Category(c ErrorCategory) bool {
	switch c {
	case TransactionConflictError:
		return true
	case TransactionError:
		return true
	case ExecutionError:
		return true
	default:
		return false
	}
}

func (e *transactionConflictError) isEdgeDBTransactionError() {}

func (e *transactionConflictError) isEdgeDBExecutionError() {}

func (e *transactionConflictError) HasTag(tag ErrorTag) bool {
	switch tag {
	case ShouldRetry:
		return true
	default:
		return false
	}
}

type transactionSerializationError struct {
	msg string
	err error
}

func (e *transactionSerializationError) Error() string {
	msg := e.msg
	if e.err != nil {
		msg = e.err.Error()
	}

	return "gel.TransactionSerializationError: " + msg
}

func (e *transactionSerializationError) Unwrap() error { return e.err }

func (e *transactionSerializationError) Category(c ErrorCategory) bool {
	switch c {
	case TransactionSerializationError:
		return true
	case TransactionConflictError:
		return true
	case TransactionError:
		return true
	case ExecutionError:
		return true
	default:
		return false
	}
}

func (e *transactionSerializationError) isEdgeDBTransactionConflictError() {}

func (e *transactionSerializationError) isEdgeDBTransactionError() {}

func (e *transactionSerializationError) isEdgeDBExecutionError() {}

func (e *transactionSerializationError) HasTag(tag ErrorTag) bool {
	switch tag {
	case ShouldRetry:
		return true
	default:
		return false
	}
}

type transactionDeadlockError struct {
	msg string
	err error
}

func (e *transactionDeadlockError) Error() string {
	msg := e.msg
	if e.err != nil {
		msg = e.err.Error()
	}

	return "gel.TransactionDeadlockError: " + msg
}

func (e *transactionDeadlockError) Unwrap() error { return e.err }

func (e *transactionDeadlockError) Category(c ErrorCategory) bool {
	switch c {
	case TransactionDeadlockError:
		return true
	case TransactionConflictError:
		return true
	case TransactionError:
		return true
	case ExecutionError:
		return true
	default:
		return false
	}
}

func (e *transactionDeadlockError) isEdgeDBTransactionConflictError() {}

func (e *transactionDeadlockError) isEdgeDBTransactionError() {}

func (e *transactionDeadlockError) isEdgeDBExecutionError() {}

func (e *transactionDeadlockError) HasTag(tag ErrorTag) bool {
	switch tag {
	case ShouldRetry:
		return true
	default:
		return false
	}
}

type watchError struct {
	msg string
	err error
}

func (e *watchError) Error() string {
	msg := e.msg
	if e.err != nil {
		msg = e.err.Error()
	}

	return "gel.WatchError: " + msg
}

func (e *watchError) Unwrap() error { return e.err }

func (e *watchError) Category(c ErrorCategory) bool {
	switch c {
	case WatchError:
		return true
	case ExecutionError:
		return true
	default:
		return false
	}
}

func (e *watchError) isEdgeDBExecutionError() {}

func (e *watchError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}

type configurationError struct {
	msg string
	err error
}

func (e *configurationError) Error() string {
	msg := e.msg
	if e.err != nil {
		msg = e.err.Error()
	}

	return "gel.ConfigurationError: " + msg
}

func (e *configurationError) Unwrap() error { return e.err }

func (e *configurationError) Category(c ErrorCategory) bool {
	switch c {
	case ConfigurationError:
		return true
	default:
		return false
	}
}

func (e *configurationError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}

type accessError struct {
	msg string
	err error
}

func (e *accessError) Error() string {
	msg := e.msg
	if e.err != nil {
		msg = e.err.Error()
	}

	return "gel.AccessError: " + msg
}

func (e *accessError) Unwrap() error { return e.err }

func (e *accessError) Category(c ErrorCategory) bool {
	switch c {
	case AccessError:
		return true
	default:
		return false
	}
}

func (e *accessError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}

type authenticationError struct {
	msg string
	err error
}

func (e *authenticationError) Error() string {
	msg := e.msg
	if e.err != nil {
		msg = e.err.Error()
	}

	return "gel.AuthenticationError: " + msg
}

func (e *authenticationError) Unwrap() error { return e.err }

func (e *authenticationError) Category(c ErrorCategory) bool {
	switch c {
	case AuthenticationError:
		return true
	case AccessError:
		return true
	default:
		return false
	}
}

func (e *authenticationError) isEdgeDBAccessError() {}

func (e *authenticationError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}

type availabilityError struct {
	msg string
	err error
}

func (e *availabilityError) Error() string {
	msg := e.msg
	if e.err != nil {
		msg = e.err.Error()
	}

	return "gel.AvailabilityError: " + msg
}

func (e *availabilityError) Unwrap() error { return e.err }

func (e *availabilityError) Category(c ErrorCategory) bool {
	switch c {
	case AvailabilityError:
		return true
	default:
		return false
	}
}

func (e *availabilityError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}

type backendUnavailableError struct {
	msg string
	err error
}

func (e *backendUnavailableError) Error() string {
	msg := e.msg
	if e.err != nil {
		msg = e.err.Error()
	}

	return "gel.BackendUnavailableError: " + msg
}

func (e *backendUnavailableError) Unwrap() error { return e.err }

func (e *backendUnavailableError) Category(c ErrorCategory) bool {
	switch c {
	case BackendUnavailableError:
		return true
	case AvailabilityError:
		return true
	default:
		return false
	}
}

func (e *backendUnavailableError) isEdgeDBAvailabilityError() {}

func (e *backendUnavailableError) HasTag(tag ErrorTag) bool {
	switch tag {
	case ShouldRetry:
		return true
	default:
		return false
	}
}

type serverOfflineError struct {
	msg string
	err error
}

func (e *serverOfflineError) Error() string {
	msg := e.msg
	if e.err != nil {
		msg = e.err.Error()
	}

	return "gel.ServerOfflineError: " + msg
}

func (e *serverOfflineError) Unwrap() error { return e.err }

func (e *serverOfflineError) Category(c ErrorCategory) bool {
	switch c {
	case ServerOfflineError:
		return true
	case AvailabilityError:
		return true
	default:
		return false
	}
}

func (e *serverOfflineError) isEdgeDBAvailabilityError() {}

func (e *serverOfflineError) HasTag(tag ErrorTag) bool {
	switch tag {
	case ShouldReconnect:
		return true
	case ShouldRetry:
		return true
	default:
		return false
	}
}

type unknownTenantError struct {
	msg string
	err error
}

func (e *unknownTenantError) Error() string {
	msg := e.msg
	if e.err != nil {
		msg = e.err.Error()
	}

	return "gel.UnknownTenantError: " + msg
}

func (e *unknownTenantError) Unwrap() error { return e.err }

func (e *unknownTenantError) Category(c ErrorCategory) bool {
	switch c {
	case UnknownTenantError:
		return true
	case AvailabilityError:
		return true
	default:
		return false
	}
}

func (e *unknownTenantError) isEdgeDBAvailabilityError() {}

func (e *unknownTenantError) HasTag(tag ErrorTag) bool {
	switch tag {
	case ShouldReconnect:
		return true
	case ShouldRetry:
		return true
	default:
		return false
	}
}

type serverBlockedError struct {
	msg string
	err error
}

func (e *serverBlockedError) Error() string {
	msg := e.msg
	if e.err != nil {
		msg = e.err.Error()
	}

	return "gel.ServerBlockedError: " + msg
}

func (e *serverBlockedError) Unwrap() error { return e.err }

func (e *serverBlockedError) Category(c ErrorCategory) bool {
	switch c {
	case ServerBlockedError:
		return true
	case AvailabilityError:
		return true
	default:
		return false
	}
}

func (e *serverBlockedError) isEdgeDBAvailabilityError() {}

func (e *serverBlockedError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}

type backendError struct {
	msg string
	err error
}

func (e *backendError) Error() string {
	msg := e.msg
	if e.err != nil {
		msg = e.err.Error()
	}

	return "gel.BackendError: " + msg
}

func (e *backendError) Unwrap() error { return e.err }

func (e *backendError) Category(c ErrorCategory) bool {
	switch c {
	case BackendError:
		return true
	default:
		return false
	}
}

func (e *backendError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}

type unsupportedBackendFeatureError struct {
	msg string
	err error
}

func (e *unsupportedBackendFeatureError) Error() string {
	msg := e.msg
	if e.err != nil {
		msg = e.err.Error()
	}

	return "gel.UnsupportedBackendFeatureError: " + msg
}

func (e *unsupportedBackendFeatureError) Unwrap() error { return e.err }

func (e *unsupportedBackendFeatureError) Category(c ErrorCategory) bool {
	switch c {
	case UnsupportedBackendFeatureError:
		return true
	case BackendError:
		return true
	default:
		return false
	}
}

func (e *unsupportedBackendFeatureError) isEdgeDBBackendError() {}

func (e *unsupportedBackendFeatureError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}

type clientError struct {
	msg string
	err error
}

func (e *clientError) Error() string {
	msg := e.msg
	if e.err != nil {
		msg = e.err.Error()
	}

	return "gel.ClientError: " + msg
}

func (e *clientError) Unwrap() error { return e.err }

func (e *clientError) Category(c ErrorCategory) bool {
	switch c {
	case ClientError:
		return true
	default:
		return false
	}
}

func (e *clientError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}

type clientConnectionError struct {
	msg string
	err error
}

func (e *clientConnectionError) Error() string {
	msg := e.msg
	if e.err != nil {
		msg = e.err.Error()
	}

	return "gel.ClientConnectionError: " + msg
}

func (e *clientConnectionError) Unwrap() error { return e.err }

func (e *clientConnectionError) Category(c ErrorCategory) bool {
	switch c {
	case ClientConnectionError:
		return true
	case ClientError:
		return true
	default:
		return false
	}
}

func (e *clientConnectionError) isEdgeDBClientError() {}

func (e *clientConnectionError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}

type clientConnectionFailedError struct {
	msg string
	err error
}

func (e *clientConnectionFailedError) Error() string {
	msg := e.msg
	if e.err != nil {
		msg = e.err.Error()
	}

	return "gel.ClientConnectionFailedError: " + msg
}

func (e *clientConnectionFailedError) Unwrap() error { return e.err }

func (e *clientConnectionFailedError) Category(c ErrorCategory) bool {
	switch c {
	case ClientConnectionFailedError:
		return true
	case ClientConnectionError:
		return true
	case ClientError:
		return true
	default:
		return false
	}
}

func (e *clientConnectionFailedError) isEdgeDBClientConnectionError() {}

func (e *clientConnectionFailedError) isEdgeDBClientError() {}

func (e *clientConnectionFailedError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}

type clientConnectionFailedTemporarilyError struct {
	msg string
	err error
}

func (e *clientConnectionFailedTemporarilyError) Error() string {
	msg := e.msg
	if e.err != nil {
		msg = e.err.Error()
	}

	return "gel.ClientConnectionFailedTemporarilyError: " + msg
}

func (e *clientConnectionFailedTemporarilyError) Unwrap() error { return e.err }

func (e *clientConnectionFailedTemporarilyError) Category(c ErrorCategory) bool {
	switch c {
	case ClientConnectionFailedTemporarilyError:
		return true
	case ClientConnectionFailedError:
		return true
	case ClientConnectionError:
		return true
	case ClientError:
		return true
	default:
		return false
	}
}

func (e *clientConnectionFailedTemporarilyError) isEdgeDBClientConnectionFailedError() {}

func (e *clientConnectionFailedTemporarilyError) isEdgeDBClientConnectionError() {}

func (e *clientConnectionFailedTemporarilyError) isEdgeDBClientError() {}

func (e *clientConnectionFailedTemporarilyError) HasTag(tag ErrorTag) bool {
	switch tag {
	case ShouldReconnect:
		return true
	case ShouldRetry:
		return true
	default:
		return false
	}
}

type clientConnectionTimeoutError struct {
	msg string
	err error
}

func (e *clientConnectionTimeoutError) Error() string {
	msg := e.msg
	if e.err != nil {
		msg = e.err.Error()
	}

	return "gel.ClientConnectionTimeoutError: " + msg
}

func (e *clientConnectionTimeoutError) Unwrap() error { return e.err }

func (e *clientConnectionTimeoutError) Category(c ErrorCategory) bool {
	switch c {
	case ClientConnectionTimeoutError:
		return true
	case ClientConnectionError:
		return true
	case ClientError:
		return true
	default:
		return false
	}
}

func (e *clientConnectionTimeoutError) isEdgeDBClientConnectionError() {}

func (e *clientConnectionTimeoutError) isEdgeDBClientError() {}

func (e *clientConnectionTimeoutError) HasTag(tag ErrorTag) bool {
	switch tag {
	case ShouldReconnect:
		return true
	case ShouldRetry:
		return true
	default:
		return false
	}
}

type clientConnectionClosedError struct {
	msg string
	err error
}

func (e *clientConnectionClosedError) Error() string {
	msg := e.msg
	if e.err != nil {
		msg = e.err.Error()
	}

	return "gel.ClientConnectionClosedError: " + msg
}

func (e *clientConnectionClosedError) Unwrap() error { return e.err }

func (e *clientConnectionClosedError) Category(c ErrorCategory) bool {
	switch c {
	case ClientConnectionClosedError:
		return true
	case ClientConnectionError:
		return true
	case ClientError:
		return true
	default:
		return false
	}
}

func (e *clientConnectionClosedError) isEdgeDBClientConnectionError() {}

func (e *clientConnectionClosedError) isEdgeDBClientError() {}

func (e *clientConnectionClosedError) HasTag(tag ErrorTag) bool {
	switch tag {
	case ShouldReconnect:
		return true
	case ShouldRetry:
		return true
	default:
		return false
	}
}

type interfaceError struct {
	msg string
	err error
}

func (e *interfaceError) Error() string {
	msg := e.msg
	if e.err != nil {
		msg = e.err.Error()
	}

	return "gel.InterfaceError: " + msg
}

func (e *interfaceError) Unwrap() error { return e.err }

func (e *interfaceError) Category(c ErrorCategory) bool {
	switch c {
	case InterfaceError:
		return true
	case ClientError:
		return true
	default:
		return false
	}
}

func (e *interfaceError) isEdgeDBClientError() {}

func (e *interfaceError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}

type queryArgumentError struct {
	msg string
	err error
}

func (e *queryArgumentError) Error() string {
	msg := e.msg
	if e.err != nil {
		msg = e.err.Error()
	}

	return "gel.QueryArgumentError: " + msg
}

func (e *queryArgumentError) Unwrap() error { return e.err }

func (e *queryArgumentError) Category(c ErrorCategory) bool {
	switch c {
	case QueryArgumentError:
		return true
	case InterfaceError:
		return true
	case ClientError:
		return true
	default:
		return false
	}
}

func (e *queryArgumentError) isEdgeDBInterfaceError() {}

func (e *queryArgumentError) isEdgeDBClientError() {}

func (e *queryArgumentError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}

type missingArgumentError struct {
	msg string
	err error
}

func (e *missingArgumentError) Error() string {
	msg := e.msg
	if e.err != nil {
		msg = e.err.Error()
	}

	return "gel.MissingArgumentError: " + msg
}

func (e *missingArgumentError) Unwrap() error { return e.err }

func (e *missingArgumentError) Category(c ErrorCategory) bool {
	switch c {
	case MissingArgumentError:
		return true
	case QueryArgumentError:
		return true
	case InterfaceError:
		return true
	case ClientError:
		return true
	default:
		return false
	}
}

func (e *missingArgumentError) isEdgeDBQueryArgumentError() {}

func (e *missingArgumentError) isEdgeDBInterfaceError() {}

func (e *missingArgumentError) isEdgeDBClientError() {}

func (e *missingArgumentError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}

type unknownArgumentError struct {
	msg string
	err error
}

func (e *unknownArgumentError) Error() string {
	msg := e.msg
	if e.err != nil {
		msg = e.err.Error()
	}

	return "gel.UnknownArgumentError: " + msg
}

func (e *unknownArgumentError) Unwrap() error { return e.err }

func (e *unknownArgumentError) Category(c ErrorCategory) bool {
	switch c {
	case UnknownArgumentError:
		return true
	case QueryArgumentError:
		return true
	case InterfaceError:
		return true
	case ClientError:
		return true
	default:
		return false
	}
}

func (e *unknownArgumentError) isEdgeDBQueryArgumentError() {}

func (e *unknownArgumentError) isEdgeDBInterfaceError() {}

func (e *unknownArgumentError) isEdgeDBClientError() {}

func (e *unknownArgumentError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}

type invalidArgumentError struct {
	msg string
	err error
}

func (e *invalidArgumentError) Error() string {
	msg := e.msg
	if e.err != nil {
		msg = e.err.Error()
	}

	return "gel.InvalidArgumentError: " + msg
}

func (e *invalidArgumentError) Unwrap() error { return e.err }

func (e *invalidArgumentError) Category(c ErrorCategory) bool {
	switch c {
	case InvalidArgumentError:
		return true
	case QueryArgumentError:
		return true
	case InterfaceError:
		return true
	case ClientError:
		return true
	default:
		return false
	}
}

func (e *invalidArgumentError) isEdgeDBQueryArgumentError() {}

func (e *invalidArgumentError) isEdgeDBInterfaceError() {}

func (e *invalidArgumentError) isEdgeDBClientError() {}

func (e *invalidArgumentError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}

type noDataError struct {
	msg string
	err error
}

func (e *noDataError) Error() string {
	msg := e.msg
	if e.err != nil {
		msg = e.err.Error()
	}

	return "gel.NoDataError: " + msg
}

func (e *noDataError) Unwrap() error { return e.err }

func (e *noDataError) Category(c ErrorCategory) bool {
	switch c {
	case NoDataError:
		return true
	case ClientError:
		return true
	default:
		return false
	}
}

func (e *noDataError) isEdgeDBClientError() {}

func (e *noDataError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}

type internalClientError struct {
	msg string
	err error
}

func (e *internalClientError) Error() string {
	msg := e.msg
	if e.err != nil {
		msg = e.err.Error()
	}

	return "gel.InternalClientError: " + msg
}

func (e *internalClientError) Unwrap() error { return e.err }

func (e *internalClientError) Category(c ErrorCategory) bool {
	switch c {
	case InternalClientError:
		return true
	case ClientError:
		return true
	default:
		return false
	}
}

func (e *internalClientError) isEdgeDBClientError() {}

func (e *internalClientError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}

func errorFromCode(code uint32, msg string) error {
	switch code {
	case 0x01_00_00_00:
		return &internalServerError{msg: msg}
	case 0x02_00_00_00:
		return &unsupportedFeatureError{msg: msg}
	case 0x03_00_00_00:
		return &protocolError{msg: msg}
	case 0x03_01_00_00:
		return &binaryProtocolError{msg: msg}
	case 0x03_01_00_01:
		return &unsupportedProtocolVersionError{msg: msg}
	case 0x03_01_00_02:
		return &typeSpecNotFoundError{msg: msg}
	case 0x03_01_00_03:
		return &unexpectedMessageError{msg: msg}
	case 0x03_02_00_00:
		return &inputDataError{msg: msg}
	case 0x03_02_01_00:
		return &parameterTypeMismatchError{msg: msg}
	case 0x03_02_02_00:
		return &stateMismatchError{msg: msg}
	case 0x03_03_00_00:
		return &resultCardinalityMismatchError{msg: msg}
	case 0x03_04_00_00:
		return &capabilityError{msg: msg}
	case 0x03_04_01_00:
		return &unsupportedCapabilityError{msg: msg}
	case 0x03_04_02_00:
		return &disabledCapabilityError{msg: msg}
	case 0x04_00_00_00:
		return &queryError{msg: msg}
	case 0x04_01_00_00:
		return &invalidSyntaxError{msg: msg}
	case 0x04_01_01_00:
		return &edgeQLSyntaxError{msg: msg}
	case 0x04_01_02_00:
		return &schemaSyntaxError{msg: msg}
	case 0x04_01_03_00:
		return &graphQLSyntaxError{msg: msg}
	case 0x04_02_00_00:
		return &invalidTypeError{msg: msg}
	case 0x04_02_01_00:
		return &invalidTargetError{msg: msg}
	case 0x04_02_01_01:
		return &invalidLinkTargetError{msg: msg}
	case 0x04_02_01_02:
		return &invalidPropertyTargetError{msg: msg}
	case 0x04_03_00_00:
		return &invalidReferenceError{msg: msg}
	case 0x04_03_00_01:
		return &unknownModuleError{msg: msg}
	case 0x04_03_00_02:
		return &unknownLinkError{msg: msg}
	case 0x04_03_00_03:
		return &unknownPropertyError{msg: msg}
	case 0x04_03_00_04:
		return &unknownUserError{msg: msg}
	case 0x04_03_00_05:
		return &unknownDatabaseError{msg: msg}
	case 0x04_03_00_06:
		return &unknownParameterError{msg: msg}
	case 0x04_03_00_07:
		return &deprecatedScopingError{msg: msg}
	case 0x04_04_00_00:
		return &schemaError{msg: msg}
	case 0x04_05_00_00:
		return &schemaDefinitionError{msg: msg}
	case 0x04_05_01_00:
		return &invalidDefinitionError{msg: msg}
	case 0x04_05_01_01:
		return &invalidModuleDefinitionError{msg: msg}
	case 0x04_05_01_02:
		return &invalidLinkDefinitionError{msg: msg}
	case 0x04_05_01_03:
		return &invalidPropertyDefinitionError{msg: msg}
	case 0x04_05_01_04:
		return &invalidUserDefinitionError{msg: msg}
	case 0x04_05_01_05:
		return &invalidDatabaseDefinitionError{msg: msg}
	case 0x04_05_01_06:
		return &invalidOperatorDefinitionError{msg: msg}
	case 0x04_05_01_07:
		return &invalidAliasDefinitionError{msg: msg}
	case 0x04_05_01_08:
		return &invalidFunctionDefinitionError{msg: msg}
	case 0x04_05_01_09:
		return &invalidConstraintDefinitionError{msg: msg}
	case 0x04_05_01_0a:
		return &invalidCastDefinitionError{msg: msg}
	case 0x04_05_02_00:
		return &duplicateDefinitionError{msg: msg}
	case 0x04_05_02_01:
		return &duplicateModuleDefinitionError{msg: msg}
	case 0x04_05_02_02:
		return &duplicateLinkDefinitionError{msg: msg}
	case 0x04_05_02_03:
		return &duplicatePropertyDefinitionError{msg: msg}
	case 0x04_05_02_04:
		return &duplicateUserDefinitionError{msg: msg}
	case 0x04_05_02_05:
		return &duplicateDatabaseDefinitionError{msg: msg}
	case 0x04_05_02_06:
		return &duplicateOperatorDefinitionError{msg: msg}
	case 0x04_05_02_07:
		return &duplicateViewDefinitionError{msg: msg}
	case 0x04_05_02_08:
		return &duplicateFunctionDefinitionError{msg: msg}
	case 0x04_05_02_09:
		return &duplicateConstraintDefinitionError{msg: msg}
	case 0x04_05_02_0a:
		return &duplicateCastDefinitionError{msg: msg}
	case 0x04_05_02_0b:
		return &duplicateMigrationError{msg: msg}
	case 0x04_06_00_00:
		return &sessionTimeoutError{msg: msg}
	case 0x04_06_01_00:
		return &idleSessionTimeoutError{msg: msg}
	case 0x04_06_02_00:
		return &queryTimeoutError{msg: msg}
	case 0x04_06_0a_00:
		return &transactionTimeoutError{msg: msg}
	case 0x04_06_0a_01:
		return &idleTransactionTimeoutError{msg: msg}
	case 0x05_00_00_00:
		return &executionError{msg: msg}
	case 0x05_01_00_00:
		return &invalidValueError{msg: msg}
	case 0x05_01_00_01:
		return &divisionByZeroError{msg: msg}
	case 0x05_01_00_02:
		return &numericOutOfRangeError{msg: msg}
	case 0x05_01_00_03:
		return &accessPolicyError{msg: msg}
	case 0x05_01_00_04:
		return &queryAssertionError{msg: msg}
	case 0x05_02_00_00:
		return &integrityError{msg: msg}
	case 0x05_02_00_01:
		return &constraintViolationError{msg: msg}
	case 0x05_02_00_02:
		return &cardinalityViolationError{msg: msg}
	case 0x05_02_00_03:
		return &missingRequiredError{msg: msg}
	case 0x05_03_00_00:
		return &transactionError{msg: msg}
	case 0x05_03_01_00:
		return &transactionConflictError{msg: msg}
	case 0x05_03_01_01:
		return &transactionSerializationError{msg: msg}
	case 0x05_03_01_02:
		return &transactionDeadlockError{msg: msg}
	case 0x05_04_00_00:
		return &watchError{msg: msg}
	case 0x06_00_00_00:
		return &configurationError{msg: msg}
	case 0x07_00_00_00:
		return &accessError{msg: msg}
	case 0x07_01_00_00:
		return &authenticationError{msg: msg}
	case 0x08_00_00_00:
		return &availabilityError{msg: msg}
	case 0x08_00_00_01:
		return &backendUnavailableError{msg: msg}
	case 0x08_00_00_02:
		return &serverOfflineError{msg: msg}
	case 0x08_00_00_03:
		return &unknownTenantError{msg: msg}
	case 0x08_00_00_04:
		return &serverBlockedError{msg: msg}
	case 0x09_00_00_00:
		return &backendError{msg: msg}
	case 0x09_00_01_00:
		return &unsupportedBackendFeatureError{msg: msg}
	case 0xff_00_00_00:
		return &clientError{msg: msg}
	case 0xff_01_00_00:
		return &clientConnectionError{msg: msg}
	case 0xff_01_01_00:
		return &clientConnectionFailedError{msg: msg}
	case 0xff_01_01_01:
		return &clientConnectionFailedTemporarilyError{msg: msg}
	case 0xff_01_02_00:
		return &clientConnectionTimeoutError{msg: msg}
	case 0xff_01_03_00:
		return &clientConnectionClosedError{msg: msg}
	case 0xff_02_00_00:
		return &interfaceError{msg: msg}
	case 0xff_02_01_00:
		return &queryArgumentError{msg: msg}
	case 0xff_02_01_01:
		return &missingArgumentError{msg: msg}
	case 0xff_02_01_02:
		return &unknownArgumentError{msg: msg}
	case 0xff_02_01_03:
		return &invalidArgumentError{msg: msg}
	case 0xff_03_00_00:
		return &noDataError{msg: msg}
	case 0xff_04_00_00:
		return &internalClientError{msg: msg}
	default:
		return &unexpectedMessageError{
			msg: fmt.Sprintf(
				"invalid error code 0x%x with message %q", code, msg,
			),
		}
	}
}
