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

package gel

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/big"
	"math/rand"
	"strconv"
	"strings"
	"testing"
	"time"

	types "github.com/edgedb/edgedb-go/internal/geltypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMissmatchedUnmarshalerType(t *testing.T) {
	ctx := context.Background()

	// Decode into wrong Unmarshaler type
	var wrongType struct {
		Val CustomInt32 `gel:"val"`
	}
	err := client.QuerySingle(ctx, `
		SELECT { val := 123_456_789_987_654_321 }`,
		&wrongType,
	)
	assert.EqualError(t, err, "gel.InvalidArgumentError: "+
		"the \"out\" argument does not match query schema: expected "+
		"struct { Val gel.CustomInt32 \"gel:\\\"val\\\"\" }.val "+
		"to be int64 or gel.OptionalInt64 got gel.CustomInt32")
	assert.Equal(t, []byte(nil), wrongType.Val.data)
}

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
		Encoded   string `gel:"encoded"`
		Decoded   int64  `gel:"decoded"`
		RoundTrip int64  `gel:"round_trip"`
		IsEqual   bool   `gel:"is_equal"`
		String    string `gel:"string"`
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
	err := client.Query(ctx, query, &results, numbers, strings)
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
	data []byte
}

func (m CustomInt64) MarshalEdgeDBInt64() ([]byte, error) {
	data := make([]byte, len(m.data))
	copy(data, m.data)
	return data, nil
}

func (m *CustomInt64) UnmarshalEdgeDBInt64(data []byte) error {
	m.data = make([]byte, len(data))
	copy(m.data, data)
	return nil
}

func TestReceiveInt64Unmarshaler(t *testing.T) {
	ctx := context.Background()
	var result struct {
		Val CustomInt64 `gel:"val"`
	}

	// Decode value
	err := client.QuerySingle(ctx, `
		SELECT { val := 123_456_789_987_654_321 }`,
		&result,
	)
	assert.NoError(t, err)
	assert.Equal(t,
		[]byte{0x01, 0xb6, 0x9b, 0x4b, 0xe0, 0x52, 0xfa, 0xb1},
		result.Val.data,
	)

	// Decode missing value
	err = client.QuerySingle(ctx, `
		SELECT { val := <OPTIONAL int64>$0 }`,
		&result,
		types.OptionalInt64{},
	)
	assert.EqualError(t, err, "gel.InvalidArgumentError: "+
		"the \"out\" argument does not match query schema: "+
		"expected gel.CustomInt64 at "+
		"struct { Val gel.CustomInt64 \"gel:\\\"val\\\"\" }.val "+
		"to be OptionalUnmarshaler interface "+
		"because the field is not required")
}

func TestSendInt64Marshaler(t *testing.T) {
	ctx := context.Background()
	var result struct {
		Val types.OptionalInt64 `gel:"val"`
	}

	// encode value into required argument
	err := client.QuerySingle(ctx, `
		SELECT { val := <int64>$0 }`,
		&result,
		CustomInt64{
			data: []byte{0x01, 0xb6, 0x9b, 0x4b, 0xe0, 0x52, 0xfa, 0xb1},
		},
	)
	assert.NoError(t, err)
	assert.Equal(
		t,
		types.NewOptionalInt64(123_456_789_987_654_321),
		result.Val,
	)

	// encode value into optional argument
	err = client.QuerySingle(ctx, `
		SELECT { val := <OPTIONAL int64>$0 }`,
		&result,
		CustomInt64{
			data: []byte{0x01, 0xb6, 0x9b, 0x4b, 0xe0, 0x52, 0xfa, 0xb1},
		},
	)
	assert.NoError(t, err)
	assert.Equal(
		t,
		types.NewOptionalInt64(123_456_789_987_654_321),
		result.Val,
	)

	// encode wrong number of bytes
	err = client.QuerySingle(ctx, `
		SELECT { val := <int64>$0 }`,
		&result,
		CustomInt64{data: []byte{0x01}},
	)
	assert.EqualError(t, err, "gel.InvalidArgumentError: "+
		"wrong number of bytes encoded by gel.CustomInt64 "+
		"at args[0] expected 8, got 1")
}

type CustomOptionalInt64 struct {
	data  []byte
	isSet bool
}

func (m CustomOptionalInt64) MarshalEdgeDBInt64() ([]byte, error) {
	if !m.isSet {
		return nil, fmt.Errorf("%T is not set", m)
	}
	data := make([]byte, len(m.data))
	copy(data, m.data)
	return data, nil
}

func (m *CustomOptionalInt64) UnmarshalEdgeDBInt64(data []byte) error {
	m.isSet = true
	m.data = make([]byte, len(data))
	copy(m.data, data)
	return nil
}

func (m *CustomOptionalInt64) SetMissing(missing bool) {
	m.isSet = !missing
	m.data = nil
}

func (m CustomOptionalInt64) Missing() bool { return !m.isSet }

func TestReceiveOptionalInt64Unmarshaler(t *testing.T) {
	ddl := `CREATE TYPE Sample { CREATE PROPERTY val -> int64; };`
	inRolledBackTx(t, ddl, func(ctx context.Context, tx *Tx) {
		var result struct {
			Val CustomOptionalInt64 `gel:"val"`
		}

		// Decode value
		err := tx.QuerySingle(ctx, `
			SELECT { val := 123_456_789_987_654_321 }`,
			&result,
		)
		assert.NoError(t, err)
		assert.Equal(t,
			[]byte{0x01, 0xb6, 0x9b, 0x4b, 0xe0, 0x52, 0xfa, 0xb1},
			result.Val.data,
		)

		// Decode missing value
		query := `WITH inserted := (INSERT Sample) SELECT inserted { val }`
		err = tx.QuerySingle(ctx, query, &result)
		assert.NoError(t, err)
		assert.Equal(t, CustomOptionalInt64{}, result.Val)
	})
}

func TestSendOptionalInt64Marshaler(t *testing.T) {
	ctx := context.Background()
	var result struct {
		Val types.OptionalInt64 `gel:"val"`
	}

	newValue := func(data []byte) CustomOptionalInt64 {
		return CustomOptionalInt64{isSet: true, data: data}
	}

	// encode value into required argument
	err := client.QuerySingle(ctx, `
		SELECT { val := <int64>$0 }`,
		&result,
		newValue([]byte{0x01, 0xb6, 0x9b, 0x4b, 0xe0, 0x52, 0xfa, 0xb1}),
	)
	assert.NoError(t, err)
	assert.Equal(
		t,
		types.NewOptionalInt64(123_456_789_987_654_321),
		result.Val,
	)

	// encode value into optional argument
	err = client.QuerySingle(ctx, `
		SELECT { val := <OPTIONAL int64>$0 }`,
		&result,
		newValue([]byte{0x01, 0xb6, 0x9b, 0x4b, 0xe0, 0x52, 0xfa, 0xb1}),
	)
	assert.NoError(t, err)
	assert.Equal(
		t,
		types.NewOptionalInt64(123_456_789_987_654_321),
		result.Val,
	)

	// encode missing value into optional argument
	err = client.QuerySingle(ctx, `
		SELECT { val := <OPTIONAL int64>$0 }`,
		&result,
		CustomOptionalInt64{},
	)
	assert.NoError(t, err)
	assert.Equal(t, types.OptionalInt64{}, result.Val)

	// encode missing value into required argument
	err = client.QuerySingle(ctx, `
		SELECT { val := <int64>$0 }`,
		&result,
		CustomOptionalInt64{},
	)
	assert.EqualError(t, err, "gel.InvalidArgumentError: "+
		"cannot encode gel.CustomOptionalInt64 at args[0] "+
		"because its value is missing")

	// encode wrong number of bytes with required argument
	err = client.QuerySingle(ctx, `
		SELECT { val := <int64>$0 }`,
		&result,
		newValue([]byte{0x01}),
	)
	assert.EqualError(t, err, "gel.InvalidArgumentError: "+
		"wrong number of bytes encoded by gel.CustomOptionalInt64 "+
		"at args[0] expected 8, got 1")

	// encode wrong number of bytes with optional argument
	err = client.QuerySingle(ctx, `
		SELECT { val := <OPTIONAL int64>$0 }`,
		&result,
		newValue([]byte{0x01}),
	)
	assert.EqualError(t, err, "gel.InvalidArgumentError: "+
		"wrong number of bytes encoded by gel.CustomOptionalInt64 "+
		"at args[0] expected 8, got 1")
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
		Encoded   string `gel:"encoded"`
		Decoded   int32  `gel:"decoded"`
		RoundTrip int32  `gel:"round_trip"`
		IsEqual   bool   `gel:"is_equal"`
		String    string `gel:"string"`
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
	err := client.Query(ctx, query, &results, numbers, strings)
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
	data []byte
}

func (m CustomInt32) MarshalEdgeDBInt32() ([]byte, error) {
	data := make([]byte, len(m.data))
	copy(data, m.data)
	return data, nil
}

func (m *CustomInt32) UnmarshalEdgeDBInt32(data []byte) error {
	m.data = make([]byte, len(data))
	copy(m.data, data)
	return nil
}

func TestReceiveInt32Unmarshaler(t *testing.T) {
	ctx := context.Background()
	var result struct {
		Val CustomInt32 `gel:"val"`
	}

	// Decode value
	err := client.QuerySingle(ctx, `SELECT { val := <int32>655_665 }`, &result)
	assert.NoError(t, err)
	assert.Equal(t,
		[]byte{0x00, 0x0a, 0x01, 0x31},
		result.Val.data,
	)

	// Decode missing value
	err = client.QuerySingle(ctx, `
		SELECT { val := <OPTIONAL int32>$0 }`,
		&result,
		types.OptionalInt32{},
	)
	assert.EqualError(t, err, "gel.InvalidArgumentError: "+
		"the \"out\" argument does not match query schema: "+
		"expected gel.CustomInt32 at "+
		"struct { Val gel.CustomInt32 \"gel:\\\"val\\\"\" }.val "+
		"to be OptionalUnmarshaler interface "+
		"because the field is not required")
}

func TestSendInt32Marshaler(t *testing.T) {
	ctx := context.Background()
	var result struct {
		Val types.OptionalInt32 `gel:"val"`
	}

	// encode value into required argument
	err := client.QuerySingle(ctx, `
		SELECT { val := <int32>$0 }`,
		&result,
		CustomInt32{data: []byte{0x00, 0x0a, 0x01, 0x31}},
	)
	assert.NoError(t, err)
	assert.Equal(t, types.NewOptionalInt32(655_665), result.Val)

	// encode value into optional argument
	err = client.QuerySingle(ctx, `
		SELECT { val := <OPTIONAL int32>$0 }`,
		&result,
		CustomInt32{data: []byte{0x00, 0x0a, 0x01, 0x31}},
	)
	assert.NoError(t, err)
	assert.Equal(t, types.NewOptionalInt32(655_665), result.Val)

	// encode wrong number of bytes
	err = client.QuerySingle(ctx, `
		SELECT { val := <int32>$0 }`,
		&result,
		CustomInt32{data: []byte{0x01}},
	)
	assert.EqualError(t, err, "gel.InvalidArgumentError: "+
		"wrong number of bytes encoded by gel.CustomInt32 "+
		"at args[0] expected 4, got 1")
}

type CustomOptionalInt32 struct {
	data  []byte
	isSet bool
}

func (m CustomOptionalInt32) MarshalEdgeDBInt32() ([]byte, error) {
	if !m.isSet {
		return nil, fmt.Errorf("%T is not set", m)
	}
	data := make([]byte, len(m.data))
	copy(data, m.data)
	return data, nil
}

func (m *CustomOptionalInt32) UnmarshalEdgeDBInt32(data []byte) error {
	m.isSet = true
	m.data = make([]byte, len(data))
	copy(m.data, data)
	return nil
}

func (m *CustomOptionalInt32) SetMissing(missing bool) {
	m.isSet = !missing
	m.data = nil
}

func (m CustomOptionalInt32) Missing() bool { return !m.isSet }

func TestReceiveOptionalInt32Unmarshaler(t *testing.T) {
	ddl := `CREATE TYPE Sample { CREATE PROPERTY val -> int32; };`
	inRolledBackTx(t, ddl, func(ctx context.Context, tx *Tx) {
		var result struct {
			Val CustomOptionalInt32 `gel:"val"`
		}

		// Decode value
		err := tx.QuerySingle(ctx, `SELECT { val := <int32>655_665 }`, &result)
		assert.NoError(t, err)
		assert.Equal(t, []byte{0x00, 0x0a, 0x01, 0x31}, result.Val.data)

		// Decode missing value
		query := `WITH inserted := (INSERT Sample) SELECT inserted { val }`
		err = tx.QuerySingle(ctx, query, &result)
		assert.NoError(t, err)
		assert.Equal(t, CustomOptionalInt32{}, result.Val)
	})
}

func TestSendOptionalInt32Marshaler(t *testing.T) {
	ctx := context.Background()
	var result struct {
		Val types.OptionalInt32 `gel:"val"`
	}

	newValue := func(data []byte) CustomOptionalInt32 {
		return CustomOptionalInt32{isSet: true, data: data}
	}

	// encode value into required argument
	err := client.QuerySingle(ctx, `
		SELECT { val := <int32>$0 }`,
		&result,
		newValue([]byte{0x00, 0x0a, 0x01, 0x31}),
	)
	assert.NoError(t, err)
	assert.Equal(t, types.NewOptionalInt32(655_665), result.Val)

	// encode value into optional argument
	err = client.QuerySingle(ctx, `
		SELECT { val := <OPTIONAL int32>$0 }`,
		&result,
		newValue([]byte{0x00, 0x0a, 0x01, 0x31}),
	)
	assert.NoError(t, err)
	assert.Equal(t, types.NewOptionalInt32(655_665), result.Val)

	// encode missing value into optional argument
	err = client.QuerySingle(ctx, `
		SELECT { val := <OPTIONAL int32>$0 }`,
		&result,
		CustomOptionalInt32{},
	)
	assert.NoError(t, err)
	assert.Equal(t, types.OptionalInt32{}, result.Val)

	// encode missing value into required argument
	err = client.QuerySingle(ctx, `
		SELECT { val := <int32>$0 }`,
		&result,
		CustomOptionalInt32{},
	)
	assert.EqualError(t, err, "gel.InvalidArgumentError: "+
		"cannot encode gel.CustomOptionalInt32 at args[0] "+
		"because its value is missing")

	// encode wrong number of bytes with required argument
	err = client.QuerySingle(ctx, `
		SELECT { val := <int32>$0 }`,
		&result,
		newValue([]byte{0x01}),
	)
	assert.EqualError(t, err, "gel.InvalidArgumentError: "+
		"wrong number of bytes encoded by gel.CustomOptionalInt32 "+
		"at args[0] expected 4, got 1")

	// encode wrong number of bytes with optional argument
	err = client.QuerySingle(ctx, `
		SELECT { val := <OPTIONAL int32>$0 }`,
		&result,
		newValue([]byte{0x01}),
	)
	assert.EqualError(t, err, "gel.InvalidArgumentError: "+
		"wrong number of bytes encoded by gel.CustomOptionalInt32 "+
		"at args[0] expected 4, got 1")
}

func TestSendAndReceiveOptionalInt32(t *testing.T) {
	ctx := context.Background()

	err := client.Tx(ctx, func(ctx context.Context, tx *Tx) error {
		e := tx.Execute(ctx, `
			CREATE TYPE Int32FieldHolder {
				CREATE PROPERTY int32 -> int32;
			};

			INSERT Int32FieldHolder;
		`)
		if e != nil {
			return e
		}

		type Result struct {
			Int32 types.OptionalInt32 `gel:"int32"`
		}

		var result Result
		e = tx.QuerySingle(ctx, `
			# decode missing optional
			SELECT Int32FieldHolder { int32 } LIMIT 1`,
			&result,
		)
		if e != nil {
			return e
		}
		assert.Equal(t, Result{}, result)

		e = tx.QuerySingle(ctx, `
			# encode unset optional
			SELECT Int32FieldHolder {
				int32 := <OPTIONAL int32>$0
			} LIMIT 1`,
			&result,
			types.OptionalInt32{},
		)
		if e != nil {
			return e
		}
		assert.Equal(t, Result{}, result)

		e = tx.QuerySingle(ctx, `
			# encode set optional
			SELECT Int32FieldHolder { int32 := <OPTIONAL int32>$0 } LIMIT 1`,
			&result,
			types.NewOptionalInt32(32),
		)
		if e != nil {
			return e
		}
		assert.Equal(t, Result{Int32: types.NewOptionalInt32(32)}, result)

		e = tx.QuerySingle(ctx, `
			# encode set optional into required argument
			SELECT Int32FieldHolder { int32 := <int32>$0 } LIMIT 1`,
			&result,
			types.NewOptionalInt32(32),
		)
		if e != nil {
			return e
		}
		assert.Equal(t, Result{Int32: types.NewOptionalInt32(32)}, result)

		e = tx.QuerySingle(ctx, `
			# encode unset optional into required argument
			SELECT Int32FieldHolder { int32 := <int32>$0 } LIMIT 1`,
			&result,
			types.OptionalInt32{},
		)
		assert.EqualError(t, e, "gel.InvalidArgumentError: "+
			"cannot encode gel.OptionalInt32 at args[0] "+
			"because its value is missing")

		return errors.New("rollback")
	})

	assert.EqualError(t, err, "rollback")
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
		Encoded   string `gel:"encoded"`
		Decoded   int16  `gel:"decoded"`
		RoundTrip int16  `gel:"round_trip"`
		IsEqual   bool   `gel:"is_equal"`
		String    string `gel:"string"`
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
	err := client.Query(ctx, query, &results, numbers, strings)
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
	data []byte
}

func (m CustomInt16) MarshalEdgeDBInt16() ([]byte, error) {
	data := make([]byte, len(m.data))
	copy(data, m.data)
	return data, nil
}

func (m *CustomInt16) UnmarshalEdgeDBInt16(data []byte) error {
	m.data = make([]byte, len(data))
	copy(m.data, data)
	return nil
}

func TestReceiveInt16Unmarshaler(t *testing.T) {
	ctx := context.Background()
	var result struct {
		Val CustomInt16 `gel:"val"`
	}

	// Decode value
	err := client.QuerySingle(ctx, `SELECT { val := <int16>6_556 }`, &result)
	assert.NoError(t, err)
	assert.Equal(t, []byte{0x19, 0x9c}, result.Val.data)

	// Decode missing value
	err = client.QuerySingle(ctx, `
		SELECT { val := <OPTIONAL int16>$0 }`,
		&result,
		types.OptionalInt16{},
	)
	assert.EqualError(t, err, "gel.InvalidArgumentError: "+
		"the \"out\" argument does not match query schema: "+
		"expected gel.CustomInt16 at "+
		"struct { Val gel.CustomInt16 \"gel:\\\"val\\\"\" }.val "+
		"to be OptionalUnmarshaler interface "+
		"because the field is not required")
}

func TestSendInt16Marshaler(t *testing.T) {
	ctx := context.Background()
	var result struct {
		Val types.OptionalInt16 `gel:"val"`
	}

	// encode value into required argument
	err := client.QuerySingle(ctx, `
		SELECT { val := <int16>$0 }`,
		&result,
		CustomInt16{data: []byte{0x19, 0x9c}},
	)
	assert.NoError(t, err)
	assert.Equal(t, types.NewOptionalInt16(6_556), result.Val)

	// encode value into optional argument
	err = client.QuerySingle(ctx, `
		SELECT { val := <OPTIONAL int16>$0 }`,
		&result,
		CustomInt16{data: []byte{0x19, 0x9c}},
	)
	assert.NoError(t, err)
	assert.Equal(t, types.NewOptionalInt16(6_556), result.Val)

	// encode wrong number of bytes
	err = client.QuerySingle(ctx, `
		SELECT { val := <int16>$0 }`,
		&result,
		CustomInt16{data: []byte{0x01}},
	)
	assert.EqualError(t, err, "gel.InvalidArgumentError: "+
		"wrong number of bytes encoded by gel.CustomInt16 "+
		"at args[0] expected 2, got 1")
}

type CustomOptionalInt16 struct {
	data  []byte
	isSet bool
}

func (m CustomOptionalInt16) MarshalEdgeDBInt16() ([]byte, error) {
	if !m.isSet {
		return nil, fmt.Errorf("%T is not set", m)
	}
	data := make([]byte, len(m.data))
	copy(data, m.data)
	return data, nil
}

func (m *CustomOptionalInt16) UnmarshalEdgeDBInt16(data []byte) error {
	m.isSet = true
	m.data = make([]byte, len(data))
	copy(m.data, data)
	return nil
}

func (m *CustomOptionalInt16) SetMissing(missing bool) {
	m.isSet = !missing
	m.data = nil
}

func (m CustomOptionalInt16) Missing() bool { return !m.isSet }

func TestReceiveOptionalInt16Unmarshaler(t *testing.T) {
	ddl := `CREATE TYPE Sample { CREATE PROPERTY val -> int16; };`
	inRolledBackTx(t, ddl, func(ctx context.Context, tx *Tx) {
		var result struct {
			Val CustomOptionalInt16 `gel:"val"`
		}

		// Decode value
		err := tx.QuerySingle(ctx, `SELECT { val := <int16>6_556 }`, &result)
		assert.NoError(t, err)
		assert.Equal(t, []byte{0x19, 0x9c}, result.Val.data)

		// Decode missing value
		query := `WITH inserted := (INSERT Sample) SELECT inserted { val }`
		err = tx.QuerySingle(ctx, query, &result)
		assert.NoError(t, err)
		assert.Equal(t, CustomOptionalInt16{}, result.Val)
	})
}

func TestSendOptionalInt16Marshaler(t *testing.T) {
	ctx := context.Background()
	var result struct {
		Val types.OptionalInt16 `gel:"val"`
	}

	newValue := func(data []byte) CustomOptionalInt16 {
		return CustomOptionalInt16{isSet: true, data: data}
	}

	// encode value into required argument
	err := client.QuerySingle(ctx, `
		SELECT { val := <int16>$0 }`,
		&result,
		newValue([]byte{0x19, 0x9c}),
	)
	assert.NoError(t, err)
	assert.Equal(t, types.NewOptionalInt16(6_556), result.Val)

	// encode value into optional argument
	err = client.QuerySingle(ctx, `
		SELECT { val := <OPTIONAL int16>$0 }`,
		&result,
		newValue([]byte{0x19, 0x9c}),
	)
	assert.NoError(t, err)
	assert.Equal(t, types.NewOptionalInt16(6_556), result.Val)

	// encode missing value into optional argument
	err = client.QuerySingle(ctx, `
		SELECT { val := <OPTIONAL int16>$0 }`,
		&result,
		CustomOptionalInt16{},
	)
	assert.NoError(t, err)
	assert.Equal(t, types.OptionalInt16{}, result.Val)

	// encode missing value into required argument
	err = client.QuerySingle(ctx, `
		SELECT { val := <int16>$0 }`,
		&result,
		CustomOptionalInt16{},
	)
	assert.EqualError(t, err, "gel.InvalidArgumentError: "+
		"cannot encode gel.CustomOptionalInt16 at args[0] "+
		"because its value is missing")

	// encode wrong number of bytes with required argument
	err = client.QuerySingle(ctx, `
		SELECT { val := <int16>$0 }`,
		&result,
		newValue([]byte{0x01}),
	)
	assert.EqualError(t, err, "gel.InvalidArgumentError: "+
		"wrong number of bytes encoded by gel.CustomOptionalInt16 "+
		"at args[0] expected 2, got 1")

	// encode wrong number of bytes with optional argument
	err = client.QuerySingle(ctx, `
		SELECT { val := <OPTIONAL int16>$0 }`,
		&result,
		newValue([]byte{0x01}),
	)
	assert.EqualError(t, err, "gel.InvalidArgumentError: "+
		"wrong number of bytes encoded by gel.CustomOptionalInt16 "+
		"at args[0] expected 2, got 1")
}

func TestSendAndReceiveOptionalInt16(t *testing.T) {
	ctx := context.Background()

	err := client.Tx(ctx, func(ctx context.Context, tx *Tx) error {
		e := tx.Execute(ctx, `
			CREATE TYPE Int16FieldHolder {
				CREATE PROPERTY int16 -> int16;
			};

			INSERT Int16FieldHolder;
		`)
		if e != nil {
			return e
		}

		type Result struct {
			Int16 types.OptionalInt16 `gel:"int16"`
		}

		var result Result
		e = tx.QuerySingle(ctx, `
			# decode missing optional
			SELECT Int16FieldHolder { int16 } LIMIT 1`,
			&result,
		)
		if e != nil {
			return e
		}
		assert.Equal(t, Result{}, result)

		e = tx.QuerySingle(ctx, `
			# encode unset optional
			SELECT Int16FieldHolder {
				int16 := <OPTIONAL int16>$0
			} LIMIT 1`,
			&result,
			types.OptionalInt16{},
		)
		if e != nil {
			return e
		}
		assert.Equal(t, Result{}, result)

		e = tx.QuerySingle(ctx, `
			# encode set optional
			SELECT Int16FieldHolder { int16 := <OPTIONAL int16>$0 } LIMIT 1`,
			&result,
			types.NewOptionalInt16(16),
		)
		if e != nil {
			return e
		}
		assert.Equal(t, Result{Int16: types.NewOptionalInt16(16)}, result)

		e = tx.QuerySingle(ctx, `
			# encode set optional into required argument
			SELECT Int16FieldHolder { int16 := <int16>$0 } LIMIT 1`,
			&result,
			types.NewOptionalInt16(16),
		)
		if e != nil {
			return e
		}
		assert.Equal(t, Result{Int16: types.NewOptionalInt16(16)}, result)

		e = tx.QuerySingle(ctx, `
			# encode unset optional into required argument
			SELECT Int16FieldHolder { int16 := <int16>$0 } LIMIT 1`,
			&result,
			types.OptionalInt16{},
		)
		assert.EqualError(t, e, "gel.InvalidArgumentError: "+
			"cannot encode gel.OptionalInt16 at args[0] "+
			"because its value is missing")

		return errors.New("rollback")
	})

	assert.EqualError(t, err, "rollback")
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
		Encoded   string `gel:"encoded"`
		Decoded   bool   `gel:"decoded"`
		RoundTrip bool   `gel:"round_trip"`
		IsEqual   bool   `gel:"is_equal"`
		String    string `gel:"string"`
	}

	samples := []bool{true, false}

	for _, i := range samples {
		s := fmt.Sprint(i)
		t.Run(s, func(t *testing.T) {
			var result Result
			err := client.QuerySingle(ctx, query, &result, i, s)
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
	data []byte
}

func (m CustomBool) MarshalEdgeDBBool() ([]byte, error) {
	data := make([]byte, len(m.data))
	copy(data, m.data)
	return data, nil
}

func (m *CustomBool) UnmarshalEdgeDBBool(data []byte) error {
	m.data = make([]byte, len(data))
	copy(m.data, data)
	return nil
}

func TestReceiveBoolUnmarshaler(t *testing.T) {
	ctx := context.Background()
	var result struct {
		Val CustomBool `gel:"val"`
	}

	// Decode value
	err := client.QuerySingle(ctx, `SELECT { val := true }`, &result)
	assert.NoError(t, err)
	assert.Equal(t, []byte{0x01}, result.Val.data)

	// Decode missing value
	err = client.QuerySingle(ctx, `
		SELECT { val := <OPTIONAL bool>$0 }`,
		&result,
		types.OptionalBool{},
	)
	assert.EqualError(t, err, "gel.InvalidArgumentError: "+
		"the \"out\" argument does not match query schema: "+
		"expected gel.CustomBool at "+
		"struct { Val gel.CustomBool \"gel:\\\"val\\\"\" }.val "+
		"to be OptionalUnmarshaler interface "+
		"because the field is not required")
}

func TestSendBoolMarshaler(t *testing.T) {
	ctx := context.Background()
	var result struct {
		Val types.OptionalBool `gel:"val"`
	}

	// encode value into required argument
	err := client.QuerySingle(ctx, `
		SELECT { val := <bool>$0 }`,
		&result,
		CustomBool{data: []byte{0x01}},
	)
	assert.NoError(t, err)
	assert.Equal(t, types.NewOptionalBool(true), result.Val)

	// encode value into optional argument
	err = client.QuerySingle(ctx, `
		SELECT { val := <OPTIONAL bool>$0 }`,
		&result,
		CustomBool{data: []byte{0x01}},
	)
	assert.NoError(t, err)
	assert.Equal(t, types.NewOptionalBool(true), result.Val)

	// encode wrong number of bytes
	err = client.QuerySingle(ctx, `
		SELECT { val := <bool>$0 }`,
		&result,
		CustomBool{data: []byte{0x01, 0x02}},
	)
	assert.EqualError(t, err, "gel.InvalidArgumentError: "+
		"wrong number of bytes encoded by gel.CustomBool "+
		"at args[0] expected 1, got 2")
}

type CustomOptionalBool struct {
	CustomBool
	isSet bool
}

func (m CustomOptionalBool) MarshalEdgeDBBool() ([]byte, error) {
	if !m.isSet {
		return nil, fmt.Errorf("%T is not set", m)
	}
	return m.CustomBool.MarshalEdgeDBBool()
}

func (m *CustomOptionalBool) UnmarshalEdgeDBBool(data []byte) error {
	m.isSet = true
	return m.CustomBool.UnmarshalEdgeDBBool(data)
}

func (m *CustomOptionalBool) SetMissing(missing bool) {
	m.isSet = !missing
	m.data = nil
}

func (m CustomOptionalBool) Missing() bool { return !m.isSet }

func TestReceiveOptionalBoolUnmarshaler(t *testing.T) {
	ddl := `CREATE TYPE Sample { CREATE PROPERTY val -> bool; };`
	inRolledBackTx(t, ddl, func(ctx context.Context, tx *Tx) {
		var result struct {
			Val CustomOptionalBool `gel:"val"`
		}

		// Decode value
		err := tx.QuerySingle(ctx, `SELECT { val := true }`, &result)
		assert.NoError(t, err)
		assert.Equal(t, []byte{0x01}, result.Val.data)

		// Decode missing value
		query := `WITH inserted := (INSERT Sample) SELECT inserted { val }`
		err = tx.QuerySingle(ctx, query, &result)
		assert.NoError(t, err)
		assert.Equal(t, CustomOptionalBool{}, result.Val)
	})
}

func TestSendOptionalBoolMarshaler(t *testing.T) {
	ctx := context.Background()
	var result struct {
		Val types.OptionalBool `gel:"val"`
	}

	newValue := func(data []byte) CustomOptionalBool {
		return CustomOptionalBool{
			isSet:      true,
			CustomBool: CustomBool{data: data},
		}
	}

	// encode value into required argument
	err := client.QuerySingle(ctx, `
		SELECT { val := <bool>$0 }`,
		&result,
		newValue([]byte{0x01}),
	)
	assert.NoError(t, err)
	assert.Equal(t, types.NewOptionalBool(true), result.Val)

	// encode value into optional argument
	err = client.QuerySingle(ctx, `
		SELECT { val := <OPTIONAL bool>$0 }`,
		&result,
		newValue([]byte{0x01}),
	)
	assert.NoError(t, err)
	assert.Equal(t, types.NewOptionalBool(true), result.Val)

	// encode missing value into optional argument
	err = client.QuerySingle(ctx, `
		SELECT { val := <OPTIONAL bool>$0 }`,
		&result,
		CustomOptionalBool{},
	)
	assert.NoError(t, err)
	assert.Equal(t, types.OptionalBool{}, result.Val)

	// encode missing value into required argument
	err = client.QuerySingle(ctx, `
		SELECT { val := <bool>$0 }`,
		&result,
		CustomOptionalBool{},
	)
	assert.EqualError(t, err, "gel.InvalidArgumentError: "+
		"cannot encode gel.CustomOptionalBool at args[0] "+
		"because its value is missing")

	// encode wrong number of bytes with required argument
	err = client.QuerySingle(ctx, `
		SELECT { val := <bool>$0 }`,
		&result,
		newValue([]byte{0x01, 0x02}),
	)
	assert.EqualError(t, err, "gel.InvalidArgumentError: "+
		"wrong number of bytes encoded by gel.CustomOptionalBool "+
		"at args[0] expected 1, got 2")

	// encode wrong number of bytes with optional argument
	err = client.QuerySingle(ctx, `
		SELECT { val := <OPTIONAL bool>$0 }`,
		&result,
		newValue([]byte{0x01, 0x02}),
	)
	assert.EqualError(t, err, "gel.InvalidArgumentError: "+
		"wrong number of bytes encoded by gel.CustomOptionalBool "+
		"at args[0] expected 1, got 2")
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
		Encoded   string  `gel:"encoded"`
		Decoded   float64 `gel:"decoded"`
		RoundTrip float64 `gel:"round_trip"`
		IsEqual   bool    `gel:"is_equal"`
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
	err := client.Query(ctx, query, &results, numbers, strings)
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
	data []byte
}

func (m CustomFloat64) MarshalEdgeDBFloat64() ([]byte, error) {
	data := make([]byte, len(m.data))
	copy(data, m.data)
	return data, nil
}

func (m *CustomFloat64) UnmarshalEdgeDBFloat64(data []byte) error {
	m.data = make([]byte, len(data))
	copy(m.data, data)
	return nil
}

func TestReceiveFloat64Unmarshaler(t *testing.T) {
	ctx := context.Background()
	var result struct {
		Val CustomFloat64 `gel:"val"`
	}

	// Decode value
	query := `SELECT { val := <float64>-15.625 }`
	err := client.QuerySingle(ctx, query, &result)
	assert.NoError(t, err)
	assert.Equal(t,
		[]byte{0xc0, 0x2f, 0x40, 0x00, 0x00, 0x00, 0x00, 0x00},
		result.Val.data,
	)

	// Decode missing value
	err = client.QuerySingle(ctx, `
		SELECT { val := <OPTIONAL float64>$0 }`,
		&result,
		types.OptionalFloat64{},
	)
	assert.EqualError(t, err, "gel.InvalidArgumentError: "+
		"the \"out\" argument does not match query schema: "+
		"expected gel.CustomFloat64 at "+
		"struct { Val gel.CustomFloat64 \"gel:\\\"val\\\"\" }.val "+
		"to be OptionalUnmarshaler interface "+
		"because the field is not required")
}

func TestSendFloat64Marshaler(t *testing.T) {
	ctx := context.Background()
	var result struct {
		Val types.OptionalFloat64 `gel:"val"`
	}

	// encode value into required argument
	err := client.QuerySingle(ctx, `
		SELECT { val := <float64>$0 }`,
		&result,
		CustomFloat64{data: []byte{
			0xc0, 0x2f, 0x40, 0x00, 0x00, 0x00, 0x00, 0x00}},
	)
	assert.NoError(t, err)

	// encode value into optional argument
	err = client.QuerySingle(ctx, `
		SELECT { val := <OPTIONAL float64>$0 }`,
		&result,
		CustomFloat64{data: []byte{
			0xc0, 0x2f, 0x40, 0x00, 0x00, 0x00, 0x00, 0x00}},
	)
	assert.NoError(t, err)
	assert.Equal(t, types.NewOptionalFloat64(-15.625), result.Val)

	// encode wrong number of bytes
	err = client.QuerySingle(ctx, `
		SELECT { val := <float64>$0 }`,
		&result,
		CustomFloat64{data: []byte{0x01}},
	)
	assert.EqualError(t, err, "gel.InvalidArgumentError: "+
		"wrong number of bytes encoded by gel.CustomFloat64 "+
		"at args[0] expected 8, got 1")
}

type CustomOptionalFloat64 struct {
	CustomFloat64
	isSet bool
}

func (m CustomOptionalFloat64) MarshalEdgeDBFloat64() ([]byte, error) {
	if !m.isSet {
		return nil, fmt.Errorf("%T is not set", m)
	}
	return m.CustomFloat64.MarshalEdgeDBFloat64()
}

func (m *CustomOptionalFloat64) UnmarshalEdgeDBFloat64(data []byte) error {
	m.isSet = true
	return m.CustomFloat64.UnmarshalEdgeDBFloat64(data)
}

func (m *CustomOptionalFloat64) SetMissing(missing bool) {
	m.isSet = !missing
	m.data = nil
}

func (m CustomOptionalFloat64) Missing() bool { return !m.isSet }

func TestReceiveOptionalFloat64Unmarshaler(t *testing.T) {
	ddl := `CREATE TYPE Sample { CREATE PROPERTY val -> float64; };`
	inRolledBackTx(t, ddl, func(ctx context.Context, tx *Tx) {
		var result struct {
			Val CustomOptionalFloat64 `gel:"val"`
		}

		// Decode value
		err := tx.QuerySingle(ctx, `
		SELECT { val := <float64>-15.625 }`,
			&result,
		)
		assert.NoError(t, err)
		assert.Equal(t,
			[]byte{0xc0, 0x2f, 0x40, 0x00, 0x00, 0x00, 0x00, 0x00},
			result.Val.data,
		)

		// Decode missing value
		query := `WITH inserted := (INSERT Sample) SELECT inserted { val }`
		err = tx.QuerySingle(ctx, query, &result)
		assert.NoError(t, err)
		assert.Equal(t, CustomOptionalFloat64{}, result.Val)
	})
}

func TestSendOptionalFloat64Marshaler(t *testing.T) {
	ctx := context.Background()
	var result struct {
		Val types.OptionalFloat64 `gel:"val"`
	}

	newValue := func(data []byte) CustomOptionalFloat64 {
		return CustomOptionalFloat64{
			isSet:         true,
			CustomFloat64: CustomFloat64{data: data},
		}
	}

	// encode value into required argument
	err := client.QuerySingle(ctx, `
		SELECT { val := <float64>$0 }`,
		&result,
		// -15.625,
		newValue([]byte{0xc0, 0x2f, 0x40, 0x00, 0x00, 0x00, 0x00, 0x00}),
	)
	assert.NoError(t, err)
	assert.Equal(t, types.NewOptionalFloat64(-15.625), result.Val)

	// encode value into optional argument
	err = client.QuerySingle(ctx, `
		SELECT { val := <OPTIONAL float64>$0 }`,
		&result,
		newValue([]byte{0xc0, 0x2f, 0x40, 0x00, 0x00, 0x00, 0x00, 0x00}),
	)
	assert.NoError(t, err)
	assert.Equal(t, types.NewOptionalFloat64(-15.625), result.Val)

	// encode missing value into optional argument
	err = client.QuerySingle(ctx, `
		SELECT { val := <OPTIONAL float64>$0 }`,
		&result,
		CustomOptionalFloat64{},
	)
	assert.NoError(t, err)
	assert.Equal(t, types.OptionalFloat64{}, result.Val)

	// encode missing value into required argument
	err = client.QuerySingle(ctx, `
		SELECT { val := <float64>$0 }`,
		&result,
		CustomOptionalFloat64{},
	)
	assert.EqualError(t, err, "gel.InvalidArgumentError: "+
		"cannot encode gel.CustomOptionalFloat64 at args[0] "+
		"because its value is missing")

	// encode wrong number of bytes with required argument
	err = client.QuerySingle(ctx, `
		SELECT { val := <float64>$0 }`,
		&result,
		newValue([]byte{0x01}),
	)
	assert.EqualError(t, err, "gel.InvalidArgumentError: "+
		"wrong number of bytes encoded by gel.CustomOptionalFloat64 "+
		"at args[0] expected 8, got 1")

	// encode wrong number of bytes with optional argument
	err = client.QuerySingle(ctx, `
		SELECT { val := <OPTIONAL float64>$0 }`,
		&result,
		newValue([]byte{0x01}),
	)
	assert.EqualError(t, err, "gel.InvalidArgumentError: "+
		"wrong number of bytes encoded by gel.CustomOptionalFloat64 "+
		"at args[0] expected 8, got 1")
}

func TestSendAndReceiveOptionalFloat64(t *testing.T) {
	ctx := context.Background()

	err := client.Tx(ctx, func(ctx context.Context, tx *Tx) error {
		e := tx.Execute(ctx, `
			CREATE TYPE Float64FieldHolder {
				CREATE PROPERTY float64 -> float64;
			};

			INSERT Float64FieldHolder;
		`)
		if e != nil {
			return e
		}

		type Result struct {
			Float64 types.OptionalFloat64 `gel:"float64"`
		}

		var result Result
		e = tx.QuerySingle(ctx, `
			# decode missing optional
			SELECT Float64FieldHolder { float64 } LIMIT 1`,
			&result,
		)
		if e != nil {
			return e
		}
		assert.Equal(t, Result{}, result)

		e = tx.QuerySingle(ctx, `
			# encode unset optional
			SELECT Float64FieldHolder {
				float64 := <OPTIONAL float64>$0
			} LIMIT 1`,
			&result,
			types.OptionalFloat64{},
		)
		if e != nil {
			return e
		}
		assert.Equal(t, Result{}, result)

		e = tx.QuerySingle(ctx, `
			# encode set optional
			SELECT Float64FieldHolder {
				float64 := <OPTIONAL float64>$0
			} LIMIT 1`,
			&result,
			types.NewOptionalFloat64(6.4),
		)
		if e != nil {
			return e
		}
		assert.Equal(t, Result{Float64: types.NewOptionalFloat64(6.4)}, result)

		e = tx.QuerySingle(ctx, `
			# encode set optional into required argument
			SELECT Float64FieldHolder { float64 := <float64>$0 } LIMIT 1`,
			&result,
			types.NewOptionalFloat64(6.4),
		)
		if e != nil {
			return e
		}
		assert.Equal(t, Result{Float64: types.NewOptionalFloat64(6.4)}, result)

		e = tx.QuerySingle(ctx, `
			# encode unset optional into required argument
			SELECT Float64FieldHolder { float64 := <float64>$0 } LIMIT 1`,
			&result,
			types.OptionalFloat64{},
		)
		assert.EqualError(t, e, "gel.InvalidArgumentError: "+
			"cannot encode gel.OptionalFloat64 at args[0] "+
			"because its value is missing")

		return errors.New("rollback")
	})

	assert.EqualError(t, err, "rollback")
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
		Encoded   string  `gel:"encoded"`
		Decoded   float32 `gel:"decoded"`
		RoundTrip float32 `gel:"round_trip"`
		IsEqual   bool    `gel:"is_equal"`
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
	err := client.Query(ctx, query, &results, numbers, strings)
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
	data []byte
}

func (m CustomFloat32) MarshalEdgeDBFloat32() ([]byte, error) {
	data := make([]byte, len(m.data))
	copy(data, m.data)
	return data, nil
}

func (m *CustomFloat32) UnmarshalEdgeDBFloat32(data []byte) error {
	m.data = make([]byte, len(data))
	copy(m.data, data)
	return nil
}

func TestReceiveFloat32Unmarshaler(t *testing.T) {
	ctx := context.Background()
	var result struct {
		Val CustomFloat32 `gel:"val"`
	}

	// Decode value
	query := `SELECT { val := <float32>-15.625 }`
	err := client.QuerySingle(ctx, query, &result)
	assert.NoError(t, err)
	assert.Equal(t, []byte{0xc1, 0x7a, 0x00, 0x00}, result.Val.data)

	// Decode missing value
	err = client.QuerySingle(ctx, `
		SELECT { val := <OPTIONAL float32>$0 }`,
		&result,
		types.OptionalFloat32{},
	)
	assert.EqualError(t, err, "gel.InvalidArgumentError: "+
		"the \"out\" argument does not match query schema: "+
		"expected gel.CustomFloat32 at "+
		"struct { Val gel.CustomFloat32 \"gel:\\\"val\\\"\" }.val "+
		"to be OptionalUnmarshaler interface "+
		"because the field is not required")
}

func TestSendFloat32Marshaler(t *testing.T) {
	ctx := context.Background()
	var result struct {
		Val types.OptionalFloat32 `gel:"val"`
	}

	// encode value into required argument
	err := client.QuerySingle(ctx, `
		SELECT { val := <float32>$0 }`,
		&result,
		CustomFloat32{data: []byte{0xc1, 0x7a, 0x00, 0x00}},
	)
	assert.NoError(t, err)
	assert.Equal(t, types.NewOptionalFloat32(-15.625), result.Val)

	// encode value into optional argument
	err = client.QuerySingle(ctx, `
		SELECT { val := <OPTIONAL float32>$0 }`,
		&result,
		CustomFloat32{data: []byte{0xc1, 0x7a, 0x00, 0x00}},
	)
	assert.NoError(t, err)
	assert.Equal(t, types.NewOptionalFloat32(-15.625), result.Val)

	// encode wrong number of bytes
	err = client.QuerySingle(ctx, `
		SELECT { val := <float32>$0 }`,
		&result,
		CustomFloat32{data: []byte{0x01}},
	)
	assert.EqualError(t, err, "gel.InvalidArgumentError: "+
		"wrong number of bytes encoded by gel.CustomFloat32 "+
		"at args[0] expected 4, got 1")
}

type CustomOptionalFloat32 struct {
	CustomFloat32
	isSet bool
}

func (m CustomOptionalFloat32) MarshalEdgeDBFloat32() ([]byte, error) {
	if !m.isSet {
		return nil, fmt.Errorf("%T is not set", m)
	}
	return m.CustomFloat32.MarshalEdgeDBFloat32()
}

func (m *CustomOptionalFloat32) UnmarshalEdgeDBFloat32(data []byte) error {
	m.isSet = true
	return m.CustomFloat32.UnmarshalEdgeDBFloat32(data)
}

func (m *CustomOptionalFloat32) SetMissing(missing bool) {
	m.isSet = !missing
	m.data = nil
}

func (m CustomOptionalFloat32) Missing() bool { return !m.isSet }

func TestReceiveOptionalFloat32Unmarshaler(t *testing.T) {
	ddl := `CREATE TYPE Sample { CREATE PROPERTY val -> float32; };`
	inRolledBackTx(t, ddl, func(ctx context.Context, tx *Tx) {
		var result struct {
			Val CustomOptionalFloat32 `gel:"val"`
		}

		// Decode value
		err := tx.QuerySingle(ctx, `
			SELECT { val := <float32>-15.625 }`,
			&result,
		)
		assert.NoError(t, err)
		assert.Equal(t, []byte{0xc1, 0x7a, 0x00, 0x00}, result.Val.data)

		// Decode missing value
		query := `WITH inserted := (INSERT Sample) SELECT inserted { val }`
		err = tx.QuerySingle(ctx, query, &result)
		assert.NoError(t, err)
		assert.Equal(t, CustomOptionalFloat32{}, result.Val)
	})
}

func TestSendOptionalFloat32Marshaler(t *testing.T) {
	ctx := context.Background()
	var result struct {
		Val types.OptionalFloat32 `gel:"val"`
	}

	newValue := func(data []byte) CustomOptionalFloat32 {
		return CustomOptionalFloat32{
			isSet:         true,
			CustomFloat32: CustomFloat32{data: data},
		}
	}

	// encode value into required argument
	err := client.QuerySingle(ctx, `
		SELECT { val := <float32>$0 }`,
		&result,
		newValue([]byte{0xc1, 0x7a, 0x00, 0x00}),
	)
	assert.NoError(t, err)
	assert.Equal(t, types.NewOptionalFloat32(-15.625), result.Val)

	// encode value into optional argument
	err = client.QuerySingle(ctx, `
		SELECT { val := <OPTIONAL float32>$0 }`,
		&result,
		newValue([]byte{0xc1, 0x7a, 0x00, 0x00}),
	)
	assert.NoError(t, err)
	assert.Equal(t, types.NewOptionalFloat32(-15.625), result.Val)

	// encode missing value into optional argument
	err = client.QuerySingle(ctx, `
		SELECT { val := <OPTIONAL float32>$0 }`,
		&result,
		CustomOptionalFloat32{},
	)
	assert.NoError(t, err)
	assert.Equal(t, types.OptionalFloat32{}, result.Val)

	// encode missing value into required argument
	err = client.QuerySingle(ctx, `
		SELECT { val := <float32>$0 }`,
		&result,
		CustomOptionalFloat32{},
	)
	assert.EqualError(t, err, "gel.InvalidArgumentError: "+
		"cannot encode gel.CustomOptionalFloat32 at args[0] "+
		"because its value is missing")

	// encode wrong number of bytes with required argument
	err = client.QuerySingle(ctx, `
		SELECT { val := <float32>$0 }`,
		&result,
		newValue([]byte{0x01}),
	)
	assert.EqualError(t, err, "gel.InvalidArgumentError: "+
		"wrong number of bytes encoded by gel.CustomOptionalFloat32 "+
		"at args[0] expected 4, got 1")

	// encode wrong number of bytes with optional argument
	err = client.QuerySingle(ctx, `
		SELECT { val := <OPTIONAL float32>$0 }`,
		&result,
		newValue([]byte{0x01}),
	)
	assert.EqualError(t, err, "gel.InvalidArgumentError: "+
		"wrong number of bytes encoded by gel.CustomOptionalFloat32 "+
		"at args[0] expected 4, got 1")
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
	err := client.Query(ctx, query, &results, samples)
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
	data := make([]byte, len(m.data))
	copy(data, m.data)
	return data, nil
}

func (m *CustomBytes) UnmarshalEdgeDBBytes(data []byte) error {
	m.data = make([]byte, len(data))
	copy(m.data, data)
	return nil
}

func TestReceiveBytesUnmarshaler(t *testing.T) {
	ctx := context.Background()
	var result struct {
		Val CustomBytes `gel:"val"`
	}

	// Decode value
	query := `SELECT { val := b'\x01\x02\x03' }`
	err := client.QuerySingle(ctx, query, &result)
	assert.NoError(t, err)
	assert.Equal(t, []byte{0x01, 0x02, 0x03}, result.Val.data)

	// Decode missing value
	err = client.QuerySingle(ctx, `
		SELECT { val := <OPTIONAL bytes>$0 }`,
		&result,
		types.OptionalBytes{},
	)
	assert.EqualError(t, err, "gel.InvalidArgumentError: "+
		"the \"out\" argument does not match query schema: "+
		"expected gel.CustomBytes at "+
		"struct { Val gel.CustomBytes \"gel:\\\"val\\\"\" }.val "+
		"to be OptionalUnmarshaler interface "+
		"because the field is not required")
}

func TestSendBytesMarshaler(t *testing.T) {
	ctx := context.Background()
	var result struct {
		Val types.OptionalBytes `gel:"val"`
	}

	// encode value into required argument
	err := client.QuerySingle(ctx, `
		SELECT { val := <bytes>$0 }`,
		&result,
		CustomBytes{data: []byte{0x01, 0x02, 0x03}},
	)
	assert.NoError(t, err)
	assert.Equal(
		t,
		types.NewOptionalBytes([]byte{0x01, 0x02, 0x03}),
		result.Val,
	)

	// encode value into optional argument
	err = client.QuerySingle(ctx, `
		SELECT { val := <OPTIONAL bytes>$0 }`,
		&result,
		CustomBytes{data: []byte{0x01, 0x02, 0x03}},
	)
	assert.NoError(t, err)
	assert.Equal(
		t,
		types.NewOptionalBytes([]byte{0x01, 0x02, 0x03}),
		result.Val,
	)

	// encode missing value into optional argument
	err = client.QuerySingle(ctx, `
		SELECT { val := <OPTIONAL bytes>$0 }`,
		&result,
		CustomOptionalBytes{},
	)
	assert.NoError(t, err)
	assert.Equal(t, types.OptionalBytes{}, result.Val)

	// encode missing value into optional argument
	err = client.QuerySingle(ctx, `
		SELECT { val := <bytes>$0 }`,
		&result,
		CustomOptionalBytes{},
	)
	assert.EqualError(t, err, "gel.InvalidArgumentError: "+
		"cannot encode gel.CustomOptionalBytes at args[0] "+
		"because its value is missing")
}

type CustomOptionalBytes struct {
	CustomBytes
	isSet bool
}

func (m CustomOptionalBytes) MarshalEdgeDBBytes() ([]byte, error) {
	if !m.isSet {
		return nil, fmt.Errorf("%T is not set", m)
	}
	return m.CustomBytes.MarshalEdgeDBBytes()
}

func (m *CustomOptionalBytes) UnmarshalEdgeDBBytes(data []byte) error {
	m.isSet = true
	return m.CustomBytes.UnmarshalEdgeDBBytes(data)
}

func (m *CustomOptionalBytes) SetMissing(missing bool) {
	m.isSet = !missing
	m.data = nil
}

func (m CustomOptionalBytes) Missing() bool { return !m.isSet }

func TestReceiveOptionalBytesUnmarshaler(t *testing.T) {
	ddl := `CREATE TYPE Sample { CREATE PROPERTY val -> bytes; };`
	inRolledBackTx(t, ddl, func(ctx context.Context, tx *Tx) {
		var result struct {
			Val CustomOptionalBytes `gel:"val"`
		}

		// Decode value
		err := tx.QuerySingle(ctx, `
			SELECT { val := b'\x01\x02\x03' }`,
			&result,
		)
		assert.NoError(t, err)
		assert.Equal(t, []byte{0x01, 0x02, 0x03}, result.Val.data)

		// Decode missing value
		query := `WITH inserted := (INSERT Sample) SELECT inserted { val }`
		err = tx.QuerySingle(ctx, query, &result)
		assert.NoError(t, err)
		assert.Equal(t, CustomOptionalBytes{}, result.Val)
	})
}

func TestSendOptionalBytesMarshaler(t *testing.T) {
	ctx := context.Background()
	var result struct {
		Val types.OptionalBytes `gel:"val"`
	}

	newValue := func(data []byte) CustomOptionalBytes {
		return CustomOptionalBytes{
			isSet:       true,
			CustomBytes: CustomBytes{data: data},
		}
	}

	// encode value into required argument
	err := client.QuerySingle(ctx, `
		SELECT { val := <bytes>$0 }`,
		&result,
		newValue([]byte{0x01, 0x02, 0x03}),
	)
	assert.NoError(t, err)
	assert.Equal(
		t,
		types.NewOptionalBytes([]byte{0x01, 0x02, 0x03}),
		result.Val,
	)

	// encode value into optional argument
	err = client.QuerySingle(ctx, `
		SELECT { val := <OPTIONAL bytes>$0 }`,
		&result,
		newValue([]byte{0x01, 0x02, 0x03}),
	)
	assert.NoError(t, err)
	assert.Equal(
		t,
		types.NewOptionalBytes([]byte{0x01, 0x02, 0x03}),
		result.Val,
	)

	// encode missing value into optional argument
	err = client.QuerySingle(ctx, `
		SELECT { val := <OPTIONAL bytes>$0 }`,
		&result,
		CustomOptionalBytes{},
	)
	assert.NoError(t, err)
	assert.Equal(t, types.OptionalBytes{}, result.Val)

	// encode missing value into required argument
	err = client.QuerySingle(ctx, `
		SELECT { val := <bytes>$0 }`,
		&result,
		CustomOptionalBytes{},
	)
	assert.EqualError(t, err, "gel.InvalidArgumentError: "+
		"cannot encode gel.CustomOptionalBytes at args[0] "+
		"because its value is missing")
}

func TestSendAndReceiveStr(t *testing.T) {
	ctx := context.Background()

	var result string
	err := client.QuerySingle(ctx, `SELECT <str>$0`, &result, "abcdef")
	require.NoError(t, err)
	assert.Equal(t, "abcdef", result, "round trip failed")
}

func TestFetchLargeStr(t *testing.T) {
	// This test is meant to stress the buffer implementation.
	ctx := context.Background()

	var result string
	err := client.QuerySingle(ctx,
		"SELECT str_repeat('a', <int64>(10^6))", &result)
	require.NoError(t, err)
	assert.Equal(t, strings.Repeat("a", 1_000_000), result)
}

type CustomStr struct {
	data []byte
}

func (m CustomStr) MarshalEdgeDBStr() ([]byte, error) {
	data := make([]byte, len(m.data))
	copy(data, m.data)
	return data, nil
}

func (m *CustomStr) UnmarshalEdgeDBStr(data []byte) error {
	m.data = make([]byte, len(data))
	copy(m.data, data)
	return nil
}

func TestReceiveStrUnmarshaler(t *testing.T) {
	ctx := context.Background()
	var result struct {
		Val CustomStr `gel:"val"`
	}

	// Decode value
	err := client.QuerySingle(ctx, `SELECT { val := 'Hi 🙂' }`, &result)
	assert.NoError(t, err)
	assert.Equal(t,
		[]byte{0x48, 0x69, 0x20, 0xf0, 0x9f, 0x99, 0x82},
		result.Val.data,
	)

	// Decode missing value
	err = client.QuerySingle(ctx, `
		SELECT { val := <OPTIONAL str>$0 }`,
		&result,
		types.OptionalStr{},
	)
	assert.EqualError(t, err, "gel.InvalidArgumentError: "+
		"the \"out\" argument does not match query schema: "+
		"expected gel.CustomStr at "+
		"struct { Val gel.CustomStr \"gel:\\\"val\\\"\" }.val "+
		"to be OptionalUnmarshaler interface "+
		"because the field is not required")
}

func TestSendStrMarshaler(t *testing.T) {
	ctx := context.Background()
	var result struct {
		Val types.OptionalStr `gel:"val"`
	}

	// encode value into required argument
	err := client.QuerySingle(ctx, `
		SELECT { val := <str>$0 }`,
		&result,
		CustomStr{
			data: []byte{0x48, 0x69, 0x20, 0xf0, 0x9f, 0x99, 0x82},
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, types.NewOptionalStr("Hi 🙂"), result.Val)

	// encode value into optional argument
	err = client.QuerySingle(ctx, `
		SELECT { val := <OPTIONAL str>$0 }`,
		&result,
		CustomStr{
			data: []byte{0x48, 0x69, 0x20, 0xf0, 0x9f, 0x99, 0x82},
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, types.NewOptionalStr("Hi 🙂"), result.Val)
}

type CustomOptionalStr struct {
	data  []byte
	isSet bool
}

func (m CustomOptionalStr) MarshalEdgeDBStr() ([]byte, error) {
	if !m.isSet {
		return nil, fmt.Errorf("%T is not set", m)
	}
	data := make([]byte, len(m.data))
	copy(data, m.data)
	return data, nil
}

func (m *CustomOptionalStr) UnmarshalEdgeDBStr(data []byte) error {
	m.isSet = true
	m.data = make([]byte, len(data))
	copy(m.data, data)
	return nil
}

func (m *CustomOptionalStr) SetMissing(missing bool) {
	m.isSet = !missing
	m.data = nil
}

func (m CustomOptionalStr) Missing() bool { return !m.isSet }

func TestReceiveOptionalStrUnmarshaler(t *testing.T) {
	ddl := `CREATE TYPE Sample { CREATE PROPERTY val -> str; };`
	inRolledBackTx(t, ddl, func(ctx context.Context, tx *Tx) {
		var result struct {
			Val CustomOptionalStr `gel:"val"`
		}

		// Decode value
		err := tx.QuerySingle(ctx, `SELECT { val := "Hi 🙂" }`, &result)
		assert.NoError(t, err)
		assert.Equal(t, []byte("Hi 🙂"), result.Val.data)

		// Decode missing value
		query := `WITH inserted := (INSERT Sample) SELECT inserted { val }`
		err = tx.QuerySingle(ctx, query, &result)
		assert.NoError(t, err)
		assert.Equal(t, CustomOptionalStr{}, result.Val)
	})
}

func TestSendOptionalStrMarshaler(t *testing.T) {
	ctx := context.Background()
	var result struct {
		Val types.OptionalStr `gel:"val"`
	}

	newValue := func(data []byte) CustomOptionalStr {
		return CustomOptionalStr{isSet: true, data: data}
	}

	// encode value into required argument
	err := client.QuerySingle(ctx, `
		SELECT { val := <str>$0 }`,
		&result,
		newValue([]byte("Hi 🙂")),
	)
	assert.NoError(t, err)
	assert.Equal(t, types.NewOptionalStr("Hi 🙂"), result.Val)

	// encode value into optional argument
	err = client.QuerySingle(ctx, `
		SELECT { val := <OPTIONAL str>$0 }`,
		&result,
		newValue([]byte("Hi 🙂")),
	)
	assert.NoError(t, err)
	assert.Equal(t, types.NewOptionalStr("Hi 🙂"), result.Val)

	// encode missing value into optional argument
	err = client.QuerySingle(ctx, `
		SELECT { val := <OPTIONAL str>$0 }`,
		&result,
		CustomOptionalStr{},
	)
	assert.NoError(t, err)
	assert.Equal(t, types.OptionalStr{}, result.Val)

	// encode missing value into required argument
	err = client.QuerySingle(ctx, `
		SELECT { val := <str>$0 }`,
		&result,
		CustomOptionalStr{},
	)
	assert.EqualError(t, err, "gel.InvalidArgumentError: "+
		"cannot encode gel.CustomOptionalStr at args[0] "+
		"because its value is missing")
}

func TestSendAndReceiveOptionalStr(t *testing.T) {
	ctx := context.Background()

	err := client.Tx(ctx, func(ctx context.Context, tx *Tx) error {
		e := tx.Execute(ctx, `
			CREATE TYPE StrFieldHolder {
				CREATE PROPERTY str -> str;
			};

			INSERT StrFieldHolder;
		`)
		if e != nil {
			return e
		}

		type Result struct {
			Str types.OptionalStr `gel:"str"`
		}

		var result Result
		e = tx.QuerySingle(ctx, `
			# decode missing optional
			SELECT StrFieldHolder { str } LIMIT 1`,
			&result,
		)
		if e != nil {
			return e
		}
		assert.Equal(t, Result{}, result)

		e = tx.QuerySingle(ctx, `
			# encode unset optional
			SELECT StrFieldHolder { str := <OPTIONAL str>$0 } LIMIT 1`,
			&result,
			types.OptionalStr{},
		)
		if e != nil {
			return e
		}
		assert.Equal(t, Result{}, result)

		e = tx.QuerySingle(ctx, `
			# encode set optional
			SELECT StrFieldHolder { str := <OPTIONAL str>$0 } LIMIT 1`,
			&result,
			types.NewOptionalStr("hello"),
		)
		if e != nil {
			return e
		}
		assert.Equal(t, Result{Str: types.NewOptionalStr("hello")}, result)

		e = tx.QuerySingle(ctx, `
			# encode set optional into required argument
			SELECT StrFieldHolder { str := <str>$0 } LIMIT 1`,
			&result,
			types.NewOptionalStr("hello"),
		)
		if e != nil {
			return e
		}
		assert.Equal(t, Result{Str: types.NewOptionalStr("hello")}, result)

		e = tx.QuerySingle(ctx, `
			# encode unset optional into required argument
			SELECT StrFieldHolder { str := <str>$0 } LIMIT 1`,
			&result,
			types.OptionalStr{},
		)
		assert.EqualError(t, e, "gel.InvalidArgumentError: "+
			"cannot encode gel.OptionalStr at args[0] "+
			"because its value is missing")

		return errors.New("rollback")
	})

	assert.EqualError(t, err, "rollback")
}

func TestSendAndReceiveJSONBytes(t *testing.T) {
	ctx := context.Background()

	strings := []string{"123", "-3.14", "true", "false", "[1, 2, 3]", "null"}

	samples := make([][]byte, len(strings))
	for i, s := range strings {
		samples[i] = []byte(s)
	}

	query := `SELECT array_unpack(<array<json>>$0)`

	var results [][]byte
	err := client.Query(ctx, query, &results, samples)
	require.NoError(t, err)
	require.Equal(t, len(samples), len(results), "wrong number of results")

	for i, s := range strings {
		t.Run(s, func(t *testing.T) {
			assert.Equal(t, samples[i], results[i])
		})
	}
}

type JSONObject struct {
	A float64 `json:"a"`
	B string  `json:"b"`
}

type OptionalJSONObject struct {
	types.Optional
	JSONObject `gel:"$inline"`
}

func TestReceiveJSONAndUnmarshal(t *testing.T) {
	// nolint:lll
	type Result struct {
		Interface                interface{}        `gel:"interface"`
		MissingInterface         interface{}        `gel:"missing_interface"`
		Scalar                   string             `gel:"scalar"`
		Slice                    []string           `gel:"slice"`
		MissingSlice             []string           `gel:"missing_slice"`
		Object                   JSONObject         `gel:"object"`
		MissingObject            OptionalJSONObject `gel:"missing_object"`
		NotMissingOptionalObject OptionalJSONObject `gel:"not_missing_optional_object"`
		MissingScalar            types.OptionalStr  `gel:"missing_scalar"`
		NotMissingOptionalScalar types.OptionalStr  `gel:"not_missing_optional_scalar"`
	}

	result := Result{
		MissingInterface: "the client should change this to nil",
		MissingSlice:     []string{"the client should change this to nil"},
	}
	result.MissingObject.SetMissing(false)
	result.MissingScalar.Set("the client should change this to missing")

	err := client.QuerySingle(
		context.Background(),
		`SELECT {
			interface := <json>123,
			missing_interface := <json>{},
			scalar := <json>"text",
			slice := to_json('["a", "b"]'),
			missing_slice := <json>{},
			object := to_json('{"a": 1, "b": "two"}'),
			missing_object := <json>{},
			not_missing_optional_object := to_json('{"a": 1, "b": "two"}'),
			missing_scalar := <json>{},
			not_missing_optional_scalar := <json>"text",
		}`,
		&result,
	)
	require.NoError(t, err)

	notMissing := OptionalJSONObject{
		JSONObject: JSONObject{
			A: float64(1),
			B: "two",
		},
	}
	notMissing.SetMissing(false)

	require.Equal(
		t,
		Result{
			Interface:                float64(123),
			MissingInterface:         nil,
			Scalar:                   "text",
			Slice:                    []string{"a", "b"},
			MissingSlice:             nil,
			Object:                   JSONObject{A: float64(1), B: "two"},
			NotMissingOptionalObject: notMissing,
			NotMissingOptionalScalar: types.NewOptionalStr("text"),
		},
		result,
	)
}

func TestReceiveJSONWrongType(t *testing.T) {
	var result string
	err := client.QuerySingle(
		context.Background(),
		`SELECT <json>123`,
		&result,
	)
	require.EqualError(
		t,
		err,
		"json: cannot unmarshal number into Go value of type string",
	)
	require.Equal(t, "", result)
}

type CustomJSON struct {
	data []byte
}

func (m CustomJSON) MarshalEdgeDBJSON() ([]byte, error) {
	data := make([]byte, len(m.data))
	copy(data, m.data)
	return data, nil
}

func (m *CustomJSON) UnmarshalEdgeDBJSON(data []byte) error {
	m.data = make([]byte, len(data))
	copy(m.data, data)
	return nil
}

func TestReceiveJSONUnmarshaler(t *testing.T) {
	ctx := context.Background()
	var result struct {
		Val CustomJSON `gel:"val"`
	}

	// Decode value
	err := client.QuerySingle(ctx, `
		SELECT { val := <json>(hello := "world") }`,
		&result,
	)
	assert.NoError(t, err)
	assert.Equal(t,
		append([]byte{0x01}, []byte(`{"hello": "world"}`)...),
		result.Val.data,
	)

	// Decode missing value
	err = client.QuerySingle(ctx, `
		SELECT { val := <OPTIONAL json>$0 }`,
		&result,
		types.OptionalBytes{},
	)
	assert.EqualError(t, err, "gel.InvalidArgumentError: "+
		"the \"out\" argument does not match query schema: "+
		"expected gel.CustomJSON at "+
		"struct { Val gel.CustomJSON \"gel:\\\"val\\\"\" }.val "+
		"to be OptionalUnmarshaler interface "+
		"because the field is not required")
}

func TestSendJSONMarshaler(t *testing.T) {
	ctx := context.Background()
	var result struct {
		Val types.OptionalBytes `gel:"val"`
	}

	// encode value into required argument
	err := client.QuerySingle(ctx, `
		SELECT { val := <json>$0 }`,
		&result,
		CustomJSON{data: append([]byte{1}, []byte(`{"hello": "world"}`)...)},
	)
	assert.NoError(t, err)
	assert.Equal(
		t,
		types.NewOptionalBytes([]byte(`{"hello": "world"}`)),
		result.Val,
	)

	// encode value into optional argument
	err = client.QuerySingle(ctx, `
		SELECT { val := <OPTIONAL json>$0 }`,
		&result,
		CustomJSON{data: append([]byte{1}, []byte(`{"hello": "world"}`)...)},
	)
	assert.NoError(t, err)
	assert.Equal(
		t,
		types.NewOptionalBytes([]byte(`{"hello": "world"}`)),
		result.Val,
	)
}

type CustomOptionalJSON struct {
	data  []byte
	isSet bool
}

func (m CustomOptionalJSON) MarshalEdgeDBJSON() ([]byte, error) {
	if !m.isSet {
		return nil, fmt.Errorf("%T is not set", m)
	}
	data := make([]byte, len(m.data))
	copy(data, m.data)
	return data, nil
}

func (m *CustomOptionalJSON) UnmarshalEdgeDBJSON(data []byte) error {
	m.isSet = true
	m.data = make([]byte, len(data))
	copy(m.data, data)
	return nil
}

func (m *CustomOptionalJSON) SetMissing(missing bool) {
	m.isSet = !missing
	m.data = nil
}

func (m CustomOptionalJSON) Missing() bool { return !m.isSet }

func TestReceiveOptionalJSONUnmarshaler(t *testing.T) {
	ddl := `CREATE TYPE Sample { CREATE PROPERTY val -> json; };`
	inRolledBackTx(t, ddl, func(ctx context.Context, tx *Tx) {
		var result struct {
			Val CustomOptionalJSON `gel:"val"`
		}

		// Decode value
		err := tx.QuerySingle(ctx, `
			SELECT { val := <json>(hello := "world") }`,
			&result,
		)
		assert.NoError(t, err)
		assert.Equal(t,
			append([]byte{0x01}, []byte(`{"hello": "world"}`)...),
			result.Val.data,
		)

		// Decode missing value
		query := `WITH inserted := (INSERT Sample) SELECT inserted { val }`
		err = tx.QuerySingle(ctx, query, &result)
		assert.NoError(t, err)
		assert.Equal(t, CustomOptionalJSON{}, result.Val)
	})
}

func TestSendOptionalJSONMarshaler(t *testing.T) {
	ctx := context.Background()
	var result struct {
		Val types.OptionalBytes `gel:"val"`
	}

	newValue := func(data []byte) CustomOptionalJSON {
		return CustomOptionalJSON{isSet: true, data: data}
	}

	// encode value into required argument
	err := client.QuerySingle(ctx, `
		SELECT { val := <json>$0 }`,
		&result,
		newValue(append([]byte{1}, []byte(`{"hello": "world"}`)...)),
	)
	assert.NoError(t, err)
	assert.Equal(
		t,
		types.NewOptionalBytes([]byte(`{"hello": "world"}`)),
		result.Val,
	)

	// encode value into optional argument
	err = client.QuerySingle(ctx, `
		SELECT { val := <OPTIONAL json>$0 }`,
		&result,
		newValue(append([]byte{1}, []byte(`{"hello": "world"}`)...)),
	)
	assert.NoError(t, err)
	assert.Equal(
		t,
		types.NewOptionalBytes([]byte(`{"hello": "world"}`)),
		result.Val,
	)

	// encode missing value into optional argument
	err = client.QuerySingle(ctx, `
		SELECT { val := <OPTIONAL json>$0 }`,
		&result,
		CustomOptionalJSON{},
	)
	assert.NoError(t, err)
	assert.Equal(t, types.OptionalBytes{}, result.Val)

	// encode missing value into required argument
	err = client.QuerySingle(ctx, `
		SELECT { val := <json>$0 }`,
		&result,
		CustomOptionalJSON{},
	)
	assert.EqualError(t, err, "gel.InvalidArgumentError: "+
		"cannot encode gel.CustomOptionalJSON at args[0] "+
		"because its value is missing")
}

func TestSendAndReceiveEnum(t *testing.T) {
	ctx := context.Background()

	type Result struct {
		Encoded   string `gel:"encoded"`
		Decoded   string `gel:"decoded"`
		RoundTrip string `gel:"round_trip"`
		IsEqual   bool   `gel:"is_equal"`
		String    string `gel:"string"`
	}

	query := `
		WITH
			e := <Color>$0,
			s := <str>$1
		SELECT (
			encoded := <str>e,
			decoded := <Color>s,
			round_trip := e,
			is_equal := <Color>s = e,
			string := <str><Color>s
		)
	`

	err := client.Tx(ctx, func(ctx context.Context, tx *Tx) error {
		e := tx.Execute(ctx,
			"CREATE SCALAR TYPE Color EXTENDING enum<Red, Green, Blue>;")
		assert.NoError(t, e)

		var result Result
		color := "Red"
		e = tx.QuerySingle(ctx, query, &result, color, color)
		require.NoError(t, e)

		assert.Equal(t, color, result.Encoded, "encoding failed")
		assert.Equal(t, color, result.Decoded, "decoding failed")
		assert.Equal(t, color, result.RoundTrip, "round trip failed")
		assert.True(t, result.IsEqual, "equality failed")
		assert.Equal(t, color, result.String)

		query = "SELECT (decoded := <Color><str>$0)"
		e = tx.QuerySingle(ctx, query, &result, "invalid")

		expected := "gel.InvalidValueError: " +
			"invalid input value for enum 'default::Color': \"invalid\""
		assert.EqualError(t, e, expected)

		return errors.New("rollback")
	})
	assert.EqualError(t, err, "rollback")
}

func TestReceiveEnumUnmarshaler(t *testing.T) {
	ctx := context.Background()
	var result struct {
		Val CustomStr `gel:"val"`
	}

	err := client.Tx(ctx, func(ctx context.Context, tx *Tx) error {
		e := tx.Execute(ctx,
			"CREATE SCALAR TYPE Color EXTENDING enum<Red, Green, Blue>;")
		assert.NoError(t, e)

		// Decode value
		e = tx.QuerySingle(ctx, `SELECT { val := <Color>'Red' }`, &result)
		assert.NoError(t, e)
		assert.Equal(t, []byte("Red"), result.Val.data)

		// Decode missing value
		e = tx.QuerySingle(ctx, `
			SELECT { val := <OPTIONAL Color>$0 }`,
			&result,
			types.OptionalStr{},
		)
		assert.EqualError(t, e, "gel.InvalidArgumentError: "+
			"the \"out\" argument does not match query schema: "+
			"expected gel.CustomStr at "+
			"struct { Val gel.CustomStr \"gel:\\\"val\\\"\" }.val "+
			"to be OptionalUnmarshaler interface "+
			"because the field is not required")

		return errors.New("rollback")
	})
	assert.EqualError(t, err, "rollback")
}

func TestSendEnumMarshaler(t *testing.T) {
	ctx := context.Background()
	var result struct {
		Val types.OptionalStr `gel:"val"`
	}

	err := client.Tx(ctx, func(ctx context.Context, tx *Tx) error {
		e := tx.Execute(ctx,
			"CREATE SCALAR TYPE Color EXTENDING enum<Red, Green, Blue>;")
		assert.NoError(t, e)

		// encode value into required argument
		e = tx.QuerySingle(ctx, `
			SELECT { val := <Color>$0 }`,
			&result,
			CustomStr{data: []byte("Red")},
		)
		assert.NoError(t, e)
		assert.Equal(t, types.NewOptionalStr("Red"), result.Val)

		// encode value into optional argument
		e = tx.QuerySingle(ctx, `
			SELECT { val := <OPTIONAL Color>$0 }`,
			&result,
			CustomStr{data: []byte("Red")},
		)
		assert.NoError(t, e)
		assert.Equal(t, types.NewOptionalStr("Red"), result.Val)

		return errors.New("rollback")
	})
	assert.EqualError(t, err, "rollback")
}

func TestReceiveOptionalEnumUnmarshaler(t *testing.T) {
	ddl := `
		CREATE SCALAR TYPE Color EXTENDING enum<Red, Green, Blue>;
		CREATE TYPE Sample {
			CREATE PROPERTY val -> Color;
		};
	`
	inRolledBackTx(t, ddl, func(ctx context.Context, tx *Tx) {
		var result struct {
			Val CustomOptionalStr `gel:"val"`
		}

		// Decode value
		err := tx.QuerySingle(ctx, `SELECT { val := <Color>'Red' }`, &result)
		assert.NoError(t, err)
		assert.Equal(t, []byte("Red"), result.Val.data)

		// Decode missing value
		query := `WITH inserted := (INSERT Sample) SELECT inserted { val }`
		err = tx.QuerySingle(ctx, query, &result)
		assert.NoError(t, err)
		assert.Equal(t, CustomOptionalStr{}, result.Val)
	})
}

func TestSendOptionalEnumMarshaler(t *testing.T) {
	ctx := context.Background()
	var result struct {
		Val types.OptionalStr `gel:"val"`
	}

	newValue := func(data []byte) CustomOptionalStr {
		return CustomOptionalStr{isSet: true, data: data}
	}

	err := client.Tx(ctx, func(ctx context.Context, tx *Tx) error {
		e := tx.Execute(ctx,
			"CREATE SCALAR TYPE Color EXTENDING enum<Red, Green, Blue>;")
		assert.NoError(t, e)

		// encode value into required argument
		e = tx.QuerySingle(ctx, `
			SELECT { val := <Color>$0 }`,
			&result,
			newValue([]byte("Red")),
		)
		assert.NoError(t, e)
		assert.Equal(t, types.NewOptionalStr("Red"), result.Val)

		// encode value into optional argument
		e = tx.QuerySingle(ctx, `
			SELECT { val := <OPTIONAL Color>$0 }`,
			&result,
			newValue([]byte("Red")),
		)
		assert.NoError(t, e)
		assert.Equal(t, types.NewOptionalStr("Red"), result.Val)

		// encode missing value into optional argument
		e = tx.QuerySingle(ctx, `
			SELECT { val := <OPTIONAL Color>$0 }`,
			&result,
			CustomOptionalStr{},
		)
		assert.NoError(t, e)
		assert.Equal(t, types.OptionalStr{}, result.Val)

		// encode missing value into required argument
		e = tx.QuerySingle(ctx, `
			SELECT { val := <Color>$0 }`,
			&result,
			CustomOptionalStr{},
		)
		assert.EqualError(t, e, "gel.InvalidArgumentError: "+
			"cannot encode gel.CustomOptionalStr at args[0] "+
			"because its value is missing")

		return errors.New("rollback")
	})
	assert.EqualError(t, err, "rollback")
}

func TestSendAndReceiveDuration(t *testing.T) {
	ctx := context.Background()

	durations := []types.Duration{
		types.Duration(0),
		types.Duration(-1),
		types.Duration(86400000000),
		types.Duration(1_000_000),
		types.Duration(3074457345618258432),
	}

	var maxDuration int64 = 3_154_000_000_000_000
	for i := 0; i < 1000; i++ {
		d := types.Duration(rand.Int63n(2*maxDuration) - maxDuration)
		durations = append(durations, d)
	}

	strings := make([]string, len(durations))
	for i := 0; i < len(strings); i++ {
		strings[i] = durations[i].String()
	}

	type Result struct {
		Decoded   types.Duration `gel:"decoded"`
		RoundTrip types.Duration `gel:"round_trip"`
		IsEqual   bool           `gel:"is_equal"`
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
	err := client.Query(ctx, query, &results, durations, strings)
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
	data []byte
}

func (m CustomDuration) MarshalEdgeDBDuration() ([]byte, error) {
	data := make([]byte, len(m.data))
	copy(data, m.data)
	return data, nil
}

func (m *CustomDuration) UnmarshalEdgeDBDuration(data []byte) error {
	m.data = make([]byte, len(data))
	copy(m.data, data)
	return nil
}

func TestReceiveDurationUnmarshaler(t *testing.T) {
	ctx := context.Background()
	var result struct {
		Val CustomDuration `gel:"val"`
	}

	// Decode value
	err := client.QuerySingle(ctx, `
		SELECT { val := <duration>'48 hours 45 minutes 7.6 seconds' }`,
		&result,
	)
	assert.NoError(t, err)
	assert.Equal(t,
		[]byte{
			0x00, 0x00, 0x00, 0x28, 0xdd, 0x11, 0x72, 0x80, // microseconds
			0x00, 0x00, 0x00, 0x00, // days
			0x00, 0x00, 0x00, 0x00, // months
		},
		result.Val.data,
	)

	// Decode missing value
	err = client.QuerySingle(ctx, `
		SELECT { val := <OPTIONAL duration>$0 }`,
		&result,
		types.OptionalDuration{},
	)
	assert.EqualError(t, err, "gel.InvalidArgumentError: "+
		"the \"out\" argument does not match query schema: "+
		"expected gel.CustomDuration at "+
		"struct { Val gel.CustomDuration \"gel:\\\"val\\\"\" }.val "+
		"to be OptionalUnmarshaler interface "+
		"because the field is not required")
}

func TestSendDurationMarshaler(t *testing.T) {
	ctx := context.Background()
	var result struct {
		Val types.OptionalDuration `gel:"val"`
	}

	// encode value into required argument
	err := client.QuerySingle(ctx, `
		SELECT { val := <duration>$0 }`,
		&result,
		CustomDuration{data: []byte{
			0x00, 0x00, 0x00, 0x28, 0xdd, 0x11, 0x72, 0x80, // microseconds
			0x00, 0x00, 0x00, 0x00, // days
			0x00, 0x00, 0x00, 0x00, // months
		}},
	)
	assert.NoError(t, err)
	assert.Equal(t, types.NewOptionalDuration(0x28dd117280), result.Val)

	// encode value into optional argument
	err = client.QuerySingle(ctx, `
		SELECT { val := <OPTIONAL duration>$0 }`,
		&result,
		CustomDuration{data: []byte{
			0x00, 0x00, 0x00, 0x28, 0xdd, 0x11, 0x72, 0x80, // microseconds
			0x00, 0x00, 0x00, 0x00, // days
			0x00, 0x00, 0x00, 0x00, // months
		}},
	)
	assert.NoError(t, err)
	assert.Equal(t, types.NewOptionalDuration(0x28dd117280), result.Val)

	// encode wrong number of bytes
	err = client.QuerySingle(ctx, `
		SELECT { val := <duration>$0 }`,
		&result,
		CustomDuration{data: []byte{0x01}},
	)
	assert.EqualError(t, err, "gel.InvalidArgumentError: "+
		"wrong number of bytes encoded by gel.CustomDuration "+
		"at args[0] expected 16, got 1")
}

type CustomOptionalDuration struct {
	data  []byte
	isSet bool
}

func (m CustomOptionalDuration) MarshalEdgeDBDuration() ([]byte, error) {
	if !m.isSet {
		return nil, fmt.Errorf("%T is not set", m)
	}
	data := make([]byte, len(m.data))
	copy(data, m.data)
	return data, nil
}

func (m *CustomOptionalDuration) UnmarshalEdgeDBDuration(data []byte) error {
	m.isSet = true
	m.data = make([]byte, len(data))
	copy(m.data, data)
	return nil
}

func (m *CustomOptionalDuration) SetMissing(missing bool) {
	m.isSet = !missing
	m.data = nil
}

func (m CustomOptionalDuration) Missing() bool { return !m.isSet }

func TestReceiveOptionalDurationUnmarshaler(t *testing.T) {
	ddl := `CREATE TYPE Sample { CREATE PROPERTY val -> duration; };`
	inRolledBackTx(t, ddl, func(ctx context.Context, tx *Tx) {
		var result struct {
			Val CustomOptionalDuration `gel:"val"`
		}

		// Decode value
		err := tx.QuerySingle(ctx, `
			SELECT { val := <duration>'48 hours 45 minutes 7.6 seconds' }`,
			&result,
		)
		assert.NoError(t, err)
		assert.Equal(t,
			[]byte{
				0x00, 0x00, 0x00, 0x28, 0xdd, 0x11, 0x72, 0x80, // microseconds
				0x00, 0x00, 0x00, 0x00, // days
				0x00, 0x00, 0x00, 0x00, // months
			},
			result.Val.data,
		)

		// Decode missing value
		query := `WITH inserted := (INSERT Sample) SELECT inserted { val }`
		err = tx.QuerySingle(ctx, query, &result)
		assert.NoError(t, err)
		assert.Equal(t, CustomOptionalDuration{}, result.Val)
	})
}

func TestSendOptionalDurationMarshaler(t *testing.T) {
	ctx := context.Background()
	var result struct {
		Val types.OptionalDuration `gel:"val"`
	}

	newValue := func(data []byte) CustomOptionalDuration {
		return CustomOptionalDuration{isSet: true, data: data}
	}

	// encode value into required argument
	err := client.QuerySingle(ctx, `
		SELECT { val := <duration>$0 }`,
		&result,
		newValue([]byte{
			0x00, 0x00, 0x00, 0x28, 0xdd, 0x11, 0x72, 0x80, // microseconds
			0x00, 0x00, 0x00, 0x00, // days
			0x00, 0x00, 0x00, 0x00, // months
		}),
	)
	assert.NoError(t, err)
	assert.Equal(t, types.NewOptionalDuration(0x28dd117280), result.Val)

	// encode value into optional argument
	err = client.QuerySingle(ctx, `
		SELECT { val := <OPTIONAL duration>$0 }`,
		&result,
		newValue([]byte{
			0x00, 0x00, 0x00, 0x28, 0xdd, 0x11, 0x72, 0x80, // microseconds
			0x00, 0x00, 0x00, 0x00, // days
			0x00, 0x00, 0x00, 0x00, // months
		}),
	)
	assert.NoError(t, err)
	assert.Equal(t, types.NewOptionalDuration(0x28dd117280), result.Val)

	// encode missing value into optional argument
	err = client.QuerySingle(ctx, `
		SELECT { val := <OPTIONAL duration>$0 }`,
		&result,
		CustomOptionalDuration{},
	)
	assert.NoError(t, err)
	assert.Equal(t, types.OptionalDuration{}, result.Val)

	// encode missing value into required argument
	err = client.QuerySingle(ctx, `
		SELECT { val := <duration>$0 }`,
		&result,
		CustomOptionalDuration{},
	)
	assert.EqualError(t, err, "gel.InvalidArgumentError: "+
		"cannot encode gel.CustomOptionalDuration at args[0] "+
		"because its value is missing")

	// encode wrong number of bytes with required argument
	err = client.QuerySingle(ctx, `
		SELECT { val := <duration>$0 }`,
		&result,
		newValue([]byte{0x01}),
	)
	assert.EqualError(t, err, "gel.InvalidArgumentError: "+
		"wrong number of bytes encoded by gel.CustomOptionalDuration "+
		"at args[0] expected 16, got 1")

	// encode wrong number of bytes with optional argument
	err = client.QuerySingle(ctx, `
		SELECT { val := <OPTIONAL duration>$0 }`,
		&result,
		newValue([]byte{0x01}),
	)
	assert.EqualError(t, err, "gel.InvalidArgumentError: "+
		"wrong number of bytes encoded by gel.CustomOptionalDuration "+
		"at args[0] expected 16, got 1")
}

func TestSendAndReceiveRelativeDuration(t *testing.T) {
	ctx := context.Background()

	var duration types.RelativeDuration
	err := client.QuerySingle(ctx,
		"SELECT <cal::relative_duration>'1y'", &duration)
	if err != nil {
		t.Skip("server version is too old for this feature")
	}

	rds := []types.RelativeDuration{
		types.NewRelativeDuration(0, 0, 0),
		types.NewRelativeDuration(0, 0, 1),
		types.NewRelativeDuration(0, 0, -1),
		types.NewRelativeDuration(0, 1, 0),
		types.NewRelativeDuration(0, -1, 0),
		types.NewRelativeDuration(1, 0, 0),
		types.NewRelativeDuration(-1, 0, 0),
		types.NewRelativeDuration(1, 1, 1),
		types.NewRelativeDuration(-1, -1, -1),
	}

	for i := 0; i < 5_000; i++ {
		rds = append(rds, types.NewRelativeDuration(
			rand.Int31n(101)-int32(50),
			rand.Int31n(1_001)-int32(500),
			rand.Int63n(2_000_000_000)-int64(1_000_000_000),
		))
	}

	type Result struct {
		RoundTrip types.RelativeDuration `gel:"round_trip"`
		Str       string                 `gel:"str"`
	}

	query := `
		WITH args := array_unpack(<array<cal::relative_duration>>$0)
		SELECT (
			round_trip := args,
			str := <str>args,
		)
	`

	var results []Result
	err = client.Query(ctx, query, &results, rds)
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
	data []byte
}

func (m CustomRelativeDuration) MarshalEdgeDBRelativeDuration() (
	[]byte, error) {
	data := make([]byte, len(m.data))
	copy(data, m.data)
	return data, nil
}

func (m *CustomRelativeDuration) UnmarshalEdgeDBRelativeDuration(
	data []byte,
) error {
	m.data = make([]byte, len(data))
	copy(m.data, data)
	return nil
}

func TestReceiveRelativeDurationUnmarshaler(t *testing.T) {
	ctx := context.Background()
	var result struct {
		Val CustomRelativeDuration `gel:"val"`
	}

	// Decode value
	err := client.QuerySingle(ctx, `
		SELECT { val := <cal::relative_duration>
			'8 months 5 days 48 hours 45 minutes 7.6 seconds'
		}`,
		&result,
	)
	assert.NoError(t, err)
	assert.Equal(t,
		[]byte{
			0x00, 0x00, 0x00, 0x28, 0xdd, 0x11, 0x72, 0x80, // microseconds
			0x00, 0x00, 0x00, 0x05, // days
			0x00, 0x00, 0x00, 0x08, // months
		},
		result.Val.data,
	)

	// Decode missing value
	err = client.QuerySingle(ctx, `
		SELECT { val := <OPTIONAL cal::relative_duration>$0 }`,
		&result,
		types.OptionalRelativeDuration{},
	)
	assert.EqualError(t, err, "gel.InvalidArgumentError: "+
		"the \"out\" argument does not match query schema: "+
		"expected gel.CustomRelativeDuration at struct "+
		"{ Val gel.CustomRelativeDuration \"gel:\\\"val\\\"\" }.val "+
		"to be OptionalUnmarshaler interface "+
		"because the field is not required")
}

func TestSendRelativeDurationMarshaler(t *testing.T) {
	ctx := context.Background()
	var result struct {
		Val types.OptionalRelativeDuration `gel:"val"`
	}

	// encode value into required argument
	err := client.QuerySingle(ctx, `
		SELECT { val := <cal::relative_duration>$0 }`,
		&result,
		CustomRelativeDuration{data: []byte{
			0x00, 0x00, 0x00, 0x28, 0xdd, 0x11, 0x72, 0x80, // microseconds
			0x00, 0x00, 0x00, 0x05, // days
			0x00, 0x00, 0x00, 0x08, // months
		}},
	)
	assert.NoError(t, err)
	assert.Equal(
		t,
		types.NewOptionalRelativeDuration(
			types.NewRelativeDuration(8, 5, 0x28dd117280),
		),
		result.Val,
	)

	// encode value into optional argument
	err = client.QuerySingle(ctx, `
		SELECT { val := <OPTIONAL cal::relative_duration>$0 }`,
		&result,
		CustomRelativeDuration{data: []byte{
			0x00, 0x00, 0x00, 0x28, 0xdd, 0x11, 0x72, 0x80, // microseconds
			0x00, 0x00, 0x00, 0x05, // days
			0x00, 0x00, 0x00, 0x08, // months
		}},
	)
	assert.NoError(t, err)
	assert.Equal(
		t,
		types.NewOptionalRelativeDuration(
			types.NewRelativeDuration(8, 5, 0x28dd117280),
		),
		result.Val,
	)

	// encode wrong number of bytes
	err = client.QuerySingle(ctx, `
		SELECT { val := <cal::relative_duration>$0 }`,
		&result,
		CustomRelativeDuration{data: []byte{0x01}},
	)
	assert.EqualError(t, err, "gel.InvalidArgumentError: "+
		"wrong number of bytes encoded by gel.CustomRelativeDuration "+
		"at args[0] expected 16, got 1")
}

type CustomOptionalRelativeDuration struct {
	data  []byte
	isSet bool
}

func (m CustomOptionalRelativeDuration) MarshalEdgeDBRelativeDuration() (
	[]byte, error) {
	if !m.isSet {
		return nil, fmt.Errorf("%T is not set", m)
	}
	data := make([]byte, len(m.data))
	copy(data, m.data)
	return data, nil
}

func (m *CustomOptionalRelativeDuration) UnmarshalEdgeDBRelativeDuration(
	data []byte,
) error {
	m.isSet = true
	m.data = make([]byte, len(data))
	copy(m.data, data)
	return nil
}

func (m *CustomOptionalRelativeDuration) SetMissing(missing bool) {
	m.isSet = !missing
	m.data = nil
}

func (m CustomOptionalRelativeDuration) Missing() bool { return !m.isSet }

func TestReceiveOptionalRelativeDurationUnmarshaler(t *testing.T) {
	ddl := `CREATE TYPE Sample {
		CREATE PROPERTY val -> cal::relative_duration;
	};`
	inRolledBackTx(t, ddl, func(ctx context.Context, tx *Tx) {
		var result struct {
			Val CustomOptionalRelativeDuration `gel:"val"`
		}

		// Decode value
		err := tx.QuerySingle(ctx, `
			SELECT { val := <cal::relative_duration>
				'8 months 5 days 48 hours 45 minutes 7.6 seconds'
			}`,
			&result,
		)
		assert.NoError(t, err)
		assert.Equal(t,
			[]byte{
				0x00, 0x00, 0x00, 0x28, 0xdd, 0x11, 0x72, 0x80, // microseconds
				0x00, 0x00, 0x00, 0x05, // days
				0x00, 0x00, 0x00, 0x08, // months
			},
			result.Val.data,
		)

		// Decode missing value
		query := `WITH inserted := (INSERT Sample) SELECT inserted { val }`
		err = tx.QuerySingle(ctx, query, &result)
		assert.NoError(t, err)
		assert.Equal(t, CustomOptionalRelativeDuration{}, result.Val)
	})
}

func TestSendOptionalRelativeDurationMarshaler(t *testing.T) {
	ctx := context.Background()
	var result struct {
		Val types.OptionalRelativeDuration `gel:"val"`
	}

	newValue := func(data []byte) CustomOptionalRelativeDuration {
		return CustomOptionalRelativeDuration{isSet: true, data: data}
	}

	// encode value into required argument
	err := client.QuerySingle(ctx, `
		SELECT { val := <cal::relative_duration>$0 }`,
		&result,
		newValue([]byte{
			0x00, 0x00, 0x00, 0x28, 0xdd, 0x11, 0x72, 0x80, // microseconds
			0x00, 0x00, 0x00, 0x05, // days
			0x00, 0x00, 0x00, 0x08, // months
		}),
	)
	assert.NoError(t, err)
	assert.Equal(
		t,
		types.NewOptionalRelativeDuration(
			types.NewRelativeDuration(8, 5, 0x28dd117280),
		),
		result.Val,
	)

	// encode value into optional argument
	err = client.QuerySingle(ctx, `
		SELECT { val := <OPTIONAL cal::relative_duration>$0 }`,
		&result,
		newValue([]byte{
			0x00, 0x00, 0x00, 0x28, 0xdd, 0x11, 0x72, 0x80, // microseconds
			0x00, 0x00, 0x00, 0x05, // days
			0x00, 0x00, 0x00, 0x08, // months
		}),
	)
	assert.NoError(t, err)
	assert.Equal(
		t,
		types.NewOptionalRelativeDuration(
			types.NewRelativeDuration(8, 5, 0x28dd117280),
		),
		result.Val,
	)

	// encode missing value into optional argument
	err = client.QuerySingle(ctx, `
		SELECT { val := <OPTIONAL cal::relative_duration>$0 }`,
		&result,
		CustomOptionalRelativeDuration{},
	)
	assert.NoError(t, err)
	assert.Equal(t, types.OptionalRelativeDuration{}, result.Val)

	// encode missing value into required argument
	err = client.QuerySingle(ctx, `
		SELECT { val := <cal::relative_duration>$0 }`,
		&result,
		CustomOptionalRelativeDuration{},
	)
	assert.EqualError(t, err, "gel.InvalidArgumentError: "+
		"cannot encode gel.CustomOptionalRelativeDuration at args[0] "+
		"because its value is missing")

	// encode wrong number of bytes with required argument
	err = client.QuerySingle(ctx, `
		SELECT { val := <cal::relative_duration>$0 }`,
		&result,
		newValue([]byte{0x01}),
	)
	assert.EqualError(t, err, "gel.InvalidArgumentError: "+
		"wrong number of bytes encoded by "+
		"gel.CustomOptionalRelativeDuration at args[0] expected 16, got 1")

	// encode wrong number of bytes with optional argument
	err = client.QuerySingle(ctx, `
		SELECT { val := <OPTIONAL cal::relative_duration>$0 }`,
		&result,
		newValue([]byte{0x01}),
	)
	assert.EqualError(t, err, "gel.InvalidArgumentError: "+
		"wrong number of bytes encoded "+
		"by gel.CustomOptionalRelativeDuration "+
		"at args[0] expected 16, got 1")
}

func TestSendAndReceiveDateDuration(t *testing.T) {
	if protocolVersion.LT(protocolVersion1p0) {
		t.Skip()
	}

	ctx := context.Background()

	var duration types.DateDuration
	err := client.QuerySingle(ctx,
		"SELECT <cal::date_duration>'1y'",
		&duration,
	)
	assert.NoError(t, err)

	rds := []types.DateDuration{
		types.NewDateDuration(0, 0),
		types.NewDateDuration(0, 1),
		types.NewDateDuration(0, -1),
		types.NewDateDuration(1, 0),
		types.NewDateDuration(-1, 0),
		types.NewDateDuration(1, 1),
		types.NewDateDuration(-1, -1),
		types.NewDateDuration(1, -1),
		types.NewDateDuration(-1, 1),
	}

	for i := 0; i < 5_000; i++ {
		rds = append(rds, types.NewDateDuration(
			rand.Int31n(101)-int32(50),
			rand.Int31n(1_001)-int32(500),
		))
	}

	type Result struct {
		RoundTrip types.DateDuration `gel:"round_trip"`
		Str       string             `gel:"str"`
	}

	query := `
		WITH args := array_unpack(<array<cal::date_duration>>$0)
		SELECT (
			round_trip := args,
			str := <str>args,
		)
	`

	var results []Result
	err = client.Query(ctx, query, &results, rds)
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

type CustomDateDuration struct {
	data []byte
}

func (m CustomDateDuration) MarshalEdgeDBDateDuration() (
	[]byte, error) {
	data := make([]byte, len(m.data))
	copy(data, m.data)
	return data, nil
}

func (m *CustomDateDuration) UnmarshalEdgeDBDateDuration(
	data []byte,
) error {
	m.data = make([]byte, len(data))
	copy(m.data, data)
	return nil
}

func TestReceiveDateDurationUnmarshaler(t *testing.T) {
	if protocolVersion.LT(protocolVersion1p0) {
		t.Skip()
	}
	ctx := context.Background()
	var result struct {
		Val CustomDateDuration `gel:"val"`
	}

	// Decode value
	err := client.QuerySingle(ctx,
		`SELECT { val := <cal::date_duration> '8 months 5 days' }`,
		&result,
	)
	assert.NoError(t, err)
	assert.Equal(t,
		[]byte{
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // microseconds
			0x00, 0x00, 0x00, 0x05, // days
			0x00, 0x00, 0x00, 0x08, // months
		},
		result.Val.data,
	)

	// Decode missing value
	err = client.QuerySingle(ctx, `
		SELECT { val := <OPTIONAL cal::date_duration>$0 }`,
		&result,
		types.OptionalDateDuration{},
	)
	assert.EqualError(t, err, "gel.InvalidArgumentError: "+
		"the \"out\" argument does not match query schema: "+
		"expected gel.CustomDateDuration at struct "+
		"{ Val gel.CustomDateDuration \"gel:\\\"val\\\"\" }.val "+
		"to be OptionalUnmarshaler interface "+
		"because the field is not required")
}

func TestSendDateDurationMarshaler(t *testing.T) {
	if protocolVersion.LT(protocolVersion1p0) {
		t.Skip()
	}

	ctx := context.Background()
	var result struct {
		Val types.OptionalDateDuration `gel:"val"`
	}

	// encode value into required argument
	err := client.QuerySingle(ctx, `
		SELECT { val := <cal::date_duration>$0 }`,
		&result,
		CustomDateDuration{data: []byte{
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // microseconds
			0x00, 0x00, 0x00, 0x05, // days
			0x00, 0x00, 0x00, 0x08, // months
		}},
	)
	assert.NoError(t, err)
	assert.Equal(t,
		types.NewOptionalDateDuration(types.NewDateDuration(8, 5)),
		result.Val,
	)

	// encode value into optional argument
	err = client.QuerySingle(ctx, `
		SELECT { val := <OPTIONAL cal::date_duration>$0 }`,
		&result,
		CustomDateDuration{data: []byte{
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // microseconds
			0x00, 0x00, 0x00, 0x05, // days
			0x00, 0x00, 0x00, 0x08, // months
		}},
	)
	assert.NoError(t, err)
	assert.Equal(t,
		types.NewOptionalDateDuration(types.NewDateDuration(8, 5)),
		result.Val,
	)

	// encode wrong number of bytes
	err = client.QuerySingle(ctx, `
		SELECT { val := <cal::date_duration>$0 }`,
		&result,
		CustomDateDuration{data: []byte{0x01}},
	)
	assert.EqualError(t, err, "gel.InvalidArgumentError: "+
		"wrong number of bytes encoded by gel.CustomDateDuration "+
		"at args[0] expected 16, got 1")
}

type CustomOptionalDateDuration struct {
	data  []byte
	isSet bool
}

func (m CustomOptionalDateDuration) MarshalEdgeDBDateDuration() (
	[]byte, error) {
	if !m.isSet {
		return nil, fmt.Errorf("%T is not set", m)
	}
	data := make([]byte, len(m.data))
	copy(data, m.data)
	return data, nil
}

func (m *CustomOptionalDateDuration) UnmarshalEdgeDBDateDuration(
	data []byte,
) error {
	m.isSet = true
	m.data = make([]byte, len(data))
	copy(m.data, data)
	return nil
}

func (m *CustomOptionalDateDuration) SetMissing(missing bool) {
	m.isSet = !missing
	m.data = nil
}

func (m CustomOptionalDateDuration) Missing() bool { return !m.isSet }

func TestReceiveOptionalDateDurationUnmarshaler(t *testing.T) {
	if protocolVersion.LT(protocolVersion1p0) {
		t.Skip()
	}

	ddl := `CREATE TYPE Sample {
		CREATE PROPERTY val -> cal::date_duration;
	};`
	inRolledBackTx(t, ddl, func(ctx context.Context, tx *Tx) {
		var result struct {
			Val CustomOptionalDateDuration `gel:"val"`
		}

		// Decode value
		err := tx.QuerySingle(ctx,
			`SELECT { val := <cal::date_duration>'8 months 5 days' }`,
			&result,
		)
		assert.NoError(t, err)
		assert.Equal(t,
			[]byte{
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // microseconds
				0x00, 0x00, 0x00, 0x05, // days
				0x00, 0x00, 0x00, 0x08, // months
			},
			result.Val.data,
		)

		// Decode missing value
		query := `WITH inserted := (INSERT Sample) SELECT inserted { val }`
		err = tx.QuerySingle(ctx, query, &result)
		assert.NoError(t, err)
		assert.Equal(t, CustomOptionalDateDuration{}, result.Val)
	})
}

func TestSendOptionalDateDurationMarshaler(t *testing.T) {
	if protocolVersion.LT(protocolVersion1p0) {
		t.Skip()
	}

	ctx := context.Background()
	var result struct {
		Val types.OptionalDateDuration `gel:"val"`
	}

	newValue := func(data []byte) CustomOptionalDateDuration {
		return CustomOptionalDateDuration{isSet: true, data: data}
	}

	// encode value into required argument
	err := client.QuerySingle(ctx, `
		SELECT { val := <cal::date_duration>$0 }`,
		&result,
		newValue([]byte{
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // microseconds
			0x00, 0x00, 0x00, 0x05, // days
			0x00, 0x00, 0x00, 0x08, // months
		}),
	)
	assert.NoError(t, err)
	assert.Equal(t,
		types.NewOptionalDateDuration(types.NewDateDuration(8, 5)),
		result.Val,
	)

	// encode value into optional argument
	err = client.QuerySingle(ctx, `
		SELECT { val := <OPTIONAL cal::date_duration>$0 }`,
		&result,
		newValue([]byte{
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // microseconds
			0x00, 0x00, 0x00, 0x05, // days
			0x00, 0x00, 0x00, 0x08, // months
		}),
	)
	assert.NoError(t, err)
	assert.Equal(t,
		types.NewOptionalDateDuration(types.NewDateDuration(8, 5)),
		result.Val,
	)

	// encode missing value into optional argument
	err = client.QuerySingle(ctx, `
		SELECT { val := <OPTIONAL cal::date_duration>$0 }`,
		&result,
		CustomOptionalDateDuration{},
	)
	assert.NoError(t, err)
	assert.Equal(t, types.OptionalDateDuration{}, result.Val)

	// encode missing value into required argument
	err = client.QuerySingle(ctx, `
		SELECT { val := <cal::date_duration>$0 }`,
		&result,
		CustomOptionalDateDuration{},
	)
	assert.EqualError(t, err, "gel.InvalidArgumentError: "+
		"cannot encode gel.CustomOptionalDateDuration at args[0] "+
		"because its value is missing")

	// encode wrong number of bytes with required argument
	err = client.QuerySingle(ctx, `
		SELECT { val := <cal::date_duration>$0 }`,
		&result,
		newValue([]byte{0x01}),
	)
	assert.EqualError(t, err, "gel.InvalidArgumentError: "+
		"wrong number of bytes encoded by "+
		"gel.CustomOptionalDateDuration at args[0] expected 16, got 1")

	// encode wrong number of bytes with optional argument
	err = client.QuerySingle(ctx, `
		SELECT { val := <OPTIONAL cal::date_duration>$0 }`,
		&result,
		newValue([]byte{0x01}),
	)
	assert.EqualError(t, err, "gel.InvalidArgumentError: "+
		"wrong number of bytes encoded "+
		"by gel.CustomOptionalDateDuration "+
		"at args[0] expected 16, got 1")
}

func TestSendAndReceiveLocalTime(t *testing.T) {
	ctx := context.Background()

	times := []types.LocalTime{
		types.NewLocalTime(0, 0, 0, 0),
		types.NewLocalTime(0, 0, 0, 1),
		types.NewLocalTime(0, 0, 0, 10),
		types.NewLocalTime(0, 0, 0, 100),
		types.NewLocalTime(0, 0, 0, 1000),
		types.NewLocalTime(0, 0, 0, 10000),
		types.NewLocalTime(0, 0, 0, 100000),
		types.NewLocalTime(0, 0, 0, 123456),
		types.NewLocalTime(0, 1, 11, 340000),
		types.NewLocalTime(5, 4, 3, 0),
		types.NewLocalTime(11, 12, 13, 0),
		types.NewLocalTime(20, 39, 57, 0),
		types.NewLocalTime(23, 59, 59, 999000),
		types.NewLocalTime(23, 59, 59, 999999),
	}

	for i := 0; i < 1_000; i++ {
		times = append(times, types.NewLocalTime(
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
		Encoded   string          `gel:"encoded"`
		Decoded   types.LocalTime `gel:"decoded"`
		RoundTrip types.LocalTime `gel:"round_trip"`
		IsEqual   bool            `gel:"is_equal"`
		String    string          `gel:"string"`
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
	err := client.Query(ctx, query, &results, times, strings)
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
	data []byte
}

func (m CustomLocalTime) MarshalEdgeDBLocalTime() ([]byte, error) {
	data := make([]byte, len(m.data))
	copy(data, m.data)
	return data, nil
}

func (m *CustomLocalTime) UnmarshalEdgeDBLocalTime(data []byte) error {
	m.data = make([]byte, len(data))
	copy(m.data, data)
	return nil
}

func TestReceiveLocalTimeUnmarshaler(t *testing.T) {
	ctx := context.Background()
	var result struct {
		Val CustomLocalTime `gel:"val"`
	}

	// Decode value
	err := client.QuerySingle(ctx, `
		SELECT { val := <cal::local_time>'12:10:00' }`,
		&result,
	)
	assert.NoError(t, err)
	assert.Equal(t,
		[]byte{0x00, 0x00, 0x00, 0x0a, 0x32, 0xae, 0xf6, 0x00},
		result.Val.data,
	)

	// Decode missing value
	err = client.QuerySingle(ctx, `
		SELECT { val := <OPTIONAL cal::local_time>$0 }`,
		&result,
		types.OptionalLocalTime{},
	)
	assert.EqualError(t, err, "gel.InvalidArgumentError: "+
		"the \"out\" argument does not match query schema: "+
		"expected gel.CustomLocalTime at "+
		"struct { Val gel.CustomLocalTime \"gel:\\\"val\\\"\" }.val "+
		"to be OptionalUnmarshaler interface "+
		"because the field is not required")
}

func TestSendLocalTimeMarshaler(t *testing.T) {
	ctx := context.Background()
	var result struct {
		Val types.OptionalLocalTime `gel:"val"`
	}

	// encode value into required argument
	err := client.QuerySingle(ctx, `
		SELECT { val := <cal::local_time>$0 }`,
		&result,
		CustomLocalTime{data: []byte{
			0x00, 0x00, 0x00, 0x0a, 0x32, 0xae, 0xf6, 0x00}},
	)
	assert.NoError(t, err)
	assert.Equal(t,
		types.NewOptionalLocalTime(types.NewLocalTime(12, 10, 0, 0)),
		result.Val,
	)

	// encode value into optional argument
	err = client.QuerySingle(ctx, `
		SELECT { val := <OPTIONAL cal::local_time>$0 }`,
		&result,
		CustomLocalTime{data: []byte{
			0x00, 0x00, 0x00, 0x0a, 0x32, 0xae, 0xf6, 0x00}},
	)
	assert.NoError(t, err)
	assert.Equal(t,
		types.NewOptionalLocalTime(types.NewLocalTime(12, 10, 0, 0)),
		result.Val,
	)

	// encode wrong number of bytes
	err = client.QuerySingle(ctx, `
		SELECT { val := <cal::local_time>$0 }`,
		&result,
		CustomLocalTime{data: []byte{0x01}},
	)
	assert.EqualError(t, err, "gel.InvalidArgumentError: "+
		"wrong number of bytes encoded by gel.CustomLocalTime "+
		"at args[0] expected 8, got 1")
}

type CustomOptionalLocalTime struct {
	data  []byte
	isSet bool
}

func (m CustomOptionalLocalTime) MarshalEdgeDBLocalTime() ([]byte, error) {
	if !m.isSet {
		return nil, fmt.Errorf("%T is not set", m)
	}
	data := make([]byte, len(m.data))
	copy(data, m.data)
	return data, nil
}

func (m *CustomOptionalLocalTime) UnmarshalEdgeDBLocalTime(data []byte) error {
	m.isSet = true
	m.data = make([]byte, len(data))
	copy(m.data, data)
	return nil
}

func (m *CustomOptionalLocalTime) SetMissing(missing bool) {
	m.isSet = !missing
	m.data = nil
}

func (m CustomOptionalLocalTime) Missing() bool { return !m.isSet }

func TestReceiveOptionalLocalTimeUnmarshaler(t *testing.T) {
	ddl := `CREATE TYPE Sample { CREATE PROPERTY val -> cal::local_time; };`
	inRolledBackTx(t, ddl, func(ctx context.Context, tx *Tx) {
		var result struct {
			Val CustomOptionalLocalTime `gel:"val"`
		}

		// Decode value
		err := tx.QuerySingle(ctx, `
			SELECT { val := <cal::local_time>'12:10:00' }`,
			&result,
		)
		assert.NoError(t, err)
		assert.Equal(t,
			[]byte{0x00, 0x00, 0x00, 0x0a, 0x32, 0xae, 0xf6, 0x00},
			result.Val.data,
		)

		// Decode missing value
		query := `WITH inserted := (INSERT Sample) SELECT inserted { val }`
		err = tx.QuerySingle(ctx, query, &result)
		assert.NoError(t, err)
		assert.Equal(t, CustomOptionalLocalTime{}, result.Val)
	})
}

func TestSendOptionalLocalTimeMarshaler(t *testing.T) {
	ctx := context.Background()
	var result struct {
		Val types.OptionalLocalTime `gel:"val"`
	}

	newValue := func(data []byte) CustomOptionalLocalTime {
		return CustomOptionalLocalTime{isSet: true, data: data}
	}

	// encode value into required argument
	err := client.QuerySingle(ctx, `
		SELECT { val := <cal::local_time>$0 }`,
		&result,
		newValue([]byte{0x00, 0x00, 0x00, 0x0a, 0x32, 0xae, 0xf6, 0x00}),
	)
	assert.NoError(t, err)
	assert.Equal(t,
		types.NewOptionalLocalTime(types.NewLocalTime(12, 10, 0, 0)),
		result.Val,
	)

	// encode value into optional argument
	err = client.QuerySingle(ctx, `
		SELECT { val := <OPTIONAL cal::local_time>$0 }`,
		&result,
		newValue([]byte{0x00, 0x00, 0x00, 0x0a, 0x32, 0xae, 0xf6, 0x00}),
	)
	assert.NoError(t, err)
	assert.Equal(t,
		types.NewOptionalLocalTime(types.NewLocalTime(12, 10, 0, 0)),
		result.Val,
	)

	// encode missing value into optional argument
	err = client.QuerySingle(ctx, `
	SELECT { val := <OPTIONAL cal::local_time>$0 }`,
		&result,
		CustomOptionalLocalTime{},
	)
	assert.NoError(t, err)
	assert.Equal(t, types.OptionalLocalTime{}, result.Val)

	// encode missing value into required argument
	err = client.QuerySingle(ctx, `
		SELECT { val := <cal::local_time>$0 }`,
		&result,
		CustomOptionalLocalTime{},
	)
	assert.EqualError(t, err, "gel.InvalidArgumentError: "+
		"cannot encode gel.CustomOptionalLocalTime at args[0] "+
		"because its value is missing")

	// encode wrong number of bytes with required argument
	err = client.QuerySingle(ctx, `
		SELECT { val := <cal::local_time>$0 }`,
		&result,
		newValue([]byte{0x01}),
	)
	assert.EqualError(t, err, "gel.InvalidArgumentError: "+
		"wrong number of bytes encoded by gel.CustomOptionalLocalTime "+
		"at args[0] expected 8, got 1")

	// encode wrong number of bytes with optional argument
	err = client.QuerySingle(ctx, `
		SELECT { val := <OPTIONAL cal::local_time>$0 }`,
		&result,
		newValue([]byte{0x01}),
	)
	assert.EqualError(t, err, "gel.InvalidArgumentError: "+
		"wrong number of bytes encoded by gel.CustomOptionalLocalTime "+
		"at args[0] expected 8, got 1")
}

func TestSendAndReceiveLocalDate(t *testing.T) {
	ctx := context.Background()

	dates := []types.LocalDate{
		types.NewLocalDate(1, 1, 1),
		types.NewLocalDate(2000, 1, 1),
		types.NewLocalDate(2019, 5, 6),
		types.NewLocalDate(4444, 12, 30),
		types.NewLocalDate(9999, 9, 9),
	}

	for i := 0; i < 1_000; i++ {
		dates = append(dates, types.NewLocalDate(
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
		Encoded   string          `gel:"encoded"`
		Decoded   types.LocalDate `gel:"decoded"`
		RoundTrip types.LocalDate `gel:"round_trip"`
		IsEqual   bool            `gel:"is_equal"`
		String    string          `gel:"string"`
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
	err := client.Query(ctx, query, &results, dates, strings)
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
	data []byte
}

func (m CustomLocalDate) MarshalEdgeDBLocalDate() ([]byte, error) {
	data := make([]byte, len(m.data))
	copy(data, m.data)
	return data, nil
}

func (m *CustomLocalDate) UnmarshalEdgeDBLocalDate(data []byte) error {
	m.data = make([]byte, len(data))
	copy(m.data, data)
	return nil
}

func TestReceiveLocalDateUnmarshaler(t *testing.T) {
	ctx := context.Background()
	var result struct {
		Val CustomLocalDate `gel:"val"`
	}

	// Decode value
	err := client.QuerySingle(ctx, `
		SELECT { val := <cal::local_date>'2019-05-06' }`,
		&result,
	)
	assert.NoError(t, err)
	assert.Equal(t, []byte{0x00, 0x00, 0x1b, 0x99}, result.Val.data)

	// Decode missing value
	err = client.QuerySingle(ctx, `
		SELECT { val := <OPTIONAL cal::local_date>$0 }`,
		&result,
		types.OptionalLocalDate{},
	)
	assert.EqualError(t, err, "gel.InvalidArgumentError: "+
		"the \"out\" argument does not match query schema: "+
		"expected gel.CustomLocalDate at "+
		"struct { Val gel.CustomLocalDate \"gel:\\\"val\\\"\" }.val "+
		"to be OptionalUnmarshaler interface "+
		"because the field is not required")
}

func TestSendLocalDateMarshaler(t *testing.T) {
	ctx := context.Background()
	var result struct {
		Val types.OptionalLocalDate `gel:"val"`
	}

	// encode value into required argument
	err := client.QuerySingle(ctx, `
		SELECT { val := <cal::local_date>$0 }`,
		&result,
		CustomLocalDate{data: []byte{0x00, 0x00, 0x1b, 0x99}},
	)
	assert.NoError(t, err)
	assert.Equal(t,
		types.NewOptionalLocalDate(types.NewLocalDate(2019, 5, 6)),
		result.Val,
	)

	// encode value into optional argument
	err = client.QuerySingle(ctx, `
		SELECT { val := <OPTIONAL cal::local_date>$0 }`,
		&result,
		CustomLocalDate{data: []byte{0x00, 0x00, 0x1b, 0x99}},
	)
	assert.NoError(t, err)
	assert.Equal(t,
		types.NewOptionalLocalDate(types.NewLocalDate(2019, 5, 6)),
		result.Val,
	)

	// encode wrong number of bytes
	err = client.QuerySingle(ctx, `
		SELECT { val := <cal::local_date>$0 }`,
		&result,
		CustomLocalDate{data: []byte{0x01}},
	)
	assert.EqualError(t, err, "gel.InvalidArgumentError: "+
		"wrong number of bytes encoded by gel.CustomLocalDate "+
		"at args[0] expected 4, got 1")
}

type CustomOptionalLocalDate struct {
	data  []byte
	isSet bool
}

func (m CustomOptionalLocalDate) MarshalEdgeDBLocalDate() ([]byte, error) {
	if !m.isSet {
		return nil, fmt.Errorf("%T is not set", m)
	}
	data := make([]byte, len(m.data))
	copy(data, m.data)
	return data, nil
}

func (m *CustomOptionalLocalDate) UnmarshalEdgeDBLocalDate(data []byte) error {
	m.isSet = true
	m.data = make([]byte, len(data))
	copy(m.data, data)
	return nil
}

func (m *CustomOptionalLocalDate) SetMissing(missing bool) {
	m.isSet = !missing
	m.data = nil
}

func (m CustomOptionalLocalDate) Missing() bool { return !m.isSet }

func TestReceiveOptionalLocalDateUnmarshaler(t *testing.T) {
	ddl := `CREATE TYPE Sample { CREATE PROPERTY val -> cal::local_date; };`
	inRolledBackTx(t, ddl, func(ctx context.Context, tx *Tx) {
		var result struct {
			Val CustomOptionalLocalDate `gel:"val"`
		}

		// Decode value
		err := tx.QuerySingle(ctx, `
			SELECT { val := <cal::local_date>'2019-05-06' }`,
			&result,
		)
		assert.NoError(t, err)
		assert.Equal(t, []byte{0x00, 0x00, 0x1b, 0x99}, result.Val.data)

		// Decode missing value
		query := `WITH inserted := (INSERT Sample) SELECT inserted { val }`
		err = tx.QuerySingle(ctx, query, &result)
		assert.NoError(t, err)
		assert.Equal(t, CustomOptionalLocalDate{}, result.Val)
	})
}

func TestSendOptionalLocalDateMarshaler(t *testing.T) {
	ctx := context.Background()
	var result struct {
		Val types.OptionalLocalDate `gel:"val"`
	}

	newValue := func(data []byte) CustomOptionalLocalDate {
		return CustomOptionalLocalDate{isSet: true, data: data}
	}

	// encode value into required argument
	err := client.QuerySingle(ctx, `
		SELECT { val := <cal::local_date>$0 }`,
		&result,
		newValue([]byte{0x00, 0x00, 0x1b, 0x99}),
	)
	assert.NoError(t, err)
	assert.Equal(t,
		types.NewOptionalLocalDate(types.NewLocalDate(2019, 5, 6)),
		result.Val,
	)

	// encode value into optional argument
	err = client.QuerySingle(ctx, `
		SELECT { val := <OPTIONAL cal::local_date>$0 }`,
		&result,
		newValue([]byte{0x00, 0x00, 0x1b, 0x99}),
	)
	assert.NoError(t, err)
	assert.Equal(t,
		types.NewOptionalLocalDate(types.NewLocalDate(2019, 5, 6)),
		result.Val,
	)

	// encode missing value into optional argument
	err = client.QuerySingle(ctx, `
		SELECT { val := <OPTIONAL cal::local_date>$0 }`,
		&result,
		CustomOptionalLocalDate{},
	)
	assert.NoError(t, err)
	assert.Equal(t, types.OptionalLocalDate{}, result.Val)

	// encode missing value into required argument
	err = client.QuerySingle(ctx, `
		SELECT { val := <cal::local_date>$0 }`,
		&result,
		CustomOptionalLocalDate{},
	)
	assert.EqualError(t, err, "gel.InvalidArgumentError: "+
		"cannot encode gel.CustomOptionalLocalDate at args[0] "+
		"because its value is missing")

	// encode wrong number of bytes with required argument
	err = client.QuerySingle(ctx, `
		SELECT { val := <cal::local_date>$0 }`,
		&result,
		newValue([]byte{0x01}),
	)
	assert.EqualError(t, err, "gel.InvalidArgumentError: "+
		"wrong number of bytes encoded by gel.CustomOptionalLocalDate "+
		"at args[0] expected 4, got 1")

	// encode wrong number of bytes with optional argument
	err = client.QuerySingle(ctx, `
		SELECT { val := <OPTIONAL cal::local_date>$0 }`,
		&result,
		newValue([]byte{0x01}),
	)
	assert.EqualError(t, err, "gel.InvalidArgumentError: "+
		"wrong number of bytes encoded by gel.CustomOptionalLocalDate "+
		"at args[0] expected 4, got 1")
}

func TestSendAndReceiveLocalDateTime(t *testing.T) {
	ctx := context.Background()

	datetimes := []types.LocalDateTime{
		types.NewLocalDateTime(2019, 5, 6, 12, 0, 0, 0),
		types.NewLocalDateTime(2018, 5, 7, 15, 1, 22, 306916),
		types.NewLocalDateTime(1, 1, 1, 1, 1, 0, 0),
		types.NewLocalDateTime(9999, 9, 9, 9, 9, 9, 0),
	}

	for i := 0; i < 1_000; i++ {
		dt := types.NewLocalDateTime(
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
		Encoded   string              `gel:"encoded"`
		Decoded   types.LocalDateTime `gel:"decoded"`
		RoundTrip types.LocalDateTime `gel:"round_trip"`
		IsEqual   bool                `gel:"is_equal"`
		String    string              `gel:"string"`
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
	err := client.Query(ctx, query, &results, datetimes, strings)
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
	data []byte
}

func (m CustomLocalDateTime) MarshalEdgeDBLocalDateTime() ([]byte, error) {
	data := make([]byte, len(m.data))
	copy(data, m.data)
	return data, nil
}

func (m *CustomLocalDateTime) UnmarshalEdgeDBLocalDateTime(data []byte) error {
	m.data = make([]byte, len(data))
	copy(m.data, data)
	return nil
}

func TestReceiveLocalDateTimeUnmarshaler(t *testing.T) {
	ctx := context.Background()
	var result struct {
		Val CustomLocalDateTime `gel:"val"`
	}

	// Decode value
	err := client.QuerySingle(ctx, `
		SELECT { val := <cal::local_datetime>'2019-05-06T12:00:00' }`,
		&result,
	)
	assert.NoError(t, err)
	assert.Equal(t,
		[]byte{0x00, 0x02, 0x2b, 0x35, 0x9b, 0xc4, 0x10, 0x00},
		result.Val.data,
	)

	// Decode missing value
	err = client.QuerySingle(ctx, `
		SELECT { val := <OPTIONAL cal::local_datetime>$0 }`,
		&result,
		types.OptionalLocalDateTime{},
	)
	assert.EqualError(t, err, "gel.InvalidArgumentError: "+
		"the \"out\" argument does not match query schema: "+
		"expected gel.CustomLocalDateTime at "+
		"struct { Val gel.CustomLocalDateTime \"gel:\\\"val\\\"\" }.val"+
		" to be OptionalUnmarshaler interface "+
		"because the field is not required")
}

func TestSendLocalDateTimeMarshaler(t *testing.T) {
	ctx := context.Background()
	var result struct {
		Val types.OptionalLocalDateTime `gel:"val"`
	}

	// encode value into required argument
	err := client.QuerySingle(ctx, `
		SELECT { val := <cal::local_datetime>$0 }`,
		&result,
		CustomLocalDateTime{data: []byte{
			0x00, 0x02, 0x2b, 0x35, 0x9b, 0xc4, 0x10, 0x00}},
	)
	assert.NoError(t, err)
	assert.Equal(
		t,
		types.NewOptionalLocalDateTime(
			types.NewLocalDateTime(2019, 5, 6, 12, 0, 0, 0),
		),
		result.Val,
	)

	// encode value into optional argument
	err = client.QuerySingle(ctx, `
		SELECT { val := <OPTIONAL cal::local_datetime>$0 }`,
		&result,
		CustomLocalDateTime{data: []byte{
			0x00, 0x02, 0x2b, 0x35, 0x9b, 0xc4, 0x10, 0x00}},
	)
	assert.NoError(t, err)
	assert.Equal(
		t,
		types.NewOptionalLocalDateTime(
			types.NewLocalDateTime(2019, 5, 6, 12, 0, 0, 0),
		),
		result.Val,
	)

	// encode wrong number of bytes
	err = client.QuerySingle(ctx, `
		SELECT { val := <cal::local_datetime>$0 }`,
		&result,
		CustomLocalDateTime{data: []byte{0x01}},
	)
	assert.EqualError(t, err, "gel.InvalidArgumentError: "+
		"wrong number of bytes encoded by gel.CustomLocalDateTime "+
		"at args[0] expected 8, got 1")
}

type CustomOptionalLocalDateTime struct {
	data  []byte
	isSet bool
}

func (m CustomOptionalLocalDateTime) MarshalEdgeDBLocalDateTime() (
	[]byte, error) {
	if !m.isSet {
		return nil, fmt.Errorf("%T is not set", m)
	}
	data := make([]byte, len(m.data))
	copy(data, m.data)
	return data, nil
}

func (m *CustomOptionalLocalDateTime) UnmarshalEdgeDBLocalDateTime(
	data []byte,
) error {
	m.isSet = true
	m.data = make([]byte, len(data))
	copy(m.data, data)
	return nil
}

func (m *CustomOptionalLocalDateTime) SetMissing(missing bool) {
	m.isSet = !missing
	m.data = nil
}

func (m CustomOptionalLocalDateTime) Missing() bool { return !m.isSet }

func TestReceiveOptionalLocalDateTimeUnmarshaler(t *testing.T) {
	ddl := `CREATE TYPE Sample {
		CREATE PROPERTY val -> cal::local_datetime;
	};`
	inRolledBackTx(t, ddl, func(ctx context.Context, tx *Tx) {
		var result struct {
			Val CustomOptionalLocalDateTime `gel:"val"`
		}

		// Decode value
		err := tx.QuerySingle(ctx,
			`SELECT { val := <cal::local_datetime>'2019-05-06T12:00:00' }`,
			&result,
		)
		assert.NoError(t, err)
		assert.Equal(t,
			[]byte{0x00, 0x02, 0x2b, 0x35, 0x9b, 0xc4, 0x10, 0x00},
			result.Val.data,
		)

		// Decode missing value
		query := `WITH inserted := (INSERT Sample) SELECT inserted { val }`
		err = tx.QuerySingle(ctx, query, &result)
		assert.NoError(t, err)
		assert.Equal(t, CustomOptionalLocalDateTime{}, result.Val)
	})
}

func TestSendOptionalLocalDateTimeMarshaler(t *testing.T) {
	ctx := context.Background()
	var result struct {
		Val types.OptionalLocalDateTime `gel:"val"`
	}

	newValue := func(data []byte) CustomOptionalLocalDateTime {
		return CustomOptionalLocalDateTime{isSet: true, data: data}
	}

	// encode value into required argument
	err := client.QuerySingle(ctx, `
		SELECT { val := <cal::local_datetime>$0 }`,
		&result,
		newValue([]byte{0x00, 0x02, 0x2b, 0x35, 0x9b, 0xc4, 0x10, 0x00}),
	)
	assert.NoError(t, err)
	assert.Equal(
		t,
		types.NewOptionalLocalDateTime(
			types.NewLocalDateTime(2019, 5, 6, 12, 0, 0, 0),
		),
		result.Val,
	)

	// encode value into optional argument
	err = client.QuerySingle(ctx, `
		SELECT { val := <OPTIONAL cal::local_datetime>$0 }`,
		&result,
		newValue([]byte{0x00, 0x02, 0x2b, 0x35, 0x9b, 0xc4, 0x10, 0x00}),
	)
	assert.NoError(t, err)
	assert.Equal(
		t,
		types.NewOptionalLocalDateTime(
			types.NewLocalDateTime(2019, 5, 6, 12, 0, 0, 0),
		),
		result.Val,
	)

	// encode missing value into optional argument
	err = client.QuerySingle(ctx, `
	SELECT { val := <OPTIONAL cal::local_datetime>$0 }`,
		&result,
		CustomOptionalLocalDateTime{},
	)
	assert.NoError(t, err)
	assert.Equal(t, types.OptionalLocalDateTime{}, result.Val)

	// encode missing value into required argument
	err = client.QuerySingle(ctx, `
		SELECT { val := <cal::local_datetime>$0 }`,
		&result,
		CustomOptionalLocalDateTime{},
	)
	assert.EqualError(t, err, "gel.InvalidArgumentError: "+
		"cannot encode gel.CustomOptionalLocalDateTime at args[0] "+
		"because its value is missing")

	// encode wrong number of bytes with required argument
	err = client.QuerySingle(ctx, `
		SELECT { val := <cal::local_datetime>$0 }`,
		&result,
		newValue([]byte{0x01}),
	)
	assert.EqualError(t, err, "gel.InvalidArgumentError: "+
		"wrong number of bytes encoded by gel.CustomOptionalLocalDateTime "+
		"at args[0] expected 8, got 1")

	// encode wrong number of bytes with optional argument
	err = client.QuerySingle(ctx, `
		SELECT { val := <OPTIONAL cal::local_datetime>$0 }`,
		&result,
		newValue([]byte{0x01}),
	)
	assert.EqualError(t, err, "gel.InvalidArgumentError: "+
		"wrong number of bytes encoded by gel.CustomOptionalLocalDateTime "+
		"at args[0] expected 8, got 1")
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
		Encoded   string    `gel:"encoded"`
		Decoded   time.Time `gel:"decoded"`
		RoundTrip time.Time `gel:"round_trip"`
		IsEqual   bool      `gel:"is_equal"`
		String    string    `gel:"string"`
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
	err := client.Query(ctx, query, &results, samples, strings)
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
	data []byte
}

func (m CustomDateTime) MarshalEdgeDBDateTime() ([]byte, error) {
	data := make([]byte, len(m.data))
	copy(data, m.data)
	return data, nil
}

func (m *CustomDateTime) UnmarshalEdgeDBDateTime(data []byte) error {
	m.data = make([]byte, len(data))
	copy(m.data, data)
	return nil
}

func TestReceiveDateTimeUnmarshaler(t *testing.T) {
	ctx := context.Background()
	var result struct {
		Val CustomDateTime `gel:"val"`
	}

	// Decode value
	err := client.QuerySingle(ctx, `
		SELECT { val := <datetime>'2019-05-06T12:00:00+00:00' }`,
		&result,
	)
	assert.NoError(t, err)
	assert.Equal(t,
		[]byte{0x00, 0x02, 0x2b, 0x35, 0x9b, 0xc4, 0x10, 0x00},
		result.Val.data,
	)

	// Decode missing value
	err = client.QuerySingle(ctx, `
		SELECT { val := <OPTIONAL datetime>$0 }`,
		&result,
		types.OptionalDateTime{},
	)
	assert.EqualError(t, err, "gel.InvalidArgumentError: "+
		"the \"out\" argument does not match query schema: "+
		"expected gel.CustomDateTime at "+
		"struct { Val gel.CustomDateTime \"gel:\\\"val\\\"\" }.val "+
		"to be OptionalUnmarshaler interface "+
		"because the field is not required")
}

func TestSendDateTimeMarshaler(t *testing.T) {
	ctx := context.Background()
	var result struct {
		Val types.OptionalDateTime `gel:"val"`
	}

	// encode value into required argument
	err := client.QuerySingle(ctx, `
		SELECT { val := <datetime>$0 }`,
		&result,
		CustomDateTime{data: []byte{
			0x00, 0x02, 0x2b, 0x35, 0x9b, 0xc4, 0x10, 0x00}},
	)
	assert.NoError(t, err)
	assert.Equal(
		t,
		types.NewOptionalDateTime(
			time.Date(2019, 5, 6, 12, 0, 0, 0, time.UTC),
		),
		result.Val,
	)

	// encode value into optional argument
	err = client.QuerySingle(ctx, `
		SELECT { val := <OPTIONAL datetime>$0 }`,
		&result,
		CustomDateTime{data: []byte{
			0x00, 0x02, 0x2b, 0x35, 0x9b, 0xc4, 0x10, 0x00}},
	)
	assert.NoError(t, err)
	assert.Equal(
		t,
		types.NewOptionalDateTime(
			time.Date(2019, 5, 6, 12, 0, 0, 0, time.UTC),
		),
		result.Val,
	)

	// encode wrong number of bytes
	err = client.QuerySingle(ctx, `
		SELECT { val := <datetime>$0 }`,
		&result,
		CustomDateTime{data: []byte{0x01}},
	)
	assert.EqualError(t, err, "gel.InvalidArgumentError: "+
		"wrong number of bytes encoded by gel.CustomDateTime "+
		"at args[0] expected 8, got 1")
}

type CustomOptionalDateTime struct {
	data  []byte
	isSet bool
}

func (m CustomOptionalDateTime) MarshalEdgeDBDateTime() ([]byte, error) {
	if !m.isSet {
		return nil, fmt.Errorf("%T is not set", m)
	}
	data := make([]byte, len(m.data))
	copy(data, m.data)
	return data, nil
}

func (m *CustomOptionalDateTime) UnmarshalEdgeDBDateTime(data []byte) error {
	m.isSet = true
	m.data = make([]byte, len(data))
	copy(m.data, data)
	return nil
}

func (m *CustomOptionalDateTime) SetMissing(missing bool) {
	m.isSet = !missing
	m.data = nil
}

func (m CustomOptionalDateTime) Missing() bool { return !m.isSet }

func TestReceiveOptionalDateTimeUnmarshaler(t *testing.T) {
	ddl := `CREATE TYPE Sample { CREATE PROPERTY val -> datetime; };`
	inRolledBackTx(t, ddl, func(ctx context.Context, tx *Tx) {
		var result struct {
			Val CustomOptionalDateTime `gel:"val"`
		}

		// Decode value
		err := tx.QuerySingle(ctx, `
			SELECT { val := <datetime>'2019-05-06T12:00:00+00:00' }`,
			&result,
		)
		assert.NoError(t, err)
		assert.Equal(t,
			[]byte{0x00, 0x02, 0x2b, 0x35, 0x9b, 0xc4, 0x10, 0x00},
			result.Val.data,
		)

		// Decode missing value
		query := `WITH inserted := (INSERT Sample) SELECT inserted { val }`
		err = tx.QuerySingle(ctx, query, &result)
		assert.NoError(t, err)
		assert.Equal(t, CustomOptionalDateTime{}, result.Val)
	})
}

func TestSendOptionalDateTimeMarshaler(t *testing.T) {
	ctx := context.Background()
	var result struct {
		Val types.OptionalDateTime `gel:"val"`
	}

	newValue := func(data []byte) CustomOptionalDateTime {
		return CustomOptionalDateTime{isSet: true, data: data}
	}

	// encode value into required argument
	err := client.QuerySingle(ctx, `
		SELECT { val := <datetime>$0 }`,
		&result,
		newValue([]byte{0x00, 0x02, 0x2b, 0x35, 0x9b, 0xc4, 0x10, 0x00}),
	)
	assert.NoError(t, err)
	assert.Equal(
		t,
		types.NewOptionalDateTime(
			time.Date(2019, 5, 6, 12, 0, 0, 0, time.UTC),
		),
		result.Val,
	)

	// encode value into optional argument
	err = client.QuerySingle(ctx, `
		SELECT { val := <OPTIONAL datetime>$0 }`,
		&result,
		newValue([]byte{0x00, 0x02, 0x2b, 0x35, 0x9b, 0xc4, 0x10, 0x00}),
	)
	assert.NoError(t, err)
	assert.Equal(
		t,
		types.NewOptionalDateTime(
			time.Date(2019, 5, 6, 12, 0, 0, 0, time.UTC),
		),
		result.Val,
	)

	// encode missing value into optional argument
	err = client.QuerySingle(ctx, `
		SELECT { val := <OPTIONAL datetime>$0 }`,
		&result,
		CustomOptionalDateTime{},
	)
	assert.NoError(t, err)
	assert.Equal(t, types.OptionalDateTime{}, result.Val)

	// encode missing value into required argument
	err = client.QuerySingle(ctx, `
		SELECT { val := <datetime>$0 }`,
		&result,
		CustomOptionalDateTime{},
	)
	assert.EqualError(t, err, "gel.InvalidArgumentError: "+
		"cannot encode gel.CustomOptionalDateTime at args[0] "+
		"because its value is missing")

	// encode wrong number of bytes with required argument
	err = client.QuerySingle(ctx, `
		SELECT { val := <datetime>$0 }`,
		&result,
		newValue([]byte{0x01}),
	)
	assert.EqualError(t, err, "gel.InvalidArgumentError: "+
		"wrong number of bytes encoded by gel.CustomOptionalDateTime "+
		"at args[0] expected 8, got 1")

	// encode wrong number of bytes with optional argument
	err = client.QuerySingle(ctx, `
		SELECT { val := <OPTIONAL datetime>$0 }`,
		&result,
		newValue([]byte{0x01}),
	)
	assert.EqualError(t, err, "gel.InvalidArgumentError: "+
		"wrong number of bytes encoded by gel.CustomOptionalDateTime "+
		"at args[0] expected 8, got 1")
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
		Encoded   string   `gel:"encoded"`
		Decoded   *big.Int `gel:"decoded"`
		RoundTrip *big.Int `gel:"round_trip"`
		IsEqual   bool     `gel:"is_equal"`
		String    string   `gel:"string"`
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
			err := client.QuerySingle(ctx, query, &result, i, s)
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

// The algorithm for decoding bigint is a summation.  If the result memory is
// not cleared before decoding the decoded value will be added to the existing
// value in memory.
func TestReuseBigIntValue(t *testing.T) {
	ctx := context.Background()
	expected := big.NewInt(123)

	var result *big.Int
	err := client.QuerySingle(ctx, "SELECT 123n", &result)
	require.NoError(t, err)
	assert.Equal(t,
		0, expected.Cmp(result),
		"%v != %v", expected.String(), result.String(),
	)

	err = client.QuerySingle(ctx, "SELECT 123n", &result)
	require.NoError(t, err)
	assert.Equal(t,
		0, expected.Cmp(result),
		"%v != %v", expected.String(), result.String(),
	)

	err = client.QuerySingle(ctx, "SELECT 123n", &result)
	require.NoError(t, err)
	assert.Equal(t,
		0, expected.Cmp(result),
		"%v != %v", expected.String(), result.String(),
	)

	var optional types.OptionalBigInt
	err = client.QuerySingle(ctx, "SELECT 123n", &optional)
	require.NoError(t, err)
	v, ok := optional.Get()
	require.True(t, ok)
	assert.Equal(t,
		0, expected.Cmp(v),
		"%v != %v", expected.String(), result.String(),
	)

	err = client.QuerySingle(ctx, "SELECT 123n", &optional)
	require.NoError(t, err)
	v, ok = optional.Get()
	require.True(t, ok)
	assert.Equal(t,
		0, expected.Cmp(v),
		"%v != %v", expected.String(), result.String(),
	)
}

type CustomBigInt struct {
	data []byte
}

func (m CustomBigInt) MarshalEdgeDBBigInt() ([]byte, error) {
	data := make([]byte, len(m.data))
	copy(data, m.data)
	return data, nil
}

func (m *CustomBigInt) UnmarshalEdgeDBBigInt(data []byte) error {
	m.data = make([]byte, len(data))
	copy(m.data, data)
	return nil
}

func TestReceiveBigIntUnmarshaler(t *testing.T) {
	ctx := context.Background()
	var result struct {
		Val CustomBigInt `gel:"val"`
	}

	// Decode value
	query := `SELECT { val := <bigint>-15000n }`
	err := client.QuerySingle(ctx, query, &result)
	assert.NoError(t, err)
	assert.Equal(t,
		[]byte{
			0x00, 0x02, // ndigits
			0x00, 0x01, // weight
			0x40, 0x00, // sign
			0x00, 0x00, // reserved
			0x00, 0x01, 0x13, 0x88, // digits
		},
		result.Val.data,
	)

	// Decode missing value
	err = client.QuerySingle(ctx, `
		SELECT { val := <OPTIONAL bigint>$0 }`,
		&result,
		types.OptionalBigInt{},
	)
	assert.EqualError(t, err, "gel.InvalidArgumentError: "+
		"the \"out\" argument does not match query schema: "+
		"expected gel.CustomBigInt at "+
		"struct { Val gel.CustomBigInt \"gel:\\\"val\\\"\" }.val "+
		"to be OptionalUnmarshaler interface "+
		"because the field is not required")
}

func TestSendBigIntMarshaler(t *testing.T) {
	ctx := context.Background()
	var result struct {
		Val types.OptionalBigInt `gel:"val"`
	}

	// encode value into required argument
	err := client.QuerySingle(ctx, `
		SELECT { val := <bigint>$0 }`,
		&result,
		CustomBigInt{data: []byte{
			0x00, 0x02, // ndigits
			0x00, 0x01, // weight
			0x40, 0x00, // sign
			0x00, 0x00, // reserved
			0x00, 0x01, 0x13, 0x88, // digits
		}},
	)
	assert.NoError(t, err)
	assert.Equal(t, types.NewOptionalBigInt(big.NewInt(-15000)), result.Val)

	// encode value into optional argument
	err = client.QuerySingle(ctx, `
		SELECT { val := <OPTIONAL bigint>$0 }`,
		&result,
		CustomBigInt{data: []byte{
			0x00, 0x02, // ndigits
			0x00, 0x01, // weight
			0x40, 0x00, // sign
			0x00, 0x00, // reserved
			0x00, 0x01, 0x13, 0x88, // digits
		}},
	)
	assert.NoError(t, err)
	assert.Equal(t, types.NewOptionalBigInt(big.NewInt(-15000)), result.Val)

	// encode wrong number of bytes
	err = client.QuerySingle(ctx, `
		SELECT { val := <bigint>$0 }`,
		&result,
		CustomBigInt{data: []byte{0x01}},
	)
	assert.EqualError(t, err, "gel.InvalidArgumentError: "+
		"wrong number of bytes encoded by gel.CustomBigInt "+
		"at args[0] expected at least 8, got 1")
}

type CustomOptionalBigInt struct {
	data  []byte
	isSet bool
}

func (m CustomOptionalBigInt) MarshalEdgeDBBigInt() ([]byte, error) {
	if !m.isSet {
		return nil, fmt.Errorf("%T is not set", m)
	}
	data := make([]byte, len(m.data))
	copy(data, m.data)
	return data, nil
}

func (m *CustomOptionalBigInt) UnmarshalEdgeDBBigInt(data []byte) error {
	m.isSet = true
	m.data = make([]byte, len(data))
	copy(m.data, data)
	return nil
}

func (m *CustomOptionalBigInt) SetMissing(missing bool) {
	m.isSet = !missing
	m.data = nil
}

func (m CustomOptionalBigInt) Missing() bool { return !m.isSet }

func TestReceiveOptionalBigIntUnmarshaler(t *testing.T) {
	ddl := `CREATE TYPE Sample { CREATE PROPERTY val -> bigint; };`
	inRolledBackTx(t, ddl, func(ctx context.Context, tx *Tx) {
		var result struct {
			Val CustomOptionalBigInt `gel:"val"`
		}

		// Decode value
		err := tx.QuerySingle(ctx,
			`SELECT { val := <bigint>-15000n }`,
			&result,
		)
		assert.NoError(t, err)
		assert.Equal(t,
			[]byte{
				0x00, 0x02, // ndigits
				0x00, 0x01, // weight
				0x40, 0x00, // sign
				0x00, 0x00, // reserved
				0x00, 0x01, 0x13, 0x88, // digits
			},
			result.Val.data,
		)

		// Decode missing value
		query := `WITH inserted := (INSERT Sample) SELECT inserted { val }`
		err = tx.QuerySingle(ctx, query, &result)
		assert.NoError(t, err)
		assert.Equal(t, CustomOptionalBigInt{}, result.Val)
	})
}

func TestSendOptionalBigIntMarshaler(t *testing.T) {
	ctx := context.Background()
	var result struct {
		Val types.OptionalBigInt `gel:"val"`
	}

	newValue := func(data []byte) CustomOptionalBigInt {
		return CustomOptionalBigInt{isSet: true, data: data}
	}

	// encode value into required argument
	err := client.QuerySingle(ctx, `
		SELECT { val := <bigint>$0 }`,
		&result,
		newValue([]byte{
			0x00, 0x02, // ndigits
			0x00, 0x01, // weight
			0x40, 0x00, // sign
			0x00, 0x00, // reserved
			0x00, 0x01, 0x13, 0x88, // digits
		}),
	)
	assert.NoError(t, err)
	assert.Equal(t, types.NewOptionalBigInt(big.NewInt(-15000)), result.Val)

	// encode value into optional argument
	err = client.QuerySingle(ctx, `
		SELECT { val := <OPTIONAL bigint>$0 }`,
		&result,
		newValue([]byte{
			0x00, 0x02, // ndigits
			0x00, 0x01, // weight
			0x40, 0x00, // sign
			0x00, 0x00, // reserved
			0x00, 0x01, 0x13, 0x88, // digits
		}),
	)
	assert.NoError(t, err)
	assert.Equal(t, types.NewOptionalBigInt(big.NewInt(-15000)), result.Val)

	// encode missing value into optional argument
	err = client.QuerySingle(ctx, `
		SELECT { val := <OPTIONAL bigint>$0 }`,
		&result,
		CustomOptionalBigInt{},
	)
	assert.NoError(t, err)
	assert.Equal(t, types.OptionalBigInt{}, result.Val)

	// encode missing value into required argument
	err = client.QuerySingle(ctx, `
		SELECT { val := <bigint>$0 }`,
		&result,
		CustomOptionalBigInt{},
	)
	assert.EqualError(t, err, "gel.InvalidArgumentError: "+
		"cannot encode gel.CustomOptionalBigInt at args[0] "+
		"because its value is missing")

	// encode wrong number of bytes with required argument
	err = client.QuerySingle(ctx, `
		SELECT { val := <bigint>$0 }`,
		&result,
		newValue([]byte{0x01}),
	)
	assert.EqualError(t, err, "gel.InvalidArgumentError: "+
		"wrong number of bytes encoded by gel.CustomOptionalBigInt "+
		"at args[0] expected at least 8, got 1")

	// encode wrong number of bytes with optional argument
	err = client.QuerySingle(ctx, `
		SELECT { val := <OPTIONAL bigint>$0 }`,
		&result,
		newValue([]byte{0x01}),
	)
	assert.EqualError(t, err, "gel.InvalidArgumentError: "+
		"wrong number of bytes encoded by gel.CustomOptionalBigInt "+
		"at args[0] expected at least 8, got 1")
}

type CustomDecimal struct {
	data []byte
}

func (m CustomDecimal) MarshalEdgeDBDecimal() ([]byte, error) {
	data := make([]byte, len(m.data))
	copy(data, m.data)
	return data, nil
}

func (m *CustomDecimal) UnmarshalEdgeDBDecimal(data []byte) error {
	m.data = make([]byte, len(data))
	copy(m.data, data)
	return nil
}

func TestReceiveDecimalUnmarshaler(t *testing.T) {
	ctx := context.Background()
	var result struct {
		Val CustomDecimal `gel:"val"`
	}

	// Decode value
	err := client.QuerySingle(ctx, `
		SELECT { val := <decimal>-15000.6250000n }`,
		&result,
	)
	assert.NoError(t, err)
	assert.Equal(t,
		[]byte{
			0x00, 0x03, // ndigits
			0x00, 0x01, // weight
			0x40, 0x00, // sign
			0x00, 0x07, // dscale
			0x00, 0x01, 0x13, 0x88, 0x18, 0x6a, // digits
		},
		result.Val.data,
	)

	// Decode missing value
	err = client.QuerySingle(ctx, `
		SELECT { val := <OPTIONAL decimal>$0 }`,
		&result,
		CustomOptionalDecimal{},
	)
	assert.EqualError(t, err, "gel.InvalidArgumentError: "+
		"the \"out\" argument does not match query schema: "+
		"expected gel.CustomDecimal at "+
		"struct { Val gel.CustomDecimal \"gel:\\\"val\\\"\" }.val "+
		"to be OptionalUnmarshaler interface "+
		"because the field is not required")
}

func TestSendDecimalMarshaler(t *testing.T) {
	ctx := context.Background()
	var result struct {
		Val CustomOptionalDecimal `gel:"val"`
	}

	// encode value into required argument
	err := client.QuerySingle(ctx, `
		SELECT { val := <decimal>$0 }`,
		&result,
		CustomDecimal{data: []byte{
			0x00, 0x03, // ndigits
			0x00, 0x01, // weight
			0x40, 0x00, // sign
			0x00, 0x07, // dscale
			0x00, 0x01, 0x13, 0x88, 0x18, 0x6a, // digits
		}},
	)
	assert.NoError(t, err)
	assert.Equal(t,
		CustomOptionalDecimal{isSet: true, data: []byte{
			0x00, 0x03, // ndigits
			0x00, 0x01, // weight
			0x40, 0x00, // sign
			0x00, 0x07, // dscale
			0x00, 0x01, 0x13, 0x88, 0x18, 0x6a, // digits
		}},
		result.Val,
	)

	// encode value into optional argument
	err = client.QuerySingle(ctx, `
		SELECT { val := <OPTIONAL decimal>$0 }`,
		&result,
		CustomDecimal{data: []byte{
			0x00, 0x03, // ndigits
			0x00, 0x01, // weight
			0x40, 0x00, // sign
			0x00, 0x07, // dscale
			0x00, 0x01, 0x13, 0x88, 0x18, 0x6a, // digits
		}},
	)
	assert.NoError(t, err)
	assert.Equal(t,
		CustomOptionalDecimal{isSet: true, data: []byte{
			0x00, 0x03, // ndigits
			0x00, 0x01, // weight
			0x40, 0x00, // sign
			0x00, 0x07, // dscale
			0x00, 0x01, 0x13, 0x88, 0x18, 0x6a, // digits
		}},
		result.Val,
	)

	// encode wrong number of bytes
	err = client.QuerySingle(ctx, `
		SELECT { val := <decimal>$0 }`,
		&result,
		CustomDecimal{data: []byte{0x01}},
	)
	assert.EqualError(t, err, "gel.InvalidArgumentError: "+
		"wrong number of bytes encoded by gel.CustomDecimal "+
		"at args[0] expected at least 8, got 1")
}

type CustomOptionalDecimal struct {
	data  []byte
	isSet bool
}

func (m CustomOptionalDecimal) MarshalEdgeDBDecimal() ([]byte, error) {
	if !m.isSet {
		return nil, fmt.Errorf("%T is not set", m)
	}
	data := make([]byte, len(m.data))
	copy(data, m.data)
	return data, nil
}

func (m *CustomOptionalDecimal) UnmarshalEdgeDBDecimal(data []byte) error {
	m.isSet = true
	m.data = make([]byte, len(data))
	copy(m.data, data)
	return nil
}

func (m *CustomOptionalDecimal) SetMissing(missing bool) {
	m.isSet = !missing
	m.data = nil
}

func (m CustomOptionalDecimal) Missing() bool { return !m.isSet }

func TestReceiveOptionalDecimalUnmarshaler(t *testing.T) {
	ddl := `CREATE TYPE Sample { CREATE PROPERTY val -> decimal; };`
	inRolledBackTx(t, ddl, func(ctx context.Context, tx *Tx) {
		var result struct {
			Val CustomOptionalDecimal `gel:"val"`
		}

		// Decode value
		err := tx.QuerySingle(ctx, `
			SELECT { val := <decimal>-15000.6250000n }`,
			&result,
		)
		assert.NoError(t, err)
		assert.Equal(t,
			[]byte{
				0x00, 0x03, // ndigits
				0x00, 0x01, // weight
				0x40, 0x00, // sign
				0x00, 0x07, // dscale
				0x00, 0x01, 0x13, 0x88, 0x18, 0x6a, // digits
			},
			result.Val.data,
		)

		// Decode missing value
		query := `WITH inserted := (INSERT Sample) SELECT inserted { val }`
		err = tx.QuerySingle(ctx, query, &result)
		assert.NoError(t, err)
		assert.Equal(t, CustomOptionalDecimal{}, result.Val)
	})
}

func TestSendOptionalDecimalMarshaler(t *testing.T) {
	ctx := context.Background()
	var result struct {
		Val CustomOptionalDecimal `gel:"val"`
	}

	// encode value into required argument
	err := client.QuerySingle(ctx, `
		SELECT { val := <decimal>$0 }`,
		&result,
		CustomDecimal{data: []byte{
			0x00, 0x03, // ndigits
			0x00, 0x01, // weight
			0x40, 0x00, // sign
			0x00, 0x07, // dscale
			0x00, 0x01, 0x13, 0x88, 0x18, 0x6a, // digits
		}},
	)
	assert.NoError(t, err)
	assert.Equal(t,
		CustomOptionalDecimal{isSet: true, data: []byte{
			0x00, 0x03, // ndigits
			0x00, 0x01, // weight
			0x40, 0x00, // sign
			0x00, 0x07, // dscale
			0x00, 0x01, 0x13, 0x88, 0x18, 0x6a, // digits
		}},
		result.Val,
	)

	// encode value into optional argument
	err = client.QuerySingle(ctx, `
		SELECT { val := <OPTIONAL decimal>$0 }`,
		&result,
		CustomDecimal{data: []byte{
			0x00, 0x03, // ndigits
			0x00, 0x01, // weight
			0x40, 0x00, // sign
			0x00, 0x07, // dscale
			0x00, 0x01, 0x13, 0x88, 0x18, 0x6a, // digits
		}},
	)
	assert.NoError(t, err)
	assert.Equal(t,
		CustomOptionalDecimal{isSet: true, data: []byte{
			0x00, 0x03, // ndigits
			0x00, 0x01, // weight
			0x40, 0x00, // sign
			0x00, 0x07, // dscale
			0x00, 0x01, 0x13, 0x88, 0x18, 0x6a, // digits
		}},
		result.Val,
	)

	// encode missing value into optional argument
	err = client.QuerySingle(ctx, `
		SELECT { val := <OPTIONAL decimal>$0 }`,
		&result,
		CustomOptionalDecimal{},
	)
	assert.NoError(t, err)
	assert.Equal(t, CustomOptionalDecimal{}, result.Val)

	// encode missing value into required argument
	err = client.QuerySingle(ctx, `
		SELECT { val := <decimal>$0 }`,
		&result,
		CustomOptionalDecimal{},
	)
	assert.EqualError(t, err, "gel.InvalidArgumentError: "+
		"cannot encode gel.CustomOptionalDecimal at args[0] "+
		"because its value is missing")

	// encode wrong number of bytes with required argument
	err = client.QuerySingle(ctx, `
		SELECT { val := <decimal>$0 }`,
		&result,
		CustomOptionalDecimal{isSet: true, data: []byte{0x01}},
	)
	assert.EqualError(t, err, "gel.InvalidArgumentError: "+
		"wrong number of bytes encoded by gel.CustomOptionalDecimal "+
		"at args[0] expected at least 8, got 1")

	// encode wrong number of bytes with optional argument
	err = client.QuerySingle(ctx, `
		SELECT { val := <OPTIONAL decimal>$0 }`,
		&result,
		CustomOptionalDecimal{isSet: true, data: []byte{0x01}},
	)
	assert.EqualError(t, err, "gel.InvalidArgumentError: "+
		"wrong number of bytes encoded by gel.CustomOptionalDecimal "+
		"at args[0] expected at least 8, got 1")
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
		Encoded   string     `gel:"encoded"`
		Decoded   types.UUID `gel:"decoded"`
		RoundTrip types.UUID `gel:"round_trip"`
		IsEqual   bool       `gel:"is_equal"`
		String    string     `gel:"string"`
	}

	samples := []string{
		"759637d8-6635-11e9-b9d4-098002d459d5",
		"00000000-0000-0000-0000-000000000000",
		"ffffffff-ffff-ffff-ffff-ffffffffffff",
	}

	for i := 0; i < 1000; i++ {
		var id types.UUID
		binary.BigEndian.PutUint64(id[:8], rand.Uint64())
		binary.BigEndian.PutUint64(id[8:], rand.Uint64())
		samples = append(samples, id.String())
	}

	for _, s := range samples {
		t.Run(s, func(t *testing.T) {
			var id types.UUID
			err := id.UnmarshalText([]byte(s))
			require.NoError(t, err)

			var result Result
			err = client.QuerySingle(ctx, query, &result, id, s)
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
	data []byte
}

func (m CustomUUID) MarshalEdgeDBUUID() ([]byte, error) {
	data := make([]byte, len(m.data))
	copy(data, m.data)
	return data, nil
}

func (m *CustomUUID) UnmarshalEdgeDBUUID(data []byte) error {
	m.data = make([]byte, len(data))
	copy(m.data, data)
	return nil
}

func TestReceiveUUIDUnmarshaler(t *testing.T) {
	ctx := context.Background()
	var result struct {
		Val CustomUUID `gel:"val"`
	}

	// Decode value
	err := client.QuerySingle(ctx, `
		SELECT { val := <uuid>'b9545c35-1fe7-485f-a6ea-f8ead251abd3' }`,
		&result,
	)
	assert.NoError(t, err)
	assert.Equal(t,
		[]byte{
			0xb9, 0x54, 0x5c, 0x35, 0x1f, 0xe7, 0x48, 0x5f,
			0xa6, 0xea, 0xf8, 0xea, 0xd2, 0x51, 0xab, 0xd3,
		},
		result.Val.data,
	)

	// Decode missing value
	err = client.QuerySingle(ctx, `
		SELECT { val := <OPTIONAL uuid>$0 }`,
		&result,
		types.OptionalUUID{},
	)
	assert.EqualError(t, err, "gel.InvalidArgumentError: "+
		"the \"out\" argument does not match query schema: "+
		"expected gel.CustomUUID at "+
		"struct { Val gel.CustomUUID \"gel:\\\"val\\\"\" }.val "+
		"to be OptionalUnmarshaler interface "+
		"because the field is not required")
}

func TestSendUUIDMarshaler(t *testing.T) {
	ctx := context.Background()
	var result struct {
		Val types.OptionalUUID `gel:"val"`
	}

	// encode value into required argument
	err := client.QuerySingle(ctx, `
		SELECT { val := <uuid>$0 }`,
		&result,
		CustomUUID{data: []byte{
			0xb9, 0x54, 0x5c, 0x35, 0x1f, 0xe7, 0x48, 0x5f,
			0xa6, 0xea, 0xf8, 0xea, 0xd2, 0x51, 0xab, 0xd3,
		}},
	)
	assert.NoError(t, err)
	assert.Equal(t,
		types.NewOptionalUUID(types.UUID{
			0xb9, 0x54, 0x5c, 0x35, 0x1f, 0xe7, 0x48, 0x5f,
			0xa6, 0xea, 0xf8, 0xea, 0xd2, 0x51, 0xab, 0xd3,
		}),
		result.Val,
	)

	// encode value into optional argument
	err = client.QuerySingle(ctx, `
		SELECT { val := <OPTIONAL uuid>$0 }`,
		&result,
		CustomUUID{data: []byte{
			0xb9, 0x54, 0x5c, 0x35, 0x1f, 0xe7, 0x48, 0x5f,
			0xa6, 0xea, 0xf8, 0xea, 0xd2, 0x51, 0xab, 0xd3,
		}},
	)
	assert.NoError(t, err)
	assert.Equal(t,
		types.NewOptionalUUID(types.UUID{
			0xb9, 0x54, 0x5c, 0x35, 0x1f, 0xe7, 0x48, 0x5f,
			0xa6, 0xea, 0xf8, 0xea, 0xd2, 0x51, 0xab, 0xd3,
		}),
		result.Val,
	)

	// encode wrong number of bytes
	err = client.QuerySingle(ctx, `
		SELECT { val := <uuid>$0 }`,
		&result,
		CustomUUID{data: []byte{0x01}},
	)
	assert.EqualError(t, err, "gel.InvalidArgumentError: "+
		"wrong number of bytes encoded by gel.CustomUUID "+
		"at args[0] expected 16, got 1")
}

type CustomOptionalUUID struct {
	data  []byte
	isSet bool
}

func (m CustomOptionalUUID) MarshalEdgeDBUUID() ([]byte, error) {
	if !m.isSet {
		return nil, fmt.Errorf("%T is not set", m)
	}
	data := make([]byte, len(m.data))
	copy(data, m.data)
	return data, nil
}

func (m *CustomOptionalUUID) UnmarshalEdgeDBUUID(data []byte) error {
	m.isSet = true
	m.data = make([]byte, len(data))
	copy(m.data, data)
	return nil
}

func (m *CustomOptionalUUID) SetMissing(missing bool) {
	m.isSet = !missing
	m.data = nil
}

func (m CustomOptionalUUID) Missing() bool { return !m.isSet }

func TestReceiveOptionalUUIDUnmarshaler(t *testing.T) {
	ddl := `CREATE TYPE Sample { CREATE PROPERTY val -> uuid; };`
	inRolledBackTx(t, ddl, func(ctx context.Context, tx *Tx) {
		var result struct {
			Val CustomOptionalUUID `gel:"val"`
		}

		// Decode value
		err := tx.QuerySingle(ctx, `
			SELECT { val := <uuid>'b9545c35-1fe7-485f-a6ea-f8ead251abd3' }`,
			&result,
		)
		assert.NoError(t, err)
		assert.Equal(t,
			[]byte{
				0xb9, 0x54, 0x5c, 0x35, 0x1f, 0xe7, 0x48, 0x5f,
				0xa6, 0xea, 0xf8, 0xea, 0xd2, 0x51, 0xab, 0xd3,
			},
			result.Val.data,
		)

		// Decode missing value
		query := `WITH inserted := (INSERT Sample) SELECT inserted { val }`
		err = tx.QuerySingle(ctx, query, &result)
		assert.NoError(t, err)
		assert.Equal(t, CustomOptionalUUID{}, result.Val)
	})
}

func TestSendOptionalUUIDMarshaler(t *testing.T) {
	ctx := context.Background()
	var result struct {
		Val types.OptionalUUID `gel:"val"`
	}

	newValue := func(data []byte) CustomOptionalUUID {
		return CustomOptionalUUID{isSet: true, data: data}
	}

	// encode value into required argument
	err := client.QuerySingle(ctx, `
		SELECT { val := <uuid>$0 }`,
		&result,
		newValue([]byte{
			0xb9, 0x54, 0x5c, 0x35, 0x1f, 0xe7, 0x48, 0x5f,
			0xa6, 0xea, 0xf8, 0xea, 0xd2, 0x51, 0xab, 0xd3,
		}),
	)
	assert.NoError(t, err)
	assert.Equal(t,
		types.NewOptionalUUID(types.UUID{
			0xb9, 0x54, 0x5c, 0x35, 0x1f, 0xe7, 0x48, 0x5f,
			0xa6, 0xea, 0xf8, 0xea, 0xd2, 0x51, 0xab, 0xd3,
		}),
		result.Val)

	// encode value into optional argument
	err = client.QuerySingle(ctx, `
		SELECT { val := <OPTIONAL uuid>$0 }`,
		&result,
		newValue([]byte{
			0xb9, 0x54, 0x5c, 0x35, 0x1f, 0xe7, 0x48, 0x5f,
			0xa6, 0xea, 0xf8, 0xea, 0xd2, 0x51, 0xab, 0xd3,
		}),
	)
	assert.NoError(t, err)
	assert.Equal(t,
		types.NewOptionalUUID(types.UUID{
			0xb9, 0x54, 0x5c, 0x35, 0x1f, 0xe7, 0x48, 0x5f,
			0xa6, 0xea, 0xf8, 0xea, 0xd2, 0x51, 0xab, 0xd3,
		}),
		result.Val,
	)

	// encode missing value into optional argument
	err = client.QuerySingle(ctx, `
		SELECT { val := <OPTIONAL uuid>$0 }`,
		&result,
		CustomOptionalUUID{},
	)
	assert.NoError(t, err)
	assert.Equal(t, types.OptionalUUID{}, result.Val)

	// encode missing value into required argument
	err = client.QuerySingle(ctx, `
		SELECT { val := <uuid>$0 }`,
		&result,
		CustomOptionalUUID{},
	)
	assert.EqualError(t, err, "gel.InvalidArgumentError: "+
		"cannot encode gel.CustomOptionalUUID at args[0] "+
		"because its value is missing")

	// encode wrong number of bytes with required argument
	err = client.QuerySingle(ctx, `
		SELECT { val := <uuid>$0 }`,
		&result,
		newValue([]byte{0x01}),
	)
	assert.EqualError(t, err, "gel.InvalidArgumentError: "+
		"wrong number of bytes encoded by gel.CustomOptionalUUID "+
		"at args[0] expected 16, got 1")

	// encode wrong number of bytes with optional argument
	err = client.QuerySingle(ctx, `
		SELECT { val := <OPTIONAL uuid>$0 }`,
		&result,
		newValue([]byte{0x01}),
	)
	assert.EqualError(t, err, "gel.InvalidArgumentError: "+
		"wrong number of bytes encoded by gel.CustomOptionalUUID "+
		"at args[0] expected 16, got 1")
}

func TestSendAndReceiveCustomScalars(t *testing.T) {
	query := `
		WITH
			x := (
				WITH
					i := enumerate(array_unpack(<array<CustomInt64>>$0)),
					s := enumerate(array_unpack(<array<str>>$1)),
				SELECT (
					i := i.1,
					s := s.1,
				)
				FILTER i.0 = s.0
			)
		SELECT (
			encoded := <str>x.i,
			decoded := <CustomInt64>x.s,
			round_trip := x.i,
			is_equal := <CustomInt64>x.s = x.i,
		)
	`

	type Result struct {
		Encoded   string `gel:"encoded"`
		Decoded   int64  `gel:"decoded"`
		RoundTrip int64  `gel:"round_trip"`
		IsEqual   bool   `gel:"is_equal"`
	}

	samples := []int64{0, 1, 9223372036854775807, -9223372036854775808}
	for i := 0; i < 1000; i++ {
		samples = append(samples, int64(rand.Uint64()))
	}

	strings := make([]string, len(samples))
	for i, n := range samples {
		strings[i] = fmt.Sprint(n)
	}

	ddl := `CREATE SCALAR TYPE CustomInt64 EXTENDING int64;`
	inRolledBackTx(t, ddl, func(c context.Context, tx *Tx) {
		var results []Result
		err := tx.Query(c, query, &results, samples, strings)
		require.NoError(t, err)
		require.Equal(t, len(samples), len(results), "unexpected result count")

		for i, s := range strings {
			t.Run(s, func(t *testing.T) {
				n := samples[i]
				r := results[i]

				assert.True(t, r.IsEqual, "equality check failed")
				assert.Equal(t, s, r.Encoded, "encoding failed")
				assert.Equal(t, n, r.Decoded, "decoding failed")
				assert.Equal(t, n, r.RoundTrip, "round trip failed")
			})
		}
	})
}

func TestReceiveCustomScalarUnmarshaler(t *testing.T) {
	ddl := `CREATE SCALAR TYPE CustomInt64 EXTENDING int64;`
	inRolledBackTx(t, ddl, func(ctx context.Context, tx *Tx) {
		var result struct {
			Val CustomInt64 `gel:"val"`
		}

		// Decode value
		err := tx.QuerySingle(ctx, `
			SELECT { val := <CustomInt64>123_456_789_987_654_321 }`,
			&result,
		)
		assert.NoError(t, err)
		assert.Equal(t,
			[]byte{0x01, 0xb6, 0x9b, 0x4b, 0xe0, 0x52, 0xfa, 0xb1},
			result.Val.data,
		)

		// Decode missing value
		err = tx.QuerySingle(ctx, `
			SELECT { val := <OPTIONAL CustomInt64>$0 }`,
			&result,
			types.OptionalInt64{},
		)
		assert.EqualError(t, err, "gel.InvalidArgumentError: "+
			"the \"out\" argument does not match query schema: "+
			"expected gel.CustomInt64 at "+
			"struct { Val gel.CustomInt64 \"gel:\\\"val\\\"\" }.val "+
			"to be OptionalUnmarshaler interface "+
			"because the field is not required")
	})
}

func TestSendCustomScalarMarshaler(t *testing.T) {
	ddl := `CREATE SCALAR TYPE CustomInt64 EXTENDING int64;`
	inRolledBackTx(t, ddl, func(ctx context.Context, tx *Tx) {
		var result struct {
			Val types.OptionalInt64 `gel:"val"`
		}

		// encode value into required argument
		err := tx.QuerySingle(ctx, `
			SELECT { val := <CustomInt64>$0 }`,
			&result,
			CustomInt64{
				data: []byte{0x01, 0xb6, 0x9b, 0x4b, 0xe0, 0x52, 0xfa, 0xb1},
			},
		)
		assert.NoError(t, err)
		assert.Equal(
			t,
			types.NewOptionalInt64(123_456_789_987_654_321),
			result.Val,
		)

		// encode value into optional argument
		err = tx.QuerySingle(ctx, `
			SELECT { val := <OPTIONAL CustomInt64>$0 }`,
			&result,
			CustomInt64{
				data: []byte{0x01, 0xb6, 0x9b, 0x4b, 0xe0, 0x52, 0xfa, 0xb1},
			},
		)
		assert.NoError(t, err)
		assert.Equal(
			t,
			types.NewOptionalInt64(123_456_789_987_654_321),
			result.Val,
		)

		// encode wrong number of bytes
		err = tx.QuerySingle(ctx, `
			SELECT { val := <CustomInt64>$0 }`,
			&result,
			CustomInt64{data: []byte{0x01}},
		)
		assert.EqualError(t, err, "gel.InvalidArgumentError: "+
			"wrong number of bytes encoded by gel.CustomInt64 "+
			"at args[0] expected 8, got 1")
	})
}

func TestReceiveOptionalCustomScalarUnmarshaler(t *testing.T) {
	ddl := `
		CREATE SCALAR TYPE CustomInt64 EXTENDING int64;
		CREATE TYPE Sample {
			CREATE PROPERTY val -> CustomInt64;
		};
	`
	inRolledBackTx(t, ddl, func(ctx context.Context, tx *Tx) {
		var result struct {
			Val CustomOptionalInt64 `gel:"val"`
		}

		// Decode value
		err := tx.QuerySingle(ctx, `
			SELECT { val := 123_456_789_987_654_321 }`,
			&result,
		)
		assert.NoError(t, err)
		assert.Equal(t,
			[]byte{0x01, 0xb6, 0x9b, 0x4b, 0xe0, 0x52, 0xfa, 0xb1},
			result.Val.data,
		)

		// Decode missing value
		query := `WITH inserted := (INSERT Sample) SELECT inserted { val }`
		err = tx.QuerySingle(ctx, query, &result)
		assert.NoError(t, err)
		assert.Equal(t, CustomOptionalInt64{}, result.Val)
	})
}

func TestSendOptionalCustomScalarMarshaler(t *testing.T) {
	ddl := `CREATE SCALAR TYPE CustomInt64 EXTENDING int64;`
	inRolledBackTx(t, ddl, func(ctx context.Context, tx *Tx) {
		var result struct {
			Val types.OptionalInt64 `gel:"val"`
		}

		newValue := func(data []byte) CustomOptionalInt64 {
			return CustomOptionalInt64{isSet: true, data: data}
		}

		// encode value into required argument
		err := tx.QuerySingle(ctx, `
			SELECT { val := <CustomInt64>$0 }`,
			&result,
			newValue([]byte{0x01, 0xb6, 0x9b, 0x4b, 0xe0, 0x52, 0xfa, 0xb1}),
		)
		assert.NoError(t, err)
		assert.Equal(
			t,
			types.NewOptionalInt64(123_456_789_987_654_321),
			result.Val,
		)

		// encode value into optional argument
		err = tx.QuerySingle(ctx, `
			SELECT { val := <OPTIONAL CustomInt64>$0 }`,
			&result,
			newValue([]byte{0x01, 0xb6, 0x9b, 0x4b, 0xe0, 0x52, 0xfa, 0xb1}),
		)
		assert.NoError(t, err)
		assert.Equal(
			t,
			types.NewOptionalInt64(123_456_789_987_654_321),
			result.Val,
		)

		// encode missing value into optional argument
		err = tx.QuerySingle(ctx, `
			SELECT { val := <OPTIONAL CustomInt64>$0 }`,
			&result,
			CustomOptionalInt64{},
		)
		assert.NoError(t, err)
		assert.Equal(t, types.OptionalInt64{}, result.Val)

		// encode missing value into required argument
		err = tx.QuerySingle(ctx, `
			SELECT { val := <CustomInt64>$0 }`,
			&result,
			CustomOptionalInt64{},
		)
		assert.EqualError(t, err, "gel.InvalidArgumentError: "+
			"cannot encode gel.CustomOptionalInt64 at args[0] "+
			"because its value is missing")

		// encode wrong number of bytes with required argument
		err = tx.QuerySingle(ctx, `
			SELECT { val := <CustomInt64>$0 }`,
			&result,
			newValue([]byte{0x01}),
		)
		assert.EqualError(t, err, "gel.InvalidArgumentError: "+
			"wrong number of bytes encoded by gel.CustomOptionalInt64 "+
			"at args[0] expected 8, got 1")

		// encode wrong number of bytes with optional argument
		err = tx.QuerySingle(ctx, `
			SELECT { val := <OPTIONAL CustomInt64>$0 }`,
			&result,
			newValue([]byte{0x01}),
		)
		assert.EqualError(t, err, "gel.InvalidArgumentError: "+
			"wrong number of bytes encoded by gel.CustomOptionalInt64 "+
			"at args[0] expected 8, got 1")
	})
}

func TestDecodeDeeplyNestedTuple(t *testing.T) {
	ctx := context.Background()
	query := "SELECT ([(1, 2), (3, 4)], (5, (6, 7)))"

	type Tuple struct {
		first  int64 `gel:"0"`
		second int64 `gel:"1"`
	}

	type OtherTuple struct {
		first  int64 `gel:"0"`
		second Tuple `gel:"1"`
	}

	type ParentTuple struct {
		first  []Tuple    `gel:"0"`
		second OtherTuple `gel:"1"`
	}

	var result ParentTuple
	err := client.QuerySingle(ctx, query, &result)
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
		ID   types.UUID `gel:"id"`
		Kind string     `gel:"kind"`
		Num  int64      `gel:"num"`
		Foo  int64      `gel:"foo"`
	}

	type Function struct {
		ID     types.UUID    `gel:"id"`
		Name   string        `gel:"name"`
		Params []Params      `gel:"params"`
		Tuple  []interface{} `gel:"tuple"`
	}

	var result Function
	err := client.QuerySingle(ctx, query, &result)
	require.NoError(t, err)
	assert.Equal(t, "std::str_repeat", result.Name)
	assert.Equal(t, 2, len(result.Params))
	assert.Equal(t, "PositionalParam", result.Params[0].Kind)
	assert.Equal(t, int64(42), result.Params[1].Foo)
}

func TestReceiveNamedTuple(t *testing.T) {
	ctx := context.Background()

	type NamedTuple struct {
		A int64 `gel:"a"`
	}

	var result NamedTuple
	err := client.QuerySingle(ctx, "SELECT (a := 1,)", &result)
	require.NoError(t, err)
	assert.Equal(t, NamedTuple{A: 1}, result)
}

func TestReceiveTuple(t *testing.T) {
	ctx := context.Background()

	var wrongType string
	err := client.QuerySingle(ctx, `SELECT ()`, &wrongType)
	require.EqualError(t, err, "gel.InvalidArgumentError: "+
		"the \"out\" argument does not match query schema: "+
		"expected string to be a struct got string")

	var emptyStruct struct{}
	err = client.QuerySingle(ctx, `SELECT ()`, &emptyStruct)
	require.NoError(t, err)

	var missingTag struct{ first int64 }
	err = client.QuerySingle(ctx, `SELECT (<int64>$0,)`, &missingTag, int64(1))
	require.EqualError(t, err, "gel.InvalidArgumentError: "+
		"the \"out\" argument does not match query schema: "+
		"expected struct { first int64 } to have a field "+
		"with the tag `gel:\"0\"`")

	type NestedTuple struct {
		second bool    `gel:"1"`
		first  float64 `gel:"0"`
	}

	type Tuple struct {
		first  int64       `gel:"0"` // nolint:structcheck
		second string      `gel:"1"` // nolint:structcheck
		third  NestedTuple `gel:"2"` // nolint:structcheck
	}

	result := []Tuple{}
	err = client.Query(ctx, `SELECT (<int64>$0,)`, &result, int64(1))
	require.NoError(t, err)
	assert.Equal(t, []Tuple{{first: 1}}, result)

	result = []Tuple{}
	err = client.Query(ctx, `SELECT {(1, "abc"), (2, "def")}`, &result)
	require.NoError(t, err)
	require.Equal(t,
		[]Tuple{
			{first: 1, second: "abc"},
			{first: 2, second: "def"},
		},
		result,
	)

	result = []Tuple{}
	err = client.Query(ctx, `SELECT (1, "abc", (2.3, true))`, &result)
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
	err := client.QuerySingle(ctx, "SELECT <array<int64>>$0", &result, "hello")
	assert.EqualError(t, err,
		"gel.InvalidArgumentError: "+
			"expected args[0] to be a slice got: string")

	type Tuple struct {
		first []int64 `gel:"0"`
	}

	var nested Tuple
	err = client.QuerySingle(ctx,
		"SELECT (<array<int64>>$0,)", &nested, []int64{1})
	require.NoError(t, err)
	assert.Equal(t, Tuple{[]int64{1}}, nested)

	// encode empty array
	err = client.QuerySingle(ctx,
		"SELECT (<array<int64>>$0,)", &nested, []int64{})
	require.NoError(t, err)
	assert.Equal(t, Tuple{[]int64{}}, nested)

	query := "SELECT <array<int64>>$0"
	err = client.QuerySingle(ctx, query, &result, []int64{1})
	require.NoError(t, err)
	assert.Equal(t, []int64{1}, result)

	arg := []int64{1, 2, 3}
	err = client.QuerySingle(ctx, "SELECT <array<int64>>$0", &result, arg)
	require.NoError(t, err)
	assert.Equal(t, []int64{1, 2, 3}, result)
}

func TestReceiveSet(t *testing.T) {
	ctx := context.Background()

	// decoding using pointers
	{
		type Function struct {
			ID   types.UUID `gel:"id"`
			Sets [][]int64  `gel:"sets"`
		}

		query := `
			SELECT schema::Function {
				id,
				sets := {[1, 2], [1]}
			}
			LIMIT 1
		`

		var result Function
		err := client.QuerySingle(ctx, query, &result)
		require.NoError(t, err)
		assert.Equal(t, [][]int64{{1, 2}, {1}}, result.Sets)
	}

	// decoding using reflect
	{
		type NestedTuple struct {
			first int64 `gel:"0"`
		}

		type Tuple struct {
			first  int64       `gel:"0"` // nolint:structcheck
			second NestedTuple `gel:"1"` // nolint:structcheck
		}

		type Function struct {
			ID   types.UUID `gel:"id"`
			Sets [][]Tuple  `gel:"sets"`
		}

		query := `
			SELECT schema::Function {
				id,
				sets := {[(1, (2,))], [(3, (4,))]}
			}
			LIMIT 1
		`

		var result Function
		err := client.QuerySingle(ctx, query, &result)
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

type OptionalTuple struct {
	Zero int64 `gel:"0"`
	One  int64 `gel:"1"`
	types.Optional
}

func TestReceiveOptionalTuple(t *testing.T) {
	ddl := `
		CREATE TYPE Sample {
			CREATE PROPERTY val -> tuple<int64, int64>;
		};
	`
	inRolledBackTx(t, ddl, func(ctx context.Context, tx *Tx) {
		var result struct {
			Val OptionalTuple `gel:"val"`
		}

		// Decode value
		err := tx.QuerySingle(ctx, `SELECT { val := (1, 2) }`, &result)
		assert.NoError(t, err)
		expected := OptionalTuple{Zero: 1, One: 2}
		expected.SetMissing(false)
		assert.Equal(t, expected, result.Val)

		// Decode missing value
		err = tx.QuerySingle(ctx, `
			WITH inserted := (INSERT Sample)
			SELECT inserted { val }`,
			&result,
		)
		assert.NoError(t, err)
		assert.Equal(t, true, result.Val.Missing())
	})
}

type OptionalNamedTuple struct {
	A     int64 `gel:"a"`
	B     int64 `gel:"b"`
	isSet bool
}

func (t *OptionalNamedTuple) SetMissing(missing bool) {
	t.isSet = !missing
}

func inRolledBackTx(
	t *testing.T,
	ddl string,
	action func(context.Context, *Tx),
) {
	ctx := context.Background()
	err := client.Tx(ctx, func(ctx context.Context, tx *Tx) error {
		e := tx.Execute(ctx, ddl)
		assert.NoError(t, e)
		if e == nil {
			action(ctx, tx)
		}
		return errors.New("rollback")
	})
	assert.EqualError(t, err, "rollback")
}

func TestReceiveOptionalNamedTuple(t *testing.T) {
	ddl := `
		CREATE TYPE Sample {
			CREATE PROPERTY val -> tuple<a: int64, b: int64>;
		};
	`
	inRolledBackTx(t, ddl, func(ctx context.Context, tx *Tx) {
		var result struct {
			Val OptionalNamedTuple `gel:"val"`
		}

		// Decode value
		err := tx.QuerySingle(ctx, `
			SELECT { val := (a := 1, b := 2) }`,
			&result,
		)
		assert.NoError(t, err)
		assert.Equal(t,
			OptionalNamedTuple{isSet: true, A: 1, B: 2},
			result.Val,
		)

		// Decode missing value
		err = tx.QuerySingle(ctx, `
			WITH inserted := (INSERT Sample)
			SELECT inserted { val }`,
			&result,
		)
		assert.NoError(t, err)
		assert.False(t, result.Val.isSet)
	})
}

func TestReceiveOptionalObject(t *testing.T) {
	ddl := `
		CREATE TYPE Nested {
			CREATE PROPERTY a -> int64;
			CREATE PROPERTY b -> int64;
		};
		CREATE TYPE Sample {
			CREATE LINK val -> Nested;
		};
	`
	inRolledBackTx(t, ddl, func(ctx context.Context, tx *Tx) {
		type OptionalObject struct {
			types.Optional
			A types.OptionalInt64 `gel:"a"`
			B types.OptionalInt64 `gel:"b"`
		}

		var result struct {
			Val OptionalObject `gel:"val"`
		}

		// Decode value
		err := tx.QuerySingle(ctx, `
			SELECT { val := { a := 1, b := 2 } }`,
			&result,
		)
		assert.NoError(t, err)
		expected := OptionalObject{
			A: types.NewOptionalInt64(1),
			B: types.NewOptionalInt64(2),
		}
		expected.SetMissing(false)
		assert.Equal(t, expected, result.Val)

		// Decode missing value
		err = tx.QuerySingle(ctx, `
			WITH inserted := (INSERT Sample)
			SELECT inserted { val: { a, b } } LIMIT 1`,
			&result,
		)
		assert.NoError(t, err)
		assert.True(t, result.Val.Missing())
	})
}

func TestReceiveOptionalArray(t *testing.T) {
	ddl := `CREATE TYPE Sample { CREATE PROPERTY val -> array<int64>; };`
	inRolledBackTx(t, ddl, func(ctx context.Context, tx *Tx) {
		var result struct {
			Val []int64 `gel:"val"`
		}

		// Decode value
		err := tx.QuerySingle(ctx, `SELECT { val := [1, 2, 3] }`, &result)
		assert.NoError(t, err)
		assert.Equal(t, []int64{1, 2, 3}, result.Val)

		// Decode missing value
		query := `WITH inserted := (INSERT Sample) SELECT inserted { val }`
		err = tx.QuerySingle(ctx, query, &result)
		assert.NoError(t, err)
		assert.Nil(t, result.Val)
	})
}

func TestSendOptionalArray(t *testing.T) {
	ctx := context.Background()
	var result struct {
		Val []int64 `gel:"val"`
	}

	// encode value into required argument
	err := client.QuerySingle(ctx, `
		SELECT ( val := <array<int64>>$0 )`,
		&result,
		[]int64{1, 2, 3},
	)
	assert.NoError(t, err)
	assert.Equal(t, []int64{1, 2, 3}, result.Val)

	// encode value into optional argument
	err = client.QuerySingle(ctx, `
		SELECT ( val := <OPTIONAL array<int64>>$0 )`,
		&result,
		[]int64{1, 2, 3},
	)
	assert.NoError(t, err)
	assert.Equal(t, []int64{1, 2, 3}, result.Val)

	// encode missing value into optional argument
	err = client.QuerySingle(ctx, `
	SELECT { val := <OPTIONAL array<int64>>$0 }`,
		&result,
		[]int64(nil),
	)
	assert.NoError(t, err)
	assert.Equal(t, []int64(nil), result.Val)

	// encode missing value into required argument
	err = client.QuerySingle(ctx, `
		SELECT <array<int64>>$0`,
		&result.Val,
		[]int64(nil),
	)
	assert.EqualError(t, err, "gel.InvalidArgumentError: "+
		"cannot encode []int64 at args[0] "+
		"because its value is missing")
}

type OtherSample struct {
	SimpleScalar CustomOptionalInt64 `gel:"simple_scalar"`
	types.Optional
}

func TestMissingObjectFields(t *testing.T) {
	ddl := `
		CREATE TYPE Sample {
			CREATE PROPERTY simple_scalar -> int64;
			CREATE PROPERTY array -> array<int64>;
			CREATE PROPERTY tuple -> tuple<int64, int64>;
			CREATE PROPERTY named_tuple -> tuple<a: int64, b: int64>;
			CREATE LINK object -> Sample;
			CREATE MULTI LINK set_ -> Sample;
		};

		# all fields are missing
		INSERT Sample;
	`
	inRolledBackTx(t, ddl, func(ctx context.Context, tx *Tx) {
		type Sample struct {
			SimpleScalar CustomOptionalInt64 `gel:"simple_scalar"`
			Array        []int64             `gel:"array"`
			Tuple        OptionalTuple       `gel:"tuple"`
			NamedTuple   OptionalNamedTuple  `gel:"named_tuple"`
			Object       OtherSample         `gel:"object"`
			Set          []Sample            `gel:"set_"`
		}

		var result Sample
		err := tx.QuerySingle(ctx, `
			SELECT Sample {
				simple_scalar,
				array,
				tuple,
				named_tuple,
				object: { simple_scalar },
				set_: { simple_scalar },
			}
			LIMIT 1`,
			&result,
		)
		assert.NoError(t, err)

		expected := Sample{
			SimpleScalar: CustomOptionalInt64{},
			Array:        nil,
			Tuple:        OptionalTuple{},
			NamedTuple:   OptionalNamedTuple{},
			Object:       OtherSample{},
			Set:          []Sample{},
		}
		assert.Equal(t, expected, result)

		err = tx.QuerySingle(ctx, `
			WITH
				object := (INSERT Sample { simple_scalar := 2 }),
				set_ := (INSERT Sample { simple_scalar := 3 }),
				inserted := (INSERT Sample {
					simple_scalar := 1,
					array := [1],
					tuple := (1, 2),
					named_tuple := (a := 1, b := 2),
					object := object,
					set_ := set_,
				})
			SELECT inserted {
				simple_scalar,
				array,
				tuple,
				named_tuple,
				object: { simple_scalar },
				set_: { simple_scalar },
			}
			LIMIT 1`,
			&result,
		)
		assert.NoError(t, err)

		expected = Sample{
			SimpleScalar: CustomOptionalInt64{
				data:  []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01},
				isSet: true,
			},
			Array: []int64{1},
			Tuple: OptionalTuple{
				Zero: int64(1),
				One:  int64(2),
			},
			NamedTuple: OptionalNamedTuple{
				A:     int64(1),
				B:     int64(2),
				isSet: true,
			},
			Object: OtherSample{SimpleScalar: CustomOptionalInt64{
				data:  []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02},
				isSet: true,
			}},
			Set: []Sample{{SimpleScalar: CustomOptionalInt64{
				data:  []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x03},
				isSet: true,
			}}},
		}
		expected.Tuple.SetMissing(false)
		expected.Object.SetMissing(false)

		assert.Equal(t, expected, result)

		type WrongType struct {
			SimpleScalar int64 `gel:"simple_scalar"`
		}

		var result2 WrongType
		err = tx.QuerySingle(ctx, `
			SELECT Sample { simple_scalar } LIMIT 1`,
			&result2,
		)
		assert.EqualError(t, err, "gel.InvalidArgumentError: "+
			`the "out" argument does not match query schema: `+
			`expected int64 at gel.WrongType.simple_scalar to be `+
			`gel.OptionalInt64 because the field is not required`)
	})
}

func TestSendAndReceiveMemory(t *testing.T) {
	ctx := context.Background()

	memories := []types.Memory{
		types.Memory(1),
		types.Memory(0),
		types.Memory(11),
		types.Memory(15),
		types.Memory(22),
		types.Memory(113),
		types.Memory(5120),
		types.Memory(110000),
		types.Memory(6291456),
		types.Memory(7516192768),
		types.Memory(346456723423),
		types.Memory(8796093022208),
		types.Memory(281474976710656),
		types.Memory(2251799813685125),
		types.Memory(9007199254740992),
		types.Memory(10133099161583616),
		types.Memory(1152921504594725865),
	}

	for i := 0; i < 1000; i++ {
		memories = append(
			memories,
			types.Memory(rand.Int63n(9223372036854775807)),
		)
	}

	strings := make([]string, len(memories))
	for i, n := range memories {
		strings[i] = fmt.Sprint(n)
	}

	type Result struct {
		Encoded   string       `gel:"encoded"`
		Decoded   types.Memory `gel:"decoded"`
		RoundTrip types.Memory `gel:"round_trip"`
		IsEqual   bool         `gel:"is_equal"`
		String    string       `gel:"string"`
	}

	query := `
		WITH
			sample := (
				WITH
					m := enumerate(array_unpack(<array<cfg::memory>>$0)),
					s := enumerate(array_unpack(<array<str>>$1)),
				SELECT (
					m := m.1,
					str := s.1,
				)
				FILTER m.0 = s.0
			)
		SELECT (
			encoded := <str>sample.m,
			decoded := <cfg::memory>sample.str,
			round_trip := sample.m,
			is_equal := <str><cfg::memory>sample.str = <str>sample.m,
			string := <str><cfg::memory>sample.str,
		)
	`

	var results []Result
	err := client.Query(ctx, query, &results, memories, strings)
	require.NoError(t, err)
	require.Equal(t, len(memories), len(results), "wrong number of results")

	for i, s := range strings {
		t.Run(strconv.Itoa(int(memories[i])), func(t *testing.T) {
			m := memories[i]
			result := results[i]

			assert.True(t, result.IsEqual, "equality check faild")
			assert.Equal(t, s, result.Encoded, "encoding failed")
			assert.Equal(t, m, result.Decoded, "decoding failed")
			assert.Equal(t, m, result.RoundTrip, "round trip failed")
			assert.Equal(t, s, result.String)
		})
	}
}

type CustomMemory struct {
	data []byte
}

func (m CustomMemory) MarshalEdgeDBMemory() ([]byte, error) {
	data := make([]byte, len(m.data))
	copy(data, m.data)
	return data, nil
}

func (m *CustomMemory) UnmarshalEdgeDBMemory(data []byte) error {
	m.data = make([]byte, len(data))
	copy(m.data, data)
	return nil
}

func TestReceiveMemoryUnmarshaler(t *testing.T) {
	ctx := context.Background()
	var result struct {
		Val CustomMemory `gel:"val"`
	}

	// Decode value
	err := client.QuerySingle(ctx, `
		SELECT { val := <cfg::memory>123_456_789_987_654_321 }`,
		&result,
	)
	assert.NoError(t, err)
	assert.Equal(t,
		[]byte{0x01, 0xb6, 0x9b, 0x4b, 0xe0, 0x52, 0xfa, 0xb1},
		result.Val.data,
	)

	// Decode missing value
	err = client.QuerySingle(ctx, `
		SELECT { val := <OPTIONAL cfg::memory>$0 }`,
		&result,
		types.OptionalMemory{},
	)
	assert.EqualError(t, err, "gel.InvalidArgumentError: "+
		"the \"out\" argument does not match query schema: "+
		"expected gel.CustomMemory at "+
		"struct { Val gel.CustomMemory \"gel:\\\"val\\\"\" }.val "+
		"to be OptionalUnmarshaler interface "+
		"because the field is not required")
}

func TestSendMemoryMarshaler(t *testing.T) {
	ctx := context.Background()
	var result struct {
		Val types.OptionalMemory `gel:"val"`
	}

	// encode value into required argument
	err := client.QuerySingle(ctx, `
		SELECT { val := <cfg::memory>$0 }`,
		&result,
		CustomMemory{
			data: []byte{0x01, 0xb6, 0x9b, 0x4b, 0xe0, 0x52, 0xfa, 0xb1},
		},
	)
	assert.NoError(t, err)
	assert.Equal(
		t,
		types.NewOptionalMemory(123_456_789_987_654_321),
		result.Val,
	)

	// encode value into optional argument
	err = client.QuerySingle(ctx, `
		SELECT { val := <OPTIONAL cfg::memory>$0 }`,
		&result,
		CustomMemory{
			data: []byte{0x01, 0xb6, 0x9b, 0x4b, 0xe0, 0x52, 0xfa, 0xb1},
		},
	)
	assert.NoError(t, err)
	assert.Equal(
		t,
		types.NewOptionalMemory(123_456_789_987_654_321),
		result.Val,
	)

	// encode wrong number of bytes
	err = client.QuerySingle(ctx, `
		SELECT { val := <cfg::memory>$0 }`,
		&result,
		CustomMemory{data: []byte{0x01}},
	)
	assert.EqualError(t, err, "gel.InvalidArgumentError: "+
		"wrong number of bytes encoded by gel.CustomMemory "+
		"at args[0] expected 8, got 1")
}

type CustomOptionalMemory struct {
	data  []byte
	isSet bool
}

func (m CustomOptionalMemory) MarshalEdgeDBMemory() ([]byte, error) {
	if !m.isSet {
		return nil, fmt.Errorf("%T is not set", m)
	}
	data := make([]byte, len(m.data))
	copy(data, m.data)
	return data, nil
}

func (m *CustomOptionalMemory) UnmarshalEdgeDBMemory(data []byte) error {
	m.isSet = true
	m.data = make([]byte, len(data))
	copy(m.data, data)
	return nil
}

func (m *CustomOptionalMemory) SetMissing(missing bool) {
	m.isSet = !missing
	m.data = nil
}

func (m CustomOptionalMemory) Missing() bool { return !m.isSet }

func TestReceiveOptionalMemoryUnmarshaler(t *testing.T) {
	ddl := `CREATE TYPE Sample { CREATE PROPERTY val -> cfg::memory; };`
	inRolledBackTx(t, ddl, func(ctx context.Context, tx *Tx) {
		var result struct {
			Val CustomOptionalMemory `gel:"val"`
		}

		// Decode value
		err := tx.QuerySingle(ctx, `
			SELECT { val := <cfg::memory>123_456_789_987_654_321 }`,
			&result,
		)
		assert.NoError(t, err)
		assert.Equal(t,
			[]byte{0x01, 0xb6, 0x9b, 0x4b, 0xe0, 0x52, 0xfa, 0xb1},
			result.Val.data,
		)

		// Decode missing value
		query := `WITH inserted := (INSERT Sample) SELECT inserted { val }`
		err = tx.QuerySingle(ctx, query, &result)
		assert.NoError(t, err)
		assert.Equal(t, CustomOptionalMemory{}, result.Val)
	})
}

func TestSendOptionalMemoryMarshaler(t *testing.T) {
	ctx := context.Background()
	var result struct {
		Val types.OptionalMemory `gel:"val"`
	}

	newValue := func(data []byte) CustomOptionalMemory {
		return CustomOptionalMemory{isSet: true, data: data}
	}

	// encode value into required argument
	err := client.QuerySingle(ctx, `
		SELECT { val := <cfg::memory>$0 }`,
		&result,
		newValue([]byte{0x01, 0xb6, 0x9b, 0x4b, 0xe0, 0x52, 0xfa, 0xb1}),
	)
	assert.NoError(t, err)
	assert.Equal(
		t,
		types.NewOptionalMemory(123_456_789_987_654_321),
		result.Val,
	)

	// encode value into optional argument
	err = client.QuerySingle(ctx, `
		SELECT { val := <OPTIONAL cfg::memory>$0 }`,
		&result,
		newValue([]byte{0x01, 0xb6, 0x9b, 0x4b, 0xe0, 0x52, 0xfa, 0xb1}),
	)
	assert.NoError(t, err)
	assert.Equal(
		t,
		types.NewOptionalMemory(123_456_789_987_654_321),
		result.Val,
	)

	// encode missing value into optional argument
	err = client.QuerySingle(ctx, `
		SELECT { val := <OPTIONAL cfg::memory>$0 }`,
		&result,
		CustomOptionalMemory{},
	)
	assert.NoError(t, err)
	assert.Equal(t, types.OptionalMemory{}, result.Val)

	// encode missing value into required argument
	err = client.QuerySingle(ctx, `
		SELECT { val := <cfg::memory>$0 }`,
		&result,
		CustomOptionalMemory{},
	)
	assert.EqualError(t, err, "gel.InvalidArgumentError: "+
		"cannot encode gel.CustomOptionalMemory at args[0] "+
		"because its value is missing")

	// encode wrong number of bytes with required argument
	err = client.QuerySingle(ctx, `
		SELECT { val := <cfg::memory>$0 }`,
		&result,
		newValue([]byte{0x01}),
	)
	assert.EqualError(t, err, "gel.InvalidArgumentError: "+
		"wrong number of bytes encoded by gel.CustomOptionalMemory "+
		"at args[0] expected 8, got 1")

	// encode wrong number of bytes with optional argument
	err = client.QuerySingle(ctx, `
		SELECT { val := <OPTIONAL cfg::memory>$0 }`,
		&result,
		newValue([]byte{0x01}),
	)
	assert.EqualError(t, err, "gel.InvalidArgumentError: "+
		"wrong number of bytes encoded by gel.CustomOptionalMemory "+
		"at args[0] expected 8, got 1")
}

func serverHasRange(t *testing.T) bool {
	var hasRange bool
	err := client.QuerySingle(
		context.Background(),
		`SELECT count((
			SELECT names := schema::ObjectType.name
			FILTER names = 'schema::Range'
		)) = 1`,
		&hasRange,
	)
	require.NoError(t, err)
	return hasRange
}

func TestSendAndReceiveRangeInt32(t *testing.T) {
	if !serverHasRange(t) {
		t.Skip("server lacks std::range support")
	}

	samples := make([]types.RangeInt32, 1_000)
	samples[0] = types.NewRangeInt32(
		types.NewOptionalInt32(1),
		types.NewOptionalInt32(1),
		true,
		false,
	)

	for i := 1; i < 1_000; i++ {
		var lower types.OptionalInt32
		var upper types.OptionalInt32

		if rand.Intn(50) != 1 {
			l := 2_147_483_647 + rand.Uint32()/2
			lower = types.NewOptionalInt32(int32(l))
		}

		if rand.Intn(50) != 1 {
			u := rand.Uint32() / 2
			upper = types.NewOptionalInt32(int32(u))
		}

		samples[i] = types.NewRangeInt32(lower, upper, true, false)
	}

	ctx := context.Background()
	var results []struct {
		Encoded    []byte                   `gel:"encoded"`
		RoundTrip  types.RangeInt32         `gel:"round_trip"`
		Missing    types.OptionalRangeInt32 `gel:"missing"`
		NotMissing types.OptionalRangeInt32 `gel:"not_missing"`
	}
	err := client.Query(
		ctx,
		`FOR x IN array_unpack(<array<range<int32>>>$0) UNION (
			SELECT {
				encoded := <json>x,
				round_trip := x,
				missing := <OPTIONAL range<int32>>$1,
				not_missing := x,
			}
		)`,
		&results,
		samples,
		types.OptionalRangeInt32{},
	)
	require.NoError(t, err)
	require.Equal(t, len(samples), len(results))

	for i, sample := range samples {
		t.Run(fmt.Sprintf("%v", i), func(t *testing.T) {
			result := results[i]
			assert.Equal(t, sample, result.RoundTrip)
			assert.Equal(t, types.OptionalRangeInt32{}, result.Missing)
			assert.Equal(
				t,
				types.NewOptionalRangeInt32(sample),
				result.NotMissing,
			)

			encoded, err := json.Marshal(sample)
			require.NoError(t, err)
			require.Equal(
				t,
				string(encoded),
				strings.ReplaceAll(string(result.Encoded), " ", ""),
			)

			var decoded types.RangeInt32
			err = json.Unmarshal(result.Encoded, &decoded)
			require.NoError(t, err)
			assert.Equal(t, sample, decoded)
		})
	}
}

func TestSendAndReceiveRangeInt64(t *testing.T) {
	if !serverHasRange(t) {
		t.Skip("server lacks std::range support")
	}

	samples := make([]types.RangeInt64, 1_000)
	samples[0] =
		types.NewRangeInt64(
			types.NewOptionalInt64(1),
			types.NewOptionalInt64(1),
			true,
			false,
		)

	for i := 1; i < 1_000; i++ {
		var lower types.OptionalInt64
		var upper types.OptionalInt64

		if rand.Intn(50) != 1 {
			l := 0x7fffffffffffffff + rand.Uint64()/2
			lower = types.NewOptionalInt64(int64(l))
		}

		if rand.Intn(50) != 1 {
			u := rand.Uint64() / 2
			upper = types.NewOptionalInt64(int64(u))
		}

		samples[i] = types.NewRangeInt64(lower, upper, true, false)
	}

	ctx := context.Background()
	var results []struct {
		Encoded    []byte                   `gel:"encoded"`
		RoundTrip  types.RangeInt64         `gel:"round_trip"`
		Missing    types.OptionalRangeInt64 `gel:"missing"`
		NotMissing types.OptionalRangeInt64 `gel:"not_missing"`
	}
	err := client.Query(
		ctx,
		`FOR x IN array_unpack(<array<range<int64>>>$0) UNION (
			SELECT {
				encoded := <json>x,
				round_trip := x,
				missing := <OPTIONAL range<int64>>$1,
				not_missing := x,
			}
		)`,
		&results,
		samples,
		types.OptionalRangeInt64{},
	)
	require.NoError(t, err)
	require.Equal(t, len(samples), len(results))

	for i, sample := range samples {
		t.Run(fmt.Sprintf("%v", i), func(t *testing.T) {
			result := results[i]
			assert.Equal(t, sample, result.RoundTrip)
			assert.Equal(t, types.OptionalRangeInt64{}, result.Missing)
			assert.Equal(
				t,
				types.NewOptionalRangeInt64(sample),
				result.NotMissing,
			)

			encoded, err := json.Marshal(sample)
			require.NoError(t, err)
			require.Equal(
				t,
				string(encoded),
				strings.ReplaceAll(string(result.Encoded), " ", ""),
			)

			var decoded types.RangeInt64
			err = json.Unmarshal(result.Encoded, &decoded)
			require.NoError(t, err)
			assert.Equal(t, sample, decoded)
		})
	}
}

func TestSendAndReceiveRangeFloat32(t *testing.T) {
	if !serverHasRange(t) {
		t.Skip("server lacks std::range support")
	}

	samples := make([]types.RangeFloat32, 1_000)
	samples[0] =
		types.NewRangeFloat32(
			types.NewOptionalFloat32(1),
			types.NewOptionalFloat32(1),
			true,
			false,
		)

	for i := 1; i < 1_000; i++ {
		l := -1_000 + 2_000*rand.Float32()
		lower := types.NewOptionalFloat32(l)
		u := l + (1_000-l)*rand.Float32()
		upper := types.NewOptionalFloat32(u)

		if rand.Intn(50) == 1 {
			lower = types.OptionalFloat32{}
		}

		if rand.Intn(50) == 1 {
			upper = types.OptionalFloat32{}
		}

		samples[i] = types.NewRangeFloat32(
			lower,
			upper,
			rand.Intn(2) == 1,
			rand.Intn(2) == 1,
		)
	}

	ctx := context.Background()
	var results []struct {
		Encoded    []byte                     `gel:"encoded"`
		RoundTrip  types.RangeFloat32         `gel:"round_trip"`
		Missing    types.OptionalRangeFloat32 `gel:"missing"`
		NotMissing types.OptionalRangeFloat32 `gel:"not_missing"`
	}
	err := client.Query(
		ctx,
		`FOR x IN array_unpack(<array<range<float32>>>$0) UNION (
			SELECT {
				encoded := <json>x,
				round_trip := x,
				missing := <OPTIONAL range<float32>>$1,
				not_missing := x,
			}
		)`,
		&results,
		samples,
		types.OptionalRangeFloat32{},
	)
	require.NoError(t, err)
	require.Equal(t, len(samples), len(results))

	for i, sample := range samples {
		t.Run(fmt.Sprintf("%v", i), func(t *testing.T) {
			result := results[i]
			assert.Equal(t, sample, result.RoundTrip)
			assert.Equal(t, types.OptionalRangeFloat32{}, result.Missing)
			assert.Equal(
				t,
				types.NewOptionalRangeFloat32(sample),
				result.NotMissing,
			)

			encoded, err := json.Marshal(sample)
			require.NoError(t, err)
			require.Equal(
				t,
				string(encoded),
				strings.ReplaceAll(string(result.Encoded), " ", ""),
			)

			var decoded types.RangeFloat32
			err = json.Unmarshal(result.Encoded, &decoded)
			require.NoError(t, err)
			assert.Equal(t, sample, decoded)
		})
	}
}

func TestSendAndReceiveRangeFloat64(t *testing.T) {
	if !serverHasRange(t) {
		t.Skip("server lacks std::range support")
	}

	samples := make([]types.RangeFloat64, 1_000)
	samples[0] =
		types.NewRangeFloat64(
			types.NewOptionalFloat64(1),
			types.NewOptionalFloat64(1),
			true,
			false,
		)

	for i := 1; i < 1_000; i++ {
		l := -1_000 + 2_000*rand.Float64()
		lower := types.NewOptionalFloat64(l)
		u := l + (1_000-l)*rand.Float64()
		upper := types.NewOptionalFloat64(u)

		if rand.Intn(50) == 1 {
			lower = types.OptionalFloat64{}
		}

		if rand.Intn(50) == 1 {
			upper = types.OptionalFloat64{}
		}

		samples[i] = types.NewRangeFloat64(
			lower,
			upper,
			rand.Intn(2) == 1,
			rand.Intn(2) == 1,
		)
	}

	ctx := context.Background()
	var results []struct {
		Encoded    []byte                     `gel:"encoded"`
		RoundTrip  types.RangeFloat64         `gel:"round_trip"`
		Missing    types.OptionalRangeFloat64 `gel:"missing"`
		NotMissing types.OptionalRangeFloat64 `gel:"not_missing"`
	}
	err := client.Query(
		ctx,
		`FOR x IN array_unpack(<array<range<float64>>>$0) UNION (
			SELECT {
				encoded := <json>x,
				round_trip := x,
				missing := <OPTIONAL range<float64>>$1,
				not_missing := x,
			}
		)`,
		&results,
		samples,
		types.OptionalRangeFloat64{},
	)
	require.NoError(t, err)
	require.Equal(t, len(samples), len(results))

	for i, sample := range samples {
		t.Run(fmt.Sprintf("%v", i), func(t *testing.T) {
			result := results[i]
			assert.Equal(t, sample, result.RoundTrip)
			assert.Equal(t, types.OptionalRangeFloat64{}, result.Missing)
			assert.Equal(
				t,
				types.NewOptionalRangeFloat64(sample),
				result.NotMissing,
			)

			encoded, err := json.Marshal(sample)
			require.NoError(t, err)
			require.Equal(
				t,
				string(encoded),
				strings.ReplaceAll(string(result.Encoded), " ", ""),
			)

			var decoded types.RangeFloat64
			err = json.Unmarshal(result.Encoded, &decoded)
			require.NoError(t, err)
			assert.Equal(t, sample, decoded)
		})
	}
}

func TestSendAndReceiveRangeDateTime(t *testing.T) {
	if !serverHasRange(t) {
		t.Skip("server lacks std::range support")
	}

	samples := make([]types.RangeDateTime, 1_000)
	samples[0] =
		types.NewRangeDateTime(
			types.NewOptionalDateTime(
				time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)),
			types.NewOptionalDateTime(
				time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)),
			true,
			false,
		)

	const maxDate = 253402300799
	const minDate = -62135596800

	for i := 1; i < 1_000; i++ {
		l := rand.Int63n(maxDate-minDate) + minDate
		u := rand.Int63n(maxDate-l) + l

		lower := types.NewOptionalDateTime(time.Unix(l, 0).UTC())
		upper := types.NewOptionalDateTime(time.Unix(u, 0).UTC())

		if rand.Intn(50) == 1 {
			lower = types.OptionalDateTime{}
		}

		if rand.Intn(50) == 1 {
			upper = types.OptionalDateTime{}
		}

		samples[i] = types.NewRangeDateTime(
			lower,
			upper,
			rand.Intn(2) == 1,
			rand.Intn(2) == 1,
		)
	}

	ctx := context.Background()
	var results []struct {
		Encoded    []byte                      `gel:"encoded"`
		RoundTrip  types.RangeDateTime         `gel:"round_trip"`
		Missing    types.OptionalRangeDateTime `gel:"missing"`
		NotMissing types.OptionalRangeDateTime `gel:"not_missing"`
	}
	err := client.Query(
		ctx,
		`FOR x IN array_unpack(<array<range<datetime>>>$0) UNION (
			SELECT {
				encoded := <json>x,
				round_trip := x,
				missing := <OPTIONAL range<datetime>>$1,
				not_missing := x,
			}
		)`,
		&results,
		samples,
		types.OptionalRangeDateTime{},
	)
	require.NoError(t, err)
	require.Equal(t, len(samples), len(results))

	for i, sample := range samples {
		t.Run(fmt.Sprintf("%v", i), func(t *testing.T) {
			result := results[i]
			assert.Equal(t, sample, result.RoundTrip)
			assert.Equal(t, types.OptionalRangeDateTime{}, result.Missing)
			assert.Equal(
				t,
				types.NewOptionalRangeDateTime(sample),
				result.NotMissing,
			)

			encoded := string(result.Encoded)
			encoded = strings.ReplaceAll(encoded, "+00:00", "Z")
			encoded = strings.ReplaceAll(encoded, " ", "")

			marshaled, err := json.Marshal(sample)
			require.NoError(t, err)
			require.Equal(t, string(marshaled), encoded)

			var decoded types.RangeDateTime
			err = json.Unmarshal([]byte(encoded), &decoded)
			require.NoError(t, err)
			assert.Equal(t, sample, decoded)
		})
	}
}

func TestSendAndReceiveRangeLocalDateTime(t *testing.T) {
	if !serverHasRange(t) {
		t.Skip("server lacks std::range support")
	}

	samples := make([]types.RangeLocalDateTime, 1_000)
	samples[0] =
		types.NewRangeLocalDateTime(
			types.NewOptionalLocalDateTime(
				types.NewLocalDateTime(1970, 1, 1, 0, 0, 0, 0)),
			types.NewOptionalLocalDateTime(
				types.NewLocalDateTime(1970, 1, 1, 0, 0, 0, 0)),
			true,
			false,
		)

	const maxDate = 30
	const minDate = 1

	for i := 1; i < 1_000; i++ {
		l := rand.Intn(maxDate-minDate) + minDate
		u := rand.Intn(maxDate-l) + l

		lower := types.NewOptionalLocalDateTime(
			types.NewLocalDateTime(1970, 1, l, 0, 0, 0, 0),
		)
		upper := types.NewOptionalLocalDateTime(
			types.NewLocalDateTime(1970, 1, u, 0, 0, 0, 0),
		)

		if rand.Intn(50) == 1 {
			lower = types.OptionalLocalDateTime{}
		}

		if rand.Intn(50) == 1 {
			upper = types.OptionalLocalDateTime{}
		}

		samples[i] = types.NewRangeLocalDateTime(
			lower,
			upper,
			rand.Intn(2) == 1,
			rand.Intn(2) == 1,
		)
	}

	ctx := context.Background()
	var results []struct {
		Encoded    []byte                           `gel:"encoded"`
		RoundTrip  types.RangeLocalDateTime         `gel:"round_trip"`
		Missing    types.OptionalRangeLocalDateTime `gel:"missing"`
		NotMissing types.OptionalRangeLocalDateTime `gel:"not_missing"`
	}
	err := client.Query(
		ctx,
		`FOR x IN array_unpack(<array<range<cal::local_datetime>>>$0) UNION (
			SELECT {
				encoded := <json>x,
				round_trip := x,
				missing := <OPTIONAL range<cal::local_datetime>>$1,
				not_missing := x,
			}
		)`,
		&results,
		samples,
		types.OptionalRangeLocalDateTime{},
	)
	require.NoError(t, err)
	require.Equal(t, len(samples), len(results))

	for i, sample := range samples {
		t.Run(fmt.Sprintf("%v", i), func(t *testing.T) {
			result := results[i]
			assert.Equal(t, sample, result.RoundTrip)
			assert.Equal(t, types.OptionalRangeLocalDateTime{}, result.Missing)
			assert.Equal(
				t,
				types.NewOptionalRangeLocalDateTime(sample),
				result.NotMissing,
			)

			encoded, err := json.Marshal(sample)
			require.NoError(t, err)
			require.Equal(
				t,
				string(encoded),
				strings.ReplaceAll(string(result.Encoded), " ", ""),
			)

			var decoded types.RangeLocalDateTime
			err = json.Unmarshal(result.Encoded, &decoded)
			require.NoError(t, err)
			assert.Equal(t, sample, decoded)
		})
	}
}

func TestSendAndReceiveRangeLocalDate(t *testing.T) {
	if !serverHasRange(t) {
		t.Skip("server lacks std::range support")
	}

	samples := make([]types.RangeLocalDate, 1_000)
	samples[0] =
		types.NewRangeLocalDate(
			types.NewOptionalLocalDate(types.NewLocalDate(1970, 1, 1)),
			types.NewOptionalLocalDate(types.NewLocalDate(1970, 1, 1)),
			true,
			false,
		)

	const maxDate = 30
	const minDate = 1

	for i := 1; i < 1_000; i++ {
		l := rand.Intn(maxDate-minDate) + minDate
		u := rand.Intn(maxDate-l) + l

		lower := types.NewOptionalLocalDate(
			types.NewLocalDate(1970, 1, l),
		)
		upper := types.NewOptionalLocalDate(
			types.NewLocalDate(1970, 1, u),
		)

		if rand.Intn(50) == 1 {
			lower = types.OptionalLocalDate{}
		}

		if rand.Intn(50) == 1 {
			upper = types.OptionalLocalDate{}
		}

		samples[i] = types.NewRangeLocalDate(lower, upper, true, false)
	}

	ctx := context.Background()
	var results []struct {
		Encoded    []byte                       `gel:"encoded"`
		RoundTrip  types.RangeLocalDate         `gel:"round_trip"`
		Missing    types.OptionalRangeLocalDate `gel:"missing"`
		NotMissing types.OptionalRangeLocalDate `gel:"not_missing"`
	}
	err := client.Query(
		ctx,
		`FOR x IN array_unpack(<array<range<cal::local_date>>>$0) UNION (
			SELECT {
				encoded := <json>x,
				round_trip := x,
				missing := <OPTIONAL range<cal::local_date>>$1,
				not_missing := x,
			}
		)`,
		&results,
		samples,
		types.OptionalRangeLocalDate{},
	)
	require.NoError(t, err)
	require.Equal(t, len(samples), len(results))

	for i, sample := range samples {
		t.Run(fmt.Sprintf("%v", i), func(t *testing.T) {
			result := results[i]
			assert.Equal(t, sample, result.RoundTrip)
			assert.Equal(t, types.OptionalRangeLocalDate{}, result.Missing)
			assert.Equal(
				t,
				types.NewOptionalRangeLocalDate(sample),
				result.NotMissing,
			)

			encoded, err := json.Marshal(sample)
			require.NoError(t, err)
			require.Equal(
				t,
				string(encoded),
				strings.ReplaceAll(string(result.Encoded), " ", ""),
			)

			var decoded types.RangeLocalDate
			err = json.Unmarshal(result.Encoded, &decoded)
			require.NoError(t, err)
			assert.Equal(t, sample, decoded)
		})
	}
}

func serverHasMultiRange(t *testing.T) bool {
	var hasMultiRange bool
	err := client.QuerySingle(
		context.Background(),
		`SELECT count((
			SELECT names := schema::ObjectType.name
			FILTER names = 'schema::MultiRange'
		)) = 1`,
		&hasMultiRange,
	)
	require.NoError(t, err)
	return hasMultiRange
}

func TestSendAndReceiveInt32MultiRange(t *testing.T) {
	if !serverHasMultiRange(t) {
		t.Skip("server lacks std::MultiRange support")
	}

	ctx := context.Background()

	var result types.MultiRangeInt32

	multiRange := make([]types.RangeInt32, 2)

	multiRange[0] = types.NewRangeInt32(
		types.NewOptionalInt32(1),
		types.NewOptionalInt32(5),
		true,
		false,
	)

	multiRange[1] = types.NewRangeInt32(
		types.NewOptionalInt32(8),
		types.NewOptionalInt32(10),
		true,
		false,
	)

	query := "SELECT <multirange<int32>>$0"
	err := client.QuerySingle(ctx, query, &result, multiRange)
	require.NoError(t, err)
	assert.Equal(t, multiRange, result)
}

func TestEmptyMultiRange(t *testing.T) {
	if !serverHasMultiRange(t) {
		t.Skip("server lacks std::MultiRange support")
	}

	ctx := context.Background()

	var result types.MultiRangeFloat32

	emptyMultiRange := []types.RangeFloat32{}

	query := "SELECT <multirange<float32>>$0"
	err := client.QuerySingle(ctx, query, &result, emptyMultiRange)
	require.NoError(t, err)
	assert.Equal(t, emptyMultiRange, result)
}

func TestCustomSequenceTypeHandling(t *testing.T) {
	ddl := `
		CREATE SCALAR TYPE SampleSequence extending std::sequence;
		CREATE TYPE Sample { CREATE PROPERTY val -> SampleSequence; };
	`
	inRolledBackTx(t, ddl, func(ctx context.Context, tx *Tx) {
		var result struct {
			Val types.OptionalInt64 `gel:"val"`
		}

		// Decode value
		err := tx.QuerySingle(ctx, `
			with inserted_sample := (insert Sample{})
			select inserted_sample { val }
			`,
			&result,
		)
		assert.NoError(t, err)

		val, isSet := result.Val.Get()
		assert.Equal(t, int64(1), val)
		assert.True(t, isSet)
	})
}
