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

/*
All caches should be LRU with max items set to 1000 or so.

descriptor cache (shared globally) mapping:
Type ID -> Type Descriptor

codec cache (conn/pool specific) mapping:
(Type ID, Go Type Ref) -> Codec

type id cache (conn/pool specific) mapping:
(Query, Expected Cardinality, IO Format, outType) -> (In Type ID, Out Type ID)

capabilities cache (conn/pool specific) mapping:
(Query, Expected Cardinality, IO Format, outType) -> capabilities

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
	"encoding/binary"
	"reflect"

	"github.com/geldata/gel-go/internal/codecs"
	types "github.com/geldata/gel-go/internal/geltypes"
	"github.com/geldata/gel-go/internal/header"
)

type codecKey struct {
	ID   types.UUID
	Type reflect.Type
}

type codecPair struct {
	in  codecs.Encoder
	out codecs.Decoder
}

type idPair struct {
	in  types.UUID
	out types.UUID
}

type queryKey struct {
	lang    Language
	cmd     string
	fmt     Format
	expCard Cardinality
	outType reflect.Type
}

func makeKey(q *query) queryKey {
	return queryKey{
		lang:    q.lang,
		cmd:     q.cmd,
		fmt:     q.fmt,
		expCard: q.expCard,
		outType: q.outType,
	}
}

func (c *protocolConnection) getCachedTypeIDs(q *query) (*idPair, bool) {
	if val, ok := c.typeIDCache.Get(makeKey(q)); ok {
		x := val.(idPair)
		return &x, true
	}

	return nil, false
}

func (c *protocolConnection) cacheTypeIDs(q *query, ids idPair) {
	c.typeIDCache.Put(makeKey(q), ids)
}

func (c *protocolConnection) cacheCapabilities0pX(
	q *query,
	headers header.Header0pX,
) {
	if capabilities, ok := headers[header.Capabilities]; ok {
		x := binary.BigEndian.Uint64(capabilities)
		if x&capabilitiesDDL != 0 {
			c.typeIDCache.Invalidate()
			c.inCodecCache.Invalidate()
			c.outCodecCache.Invalidate()
			c.capabilitiesCache.Invalidate()
		}
		c.capabilitiesCache.Put(makeKey(q), x)
	}
}

func (c *protocolConnection) cacheCapabilities1pX(
	q *query,
	capabilities uint64,
) {
	if capabilities&capabilitiesDDL != 0 {
		c.typeIDCache.Invalidate()
		c.inCodecCache.Invalidate()
		c.outCodecCache.Invalidate()
		c.capabilitiesCache.Invalidate()
	}
	c.capabilitiesCache.Put(makeKey(q), capabilities)
}

func (c *reconnectingConn) getCachedCapabilities(q *query) (uint64, bool) {
	if val, ok := c.capabilitiesCache.Get(makeKey(q)); ok {
		x := val.(uint64)
		return x, true
	}

	return 0, false
}
