package types

// UUID a universally unique identifier
// https://www.edgedb.com/docs/datamodel/scalars/uuid#type::std::uuid
type UUID [16]byte

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
