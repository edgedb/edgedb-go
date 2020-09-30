package edgedb

import (
	"fmt"
	"testing"

	"github.com/fmoor/edgedb-golang/edgedb/types"
	"github.com/stretchr/testify/assert"
)

func TestReflectInt64(t *testing.T) {
	options := ConnConfig{"edgedb", "edgedb"}
	edb, _ := Connect(options)
	defer edb.Close()

	out := make([]int64, 0)
	err := edb.Query("SELECT <int64>7", &out)
	assert.Nil(t, err)
	assert.Equal(t, []int64{7}, out)
}

var exampleID types.UUID = [16]byte{
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
}

func TestReflectStruct(t *testing.T) {
	options := ConnConfig{"edgedb", "edgedb"}
	edb, _ := Connect(options)
	defer edb.Close()

	type Database struct {
		Name string     `edgedb:"name"`
		ID   types.UUID `edgedb:"id"`
	}

	out := make([]Database, 0)
	err := edb.Query("SELECT sys::Database{name, id}", &out)
	fmt.Println(out)
	assert.Nil(t, err)

	expectedNames := []string{"edgedb", "edgedb0"}
	assert.Equal(t, len(expectedNames), len(out))
	for i := 0; i < len(out); i++ {
		db := out[i]
		assert.IsType(t, exampleID, db.ID)
		assert.Equal(t, expectedNames[i], db.Name)
	}
}
