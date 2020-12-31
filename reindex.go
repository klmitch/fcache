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

import "reflect"

// keyMap is used by Reindex to detect duplicate or missing keys and
// to capture the old and new keys.
type keyMap struct {
	idx    index       // Index
	old    interface{} // Old key
	new    interface{} // New key
	detect bool        // Used to detect incongruent keys
}

// fillKeyMap is a helper for Reindex that fills in the "old" side of
// the keyMap given the existing cache entry.  The cache MUST be
// locked upon entry to this method.
func (fc *FCache) fillKeyMap(ent *entry) (map[interface{}]*keyMap, error) {
	indexes := map[interface{}]*keyMap{}
	for _, k := range ent.content.Keys {
		// Make sure it's a defined index
		idx, ok := fc.indexes[k.Index]
		if !ok {
			continue
		}

		// Filter out duplications
		if _, ok := indexes[k.Index]; ok {
			continue
		}

		// Make sure it's this entry that's there
		if tmp, ok := idx.entries[k.Key]; !ok || !reflect.DeepEqual(ent.content, tmp.content) {
			return nil, ErrEntryNotFound
		}

		indexes[k.Index] = &keyMap{
			idx:    idx,
			old:    k.Key,
			detect: true,
		}
	}

	return indexes, nil
}

// finishKeyMap completes the work of fillKeyMap by walking through
// the list of new keys, detecting duplications and handling them
// appropriately.  The cache MUST be locked upon entry to this method.
func (fc *FCache) finishKeyMap(indexes map[interface{}]*keyMap, new []Key) error {
	for _, k := range new {
		// Detect index in new that's not in old
		km, ok := indexes[k.Index]
		if !ok || !km.detect {
			return ErrIncongruentKeys
		}

		// Ensure we skip if old and new are the same
		if reflect.DeepEqual(km.old, k.Key) {
			delete(indexes, k.Index)
			continue
		}

		// Ensure we detect duplicates in later iterations
		km.new = k.Key
		km.detect = false
	}

	return nil
}

// remap is a helper that applies the finalized key map to the cache.
// The cache MUST be locked upon entry to this method.
func (fc *FCache) remap(indexes map[interface{}]*keyMap, ent *entry) {
	keys := make([]Key, len(ent.content.Keys))

	// Step through the existing keys
	for i, k := range ent.content.Keys {
		// Handle unmutated key
		km, ok := indexes[k.Index]
		if !ok {
			keys[i] = k
			continue
		}

		// OK, construct the new key
		keys[i] = Key{
			Index: k.Index,
			Key:   km.new,
		}

		// Delete the old entry
		delete(km.idx.entries, km.old)

		// Check for a squatter
		e, ok := km.idx.entries[km.new]
		if ok {
			// Try to complete the squatter
			if e.content == nil {
				defer e.complete(ent.content)
				continue
			}

			// OK, have to evict the old entry
			fc.evict(e.content.Keys)
		}

		// Replace with the new entry
		km.idx.entries[km.new] = ent
	}

	// Update the entry keys
	ent.content.Keys = keys
}

// Reindex reindexes an existing entry in the cache--that is, it
// changes the set of old keys to a set of new keys.  It should be
// passed the list of new keys and appropriate options to find the
// object to reindex in the cache.
func (fc *FCache) Reindex(newKeys []Key, opts ...LookupOption) error {
	// Process the options
	o, err := procLookupOpts(opts)
	if err != nil {
		return err
	}

	// Lock the cache
	fc.Lock()
	defer fc.Unlock()

	// Look for the index of the primary key
	idx, ok := fc.indexes[o.key.Index]
	if !ok {
		return ErrBadIndex
	}

	// Find the existing entry
	ent, ok := idx.entries[o.key.Key]
	if !ok || ent.content == nil {
		return ErrNotCached
	}

	// Construct the initial keymap
	indexes, err := fc.fillKeyMap(ent)
	if err != nil {
		return err
	}

	// Finish the keymap construction
	if err = fc.finishKeyMap(indexes, newKeys); err != nil {
		return err
	}

	// Perform the remap
	fc.remap(indexes, ent)

	return nil
}
