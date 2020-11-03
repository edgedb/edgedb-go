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
	"fmt"
	"reflect"
)

func fieldByTag(t reflect.Type, name string) (reflect.StructField, bool) {
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.Tag.Get("edgedb") == name {
			return field, true
		}
	}

	return reflect.StructField{}, false
}

func StructField(t reflect.Type, name string) (reflect.StructField, bool) {
	if f, ok := fieldByTag(t, name); ok {
		return f, true
	}

	if f, ok := t.FieldByName(name); ok {
		return f, true
	}

	return reflect.StructField{}, false
}

func ValueOf(i interface{}) (reflect.Value, error) {
	v := reflect.ValueOf(i)
	if v.Kind() != reflect.Ptr {
		return reflect.Value{}, fmt.Errorf(
			"out must be a pointer, got %v",
			v.Type(),
		)
	}

	e := v.Elem()
	if !e.IsValid() {
		return reflect.Value{}, fmt.Errorf(
			"out must point to a valid value, got %v",
			i,
		)
	}

	return e, nil
}

func ValueOfSlice(i interface{}) (reflect.Value, error) {
	v, err := ValueOf(i)
	if err != nil {
		return v, err
	}

	if v.Kind() != reflect.Slice {
		return reflect.Value{}, fmt.Errorf(
			"out must be a pointer to a slice, got %v",
			reflect.ValueOf(i).Type(),
		)
	}

	return v, nil
}
