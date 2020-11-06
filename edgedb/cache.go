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

	"github.com/edgedb/edgedb-go/edgedb/cache"
	"github.com/edgedb/edgedb-go/edgedb/protocol/codecs"
	"github.com/edgedb/edgedb-go/edgedb/types"
)

var descCache = cache.New(1_000)

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

type codecKey struct {
	inID    types.UUID
	outID   types.UUID
	outType reflect.Type
}

func (c *Client) getCodecs(
	ids idPair,
	tp reflect.Type,
) (cdcs codecPair, ok bool) {
	key := codecKey{inID: ids.in, outID: ids.out, outType: tp}
	out, ok := c.codecCache.Get(key)
	if !ok {
		return cdcs, false
	}

	key = codecKey{inID: ids.in, outID: ids.out}
	in, ok := c.codecCache.Get(key)
	if !ok {
		return cdcs, false
	}

	cdcs.out = out.(codecs.Codec)
	cdcs.in = in.(codecs.Codec)
	return cdcs, false
}

func (c *Client) putCodecs(ids idPair, tp reflect.Type, cdcs codecPair) {
	key := codecKey{inID: ids.in, outID: ids.out, outType: tp}
	c.codecCache.Put(key, cdcs.out)

	key = codecKey{inID: ids.in, outID: ids.out}
	c.codecCache.Put(key, cdcs.in)
}

func getDescriptors(ids idPair) (descs descPair, ok bool) {
	val, ok := descCache.Get(ids.in)
	if !ok {
		return descs, ok
	}
	descs.in = val.([]byte)

	val, ok = descCache.Get(ids.out)
	if !ok {
		return descs, ok
	}

	descs.out = val.([]byte)
	return descs, ok
}

func putDescriptors(ids idPair, descs descPair) {
	descCache.Put(ids.in, descs.in)
	descCache.Put(ids.out, descs.out)
}
