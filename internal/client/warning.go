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

package edgedb

import (
	"errors"
	"log"
)

// Warning is used to decode warnings in the protocol.
type Warning struct {
	Code    uint32 `json:"code"`
	Message string `json:"message"`
}

// LogWarnings is an edgedb.WarningHandler that logs warnings.
func LogWarnings(errors []error) error {
	for _, err := range errors {
		log.Println("EdgeDB warning:", err.Error())
	}

	return nil
}

// WarningsAsErrors is an edgedb.WarningHandler that returns warnings as
// errors.
func WarningsAsErrors(warnings []error) error {
	return errors.Join(warnings...)
}
