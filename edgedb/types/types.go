package types

import "fmt"

// UUID a universally unique identifier
// https://www.edgedb.com/docs/datamodel/scalars/uuid#type::std::uuid
type UUID [16]byte

func (id UUID) String() string {
	// todo format string same as EdgeDB
	// https://www.edgedb.com/docs/internals/protocol/dataformats#std-uuid
	return fmt.Sprintf("% x", id[:])
}

// Set https://www.edgedb.com/docs/edgeql/overview#everything-is-a-set
type Set []interface{}

// Object https://www.edgedb.com/docs/datamodel/objects#type::std::Object
type Object map[string]interface{}

// Array https://www.edgedb.com/docs/datamodel/colltypes#type::std::array
type Array []interface{}

// Tuple https://www.edgedb.com/docs/datamodel/colltypes#type::std::tuple
type Tuple []interface{}

// NamedTuple ?
type NamedTuple map[string]interface{}
