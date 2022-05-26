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

package message

// Message types sent by server
const (
	Authentication         = 0x52
	CommandComplete        = 0x43
	CommandDataDescription = 0x54
	Data                   = 0x44
	DumpBlock              = 0x3d
	DumpHeader             = 0x40
	ErrorResponse          = 0x45
	LogMessage             = 0x4c
	ParameterStatus        = 0x53
	ParseComplete          = 0x31
	ReadyForCommand        = 0x5a
	RestoreReady           = 0x2b
	ServerHandshake        = 0x76
	ServerKeyData          = 0x4b
)

// Message types sent by client
const (
	AuthenticationSASLInitialResponse = 0x70
	AuthenticationSASLResponse        = 0x72
	ClientHandshake                   = 0x56
	DescribeStatement                 = 0x44
	Dump                              = 0x3e
	Execute0pX                        = 0x45
	ExecuteScript                     = 0x51
	Flush                             = 0x48
	Execute                           = 0x4f
	Parse                             = 0x50
	Restore                           = 0x3c
	RestoreBlock                      = 0x3d
	RestoreEOF                        = 0x2e
	Sync                              = 0x53
	Terminate                         = 0x58
)
