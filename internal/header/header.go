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

package header

import "encoding/binary"

// Header0pX is a binary protocol header
type Header0pX map[uint16][]byte

const (
	// AllowCapabilities tells the server what capabilities it should allow.
	AllowCapabilities uint16 = 0xFF04
	allCapabilities   uint64 = 0xffffffffffffffff

	// ExplicitObjectIDs tells the server not to inject object ids.
	ExplicitObjectIDs = 0xFF05

	// AllowCapabilitieTransaction represents the transaction capability
	// in the AllowCapabilities header.
	AllowCapabilitieTransaction uint64 = 0b100

	// Capabilities is returned in PrepareComplete and CommandDataDescription
	// messages.
	Capabilities uint16 = 0x1001
)

// NewAllowCapabilitiesWithout returns an AllowCapabilities header value
// with the bits set in mask masked off.
func NewAllowCapabilitiesWithout(mask uint64) []byte {
	bts := make([]byte, 8)
	binary.BigEndian.PutUint64(bts, allCapabilities^mask)
	return bts
}
