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
// run 'make errors' to regenerate

package edgedb

import "fmt"

const (
	// ShouldRetry is an error tag.
	ShouldRetry ErrorTag = "SHOULD_RETRY"
	// ShouldReconnect is an error tag.
	ShouldReconnect ErrorTag = "SHOULD_RECONNECT"
)

// InternalServerError is an error.
type InternalServerError interface {
	Error
	isEdgeDBInternalServerError()
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

func (e *internalServerError) isEdgeDBInternalServerError() {}

func (e *internalServerError) isEdgeDBError() {}

func (e *internalServerError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}
// UnsupportedFeatureError is an error.
type UnsupportedFeatureError interface {
	Error
	isEdgeDBUnsupportedFeatureError()
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

func (e *unsupportedFeatureError) isEdgeDBUnsupportedFeatureError() {}

func (e *unsupportedFeatureError) isEdgeDBError() {}

func (e *unsupportedFeatureError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}
// ProtocolError is an error.
type ProtocolError interface {
	Error
	isEdgeDBProtocolError()
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

func (e *protocolError) isEdgeDBProtocolError() {}

func (e *protocolError) isEdgeDBError() {}

func (e *protocolError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}
// BinaryProtocolError is an error.
type BinaryProtocolError interface {
	ProtocolError
	isEdgeDBBinaryProtocolError()
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

func (e *binaryProtocolError) isEdgeDBBinaryProtocolError() {}

func (e *binaryProtocolError) isEdgeDBProtocolError() {}

func (e *binaryProtocolError) isEdgeDBError() {}

func (e *binaryProtocolError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}
// UnsupportedProtocolVersionError is an error.
type UnsupportedProtocolVersionError interface {
	BinaryProtocolError
	isEdgeDBUnsupportedProtocolVersionError()
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

func (e *unsupportedProtocolVersionError) isEdgeDBUnsupportedProtocolVersionError() {}

func (e *unsupportedProtocolVersionError) isEdgeDBBinaryProtocolError() {}

func (e *unsupportedProtocolVersionError) isEdgeDBProtocolError() {}

func (e *unsupportedProtocolVersionError) isEdgeDBError() {}

func (e *unsupportedProtocolVersionError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}
// TypeSpecNotFoundError is an error.
type TypeSpecNotFoundError interface {
	BinaryProtocolError
	isEdgeDBTypeSpecNotFoundError()
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

func (e *typeSpecNotFoundError) isEdgeDBTypeSpecNotFoundError() {}

func (e *typeSpecNotFoundError) isEdgeDBBinaryProtocolError() {}

func (e *typeSpecNotFoundError) isEdgeDBProtocolError() {}

func (e *typeSpecNotFoundError) isEdgeDBError() {}

func (e *typeSpecNotFoundError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}
// UnexpectedMessageError is an error.
type UnexpectedMessageError interface {
	BinaryProtocolError
	isEdgeDBUnexpectedMessageError()
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

func (e *unexpectedMessageError) isEdgeDBUnexpectedMessageError() {}

func (e *unexpectedMessageError) isEdgeDBBinaryProtocolError() {}

func (e *unexpectedMessageError) isEdgeDBProtocolError() {}

func (e *unexpectedMessageError) isEdgeDBError() {}

func (e *unexpectedMessageError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}
// InputDataError is an error.
type InputDataError interface {
	ProtocolError
	isEdgeDBInputDataError()
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

func (e *inputDataError) isEdgeDBInputDataError() {}

func (e *inputDataError) isEdgeDBProtocolError() {}

func (e *inputDataError) isEdgeDBError() {}

func (e *inputDataError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}
// ResultCardinalityMismatchError is an error.
type ResultCardinalityMismatchError interface {
	ProtocolError
	isEdgeDBResultCardinalityMismatchError()
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

func (e *resultCardinalityMismatchError) isEdgeDBResultCardinalityMismatchError() {}

func (e *resultCardinalityMismatchError) isEdgeDBProtocolError() {}

func (e *resultCardinalityMismatchError) isEdgeDBError() {}

func (e *resultCardinalityMismatchError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}
// CapabilityError is an error.
type CapabilityError interface {
	ProtocolError
	isEdgeDBCapabilityError()
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

	return "edgedb.CapabilityError: " + msg
}

func (e *capabilityError) Unwrap() error { return e.err }

func (e *capabilityError) isEdgeDBCapabilityError() {}

func (e *capabilityError) isEdgeDBProtocolError() {}

func (e *capabilityError) isEdgeDBError() {}

func (e *capabilityError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}
// UnsupportedCapabilityError is an error.
type UnsupportedCapabilityError interface {
	CapabilityError
	isEdgeDBUnsupportedCapabilityError()
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

	return "edgedb.UnsupportedCapabilityError: " + msg
}

func (e *unsupportedCapabilityError) Unwrap() error { return e.err }

func (e *unsupportedCapabilityError) isEdgeDBUnsupportedCapabilityError() {}

func (e *unsupportedCapabilityError) isEdgeDBCapabilityError() {}

func (e *unsupportedCapabilityError) isEdgeDBProtocolError() {}

func (e *unsupportedCapabilityError) isEdgeDBError() {}

func (e *unsupportedCapabilityError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}
// DisabledCapabilityError is an error.
type DisabledCapabilityError interface {
	CapabilityError
	isEdgeDBDisabledCapabilityError()
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

	return "edgedb.DisabledCapabilityError: " + msg
}

func (e *disabledCapabilityError) Unwrap() error { return e.err }

func (e *disabledCapabilityError) isEdgeDBDisabledCapabilityError() {}

func (e *disabledCapabilityError) isEdgeDBCapabilityError() {}

func (e *disabledCapabilityError) isEdgeDBProtocolError() {}

func (e *disabledCapabilityError) isEdgeDBError() {}

func (e *disabledCapabilityError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}
// QueryError is an error.
type QueryError interface {
	Error
	isEdgeDBQueryError()
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

func (e *queryError) isEdgeDBQueryError() {}

func (e *queryError) isEdgeDBError() {}

func (e *queryError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}
// InvalidSyntaxError is an error.
type InvalidSyntaxError interface {
	QueryError
	isEdgeDBInvalidSyntaxError()
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

func (e *invalidSyntaxError) isEdgeDBInvalidSyntaxError() {}

func (e *invalidSyntaxError) isEdgeDBQueryError() {}

func (e *invalidSyntaxError) isEdgeDBError() {}

func (e *invalidSyntaxError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}
// EdgeQLSyntaxError is an error.
type EdgeQLSyntaxError interface {
	InvalidSyntaxError
	isEdgeDBEdgeQLSyntaxError()
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

func (e *edgeQLSyntaxError) isEdgeDBEdgeQLSyntaxError() {}

func (e *edgeQLSyntaxError) isEdgeDBInvalidSyntaxError() {}

func (e *edgeQLSyntaxError) isEdgeDBQueryError() {}

func (e *edgeQLSyntaxError) isEdgeDBError() {}

func (e *edgeQLSyntaxError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}
// SchemaSyntaxError is an error.
type SchemaSyntaxError interface {
	InvalidSyntaxError
	isEdgeDBSchemaSyntaxError()
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

func (e *schemaSyntaxError) isEdgeDBSchemaSyntaxError() {}

func (e *schemaSyntaxError) isEdgeDBInvalidSyntaxError() {}

func (e *schemaSyntaxError) isEdgeDBQueryError() {}

func (e *schemaSyntaxError) isEdgeDBError() {}

func (e *schemaSyntaxError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}
// GraphQLSyntaxError is an error.
type GraphQLSyntaxError interface {
	InvalidSyntaxError
	isEdgeDBGraphQLSyntaxError()
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

func (e *graphQLSyntaxError) isEdgeDBGraphQLSyntaxError() {}

func (e *graphQLSyntaxError) isEdgeDBInvalidSyntaxError() {}

func (e *graphQLSyntaxError) isEdgeDBQueryError() {}

func (e *graphQLSyntaxError) isEdgeDBError() {}

func (e *graphQLSyntaxError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}
// InvalidTypeError is an error.
type InvalidTypeError interface {
	QueryError
	isEdgeDBInvalidTypeError()
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

func (e *invalidTypeError) isEdgeDBInvalidTypeError() {}

func (e *invalidTypeError) isEdgeDBQueryError() {}

func (e *invalidTypeError) isEdgeDBError() {}

func (e *invalidTypeError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}
// InvalidTargetError is an error.
type InvalidTargetError interface {
	InvalidTypeError
	isEdgeDBInvalidTargetError()
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

func (e *invalidTargetError) isEdgeDBInvalidTargetError() {}

func (e *invalidTargetError) isEdgeDBInvalidTypeError() {}

func (e *invalidTargetError) isEdgeDBQueryError() {}

func (e *invalidTargetError) isEdgeDBError() {}

func (e *invalidTargetError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}
// InvalidLinkTargetError is an error.
type InvalidLinkTargetError interface {
	InvalidTargetError
	isEdgeDBInvalidLinkTargetError()
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

func (e *invalidLinkTargetError) isEdgeDBInvalidLinkTargetError() {}

func (e *invalidLinkTargetError) isEdgeDBInvalidTargetError() {}

func (e *invalidLinkTargetError) isEdgeDBInvalidTypeError() {}

func (e *invalidLinkTargetError) isEdgeDBQueryError() {}

func (e *invalidLinkTargetError) isEdgeDBError() {}

func (e *invalidLinkTargetError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}
// InvalidPropertyTargetError is an error.
type InvalidPropertyTargetError interface {
	InvalidTargetError
	isEdgeDBInvalidPropertyTargetError()
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

func (e *invalidPropertyTargetError) isEdgeDBInvalidPropertyTargetError() {}

func (e *invalidPropertyTargetError) isEdgeDBInvalidTargetError() {}

func (e *invalidPropertyTargetError) isEdgeDBInvalidTypeError() {}

func (e *invalidPropertyTargetError) isEdgeDBQueryError() {}

func (e *invalidPropertyTargetError) isEdgeDBError() {}

func (e *invalidPropertyTargetError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}
// InvalidReferenceError is an error.
type InvalidReferenceError interface {
	QueryError
	isEdgeDBInvalidReferenceError()
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

func (e *invalidReferenceError) isEdgeDBInvalidReferenceError() {}

func (e *invalidReferenceError) isEdgeDBQueryError() {}

func (e *invalidReferenceError) isEdgeDBError() {}

func (e *invalidReferenceError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}
// UnknownModuleError is an error.
type UnknownModuleError interface {
	InvalidReferenceError
	isEdgeDBUnknownModuleError()
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

func (e *unknownModuleError) isEdgeDBUnknownModuleError() {}

func (e *unknownModuleError) isEdgeDBInvalidReferenceError() {}

func (e *unknownModuleError) isEdgeDBQueryError() {}

func (e *unknownModuleError) isEdgeDBError() {}

func (e *unknownModuleError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}
// UnknownLinkError is an error.
type UnknownLinkError interface {
	InvalidReferenceError
	isEdgeDBUnknownLinkError()
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

func (e *unknownLinkError) isEdgeDBUnknownLinkError() {}

func (e *unknownLinkError) isEdgeDBInvalidReferenceError() {}

func (e *unknownLinkError) isEdgeDBQueryError() {}

func (e *unknownLinkError) isEdgeDBError() {}

func (e *unknownLinkError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}
// UnknownPropertyError is an error.
type UnknownPropertyError interface {
	InvalidReferenceError
	isEdgeDBUnknownPropertyError()
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

func (e *unknownPropertyError) isEdgeDBUnknownPropertyError() {}

func (e *unknownPropertyError) isEdgeDBInvalidReferenceError() {}

func (e *unknownPropertyError) isEdgeDBQueryError() {}

func (e *unknownPropertyError) isEdgeDBError() {}

func (e *unknownPropertyError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}
// UnknownUserError is an error.
type UnknownUserError interface {
	InvalidReferenceError
	isEdgeDBUnknownUserError()
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

func (e *unknownUserError) isEdgeDBUnknownUserError() {}

func (e *unknownUserError) isEdgeDBInvalidReferenceError() {}

func (e *unknownUserError) isEdgeDBQueryError() {}

func (e *unknownUserError) isEdgeDBError() {}

func (e *unknownUserError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}
// UnknownDatabaseError is an error.
type UnknownDatabaseError interface {
	InvalidReferenceError
	isEdgeDBUnknownDatabaseError()
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

func (e *unknownDatabaseError) isEdgeDBUnknownDatabaseError() {}

func (e *unknownDatabaseError) isEdgeDBInvalidReferenceError() {}

func (e *unknownDatabaseError) isEdgeDBQueryError() {}

func (e *unknownDatabaseError) isEdgeDBError() {}

func (e *unknownDatabaseError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}
// UnknownParameterError is an error.
type UnknownParameterError interface {
	InvalidReferenceError
	isEdgeDBUnknownParameterError()
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

func (e *unknownParameterError) isEdgeDBUnknownParameterError() {}

func (e *unknownParameterError) isEdgeDBInvalidReferenceError() {}

func (e *unknownParameterError) isEdgeDBQueryError() {}

func (e *unknownParameterError) isEdgeDBError() {}

func (e *unknownParameterError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}
// SchemaError is an error.
type SchemaError interface {
	QueryError
	isEdgeDBSchemaError()
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

func (e *schemaError) isEdgeDBSchemaError() {}

func (e *schemaError) isEdgeDBQueryError() {}

func (e *schemaError) isEdgeDBError() {}

func (e *schemaError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}
// SchemaDefinitionError is an error.
type SchemaDefinitionError interface {
	QueryError
	isEdgeDBSchemaDefinitionError()
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

func (e *schemaDefinitionError) isEdgeDBSchemaDefinitionError() {}

func (e *schemaDefinitionError) isEdgeDBQueryError() {}

func (e *schemaDefinitionError) isEdgeDBError() {}

func (e *schemaDefinitionError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}
// InvalidDefinitionError is an error.
type InvalidDefinitionError interface {
	SchemaDefinitionError
	isEdgeDBInvalidDefinitionError()
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

func (e *invalidDefinitionError) isEdgeDBInvalidDefinitionError() {}

func (e *invalidDefinitionError) isEdgeDBSchemaDefinitionError() {}

func (e *invalidDefinitionError) isEdgeDBQueryError() {}

func (e *invalidDefinitionError) isEdgeDBError() {}

func (e *invalidDefinitionError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}
// InvalidModuleDefinitionError is an error.
type InvalidModuleDefinitionError interface {
	InvalidDefinitionError
	isEdgeDBInvalidModuleDefinitionError()
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

func (e *invalidModuleDefinitionError) isEdgeDBInvalidModuleDefinitionError() {}

func (e *invalidModuleDefinitionError) isEdgeDBInvalidDefinitionError() {}

func (e *invalidModuleDefinitionError) isEdgeDBSchemaDefinitionError() {}

func (e *invalidModuleDefinitionError) isEdgeDBQueryError() {}

func (e *invalidModuleDefinitionError) isEdgeDBError() {}

func (e *invalidModuleDefinitionError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}
// InvalidLinkDefinitionError is an error.
type InvalidLinkDefinitionError interface {
	InvalidDefinitionError
	isEdgeDBInvalidLinkDefinitionError()
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

func (e *invalidLinkDefinitionError) isEdgeDBInvalidLinkDefinitionError() {}

func (e *invalidLinkDefinitionError) isEdgeDBInvalidDefinitionError() {}

func (e *invalidLinkDefinitionError) isEdgeDBSchemaDefinitionError() {}

func (e *invalidLinkDefinitionError) isEdgeDBQueryError() {}

func (e *invalidLinkDefinitionError) isEdgeDBError() {}

func (e *invalidLinkDefinitionError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}
// InvalidPropertyDefinitionError is an error.
type InvalidPropertyDefinitionError interface {
	InvalidDefinitionError
	isEdgeDBInvalidPropertyDefinitionError()
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

func (e *invalidPropertyDefinitionError) isEdgeDBInvalidPropertyDefinitionError() {}

func (e *invalidPropertyDefinitionError) isEdgeDBInvalidDefinitionError() {}

func (e *invalidPropertyDefinitionError) isEdgeDBSchemaDefinitionError() {}

func (e *invalidPropertyDefinitionError) isEdgeDBQueryError() {}

func (e *invalidPropertyDefinitionError) isEdgeDBError() {}

func (e *invalidPropertyDefinitionError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}
// InvalidUserDefinitionError is an error.
type InvalidUserDefinitionError interface {
	InvalidDefinitionError
	isEdgeDBInvalidUserDefinitionError()
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

func (e *invalidUserDefinitionError) isEdgeDBInvalidUserDefinitionError() {}

func (e *invalidUserDefinitionError) isEdgeDBInvalidDefinitionError() {}

func (e *invalidUserDefinitionError) isEdgeDBSchemaDefinitionError() {}

func (e *invalidUserDefinitionError) isEdgeDBQueryError() {}

func (e *invalidUserDefinitionError) isEdgeDBError() {}

func (e *invalidUserDefinitionError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}
// InvalidDatabaseDefinitionError is an error.
type InvalidDatabaseDefinitionError interface {
	InvalidDefinitionError
	isEdgeDBInvalidDatabaseDefinitionError()
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

func (e *invalidDatabaseDefinitionError) isEdgeDBInvalidDatabaseDefinitionError() {}

func (e *invalidDatabaseDefinitionError) isEdgeDBInvalidDefinitionError() {}

func (e *invalidDatabaseDefinitionError) isEdgeDBSchemaDefinitionError() {}

func (e *invalidDatabaseDefinitionError) isEdgeDBQueryError() {}

func (e *invalidDatabaseDefinitionError) isEdgeDBError() {}

func (e *invalidDatabaseDefinitionError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}
// InvalidOperatorDefinitionError is an error.
type InvalidOperatorDefinitionError interface {
	InvalidDefinitionError
	isEdgeDBInvalidOperatorDefinitionError()
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

func (e *invalidOperatorDefinitionError) isEdgeDBInvalidOperatorDefinitionError() {}

func (e *invalidOperatorDefinitionError) isEdgeDBInvalidDefinitionError() {}

func (e *invalidOperatorDefinitionError) isEdgeDBSchemaDefinitionError() {}

func (e *invalidOperatorDefinitionError) isEdgeDBQueryError() {}

func (e *invalidOperatorDefinitionError) isEdgeDBError() {}

func (e *invalidOperatorDefinitionError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}
// InvalidAliasDefinitionError is an error.
type InvalidAliasDefinitionError interface {
	InvalidDefinitionError
	isEdgeDBInvalidAliasDefinitionError()
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

func (e *invalidAliasDefinitionError) isEdgeDBInvalidAliasDefinitionError() {}

func (e *invalidAliasDefinitionError) isEdgeDBInvalidDefinitionError() {}

func (e *invalidAliasDefinitionError) isEdgeDBSchemaDefinitionError() {}

func (e *invalidAliasDefinitionError) isEdgeDBQueryError() {}

func (e *invalidAliasDefinitionError) isEdgeDBError() {}

func (e *invalidAliasDefinitionError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}
// InvalidFunctionDefinitionError is an error.
type InvalidFunctionDefinitionError interface {
	InvalidDefinitionError
	isEdgeDBInvalidFunctionDefinitionError()
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

func (e *invalidFunctionDefinitionError) isEdgeDBInvalidFunctionDefinitionError() {}

func (e *invalidFunctionDefinitionError) isEdgeDBInvalidDefinitionError() {}

func (e *invalidFunctionDefinitionError) isEdgeDBSchemaDefinitionError() {}

func (e *invalidFunctionDefinitionError) isEdgeDBQueryError() {}

func (e *invalidFunctionDefinitionError) isEdgeDBError() {}

func (e *invalidFunctionDefinitionError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}
// InvalidConstraintDefinitionError is an error.
type InvalidConstraintDefinitionError interface {
	InvalidDefinitionError
	isEdgeDBInvalidConstraintDefinitionError()
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

func (e *invalidConstraintDefinitionError) isEdgeDBInvalidConstraintDefinitionError() {}

func (e *invalidConstraintDefinitionError) isEdgeDBInvalidDefinitionError() {}

func (e *invalidConstraintDefinitionError) isEdgeDBSchemaDefinitionError() {}

func (e *invalidConstraintDefinitionError) isEdgeDBQueryError() {}

func (e *invalidConstraintDefinitionError) isEdgeDBError() {}

func (e *invalidConstraintDefinitionError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}
// InvalidCastDefinitionError is an error.
type InvalidCastDefinitionError interface {
	InvalidDefinitionError
	isEdgeDBInvalidCastDefinitionError()
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

func (e *invalidCastDefinitionError) isEdgeDBInvalidCastDefinitionError() {}

func (e *invalidCastDefinitionError) isEdgeDBInvalidDefinitionError() {}

func (e *invalidCastDefinitionError) isEdgeDBSchemaDefinitionError() {}

func (e *invalidCastDefinitionError) isEdgeDBQueryError() {}

func (e *invalidCastDefinitionError) isEdgeDBError() {}

func (e *invalidCastDefinitionError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}
// DuplicateDefinitionError is an error.
type DuplicateDefinitionError interface {
	SchemaDefinitionError
	isEdgeDBDuplicateDefinitionError()
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

func (e *duplicateDefinitionError) isEdgeDBDuplicateDefinitionError() {}

func (e *duplicateDefinitionError) isEdgeDBSchemaDefinitionError() {}

func (e *duplicateDefinitionError) isEdgeDBQueryError() {}

func (e *duplicateDefinitionError) isEdgeDBError() {}

func (e *duplicateDefinitionError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}
// DuplicateModuleDefinitionError is an error.
type DuplicateModuleDefinitionError interface {
	DuplicateDefinitionError
	isEdgeDBDuplicateModuleDefinitionError()
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

func (e *duplicateModuleDefinitionError) isEdgeDBDuplicateModuleDefinitionError() {}

func (e *duplicateModuleDefinitionError) isEdgeDBDuplicateDefinitionError() {}

func (e *duplicateModuleDefinitionError) isEdgeDBSchemaDefinitionError() {}

func (e *duplicateModuleDefinitionError) isEdgeDBQueryError() {}

func (e *duplicateModuleDefinitionError) isEdgeDBError() {}

func (e *duplicateModuleDefinitionError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}
// DuplicateLinkDefinitionError is an error.
type DuplicateLinkDefinitionError interface {
	DuplicateDefinitionError
	isEdgeDBDuplicateLinkDefinitionError()
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

func (e *duplicateLinkDefinitionError) isEdgeDBDuplicateLinkDefinitionError() {}

func (e *duplicateLinkDefinitionError) isEdgeDBDuplicateDefinitionError() {}

func (e *duplicateLinkDefinitionError) isEdgeDBSchemaDefinitionError() {}

func (e *duplicateLinkDefinitionError) isEdgeDBQueryError() {}

func (e *duplicateLinkDefinitionError) isEdgeDBError() {}

func (e *duplicateLinkDefinitionError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}
// DuplicatePropertyDefinitionError is an error.
type DuplicatePropertyDefinitionError interface {
	DuplicateDefinitionError
	isEdgeDBDuplicatePropertyDefinitionError()
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

func (e *duplicatePropertyDefinitionError) isEdgeDBDuplicatePropertyDefinitionError() {}

func (e *duplicatePropertyDefinitionError) isEdgeDBDuplicateDefinitionError() {}

func (e *duplicatePropertyDefinitionError) isEdgeDBSchemaDefinitionError() {}

func (e *duplicatePropertyDefinitionError) isEdgeDBQueryError() {}

func (e *duplicatePropertyDefinitionError) isEdgeDBError() {}

func (e *duplicatePropertyDefinitionError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}
// DuplicateUserDefinitionError is an error.
type DuplicateUserDefinitionError interface {
	DuplicateDefinitionError
	isEdgeDBDuplicateUserDefinitionError()
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

func (e *duplicateUserDefinitionError) isEdgeDBDuplicateUserDefinitionError() {}

func (e *duplicateUserDefinitionError) isEdgeDBDuplicateDefinitionError() {}

func (e *duplicateUserDefinitionError) isEdgeDBSchemaDefinitionError() {}

func (e *duplicateUserDefinitionError) isEdgeDBQueryError() {}

func (e *duplicateUserDefinitionError) isEdgeDBError() {}

func (e *duplicateUserDefinitionError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}
// DuplicateDatabaseDefinitionError is an error.
type DuplicateDatabaseDefinitionError interface {
	DuplicateDefinitionError
	isEdgeDBDuplicateDatabaseDefinitionError()
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

func (e *duplicateDatabaseDefinitionError) isEdgeDBDuplicateDatabaseDefinitionError() {}

func (e *duplicateDatabaseDefinitionError) isEdgeDBDuplicateDefinitionError() {}

func (e *duplicateDatabaseDefinitionError) isEdgeDBSchemaDefinitionError() {}

func (e *duplicateDatabaseDefinitionError) isEdgeDBQueryError() {}

func (e *duplicateDatabaseDefinitionError) isEdgeDBError() {}

func (e *duplicateDatabaseDefinitionError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}
// DuplicateOperatorDefinitionError is an error.
type DuplicateOperatorDefinitionError interface {
	DuplicateDefinitionError
	isEdgeDBDuplicateOperatorDefinitionError()
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

func (e *duplicateOperatorDefinitionError) isEdgeDBDuplicateOperatorDefinitionError() {}

func (e *duplicateOperatorDefinitionError) isEdgeDBDuplicateDefinitionError() {}

func (e *duplicateOperatorDefinitionError) isEdgeDBSchemaDefinitionError() {}

func (e *duplicateOperatorDefinitionError) isEdgeDBQueryError() {}

func (e *duplicateOperatorDefinitionError) isEdgeDBError() {}

func (e *duplicateOperatorDefinitionError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}
// DuplicateViewDefinitionError is an error.
type DuplicateViewDefinitionError interface {
	DuplicateDefinitionError
	isEdgeDBDuplicateViewDefinitionError()
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

func (e *duplicateViewDefinitionError) isEdgeDBDuplicateViewDefinitionError() {}

func (e *duplicateViewDefinitionError) isEdgeDBDuplicateDefinitionError() {}

func (e *duplicateViewDefinitionError) isEdgeDBSchemaDefinitionError() {}

func (e *duplicateViewDefinitionError) isEdgeDBQueryError() {}

func (e *duplicateViewDefinitionError) isEdgeDBError() {}

func (e *duplicateViewDefinitionError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}
// DuplicateFunctionDefinitionError is an error.
type DuplicateFunctionDefinitionError interface {
	DuplicateDefinitionError
	isEdgeDBDuplicateFunctionDefinitionError()
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

func (e *duplicateFunctionDefinitionError) isEdgeDBDuplicateFunctionDefinitionError() {}

func (e *duplicateFunctionDefinitionError) isEdgeDBDuplicateDefinitionError() {}

func (e *duplicateFunctionDefinitionError) isEdgeDBSchemaDefinitionError() {}

func (e *duplicateFunctionDefinitionError) isEdgeDBQueryError() {}

func (e *duplicateFunctionDefinitionError) isEdgeDBError() {}

func (e *duplicateFunctionDefinitionError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}
// DuplicateConstraintDefinitionError is an error.
type DuplicateConstraintDefinitionError interface {
	DuplicateDefinitionError
	isEdgeDBDuplicateConstraintDefinitionError()
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

func (e *duplicateConstraintDefinitionError) isEdgeDBDuplicateConstraintDefinitionError() {}

func (e *duplicateConstraintDefinitionError) isEdgeDBDuplicateDefinitionError() {}

func (e *duplicateConstraintDefinitionError) isEdgeDBSchemaDefinitionError() {}

func (e *duplicateConstraintDefinitionError) isEdgeDBQueryError() {}

func (e *duplicateConstraintDefinitionError) isEdgeDBError() {}

func (e *duplicateConstraintDefinitionError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}
// DuplicateCastDefinitionError is an error.
type DuplicateCastDefinitionError interface {
	DuplicateDefinitionError
	isEdgeDBDuplicateCastDefinitionError()
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

func (e *duplicateCastDefinitionError) isEdgeDBDuplicateCastDefinitionError() {}

func (e *duplicateCastDefinitionError) isEdgeDBDuplicateDefinitionError() {}

func (e *duplicateCastDefinitionError) isEdgeDBSchemaDefinitionError() {}

func (e *duplicateCastDefinitionError) isEdgeDBQueryError() {}

func (e *duplicateCastDefinitionError) isEdgeDBError() {}

func (e *duplicateCastDefinitionError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}
// QueryTimeoutError is an error.
type QueryTimeoutError interface {
	QueryError
	isEdgeDBQueryTimeoutError()
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

func (e *queryTimeoutError) isEdgeDBQueryTimeoutError() {}

func (e *queryTimeoutError) isEdgeDBQueryError() {}

func (e *queryTimeoutError) isEdgeDBError() {}

func (e *queryTimeoutError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}
// ExecutionError is an error.
type ExecutionError interface {
	Error
	isEdgeDBExecutionError()
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

func (e *executionError) isEdgeDBExecutionError() {}

func (e *executionError) isEdgeDBError() {}

func (e *executionError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}
// InvalidValueError is an error.
type InvalidValueError interface {
	ExecutionError
	isEdgeDBInvalidValueError()
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

func (e *invalidValueError) isEdgeDBInvalidValueError() {}

func (e *invalidValueError) isEdgeDBExecutionError() {}

func (e *invalidValueError) isEdgeDBError() {}

func (e *invalidValueError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}
// DivisionByZeroError is an error.
type DivisionByZeroError interface {
	InvalidValueError
	isEdgeDBDivisionByZeroError()
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

func (e *divisionByZeroError) isEdgeDBDivisionByZeroError() {}

func (e *divisionByZeroError) isEdgeDBInvalidValueError() {}

func (e *divisionByZeroError) isEdgeDBExecutionError() {}

func (e *divisionByZeroError) isEdgeDBError() {}

func (e *divisionByZeroError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}
// NumericOutOfRangeError is an error.
type NumericOutOfRangeError interface {
	InvalidValueError
	isEdgeDBNumericOutOfRangeError()
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

func (e *numericOutOfRangeError) isEdgeDBNumericOutOfRangeError() {}

func (e *numericOutOfRangeError) isEdgeDBInvalidValueError() {}

func (e *numericOutOfRangeError) isEdgeDBExecutionError() {}

func (e *numericOutOfRangeError) isEdgeDBError() {}

func (e *numericOutOfRangeError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}
// IntegrityError is an error.
type IntegrityError interface {
	ExecutionError
	isEdgeDBIntegrityError()
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

func (e *integrityError) isEdgeDBIntegrityError() {}

func (e *integrityError) isEdgeDBExecutionError() {}

func (e *integrityError) isEdgeDBError() {}

func (e *integrityError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}
// ConstraintViolationError is an error.
type ConstraintViolationError interface {
	IntegrityError
	isEdgeDBConstraintViolationError()
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

func (e *constraintViolationError) isEdgeDBConstraintViolationError() {}

func (e *constraintViolationError) isEdgeDBIntegrityError() {}

func (e *constraintViolationError) isEdgeDBExecutionError() {}

func (e *constraintViolationError) isEdgeDBError() {}

func (e *constraintViolationError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}
// CardinalityViolationError is an error.
type CardinalityViolationError interface {
	IntegrityError
	isEdgeDBCardinalityViolationError()
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

func (e *cardinalityViolationError) isEdgeDBCardinalityViolationError() {}

func (e *cardinalityViolationError) isEdgeDBIntegrityError() {}

func (e *cardinalityViolationError) isEdgeDBExecutionError() {}

func (e *cardinalityViolationError) isEdgeDBError() {}

func (e *cardinalityViolationError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}
// MissingRequiredError is an error.
type MissingRequiredError interface {
	IntegrityError
	isEdgeDBMissingRequiredError()
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

func (e *missingRequiredError) isEdgeDBMissingRequiredError() {}

func (e *missingRequiredError) isEdgeDBIntegrityError() {}

func (e *missingRequiredError) isEdgeDBExecutionError() {}

func (e *missingRequiredError) isEdgeDBError() {}

func (e *missingRequiredError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}
// TransactionError is an error.
type TransactionError interface {
	ExecutionError
	isEdgeDBTransactionError()
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

func (e *transactionError) isEdgeDBTransactionError() {}

func (e *transactionError) isEdgeDBExecutionError() {}

func (e *transactionError) isEdgeDBError() {}

func (e *transactionError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}
// TransactionSerializationError is an error.
type TransactionSerializationError interface {
	TransactionError
	isEdgeDBTransactionSerializationError()
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

func (e *transactionSerializationError) isEdgeDBTransactionSerializationError() {}

func (e *transactionSerializationError) isEdgeDBTransactionError() {}

func (e *transactionSerializationError) isEdgeDBExecutionError() {}

func (e *transactionSerializationError) isEdgeDBError() {}

func (e *transactionSerializationError) HasTag(tag ErrorTag) bool {
	switch tag {
	case ShouldRetry:
		return true
	default:
		return false
	}
}
// TransactionDeadlockError is an error.
type TransactionDeadlockError interface {
	TransactionError
	isEdgeDBTransactionDeadlockError()
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

func (e *transactionDeadlockError) isEdgeDBTransactionDeadlockError() {}

func (e *transactionDeadlockError) isEdgeDBTransactionError() {}

func (e *transactionDeadlockError) isEdgeDBExecutionError() {}

func (e *transactionDeadlockError) isEdgeDBError() {}

func (e *transactionDeadlockError) HasTag(tag ErrorTag) bool {
	switch tag {
	case ShouldRetry:
		return true
	default:
		return false
	}
}
// ConfigurationError is an error.
type ConfigurationError interface {
	Error
	isEdgeDBConfigurationError()
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

func (e *configurationError) isEdgeDBConfigurationError() {}

func (e *configurationError) isEdgeDBError() {}

func (e *configurationError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}
// AccessError is an error.
type AccessError interface {
	Error
	isEdgeDBAccessError()
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

func (e *accessError) isEdgeDBAccessError() {}

func (e *accessError) isEdgeDBError() {}

func (e *accessError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}
// AuthenticationError is an error.
type AuthenticationError interface {
	AccessError
	isEdgeDBAuthenticationError()
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

func (e *authenticationError) isEdgeDBAuthenticationError() {}

func (e *authenticationError) isEdgeDBAccessError() {}

func (e *authenticationError) isEdgeDBError() {}

func (e *authenticationError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}
// ClientError is an error.
type ClientError interface {
	Error
	isEdgeDBClientError()
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

func (e *clientError) isEdgeDBClientError() {}

func (e *clientError) isEdgeDBError() {}

func (e *clientError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}
// ClientConnectionError is an error.
type ClientConnectionError interface {
	ClientError
	isEdgeDBClientConnectionError()
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

func (e *clientConnectionError) isEdgeDBClientConnectionError() {}

func (e *clientConnectionError) isEdgeDBClientError() {}

func (e *clientConnectionError) isEdgeDBError() {}

func (e *clientConnectionError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}
// ClientConnectionFailedError is an error.
type ClientConnectionFailedError interface {
	ClientConnectionError
	isEdgeDBClientConnectionFailedError()
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

	return "edgedb.ClientConnectionFailedError: " + msg
}

func (e *clientConnectionFailedError) Unwrap() error { return e.err }

func (e *clientConnectionFailedError) isEdgeDBClientConnectionFailedError() {}

func (e *clientConnectionFailedError) isEdgeDBClientConnectionError() {}

func (e *clientConnectionFailedError) isEdgeDBClientError() {}

func (e *clientConnectionFailedError) isEdgeDBError() {}

func (e *clientConnectionFailedError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}
// ClientConnectionFailedTemporarilyError is an error.
type ClientConnectionFailedTemporarilyError interface {
	ClientConnectionFailedError
	isEdgeDBClientConnectionFailedTemporarilyError()
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

	return "edgedb.ClientConnectionFailedTemporarilyError: " + msg
}

func (e *clientConnectionFailedTemporarilyError) Unwrap() error { return e.err }

func (e *clientConnectionFailedTemporarilyError) isEdgeDBClientConnectionFailedTemporarilyError() {}

func (e *clientConnectionFailedTemporarilyError) isEdgeDBClientConnectionFailedError() {}

func (e *clientConnectionFailedTemporarilyError) isEdgeDBClientConnectionError() {}

func (e *clientConnectionFailedTemporarilyError) isEdgeDBClientError() {}

func (e *clientConnectionFailedTemporarilyError) isEdgeDBError() {}

func (e *clientConnectionFailedTemporarilyError) HasTag(tag ErrorTag) bool {
	switch tag {
	case ShouldRetry:
		return true
	case ShouldReconnect:
		return true
	default:
		return false
	}
}
// ClientConnectionTimeoutError is an error.
type ClientConnectionTimeoutError interface {
	ClientConnectionError
	isEdgeDBClientConnectionTimeoutError()
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

	return "edgedb.ClientConnectionTimeoutError: " + msg
}

func (e *clientConnectionTimeoutError) Unwrap() error { return e.err }

func (e *clientConnectionTimeoutError) isEdgeDBClientConnectionTimeoutError() {}

func (e *clientConnectionTimeoutError) isEdgeDBClientConnectionError() {}

func (e *clientConnectionTimeoutError) isEdgeDBClientError() {}

func (e *clientConnectionTimeoutError) isEdgeDBError() {}

func (e *clientConnectionTimeoutError) HasTag(tag ErrorTag) bool {
	switch tag {
	case ShouldRetry:
		return true
	case ShouldReconnect:
		return true
	default:
		return false
	}
}
// ClientConnectionClosedError is an error.
type ClientConnectionClosedError interface {
	ClientConnectionError
	isEdgeDBClientConnectionClosedError()
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

	return "edgedb.ClientConnectionClosedError: " + msg
}

func (e *clientConnectionClosedError) Unwrap() error { return e.err }

func (e *clientConnectionClosedError) isEdgeDBClientConnectionClosedError() {}

func (e *clientConnectionClosedError) isEdgeDBClientConnectionError() {}

func (e *clientConnectionClosedError) isEdgeDBClientError() {}

func (e *clientConnectionClosedError) isEdgeDBError() {}

func (e *clientConnectionClosedError) HasTag(tag ErrorTag) bool {
	switch tag {
	case ShouldRetry:
		return true
	case ShouldReconnect:
		return true
	default:
		return false
	}
}
// InterfaceError is an error.
type InterfaceError interface {
	ClientError
	isEdgeDBInterfaceError()
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

func (e *interfaceError) isEdgeDBInterfaceError() {}

func (e *interfaceError) isEdgeDBClientError() {}

func (e *interfaceError) isEdgeDBError() {}

func (e *interfaceError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}
// QueryArgumentError is an error.
type QueryArgumentError interface {
	InterfaceError
	isEdgeDBQueryArgumentError()
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

func (e *queryArgumentError) isEdgeDBQueryArgumentError() {}

func (e *queryArgumentError) isEdgeDBInterfaceError() {}

func (e *queryArgumentError) isEdgeDBClientError() {}

func (e *queryArgumentError) isEdgeDBError() {}

func (e *queryArgumentError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}
// MissingArgumentError is an error.
type MissingArgumentError interface {
	QueryArgumentError
	isEdgeDBMissingArgumentError()
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

func (e *missingArgumentError) isEdgeDBMissingArgumentError() {}

func (e *missingArgumentError) isEdgeDBQueryArgumentError() {}

func (e *missingArgumentError) isEdgeDBInterfaceError() {}

func (e *missingArgumentError) isEdgeDBClientError() {}

func (e *missingArgumentError) isEdgeDBError() {}

func (e *missingArgumentError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}
// UnknownArgumentError is an error.
type UnknownArgumentError interface {
	QueryArgumentError
	isEdgeDBUnknownArgumentError()
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

func (e *unknownArgumentError) isEdgeDBUnknownArgumentError() {}

func (e *unknownArgumentError) isEdgeDBQueryArgumentError() {}

func (e *unknownArgumentError) isEdgeDBInterfaceError() {}

func (e *unknownArgumentError) isEdgeDBClientError() {}

func (e *unknownArgumentError) isEdgeDBError() {}

func (e *unknownArgumentError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}
// InvalidArgumentError is an error.
type InvalidArgumentError interface {
	QueryArgumentError
	isEdgeDBInvalidArgumentError()
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

func (e *invalidArgumentError) isEdgeDBInvalidArgumentError() {}

func (e *invalidArgumentError) isEdgeDBQueryArgumentError() {}

func (e *invalidArgumentError) isEdgeDBInterfaceError() {}

func (e *invalidArgumentError) isEdgeDBClientError() {}

func (e *invalidArgumentError) isEdgeDBError() {}

func (e *invalidArgumentError) HasTag(tag ErrorTag) bool {
	switch tag {
	default:
		return false
	}
}
// NoDataError is an error.
type NoDataError interface {
	ClientError
	isEdgeDBNoDataError()
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

func (e *noDataError) isEdgeDBNoDataError() {}

func (e *noDataError) isEdgeDBClientError() {}

func (e *noDataError) isEdgeDBError() {}

func (e *noDataError) HasTag(tag ErrorTag) bool {
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
	default:
		panic(fmt.Sprintf("invalid error code 0x%x", code))
	}
}