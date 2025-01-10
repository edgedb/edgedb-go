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
	"encoding/binary"
	"testing"
	"time"

	"github.com/geldata/gel-go/internal/buff"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRoundingGoTime(t *testing.T) {
	cases := []struct {
		name     string
		input    string
		expected int64
	}{
		{
			name:     "maximum value",
			input:    "9999-12-31T23:59:59.999999499Z",
			expected: 252455615999999999,
		},
		{
			name:     "minimum value",
			input:    "0001-01-01T00:00:00.000000Z",
			expected: -63082281600000000,
		},
		{
			// remainder: 500
			// rounding: +1
			input:    "1814-03-09T01:02:03.000005500Z",
			expected: -5863791476999994,
			name:     "negative unix timestamp round up",
		},
		{
			// remainder: 501
			// rounding: -1
			input:    "1814-03-09T01:02:03.000005501Z",
			expected: -5863791476999994,
			name:     "negative unix timestamp 5501",
		},
		{
			// remainder: 499
			input:    "1814-03-09T01:02:03.000005499Z",
			expected: -5863791476999995,
			name:     "negative unix timestamp 5499",
		},
		{
			// remainder: 500
			input:    "1856-08-27T01:02:03.000004500Z",
			expected: -4523554676999996,
			name:     "negative unix timestamp round down",
		},
		{
			// remainder: 501
			// rounding: +1
			input:    "1856-08-27T01:02:03.000004501Z",
			expected: -4523554676999995,
			name:     "negative unix timestamp 4501",
		},
		{
			// remainder: 499
			input:    "1856-08-27T01:02:03.000004499Z",
			expected: -4523554676999996,
			name:     "negative unix timestamp 4499",
		},
		{
			// remainder: 500
			input:    "1969-12-31T23:59:59.999999500Z",
			expected: -946684800000000,
			name:     "unix timestamp to zero",
		},
		{
			// remainder: 500
			// rounding: +1
			input:    "1997-07-05T01:02:03.000009500Z",
			expected: -78620276999990,
			name:     "negative postgres timestamp round up",
		},
		{
			// remainder: 500
			// rounding: +1
			input:    "1997-07-05T01:02:03.000009500Z",
			expected: -78620276999990,
			name:     "negative postgres timestamp 9501",
		},
		{
			// remainder: 499
			input:    "1997-07-05T01:02:03.000009499Z",
			expected: -78620276999991,
			name:     "negative postgres timestamp 9499",
		},
		{
			// remainder: 500
			input:    "1997-07-05T01:02:03.000000500Z",
			expected: -78620277000000,
			name:     "negative postgres timestamp round down",
		},
		{
			// remainder: 501
			// rounding: -1
			input:    "1997-07-05T01:02:03.000000501Z",
			expected: -78620276999999,
			name:     "negative postgres timestamp 501",
		},
		{
			input:    "1997-07-05T01:02:03.000000499Z",
			expected: -78620277000000,
			name:     "negative postgres timestamp 499",
		},
		{
			input:    "1999-12-31T23:59:59.999999500Z",
			expected: 0,
			name:     "postgres timestamp to zero",
		},
		{
			input:    "2014-02-27T00:00:00.000001500Z",
			expected: 446774400000002,
			name:     "positive timestamp round up",
		},
		{
			input:    "2014-02-27T00:00:00.000001501Z",
			expected: 446774400000002,
			name:     "positive timestamp 1501",
		},
		{
			input:    "2014-02-27T00:00:00.000001499Z",
			expected: 446774400000001,
			name:     "positive timestamp 1499",
		},
		{
			input:    "2022-02-24T05:43:03.000002500Z",
			expected: 698996583000002,
			name:     "positive timestamp round down",
		},
		{
			input:    "2022-02-24T05:43:03.000002501Z",
			expected: 698996583000003,
			name:     "positive timestamp 2501",
		},
		{
			input:    "2022-02-24T05:43:03.000002499Z",
			expected: 698996583000002,
			name:     "positive timestamp 2499",
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			val, err := time.Parse(
				"2006-01-02T15:04:05.999999999Z",
				test.input,
			)
			require.NoError(t, err)

			data := make([]byte, 12)
			buf := buff.NewWriter(data)
			codec := DateTimeCodec{}
			err = codec.Encode(buf, val, "path-root", true)
			require.NoError(t, err)

			assert.Equal(
				t,
				test.expected,
				int64(binary.BigEndian.Uint64(data[4:])),
				"intput: %s", test.input,
			)
		})
	}
}
