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

package gel

//go:generate go run golang.org/x/tools/cmd/stringer@v0.25.0 -type Message

// Message is a protocol message type.
type Message uint8

// Message types sent by server
const (
	Authentication         Message = 0x52
	CommandComplete        Message = 0x43
	CommandDataDescription Message = 0x54
	Data                   Message = 0x44
	DumpBlock              Message = 0x3d
	DumpHeader             Message = 0x40
	ErrorResponse          Message = 0x45
	LogMessage             Message = 0x4c
	ParameterStatus        Message = 0x53
	ParseComplete          Message = 0x31
	ReadyForCommand        Message = 0x5a
	RestoreReady           Message = 0x2b
	ServerHandshake        Message = 0x76
	ServerKeyData          Message = 0x4b
	StateDataDescription   Message = 0x73
)

// Message types sent by client
const (
	AuthenticationSASLInitialResponse Message = 0x70
	AuthenticationSASLResponse        Message = 0x72
	ClientHandshake                   Message = 0x56
	DescribeStatement                 Message = 0x44
	Dump                              Message = 0x3e
	Execute0pX                        Message = 0x45
	ExecuteScript                     Message = 0x51
	Flush                             Message = 0x48
	Execute                           Message = 0x4f
	Parse                             Message = 0x50
	Restore                           Message = 0x3c
	RestoreBlock                      Message = 0x3d
	RestoreEOF                        Message = 0x2e
	Sync                              Message = 0x53
	Terminate                         Message = 0x58
)
