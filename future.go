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

// Future is returned by LookupFuture and represents the promise to
// return the desired entry at a future point in time.  Callers may
// wait upon the Future at any time and retrieve the result of the
// lookup.
//
// Note that the result of calling both Wait and Channel is undefined;
// callers are strongly encouraged to call one or the other, but not
// both.
type Future struct {
	fc       *FCache      // The cache the future is from
	ent      *entry       // The actual entry in the cache
	result   <-chan Entry // Channel to receive the result
	cookie   uint64       // A unique identifier for this future
	canceled bool         // A flag indicating cancelation
}

// wait is the internal implementation of waiting on the future.
func (f *Future) wait(ctx context.Context) Entry {
	// Allow canceling from the context
	select {
	case result := <-f.result:
		return result

	case <-ctx.Done():
		return Entry{
			Error: ctx.Err(),
		}
	}
}

// WaitWithContext waits for the future to be completed and returns
// the desired result.  It accepts a context.Context to specify a
// context that may be used to cancel the wait.  Note that the future
// remains active, and a later call to Wait or WaitWithContext may be
// made.
func (f *Future) WaitWithContext(ctx context.Context) (interface{}, error) {
	// If the future has been canceled, tell them
	if f.canceled {
		return nil, ErrFutureCanceled
	}

	// If we have a result channel, simply wait on it
	if f.result != nil {
		// If the result is empty, the channel has been closed
		if ent := f.wait(ctx); ent.Object != nil || ent.Error != nil {
			return ent.Object, ent.Error
		}
	}

	// No result channel; lock the cache and get the entry
	// contents
	f.fc.Lock()
	defer f.fc.Unlock()
	if f.ent.content == nil {
		// Hmm, probably shouldn't happen...
		return nil, ErrNotCached
	}

	return f.ent.content.Object, f.ent.content.Error
}

// Wait waits for the future to be completed and returns the desired
// result.  To pass a context that may be used to cancel a wait, use
// the WaitWithContext method.
func (f *Future) Wait() (interface{}, error) {
	return f.WaitWithContext(context.Background())
}

// Cancel signals that we are not interested in the future anymore.
// Note that calling this method does not cancel any pending factory
// function calls.
func (f *Future) Cancel() {
	if !f.canceled {
		f.fc.Lock()
		defer f.fc.Unlock()
		delete(f.ent.reqs, f.cookie)
		f.result = nil
		f.canceled = true
	}
}

// Channel returns a channel that the caller may receive from to
// receive the result.
func (f *Future) Channel() <-chan Entry {
	// If the future was canceled, return nil
	if f.canceled {
		return nil
	}

	// If the result channel exists, return it
	if f.result != nil {
		return f.result
	}

	// Lock the cache
	f.fc.Lock()
	defer f.fc.Unlock()

	// If there's no content, we can't synthesize a channel
	if f.ent.content == nil {
		return nil
	}

	// Synthesize a channel with the desired contents
	result := make(chan Entry, 1)
	result <- *f.ent.content
	close(result)
	return result
}
