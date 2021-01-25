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

// Package soc has utilities for working with sockets.
package soc

import (
	"errors"
	"net"
)

const minChunkSize = 5

// Data is the bytes & error that were read from a socket.
// When data's bytes are no longer used data must be released.
type Data struct {
	Buf     []byte
	Err     error
	release func()
}

// Release frees the underlying array to be reused.
func (d *Data) Release() {
	if d.release != nil {
		d.release()
	}
}

// IsPermanentNetErr returns true if the error is not a temporary net.Error.
func IsPermanentNetErr(err error) bool {
	if err == nil {
		return false
	}

	var netErr net.Error
	if errors.As(err, &netErr) {
		return !netErr.Temporary()
	}

	return false
}

// Read reads a socket sending the read data to toBeDeserialized.
// This should be run in it's own go routine.
func Read(conn net.Conn, freeMemory *MemPool, toBeDeserialized chan *Data) {
	mkRelease := func(slab []byte) func() {
		return func() {
			freeMemory.Release(slab)
		}
	}

	for {
		slab := freeMemory.Acquire()
		buf := slab

		for len(buf) >= minChunkSize {
			n, err := conn.Read(buf)

			data := &Data{Buf: buf[:n:n]}
			buf = buf[n:]

			// releasing the last chunk of data written to the slab
			// releases the whole slab.
			if err != nil || len(buf) < minChunkSize {
				data.release = mkRelease(slab)
			}

			toBeDeserialized <- data

			if err != nil {
				toBeDeserialized <- &Data{Err: err}
			}

			// shut down on permanent error
			if IsPermanentNetErr(err) {
				return
			}
		}
	}
}
