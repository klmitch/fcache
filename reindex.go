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

// fillKeyMap is a helper for Reindex that fills in a keyMap given the
// list of old keys to map from.  The cache MUST be locked upon entry
// to this method.
func (fc *FCache) fillKeyMap(obj interface{}, old []Key) (map[interface{}]*keyMap, error) {
	indexes := map[interface{}]*keyMap{}
	for _, k := range old {
		// Make sure it's a defined index
		idx, ok := fc.indexes[k.Index]
		if !ok {
			return nil, ErrBadIndex
		}

		// Filter out duplications
		if _, ok := indexes[k.Index]; ok {
			return nil, ErrIncongruentKeys
		}

		// Make sure it's this entry that's there
		if ent, ok := idx.entries[k.Key]; !ok || ent.content == nil || ent.content.Object != obj {
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
// the set of new keys, detecting duplications and handling them
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
func (fc *FCache) remap(indexes map[interface{}]*keyMap, ent *Entry) {
	newEnt := &entry{
		content: ent,
	}

	// Scan through the indexes
	for _, km := range indexes {
		// Delete the old entry
		delete(km.idx.entries, km.old)

		// Check for a squatter
		e, ok := km.idx.entries[km.new]
		if ok {
			// OK, check if we can complete the
			// squatter...
			if e.content == nil {
				e.complete(ent)
				continue
			}

			// Evict the old entry
			fc.evict(e.content.Keys)
		}

		// Replace with the new entry
		km.idx.entries[km.new] = newEnt
	}
}

// Reindex reindexes an object in the cache.  It should be passed the
// object in question and the lists of old and new cache keys; these
// key lists MUST be complete.
func (fc *FCache) Reindex(obj interface{}, old []Key, new []Key) error {
	// Lock the cache
	fc.Lock()
	defer fc.Unlock()

	// Make sure the old and new key lists are congruent
	if len(old) != len(new) {
		return ErrIncongruentKeys
	}
	indexes, err := fc.fillKeyMap(obj, old)
	if err != nil {
		return err
	}
	if err = fc.finishKeyMap(indexes, new); err != nil {
		return err
	}

	// Now remap
	fc.remap(indexes, &Entry{
		Object: obj,
		Keys:   new,
	})

	return nil
}
