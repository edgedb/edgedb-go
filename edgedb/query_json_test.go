package edgedb

import (
	"testing"

	"github.com/fmoor/edgedb-golang/edgedb/options"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQueryJSON(t *testing.T) {
	opts := options.FromDSN("edgedb://edgedb@localhost:5656/edgedb")
	conn, err := Connect(opts)
	require.Nil(t, err)
	defer conn.Close()

	result, err := conn.QueryJSON(
		"SELECT {(a := 0, b := <int64>$0), (a := 42, b := <int64>$1)}",
		int64(1),
		int64(2),
	)

	// casting to string makes error message more helpful
	// when this test fails
	actual := string(result)

	assert.Nil(t, err)
	assert.Equal(t, "[{\"a\":0,\"b\":1},{\"a\":42,\"b\":2}]", actual)
}
