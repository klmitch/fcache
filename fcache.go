// Copyright (c) 2020 Kevin L. Mitchell
//
// Licensed under the Apache License, Version 2.0 (the "License"); you
// may not use this file except in compliance with the License.  You
// may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
// implied.  See the License for the specific language governing
// permissions and limitations under the License.

// Package fcache provides an implementation of a thread-safe, context
// aware, multi-index object and error cache.
package fcache

import (
	"sync"
)

// FCache describes a future cache.  A future cache is a cache that is
// thread-safe, contains multiple indexes, and which enables
// decoupling requests from lookups.
type FCache struct {
	sync.Mutex

	indexes map[interface{}]index // The cache indexes
}

// New constructs a new FCache object and returns it.  At least one
// Index must be passed, and all indexes must define both the index
// key and the factory function to call when the requested entry does
// not exist in the cache.
func New(indexes ...Index) (*FCache, error) {
	// Make sure we have at least one index
	if len(indexes) < 1 {
		return nil, ErrMissingIndex
	}

	// Construct the cache
	fc := &FCache{
		indexes: map[interface{}]index{},
	}

	// Process all the indexes
	for _, idx := range indexes {
		if _, ok := fc.indexes[idx.Index]; ok {
			return nil, ErrDuplicateOption
		}
		if idx.Factory == nil {
			return nil, ErrMissingFactory
		}

		fc.indexes[idx.Index] = index{
			factory: idx.Factory,
			entries: map[interface{}]*entry{},
		}
	}

	return fc, nil
}
