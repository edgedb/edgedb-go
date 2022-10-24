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

package snc

import (
	"math/rand"
	"sync"
	"time"
)

// NewRand returns a Rand.
func NewRand() *Rand {
	return &Rand{rnd: rand.New(rand.NewSource(time.Now().UnixNano()))}
}

// Rand is a concurrency safe rand.Rand. Use NewRand to create a valid Rand.
type Rand struct {
	mx  sync.Mutex
	rnd *rand.Rand
}

// Intn is the same as rand.Rand.Intn
func (r *Rand) Intn(n int) int {
	r.mx.Lock()
	defer r.mx.Unlock()
	return r.rnd.Intn(n)
}

// Float64 is the same as rand.Rand.Float64
func (r *Rand) Float64() float64 {
	r.mx.Lock()
	defer r.mx.Unlock()
	return r.rnd.Float64()
}
