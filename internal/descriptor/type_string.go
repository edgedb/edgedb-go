// Code generated by "stringer -type Type"; DO NOT EDIT.

package descriptor

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[Set-0]
	_ = x[Object-1]
	_ = x[BaseScalar-2]
	_ = x[Scalar-3]
	_ = x[Tuple-4]
	_ = x[NamedTuple-5]
	_ = x[Array-6]
	_ = x[Enum-7]
	_ = x[InputShape-8]
	_ = x[Range-9]
	_ = x[ObjectShape-10]
	_ = x[Compound-11]
}

const _Type_name = "SetObjectBaseScalarScalarTupleNamedTupleArrayEnumInputShapeRangeObjectShapeCompound"

var _Type_index = [...]uint8{0, 3, 9, 19, 25, 30, 40, 45, 49, 59, 64, 75, 83}

func (i Type) String() string {
	if i >= Type(len(_Type_index)-1) {
		return "Type(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _Type_name[_Type_index[i]:_Type_index[i+1]]
}
