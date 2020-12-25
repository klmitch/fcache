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

import (
	"context"
	"sync/atomic"
)

// Factory describes a function that may be used to construct an
// object when it is not found in the specified index.  The factory is
// called with a context.Context object, which may be used to cancel
// the factory function, and must return the entry for the constructed
// object.  Wrapping the error in a PermanentError will cause the
// error to be cached in the index.
type Factory func(ctx context.Context, key Key) *Entry

// Key describes a cache key.  A cache key is a two-ple struct,
// consisting of the name of an index and a key within that index for
// the object.  If the key does not exist in the index, the factory
// associated with that index may be called to construct the desired
// object.
type Key struct {
	Index interface{} // Key describing the index
	Key   interface{} // Key within the index
}

// Entry describes the object (or permanent error), including its
// index keys.
type Entry struct {
	Object interface{} // The object
	Error  error       // An error encountered by the factory
	Keys   []Key       // A list of keys associated with the object
}

// Index describes an index.  At least one of these structures must be
// passed to New to construct an FCache object.  Each Index must have
// both the index key and the factory function.
type Index struct {
	Index   interface{} // Key describing the index
	Factory Factory     // The factory function for the index
}

// entry contains the internal index entry, which also contains
// information about pending requests and a cancelation function.
type entry struct {
	content *Entry                  // The contents of the entry
	reqs    map[uint64]chan<- Entry // Pending waiting requests
	cancel  context.CancelFunc      // Function to cancel request
}

// index contains a single index.  An FCache contains one or more such
// indexes.
type index struct {
	factory Factory                // The factory that fetches the object
	entries map[interface{}]*entry // The entries in the index
}

// newEntry constructs a new index entry, complete with a cancel
// function.  It does not launch the factory; the consumer must do
// that.  Returns the entry and the context to use.
func newEntry() (*entry, context.Context) {
	// Create the context
	ctx, cancelFunc := context.WithCancel(context.Background())
	return &entry{
		cancel: cancelFunc,
	}, ctx
}

// reqCounter is a counter that is atomically incremented.  It is used
// to provide a stream of cookies, which may be used for canceling
// specific requests.
var reqCounter uint64

// makeFuture constructs a Future from the entry.
func (e *entry) makeFuture(fc *FCache) *Future {
	// Make a channel if we're not completed
	var resultChan chan Entry
	var cookie uint64
	if e.content == nil {
		resultChan = make(chan Entry, 1)
		cookie = atomic.AddUint64(&reqCounter, 1)
		if e.reqs == nil {
			e.reqs = map[uint64]chan<- Entry{}
		}
		e.reqs[cookie] = resultChan
	}

	return &Future{
		fc:     fc,
		ent:    e,
		result: resultChan,
		cookie: cookie,
	}
}

// complete updates the entry with the proper contents.  It returns a
// boolean true value if the index entry should be removed, e.g., if
// the error is non-nil and is not a permanent error.  This call will
// cancel any pending operations.
func (e *entry) complete(ent *Entry) bool {
	// Do nothing if the entry is already complete
	if e.content != nil {
		return false
	}

	// Cancel any pending request
	if e.cancel != nil {
		e.cancel()
		e.cancel = nil
	}

	// Save the content
	e.content = ent

	// Pass it on to all pending requests and close the channels
	if e.reqs != nil {
		for _, req := range e.reqs {
			req <- *ent
			close(req)
		}

		// Clear the map of requests
		e.reqs = nil
	}

	// Check if the entry needs to be removed
	return ent.Error != nil && !IsPermanent(ent.Error)
}
