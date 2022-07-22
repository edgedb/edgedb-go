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

package state

import (
	"github.com/edgedb/edgedb-go/internal/codecs"
	"github.com/edgedb/edgedb-go/internal/descriptor"
)

func buildSetOfArrayCodec(
	desc descriptor.Descriptor,
	path codecs.Path,
) (codecs.Encoder, error) {
	child, err := BuildEncoder(desc.Fields[0].Desc, path)
	if err != nil {
		return nil, err
	}

	return &setOfArrayCodec{arrayOrSetEncoder{desc.ID, child}}, nil
}

type setOfArrayCodec struct {
	arrayOrSetEncoder
}
