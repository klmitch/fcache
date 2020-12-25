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

// evict clears entries from the cache.  The cache MUST be locked upon
// entry to this method.
func (fc *FCache) evict(keys []Key) {
	// Walk through the keys
	for _, k := range keys {
		// Skip indexes we don't know about
		idx, ok := fc.indexes[k.Index]
		if !ok {
			continue
		}

		// Clear out only completed entries
		if e, ok := idx.entries[k.Key]; ok && e.content != nil {
			delete(idx.entries, k.Key)
		}
	}
}

// Evict removes a specific entry in the cache.  The options specify
// which entry to evict.
func (fc *FCache) Evict(opts ...LookupOption) error {
	// Lock the cache
	fc.Lock()
	defer fc.Unlock()

	// Process the options
	o, err := procLookupOpts(opts)
	if err != nil {
		return err
	}

	// Look for the index
	idx, ok := fc.indexes[o.key.Index]
	if !ok {
		return ErrBadIndex
	}

	// Check to see if there's an entry
	ent, ok := idx.entries[o.key.Key]
	if !ok || ent.content == nil {
		// Not present, do nothing
		return nil
	}

	// Evict the entry
	fc.evict(ent.content.Keys)

	return nil
}
