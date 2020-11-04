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
	"github.com/edgedb/edgedb-go/edgedb/protocol/cardinality"
	"github.com/edgedb/edgedb-go/edgedb/protocol/format"
)

type query struct {
	cmd     string
	fmt     uint8
	expCard uint8
	args    []interface{}
}

func (q *query) flat() bool {
	if q.expCard != cardinality.Many {
		return true
	}

	if q.fmt == format.JSON {
		return true
	}

	return false
}
