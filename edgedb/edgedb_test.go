package main

import (
	"testing"
	"time"

	"github.com/fmoor/edgedb-golang/edgedb/protocol"
	"github.com/stretchr/testify/assert"
)

type testCase struct {
	query  string
	result interface{}
}

var testCases = []testCase{
	testCase{"SELECT 1;", []interface{}{int64(1)}},
	testCase{"SELECT <int16>1;", []interface{}{int16(1)}},
	testCase{"SELECT <int32>1;", []interface{}{int32(1)}},
	testCase{"SELECT <int64>1;", []interface{}{int64(1)}},
	testCase{"SELECT <float32>1;", []interface{}{float32(1.0)}},
	testCase{"SELECT <float64>1;", []interface{}{1.0}},
	testCase{"SELECT 'hello';", []interface{}{"hello"}},
	testCase{"SELECT b'world';", []interface{}{[]byte{119, 111, 114, 108, 100}}},
	testCase{"SELECT [1, 2, 3];", []interface{}{[]interface{}{int64(1), int64(2), int64(3)}}},
	testCase{"SELECT ('foo', [1, 2, 3]);", []interface{}{[]interface{}{"foo", []interface{}{int64(1), int64(2), int64(3)}}}},
	testCase{"SELECT true;", []interface{}{true}},
	testCase{"SELECT false;", []interface{}{false}},
	testCase{
		"SELECT <datetime>'2018-05-07T15:01:22.306916+00';",
		[]interface{}{time.Date(2018, 5, 7, 15, 1, 22, 306_916_000, time.UTC)},
	},
	testCase{
		"SELECT <cal::local_datetime>'2018-05-07T15:01:22.306916';",
		[]interface{}{time.Date(2018, 5, 7, 15, 1, 22, 306_916_000, time.UTC)},
	},
	testCase{
		"SELECT <cal::local_date>'2018-05-07';",
		[]interface{}{time.Date(2018, 5, 7, 0, 0, 0, 0, time.UTC)},
	},
	testCase{
		"SELECT <cal::local_time>'15:01:22.306916';",
		[]interface{}{time.Duration(54082306916000)},
	},
	testCase{
		"SELECT <duration>'48 hours 45 minutes';",
		[]interface{}{time.Duration(175500000000000)},
	},
	testCase{"SELECT <json>42;", []interface{}{float64(42)}},
	testCase{
		"SELECT sys::Database{ name } FILTER .name = 'edgedb';",
		[]interface{}{map[string]interface{}{"id": protocol.UUID("2ea37081-e7f3-11ea-b9d3-1934-52ed3b14"), "name": "edgedb"}},
	},
}

func TestQueries(t *testing.T) {
	edb, _ := Connect("edgedb")
	defer edb.Close()

	for _, query := range testCases {
		t.Run(query.query, func(t *testing.T) {
			result, err := edb.Query(query.query)
			assert.Nil(t, err)
			assert.Equal(t, query.result, result)
		})
	}
}
