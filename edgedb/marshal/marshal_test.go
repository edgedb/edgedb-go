package marshal

import (
	"reflect"
	"testing"

	"github.com/fmoor/edgedb-golang/edgedb/types"
	"github.com/stretchr/testify/assert"
)

func TestMarshalSetOfScalar(t *testing.T) {
	var result interface{} = &[]int64{}
	input := types.Set{int64(3), int64(5), int64(8)}
	Marshal(&result, input)
	assert.Equal(t, []int64{3, 5, 8}, *(result.(*[]int64)))
}

func TestMarshalSetOfObject(t *testing.T) {

	type Database struct {
		Name string     `edgedb:"name"`
		ID   types.UUID `edgedb:"id"`
	}

	input := types.Set{
		types.Object{
			"name": "edgedb",
			"id":   types.UUID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		},
		types.Object{
			"name": "tutorial",
			"id":   types.UUID{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
		},
	}

	expected := []Database{
		Database{
			Name: "edgedb",
			ID:   types.UUID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		},
		Database{
			Name: "tutorial",
			ID:   types.UUID{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
		},
	}

	var result interface{} = &[]Database{}
	Marshal(&result, input)
	assert.Equal(t, expected, *(result.(*[]Database)))
}

func TestSetNilScalar(t *testing.T) {
	var out int64
	in := types.Set{}

	setScalar(reflect.ValueOf(out), reflect.ValueOf(in))

	assert.Equal(t, int64(0), out)
}

func TestSetScalar(t *testing.T) {
	out := int64(0)
	in := int64(27)

	ov := reflect.ValueOf(&out)
	setScalar(ov.Elem(), reflect.ValueOf(in))

	assert.Equal(t, int64(27), out)
}
