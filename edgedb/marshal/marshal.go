package marshal

import (
	"reflect"
)

func Marshal(out *interface{}, in interface{}) {
	ov := reflect.ValueOf(out).Elem().Elem().Elem()
	iv := reflect.ValueOf(in)
	setValue(ov, iv)
}

func setValue(out reflect.Value, in reflect.Value) {
	if in.Kind() == reflect.Interface {
		in = in.Elem()
	}

	switch out.Kind() {
	case reflect.Struct:
		setStruct(out, in)
	case reflect.Slice:
		setSlice(out, in)
	default:
		out.Set(in)
	}
}

func setSlice(out reflect.Value, in reflect.Value) {
	tmp := reflect.MakeSlice(out.Type(), in.Len(), in.Len())

	for i := 0; i < in.Len(); i++ {
		setValue(tmp.Index(i), in.Index(i))
	}

	out.Set(tmp)
}

func setStruct(out reflect.Value, in reflect.Value) {
	iter := in.MapRange()
	for iter.Next() {
		setField(out, in, iter.Key())
	}
}

func setField(out reflect.Value, in reflect.Value, name reflect.Value) {
	fieldName := name.Interface().(string)
	outField := fieldByTag(out, fieldName)
	inField := in.MapIndex(name)
	if outField.IsValid() {
		setValue(outField, inField)
	}
}

func fieldByTag(out reflect.Value, name string) reflect.Value {
	sType := out.Type()
	for i := 0; i < sType.NumField(); i++ {
		field := sType.Field(i)
		if field.Tag.Get("edgedb") == name {
			return out.FieldByName(field.Name)
		}
	}
	return reflect.Value{}
}
