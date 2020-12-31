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

// wrapErrorFromCode wraps an error in an edgedb error type.
func wrapErrorFromCode(code uint32, err error) error {
	if err == nil {
		return nil
	}

	switch code {
	case internalServerErrorCode:
		return &InternalServerError{&baseError{err: wrapError(err)}}
	case unsupportedFeatureErrorCode:
		return &UnsupportedFeatureError{&baseError{err: wrapError(err)}}
	case protocolErrorCode:
		return &ProtocolError{&baseError{err: wrapError(err)}}
	case binaryProtocolErrorCode:
		next := wrapErrorFromCode(protocolErrorCode, err)
		return &BinaryProtocolError{&baseError{err: next}}
	case unsupportedProtocolVersionErrorCode:
		next := wrapErrorFromCode(binaryProtocolErrorCode, err)
		return &UnsupportedProtocolVersionError{&baseError{err: next}}
	case typeSpecNotFoundErrorCode:
		next := wrapErrorFromCode(binaryProtocolErrorCode, err)
		return &TypeSpecNotFoundError{&baseError{err: next}}
	case unexpectedMessageErrorCode:
		next := wrapErrorFromCode(binaryProtocolErrorCode, err)
		return &UnexpectedMessageError{&baseError{err: next}}
	case inputDataErrorCode:
		next := wrapErrorFromCode(protocolErrorCode, err)
		return &InputDataError{&baseError{err: next}}
	case resultCardinalityMismatchErrorCode:
		next := wrapErrorFromCode(protocolErrorCode, err)
		return &ResultCardinalityMismatchError{&baseError{err: next}}
	case queryErrorCode:
		return &QueryError{&baseError{err: wrapError(err)}}
	case invalidSyntaxErrorCode:
		next := wrapErrorFromCode(queryErrorCode, err)
		return &InvalidSyntaxError{&baseError{err: next}}
	case edgeQLSyntaxErrorCode:
		next := wrapErrorFromCode(invalidSyntaxErrorCode, err)
		return &EdgeQLSyntaxError{&baseError{err: next}}
	case schemaSyntaxErrorCode:
		next := wrapErrorFromCode(invalidSyntaxErrorCode, err)
		return &SchemaSyntaxError{&baseError{err: next}}
	case graphQLSyntaxErrorCode:
		next := wrapErrorFromCode(invalidSyntaxErrorCode, err)
		return &GraphQLSyntaxError{&baseError{err: next}}
	case invalidTypeErrorCode:
		next := wrapErrorFromCode(queryErrorCode, err)
		return &InvalidTypeError{&baseError{err: next}}
	case invalidTargetErrorCode:
		next := wrapErrorFromCode(invalidTypeErrorCode, err)
		return &InvalidTargetError{&baseError{err: next}}
	case invalidLinkTargetErrorCode:
		next := wrapErrorFromCode(invalidTargetErrorCode, err)
		return &InvalidLinkTargetError{&baseError{err: next}}
	case invalidPropertyTargetErrorCode:
		next := wrapErrorFromCode(invalidTargetErrorCode, err)
		return &InvalidPropertyTargetError{&baseError{err: next}}
	case invalidReferenceErrorCode:
		next := wrapErrorFromCode(queryErrorCode, err)
		return &InvalidReferenceError{&baseError{err: next}}
	case unknownModuleErrorCode:
		next := wrapErrorFromCode(invalidReferenceErrorCode, err)
		return &UnknownModuleError{&baseError{err: next}}
	case unknownLinkErrorCode:
		next := wrapErrorFromCode(invalidReferenceErrorCode, err)
		return &UnknownLinkError{&baseError{err: next}}
	case unknownPropertyErrorCode:
		next := wrapErrorFromCode(invalidReferenceErrorCode, err)
		return &UnknownPropertyError{&baseError{err: next}}
	case unknownUserErrorCode:
		next := wrapErrorFromCode(invalidReferenceErrorCode, err)
		return &UnknownUserError{&baseError{err: next}}
	case unknownDatabaseErrorCode:
		next := wrapErrorFromCode(invalidReferenceErrorCode, err)
		return &UnknownDatabaseError{&baseError{err: next}}
	case unknownParameterErrorCode:
		next := wrapErrorFromCode(invalidReferenceErrorCode, err)
		return &UnknownParameterError{&baseError{err: next}}
	case schemaErrorCode:
		next := wrapErrorFromCode(queryErrorCode, err)
		return &SchemaError{&baseError{err: next}}
	case schemaDefinitionErrorCode:
		next := wrapErrorFromCode(queryErrorCode, err)
		return &SchemaDefinitionError{&baseError{err: next}}
	case invalidDefinitionErrorCode:
		next := wrapErrorFromCode(schemaDefinitionErrorCode, err)
		return &InvalidDefinitionError{&baseError{err: next}}
	case invalidModuleDefinitionErrorCode:
		next := wrapErrorFromCode(invalidDefinitionErrorCode, err)
		return &InvalidModuleDefinitionError{&baseError{err: next}}
	case invalidLinkDefinitionErrorCode:
		next := wrapErrorFromCode(invalidDefinitionErrorCode, err)
		return &InvalidLinkDefinitionError{&baseError{err: next}}
	case invalidPropertyDefinitionErrorCode:
		next := wrapErrorFromCode(invalidDefinitionErrorCode, err)
		return &InvalidPropertyDefinitionError{&baseError{err: next}}
	case invalidUserDefinitionErrorCode:
		next := wrapErrorFromCode(invalidDefinitionErrorCode, err)
		return &InvalidUserDefinitionError{&baseError{err: next}}
	case invalidDatabaseDefinitionErrorCode:
		next := wrapErrorFromCode(invalidDefinitionErrorCode, err)
		return &InvalidDatabaseDefinitionError{&baseError{err: next}}
	case invalidOperatorDefinitionErrorCode:
		next := wrapErrorFromCode(invalidDefinitionErrorCode, err)
		return &InvalidOperatorDefinitionError{&baseError{err: next}}
	case invalidAliasDefinitionErrorCode:
		next := wrapErrorFromCode(invalidDefinitionErrorCode, err)
		return &InvalidAliasDefinitionError{&baseError{err: next}}
	case invalidFunctionDefinitionErrorCode:
		next := wrapErrorFromCode(invalidDefinitionErrorCode, err)
		return &InvalidFunctionDefinitionError{&baseError{err: next}}
	case invalidConstraintDefinitionErrorCode:
		next := wrapErrorFromCode(invalidDefinitionErrorCode, err)
		return &InvalidConstraintDefinitionError{&baseError{err: next}}
	case invalidCastDefinitionErrorCode:
		next := wrapErrorFromCode(invalidDefinitionErrorCode, err)
		return &InvalidCastDefinitionError{&baseError{err: next}}
	case duplicateDefinitionErrorCode:
		next := wrapErrorFromCode(schemaDefinitionErrorCode, err)
		return &DuplicateDefinitionError{&baseError{err: next}}
	case duplicateModuleDefinitionErrorCode:
		next := wrapErrorFromCode(duplicateDefinitionErrorCode, err)
		return &DuplicateModuleDefinitionError{&baseError{err: next}}
	case duplicateLinkDefinitionErrorCode:
		next := wrapErrorFromCode(duplicateDefinitionErrorCode, err)
		return &DuplicateLinkDefinitionError{&baseError{err: next}}
	case duplicatePropertyDefinitionErrorCode:
		next := wrapErrorFromCode(duplicateDefinitionErrorCode, err)
		return &DuplicatePropertyDefinitionError{&baseError{err: next}}
	case duplicateUserDefinitionErrorCode:
		next := wrapErrorFromCode(duplicateDefinitionErrorCode, err)
		return &DuplicateUserDefinitionError{&baseError{err: next}}
	case duplicateDatabaseDefinitionErrorCode:
		next := wrapErrorFromCode(duplicateDefinitionErrorCode, err)
		return &DuplicateDatabaseDefinitionError{&baseError{err: next}}
	case duplicateOperatorDefinitionErrorCode:
		next := wrapErrorFromCode(duplicateDefinitionErrorCode, err)
		return &DuplicateOperatorDefinitionError{&baseError{err: next}}
	case duplicateViewDefinitionErrorCode:
		next := wrapErrorFromCode(duplicateDefinitionErrorCode, err)
		return &DuplicateViewDefinitionError{&baseError{err: next}}
	case duplicateFunctionDefinitionErrorCode:
		next := wrapErrorFromCode(duplicateDefinitionErrorCode, err)
		return &DuplicateFunctionDefinitionError{&baseError{err: next}}
	case duplicateConstraintDefinitionErrorCode:
		next := wrapErrorFromCode(duplicateDefinitionErrorCode, err)
		return &DuplicateConstraintDefinitionError{&baseError{err: next}}
	case duplicateCastDefinitionErrorCode:
		next := wrapErrorFromCode(duplicateDefinitionErrorCode, err)
		return &DuplicateCastDefinitionError{&baseError{err: next}}
	case queryTimeoutErrorCode:
		next := wrapErrorFromCode(queryErrorCode, err)
		return &QueryTimeoutError{&baseError{err: next}}
	case executionErrorCode:
		return &ExecutionError{&baseError{err: wrapError(err)}}
	case invalidValueErrorCode:
		next := wrapErrorFromCode(executionErrorCode, err)
		return &InvalidValueError{&baseError{err: next}}
	case divisionByZeroErrorCode:
		next := wrapErrorFromCode(invalidValueErrorCode, err)
		return &DivisionByZeroError{&baseError{err: next}}
	case numericOutOfRangeErrorCode:
		next := wrapErrorFromCode(invalidValueErrorCode, err)
		return &NumericOutOfRangeError{&baseError{err: next}}
	case integrityErrorCode:
		next := wrapErrorFromCode(executionErrorCode, err)
		return &IntegrityError{&baseError{err: next}}
	case constraintViolationErrorCode:
		next := wrapErrorFromCode(integrityErrorCode, err)
		return &ConstraintViolationError{&baseError{err: next}}
	case cardinalityViolationErrorCode:
		next := wrapErrorFromCode(integrityErrorCode, err)
		return &CardinalityViolationError{&baseError{err: next}}
	case missingRequiredErrorCode:
		next := wrapErrorFromCode(integrityErrorCode, err)
		return &MissingRequiredError{&baseError{err: next}}
	case transactionErrorCode:
		next := wrapErrorFromCode(executionErrorCode, err)
		return &TransactionError{&baseError{err: next}}
	case transactionSerializationErrorCode:
		next := wrapErrorFromCode(transactionErrorCode, err)
		return &TransactionSerializationError{&baseError{err: next}}
	case transactionDeadlockErrorCode:
		next := wrapErrorFromCode(transactionErrorCode, err)
		return &TransactionDeadlockError{&baseError{err: next}}
	case configurationErrorCode:
		return &ConfigurationError{&baseError{err: wrapError(err)}}
	case accessErrorCode:
		return &AccessError{&baseError{err: wrapError(err)}}
	case authenticationErrorCode:
		next := wrapErrorFromCode(accessErrorCode, err)
		return &AuthenticationError{&baseError{err: next}}
	case clientErrorCode:
		return &ClientError{&baseError{err: wrapError(err)}}
	case clientConnectionErrorCode:
		next := wrapErrorFromCode(clientErrorCode, err)
		return &ClientConnectionError{&baseError{err: next}}
	case interfaceErrorCode:
		next := wrapErrorFromCode(clientErrorCode, err)
		return &InterfaceError{&baseError{err: next}}
	case queryArgumentErrorCode:
		next := wrapErrorFromCode(interfaceErrorCode, err)
		return &QueryArgumentError{&baseError{err: next}}
	case missingArgumentErrorCode:
		next := wrapErrorFromCode(queryArgumentErrorCode, err)
		return &MissingArgumentError{&baseError{err: next}}
	case unknownArgumentErrorCode:
		next := wrapErrorFromCode(queryArgumentErrorCode, err)
		return &UnknownArgumentError{&baseError{err: next}}
	case invalidArgumentErrorCode:
		next := wrapErrorFromCode(queryArgumentErrorCode, err)
		return &InvalidArgumentError{&baseError{err: next}}
	case noDataErrorCode:
		next := wrapErrorFromCode(clientErrorCode, err)
		return &NoDataError{&baseError{err: next}}
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
