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

type node struct {
	key  interface{}
	val  interface{}
	prev *node
	next *node
}

// Cache is a thread safe LRU cache with O(1) time operations.
type Cache struct {
	cap int
	ch  chan map[interface{}]*node

	// root.prev is the tail
	// root.next is the head
	root node
}

// New returns a new cache.
func New(cap int) *Cache {
	ch := make(chan map[interface{}]*node, 1)
	ch <- make(map[interface{}]*node, cap)

	c := Cache{cap: cap, ch: ch}
	c.root.next = &c.root
	c.root.prev = &c.root

	return &c
}

// Get returns a value from the cache.
func (c *Cache) Get(id interface{}) (interface{}, bool) {
	// ensure map is only used by one go routine at a time.
	m := <-c.ch
	defer func() { c.ch <- m }()

	if n, ok := m[id]; ok {
		c.moveToFront(n)
		return n.val, true
	}

	return nil, false
}

// Put adds a value to the cache.
func (c *Cache) Put(key interface{}, val interface{}) {
	// ensure map is only used by one go routine at a time.
	m := <-c.ch
	defer func() { c.ch <- m }()

	if n, ok := m[key]; ok {
		n.val = val
		c.moveToFront(n)
		return
	}

	for len(m) >= c.cap {
		oldest := c.root.prev
		delete(m, oldest.key)
		c.remove(oldest)
	}

	n := &node{key: key, val: val}
	c.pushFront(n)
	m[key] = n
}

func (c *Cache) moveToFront(n *node) {
	n.prev.next = n.next
	n.next.prev = n.prev

	n.prev = &c.root
	n.next = c.root.next
	n.prev.next = n
	n.next.prev = n
}

func (c *Cache) remove(n *node) {
	n.prev.next = n.next
	n.next.prev = n.prev

	// avoid memory leaks
	n.prev = nil
	n.next = nil
}

func (c *Cache) pushFront(n *node) {
	n.prev = &c.root
	n.next = c.root.next
	n.prev.next = n
	n.next.prev = n
}
