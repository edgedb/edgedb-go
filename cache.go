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

/*
All caches should be LRU with max items set to 1000 or so.

descriptor cache (shared globally) mapping:
Type ID -> Type Descriptor

codec cache (conn/pool specific) mapping:
(Type ID, Go Type Ref) -> Codec

type id cache (conn/pool) mapping:
(Query, Expected Cardinality, IO Format) -> (In Type ID, Out Type ID)

Optimistic execute flow:
1. check type id cache for (eql, expCard, format).
	- if cache miss then do prepare/execute flow instead of optimistic

2. check codec cache for (out Type ID, out.Type()).
	- if cache miss then check descriptor cache for out Type ID
		- if cache miss then do prepare/execute flow instead of optimistic
		- else build typed codec for out.Type()

3. use codecs in optimistic execute
*/

import (
	"reflect"

	"github.com/edgedb/edgedb-go/edgedb/cache"
	"github.com/edgedb/edgedb-go/edgedb/protocol/codecs"
	"github.com/edgedb/edgedb-go/edgedb/types"
)

var descCache = cache.New(1_000)

type codecKey struct {
	ID   types.UUID
	Type reflect.Type
}

type codecPair struct {
	in  codecs.Codec
	out codecs.Codec
}

type descPair struct {
	in  []byte
	out []byte
}

type idPair struct {
	in  types.UUID
	out types.UUID
}

type queryKey struct {
	cmd     string
	fmt     uint8
	expCard uint8
}

func (c *Client) getTypeIDs(q query) (idPair, bool) {
	key := queryKey{
		cmd:     q.cmd,
		fmt:     q.fmt,
		expCard: q.expCard,
	}

	if val, ok := c.typeIDCache.Get(key); ok {
		return val.(idPair), true
	}

	return idPair{}, false
}

func (c *Client) putTypeIDs(q query, ids idPair) {
	key := queryKey{
		cmd:     q.cmd,
		fmt:     q.fmt,
		expCard: q.expCard,
	}
	c.typeIDCache.Put(key, ids)
}
