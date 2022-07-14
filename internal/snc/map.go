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

package snc

import "sync"

// NewServerSettings returns an empty ServerSettings.
func NewServerSettings() *ServerSettings {
	return &ServerSettings{settings: make(map[string]interface{})}
}

// ServerSettings is a concurrency safe map. A ServerSettings must
// not be copied after first use. Use NewServerSettings() instead of creating
// ServerSettings manually.
type ServerSettings struct {
	settings map[string]interface{}
	mx       sync.RWMutex
}

// GetOk returns the value for key.
func (s *ServerSettings) GetOk(key string) (interface{}, bool) {
	s.mx.RLock()
	defer s.mx.RUnlock()
	val, ok := s.settings[key]
	return val, ok
}

// Get returns the value for key.
func (s *ServerSettings) Get(key string) interface{} {
	s.mx.RLock()
	defer s.mx.RUnlock()
	return s.settings[key]
}

// Set sets the value for key.
func (s *ServerSettings) Set(key string, val interface{}) {
	s.mx.Lock()
	defer s.mx.Unlock()
	s.settings[key] = val
}
