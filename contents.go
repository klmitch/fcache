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

// Contents returns all completed entries in the specified cache
// index.  Only completed entries are returned; any uncompleted
// entries are skipped.  What is returned is a list of Entry
// structures; this allows Contents to return cached errors.
func (fc *FCache) Contents(index interface{}) ([]Entry, error) {
	// Lock the cache
	fc.Lock()
	defer fc.Unlock()

	// Look for the index
	idx, ok := fc.indexes[index]
	if !ok {
		return nil, ErrBadIndex
	}

	// Initialize a container for entries
	result := make([]Entry, 0, len(idx.entries))
	for _, ent := range idx.entries {
		if ent.content != nil {
			result = append(result, *ent.content)
		}
	}

	return result, nil
}

// ContentsFuture is similar to Contents, but returns Future instances
// for all entries in the specified cache index, including pending
// entries.
func (fc *FCache) ContentsFuture(index interface{}) ([]*Future, error) {
	// Lock the cache
	fc.Lock()
	defer fc.Unlock()

	// Look for the index
	idx, ok := fc.indexes[index]
	if !ok {
		return nil, ErrBadIndex
	}

	// Initialize a container for the futures
	result := make([]*Future, 0, len(idx.entries))
	for _, ent := range idx.entries {
		result = append(result, ent.makeFuture(fc))
	}

	return result, nil
}
