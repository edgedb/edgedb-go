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

package edgedb

import "fmt"

type borrowable struct {
	reason string
}

func (b *borrowable) assertUnborrowed() error {
	switch b.reason {
	case "transaction":
		return &interfaceError{
			msg: "Connection is borrowed for a transaction. " +
				"Use the methods on transaction object instead.",
		}
	case "":
		return nil
	default:
		panic(fmt.Sprintf("unexpected reason: %q", b.reason))
	}
}

func (b *borrowable) borrow(reason string) error {
	if b.reason != "" {
		msg := "connection is already borrowed for " + b.reason
		return &interfaceError{msg: msg}
	}

	if reason != "transaction" {
		panic(fmt.Sprintf("unexpected reason: %q", reason))
	}

	b.reason = reason
	return nil
}

func (b *borrowable) unborrow() {
	if b.reason == "" {
		panic("not currently borrowed, can not unborrow")
	}

	b.reason = ""
}
