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

// InternalServerError is an error.
type InternalServerError interface {
	Error
	isInternalServerError()
}

type internalServerError struct {
	msg string
	err error
}

func (e *internalServerError) Error() string {
	msg := e.msg
	if e.err != nil {
		msg = e.err.Error()
	}

	return "edgedb.InternalServerError: " + msg
}

func (e *internalServerError) Unwrap() error { return e.err }

func (e *internalServerError) isInternalServerError() {}

func (e *internalServerError) isError() {}

// UnsupportedFeatureError is an error.
type UnsupportedFeatureError interface {
	Error
	isUnsupportedFeatureError()
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

	return "edgedb.UnsupportedFeatureError: " + msg
}

func (e *unsupportedFeatureError) Unwrap() error { return e.err }

func (e *unsupportedFeatureError) isUnsupportedFeatureError() {}

func (e *unsupportedFeatureError) isError() {}

// ProtocolError is an error.
type ProtocolError interface {
	Error
	isProtocolError()
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

	return "edgedb.ProtocolError: " + msg
}

func (e *protocolError) Unwrap() error { return e.err }

func (e *protocolError) isProtocolError() {}

func (e *protocolError) isError() {}

// BinaryProtocolError is an error.
type BinaryProtocolError interface {
	ProtocolError
	isBinaryProtocolError()
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

	return "edgedb.BinaryProtocolError: " + msg
}

func (e *binaryProtocolError) Unwrap() error { return e.err }

func (e *binaryProtocolError) isBinaryProtocolError() {}

func (e *binaryProtocolError) isProtocolError() {}

func (e *binaryProtocolError) isError() {}

// UnsupportedProtocolVersionError is an error.
type UnsupportedProtocolVersionError interface {
	BinaryProtocolError
	isUnsupportedProtocolVersionError()
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

	return "edgedb.UnsupportedProtocolVersionError: " + msg
}

func (e *unsupportedProtocolVersionError) Unwrap() error { return e.err }

func (e *unsupportedProtocolVersionError) isUnsupportedProtocolVersionError() {}

func (e *unsupportedProtocolVersionError) isBinaryProtocolError() {}

func (e *unsupportedProtocolVersionError) isProtocolError() {}

func (e *unsupportedProtocolVersionError) isError() {}

// TypeSpecNotFoundError is an error.
type TypeSpecNotFoundError interface {
	BinaryProtocolError
	isTypeSpecNotFoundError()
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

	return "edgedb.TypeSpecNotFoundError: " + msg
}

func (e *typeSpecNotFoundError) Unwrap() error { return e.err }

func (e *typeSpecNotFoundError) isTypeSpecNotFoundError() {}

func (e *typeSpecNotFoundError) isBinaryProtocolError() {}

func (e *typeSpecNotFoundError) isProtocolError() {}

func (e *typeSpecNotFoundError) isError() {}

// UnexpectedMessageError is an error.
type UnexpectedMessageError interface {
	BinaryProtocolError
	isUnexpectedMessageError()
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

	return "edgedb.UnexpectedMessageError: " + msg
}

func (e *unexpectedMessageError) Unwrap() error { return e.err }

func (e *unexpectedMessageError) isUnexpectedMessageError() {}

func (e *unexpectedMessageError) isBinaryProtocolError() {}

func (e *unexpectedMessageError) isProtocolError() {}

func (e *unexpectedMessageError) isError() {}

// InputDataError is an error.
type InputDataError interface {
	ProtocolError
	isInputDataError()
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

	return "edgedb.InputDataError: " + msg
}

func (e *inputDataError) Unwrap() error { return e.err }

func (e *inputDataError) isInputDataError() {}

func (e *inputDataError) isProtocolError() {}

func (e *inputDataError) isError() {}

// ResultCardinalityMismatchError is an error.
type ResultCardinalityMismatchError interface {
	ProtocolError
	isResultCardinalityMismatchError()
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

	return "edgedb.ResultCardinalityMismatchError: " + msg
}

func (e *resultCardinalityMismatchError) Unwrap() error { return e.err }

func (e *resultCardinalityMismatchError) isResultCardinalityMismatchError() {}

func (e *resultCardinalityMismatchError) isProtocolError() {}

func (e *resultCardinalityMismatchError) isError() {}

// QueryError is an error.
type QueryError interface {
	Error
	isQueryError()
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

	return "edgedb.QueryError: " + msg
}

func (e *queryError) Unwrap() error { return e.err }

func (e *queryError) isQueryError() {}

func (e *queryError) isError() {}

// InvalidSyntaxError is an error.
type InvalidSyntaxError interface {
	QueryError
	isInvalidSyntaxError()
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

	return "edgedb.InvalidSyntaxError: " + msg
}

func (e *invalidSyntaxError) Unwrap() error { return e.err }

func (e *invalidSyntaxError) isInvalidSyntaxError() {}

func (e *invalidSyntaxError) isQueryError() {}

func (e *invalidSyntaxError) isError() {}

// EdgeQLSyntaxError is an error.
type EdgeQLSyntaxError interface {
	InvalidSyntaxError
	isEdgeQLSyntaxError()
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

	return "edgedb.EdgeQLSyntaxError: " + msg
}

func (e *edgeQLSyntaxError) Unwrap() error { return e.err }

func (e *edgeQLSyntaxError) isEdgeQLSyntaxError() {}

func (e *edgeQLSyntaxError) isInvalidSyntaxError() {}

func (e *edgeQLSyntaxError) isQueryError() {}

func (e *edgeQLSyntaxError) isError() {}

// SchemaSyntaxError is an error.
type SchemaSyntaxError interface {
	InvalidSyntaxError
	isSchemaSyntaxError()
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

	return "edgedb.SchemaSyntaxError: " + msg
}

func (e *schemaSyntaxError) Unwrap() error { return e.err }

func (e *schemaSyntaxError) isSchemaSyntaxError() {}

func (e *schemaSyntaxError) isInvalidSyntaxError() {}

func (e *schemaSyntaxError) isQueryError() {}

func (e *schemaSyntaxError) isError() {}

// GraphQLSyntaxError is an error.
type GraphQLSyntaxError interface {
	InvalidSyntaxError
	isGraphQLSyntaxError()
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

	return "edgedb.GraphQLSyntaxError: " + msg
}

func (e *graphQLSyntaxError) Unwrap() error { return e.err }

func (e *graphQLSyntaxError) isGraphQLSyntaxError() {}

func (e *graphQLSyntaxError) isInvalidSyntaxError() {}

func (e *graphQLSyntaxError) isQueryError() {}

func (e *graphQLSyntaxError) isError() {}

// InvalidTypeError is an error.
type InvalidTypeError interface {
	QueryError
	isInvalidTypeError()
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

	return "edgedb.InvalidTypeError: " + msg
}

func (e *invalidTypeError) Unwrap() error { return e.err }

func (e *invalidTypeError) isInvalidTypeError() {}

func (e *invalidTypeError) isQueryError() {}

func (e *invalidTypeError) isError() {}

// InvalidTargetError is an error.
type InvalidTargetError interface {
	InvalidTypeError
	isInvalidTargetError()
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

	return "edgedb.InvalidTargetError: " + msg
}

func (e *invalidTargetError) Unwrap() error { return e.err }

func (e *invalidTargetError) isInvalidTargetError() {}

func (e *invalidTargetError) isInvalidTypeError() {}

func (e *invalidTargetError) isQueryError() {}

func (e *invalidTargetError) isError() {}

// InvalidLinkTargetError is an error.
type InvalidLinkTargetError interface {
	InvalidTargetError
	isInvalidLinkTargetError()
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

	return "edgedb.InvalidLinkTargetError: " + msg
}

func (e *invalidLinkTargetError) Unwrap() error { return e.err }

func (e *invalidLinkTargetError) isInvalidLinkTargetError() {}

func (e *invalidLinkTargetError) isInvalidTargetError() {}

func (e *invalidLinkTargetError) isInvalidTypeError() {}

func (e *invalidLinkTargetError) isQueryError() {}

func (e *invalidLinkTargetError) isError() {}

// InvalidPropertyTargetError is an error.
type InvalidPropertyTargetError interface {
	InvalidTargetError
	isInvalidPropertyTargetError()
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

	return "edgedb.InvalidPropertyTargetError: " + msg
}

func (e *invalidPropertyTargetError) Unwrap() error { return e.err }

func (e *invalidPropertyTargetError) isInvalidPropertyTargetError() {}

func (e *invalidPropertyTargetError) isInvalidTargetError() {}

func (e *invalidPropertyTargetError) isInvalidTypeError() {}

func (e *invalidPropertyTargetError) isQueryError() {}

func (e *invalidPropertyTargetError) isError() {}

// InvalidReferenceError is an error.
type InvalidReferenceError interface {
	QueryError
	isInvalidReferenceError()
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

	return "edgedb.InvalidReferenceError: " + msg
}

func (e *invalidReferenceError) Unwrap() error { return e.err }

func (e *invalidReferenceError) isInvalidReferenceError() {}

func (e *invalidReferenceError) isQueryError() {}

func (e *invalidReferenceError) isError() {}

// UnknownModuleError is an error.
type UnknownModuleError interface {
	InvalidReferenceError
	isUnknownModuleError()
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

	return "edgedb.UnknownModuleError: " + msg
}

func (e *unknownModuleError) Unwrap() error { return e.err }

func (e *unknownModuleError) isUnknownModuleError() {}

func (e *unknownModuleError) isInvalidReferenceError() {}

func (e *unknownModuleError) isQueryError() {}

func (e *unknownModuleError) isError() {}

// UnknownLinkError is an error.
type UnknownLinkError interface {
	InvalidReferenceError
	isUnknownLinkError()
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

	return "edgedb.UnknownLinkError: " + msg
}

func (e *unknownLinkError) Unwrap() error { return e.err }

func (e *unknownLinkError) isUnknownLinkError() {}

func (e *unknownLinkError) isInvalidReferenceError() {}

func (e *unknownLinkError) isQueryError() {}

func (e *unknownLinkError) isError() {}

// UnknownPropertyError is an error.
type UnknownPropertyError interface {
	InvalidReferenceError
	isUnknownPropertyError()
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

	return "edgedb.UnknownPropertyError: " + msg
}

func (e *unknownPropertyError) Unwrap() error { return e.err }

func (e *unknownPropertyError) isUnknownPropertyError() {}

func (e *unknownPropertyError) isInvalidReferenceError() {}

func (e *unknownPropertyError) isQueryError() {}

func (e *unknownPropertyError) isError() {}

// UnknownUserError is an error.
type UnknownUserError interface {
	InvalidReferenceError
	isUnknownUserError()
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

	return "edgedb.UnknownUserError: " + msg
}

func (e *unknownUserError) Unwrap() error { return e.err }

func (e *unknownUserError) isUnknownUserError() {}

func (e *unknownUserError) isInvalidReferenceError() {}

func (e *unknownUserError) isQueryError() {}

func (e *unknownUserError) isError() {}

// UnknownDatabaseError is an error.
type UnknownDatabaseError interface {
	InvalidReferenceError
	isUnknownDatabaseError()
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

	return "edgedb.UnknownDatabaseError: " + msg
}

func (e *unknownDatabaseError) Unwrap() error { return e.err }

func (e *unknownDatabaseError) isUnknownDatabaseError() {}

func (e *unknownDatabaseError) isInvalidReferenceError() {}

func (e *unknownDatabaseError) isQueryError() {}

func (e *unknownDatabaseError) isError() {}

// UnknownParameterError is an error.
type UnknownParameterError interface {
	InvalidReferenceError
	isUnknownParameterError()
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

	return "edgedb.UnknownParameterError: " + msg
}

func (e *unknownParameterError) Unwrap() error { return e.err }

func (e *unknownParameterError) isUnknownParameterError() {}

func (e *unknownParameterError) isInvalidReferenceError() {}

func (e *unknownParameterError) isQueryError() {}

func (e *unknownParameterError) isError() {}

// SchemaError is an error.
type SchemaError interface {
	QueryError
	isSchemaError()
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

	return "edgedb.SchemaError: " + msg
}

func (e *schemaError) Unwrap() error { return e.err }

func (e *schemaError) isSchemaError() {}

func (e *schemaError) isQueryError() {}

func (e *schemaError) isError() {}

// SchemaDefinitionError is an error.
type SchemaDefinitionError interface {
	QueryError
	isSchemaDefinitionError()
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

	return "edgedb.SchemaDefinitionError: " + msg
}

func (e *schemaDefinitionError) Unwrap() error { return e.err }

func (e *schemaDefinitionError) isSchemaDefinitionError() {}

func (e *schemaDefinitionError) isQueryError() {}

func (e *schemaDefinitionError) isError() {}

// InvalidDefinitionError is an error.
type InvalidDefinitionError interface {
	SchemaDefinitionError
	isInvalidDefinitionError()
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

	return "edgedb.InvalidDefinitionError: " + msg
}

func (e *invalidDefinitionError) Unwrap() error { return e.err }

func (e *invalidDefinitionError) isInvalidDefinitionError() {}

func (e *invalidDefinitionError) isSchemaDefinitionError() {}

func (e *invalidDefinitionError) isQueryError() {}

func (e *invalidDefinitionError) isError() {}

// InvalidModuleDefinitionError is an error.
type InvalidModuleDefinitionError interface {
	InvalidDefinitionError
	isInvalidModuleDefinitionError()
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

	return "edgedb.InvalidModuleDefinitionError: " + msg
}

func (e *invalidModuleDefinitionError) Unwrap() error { return e.err }

func (e *invalidModuleDefinitionError) isInvalidModuleDefinitionError() {}

func (e *invalidModuleDefinitionError) isInvalidDefinitionError() {}

func (e *invalidModuleDefinitionError) isSchemaDefinitionError() {}

func (e *invalidModuleDefinitionError) isQueryError() {}

func (e *invalidModuleDefinitionError) isError() {}

// InvalidLinkDefinitionError is an error.
type InvalidLinkDefinitionError interface {
	InvalidDefinitionError
	isInvalidLinkDefinitionError()
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

	return "edgedb.InvalidLinkDefinitionError: " + msg
}

func (e *invalidLinkDefinitionError) Unwrap() error { return e.err }

func (e *invalidLinkDefinitionError) isInvalidLinkDefinitionError() {}

func (e *invalidLinkDefinitionError) isInvalidDefinitionError() {}

func (e *invalidLinkDefinitionError) isSchemaDefinitionError() {}

func (e *invalidLinkDefinitionError) isQueryError() {}

func (e *invalidLinkDefinitionError) isError() {}

// InvalidPropertyDefinitionError is an error.
type InvalidPropertyDefinitionError interface {
	InvalidDefinitionError
	isInvalidPropertyDefinitionError()
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

	return "edgedb.InvalidPropertyDefinitionError: " + msg
}

func (e *invalidPropertyDefinitionError) Unwrap() error { return e.err }

func (e *invalidPropertyDefinitionError) isInvalidPropertyDefinitionError() {}

func (e *invalidPropertyDefinitionError) isInvalidDefinitionError() {}

func (e *invalidPropertyDefinitionError) isSchemaDefinitionError() {}

func (e *invalidPropertyDefinitionError) isQueryError() {}

func (e *invalidPropertyDefinitionError) isError() {}

// InvalidUserDefinitionError is an error.
type InvalidUserDefinitionError interface {
	InvalidDefinitionError
	isInvalidUserDefinitionError()
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

	return "edgedb.InvalidUserDefinitionError: " + msg
}

func (e *invalidUserDefinitionError) Unwrap() error { return e.err }

func (e *invalidUserDefinitionError) isInvalidUserDefinitionError() {}

func (e *invalidUserDefinitionError) isInvalidDefinitionError() {}

func (e *invalidUserDefinitionError) isSchemaDefinitionError() {}

func (e *invalidUserDefinitionError) isQueryError() {}

func (e *invalidUserDefinitionError) isError() {}

// InvalidDatabaseDefinitionError is an error.
type InvalidDatabaseDefinitionError interface {
	InvalidDefinitionError
	isInvalidDatabaseDefinitionError()
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

	return "edgedb.InvalidDatabaseDefinitionError: " + msg
}

func (e *invalidDatabaseDefinitionError) Unwrap() error { return e.err }

func (e *invalidDatabaseDefinitionError) isInvalidDatabaseDefinitionError() {}

func (e *invalidDatabaseDefinitionError) isInvalidDefinitionError() {}

func (e *invalidDatabaseDefinitionError) isSchemaDefinitionError() {}

func (e *invalidDatabaseDefinitionError) isQueryError() {}

func (e *invalidDatabaseDefinitionError) isError() {}

// InvalidOperatorDefinitionError is an error.
type InvalidOperatorDefinitionError interface {
	InvalidDefinitionError
	isInvalidOperatorDefinitionError()
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

	return "edgedb.InvalidOperatorDefinitionError: " + msg
}

func (e *invalidOperatorDefinitionError) Unwrap() error { return e.err }

func (e *invalidOperatorDefinitionError) isInvalidOperatorDefinitionError() {}

func (e *invalidOperatorDefinitionError) isInvalidDefinitionError() {}

func (e *invalidOperatorDefinitionError) isSchemaDefinitionError() {}

func (e *invalidOperatorDefinitionError) isQueryError() {}

func (e *invalidOperatorDefinitionError) isError() {}

// InvalidAliasDefinitionError is an error.
type InvalidAliasDefinitionError interface {
	InvalidDefinitionError
	isInvalidAliasDefinitionError()
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

	return "edgedb.InvalidAliasDefinitionError: " + msg
}

func (e *invalidAliasDefinitionError) Unwrap() error { return e.err }

func (e *invalidAliasDefinitionError) isInvalidAliasDefinitionError() {}

func (e *invalidAliasDefinitionError) isInvalidDefinitionError() {}

func (e *invalidAliasDefinitionError) isSchemaDefinitionError() {}

func (e *invalidAliasDefinitionError) isQueryError() {}

func (e *invalidAliasDefinitionError) isError() {}

// InvalidFunctionDefinitionError is an error.
type InvalidFunctionDefinitionError interface {
	InvalidDefinitionError
	isInvalidFunctionDefinitionError()
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

	return "edgedb.InvalidFunctionDefinitionError: " + msg
}

func (e *invalidFunctionDefinitionError) Unwrap() error { return e.err }

func (e *invalidFunctionDefinitionError) isInvalidFunctionDefinitionError() {}

func (e *invalidFunctionDefinitionError) isInvalidDefinitionError() {}

func (e *invalidFunctionDefinitionError) isSchemaDefinitionError() {}

func (e *invalidFunctionDefinitionError) isQueryError() {}

func (e *invalidFunctionDefinitionError) isError() {}

// InvalidConstraintDefinitionError is an error.
type InvalidConstraintDefinitionError interface {
	InvalidDefinitionError
	isInvalidConstraintDefinitionError()
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

	return "edgedb.InvalidConstraintDefinitionError: " + msg
}

func (e *invalidConstraintDefinitionError) Unwrap() error { return e.err }

func (e *invalidConstraintDefinitionError) isInvalidConstraintDefinitionError() {}

func (e *invalidConstraintDefinitionError) isInvalidDefinitionError() {}

func (e *invalidConstraintDefinitionError) isSchemaDefinitionError() {}

func (e *invalidConstraintDefinitionError) isQueryError() {}

func (e *invalidConstraintDefinitionError) isError() {}

// InvalidCastDefinitionError is an error.
type InvalidCastDefinitionError interface {
	InvalidDefinitionError
	isInvalidCastDefinitionError()
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

	return "edgedb.InvalidCastDefinitionError: " + msg
}

func (e *invalidCastDefinitionError) Unwrap() error { return e.err }

func (e *invalidCastDefinitionError) isInvalidCastDefinitionError() {}

func (e *invalidCastDefinitionError) isInvalidDefinitionError() {}

func (e *invalidCastDefinitionError) isSchemaDefinitionError() {}

func (e *invalidCastDefinitionError) isQueryError() {}

func (e *invalidCastDefinitionError) isError() {}

// DuplicateDefinitionError is an error.
type DuplicateDefinitionError interface {
	SchemaDefinitionError
	isDuplicateDefinitionError()
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

	return "edgedb.DuplicateDefinitionError: " + msg
}

func (e *duplicateDefinitionError) Unwrap() error { return e.err }

func (e *duplicateDefinitionError) isDuplicateDefinitionError() {}

func (e *duplicateDefinitionError) isSchemaDefinitionError() {}

func (e *duplicateDefinitionError) isQueryError() {}

func (e *duplicateDefinitionError) isError() {}

// DuplicateModuleDefinitionError is an error.
type DuplicateModuleDefinitionError interface {
	DuplicateDefinitionError
	isDuplicateModuleDefinitionError()
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

	return "edgedb.DuplicateModuleDefinitionError: " + msg
}

func (e *duplicateModuleDefinitionError) Unwrap() error { return e.err }

func (e *duplicateModuleDefinitionError) isDuplicateModuleDefinitionError() {}

func (e *duplicateModuleDefinitionError) isDuplicateDefinitionError() {}

func (e *duplicateModuleDefinitionError) isSchemaDefinitionError() {}

func (e *duplicateModuleDefinitionError) isQueryError() {}

func (e *duplicateModuleDefinitionError) isError() {}

// DuplicateLinkDefinitionError is an error.
type DuplicateLinkDefinitionError interface {
	DuplicateDefinitionError
	isDuplicateLinkDefinitionError()
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

	return "edgedb.DuplicateLinkDefinitionError: " + msg
}

func (e *duplicateLinkDefinitionError) Unwrap() error { return e.err }

func (e *duplicateLinkDefinitionError) isDuplicateLinkDefinitionError() {}

func (e *duplicateLinkDefinitionError) isDuplicateDefinitionError() {}

func (e *duplicateLinkDefinitionError) isSchemaDefinitionError() {}

func (e *duplicateLinkDefinitionError) isQueryError() {}

func (e *duplicateLinkDefinitionError) isError() {}

// DuplicatePropertyDefinitionError is an error.
type DuplicatePropertyDefinitionError interface {
	DuplicateDefinitionError
	isDuplicatePropertyDefinitionError()
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

	return "edgedb.DuplicatePropertyDefinitionError: " + msg
}

func (e *duplicatePropertyDefinitionError) Unwrap() error { return e.err }

func (e *duplicatePropertyDefinitionError) isDuplicatePropertyDefinitionError() {}

func (e *duplicatePropertyDefinitionError) isDuplicateDefinitionError() {}

func (e *duplicatePropertyDefinitionError) isSchemaDefinitionError() {}

func (e *duplicatePropertyDefinitionError) isQueryError() {}

func (e *duplicatePropertyDefinitionError) isError() {}

// DuplicateUserDefinitionError is an error.
type DuplicateUserDefinitionError interface {
	DuplicateDefinitionError
	isDuplicateUserDefinitionError()
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

	return "edgedb.DuplicateUserDefinitionError: " + msg
}

func (e *duplicateUserDefinitionError) Unwrap() error { return e.err }

func (e *duplicateUserDefinitionError) isDuplicateUserDefinitionError() {}

func (e *duplicateUserDefinitionError) isDuplicateDefinitionError() {}

func (e *duplicateUserDefinitionError) isSchemaDefinitionError() {}

func (e *duplicateUserDefinitionError) isQueryError() {}

func (e *duplicateUserDefinitionError) isError() {}

// DuplicateDatabaseDefinitionError is an error.
type DuplicateDatabaseDefinitionError interface {
	DuplicateDefinitionError
	isDuplicateDatabaseDefinitionError()
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

	return "edgedb.DuplicateDatabaseDefinitionError: " + msg
}

func (e *duplicateDatabaseDefinitionError) Unwrap() error { return e.err }

func (e *duplicateDatabaseDefinitionError) isDuplicateDatabaseDefinitionError() {}

func (e *duplicateDatabaseDefinitionError) isDuplicateDefinitionError() {}

func (e *duplicateDatabaseDefinitionError) isSchemaDefinitionError() {}

func (e *duplicateDatabaseDefinitionError) isQueryError() {}

func (e *duplicateDatabaseDefinitionError) isError() {}

// DuplicateOperatorDefinitionError is an error.
type DuplicateOperatorDefinitionError interface {
	DuplicateDefinitionError
	isDuplicateOperatorDefinitionError()
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

	return "edgedb.DuplicateOperatorDefinitionError: " + msg
}

func (e *duplicateOperatorDefinitionError) Unwrap() error { return e.err }

func (e *duplicateOperatorDefinitionError) isDuplicateOperatorDefinitionError() {}

func (e *duplicateOperatorDefinitionError) isDuplicateDefinitionError() {}

func (e *duplicateOperatorDefinitionError) isSchemaDefinitionError() {}

func (e *duplicateOperatorDefinitionError) isQueryError() {}

func (e *duplicateOperatorDefinitionError) isError() {}

// DuplicateViewDefinitionError is an error.
type DuplicateViewDefinitionError interface {
	DuplicateDefinitionError
	isDuplicateViewDefinitionError()
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

	return "edgedb.DuplicateViewDefinitionError: " + msg
}

func (e *duplicateViewDefinitionError) Unwrap() error { return e.err }

func (e *duplicateViewDefinitionError) isDuplicateViewDefinitionError() {}

func (e *duplicateViewDefinitionError) isDuplicateDefinitionError() {}

func (e *duplicateViewDefinitionError) isSchemaDefinitionError() {}

func (e *duplicateViewDefinitionError) isQueryError() {}

func (e *duplicateViewDefinitionError) isError() {}

// DuplicateFunctionDefinitionError is an error.
type DuplicateFunctionDefinitionError interface {
	DuplicateDefinitionError
	isDuplicateFunctionDefinitionError()
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

	return "edgedb.DuplicateFunctionDefinitionError: " + msg
}

func (e *duplicateFunctionDefinitionError) Unwrap() error { return e.err }

func (e *duplicateFunctionDefinitionError) isDuplicateFunctionDefinitionError() {}

func (e *duplicateFunctionDefinitionError) isDuplicateDefinitionError() {}

func (e *duplicateFunctionDefinitionError) isSchemaDefinitionError() {}

func (e *duplicateFunctionDefinitionError) isQueryError() {}

func (e *duplicateFunctionDefinitionError) isError() {}

// DuplicateConstraintDefinitionError is an error.
type DuplicateConstraintDefinitionError interface {
	DuplicateDefinitionError
	isDuplicateConstraintDefinitionError()
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

	return "edgedb.DuplicateConstraintDefinitionError: " + msg
}

func (e *duplicateConstraintDefinitionError) Unwrap() error { return e.err }

func (e *duplicateConstraintDefinitionError) isDuplicateConstraintDefinitionError() {}

func (e *duplicateConstraintDefinitionError) isDuplicateDefinitionError() {}

func (e *duplicateConstraintDefinitionError) isSchemaDefinitionError() {}

func (e *duplicateConstraintDefinitionError) isQueryError() {}

func (e *duplicateConstraintDefinitionError) isError() {}

// DuplicateCastDefinitionError is an error.
type DuplicateCastDefinitionError interface {
	DuplicateDefinitionError
	isDuplicateCastDefinitionError()
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

	return "edgedb.DuplicateCastDefinitionError: " + msg
}

func (e *duplicateCastDefinitionError) Unwrap() error { return e.err }

func (e *duplicateCastDefinitionError) isDuplicateCastDefinitionError() {}

func (e *duplicateCastDefinitionError) isDuplicateDefinitionError() {}

func (e *duplicateCastDefinitionError) isSchemaDefinitionError() {}

func (e *duplicateCastDefinitionError) isQueryError() {}

func (e *duplicateCastDefinitionError) isError() {}

// QueryTimeoutError is an error.
type QueryTimeoutError interface {
	QueryError
	isQueryTimeoutError()
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

	return "edgedb.QueryTimeoutError: " + msg
}

func (e *queryTimeoutError) Unwrap() error { return e.err }

func (e *queryTimeoutError) isQueryTimeoutError() {}

func (e *queryTimeoutError) isQueryError() {}

func (e *queryTimeoutError) isError() {}

// ExecutionError is an error.
type ExecutionError interface {
	Error
	isExecutionError()
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

	return "edgedb.ExecutionError: " + msg
}

func (e *executionError) Unwrap() error { return e.err }

func (e *executionError) isExecutionError() {}

func (e *executionError) isError() {}

// InvalidValueError is an error.
type InvalidValueError interface {
	ExecutionError
	isInvalidValueError()
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

	return "edgedb.InvalidValueError: " + msg
}

func (e *invalidValueError) Unwrap() error { return e.err }

func (e *invalidValueError) isInvalidValueError() {}

func (e *invalidValueError) isExecutionError() {}

func (e *invalidValueError) isError() {}

// DivisionByZeroError is an error.
type DivisionByZeroError interface {
	InvalidValueError
	isDivisionByZeroError()
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

	return "edgedb.DivisionByZeroError: " + msg
}

func (e *divisionByZeroError) Unwrap() error { return e.err }

func (e *divisionByZeroError) isDivisionByZeroError() {}

func (e *divisionByZeroError) isInvalidValueError() {}

func (e *divisionByZeroError) isExecutionError() {}

func (e *divisionByZeroError) isError() {}

// NumericOutOfRangeError is an error.
type NumericOutOfRangeError interface {
	InvalidValueError
	isNumericOutOfRangeError()
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

	return "edgedb.NumericOutOfRangeError: " + msg
}

func (e *numericOutOfRangeError) Unwrap() error { return e.err }

func (e *numericOutOfRangeError) isNumericOutOfRangeError() {}

func (e *numericOutOfRangeError) isInvalidValueError() {}

func (e *numericOutOfRangeError) isExecutionError() {}

func (e *numericOutOfRangeError) isError() {}

// IntegrityError is an error.
type IntegrityError interface {
	ExecutionError
	isIntegrityError()
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

	return "edgedb.IntegrityError: " + msg
}

func (e *integrityError) Unwrap() error { return e.err }

func (e *integrityError) isIntegrityError() {}

func (e *integrityError) isExecutionError() {}

func (e *integrityError) isError() {}

// ConstraintViolationError is an error.
type ConstraintViolationError interface {
	IntegrityError
	isConstraintViolationError()
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

	return "edgedb.ConstraintViolationError: " + msg
}

func (e *constraintViolationError) Unwrap() error { return e.err }

func (e *constraintViolationError) isConstraintViolationError() {}

func (e *constraintViolationError) isIntegrityError() {}

func (e *constraintViolationError) isExecutionError() {}

func (e *constraintViolationError) isError() {}

// CardinalityViolationError is an error.
type CardinalityViolationError interface {
	IntegrityError
	isCardinalityViolationError()
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

	return "edgedb.CardinalityViolationError: " + msg
}

func (e *cardinalityViolationError) Unwrap() error { return e.err }

func (e *cardinalityViolationError) isCardinalityViolationError() {}

func (e *cardinalityViolationError) isIntegrityError() {}

func (e *cardinalityViolationError) isExecutionError() {}

func (e *cardinalityViolationError) isError() {}

// MissingRequiredError is an error.
type MissingRequiredError interface {
	IntegrityError
	isMissingRequiredError()
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

	return "edgedb.MissingRequiredError: " + msg
}

func (e *missingRequiredError) Unwrap() error { return e.err }

func (e *missingRequiredError) isMissingRequiredError() {}

func (e *missingRequiredError) isIntegrityError() {}

func (e *missingRequiredError) isExecutionError() {}

func (e *missingRequiredError) isError() {}

// TransactionError is an error.
type TransactionError interface {
	ExecutionError
	isTransactionError()
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

	return "edgedb.TransactionError: " + msg
}

func (e *transactionError) Unwrap() error { return e.err }

func (e *transactionError) isTransactionError() {}

func (e *transactionError) isExecutionError() {}

func (e *transactionError) isError() {}

// TransactionSerializationError is an error.
type TransactionSerializationError interface {
	TransactionError
	isTransactionSerializationError()
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

	return "edgedb.TransactionSerializationError: " + msg
}

func (e *transactionSerializationError) Unwrap() error { return e.err }

func (e *transactionSerializationError) isTransactionSerializationError() {}

func (e *transactionSerializationError) isTransactionError() {}

func (e *transactionSerializationError) isExecutionError() {}

func (e *transactionSerializationError) isError() {}

// TransactionDeadlockError is an error.
type TransactionDeadlockError interface {
	TransactionError
	isTransactionDeadlockError()
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

	return "edgedb.TransactionDeadlockError: " + msg
}

func (e *transactionDeadlockError) Unwrap() error { return e.err }

func (e *transactionDeadlockError) isTransactionDeadlockError() {}

func (e *transactionDeadlockError) isTransactionError() {}

func (e *transactionDeadlockError) isExecutionError() {}

func (e *transactionDeadlockError) isError() {}

// ConfigurationError is an error.
type ConfigurationError interface {
	Error
	isConfigurationError()
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

	return "edgedb.ConfigurationError: " + msg
}

func (e *configurationError) Unwrap() error { return e.err }

func (e *configurationError) isConfigurationError() {}

func (e *configurationError) isError() {}

// AccessError is an error.
type AccessError interface {
	Error
	isAccessError()
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

	return "edgedb.AccessError: " + msg
}

func (e *accessError) Unwrap() error { return e.err }

func (e *accessError) isAccessError() {}

func (e *accessError) isError() {}

// AuthenticationError is an error.
type AuthenticationError interface {
	AccessError
	isAuthenticationError()
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

	return "edgedb.AuthenticationError: " + msg
}

func (e *authenticationError) Unwrap() error { return e.err }

func (e *authenticationError) isAuthenticationError() {}

func (e *authenticationError) isAccessError() {}

func (e *authenticationError) isError() {}

// LogMessage is an error.
type LogMessage interface {
	Error
	isLogMessage()
}

type logMessage struct {
	msg string
	err error
}

func (e *logMessage) Error() string {
	msg := e.msg
	if e.err != nil {
		msg = e.err.Error()
	}

	return "edgedb.LogMessage: " + msg
}

func (e *logMessage) Unwrap() error { return e.err }

func (e *logMessage) isLogMessage() {}

func (e *logMessage) isError() {}

// WarningMessage is an error.
type WarningMessage interface {
	Error
	isWarningMessage()
}

type warningMessage struct {
	msg string
	err error
}

func (e *warningMessage) Error() string {
	msg := e.msg
	if e.err != nil {
		msg = e.err.Error()
	}

	return "edgedb.WarningMessage: " + msg
}

func (e *warningMessage) Unwrap() error { return e.err }

func (e *warningMessage) isWarningMessage() {}

func (e *warningMessage) isError() {}

// ClientError is an error.
type ClientError interface {
	Error
	isClientError()
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

	return "edgedb.ClientError: " + msg
}

func (e *clientError) Unwrap() error { return e.err }

func (e *clientError) isClientError() {}

func (e *clientError) isError() {}

// ClientConnectionError is an error.
type ClientConnectionError interface {
	ClientError
	isClientConnectionError()
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

	return "edgedb.ClientConnectionError: " + msg
}

func (e *clientConnectionError) Unwrap() error { return e.err }

func (e *clientConnectionError) isClientConnectionError() {}

func (e *clientConnectionError) isClientError() {}

func (e *clientConnectionError) isError() {}

// InterfaceError is an error.
type InterfaceError interface {
	ClientError
	isInterfaceError()
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

	return "edgedb.InterfaceError: " + msg
}

func (e *interfaceError) Unwrap() error { return e.err }

func (e *interfaceError) isInterfaceError() {}

func (e *interfaceError) isClientError() {}

func (e *interfaceError) isError() {}

// QueryArgumentError is an error.
type QueryArgumentError interface {
	InterfaceError
	isQueryArgumentError()
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

	return "edgedb.QueryArgumentError: " + msg
}

func (e *queryArgumentError) Unwrap() error { return e.err }

func (e *queryArgumentError) isQueryArgumentError() {}

func (e *queryArgumentError) isInterfaceError() {}

func (e *queryArgumentError) isClientError() {}

func (e *queryArgumentError) isError() {}

// MissingArgumentError is an error.
type MissingArgumentError interface {
	QueryArgumentError
	isMissingArgumentError()
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

	return "edgedb.MissingArgumentError: " + msg
}

func (e *missingArgumentError) Unwrap() error { return e.err }

func (e *missingArgumentError) isMissingArgumentError() {}

func (e *missingArgumentError) isQueryArgumentError() {}

func (e *missingArgumentError) isInterfaceError() {}

func (e *missingArgumentError) isClientError() {}

func (e *missingArgumentError) isError() {}

// UnknownArgumentError is an error.
type UnknownArgumentError interface {
	QueryArgumentError
	isUnknownArgumentError()
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

	return "edgedb.UnknownArgumentError: " + msg
}

func (e *unknownArgumentError) Unwrap() error { return e.err }

func (e *unknownArgumentError) isUnknownArgumentError() {}

func (e *unknownArgumentError) isQueryArgumentError() {}

func (e *unknownArgumentError) isInterfaceError() {}

func (e *unknownArgumentError) isClientError() {}

func (e *unknownArgumentError) isError() {}

// InvalidArgumentError is an error.
type InvalidArgumentError interface {
	QueryArgumentError
	isInvalidArgumentError()
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

	return "edgedb.InvalidArgumentError: " + msg
}

func (e *invalidArgumentError) Unwrap() error { return e.err }

func (e *invalidArgumentError) isInvalidArgumentError() {}

func (e *invalidArgumentError) isQueryArgumentError() {}

func (e *invalidArgumentError) isInterfaceError() {}

func (e *invalidArgumentError) isClientError() {}

func (e *invalidArgumentError) isError() {}

// NoDataError is an error.
type NoDataError interface {
	ClientError
	isNoDataError()
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

	return "edgedb.NoDataError: " + msg
}

func (e *noDataError) Unwrap() error { return e.err }

func (e *noDataError) isNoDataError() {}

func (e *noDataError) isClientError() {}

func (e *noDataError) isError() {}

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
	case 0x03_03_00_00:
		return &resultCardinalityMismatchError{msg: msg}
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
	case 0x04_06_00_00:
		return &queryTimeoutError{msg: msg}
	case 0x05_00_00_00:
		return &executionError{msg: msg}
	case 0x05_01_00_00:
		return &invalidValueError{msg: msg}
	case 0x05_01_00_01:
		return &divisionByZeroError{msg: msg}
	case 0x05_01_00_02:
		return &numericOutOfRangeError{msg: msg}
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
	case 0x05_03_00_01:
		return &transactionSerializationError{msg: msg}
	case 0x05_03_00_02:
		return &transactionDeadlockError{msg: msg}
	case 0x06_00_00_00:
		return &configurationError{msg: msg}
	case 0x07_00_00_00:
		return &accessError{msg: msg}
	case 0x07_01_00_00:
		return &authenticationError{msg: msg}
	case 0xf0_00_00_00:
		return &logMessage{msg: msg}
	case 0xf0_01_00_00:
		return &warningMessage{msg: msg}
	case 0xff_00_00_00:
		return &clientError{msg: msg}
	case 0xff_01_00_00:
		return &clientConnectionError{msg: msg}
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
	default:
		panic(fmt.Sprintf("invalid error code 0x%x", code))
	}
}
