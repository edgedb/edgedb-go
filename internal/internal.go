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

package internal

// ProtocolVersion represents an EdgeDB protocol version
type ProtocolVersion struct {
	Major uint16
	Minor uint16
}

// GT returns true if v > other
func (v ProtocolVersion) GT(other ProtocolVersion) bool {
	switch {
	case v.Major > other.Major:
		return true
	case v.Major < other.Major:
		return false
	default:
		return v.Minor > other.Minor
	}
}

// GTE returns true if v >= other
func (v ProtocolVersion) GTE(other ProtocolVersion) bool {
	if v == other {
		return true
	}

	return v.GT(other)
}

// LT returns true if v < other
func (v ProtocolVersion) LT(other ProtocolVersion) bool {
	switch {
	case v.Major < other.Major:
		return true
	case v.Major > other.Major:
		return false
	default:
		return v.Minor < other.Minor
	}
}
