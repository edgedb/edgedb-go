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

package edgedb

import (
	"context"
	"encoding/binary"
	"fmt"
	"math"
	"math/big"
	"math/rand"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSendAndReceiveInt64(t *testing.T) {
	ctx := context.Background()

	numbers := []int64{
		-1,
		1,
		0,
		11,
		-11,
		15,
		22,
		113,
		-11111,
		110000,
		-1100000,
		346456723423,
		-346456723423,
		281474976710656,
		2251799813685125,
		9007199254740992,
		-2251799813685125,
		1152921504594725865,
		-1152921504594725865,
	}

	for i := 0; i < 1000; i++ {
		numbers = append(numbers, int64(rand.Uint64()))
	}

	strings := make([]string, len(numbers))
	for i, n := range numbers {
		strings[i] = fmt.Sprint(n)
	}

	type Result struct {
		Encoded   string `edgedb:"encoded"`
		Decoded   int64  `edgedb:"decoded"`
		RoundTrip int64  `edgedb:"round_trip"`
		IsEqual   bool   `edgedb:"is_equal"`
		String    string `edgedb:"string"`
	}

	query := `
		WITH
			x := (
				WITH
					n := enumerate(array_unpack(<array<int64>>$0)),
					s := enumerate(array_unpack(<array<str>>$1)),
				SELECT (
					n := n.1,
					s := s.1,
				)
				FILTER n.0 = s.0
			)
		SELECT (
			encoded := <str>x.n,
			decoded := <int64>x.s,
			round_trip := x.n,
			is_equal := <int64>x.s = x.n,
			string := <str><int64>x.s,
		)
	`

	var results []Result
	err := conn.Query(ctx, query, &results, numbers, strings)
	require.NoError(t, err)
	require.Equal(t, len(numbers), len(results), "unexpected result count")

	for i, s := range strings {
		t.Run(s, func(t *testing.T) {
			n := numbers[i]
			r := results[i]

			assert.True(t, r.IsEqual, "equality check faild")
			assert.Equal(t, s, r.Encoded, "encoding failed")
			assert.Equal(t, n, r.Decoded, "decoding failed")
			assert.Equal(t, n, r.RoundTrip, "round trip failed")
			assert.Equal(t, s, r.String)
		})
	}
}

type CustomInt64 struct {
	data [8]byte
}

func (m CustomInt64) MarshalEdgeDBInt64() ([]byte, error) {
	return m.data[:], nil
}

func (m *CustomInt64) UnmarshalEdgeDBInt64(data []byte) error {
	copy(m.data[:], data)
	return nil
}

func TestSendAndReceiveInt64Marshaler(t *testing.T) {
	ctx := context.Background()

	query := `SELECT (
		encoded := <int64>$0,
		decoded := 123_456_789_987_654_321,
	)`

	type Result struct {
		Encoded int64       `edgedb:"encoded"`
		Decoded CustomInt64 `edgedb:"decoded"`
	}

	data := [8]byte{0x01, 0xb6, 0x9b, 0x4b, 0xe0, 0x52, 0xfa, 0xb1}
	arg := &CustomInt64{}
	copy(arg.data[:], data[:])

	var result Result
	err := conn.QueryOne(ctx, query, &result, arg)
	require.NoError(t, err)
	assert.Equal(t,
		Result{
			Encoded: 123_456_789_987_654_321,
			Decoded: CustomInt64{data},
		},
		result,
	)
}

func TestSendAndReceiveInt32(t *testing.T) {
	ctx := context.Background()

	numbers := []int32{-1, 0, 1, 10, 2147483647}
	for i := 0; i < 1000; i++ {
		numbers = append(numbers, int32(rand.Uint32()))
	}

	strings := make([]string, len(numbers))
	for i, n := range numbers {
		strings[i] = fmt.Sprint(n)
	}

	type Result struct {
		Encoded   string `edgedb:"encoded"`
		Decoded   int32  `edgedb:"decoded"`
		RoundTrip int32  `edgedb:"round_trip"`
		IsEqual   bool   `edgedb:"is_equal"`
		String    string `edgedb:"string"`
	}

	query := `
		WITH
			x := (
				WITH
					n := enumerate(array_unpack(<array<int32>>$0)),
					s := enumerate(array_unpack(<array<str>>$1)),
				SELECT (
					n := n.1,
					s := s.1,
				)
				FILTER n.0 = s.0
			)
		SELECT (
			encoded := <str>x.n,
			decoded := <int32>x.s,
			round_trip := x.n,
			is_equal := <int32>x.s = x.n,
			string := <str><int32>x.s,
		)
	`

	var results []Result
	err := conn.Query(ctx, query, &results, numbers, strings)
	require.NoError(t, err)
	require.Equal(t, len(numbers), len(results), "wrong number of results")

	for i, s := range strings {
		t.Run(s, func(t *testing.T) {
			n := numbers[i]
			r := results[i]

			assert.True(t, r.IsEqual, "equality check faild")
			assert.Equal(t, s, r.Encoded, "encoding failed")
			assert.Equal(t, n, r.Decoded, "decoding failed")
			assert.Equal(t, n, r.RoundTrip)
			assert.Equal(t, s, r.String)
		})
	}
}

type CustomInt32 struct {
	data [4]byte
}

func (m CustomInt32) MarshalEdgeDBInt32() ([]byte, error) {
	return m.data[:], nil
}

func (m *CustomInt32) UnmarshalEdgeDBInt32(data []byte) error {
	copy(m.data[:], data)
	return nil
}

func TestSendAndReceiveInt32Marshaler(t *testing.T) {
	ctx := context.Background()

	query := `SELECT (
		encoded := <int32>$0,
		decoded := <int32>655_665,
	)`

	type Result struct {
		Encoded int32       `edgedb:"encoded"`
		Decoded CustomInt32 `edgedb:"decoded"`
	}

	data := [4]byte{0x00, 0x0a, 0x01, 0x31}
	arg := &CustomInt32{}
	copy(arg.data[:], data[:])

	var result Result
	err := conn.QueryOne(ctx, query, &result, arg)
	require.NoError(t, err)
	assert.Equal(t,
		Result{
			Encoded: 655_665,
			Decoded: CustomInt32{data},
		},
		result,
	)
}

func TestSendAndReceiveInt16(t *testing.T) {
	ctx := context.Background()

	numbers := []int16{-1, 0, 1, 10, 15, 22, -1111}
	for i := 0; i < 1000; i++ {
		numbers = append(numbers, int16(rand.Uint32()))
	}

	strings := make([]string, len(numbers))
	for i, n := range numbers {
		strings[i] = fmt.Sprint(n)
	}

	type Result struct {
		Encoded   string `edgedb:"encoded"`
		Decoded   int16  `edgedb:"decoded"`
		RoundTrip int16  `edgedb:"round_trip"`
		IsEqual   bool   `edgedb:"is_equal"`
		String    string `edgedb:"string"`
	}

	query := `
		WITH
			x := (
				WITH
					n := enumerate(array_unpack(<array<int16>>$0)),
					s := enumerate(array_unpack(<array<str>>$1)),
				SELECT (
					n := n.1,
					s := s.1,
				)
				FILTER n.0 = s.0
			)
		SELECT (
			encoded := <str>x.n,
			decoded := <int16>x.s,
			round_trip := x.n,
			is_equal := <int16>x.s = x.n,
			string := <str><int16>x.s,
		)
	`

	var results []Result
	err := conn.Query(ctx, query, &results, numbers, strings)
	require.NoError(t, err)
	require.Equal(t, len(numbers), len(results), "wrong number of results")

	for i, s := range strings {
		t.Run(s, func(t *testing.T) {
			n := numbers[i]
			r := results[i]

			assert.True(t, r.IsEqual, "equality check faild")
			assert.Equal(t, s, r.Encoded, "encoding failed")
			assert.Equal(t, n, r.Decoded, "decoding failed")
			assert.Equal(t, n, r.RoundTrip, "round trip failed")
			assert.Equal(t, s, r.String)
		})
	}
}

type CustomInt16 struct {
	data [2]byte
}

func (m CustomInt16) MarshalEdgeDBInt16() ([]byte, error) {
	return m.data[:], nil
}

func (m *CustomInt16) UnmarshalEdgeDBInt16(data []byte) error {
	copy(m.data[:], data)
	return nil
}

func TestSendAndReceiveInt16Marshaler(t *testing.T) {
	ctx := context.Background()

	query := `SELECT (
		encoded := <int16>$0,
		decoded := <int16>6556,
	)`

	type Result struct {
		Encoded int16       `edgedb:"encoded"`
		Decoded CustomInt16 `edgedb:"decoded"`
	}

	data := [2]byte{0x19, 0x9c}
	arg := &CustomInt16{}
	copy(arg.data[:], data[:])

	var result Result
	err := conn.QueryOne(ctx, query, &result, arg)
	require.NoError(t, err)
	assert.Equal(t,
		Result{
			Encoded: 6556,
			Decoded: CustomInt16{data},
		},
		result,
	)
}

func TestSendAndReceiveBool(t *testing.T) {
	ctx := context.Background()

	query := `
		WITH
			i := <bool>$0,
			s := <str>$1,
		SELECT (
			encoded := <str>i,
			decoded := <bool>s,
			round_trip := i,
			is_equal := <bool>s = i,
			string := <str><bool>s,
		)
	`

	type Result struct {
		Encoded   string `edgedb:"encoded"`
		Decoded   bool   `edgedb:"decoded"`
		RoundTrip bool   `edgedb:"round_trip"`
		IsEqual   bool   `edgedb:"is_equal"`
		String    string `edgedb:"string"`
	}

	samples := []bool{true, false}

	for _, i := range samples {
		s := fmt.Sprint(i)
		t.Run(s, func(t *testing.T) {
			var result Result
			err := conn.QueryOne(ctx, query, &result, i, s)
			assert.NoError(t, err)

			assert.True(t, result.IsEqual, "equality check faild")
			assert.Equal(t, s, result.Encoded, "encoding failed")
			assert.Equal(t, i, result.Decoded, "decoding failed")
			assert.Equal(t, i, result.RoundTrip)
			assert.Equal(t, s, result.String)
		})
	}
}

type CustomBool struct {
	data [1]byte
}

func (m CustomBool) MarshalEdgeDBBool() ([]byte, error) {
	return m.data[:], nil
}

func (m *CustomBool) UnmarshalEdgeDBBool(data []byte) error {
	copy(m.data[:], data)
	return nil
}

func TestSendAndReceiveBoolMarshaler(t *testing.T) {
	ctx := context.Background()

	query := `SELECT (
		encoded := <bool>$0,
		decoded := <bool>true,
	)`

	type Result struct {
		Encoded bool       `edgedb:"encoded"`
		Decoded CustomBool `edgedb:"decoded"`
	}

	data := [1]byte{0x01}
	arg := &CustomBool{}
	copy(arg.data[:], data[:])

	var result Result
	err := conn.QueryOne(ctx, query, &result, arg)
	require.NoError(t, err)
	assert.Equal(t,
		Result{
			Encoded: true,
			Decoded: CustomBool{data},
		},
		result,
	)
}

func TestSendAndReceiveFloat64(t *testing.T) {
	ctx := context.Background()

	numbers := []float64{0, 1, 123.2, -1.1}
	for i := 0; i < 1000; i++ {
		n := math.Float64frombits(rand.Uint64())

		// NaN is not equal to itself so assertions will fail.
		if !math.IsNaN(n) {
			numbers = append(numbers, n)
		}
	}

	strings := make([]string, len(numbers))
	for i, n := range numbers {
		strings[i] = fmt.Sprint(n)
	}

	type Result struct {
		Encoded   string  `edgedb:"encoded"`
		Decoded   float64 `edgedb:"decoded"`
		RoundTrip float64 `edgedb:"round_trip"`
		IsEqual   bool    `edgedb:"is_equal"`
	}

	query := `
		WITH
			x := (
				WITH
					n := enumerate(array_unpack(<array<float64>>$0)),
					s := enumerate(array_unpack(<array<str>>$1)),
				SELECT (
					n := n.1,
					s := s.1,
				)
				FILTER n.0 = s.0
			)
		SELECT (
			encoded := <str>x.n,
			decoded := <float64>x.s,
			round_trip := x.n,
			is_equal := <float64>x.s = x.n,
		)
	`

	var results []Result
	err := conn.Query(ctx, query, &results, numbers, strings)
	require.NoError(t, err)
	require.Equal(t, len(numbers), len(results), "wrong number of results")

	for i, s := range strings {
		t.Run(s, func(t *testing.T) {
			n := numbers[i]
			r := results[i]

			encoded, err := strconv.ParseFloat(r.Encoded, 64)
			require.NoError(t, err)

			assert.True(t, r.IsEqual, "equality check faild")
			assert.Equal(t, n, encoded, "encoding failed")
			assert.Equal(t, n, r.Decoded, "decoding failed")
			assert.Equal(t, n, r.RoundTrip, "round trip failed")
		})
	}
}

type CustomFloat64 struct {
	data [8]byte
}

func (m CustomFloat64) MarshalEdgeDBFloat64() ([]byte, error) {
	return m.data[:], nil
}

func (m *CustomFloat64) UnmarshalEdgeDBFloat64(data []byte) error {
	copy(m.data[:], data)
	return nil
}

func TestSendAndReceiveFloat64Marshaler(t *testing.T) {
	ctx := context.Background()

	query := `SELECT (
		encoded := <float64>$0,
		decoded := <float64>-15.625,
	)`

	type Result struct {
		Encoded float64       `edgedb:"encoded"`
		Decoded CustomFloat64 `edgedb:"decoded"`
	}

	data := [8]byte{0xc0, 0x2f, 0x40, 0x00, 0x00, 0x00, 0x00, 0x00}
	arg := &CustomFloat64{}
	copy(arg.data[:], data[:])

	var result Result
	err := conn.QueryOne(ctx, query, &result, arg)
	require.NoError(t, err)
	assert.Equal(t,
		Result{
			Encoded: -15.625,
			Decoded: CustomFloat64{data},
		},
		result,
	)
}

func TestSendAndReceiveFloat32(t *testing.T) {
	ctx := context.Background()

	numbers := []float32{0, 1, 123.2, -1.1}
	for i := 0; i < 1000; i++ {
		n := math.Float32frombits(rand.Uint32())

		// NaN is not equal to itself so assertions will fail.
		if !math.IsNaN(float64(n)) {
			numbers = append(numbers, n)
		}
	}

	strings := make([]string, len(numbers))
	for i, n := range numbers {
		strings[i] = fmt.Sprint(n)
	}

	type Result struct {
		Encoded   string  `edgedb:"encoded"`
		Decoded   float32 `edgedb:"decoded"`
		RoundTrip float32 `edgedb:"round_trip"`
		IsEqual   bool    `edgedb:"is_equal"`
	}

	query := `
		WITH
			x := (
				WITH
					n := enumerate(array_unpack(<array<float32>>$0)),
					s := enumerate(array_unpack(<array<str>>$1)),
				SELECT (
					n := n.1,
					s := s.1,
				)
				FILTER n.0 = s.0
			)
		SELECT (
			encoded := <str><float32>x.n,
			decoded := <float32>x.s,
			round_trip := x.n,
			is_equal := <float32>x.s = x.n,
		)
	`

	var results []Result
	err := conn.Query(ctx, query, &results, numbers, strings)
	require.NoError(t, err)
	require.Equal(t, len(numbers), len(results), "wrong number of results")

	for i, s := range strings {
		t.Run(s, func(t *testing.T) {
			n := numbers[i]
			r := results[i]

			encoded, err := strconv.ParseFloat(r.Encoded, 32)
			require.NoError(t, err)

			assert.True(t, r.IsEqual, "equality check faild")
			assert.Equal(t, n, float32(encoded), "encoding failed")
			assert.Equal(t, n, r.Decoded, "decoding failed")
			assert.Equal(t, n, r.RoundTrip, "round trip failed")
		})
	}
}

type CustomFloat32 struct {
	data [4]byte
}

func (m CustomFloat32) MarshalEdgeDBFloat32() ([]byte, error) {
	return m.data[:], nil
}

func (m *CustomFloat32) UnmarshalEdgeDBFloat32(data []byte) error {
	copy(m.data[:], data)
	return nil
}

func TestSendAndReceiveFloat32Marshaler(t *testing.T) {
	ctx := context.Background()

	query := `SELECT (
		encoded := <float32>$0,
		decoded := <float32>-15.625,
	)`

	type Result struct {
		Encoded float32       `edgedb:"encoded"`
		Decoded CustomFloat32 `edgedb:"decoded"`
	}

	data := [4]byte{0xc1, 0x7a, 0x00, 0x00}
	arg := &CustomFloat32{}
	copy(arg.data[:], data[:])

	var result Result
	err := conn.QueryOne(ctx, query, &result, arg)
	require.NoError(t, err)
	assert.Equal(t,
		Result{
			Encoded: -15.625,
			Decoded: CustomFloat32{data},
		},
		result,
	)
}

func TestSendAndReceiveBytes(t *testing.T) {
	ctx := context.Background()

	samples := [][]byte{
		[]byte("abcdef"),
	}

	for i := 0; i < 1000; i++ {
		n := rand.Intn(999) + 1
		b := make([]byte, n)

		for i := 0; i < n; i++ {
			b[i] = uint8(rand.Uint32())
		}

		samples = append(samples, b)
	}

	query := `SELECT array_unpack(<array<bytes>>$0)`

	var results [][]byte
	err := conn.Query(ctx, query, &results, samples)
	require.NoError(t, err)
	require.Equal(t, len(samples), len(results), "wrong number of results")

	for i, b := range samples {
		t.Run(string(b), func(t *testing.T) {
			assert.Equal(t, b, results[i])
		})
	}
}

type CustomBytes struct {
	data []byte
}

func (m CustomBytes) MarshalEdgeDBBytes() ([]byte, error) {
	return m.data, nil
}

func (m *CustomBytes) UnmarshalEdgeDBBytes(data []byte) error {
	m.data = data
	return nil
}

func TestSendAndReceiveBytesMarshaler(t *testing.T) {
	ctx := context.Background()

	query := `SELECT (
		encoded := <bytes>$0,
		decoded := b'\x01\x02\x03',
	)`

	type Result struct {
		Encoded []byte      `edgedb:"encoded"`
		Decoded CustomBytes `edgedb:"decoded"`
	}

	data := []byte{0x01, 0x02, 0x03}
	arg := &CustomBytes{make([]byte, len(data))}
	copy(arg.data, data)

	var result Result
	err := conn.QueryOne(ctx, query, &result, arg)
	require.NoError(t, err)
	assert.Equal(t,
		Result{
			Encoded: data,
			Decoded: CustomBytes{data},
		},
		result,
	)
}

func TestSendAndReceiveStr(t *testing.T) {
	ctx := context.Background()

	var result string
	err := conn.QueryOne(ctx, `SELECT <str>$0`, &result, "abcdef")
	require.NoError(t, err)
	assert.Equal(t, "abcdef", result, "round trip failed")
}

func TestFetchLargeStr(t *testing.T) {
	// This test is meant to stress the buffer implementation.
	ctx := context.Background()

	var result string
	err := conn.QueryOne(ctx, "SELECT str_repeat('a', <int64>(10^6))", &result)
	require.NoError(t, err)
	assert.Equal(t, strings.Repeat("a", 1_000_000), result)
}

type CustomStr struct {
	data []byte
}

func (m CustomStr) MarshalEdgeDBStr() ([]byte, error) {
	return m.data, nil
}

func (m *CustomStr) UnmarshalEdgeDBStr(data []byte) error {
	m.data = data
	return nil
}

func TestSendAndReceiveStrMarshaler(t *testing.T) {
	ctx := context.Background()

	query := `SELECT (
		encoded := <str>$0,
		decoded := 'Hello! 🙂',
	)`

	type Result struct {
		Encoded string    `edgedb:"encoded"`
		Decoded CustomStr `edgedb:"decoded"`
	}

	data := []byte{
		0x48, 0x65, 0x6c, 0x6c, 0x6f, 0x21, 0x20, 0xf0, 0x9f, 0x99, 0x82,
	}
	arg := &CustomStr{make([]byte, len(data))}
	copy(arg.data, data)

	var result Result
	err := conn.QueryOne(ctx, query, &result, arg)
	require.NoError(t, err)
	assert.Equal(t,
		Result{
			Encoded: "Hello! 🙂",
			Decoded: CustomStr{data},
		},
		result,
	)
}

func TestSendAndReceiveJSON(t *testing.T) {
	ctx := context.Background()

	strings := []string{"123", "-3.14", "true", "false", "[1, 2, 3]", "null"}

	samples := make([][]byte, len(strings))
	for i, s := range strings {
		samples[i] = []byte(s)
	}

	query := `SELECT array_unpack(<array<json>>$0)`

	var results [][]byte
	err := conn.Query(ctx, query, &results, samples)
	require.NoError(t, err)
	require.Equal(t, len(samples), len(results), "wrong number of results")

	for i, s := range strings {
		t.Run(s, func(t *testing.T) {
			assert.Equal(t, samples[i], results[i])
		})
	}
}

type CustomJSON struct {
	data []byte
}

func (m CustomJSON) MarshalEdgeDBJSON() ([]byte, error) {
	return m.data, nil
}

func (m *CustomJSON) UnmarshalEdgeDBJSON(data []byte) error {
	m.data = data
	return nil
}

func TestSendAndReceiveJSONMarshaler(t *testing.T) {
	ctx := context.Background()

	query := `SELECT (
		encoded := to_str(<json>$0),
		decoded := <json>(hello := "world"),
	)`

	type Result struct {
		Encoded string     `edgedb:"encoded"`
		Decoded CustomJSON `edgedb:"decoded"`
	}

	data := append([]byte{1}, []byte(`{"hello": "world"}`)...)
	arg := &CustomJSON{make([]byte, len(data))}
	copy(arg.data, data)

	var result Result
	err := conn.QueryOne(ctx, query, &result, arg)
	require.NoError(t, err)
	assert.Equal(t,
		Result{
			Encoded: `{"hello": "world"}`,
			Decoded: CustomJSON{data},
		},
		result,
	)
}

func TestSendAndReceiveEnum(t *testing.T) {
	ctx := context.Background()

	type Result struct {
		Encoded   string `edgedb:"encoded"`
		Decoded   string `edgedb:"decoded"`
		RoundTrip string `edgedb:"round_trip"`
		IsEqual   bool   `edgedb:"is_equal"`
		String    string `edgedb:"string"`
	}

	query := `
		WITH
			e := <ColorEnum>$0,
			s := <str>$1
		SELECT (
			encoded := <str>e,
			decoded := <ColorEnum>s,
			round_trip := e,
			is_equal := <ColorEnum>s = e,
			string := <str><ColorEnum>s
		)
	`

	var result Result
	color := "Red"
	err := conn.QueryOne(ctx, query, &result, color, color)
	require.NoError(t, err)

	assert.Equal(t, color, result.Encoded, "encoding failed")
	assert.Equal(t, color, result.Decoded, "decoding failed")
	assert.Equal(t, color, result.RoundTrip, "round trip failed")
	assert.True(t, result.IsEqual, "equality failed")
	assert.Equal(t, color, result.String)

	query = "SELECT (decoded := <ColorEnum><str>$0)"
	err = conn.QueryOne(ctx, query, &result, "invalid")

	expected := "edgedb.InvalidValueError: " +
		"invalid input value for enum 'default::ColorEnum': \"invalid\""
	assert.EqualError(t, err, expected)
}

type CustomEnum struct {
	data []byte
}

func (m CustomEnum) MarshalEdgeDBStr() ([]byte, error) {
	return m.data, nil
}

func (m *CustomEnum) UnmarshalEdgeDBStr(data []byte) error {
	m.data = data
	return nil
}

func TestSendAndReceiveEnumMarshaler(t *testing.T) {
	ctx := context.Background()

	query := `SELECT (
		encoded := <ColorEnum>$0,
		decoded := <ColorEnum>'Red',
	)`

	type Result struct {
		Encoded string     `edgedb:"encoded"`
		Decoded CustomEnum `edgedb:"decoded"`
	}
	data := []byte{0x52, 0x65, 0x64}
	arg := &CustomEnum{make([]byte, len(data))}
	copy(arg.data, data)

	var result Result
	err := conn.QueryOne(ctx, query, &result, arg)
	require.NoError(t, err)
	assert.Equal(t,
		Result{
			Encoded: "Red",
			Decoded: CustomEnum{data},
		},
		result,
	)
}

func TestSendAndReceiveDuration(t *testing.T) {
	ctx := context.Background()

	durations := []Duration{
		Duration(0),
		Duration(-1),
		Duration(86400000000),
		Duration(1_000_000),
		Duration(3074457345618258432),
	}

	var maxDuration int64 = 3_154_000_000_000_000
	for i := 0; i < 1000; i++ {
		d := Duration(rand.Int63n(2*maxDuration) - maxDuration)
		durations = append(durations, d)
	}

	strings := make([]string, len(durations))
	for i := 0; i < len(strings); i++ {
		strings[i] = durations[i].String()
	}

	type Result struct {
		Decoded   Duration `edgedb:"decoded"`
		RoundTrip Duration `edgedb:"round_trip"`
		IsEqual   bool     `edgedb:"is_equal"`
	}

	query := `
		WITH
			sample := (
				WITH
					d := enumerate(array_unpack(<array<duration>>$0)),
					s := enumerate(array_unpack(<array<str>>$1)),
				SELECT (
					d := d.1,
					str := s.1,
				)
				FILTER d.0 = s.0
			)
		SELECT (
			decoded := <duration>sample.str,
			round_trip := sample.d,
			is_equal := <duration>sample.str = sample.d,
		)
	`

	var results []Result
	err := conn.Query(ctx, query, &results, durations, strings)
	require.NoError(t, err)
	require.Equal(t, len(durations), len(results), "wrong number of results")

	for i, s := range strings {
		t.Run(s, func(t *testing.T) {
			d := durations[i]
			result := results[i]
			assert.True(t, result.IsEqual, "equality check faild")
			assert.Equal(t, d, result.RoundTrip, "round trip failed")
			assert.Equal(t, d, result.Decoded, "decoding failed")
		})
	}
}

type CustomDuration struct {
	data [16]byte
}

func (m CustomDuration) MarshalEdgeDBDuration() ([]byte, error) {
	return m.data[:], nil
}

func (m *CustomDuration) UnmarshalEdgeDBDuration(data []byte) error {
	copy(m.data[:], data)
	return nil
}

func TestSendAndReceiveDurationMarshaler(t *testing.T) {
	ctx := context.Background()

	query := `SELECT (
		encoded := <duration>$0,
		decoded := <duration>'48 hours 45 minutes 7.6 seconds',
	)`

	type Result struct {
		Encoded Duration       `edgedb:"encoded"`
		Decoded CustomDuration `edgedb:"decoded"`
	}

	data := [16]byte{
		0x00, 0x00, 0x00, 0x28, 0xdd, 0x11, 0x72, 0x80, // microseconds
		0x00, 0x00, 0x00, 0x00, // days
		0x00, 0x00, 0x00, 0x00, // months
	}
	arg := &CustomDuration{}
	copy(arg.data[:], data[:])

	var result Result
	err := conn.QueryOne(ctx, query, &result, arg)
	require.NoError(t, err)
	assert.Equal(t,
		Result{
			Encoded: Duration(0x28dd117280),
			Decoded: CustomDuration{data},
		},
		result,
	)
}

func TestSendAndReceiveRelativeDuration(t *testing.T) {
	ctx := context.Background()

	var duration RelativeDuration
	err := conn.QueryOne(ctx, "SELECT <cal::relative_duration>'1y'", &duration)
	if err != nil {
		t.Skip("server version is too old for this feature")
	}

	rds := []RelativeDuration{
		NewRelativeDuration(0, 0, 0),
		NewRelativeDuration(0, 0, 1),
		NewRelativeDuration(0, 0, -1),
		NewRelativeDuration(0, 1, 0),
		NewRelativeDuration(0, -1, 0),
		NewRelativeDuration(1, 0, 0),
		NewRelativeDuration(-1, 0, 0),
		NewRelativeDuration(1, 1, 1),
		NewRelativeDuration(-1, -1, -1),
	}

	for i := 0; i < 5_000; i++ {
		rds = append(rds, NewRelativeDuration(
			rand.Int31n(101)-int32(50),
			rand.Int31n(1_001)-int32(500),
			rand.Int63n(2_000_000_000)-int64(1_000_000_000),
		))
	}

	type Result struct {
		RoundTrip RelativeDuration `edgedb:"round_trip"`
		Str       string           `edgedb:"str"`
	}

	query := `
		WITH args := array_unpack(<array<cal::relative_duration>>$0)
		SELECT (
			round_trip := args,
			str := <str>args,
		)
	`

	var results []Result
	err = conn.Query(ctx, query, &results, rds)
	require.NoError(t, err)
	require.Equal(t, len(rds), len(results), "wrong number of results")

	for i, rd := range rds {
		t.Run(rd.String(), func(t *testing.T) {
			result := results[i]
			assert.Equal(t, rd, result.RoundTrip, "round trip failed")
			assert.Equal(t, rd.String(), result.Str, "incorrect String() val")
		})
	}
}

type CustomRelativeDuration struct {
	data [16]byte
}

func (m CustomRelativeDuration) MarshalEdgeDBRelativeDuration() (
	[]byte, error) {
	return m.data[:], nil
}

func (m *CustomRelativeDuration) UnmarshalEdgeDBRelativeDuration(
	data []byte,
) error {
	copy(m.data[:], data)
	return nil
}

func TestSendAndReceiveRelativeDurationMarshaler(t *testing.T) {
	ctx := context.Background()

	var duration RelativeDuration
	err := conn.QueryOne(ctx, "SELECT <cal::relative_duration>'1y'", &duration)
	if err != nil {
		t.Skip("server version is too old for this feature")
	}

	query := `SELECT (
		encoded := <cal::relative_duration>$0,
		decoded := <cal::relative_duration>
			'8 months 5 days 48 hours 45 minutes 7.6 seconds',
	)`

	type Result struct {
		Encoded RelativeDuration       `edgedb:"encoded"`
		Decoded CustomRelativeDuration `edgedb:"decoded"`
	}

	data := [16]byte{
		0x00, 0x00, 0x00, 0x28, 0xdd, 0x11, 0x72, 0x80, // microseconds
		0x00, 0x00, 0x00, 0x05, // days
		0x00, 0x00, 0x00, 0x08, // months
	}
	arg := &CustomRelativeDuration{}
	copy(arg.data[:], data[:])

	var result Result
	err = conn.QueryOne(ctx, query, &result, arg)
	require.NoError(t, err)
	assert.Equal(t,
		Result{
			Encoded: NewRelativeDuration(8, 5, 0x28dd117280),
			Decoded: CustomRelativeDuration{data},
		},
		result,
	)
}

func TestSendAndReceiveLocalTime(t *testing.T) {
	ctx := context.Background()

	times := []LocalTime{
		NewLocalTime(0, 0, 0, 0),
		NewLocalTime(0, 0, 0, 1),
		NewLocalTime(0, 0, 0, 10),
		NewLocalTime(0, 0, 0, 100),
		NewLocalTime(0, 0, 0, 1000),
		NewLocalTime(0, 0, 0, 10000),
		NewLocalTime(0, 0, 0, 100000),
		NewLocalTime(0, 0, 0, 123456),
		NewLocalTime(0, 1, 11, 340000),
		NewLocalTime(5, 4, 3, 0),
		NewLocalTime(11, 12, 13, 0),
		NewLocalTime(20, 39, 57, 0),
		NewLocalTime(23, 59, 59, 999000),
		NewLocalTime(23, 59, 59, 999999),
	}

	for i := 0; i < 1_000; i++ {
		times = append(times, NewLocalTime(
			rand.Intn(24),
			rand.Intn(60),
			rand.Intn(60),
			rand.Intn(1_000_000),
		))
	}

	strings := make([]string, len(times))
	for i, t := range times {
		strings[i] = t.String()
	}

	type Result struct {
		Encoded   string    `edgedb:"encoded"`
		Decoded   LocalTime `edgedb:"decoded"`
		RoundTrip LocalTime `edgedb:"round_trip"`
		IsEqual   bool      `edgedb:"is_equal"`
		String    string    `edgedb:"string"`
	}

	query := `
		WITH
			x := (
				WITH
					t := enumerate(array_unpack(<array<cal::local_time>>$0)),
					s := enumerate(array_unpack(<array<str>>$1)),
				SELECT (
					t := t.1,
					s := s.1,
				)
				FILTER t.0 = s.0
			)
		SELECT (
			encoded := <str>x.t,
			decoded := <cal::local_time>x.s,
			round_trip := x.t,
			is_equal := <cal::local_time>x.s = x.t,
			string := <str><cal::local_time><str>x.s,
		)
	`

	var results []Result
	err := conn.Query(ctx, query, &results, times, strings)
	require.NoError(t, err)

	for i, s := range strings {
		t.Run(s, func(t *testing.T) {
			time := times[i]
			r := results[i]

			assert.Equal(t, time, r.RoundTrip, "round trip failed")
			assert.Equal(t, time, r.Decoded, "decode is wrong")
			assert.Equal(t, s, r.Encoded, "encode is wrong")
			assert.True(t, r.IsEqual, "equality failed")
			assert.Equal(t, s, r.String)
		})
	}
}

type CustomLocalTime struct {
	data [8]byte
}

func (m CustomLocalTime) MarshalEdgeDBLocalTime() ([]byte, error) {
	return m.data[:], nil
}

func (m *CustomLocalTime) UnmarshalEdgeDBLocalTime(data []byte) error {
	copy(m.data[:], data)
	return nil
}

func TestSendAndReceiveLocalTimeMarshaler(t *testing.T) {
	ctx := context.Background()

	query := `SELECT (
		encoded := <str><cal::local_time>$0,
		decoded := <cal::local_time>'12:10:00',
	)`

	type Result struct {
		Encoded string          `edgedb:"encoded"`
		Decoded CustomLocalTime `edgedb:"decoded"`
	}

	data := [8]byte{0x00, 0x00, 0x00, 0x0a, 0x32, 0xae, 0xf6, 0x00}
	arg := &CustomLocalTime{}
	copy(arg.data[:], data[:])

	var result Result
	err := conn.QueryOne(ctx, query, &result, arg)
	require.NoError(t, err)
	assert.Equal(t,
		Result{
			Encoded: "12:10:00",
			Decoded: CustomLocalTime{data},
		},
		result,
	)
}

func TestSendAndReceiveLocalDate(t *testing.T) {
	ctx := context.Background()

	dates := []LocalDate{
		NewLocalDate(1, 1, 1),
		NewLocalDate(2000, 1, 1),
		NewLocalDate(2019, 5, 6),
		NewLocalDate(4444, 12, 30),
		NewLocalDate(9999, 9, 9),
	}

	for i := 0; i < 1_000; i++ {
		dates = append(dates, NewLocalDate(
			rand.Intn(9999)+1,
			time.Month(rand.Intn(12)+1),
			rand.Intn(30)+1,
		))
	}

	strings := make([]string, len(dates))
	for i, d := range dates {
		strings[i] = d.String()
	}

	type Result struct {
		Encoded   string    `edgedb:"encoded"`
		Decoded   LocalDate `edgedb:"decoded"`
		RoundTrip LocalDate `edgedb:"round_trip"`
		IsEqual   bool      `edgedb:"is_equal"`
		String    string    `edgedb:"string"`
	}

	query := `
		WITH
			x := (
				WITH
					d := enumerate(array_unpack(<array<cal::local_date>>$0)),
					s := enumerate(array_unpack(<array<str>>$1)),
				SELECT (
					d := d.1,
					s := s.1,
				)
				FILTER d.0 = s.0
			)
		SELECT (
			encoded := <str>x.d,
			decoded := <cal::local_date>x.s,
			round_trip := x.d,
			is_equal := <cal::local_date>x.s = x.d,
			string := <str><cal::local_date>x.s,
		)
	`

	var results []Result
	err := conn.Query(ctx, query, &results, dates, strings)
	require.NoError(t, err)
	require.Equal(t, len(dates), len(results))

	for i, s := range strings {
		t.Run(s, func(t *testing.T) {
			d := dates[i]
			r := results[i]

			assert.Equal(t, d, r.RoundTrip, "round trip failed")
			assert.Equal(t, d, r.Decoded, "decode is wrong")
			assert.Equal(t, s, r.Encoded, "encode is wrong")
			assert.True(t, r.IsEqual, "equality failed")
			assert.Equal(t, s, r.String)
		})
	}
}

type CustomLocalDate struct {
	data [4]byte
}

func (m CustomLocalDate) MarshalEdgeDBLocalDate() ([]byte, error) {
	return m.data[:], nil
}

func (m *CustomLocalDate) UnmarshalEdgeDBLocalDate(data []byte) error {
	copy(m.data[:], data)
	return nil
}

func TestSendAndReceiveLocalDateMarshaler(t *testing.T) {
	ctx := context.Background()

	query := `SELECT (
		encoded := <str><cal::local_date>$0,
		decoded := <cal::local_date>'2019-05-06',
	)`

	type Result struct {
		Encoded string          `edgedb:"encoded"`
		Decoded CustomLocalDate `edgedb:"decoded"`
	}

	data := [4]byte{0x00, 0x00, 0x1b, 0x99}
	arg := &CustomLocalDate{}
	copy(arg.data[:], data[:])

	var result Result
	err := conn.QueryOne(ctx, query, &result, arg)
	require.NoError(t, err)
	assert.Equal(t,
		Result{
			Encoded: "2019-05-06",
			Decoded: CustomLocalDate{data},
		},
		result,
	)
}

func TestSendAndReceiveLocalDateTime(t *testing.T) {
	ctx := context.Background()

	datetimes := []LocalDateTime{
		NewLocalDateTime(2019, 5, 6, 12, 0, 0, 0),
		NewLocalDateTime(2018, 5, 7, 15, 1, 22, 306916),
		NewLocalDateTime(1, 1, 1, 1, 1, 0, 0),
		NewLocalDateTime(9999, 9, 9, 9, 9, 9, 0),
	}

	for i := 0; i < 1_000; i++ {
		dt := NewLocalDateTime(
			rand.Intn(9999)+1,
			time.Month(rand.Intn(12))+1,
			rand.Intn(30)+1,
			rand.Intn(24),
			rand.Intn(60),
			rand.Intn(60),
			rand.Intn(1_000_000),
		)

		datetimes = append(datetimes, dt)
	}

	strings := make([]string, len(datetimes))
	for i, t := range datetimes {
		strings[i] = t.String()
	}

	type Result struct {
		Encoded   string        `edgedb:"encoded"`
		Decoded   LocalDateTime `edgedb:"decoded"`
		RoundTrip LocalDateTime `edgedb:"round_trip"`
		IsEqual   bool          `edgedb:"is_equal"`
		String    string        `edgedb:"string"`
	}

	query := `
		WITH
			x := (
				WITH
					dt := enumerate(array_unpack(
						<array<cal::local_datetime>>$0
					)),
					s := enumerate(array_unpack(<array<str>>$1)),
				SELECT (
					dt := dt.1,
					s := s.1,
				)
				FILTER dt.0 = s.0
			)
		SELECT (
			encoded := <str>x.dt,
			decoded := <cal::local_datetime>x.s,
			round_trip := x.dt,
			is_equal := <cal::local_datetime>x.s = x.dt,
			string := <str><cal::local_datetime>x.s,
		)
	`

	var results []Result
	err := conn.Query(ctx, query, &results, datetimes, strings)
	require.NoError(t, err)
	require.Equal(t, len(datetimes), len(results), "wrong number of results")

	for i, s := range strings {
		t.Run(s, func(t *testing.T) {
			dt := datetimes[i]
			r := results[i]

			assert.True(t, r.IsEqual, "equality check faild")
			assert.Equal(t, s, r.Encoded, "encoding failed")
			assert.Equal(t, dt, r.Decoded)
			assert.Equal(t, dt, r.RoundTrip)
			assert.Equal(t, s, r.String)
		})
	}
}

type CustomLocalDateTime struct {
	data [8]byte
}

func (m CustomLocalDateTime) MarshalEdgeDBLocalDateTime() ([]byte, error) {
	return m.data[:], nil
}

func (m *CustomLocalDateTime) UnmarshalEdgeDBLocalDateTime(data []byte) error {
	copy(m.data[:], data)
	return nil
}

func TestSendAndReceiveLocalDateTimeMarshaler(t *testing.T) {
	ctx := context.Background()

	query := `SELECT (
		encoded := <str><cal::local_datetime>$0,
		decoded := <cal::local_datetime>'2019-05-06T12:00:00',
	)`

	type Result struct {
		Encoded string              `edgedb:"encoded"`
		Decoded CustomLocalDateTime `edgedb:"decoded"`
	}

	data := [8]byte{0x00, 0x02, 0x2b, 0x35, 0x9b, 0xc4, 0x10, 0x00}
	arg := &CustomLocalDateTime{}
	copy(arg.data[:], data[:])

	var result Result
	err := conn.QueryOne(ctx, query, &result, arg)
	require.NoError(t, err)
	assert.Equal(t,
		Result{
			Encoded: "2019-05-06T12:00:00",
			Decoded: CustomLocalDateTime{data},
		},
		result,
	)
}

func TestSendAndReceiveDateTime(t *testing.T) {
	ctx := context.Background()
	format := "2006-01-02T15:04:05.999999-07:00"

	samples := []time.Time{
		time.Date(2019, 5, 6, 12, 0, 0, 0, time.UTC),
		time.Date(1986, 4, 26, 1, 23, 40, 1_000, time.FixedZone("", -25200)),
		time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(9999, 9, 9, 9, 9, 0, 0, time.FixedZone("", 32400)),
	}

	const maxDate = 253402300799
	const minDate = -62135596800

	for i := 0; i < 1000; i++ {
		samples = append(samples, time.Unix(
			rand.Int63n(maxDate-minDate)+minDate,
			1_000*rand.Int63n(1_000_000),
		))
	}

	strings := make([]string, len(samples))
	for i, t := range samples {
		strings[i] = t.UTC().Format(format)
	}

	type Result struct {
		Encoded   string    `edgedb:"encoded"`
		Decoded   time.Time `edgedb:"decoded"`
		RoundTrip time.Time `edgedb:"round_trip"`
		IsEqual   bool      `edgedb:"is_equal"`
		String    string    `edgedb:"string"`
	}

	query := `
		WITH
			x := (
				WITH
					dt := enumerate(array_unpack(<array<datetime>>$0)),
					s := enumerate(array_unpack(<array<str>>$1)),
				SELECT (
					dt := dt.1,
					s := s.1,
				)
				FILTER dt.0 = s.0
			)
		SELECT (
			encoded := <str>x.dt,
			decoded := <datetime>x.s,
			round_trip := x.dt,
			is_equal := <datetime>x.s = x.dt,
			string := <str><datetime>x.s,
		)
	`

	var results []Result
	err := conn.Query(ctx, query, &results, samples, strings)
	require.NoError(t, err)
	require.Equal(t, len(samples), len(results), "wrong number of results")

	for i, s := range strings {
		t.Run(s, func(t *testing.T) {
			dt := samples[i].UTC()
			r := results[i]

			assert.True(t, r.IsEqual, "equality check faild: %v", dt.Unix())
			assert.Equal(t, s, r.Encoded, "encoding failed")
			assert.Equal(t, s, r.String, "string failed")
			assert.True(t,
				dt.Equal(r.Decoded),
				"decoding failed: %v != %v", dt, r.Decoded,
			)
			assert.True(t,
				dt.Equal(r.RoundTrip),
				"round trip failed: %v != %v", dt, r.RoundTrip,
			)
		})
	}
}

type CustomDateTime struct {
	data [8]byte
}

func (m CustomDateTime) MarshalEdgeDBDateTime() ([]byte, error) {
	return m.data[:], nil
}

func (m *CustomDateTime) UnmarshalEdgeDBDateTime(data []byte) error {
	copy(m.data[:], data)
	return nil
}

func TestSendAndReceiveDateTimeMarshaler(t *testing.T) {
	ctx := context.Background()

	query := `SELECT (
		encoded := <str><datetime>$0,
		decoded := <datetime>'2019-05-06T12:00:00+00:00',
	)`

	type Result struct {
		Encoded string         `edgedb:"encoded"`
		Decoded CustomDateTime `edgedb:"decoded"`
	}

	data := [8]byte{0x00, 0x02, 0x2b, 0x35, 0x9b, 0xc4, 0x10, 0x00}
	arg := &CustomDateTime{}
	copy(arg.data[:], data[:])

	var result Result
	err := conn.QueryOne(ctx, query, &result, arg)
	require.NoError(t, err)
	assert.Equal(t,
		Result{
			Encoded: "2019-05-06T12:00:00+00:00",
			Decoded: CustomDateTime{data},
		},
		result,
	)
}

func TestSendAndReceiveBigInt(t *testing.T) {
	ctx := context.Background()

	query := `
		WITH
			i := <bigint>$0,
			s := <str>$1
		SELECT (
			encoded := <str>i,
			decoded := <bigint>s,
			round_trip := i,
			is_equal := <bigint>s = i,
			string := <str><bigint>s,
		)
	`

	type Result struct {
		Encoded   string   `edgedb:"encoded"`
		Decoded   *big.Int `edgedb:"decoded"`
		RoundTrip *big.Int `edgedb:"round_trip"`
		IsEqual   bool     `edgedb:"is_equal"`
		String    string   `edgedb:"string"`
	}

	samples := []string{
		"0",
		"1",
		"-1",
		"11",
		"-11",
		"123",
		"-123",
		"123789",
		"-123789",
		"19876",
		"-19876",
		"19876",
		"-19876",
		"11001200000031231238172638172637981268371628312300000000",
		"-11001231231238172638172637981268371628312300",
		"198761239812739812739801279371289371932",
		"-198761182763908473812974620938742386",
		"98761239812739812739801279371289371932",
		"-98761182763908473812974620938742386",
		"8761239812739812739801279371289371932",
		"-8761182763908473812974620938742386",
		"761239812739812739801279371289371932",
		"-761182763908473812974620938742386",
		"61239812739812739801279371289371932",
		"-61182763908473812974620938742386",
		"1239812739812739801279371289371932",
		"-1182763908473812974620938742386",
		"9812739812739801279371289371932",
		"-3908473812974620938742386",
		"98127373373209",
		"-4620938742386",
		"100000000000",
		"-100000000000",
		"10000000000",
		"-10000000000",
		"1000000000",
		"-1000000000",
		"100000000",
		"-100000000",
		"10000000",
		"-10000000",
		"1000000",
		"-1000000",
		"100000",
		"-100000",
		"10000",
		"-10000",
		"1000",
		"-1000",
		"100",
		"-100",
		"10",
		"-10",
		"100030000010",
		"-100000600004",
		"10000000100",
		"-10030000000",
		"1000040000",
		"-1000000000",
		"1010000001",
		"-1000000001",
		"1001001000",
		"-10000099",
		"99999",
		"9999",
		"999",
		"1011",
		"1009",
		"1709",
	}

	// Generate random bigints
	for i := 0; i < 1000; i++ {
		n := rand.Intn(30) + 1
		num := make([]byte, n)

		for j := 0; j < n; j++ {
			num[j] = "0123456789"[rand.Intn(10)]
		}

		t := strings.TrimLeft(string(num), "0")
		if t == "" {
			continue
		}

		// 33% chance for a negative number
		if rand.Intn(3) == 0 {
			t = "-" + t
		}

		samples = append(samples, t)
	}

	// Generate more random bigints consisting from mostly 0s
	for i := 0; i < 1000; i++ {
		n := rand.Intn(50) + 1
		num := make([]byte, n)

		for j := 0; j < n; j++ {
			k := rand.Intn(10)
			num[j] = "00000000000000000000000000000000000123456789"[k]
		}

		t := strings.TrimLeft(string(num), "0")
		if t == "" {
			continue
		}

		// 33% chance for a negative number
		if rand.Intn(3) == 0 {
			t = "-" + t
		}

		samples = append(samples, t)
	}

	for _, s := range samples {
		t.Run(s, func(t *testing.T) {
			i, ok := (&big.Int{}).SetString(s, 10)
			require.True(t, ok, "invalid big.Int literal: %v", s)
			require.Equal(t, s, i.String())

			var result Result
			err := conn.QueryOne(ctx, query, &result, i, s)
			assert.NoError(t, err)

			assert.True(t, result.IsEqual, "equality check faild")
			assert.Equal(t, s, result.Encoded, "encoding failed")
			assert.Equal(t, i, result.Decoded)
			assert.Equal(t, i, result.RoundTrip)
			assert.Equal(t, s, result.String)
			require.Equal(t, s, i.String(), "argument was mutated")
		})
	}
}

type CustomBigInt struct {
	data []byte
}

func (m CustomBigInt) MarshalEdgeDBBigInt() ([]byte, error) {
	return m.data, nil
}

func (m *CustomBigInt) UnmarshalEdgeDBBigInt(data []byte) error {
	m.data = data
	return nil
}

func TestSendAndReceiveBigIntMarshaler(t *testing.T) {
	ctx := context.Background()

	query := `SELECT (
		encoded := <str><bigint>$0,
		decoded := <bigint>-15000n,
	)`

	type Result struct {
		Encoded string       `edgedb:"encoded"`
		Decoded CustomBigInt `edgedb:"decoded"`
	}

	data := []byte{
		0x00, 0x02, // ndigits
		0x00, 0x01, // weight
		0x40, 0x00, // sign
		0x00, 0x00, // reserved
		0x00, 0x01, 0x13, 0x88, // digits
	}
	arg := &CustomBigInt{make([]byte, len(data))}
	copy(arg.data, data)

	var result Result
	err := conn.QueryOne(ctx, query, &result, arg)
	require.NoError(t, err)
	assert.Equal(t,
		Result{
			Encoded: `-15000`,
			Decoded: CustomBigInt{data},
		},
		result,
	)
}

type CustomDecimal struct {
	data []byte
}

func (d CustomDecimal) MarshalEdgeDBDecimal() ([]byte, error) {
	return d.data, nil
}

func (d *CustomDecimal) UnmarshalEdgeDBDecimal(data []byte) error {
	d.data = data
	return nil
}

func TestSendAndReceiveDecimalMarshaler(t *testing.T) {
	ctx := context.Background()

	data := []byte{
		0x00, 0x03, // ndigits
		0x00, 0x01, // weight
		0x40, 0x00, // sign
		0x00, 0x07, // dscale
		0x00, 0x01, 0x13, 0x88, 0x18, 0x6a, // digits
	}

	arg := CustomDecimal{make([]byte, len(data))}
	copy(arg.data, data)

	type Result struct {
		Decoded CustomDecimal `edgedb:"decoded"`
		Encoded string        `edgedb:"encoded"`
	}

	query := `SELECT (
		decoded := -15000.6250000n,
		encoded := <str><decimal>$0,
	)`

	var result Result
	err := conn.QueryOne(ctx, query, &result, arg)
	require.NoError(t, err)

	expected := CustomDecimal{make([]byte, len(data))}
	copy(expected.data, data)
	assert.Equal(t, Result{expected, "-15000.6250000"}, result)
}

func TestSendAndReceiveUUID(t *testing.T) {
	ctx := context.Background()

	query := `
		WITH
			id := <uuid>$0,
			s := <str>$1
		SELECT (
			encoded := <str>id,
			decoded := <uuid>s,
			round_trip := id,
			is_equal := <uuid>s = id,
			string := <str><uuid>s,
		)
	`

	type Result struct {
		Encoded   string `edgedb:"encoded"`
		Decoded   UUID   `edgedb:"decoded"`
		RoundTrip UUID   `edgedb:"round_trip"`
		IsEqual   bool   `edgedb:"is_equal"`
		String    string `edgedb:"string"`
	}

	samples := []string{
		"759637d8-6635-11e9-b9d4-098002d459d5",
		"00000000-0000-0000-0000-000000000000",
		"ffffffff-ffff-ffff-ffff-ffffffffffff",
	}

	for i := 0; i < 1000; i++ {
		var id UUID
		binary.BigEndian.PutUint64(id[:8], rand.Uint64())
		binary.BigEndian.PutUint64(id[8:], rand.Uint64())
		samples = append(samples, id.String())
	}

	for _, s := range samples {
		t.Run(s, func(t *testing.T) {
			var id UUID
			err := id.UnmarshalText([]byte(s))
			require.NoError(t, err)

			var result Result
			err = conn.QueryOne(ctx, query, &result, id, s)
			assert.NoError(t, err)

			assert.True(t, result.IsEqual, "equality check faild")
			assert.Equal(t, s, result.Encoded, "encoding failed")
			assert.Equal(t, id, result.Decoded)
			assert.Equal(t, id, result.RoundTrip)
			assert.Equal(t, s, result.String)
			require.Equal(t, s, id.String(), "argument was mutated")
		})
	}
}

type CustomUUID struct {
	data [16]byte
}

func (m CustomUUID) MarshalEdgeDBUUID() ([]byte, error) {
	return m.data[:], nil
}

func (m *CustomUUID) UnmarshalEdgeDBUUID(data []byte) error {
	copy(m.data[:], data)
	return nil
}

func TestSendAndReceiveUUIDMarshaler(t *testing.T) {
	ctx := context.Background()

	query := `SELECT (
		encoded := <str><uuid>$0,
		decoded := <uuid>'b9545c35-1fe7-485f-a6ea-f8ead251abd3',
	)`

	type Result struct {
		Encoded string     `edgedb:"encoded"`
		Decoded CustomUUID `edgedb:"decoded"`
	}

	data := [16]byte{
		0xb9, 0x54, 0x5c, 0x35, 0x1f, 0xe7, 0x48, 0x5f,
		0xa6, 0xea, 0xf8, 0xea, 0xd2, 0x51, 0xab, 0xd3,
	}
	arg := &CustomUUID{}
	copy(arg.data[:], data[:])

	var result Result
	err := conn.QueryOne(ctx, query, &result, arg)
	require.NoError(t, err)
	assert.Equal(t,
		Result{
			Encoded: "b9545c35-1fe7-485f-a6ea-f8ead251abd3",
			Decoded: CustomUUID{data},
		},
		result,
	)
}

func TestSendAndReceiveCustomScalars(t *testing.T) {
	ctx := context.Background()

	query := `
		WITH
			i := <CustomInt64>$0,
			s := <str>$1,
		SELECT (
			encoded := <str>i,
			decoded := <CustomInt64>s,
			round_trip := i,
			is_equal := i = <CustomInt64>s,
		)
	`

	type Result struct {
		Encoded   string `edgedb:"encoded"`
		Decoded   int64  `edgedb:"decoded"`
		RoundTrip int64  `edgedb:"round_trip"`
		IsEqual   bool   `edgedb:"is_equal"`
	}

	samples := []int64{0, 1, 9223372036854775807, -9223372036854775808}

	for i := 0; i < 1000; i++ {
		samples = append(samples, int64(rand.Uint64()))
	}

	for _, i := range samples {
		s := fmt.Sprint(i)
		t.Run(s, func(t *testing.T) {
			var result Result
			err := conn.QueryOne(ctx, query, &result, i, s)

			assert.NoError(t, err)
			assert.Equal(t, s, result.Encoded)
			assert.Equal(t, i, result.Decoded)
			assert.Equal(t, i, result.Decoded)
			assert.True(t, result.IsEqual)
		})
	}
}

type CustomScalar struct {
	data [8]byte
}

func (m CustomScalar) MarshalEdgeDBInt64() ([]byte, error) {
	return m.data[:], nil
}

func (m *CustomScalar) UnmarshalEdgeDBInt64(data []byte) error {
	copy(m.data[:], data)
	return nil
}

func TestSendAndReceiveCustomScalarMarshaler(t *testing.T) {
	ctx := context.Background()

	query := `SELECT (
		encoded := <CustomInt64>$0,
		decoded := <CustomInt64>123_456_789_987_654_321,
	)`

	type Result struct {
		Encoded int64        `edgedb:"encoded"`
		Decoded CustomScalar `edgedb:"decoded"`
	}

	data := [8]byte{0x01, 0xb6, 0x9b, 0x4b, 0xe0, 0x52, 0xfa, 0xb1}
	arg := &CustomScalar{}
	copy(arg.data[:], data[:])

	var result Result
	err := conn.QueryOne(ctx, query, &result, arg)
	require.NoError(t, err)
	assert.Equal(t,
		Result{
			Encoded: 123_456_789_987_654_321,
			Decoded: CustomScalar{data},
		},
		result,
	)
}

func TestDecodeDeeplyNestedTuple(t *testing.T) {
	ctx := context.Background()
	query := "SELECT ([(1, 2), (3, 4)], (5, (6, 7)))"

	type Tuple struct {
		first  int64 `edgedb:"0"`
		second int64 `edgedb:"1"`
	}

	type OtherTuple struct {
		first  int64 `edgedb:"0"`
		second Tuple `edgedb:"1"`
	}

	type ParentTuple struct {
		first  []Tuple    `edgedb:"0"`
		second OtherTuple `edgedb:"1"`
	}

	var result ParentTuple
	err := conn.QueryOne(ctx, query, &result)
	require.NoError(t, err)

	expected := ParentTuple{
		first: []Tuple{
			{1, 2},
			{3, 4},
		},
		second: OtherTuple{5, Tuple{6, 7}},
	}

	assert.Equal(t, expected, result)
}

func TestReceiveObject(t *testing.T) {
	ctx := context.Background()

	query := `
		SELECT schema::Function {
			name,
			params: {
				kind,
				num,
				foo := 42,
			} ORDER BY .num ASC
		}
		FILTER .name = 'std::str_repeat'
		LIMIT 1
	`

	type Params struct {
		ID   UUID   `edgedb:"id"`
		Kind string `edgedb:"kind"`
		Num  int64  `edgedb:"num"`
		Foo  int64  `edgedb:"foo"`
	}

	type Function struct {
		ID     UUID          `edgedb:"id"`
		Name   string        `edgedb:"name"`
		Params []Params      `edgedb:"params"`
		Tuple  []interface{} `edgedb:"tuple"`
	}

	var result Function
	err := conn.QueryOne(ctx, query, &result)
	require.NoError(t, err)
	assert.Equal(t, "std::str_repeat", result.Name)
	assert.Equal(t, 2, len(result.Params))
	assert.Equal(t, "PositionalParam", result.Params[0].Kind)
	assert.Equal(t, int64(42), result.Params[1].Foo)
}

func TestReceiveNamedTuple(t *testing.T) {
	ctx := context.Background()

	type NamedTuple struct {
		A int64 `edgedb:"a"`
	}

	var result NamedTuple
	err := conn.QueryOne(ctx, "SELECT (a := 1,)", &result)
	require.NoError(t, err)
	assert.Equal(t, NamedTuple{A: 1}, result)
}

func TestReceiveTuple(t *testing.T) {
	ctx := context.Background()

	var wrongType string
	err := conn.QueryOne(ctx, `SELECT ()`, &wrongType)
	require.EqualError(t, err, "edgedb.UnsupportedFeatureError: "+
		"the \"out\" argument does not match query schema: "+
		"expected string to be a struct got string")

	var emptyStruct struct{}
	err = conn.QueryOne(ctx, `SELECT ()`, &emptyStruct)
	require.NoError(t, err)

	var missingTag struct{ first int64 }
	err = conn.QueryOne(ctx, `SELECT (<int64>$0,)`, &missingTag, int64(1))
	require.EqualError(t, err, "edgedb.UnsupportedFeatureError: "+
		"the \"out\" argument does not match query schema: "+
		"expected struct { first int64 } to have a field "+
		"with the tag `edgedb:\"0\"`")

	type NestedTuple struct {
		second bool    `edgedb:"1"`
		first  float64 `edgedb:"0"`
	}

	type Tuple struct {
		first  int64       `edgedb:"0"` // nolint:structcheck
		second string      `edgedb:"1"` // nolint:structcheck
		third  NestedTuple `edgedb:"2"` // nolint:structcheck
	}

	result := []Tuple{}
	err = conn.Query(ctx, `SELECT (<int64>$0,)`, &result, int64(1))
	require.NoError(t, err)
	assert.Equal(t, []Tuple{{first: 1}}, result)

	result = []Tuple{}
	err = conn.Query(ctx, `SELECT {(1, "abc"), (2, "def")}`, &result)
	require.NoError(t, err)
	require.Equal(t,
		[]Tuple{
			{first: 1, second: "abc"},
			{first: 2, second: "def"},
		},
		result,
	)

	result = []Tuple{}
	err = conn.Query(ctx, `SELECT (1, "abc", (2.3, true))`, &result)
	require.NoError(t, err)
	require.Equal(t,
		[]Tuple{{
			1,
			"abc",
			NestedTuple{
				first:  2.3,
				second: true,
			},
		}},
		result,
	)
}

func TestSendAndReceiveArray(t *testing.T) {
	ctx := context.Background()

	var result []int64
	err := conn.QueryOne(ctx, "SELECT <array<int64>>$0", &result, "hello")
	assert.EqualError(t, err,
		"edgedb.InvalidArgumentError: "+
			"expected args[0] to be a slice got: string")

	type Tuple struct {
		first []int64 `edgedb:"0"`
	}

	var nested Tuple
	err = conn.QueryOne(ctx, "SELECT (<array<int64>>$0,)", &nested, []int64{1})
	require.NoError(t, err)
	assert.Equal(t, Tuple{[]int64{1}}, nested)

	err = conn.QueryOne(ctx, "SELECT <array<int64>>$0", &result, []int64(nil))
	require.NoError(t, err)
	assert.Equal(t, []int64(nil), result)

	err = conn.QueryOne(ctx, "SELECT <array<int64>>$0", &result, []int64{1})
	require.NoError(t, err)
	assert.Equal(t, []int64{1}, result)

	arg := []int64{1, 2, 3}
	err = conn.QueryOne(ctx, "SELECT <array<int64>>$0", &result, arg)
	require.NoError(t, err)
	assert.Equal(t, []int64{1, 2, 3}, result)
}

func TestReceiveSet(t *testing.T) {
	ctx := context.Background()

	// decoding using pointers
	{
		type Function struct {
			ID   UUID      `edgedb:"id"`
			Sets [][]int64 `edgedb:"sets"`
		}

		query := `
			SELECT schema::Function {
				id,
				sets := {[1, 2], [1]}
			}
			LIMIT 1
		`

		var result Function
		err := conn.QueryOne(ctx, query, &result)
		require.NoError(t, err)
		assert.Equal(t, [][]int64{{1, 2}, {1}}, result.Sets)
	}

	// decoding using reflect
	{
		type NestedTuple struct {
			first int64 `edgedb:"0"`
		}

		type Tuple struct {
			first  int64       `edgedb:"0"` // nolint:structcheck
			second NestedTuple `edgedb:"1"` // nolint:structcheck
		}

		type Function struct {
			ID   UUID      `edgedb:"id"`
			Sets [][]Tuple `edgedb:"sets"`
		}

		query := `
			SELECT schema::Function {
				id,
				sets := {[(1, (2,))], [(3, (4,))]}
			}
			LIMIT 1
		`

		var result Function
		err := conn.QueryOne(ctx, query, &result)
		require.NoError(t, err)
		assert.Equal(t,
			[][]Tuple{
				{{1, NestedTuple{2}}},
				{{3, NestedTuple{4}}},
			},
			result.Sets,
		)
	}
}
