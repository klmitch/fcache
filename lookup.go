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

// entryWait is a helper that simply waits on an entry.  This function
// MUST NOT be called with the cache locked.
func entryWait(ctx context.Context, ent *entry, resultChan <-chan Entry, cookie uint64) Entry {
	// Ensure we have a context
	if ctx == nil {
		ctx = context.Background()
	}

	// Allow canceling from the context
	select {
	case result := <-resultChan:
		return result

	case <-ctx.Done():
		ent.cancelReq(cookie)
		return Entry{
			Error: ctx.Err(),
		}
	}
}

// manufacture calls the index factory function.  It MUST be called as
// a goroutine, and the mutex MUST NOT be locked.  It will invoke the
// factory, then lock the mutex and complete the appropriate entry or
// entries in the cache.
func (fc *FCache) manufacture(ctx context.Context, key Key, factory Factory) {
	// Invoke the factory
	ent := factory(ctx, key)

	// Lock the cache
	fc.Lock()
	defer fc.Unlock()

	// Insert the object into the appropriate indexes
	fc.insert(ent)
}

// insert inserts the entry into the cache, constructing index entries
// as required.  The cache MUST be locked upon entry to this method.
func (fc *FCache) insert(ent *Entry) {
	// Pre-create the entry, if appropriate
	var newE *entry
	if ent.Error == nil || IsPermanent(ent.Error) {
		newE = &entry{
			content: ent,
		}
	}

	// Walk through the keys
	for _, k := range ent.Keys {
		// Skip indexes we don't know about
		idx, ok := fc.indexes[k.Index]
		if !ok {
			continue
		}

		// Complete the entry
		if e, ok := idx.entries[k.Key]; ok {
			if e.complete(ent) {
				delete(idx.entries, k.Key)
			}
		} else if ent != nil {
			idx.entries[k.Key] = newE
		}
	}
}

// Lookup looks up an entry in the cache.  The options specify which
// entry to look up.  The desired entry is returned, or the factory
// function invoked to construct it.
func (fc *FCache) Lookup(opts ...Option) (interface{}, error) {
	// Lock the cache
	fc.Lock()

	// Process the options
	o, err := procOpts(opts)
	if err != nil {
		fc.Unlock()
		return nil, err
	}

	// Look for the index
	idx, ok := fc.indexes[o.key.Index]
	if !ok {
		fc.Unlock()
		return nil, ErrBadIndex
	}

	// Find an existing entry
	ent, ok := idx.entries[o.key.Key]
	if !ok {
		// Not present; insert entry if we were passed one
		if o.ent != nil {
			fc.insert(o.ent)
			fc.Unlock()
			return o.ent.Object, o.ent.Error
		}

		// Only searching the cache?
		if o.only {
			fc.Unlock()
			return nil, ErrNotCached
		}

		// Construct a new entry
		var ctx context.Context
		ent, ctx = newEntry()
		idx.entries[o.key.Key] = ent

		// Make sure to run the factory
		go fc.manufacture(ctx, *o.key, idx.factory)
	}

	// If it's complete, return it
	if ent.content != nil {
		defer fc.Unlock()
		return ent.content.Object, ent.content.Error
	}

	// If we're only searching the cache, stop here
	if o.only {
		fc.Unlock()
		return nil, ErrNotCached
	}

	// Wait on the entry
	resultChan, cookie := ent.request()
	fc.Unlock()
	content := entryWait(o.ctx, ent, resultChan, cookie)
	return content.Object, content.Error
}
