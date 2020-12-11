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

package edgedb

import (
	"errors"
	"fmt"
	"log"

	"github.com/edgedb/edgedb-go/internal/buff"
)

var (
	// Error is wrapped by all edgedb errors.
	Error error = errors.New("")

	// ErrReleasedTwice is returned if a PoolConn is released more than once.
	ErrReleasedTwice = fmt.Errorf(
		"connection released more than once%w", Error,
	)

	// ErrZeroResults is returned when a query has no results.
	ErrZeroResults = fmt.Errorf("zero results%w", Error)

	// ErrPoolClosed is returned by operations on closed pools.
	ErrPoolClosed error = fmt.Errorf("pool closed%w", Error)

	// ErrContextExpired is returned when an expired context is used.
	ErrContextExpired error = fmt.Errorf("context expired%w", Error)

	// ErrBadConfig is wrapped
	// when a function returning Options encounters an error.
	ErrBadConfig error = fmt.Errorf("%w", Error)

	// ErrClientFault ...
	ErrClientFault error = fmt.Errorf("%w", Error)

	// ErrInterfaceViolation ...
	ErrInterfaceViolation error = fmt.Errorf("%w", ErrClientFault)
)

func decodeError(r *buff.Reader) error {
	r.Discard(5) // skip severity & code
	err := fmt.Errorf("%v%w", r.PopString(), Error)

	n := int(r.PopUint16())
	headers := make(map[uint16]string, n)

	for i := 0; i < n; i++ {
		headers[r.PopUint16()] = r.PopString()
	}

	// todo do something with headers
	log.Println(headers)
	return err
}

type wrappedManyError struct {
	msg     string
	wrapped []error
}

func (e *wrappedManyError) Error() string {
	return e.msg
}

func (e *wrappedManyError) Is(target error) bool {
	for _, err := range e.wrapped {
		if errors.Is(err, target) {
			return true
		}
	}

	return false
}

func wrapAll(errs ...error) error {
	err := &wrappedManyError{}
	for _, e := range errs {
		if e != nil {
			err.wrapped = append(err.wrapped, e)
		}
	}

	if len(err.wrapped) == 0 {
		return nil
	}

	if len(err.wrapped) == 1 {
		return err.wrapped[0]
	}

	err.msg = err.wrapped[0].Error()
	for _, e := range err.wrapped[1:] {
		err.msg += "; " + e.Error()
	}

	return err
}
