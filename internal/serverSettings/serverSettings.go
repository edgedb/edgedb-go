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

package serverSettings

import (
	"fmt"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/edgedb/edgedb-go/internal"
	"github.com/edgedb/edgedb-go/internal/edgedbtypes"
)

const defaultIdleConnectionTimeout = 30 * time.Second

var (
	defaultConcurrency = max(4, runtime.NumCPU())
)

func max(a, b int) int {
	if a > b {
		return a
	}

	return b
}

// ServerSettings unifies storage and access on server settings.
type ServerSettings struct {
	mu                       sync.RWMutex
	suggestedPoolConcurrency edgedbtypes.OptionalInt64
	idleConnectionTimeout    edgedbtypes.OptionalDuration
}

// GetPoolConcurrency returns a user/server provided value, or a default.
func (s *ServerSettings) GetPoolConcurrency() int {
	s.mu.RLock()
	concurrency, ok := s.suggestedPoolConcurrency.Get()
	s.mu.RUnlock()
	if ok {
		return int(concurrency)
	}
	return defaultConcurrency
}

// GetIdleConnectionTimeout returns a user/server provided value, or a default.
func (s *ServerSettings) GetIdleConnectionTimeout() time.Duration {
	s.mu.RLock()
	d, ok := s.idleConnectionTimeout.Get()
	s.mu.RUnlock()
	if ok {
		return d.ToStdDuration()
	}
	return defaultIdleConnectionTimeout
}

// Set parses and populates a user/server provided value.
func (s *ServerSettings) Set(
	key string, blob []byte, v internal.ProtocolVersion,
) error {
	switch key {
	case "suggested_pool_concurrency":
		c, err := strconv.Atoi(string(blob))
		if err != nil {
			return fmt.Errorf("invalid suggested_pool_concurrency: %w", err)
		}
		s.mu.Lock()
		s.suggestedPoolConcurrency.Set(int64(c))
		s.mu.Unlock()
	case "system_config":
		c, err := parseSystemConfig(blob, v)
		if err != nil {
			return fmt.Errorf("invalid system_config: %w", err)
		}
		s.mu.Lock()
		s.idleConnectionTimeout.Set(c.SessionIdleTimeout)
		s.mu.Unlock()
	}
	return nil
}

// WithSuggestedPoolConcurrency is used for testing only.
func WithSuggestedPoolConcurrency(c int64) *ServerSettings {
	return &ServerSettings{
		suggestedPoolConcurrency: *(&edgedbtypes.OptionalInt64{}).Set(c),
	}
}
