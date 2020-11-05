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
	"fmt"
	"reflect"

	"github.com/edgedb/edgedb-go/edgedb/cache"
	"github.com/edgedb/edgedb-go/edgedb/protocol/codecs"
	"github.com/edgedb/edgedb-go/edgedb/types"
)

var (
	// don't use these caches directly.
	// instead use the utility functions below.

	// todo what should codec cache capacity be?
	codecCache     = cache.New(100)
	descQueryCache = cache.New(100)
	descCache      = cache.New(200)
)

type codecPair struct {
	In  codecs.Codec
	Out codecs.Codec
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
	tp      reflect.Type
}

func newQueryKey(q query, tp reflect.Type) queryKey {
	return queryKey{
		cmd:     q.cmd,
		fmt:     q.fmt,
		expCard: q.expCard,
		tp:      tp,
	}
}

func getCodecs(q query, tp reflect.Type) (cdcs codecPair, ok bool) {
	key := newQueryKey(q, tp)
	val, ok := codecCache.Get(key)

	if !ok {
		return cdcs, ok
	}

	return val.(codecPair), ok
}

func putCodecs(q query, tp reflect.Type, cdcs codecPair) {
	key := newQueryKey(q, tp)
	fmt.Println(key)
	codecCache.Put(key, cdcs)
}

func getDescriptors(q query) (descs descPair, ok bool) {
	key := newQueryKey(q, nil)
	val, ok := descQueryCache.Get(key)

	if !ok {
		return descs, ok
	}

	return val.(descPair), ok
}

func putDescriptors(q query, descs descPair) {
	key := newQueryKey(q, nil)
	descQueryCache.Put(key, descs)
}

func getDescriptorsByID(ids idPair) (descs descPair, ok bool) {
	inVal, ok := descCache.Get(ids.in)
	if !ok {
		return descs, ok
	}

	outVal, ok := descCache.Get(ids.out)
	if !ok {
		return descs, ok
	}

	descs.in = inVal.([]byte)
	descs.out = outVal.([]byte)
	return descs, ok
}

func putDescriptorsByID(ids idPair, pair descPair) {
	descCache.Put(ids.in, pair.in)
	descCache.Put(ids.out, pair.out)
}
