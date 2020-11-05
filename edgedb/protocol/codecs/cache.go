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

import "reflect"

// CacheKey is used to look up codecs in the cache.
type CacheKey struct {
	Command      string
	Format       uint8
	ExpectedCard uint8
	Type         reflect.Type
}

// CodecPair is used to cache in & out codec pairs
type CodecPair struct {
	In  Codec
	Out Codec
}

type node struct {
	key    CacheKey
	codecs *CodecPair
	prev   *node
	next   *node
}

type cache struct {
	cap int
	m   map[CacheKey]*node

	// root.prev is the tail
	// root.next is the head
	root node
}

func makeCache(cap int) *cache {
	c := cache{
		cap: cap,
		m:   make(map[CacheKey]*node, cap),
	}

	c.root.next = &c.root
	c.root.prev = &c.root
	return &c
}

func (c *cache) get(key CacheKey) *CodecPair {
	if node, ok := c.m[key]; ok {
		c.moveToFront(node)
		return node.codecs
	}

	return nil
}

func (c *cache) put(key CacheKey, codecs *CodecPair) {
	if n, ok := c.m[key]; ok {
		n.codecs = codecs
		c.moveToFront(n)
		return
	}

	if len(c.m) >= c.cap {
		oldest := c.root.prev
		delete(c.m, oldest.key)
		c.remove(oldest)
	}

	n := &node{key: key, codecs: codecs}
	c.pushFront(n)
	c.m[key] = n
}

func (c *cache) moveToFront(n *node) {
	n.prev.next = n.next
	n.next.prev = n.prev

	n.prev = &c.root
	n.next = c.root.next
	n.prev.next = n
	n.next.prev = n
}

func (c *cache) remove(n *node) {
	n.prev.next = n.next
	n.next.prev = n.prev

	// avoid memory leaks
	n.prev = nil
	n.next = nil
}

func (c *cache) pushFront(n *node) {
	n.prev = &c.root
	n.next = c.root.next
	n.prev.next = n
	n.next.prev = n
}

// Cache is a thread safe LRU codec cache.
type Cache struct {
	ch chan *cache
}

// MakeCache returns a new Cache.
func MakeCache(cap int) Cache {
	s := Cache{ch: make(chan *cache, 1)}
	s.ch <- makeCache(cap)
	return s
}

// Get returns the key's codec or nil if key is not present.
func (s *Cache) Get(key CacheKey) *CodecPair {
	cache := <-s.ch
	defer func() { s.ch <- cache }()

	return cache.get(key)
}

// Put adds codec to the cache.
func (s *Cache) Put(key CacheKey, codecs *CodecPair) {
	cache := <-s.ch
	defer func() { s.ch <- cache }()

	cache.put(key, codecs)
}
