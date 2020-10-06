package edgedb

import (
	"testing"

	"github.com/fmoor/edgedb-golang/edgedb/options"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQueryOne(t *testing.T) {
	opts := options.FromDSN("edgedb://edgedb@localhost:5656/edgedb")
	conn, err := Connect(opts)
	require.Nil(t, err)
	defer conn.Close()

	var result int64
	err = conn.QueryOne("SELECT 42", &result)

	assert.Nil(t, err)
	assert.Equal(t, int64(42), result)
}
