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

package cache

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCacheGetMissingKey(t *testing.T) {
	cache := New(1)
	val, ok := cache.Get("key")
	assert.False(t, ok)
	assert.Nil(t, val)
}

func TestCachePutNewKey(t *testing.T) {
	cache := New(1)

	cache.Put("key", "val")
	val, ok := cache.Get("key")

	require.True(t, ok)
	assert.Equal(t, "val", val)
}

func TestCachePutMoreThanCapacity(t *testing.T) {
	cache := New(1)

	cache.Put(1, "one")
	cache.Put(2, "two")

	val, ok := cache.Get(1)
	require.False(t, ok)
	require.Nil(t, val)

	val, ok = cache.Get(2)
	require.True(t, ok)
	assert.Equal(t, "two", val)
}

func TestCachePutConcurencySafe(t *testing.T) {
	// running this test with the race detector enabled
	// is likely to expose race conditions.
	// https://golang.org/doc/articles/race_detector.html

	done := make(chan struct{}, 20)
	cache := New(10)

	for i := 0; i < 20; i++ {
		go func() {
			for i := 0; i < 1000; i++ {
				cache.Put("key", "val")
			}
			done <- struct{}{}
		}()
	}

	for i := 0; i < 20; i++ {
		<-done
	}
}
