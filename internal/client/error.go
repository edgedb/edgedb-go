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

package gel

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"syscall"

	"github.com/edgedb/edgedb-go/internal/buff"
)

var (
	errNoTOMLFound             = errors.New("no gel.toml found")
	errZeroResults       error = &noDataError{msg: "zero results"}
	errStateNotSupported       = &interfaceError{msg: "client methods " +
		"WithConfig, WithGlobals, and WithModuleAliases " +
		"are not supported by the server. " +
		"Upgrade your server to version 2.0 or greater " +
		"to use these features."}
)

// ErrorTag is the argument type to Error.HasTag().
type ErrorTag string

// ErrorCategory values represent Gel's error types.
type ErrorCategory string

// Error is the error type returned from gel.
type Error interface {
	Error() string
	Unwrap() error

	// HasTag returns true if the error is marked with the supplied tag.
	HasTag(ErrorTag) bool

	// Category returns true if the error is in the provided category.
	Category(ErrorCategory) bool
}

// firstError returns the first non nil error or nil.
func firstError(a, b error) error {
	if a != nil {
		return a
	}

	return b
}

const (
	hint          = 0x0001
	positionStart = 0xfff1
	lineStart     = 0xfff3
)

func positionFromHeaders(headers map[uint16]string) (*int, *int, error) {
	lineNoRaw, ok := headers[lineStart]
	if !ok {
		return nil, nil, nil
	}

	byteNoRaw, ok := headers[positionStart]
	if !ok {
		return nil, nil, nil
	}

	lineNo, err := strconv.Atoi(lineNoRaw)
	if err != nil {
		return nil, nil, &binaryProtocolError{
			err: fmt.Errorf("decode lineNo: %q: %w", lineNoRaw, err),
		}
	}
	byteNo, err := strconv.Atoi(byteNoRaw)
	if err != nil {
		return nil, nil, &binaryProtocolError{
			err: fmt.Errorf("decode byteNo: %q: %w", byteNoRaw, err),
		}
	}

	return &lineNo, &byteNo, nil
}

// decodeErrorResponseMsg decodes an error response
// https://www.edgedb.com/docs/internals/protocol/messages#errorresponse
func decodeErrorResponseMsg(r *buff.Reader, query string) error {
	r.Discard(1) // severity
	w := Warning{
		Code:    r.PopUint32(),
		Message: r.PopString(),
	}

	n := int(r.PopUint16())
	headers := make(map[uint16]string, n)
	for i := 0; i < n; i++ {
		headers[r.PopUint16()] = r.PopString()
	}

	var err error
	w.Line, w.Start, err = positionFromHeaders(headers)
	if err != nil {
		return errors.Join(w.Err(query), err)
	}

	w.Hint = headers[hint]
	return w.Err(query)
}

type wrappedManyError struct {
	msg  string
	errs []error
}

func (e *wrappedManyError) Error() string {
	return e.msg
}

func (e *wrappedManyError) Is(target error) bool {
	for _, err := range e.errs {
		if errors.Is(err, target) {
			return true
		}
	}

	return false
}

func (e *wrappedManyError) As(target interface{}) bool {
	for _, err := range e.errs {
		if errors.As(err, target) {
			return true
		}
	}

	return false
}

func wrapAll(errs ...error) error {
	err := &wrappedManyError{}
	for _, e := range errs {
		if e != nil {
			err.errs = append(err.errs, e)
		}
	}

	if len(err.errs) == 0 {
		return nil
	}

	if len(err.errs) == 1 {
		return err.errs[0]
	}

	err.msg = err.errs[0].Error()
	for _, e := range err.errs[1:] {
		err.msg += "; " + e.Error()
	}

	return err
}

func isClientConnectionError(err error) bool {
	var edbErr Error
	return errors.As(err, &edbErr) && edbErr.Category(ClientConnectionError)
}

func wrapNetError(err error) error {
	var errEDB Error
	var errNetOp *net.OpError
	var errDSN *net.DNSError

	switch {
	case err == nil:
		return err
	case errors.As(err, &errEDB):
		return err

	case errors.Is(err, context.Canceled):
		fallthrough
	case errors.Is(err, context.DeadlineExceeded):
		fallthrough
	case errors.As(err, &errNetOp) && errNetOp.Timeout():
		return &clientConnectionTimeoutError{err: err}

	case errors.Is(err, io.EOF):
		fallthrough
	case errors.Is(err, syscall.ECONNREFUSED):
		fallthrough
	case errors.Is(err, syscall.ECONNABORTED):
		fallthrough
	case errors.Is(err, syscall.ECONNRESET):
		fallthrough
	case errors.Is(err, syscall.EADDRINUSE):
		fallthrough
	case errors.As(err, &errDSN):
		fallthrough
	case errors.Is(err, syscall.ENOENT):
		return &clientConnectionFailedTemporarilyError{err: err}

	case errors.Is(err, net.ErrClosed):
		return &clientConnectionClosedError{err: err}

	default:
		return &clientConnectionFailedError{err: err}
	}
}

func invalidTLSSecurity(val string) error {
	return fmt.Errorf(
		"invalid TLSSecurity value: expected one of %v, got: %q",
		englishList(
			[]string{"insecure", "no_host_verification", "strict"},
			"or"),
		val,
	)
}
