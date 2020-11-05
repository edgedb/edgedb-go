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

package codecs

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCacheGetMissingKey(t *testing.T) {
	cache := MakeCache(1)
	result := cache.Get(CacheKey{})
	assert.Nil(t, result)
}

func TestCachePutNewKey(t *testing.T) {
	cache := MakeCache(1)
	key := CacheKey{}
	require.Nil(t, cache.Get(key))

	pair := &CodecPair{JSONBytes, JSONBytes}
	cache.Put(key, pair)
	result := cache.Get(key)

	require.Equal(t, pair, result)
}

func TestCachePutMoreThanCapacity(t *testing.T) {
	cache := MakeCache(1)
	k0 := CacheKey{Format: 0}
	k1 := CacheKey{Format: 1}
	pair := &CodecPair{JSONBytes, JSONBytes}

	cache.Put(k0, pair)
	cache.Put(k1, pair)

	require.Nil(t, cache.Get(k0))
	assert.Equal(t, cache.Get(k1), pair)
}
