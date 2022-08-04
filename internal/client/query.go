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
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"reflect"

	types "github.com/edgedb/edgedb-go/internal/edgedbtypes"
	"github.com/edgedb/edgedb-go/internal/header"
	"github.com/edgedb/edgedb-go/internal/introspect"
)

type query struct {
	out          reflect.Value
	outType      reflect.Type
	method       string
	cmd          string
	fmt          Format
	expCard      Cardinality
	args         []interface{}
	capabilities uint64
	state        map[string]interface{}
}

func (q *query) flat() bool {
	if q.expCard != Many {
		return true
	}

	if q.fmt == JSON {
		return true
	}

	return false
}

func (q *query) headers0pX() header.Header {
	bts := make([]byte, 8)
	binary.BigEndian.PutUint64(bts, q.capabilities)

	return header.Header{header.AllowCapabilities: bts}
}

// newQuery returns a new granular flow query.
func newQuery(
	method, cmd string,
	args []interface{},
	capabilities uint64,
	state map[string]interface{},
	out interface{},
) (*query, error) {
	var (
		expCard Cardinality
		frmt    Format
	)

	switch method {
	case "Execute":
		return &query{
			method:       method,
			cmd:          cmd,
			fmt:          Null,
			expCard:      Many,
			args:         args,
			capabilities: capabilities,
			state:        state,
		}, nil
	case "Query":
		expCard = Many
		frmt = Binary
	case "QuerySingle":
		expCard = AtMostOne
		frmt = Binary
	case "QueryJSON":
		expCard = Many
		frmt = JSON
	case "QuerySingleJSON":
		expCard = AtMostOne
		frmt = JSON
	default:
		return nil, fmt.Errorf("unknown query method %q", method)
	}

	q := query{
		method:       method,
		cmd:          cmd,
		fmt:          frmt,
		expCard:      expCard,
		args:         args,
		capabilities: capabilities,
		state:        state,
	}

	var err error

	if frmt == JSON || expCard == AtMostOne {
		q.out, err = introspect.ValueOf(out)
	} else {
		q.out, err = introspect.ValueOfSlice(out)
		if err == nil {
			q.out.SetLen(0)
		}
	}

	if err != nil {
		return &query{}, &interfaceError{err: err}
	}

	q.outType = q.out.Type()
	if !q.flat() {
		q.outType = q.outType.Elem()
	}

	return &q, nil
}

type queryable interface {
	capabilities1pX() uint64
	granularFlow(context.Context, *query) error
}

type unseter interface {
	Unset()
}

func runQuery(
	ctx context.Context,
	c queryable,
	method, cmd string,
	out interface{},
	args []interface{},
	state map[string]interface{},
) error {
	if method == "QuerySingleJSON" {
		switch out.(type) {
		case *[]byte, *types.OptionalBytes:
		default:
			return &interfaceError{msg: fmt.Sprintf(
				`the "out" argument must be *[]byte or *OptionalBytes, got %T`,
				out)}
		}
	}

	q, err := newQuery(method, cmd, args, c.capabilities1pX(), state, out)
	if err != nil {
		return err
	}

	err = c.granularFlow(ctx, q)

	var edbErr Error
	if errors.As(err, &edbErr) &&
		edbErr.Category(NoDataError) &&
		(q.method == "QuerySingle" || q.method == "QuerySingleJSON") {
		if opt, ok := out.(unseter); ok {
			opt.Unset()
			return nil
		}
	}

	return err
}

func copyState(in map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(in))

	for k, v := range in {
		switch val := v.(type) {
		case map[string]interface{}:
			out[k] = copyState(val)
		case []interface{}:
			out[k] = copyStateSlice(val)
		default:
			out[k] = val
		}
	}

	return out
}

func copyStateSlice(in []interface{}) []interface{} {
	out := make([]interface{}, len(in))

	for i, v := range in {
		switch val := v.(type) {
		case map[string]interface{}:
			out[i] = copyState(val)
		case []interface{}:
			out[i] = copyStateSlice(val)
		default:
			out[i] = val
		}
	}

	return out
}
