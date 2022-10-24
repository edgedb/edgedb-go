// This source file is part of the EdgeDB open source project.
//
// Copyright EdgeDB Inc. and the EdgeDB authors.
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

package errgen

import (
	"regexp"
	"strings"
)

// Type represents and EdgeDB error type.
type Type struct {
	Code      [4]uint8
	Name      string
	Ancestors []string
	Tags      []Tag
}

// PrivateName returns the private go name for this error type.
func (t *Type) PrivateName() string {
	return strings.ToLower(t.Name[0:1]) + t.Name[1:]
}

func parseType(typ []interface{}, lookup map[string]string) *Type {
	name := typ[0].(string)
	errType := &Type{
		Code: [4]uint8{
			uint8(typ[2].(float64)),
			uint8(typ[3].(float64)),
			uint8(typ[4].(float64)),
			uint8(typ[5].(float64)),
		},
		Name: name,
	}

	for _, tag := range typ[6].([]interface{}) {
		errType.Tags = append(errType.Tags, Tag(tag.(string)))
	}

	parent := lookup[name]
	for parent != "" {
		errType.Ancestors = append(errType.Ancestors, parent)
		parent = lookup[parent]
	}

	return errType
}

// ParseTypes extracts the error types from edb gen-errors-json --client output
func ParseTypes(data [][]interface{}) []*Type {
	lookup := make(map[string]string, len(data))
	for _, t := range data {
		name := t[0].(string)
		if !strings.HasSuffix(name, "Error") {
			continue
		}

		parent, _ := t[1].(string)
		lookup[name] = parent
	}

	types := make([]*Type, 0, len(data))
	for _, d := range data {
		typ := parseType(d, lookup)
		if !strings.HasSuffix(typ.Name, "Error") {
			continue
		}

		types = append(types, parseType(d, lookup))
	}

	return types
}

// Tag represents an EdgeDB error tag.
type Tag string

// Identifyer returns the MixedCaps version of the tag.
func (t Tag) Identifyer() string {
	re := regexp.MustCompile(`[A-Z]+`)

	b := re.ReplaceAllFunc([]byte(t), func(b []byte) []byte {
		s := strings.ToLower(string(b[1:]))
		return append(b[0:1], []byte(s)...)
	})

	return strings.ReplaceAll(string(b), "_", "")
}

// ParseTags returns a list of unique tags.
func ParseTags(data [][]interface{}) []Tag {
	uniqueTags := map[Tag]interface{}{}

	for _, t := range data {
		for _, tagName := range t[6].([]interface{}) {
			uniqueTags[Tag(tagName.(string))] = nil
		}
	}

	tags := make([]Tag, 0, len(uniqueTags))
	for tag := range uniqueTags {
		tags = append(tags, tag)
	}

	return tags
}
