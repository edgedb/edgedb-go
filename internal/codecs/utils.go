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

package codecs

import (
	"fmt"

	"github.com/geldata/gel-go/internal/buff"
)

func encodeOptional(
	w *buff.Writer,
	missingValue, requiredArgument bool,
	encode, missingValueError func() error,
) error {
	switch {
	case missingValue && requiredArgument:
		return missingValueError()
	case missingValue:
		w.PushUint32(0xffffffff)
		return nil
	default:
		return encode()
	}
}

func encodeMarshaler(
	w *buff.Writer,
	val interface{},
	marshal func() ([]byte, error),
	expectedDataLen int,
	path Path,
) error {
	data, err := marshal()
	if err != nil {
		return err
	}
	if len(data) != expectedDataLen {
		return wrongNumberOfBytesError(val, path, expectedDataLen, len(data))
	}
	w.PushUint32(uint32(expectedDataLen))
	w.PushBytes(data)
	return nil
}

func missingValueError(val interface{}, path Path) error {
	var name string
	switch in := val.(type) {
	case string:
		name = in
	default:
		name = fmt.Sprintf("%T", in)
	}

	return fmt.Errorf(
		"cannot encode %v at %v because its value is missing",
		name, path)
}

func wrongNumberOfBytesError(
	val interface{},
	path Path,
	expected interface{},
	actual int,
) error {
	return fmt.Errorf(
		"wrong number of bytes encoded by %T at %v expected %v, got %v",
		val, path, expected, actual)
}
