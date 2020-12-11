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

func repeatingBenchmarkData(n int, d []byte) []byte {
	buf := make([]byte, n*len(d))

	for i := 0; i < n*len(d); i += len(d) {
		copy(buf[i:], d)
	}

	return buf
}

type writeFixture struct {
	written []byte
}

func (w *writeFixture) Write(b []byte) (int, error) {
	if len(w.written) > 0 {
		panic("Write called more than once")
	}
	w.written = make([]byte, len(b))
	return copy(w.written, b), nil
}
