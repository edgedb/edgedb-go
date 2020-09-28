package edgedb

import (
	"testing"
	"time"

	"github.com/fmoor/edgedb-golang/edgedb/protocol"
	"github.com/stretchr/testify/assert"
)

type decodeTestCase struct {
	query  string
	result []interface{}
}

var decodeTestCases = []decodeTestCase{
	decodeTestCase{"SELECT <int16>1;", []interface{}{int16(1)}},
	decodeTestCase{"SELECT <int32>1;", []interface{}{int32(1)}},
	decodeTestCase{"SELECT <int64>1;", []interface{}{int64(1)}},
	decodeTestCase{"SELECT <float32>1;", []interface{}{float32(1.0)}},
	decodeTestCase{"SELECT <float64>1;", []interface{}{1.0}},
	decodeTestCase{"SELECT 'hello';", []interface{}{"hello"}},
	decodeTestCase{"SELECT b'world';", []interface{}{[]byte{119, 111, 114, 108, 100}}},
	decodeTestCase{"SELECT [1, 2, 3];", []interface{}{[]interface{}{int64(1), int64(2), int64(3)}}},
	decodeTestCase{"SELECT ('foo', [1, 2, 3]);", []interface{}{[]interface{}{"foo", []interface{}{int64(1), int64(2), int64(3)}}}},
	decodeTestCase{"SELECT true;", []interface{}{true}},
	decodeTestCase{"SELECT false;", []interface{}{false}},
	decodeTestCase{
		"SELECT <datetime>'2018-05-07T15:01:22.306916+00';",
		[]interface{}{time.Date(2018, 5, 7, 15, 1, 22, 306_916_000, time.UTC)},
	},
	decodeTestCase{
		"SELECT <cal::local_datetime>'2018-05-07T15:01:22.306916';",
		[]interface{}{time.Date(2018, 5, 7, 15, 1, 22, 306_916_000, time.UTC)},
	},
	decodeTestCase{
		"SELECT <cal::local_date>'2018-05-07';",
		[]interface{}{time.Date(2018, 5, 7, 0, 0, 0, 0, time.UTC)},
	},
	decodeTestCase{
		"SELECT <cal::local_time>'15:01:22.306916';",
		[]interface{}{time.Duration(54082306916000)},
	},
	decodeTestCase{
		"SELECT <duration>'48 hours 45 minutes';",
		[]interface{}{time.Duration(175500000000000)},
	},
	decodeTestCase{"SELECT <json>42;", []interface{}{float64(42)}},
	decodeTestCase{
		"SELECT sys::Database{ name } FILTER .name = 'edgedb';",
		[]interface{}{map[string]interface{}{"id": protocol.UUID("2ea37081-e7f3-11ea-b9d3-1934-52ed3b14"), "name": "edgedb"}},
	},
}

func TestDecodeQueries(t *testing.T) {
	options := ConnConfig{"edgedb", "edgedb"}
	edb, _ := Connect(options)
	defer edb.Close()

	for _, testCase := range decodeTestCases {
		t.Run(testCase.query, func(t *testing.T) {
			result, err := edb.Query(testCase.query)
			assert.Nil(t, err)
			assert.Equal(t, testCase.result, result)
		})
	}
}
