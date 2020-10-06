package edgedb

import (
	"testing"

	"github.com/fmoor/edgedb-golang/edgedb/options"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNamedQueryArguments(t *testing.T) {
	opts := options.FromDSN("edgedb://edgedb@localhost:5656/edgedb")
	conn, err := Connect(opts)
	require.Nil(t, err)
	defer conn.Close()

	result := [][]int64{}
	err = conn.Query(
		"SELECT [<int64>$first, <int64>$second]",
		&result,
		map[string]interface{}{
			"first":  int64(5),
			"second": int64(8),
		},
	)

	assert.Nil(t, err)
	assert.Equal(t, [][]int64{[]int64{5, 8}}, result)
}

func TestNumberedQueryArguments(t *testing.T) {
	opts := options.FromDSN("edgedb://edgedb@localhost:5656/edgedb")
	conn, err := Connect(opts)
	require.Nil(t, err)
	defer conn.Close()

	result := [][]int64{}
	err = conn.Query(
		"SELECT [<int64>$0, <int64>$1]",
		&result,
		int64(5),
		int64(8),
	)

	assert.Nil(t, err)
	assert.Equal(t, [][]int64{[]int64{5, 8}}, result)
}
