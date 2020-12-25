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

// manufacture calls the index factory function.  It MUST be called as
// a goroutine.  It will invoke the factory, then lock the mutex and
// complete the appropriate entry or entries in the cache.
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
func (fc *FCache) insert(ent *Entry) *entry {
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

	return newE
}

// lookup looks up an entry in the cache and returns a Future.  The
// Lookup and LookupFuture methods use lookup to perform the actual
// lookup.
func (fc *FCache) lookup(o lookupOptions) (*Future, error) {
	// Lock the cache
	fc.Lock()
	defer fc.Unlock()

	// Look for the index
	idx, ok := fc.indexes[o.key.Index]
	if !ok {
		return nil, ErrBadIndex
	}

	// Find an existing entry, constructing it if needed
	ent, ok := idx.entries[o.key.Key]
	if !ok {
		// Not present; insert entry if one was passed
		if o.ent != nil {
			e := fc.insert(o.ent)
			return e.makeFuture(fc), nil
		}

		// Only searching the cache?
		if o.only {
			return nil, ErrNotCached
		}

		// Construct a new entry
		var ctx context.Context
		ent, ctx = newEntry()
		idx.entries[o.key.Key] = ent

		// Manufacture the entry
		go fc.manufacture(ctx, *o.key, idx.factory)
	}

	// If the entry is incomplete and we're only searching the
	// cache, stop here
	if ent.content == nil && o.only {
		return nil, ErrNotCached
	}

	// Construct and return a future
	return ent.makeFuture(fc), nil
}

// Lookup looks up an entry in the cache and returns it.  The options
// specify which entry to look up.  If necessary, the index factory
// function will be invoked to construct the entry.  This method waits
// for the object to be constructed, unless the SearchCache option is
// provided; use LookupFuture to instead return a Future.
func (fc *FCache) Lookup(opts ...LookupOption) (interface{}, error) {
	// Process the options
	o, err := procLookupOpts(opts)
	if err != nil {
		return nil, err
	}

	// Perform the lookup
	f, err := fc.lookup(o)
	if err != nil {
		return nil, err
	}

	// Wait on the future
	defer f.Cancel()
	return f.WaitWithContext(o.ctx)
}

// LookupFuture looks up an entry in the cache and returns a Future,
// which is a promise to provide the details of the entry lookup at a
// future date.  The options specify which entry to look up.  If
// necessary, the index factory function will be invoked to construct
// the entry.
func (fc *FCache) LookupFuture(opts ...LookupOption) (*Future, error) {
	// Process the options
	o, err := procLookupOpts(opts)
	if err != nil {
		return nil, err
	}

	// Perform the lookup and return the future
	return fc.lookup(o)
}
