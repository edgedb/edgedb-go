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

package buff

import "sync"

// DoneReadingSignal is a convenient type to use with buff.Reader.Next()
type DoneReadingSignal struct {
	Chan chan struct{}
	once sync.Once
}

// NewSignal returns a new DoneReadingSignal.
// Only use the returned object once per Reader.Next() for loop.
func NewSignal() *DoneReadingSignal {
	return &DoneReadingSignal{Chan: make(chan struct{}, 1)}
}

// Signal sends on Chan the first time Signal is called.
// Subsequent calls are no-op.
func (d *DoneReadingSignal) Signal() {
	d.once.Do(func() { d.Chan <- struct{}{} })
}
