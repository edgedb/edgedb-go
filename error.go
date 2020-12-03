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

	"github.com/edgedb/edgedb-go/protocol/buff"
)

var (
	// Error is wrapped by all errors returned from the server.
	Error = errors.New("")
	// todo error API (hierarchy and wrap all returned errors)

	// ErrReleasedTwice is returned if a PoolConn is released more than once.
	ErrReleasedTwice = fmt.Errorf(
		"connection released more than once%w", Error,
	)

	// ErrorZeroResults is returned when a query has no results.
	ErrorZeroResults = fmt.Errorf("zero results%w", Error)

	// ErrorPoolClosed is returned by operations on closed pools.
	ErrorPoolClosed error = fmt.Errorf("pool closed%w", Error)

	// ErrorConnsInUse is returned when all connects are in use.
	ErrorConnsInUse error = fmt.Errorf("all connections in use%w", Error)

	// ErrorContextExpired is returned when an expired context is used.
	ErrorContextExpired error = fmt.Errorf("context expired%w", Error)

	// ErrorConfiguration is returned when invalid configuration is received.
	ErrorConfiguration error = fmt.Errorf("%w", Error)
)

func decodeError(buf *buff.Buff) error {
	buf.Discard(5) // skip severity & code
	return fmt.Errorf("%v%w", buf.PopString(), Error)
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

	err.msg = err.wrapped[0].Error()
	for _, e := range err.wrapped[1:] {
		err.msg += "; " + e.Error()
	}

	return err
}
