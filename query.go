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
	"fmt"
	"reflect"

	"github.com/edgedb/edgedb-go/internal/cardinality"
	"github.com/edgedb/edgedb-go/internal/format"
	"github.com/edgedb/edgedb-go/internal/introspect"
)

// sfQuery is a script flow query
type sfQuery struct {
	cmd     string
	headers msgHeaders
}

type msgHeaders map[uint16][]byte

// gfQuery is a granular flow query
type gfQuery struct {
	out     reflect.Value
	outType reflect.Type
	method  string
	cmd     string
	fmt     uint8
	expCard uint8
	args    []interface{}
	headers msgHeaders
}

// newQuery returns a new granular flow query.
func newQuery(
	method, cmd string,
	args []interface{},
	headers msgHeaders,
	out interface{},
) (*gfQuery, error) {
	var (
		expCard uint8
		frmt    uint8
	)

	switch method {
	case "Query":
		expCard = cardinality.Many
		frmt = format.Binary
	case "QuerySingle":
		expCard = cardinality.AtMostOne
		frmt = format.Binary
	case "QueryJSON":
		expCard = cardinality.Many
		frmt = format.JSON
	case "QuerySingleJSON":
		expCard = cardinality.AtMostOne
		frmt = format.JSON
	default:
		return nil, fmt.Errorf("unknown query method %q", method)
	}

	q := gfQuery{
		method:  method,
		cmd:     cmd,
		fmt:     frmt,
		expCard: expCard,
		args:    args,
		headers: headers,
	}

	var err error

	if frmt == format.JSON || expCard == cardinality.AtMostOne {
		q.out, err = introspect.ValueOf(out)
	} else {
		q.out, err = introspect.ValueOfSlice(out)
		if err == nil {
			q.out.SetLen(0)
		}
	}

	if err != nil {
		return &gfQuery{}, err
	}

	q.outType = q.out.Type()
	if !q.flat() {
		q.outType = q.outType.Elem()
	}

	return &q, nil
}

func (q *gfQuery) flat() bool {
	if q.expCard != cardinality.Many {
		return true
	}

	if q.fmt == format.JSON {
		return true
	}

	return false
}

type queryable interface {
	headers() msgHeaders
	granularFlow(context.Context, *gfQuery) error
}

func runQuery(
	ctx context.Context,
	c queryable,
	method, cmd string,
	out interface{},
	args []interface{},
) error {
	q, err := newQuery(method, cmd, args, c.headers(), out)
	if err != nil {
		return err
	}

	return c.granularFlow(ctx, q)
}
