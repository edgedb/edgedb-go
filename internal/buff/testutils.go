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

package buff

import (
	"encoding/binary"
)

// makes a message with n 0xff bytes.
func newBenchmarkMessage(n int) []byte {
	buf := make([]byte, 5+n)
	binary.BigEndian.PutUint32(buf[1:5], uint32(4+n))
	for i := 5; i < n; i++ {
		buf[i] = 0xff
	}

	return buf
}

func newBenchmarkWriter(size int) *Writer {
	w := NewWriter()
	w.buf = make([]byte, size)[:0]
	return w
}

type writerFixture struct {
	written []byte
}

func (w *writerFixture) Write(b []byte) (int, error) {
	w.written = make([]byte, len(b))
	return copy(w.written, b), nil
}
