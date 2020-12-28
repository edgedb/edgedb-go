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

// This file is auto generated. Do not edit!

package edgedb

import "fmt"

const (
	internalServerErrorCode uint32 = 0x01_00_00_00
	unsupportedFeatureErrorCode uint32 = 0x02_00_00_00
	protocolErrorCode uint32 = 0x03_00_00_00
	binaryProtocolErrorCode uint32 = 0x03_01_00_00
	unsupportedProtocolVersionErrorCode uint32 = 0x03_01_00_01
	typeSpecNotFoundErrorCode uint32 = 0x03_01_00_02
	unexpectedMessageErrorCode uint32 = 0x03_01_00_03
	inputDataErrorCode uint32 = 0x03_02_00_00
	resultCardinalityMismatchErrorCode uint32 = 0x03_03_00_00
	queryErrorCode uint32 = 0x04_00_00_00
	invalidSyntaxErrorCode uint32 = 0x04_01_00_00
	edgeQLSyntaxErrorCode uint32 = 0x04_01_01_00
	schemaSyntaxErrorCode uint32 = 0x04_01_02_00
	graphQLSyntaxErrorCode uint32 = 0x04_01_03_00
	invalidTypeErrorCode uint32 = 0x04_02_00_00
	invalidTargetErrorCode uint32 = 0x04_02_01_00
	invalidLinkTargetErrorCode uint32 = 0x04_02_01_01
	invalidPropertyTargetErrorCode uint32 = 0x04_02_01_02
	invalidReferenceErrorCode uint32 = 0x04_03_00_00
	unknownModuleErrorCode uint32 = 0x04_03_00_01
	unknownLinkErrorCode uint32 = 0x04_03_00_02
	unknownPropertyErrorCode uint32 = 0x04_03_00_03
	unknownUserErrorCode uint32 = 0x04_03_00_04
	unknownDatabaseErrorCode uint32 = 0x04_03_00_05
	unknownParameterErrorCode uint32 = 0x04_03_00_06
	schemaErrorCode uint32 = 0x04_04_00_00
	schemaDefinitionErrorCode uint32 = 0x04_05_00_00
	invalidDefinitionErrorCode uint32 = 0x04_05_01_00
	invalidModuleDefinitionErrorCode uint32 = 0x04_05_01_01
	invalidLinkDefinitionErrorCode uint32 = 0x04_05_01_02
	invalidPropertyDefinitionErrorCode uint32 = 0x04_05_01_03
	invalidUserDefinitionErrorCode uint32 = 0x04_05_01_04
	invalidDatabaseDefinitionErrorCode uint32 = 0x04_05_01_05
	invalidOperatorDefinitionErrorCode uint32 = 0x04_05_01_06
	invalidAliasDefinitionErrorCode uint32 = 0x04_05_01_07
	invalidFunctionDefinitionErrorCode uint32 = 0x04_05_01_08
	invalidConstraintDefinitionErrorCode uint32 = 0x04_05_01_09
	invalidCastDefinitionErrorCode uint32 = 0x04_05_01_0a
	duplicateDefinitionErrorCode uint32 = 0x04_05_02_00
	duplicateModuleDefinitionErrorCode uint32 = 0x04_05_02_01
	duplicateLinkDefinitionErrorCode uint32 = 0x04_05_02_02
	duplicatePropertyDefinitionErrorCode uint32 = 0x04_05_02_03
	duplicateUserDefinitionErrorCode uint32 = 0x04_05_02_04
	duplicateDatabaseDefinitionErrorCode uint32 = 0x04_05_02_05
	duplicateOperatorDefinitionErrorCode uint32 = 0x04_05_02_06
	duplicateViewDefinitionErrorCode uint32 = 0x04_05_02_07
	duplicateFunctionDefinitionErrorCode uint32 = 0x04_05_02_08
	duplicateConstraintDefinitionErrorCode uint32 = 0x04_05_02_09
	duplicateCastDefinitionErrorCode uint32 = 0x04_05_02_0a
	queryTimeoutErrorCode uint32 = 0x04_06_00_00
	executionErrorCode uint32 = 0x05_00_00_00
	invalidValueErrorCode uint32 = 0x05_01_00_00
	divisionByZeroErrorCode uint32 = 0x05_01_00_01
	numericOutOfRangeErrorCode uint32 = 0x05_01_00_02
	integrityErrorCode uint32 = 0x05_02_00_00
	constraintViolationErrorCode uint32 = 0x05_02_00_01
	cardinalityViolationErrorCode uint32 = 0x05_02_00_02
	missingRequiredErrorCode uint32 = 0x05_02_00_03
	transactionErrorCode uint32 = 0x05_03_00_00
	transactionSerializationErrorCode uint32 = 0x05_03_00_01
	transactionDeadlockErrorCode uint32 = 0x05_03_00_02
	configurationErrorCode uint32 = 0x06_00_00_00
	accessErrorCode uint32 = 0x07_00_00_00
	authenticationErrorCode uint32 = 0x07_01_00_00
	clientErrorCode uint32 = 0xff_00_00_00
	clientConnectionErrorCode uint32 = 0xff_01_00_00
	interfaceErrorCode uint32 = 0xff_02_00_00
	queryArgumentErrorCode uint32 = 0xff_02_01_00
	missingArgumentErrorCode uint32 = 0xff_02_01_01
	unknownArgumentErrorCode uint32 = 0xff_02_01_02
	invalidArgumentErrorCode uint32 = 0xff_02_01_03
	noDataErrorCode uint32 = 0xff_03_00_00
)

// newErrorFromCode returns a new edgedb error.
func newErrorFromCode(code uint32, msg string) error {
	switch code {
	case internalServerErrorCode:
		tail := &baseError{msg: "edgedb: " + msg}
		base := &baseError{err: &Error{tail}}
		return &InternalServerError{base}
	case unsupportedFeatureErrorCode:
		tail := &baseError{msg: "edgedb: " + msg}
		base := &baseError{err: &Error{tail}}
		return &UnsupportedFeatureError{base}
	case protocolErrorCode:
		tail := &baseError{msg: "edgedb: " + msg}
		base := &baseError{err: &Error{tail}}
		return &ProtocolError{base}
	case binaryProtocolErrorCode:
		base := &baseError{err: newErrorFromCode(protocolErrorCode, msg)}
		return &BinaryProtocolError{base}
	case unsupportedProtocolVersionErrorCode:
		base := &baseError{err: newErrorFromCode(binaryProtocolErrorCode, msg)}
		return &UnsupportedProtocolVersionError{base}
	case typeSpecNotFoundErrorCode:
		base := &baseError{err: newErrorFromCode(binaryProtocolErrorCode, msg)}
		return &TypeSpecNotFoundError{base}
	case unexpectedMessageErrorCode:
		base := &baseError{err: newErrorFromCode(binaryProtocolErrorCode, msg)}
		return &UnexpectedMessageError{base}
	case inputDataErrorCode:
		base := &baseError{err: newErrorFromCode(protocolErrorCode, msg)}
		return &InputDataError{base}
	case resultCardinalityMismatchErrorCode:
		base := &baseError{err: newErrorFromCode(protocolErrorCode, msg)}
		return &ResultCardinalityMismatchError{base}
	case queryErrorCode:
		tail := &baseError{msg: "edgedb: " + msg}
		base := &baseError{err: &Error{tail}}
		return &QueryError{base}
	case invalidSyntaxErrorCode:
		base := &baseError{err: newErrorFromCode(queryErrorCode, msg)}
		return &InvalidSyntaxError{base}
	case edgeQLSyntaxErrorCode:
		base := &baseError{err: newErrorFromCode(invalidSyntaxErrorCode, msg)}
		return &EdgeQLSyntaxError{base}
	case schemaSyntaxErrorCode:
		base := &baseError{err: newErrorFromCode(invalidSyntaxErrorCode, msg)}
		return &SchemaSyntaxError{base}
	case graphQLSyntaxErrorCode:
		base := &baseError{err: newErrorFromCode(invalidSyntaxErrorCode, msg)}
		return &GraphQLSyntaxError{base}
	case invalidTypeErrorCode:
		base := &baseError{err: newErrorFromCode(queryErrorCode, msg)}
		return &InvalidTypeError{base}
	case invalidTargetErrorCode:
		base := &baseError{err: newErrorFromCode(invalidTypeErrorCode, msg)}
		return &InvalidTargetError{base}
	case invalidLinkTargetErrorCode:
		base := &baseError{err: newErrorFromCode(invalidTargetErrorCode, msg)}
		return &InvalidLinkTargetError{base}
	case invalidPropertyTargetErrorCode:
		base := &baseError{err: newErrorFromCode(invalidTargetErrorCode, msg)}
		return &InvalidPropertyTargetError{base}
	case invalidReferenceErrorCode:
		base := &baseError{err: newErrorFromCode(queryErrorCode, msg)}
		return &InvalidReferenceError{base}
	case unknownModuleErrorCode:
		base := &baseError{err: newErrorFromCode(invalidReferenceErrorCode, msg)}
		return &UnknownModuleError{base}
	case unknownLinkErrorCode:
		base := &baseError{err: newErrorFromCode(invalidReferenceErrorCode, msg)}
		return &UnknownLinkError{base}
	case unknownPropertyErrorCode:
		base := &baseError{err: newErrorFromCode(invalidReferenceErrorCode, msg)}
		return &UnknownPropertyError{base}
	case unknownUserErrorCode:
		base := &baseError{err: newErrorFromCode(invalidReferenceErrorCode, msg)}
		return &UnknownUserError{base}
	case unknownDatabaseErrorCode:
		base := &baseError{err: newErrorFromCode(invalidReferenceErrorCode, msg)}
		return &UnknownDatabaseError{base}
	case unknownParameterErrorCode:
		base := &baseError{err: newErrorFromCode(invalidReferenceErrorCode, msg)}
		return &UnknownParameterError{base}
	case schemaErrorCode:
		base := &baseError{err: newErrorFromCode(queryErrorCode, msg)}
		return &SchemaError{base}
	case schemaDefinitionErrorCode:
		base := &baseError{err: newErrorFromCode(queryErrorCode, msg)}
		return &SchemaDefinitionError{base}
	case invalidDefinitionErrorCode:
		base := &baseError{err: newErrorFromCode(schemaDefinitionErrorCode, msg)}
		return &InvalidDefinitionError{base}
	case invalidModuleDefinitionErrorCode:
		base := &baseError{err: newErrorFromCode(invalidDefinitionErrorCode, msg)}
		return &InvalidModuleDefinitionError{base}
	case invalidLinkDefinitionErrorCode:
		base := &baseError{err: newErrorFromCode(invalidDefinitionErrorCode, msg)}
		return &InvalidLinkDefinitionError{base}
	case invalidPropertyDefinitionErrorCode:
		base := &baseError{err: newErrorFromCode(invalidDefinitionErrorCode, msg)}
		return &InvalidPropertyDefinitionError{base}
	case invalidUserDefinitionErrorCode:
		base := &baseError{err: newErrorFromCode(invalidDefinitionErrorCode, msg)}
		return &InvalidUserDefinitionError{base}
	case invalidDatabaseDefinitionErrorCode:
		base := &baseError{err: newErrorFromCode(invalidDefinitionErrorCode, msg)}
		return &InvalidDatabaseDefinitionError{base}
	case invalidOperatorDefinitionErrorCode:
		base := &baseError{err: newErrorFromCode(invalidDefinitionErrorCode, msg)}
		return &InvalidOperatorDefinitionError{base}
	case invalidAliasDefinitionErrorCode:
		base := &baseError{err: newErrorFromCode(invalidDefinitionErrorCode, msg)}
		return &InvalidAliasDefinitionError{base}
	case invalidFunctionDefinitionErrorCode:
		base := &baseError{err: newErrorFromCode(invalidDefinitionErrorCode, msg)}
		return &InvalidFunctionDefinitionError{base}
	case invalidConstraintDefinitionErrorCode:
		base := &baseError{err: newErrorFromCode(invalidDefinitionErrorCode, msg)}
		return &InvalidConstraintDefinitionError{base}
	case invalidCastDefinitionErrorCode:
		base := &baseError{err: newErrorFromCode(invalidDefinitionErrorCode, msg)}
		return &InvalidCastDefinitionError{base}
	case duplicateDefinitionErrorCode:
		base := &baseError{err: newErrorFromCode(schemaDefinitionErrorCode, msg)}
		return &DuplicateDefinitionError{base}
	case duplicateModuleDefinitionErrorCode:
		base := &baseError{err: newErrorFromCode(duplicateDefinitionErrorCode, msg)}
		return &DuplicateModuleDefinitionError{base}
	case duplicateLinkDefinitionErrorCode:
		base := &baseError{err: newErrorFromCode(duplicateDefinitionErrorCode, msg)}
		return &DuplicateLinkDefinitionError{base}
	case duplicatePropertyDefinitionErrorCode:
		base := &baseError{err: newErrorFromCode(duplicateDefinitionErrorCode, msg)}
		return &DuplicatePropertyDefinitionError{base}
	case duplicateUserDefinitionErrorCode:
		base := &baseError{err: newErrorFromCode(duplicateDefinitionErrorCode, msg)}
		return &DuplicateUserDefinitionError{base}
	case duplicateDatabaseDefinitionErrorCode:
		base := &baseError{err: newErrorFromCode(duplicateDefinitionErrorCode, msg)}
		return &DuplicateDatabaseDefinitionError{base}
	case duplicateOperatorDefinitionErrorCode:
		base := &baseError{err: newErrorFromCode(duplicateDefinitionErrorCode, msg)}
		return &DuplicateOperatorDefinitionError{base}
	case duplicateViewDefinitionErrorCode:
		base := &baseError{err: newErrorFromCode(duplicateDefinitionErrorCode, msg)}
		return &DuplicateViewDefinitionError{base}
	case duplicateFunctionDefinitionErrorCode:
		base := &baseError{err: newErrorFromCode(duplicateDefinitionErrorCode, msg)}
		return &DuplicateFunctionDefinitionError{base}
	case duplicateConstraintDefinitionErrorCode:
		base := &baseError{err: newErrorFromCode(duplicateDefinitionErrorCode, msg)}
		return &DuplicateConstraintDefinitionError{base}
	case duplicateCastDefinitionErrorCode:
		base := &baseError{err: newErrorFromCode(duplicateDefinitionErrorCode, msg)}
		return &DuplicateCastDefinitionError{base}
	case queryTimeoutErrorCode:
		base := &baseError{err: newErrorFromCode(queryErrorCode, msg)}
		return &QueryTimeoutError{base}
	case executionErrorCode:
		tail := &baseError{msg: "edgedb: " + msg}
		base := &baseError{err: &Error{tail}}
		return &ExecutionError{base}
	case invalidValueErrorCode:
		base := &baseError{err: newErrorFromCode(executionErrorCode, msg)}
		return &InvalidValueError{base}
	case divisionByZeroErrorCode:
		base := &baseError{err: newErrorFromCode(invalidValueErrorCode, msg)}
		return &DivisionByZeroError{base}
	case numericOutOfRangeErrorCode:
		base := &baseError{err: newErrorFromCode(invalidValueErrorCode, msg)}
		return &NumericOutOfRangeError{base}
	case integrityErrorCode:
		base := &baseError{err: newErrorFromCode(executionErrorCode, msg)}
		return &IntegrityError{base}
	case constraintViolationErrorCode:
		base := &baseError{err: newErrorFromCode(integrityErrorCode, msg)}
		return &ConstraintViolationError{base}
	case cardinalityViolationErrorCode:
		base := &baseError{err: newErrorFromCode(integrityErrorCode, msg)}
		return &CardinalityViolationError{base}
	case missingRequiredErrorCode:
		base := &baseError{err: newErrorFromCode(integrityErrorCode, msg)}
		return &MissingRequiredError{base}
	case transactionErrorCode:
		base := &baseError{err: newErrorFromCode(executionErrorCode, msg)}
		return &TransactionError{base}
	case transactionSerializationErrorCode:
		base := &baseError{err: newErrorFromCode(transactionErrorCode, msg)}
		return &TransactionSerializationError{base}
	case transactionDeadlockErrorCode:
		base := &baseError{err: newErrorFromCode(transactionErrorCode, msg)}
		return &TransactionDeadlockError{base}
	case configurationErrorCode:
		tail := &baseError{msg: "edgedb: " + msg}
		base := &baseError{err: &Error{tail}}
		return &ConfigurationError{base}
	case accessErrorCode:
		tail := &baseError{msg: "edgedb: " + msg}
		base := &baseError{err: &Error{tail}}
		return &AccessError{base}
	case authenticationErrorCode:
		base := &baseError{err: newErrorFromCode(accessErrorCode, msg)}
		return &AuthenticationError{base}
	case clientErrorCode:
		tail := &baseError{msg: "edgedb: " + msg}
		base := &baseError{err: &Error{tail}}
		return &ClientError{base}
	case clientConnectionErrorCode:
		base := &baseError{err: newErrorFromCode(clientErrorCode, msg)}
		return &ClientConnectionError{base}
	case interfaceErrorCode:
		base := &baseError{err: newErrorFromCode(clientErrorCode, msg)}
		return &InterfaceError{base}
	case queryArgumentErrorCode:
		base := &baseError{err: newErrorFromCode(interfaceErrorCode, msg)}
		return &QueryArgumentError{base}
	case missingArgumentErrorCode:
		base := &baseError{err: newErrorFromCode(queryArgumentErrorCode, msg)}
		return &MissingArgumentError{base}
	case unknownArgumentErrorCode:
		base := &baseError{err: newErrorFromCode(queryArgumentErrorCode, msg)}
		return &UnknownArgumentError{base}
	case invalidArgumentErrorCode:
		base := &baseError{err: newErrorFromCode(queryArgumentErrorCode, msg)}
		return &InvalidArgumentError{base}
	case noDataErrorCode:
		base := &baseError{err: newErrorFromCode(clientErrorCode, msg)}
		return &NoDataError{base}
	default:
		panic(fmt.Sprintf("unknown error code: %v", code))
	}
}

// InternalServerError is an error.
type InternalServerError struct {
	*baseError
}

// UnsupportedFeatureError is an error.
type UnsupportedFeatureError struct {
	*baseError
}

// ProtocolError is an error.
type ProtocolError struct {
	*baseError
}

// BinaryProtocolError is an error.
type BinaryProtocolError struct {
	*baseError
}

// UnsupportedProtocolVersionError is an error.
type UnsupportedProtocolVersionError struct {
	*baseError
}

// TypeSpecNotFoundError is an error.
type TypeSpecNotFoundError struct {
	*baseError
}

// UnexpectedMessageError is an error.
type UnexpectedMessageError struct {
	*baseError
}

// InputDataError is an error.
type InputDataError struct {
	*baseError
}

// ResultCardinalityMismatchError is an error.
type ResultCardinalityMismatchError struct {
	*baseError
}

// QueryError is an error.
type QueryError struct {
	*baseError
}

// InvalidSyntaxError is an error.
type InvalidSyntaxError struct {
	*baseError
}

// EdgeQLSyntaxError is an error.
type EdgeQLSyntaxError struct {
	*baseError
}

// SchemaSyntaxError is an error.
type SchemaSyntaxError struct {
	*baseError
}

// GraphQLSyntaxError is an error.
type GraphQLSyntaxError struct {
	*baseError
}

// InvalidTypeError is an error.
type InvalidTypeError struct {
	*baseError
}

// InvalidTargetError is an error.
type InvalidTargetError struct {
	*baseError
}

// InvalidLinkTargetError is an error.
type InvalidLinkTargetError struct {
	*baseError
}

// InvalidPropertyTargetError is an error.
type InvalidPropertyTargetError struct {
	*baseError
}

// InvalidReferenceError is an error.
type InvalidReferenceError struct {
	*baseError
}

// UnknownModuleError is an error.
type UnknownModuleError struct {
	*baseError
}

// UnknownLinkError is an error.
type UnknownLinkError struct {
	*baseError
}

// UnknownPropertyError is an error.
type UnknownPropertyError struct {
	*baseError
}

// UnknownUserError is an error.
type UnknownUserError struct {
	*baseError
}

// UnknownDatabaseError is an error.
type UnknownDatabaseError struct {
	*baseError
}

// UnknownParameterError is an error.
type UnknownParameterError struct {
	*baseError
}

// SchemaError is an error.
type SchemaError struct {
	*baseError
}

// SchemaDefinitionError is an error.
type SchemaDefinitionError struct {
	*baseError
}

// InvalidDefinitionError is an error.
type InvalidDefinitionError struct {
	*baseError
}

// InvalidModuleDefinitionError is an error.
type InvalidModuleDefinitionError struct {
	*baseError
}

// InvalidLinkDefinitionError is an error.
type InvalidLinkDefinitionError struct {
	*baseError
}

// InvalidPropertyDefinitionError is an error.
type InvalidPropertyDefinitionError struct {
	*baseError
}

// InvalidUserDefinitionError is an error.
type InvalidUserDefinitionError struct {
	*baseError
}

// InvalidDatabaseDefinitionError is an error.
type InvalidDatabaseDefinitionError struct {
	*baseError
}

// InvalidOperatorDefinitionError is an error.
type InvalidOperatorDefinitionError struct {
	*baseError
}

// InvalidAliasDefinitionError is an error.
type InvalidAliasDefinitionError struct {
	*baseError
}

// InvalidFunctionDefinitionError is an error.
type InvalidFunctionDefinitionError struct {
	*baseError
}

// InvalidConstraintDefinitionError is an error.
type InvalidConstraintDefinitionError struct {
	*baseError
}

// InvalidCastDefinitionError is an error.
type InvalidCastDefinitionError struct {
	*baseError
}

// DuplicateDefinitionError is an error.
type DuplicateDefinitionError struct {
	*baseError
}

// DuplicateModuleDefinitionError is an error.
type DuplicateModuleDefinitionError struct {
	*baseError
}

// DuplicateLinkDefinitionError is an error.
type DuplicateLinkDefinitionError struct {
	*baseError
}

// DuplicatePropertyDefinitionError is an error.
type DuplicatePropertyDefinitionError struct {
	*baseError
}

// DuplicateUserDefinitionError is an error.
type DuplicateUserDefinitionError struct {
	*baseError
}

// DuplicateDatabaseDefinitionError is an error.
type DuplicateDatabaseDefinitionError struct {
	*baseError
}

// DuplicateOperatorDefinitionError is an error.
type DuplicateOperatorDefinitionError struct {
	*baseError
}

// DuplicateViewDefinitionError is an error.
type DuplicateViewDefinitionError struct {
	*baseError
}

// DuplicateFunctionDefinitionError is an error.
type DuplicateFunctionDefinitionError struct {
	*baseError
}

// DuplicateConstraintDefinitionError is an error.
type DuplicateConstraintDefinitionError struct {
	*baseError
}

// DuplicateCastDefinitionError is an error.
type DuplicateCastDefinitionError struct {
	*baseError
}

// QueryTimeoutError is an error.
type QueryTimeoutError struct {
	*baseError
}

// ExecutionError is an error.
type ExecutionError struct {
	*baseError
}

// InvalidValueError is an error.
type InvalidValueError struct {
	*baseError
}

// DivisionByZeroError is an error.
type DivisionByZeroError struct {
	*baseError
}

// NumericOutOfRangeError is an error.
type NumericOutOfRangeError struct {
	*baseError
}

// IntegrityError is an error.
type IntegrityError struct {
	*baseError
}

// ConstraintViolationError is an error.
type ConstraintViolationError struct {
	*baseError
}

// CardinalityViolationError is an error.
type CardinalityViolationError struct {
	*baseError
}

// MissingRequiredError is an error.
type MissingRequiredError struct {
	*baseError
}

// TransactionError is an error.
type TransactionError struct {
	*baseError
}

// TransactionSerializationError is an error.
type TransactionSerializationError struct {
	*baseError
}

// TransactionDeadlockError is an error.
type TransactionDeadlockError struct {
	*baseError
}

// ConfigurationError is an error.
type ConfigurationError struct {
	*baseError
}

// AccessError is an error.
type AccessError struct {
	*baseError
}

// AuthenticationError is an error.
type AuthenticationError struct {
	*baseError
}

// ClientError is an error.
type ClientError struct {
	*baseError
}

// ClientConnectionError is an error.
type ClientConnectionError struct {
	*baseError
}

// InterfaceError is an error.
type InterfaceError struct {
	*baseError
}

// QueryArgumentError is an error.
type QueryArgumentError struct {
	*baseError
}

// MissingArgumentError is an error.
type MissingArgumentError struct {
	*baseError
}

// UnknownArgumentError is an error.
type UnknownArgumentError struct {
	*baseError
}

// InvalidArgumentError is an error.
type InvalidArgumentError struct {
	*baseError
}

// NoDataError is an error.
type NoDataError struct {
	*baseError
}
