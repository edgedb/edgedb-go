package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type encodeTestCase struct {
	query    string
	argument interface{}
	result   []interface{}
}

var encodeTestCases = []encodeTestCase{
	// todo test other base scalar types
	encodeTestCase{"SELECT <bool>$val", true, []interface{}{true}},
	encodeTestCase{"SELECT <bool>$val", false, []interface{}{false}},
	encodeTestCase{"SELECT <int16>$val", int16(1), []interface{}{int16(1)}},
	encodeTestCase{"SELECT <int32>$val", int32(1), []interface{}{int32(1)}},
	encodeTestCase{"SELECT <int64>$val", int64(1), []interface{}{int64(1)}},
	encodeTestCase{"SELECT <float64>$val", float64(1), []interface{}{float64(1)}},
	encodeTestCase{"SELECT <float32>$val", float32(1), []interface{}{float32(1)}},
	encodeTestCase{"SELECT <str>$val", "hello", []interface{}{"hello"}},
	encodeTestCase{"SELECT <bytes>$val", []uint8{1, 2, 3}, []interface{}{[]uint8{1, 2, 3}}},
	encodeTestCase{
		"SELECT <array<int64>>$val",
		[]interface{}{int64(1), int64(2), int64(3)},
		[]interface{}{[]interface{}{int64(1), int64(2), int64(3)}},
	},
}

func TestEncodeQueries(t *testing.T) {
	options := ConnConfig{"edgedb", "edgedb"}
	edb, _ := Connect(options)
	defer edb.Close()

	for _, testCase := range encodeTestCases {
		t.Run(testCase.query, func(t *testing.T) {
			result, err := edb.QueryWithArgs(
				testCase.query,
				map[string]interface{}{"val": testCase.argument},
			)
			assert.Nil(t, err)
			assert.Equal(t, testCase.result, result)
		})
	}
}
