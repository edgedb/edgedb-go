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

package protocol

import (
	"testing"
)

func BenchmarkPushUint16(b *testing.B) {
	data := [100]byte{}
	var n uint16 = 371

	for i := 0; i < b.N; i++ {
		msg := data[:0]
		PushUint16(&msg, n)
	}
}

func BenchmarkPushUint32(b *testing.B) {
	data := [100]byte{}
	var n uint32 = 2_147_483_647

	for i := 0; i < b.N; i++ {
		msg := data[:0]
		PushUint32(&msg, n)
	}
}

func BenchmarkPushUint64(b *testing.B) {
	data := [100]byte{}
	var n uint64 = 9_223_372_036_854_775_807

	for i := 0; i < b.N; i++ {
		msg := data[:0]
		PushUint64(&msg, n)
	}
}

func BenchmarkPopBytes(b *testing.B) {
	data := [12]byte{
		0, 0, 0, 8,
		1, 2, 3, 4,
		5, 6, 7, 8,
	}

	for i := 0; i < b.N; i++ {
		msg := data[:]
		PopBytes(&msg)
	}
}

func BenchmarkPopMessage(b *testing.B) {
	data := [22]byte{
		'C',
		0, 0, 0, 16,
		0, 0, 0, 0,
		0, 0, 0, 0,
		0, 0, 0, 0,
		'S',
		0, 0, 0, 4,
	}
	for i := 0; i < b.N; i++ {
		msg := data[:]
		PopMessage(&msg)
	}
}
