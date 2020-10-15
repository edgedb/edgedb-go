// This source file is part of the EdgeDB open source project.
//
// Copyright 2020-present EdgeDB Inc. and the EdgeDB authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package marshal

import (
	"reflect"
)

// Marshal populates out with values from in
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
		setScalar(out, in)
	}
}

func setScalar(out reflect.Value, in reflect.Value) {
	if in.Kind() == reflect.Slice {
		// assume in's value is an empty slice
		// which represents a null value
		// https://www.edgedb.com/docs/internals/protocol/dataformats
		return
	}
	out.Set(in)
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
