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

package fcache

import "context"

// CleanFlags is a set of bit flags that may be passed to FCache.Clean to indicate what should be cleaned from the cache.
type CleanFlags uint8

// Valid flags for CleanFlags.
const (
	CleanObjects CleanFlags = 1 << iota // Clean objects from cache
	CleanErrors                         // Clean cached errors
	CleanPending                        // Clean pending requests and cancel them
)

// Clean is used to clean things out of the cache.  The specific
// things to clean up are specified through the CleanFlags bitmask
// that is passed in.
func (fc *FCache) Clean(flags CleanFlags) {
	// Lock the cache
	fc.Lock()
	defer fc.Unlock()

	// Clear the desired objects
	for _, idx := range fc.indexes {
		// Find the entries to remove
		toRemove := []interface{}{}
		for key, ent := range idx.entries {
			if ent.content == nil {
				if (flags & CleanPending) != 0 {
					ent.complete(&Entry{
						Error: context.Canceled,
					})
					toRemove = append(toRemove, key)
				}
			} else if (flags&CleanObjects) != 0 && ent.content.Object != nil {
				toRemove = append(toRemove, key)
			} else if (flags&CleanErrors) != 0 && ent.content.Error != nil {
				toRemove = append(toRemove, key)
			}
		}

		// Remove the entries
		for _, k := range toRemove {
			delete(idx.entries, k)
		}
	}
}
