package edgedb

import (
	"github.com/edgedb/edgedb-go/edgedb/protocol/cardinality"
	"github.com/edgedb/edgedb-go/edgedb/protocol/format"
)

type query struct {
	cmd  string
	fmt  uint8
	card uint8
	args []interface{}
}

func (q *query) flat() bool {
	if q.card != cardinality.Many {
		return true
	}

	if q.fmt == format.JSON {
		return true
	}

	return false
}
