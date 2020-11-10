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

	"github.com/edgedb/edgedb-go/edgedb/types"
	"github.com/stretchr/testify/assert"
)

func TestPopUint8(t *testing.T) {
	bts := []byte{10}
	result := PopUint8(&bts)

	assert.Equal(t, uint8(10), result)
	assert.Equal(t, []byte{}, bts)
}

func TestPushUint8(t *testing.T) {
	bts := []byte{}
	PushUint8(&bts, 7)

	assert.Equal(t, []byte{7}, bts)
}

func TestPopUint16(t *testing.T) {
	bts := []byte{0xa, 0x3}
	result := PopUint16(&bts)

	assert.Equal(t, uint16(0xa03), result)
	assert.Equal(t, []byte{}, bts)
}

func TestPushUint16(t *testing.T) {
	bts := []byte{}
	PushUint16(&bts, 0xa03)

	assert.Equal(t, []byte{0xa, 0x3}, bts)
}

func TestPopUint32(t *testing.T) {
	bts := []byte{0, 0, 0, 37}
	result := PopUint32(&bts)

	assert.Equal(t, uint32(37), result)
	assert.Equal(t, []byte{}, bts)
}

func TestPeekUint32(t *testing.T) {
	bts := []byte{1, 2, 3, 4}
	result := PeekUint32(&bts)

	assert.Equal(t, uint32(0x1020304), result)
	assert.Equal(t, []byte{1, 2, 3, 4}, bts)
}

func TestPushUint32(t *testing.T) {
	bts := []byte{}
	PushUint32(&bts, 0x1020304)

	assert.Equal(t, []byte{1, 2, 3, 4}, bts)
}

func TestPopUint64(t *testing.T) {
	bts := []byte{1, 2, 3, 4, 5, 6, 7, 8}
	result := PopUint64(&bts)

	assert.Equal(t, uint64(0x102030405060708), result)
	assert.Equal(t, []byte{}, bts)
}

func TestPushUint64(t *testing.T) {
	bts := []byte{}
	PushUint64(&bts, 0x102030405060708)

	expected := []byte{1, 2, 3, 4, 5, 6, 7, 8}
	assert.Equal(t, expected, bts)
}

func TestPopBytes(t *testing.T) {
	bts := []byte{0, 0, 0, 1, 32}
	result := PopBytes(&bts)

	assert.Equal(t, []byte{32}, result)
	assert.Equal(t, []byte{}, bts)
}

func TestPushBytes(t *testing.T) {
	bts := []byte{}
	PushBytes(&bts, []byte{7, 5})

	expected := []byte{0, 0, 0, 2, 7, 5}
	assert.Equal(t, expected, bts)
}

func TestPopString(t *testing.T) {
	bts := []byte{0, 0, 0, 3, 102, 111, 111}
	result := PopString(&bts)

	assert.Equal(t, "foo", result)
	assert.Equal(t, []byte{}, bts)
}

func TestPushString(t *testing.T) {
	bts := []byte{}
	PushString(&bts, "foo")

	expected := []byte{0, 0, 0, 3, 102, 111, 111}
	assert.Equal(t, expected, bts)
}

func TestPopUUID(t *testing.T) {
	bts := []byte{1, 2, 3, 4, 5, 6, 7, 8, 8, 7, 6, 5, 4, 3, 2, 1}
	result := PopUUID(&bts)

	expected := types.UUID{1, 2, 3, 4, 5, 6, 7, 8, 8, 7, 6, 5, 4, 3, 2, 1}

	assert.Equal(t, expected, result)
	assert.Equal(t, []byte{}, bts)
}

func TestPopMessage(t *testing.T) {
	bts := []byte{32, 0, 0, 0, 5, 6}
	result := PopMessage(&bts)
	expected := []byte{32, 0, 0, 0, 5, 6}

	assert.Equal(t, expected, result)
	assert.Equal(t, []byte{}, bts)
}

func TestPutMsgLength(t *testing.T) {
	msg := []byte{0x50, 0, 0, 0, 0, 1, 2, 3, 4}
	PutMsgLength(msg)

	expected := []byte{0x50, 0, 0, 0, 8, 1, 2, 3, 4}
	assert.Equal(t, expected, msg)
}
