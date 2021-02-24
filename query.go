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
	"reflect"

	"github.com/edgedb/edgedb-go/internal/cardinality"
	"github.com/edgedb/edgedb-go/internal/format"
	"github.com/edgedb/edgedb-go/internal/marshal"
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
	cmd     string
	fmt     uint8
	expCard uint8
	args    []interface{}
	headers msgHeaders
}

// newQuery returns a new granular flow query.
func newQuery(
	cmd string,
	fmt, expCard uint8,
	args []interface{},
	headers msgHeaders,
	out interface{},
) (*gfQuery, error) {
	q := gfQuery{
		cmd:     cmd,
		fmt:     fmt,
		expCard: expCard,
		args:    args,
		headers: headers,
	}

	var err error

	if fmt == format.JSON || expCard == cardinality.One {
		q.out, err = marshal.ValueOf(out)
	} else {
		q.out, err = marshal.ValueOfSlice(out)
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
