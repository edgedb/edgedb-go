package edgedb

import (
	"math/rand"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTransactionSaves(t *testing.T) {
	tx, err := conn.Transaction()
	assert.Nil(t, err)

	err = tx.Start()
	require.Nil(t, err)

	name := "test" + strconv.Itoa(rand.Int())
	// todo maybe clean up the random entry :thinking:
	err = conn.Query("INSERT User{ name := <str>$0 }", (*interface{})(nil), name)
	assert.Nil(t, err)

	err = tx.Commit()
	require.Nil(t, err)

	var result string
	err = conn.QueryOne(`
			SELECT User.name
			FILTER User.name = <str>$0;
		`,
		&result,
		name,
	)

	assert.Nil(t, err)
	assert.Equal(t, name, result)
}

func TestTransactionRollsBack(t *testing.T) {
	tx, err := conn.Transaction()
	assert.Nil(t, err)

	err = tx.Start()
	require.Nil(t, err)

	name := "test" + strconv.Itoa(rand.Int())
	// todo maybe clean up the random entry :thinking:
	err = conn.Query("INSERT User{ name := <str>$0 }", (*interface{})(nil), name)
	assert.Nil(t, err)

	err = tx.RollBack()
	require.Nil(t, err)

	var result string
	err = conn.QueryOne(`
			SELECT User.name
			FILTER User.name = <str>$0;
		`,
		&result,
		name,
	)

	assert.Equal(t, ErrorZeroResults, err)

}
