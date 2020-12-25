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

// LookupOption identifies an option that may be passed to the
// FCache.Lookup and FCache.Evict methods.
type LookupOption interface {
	// apply simply applies the option.
	apply(o *lookupOptions) error
}

// lookupOptions contains the consolidated options for a cache lookup
// or invalidation operation.
type lookupOptions struct {
	ent  *Entry          // Specific entry to look up or cache
	key  *Key            // Key to look up
	only bool            // Flag to allow the miss and return an error
	ctx  context.Context // Context to monitor for cancellation
}

// procLookupOpts processes a list of options and returns a
// constructed options structure.
func procLookupOpts(opts []LookupOption) (lookupOptions, error) {
	result := lookupOptions{
		ctx: context.Background(),
	}

	// Apply the options
	for _, opt := range opts {
		if err := opt.apply(&result); err != nil {
			return lookupOptions{}, err
		}
	}

	// Make sure we have a key
	if result.key == nil {
		return lookupOptions{}, ErrNoKey
	}

	return result, nil
}

// byEntryOption is a LookupOption that specifies an object to look
// up.  If the object is in the cache, the cached version (rather than
// the passed version) will be returned; otherwise, the specified
// object will be added to the cache and returned.
type byEntryOption struct {
	Ent Entry // The entry
}

// apply simply applies the option.
func (opt byEntryOption) apply(o *lookupOptions) error {
	if o.key != nil {
		return ErrDuplicateOption
	}
	o.ent = &opt.Ent
	o.key = &o.ent.Keys[0]
	return nil
}

// ByEntry returns a LookupOption that specifies an object to look up.
// If the object is in the cache, the cached version (rather than the
// passed version) will be returned; otherwise, the specified object
// will be added to the cache and returned.  For a call to Evict, the
// object itself is ignored, and only the first key is looked up and
// used to evict whatever is in the cache.
func ByEntry(ent Entry) LookupOption {
	return byEntryOption{
		Ent: ent,
	}
}

// byKeyOption is a LookupOption that specifies a cache key to look
// up.  If the object exists in the cache, the cached version will be
// returned; otherwise, the index's factory function will be called
// (unless SearchCache is also provided).
type byKeyOption struct {
	Key Key // The key
}

// apply simply applies the option.
func (opt byKeyOption) apply(o *lookupOptions) error {
	if o.key != nil {
		return ErrDuplicateOption
	}
	o.key = &opt.Key
	return nil
}

// ByKey returns a LookupOption that specifies a cache key to look up.
// If the object exists in the cache, the cached version will be
// returned; otherwise, the index's factory function will be called
// (unless SearchCache is also provided).
func ByKey(key Key) LookupOption {
	return byKeyOption{
		Key: key,
	}
}

// searchCacheOption is a LookupOption that specifies that only the
// cache should be searched.  If specified, the index factory function
// will not be called, even if the key is not found.
type searchCacheOption bool

// apply simply applies the option.
func (opt searchCacheOption) apply(o *lookupOptions) error {
	o.only = bool(opt)
	return nil
}

// SearchCache is a LookupOption that specifies that only the cache
// should be searched.  If specified, the index factory function will
// not be called, even if the key is not found.
var SearchCache searchCacheOption = true

// withContextOption is a LookupOption that specifies a
// context.Context for the lookup.
type withContextOption struct {
	Ctx context.Context // The context
}

// apply simply applies the option.
func (opt withContextOption) apply(o *lookupOptions) error {
	if o.ctx != nil {
		return ErrDuplicateOption
	}
	o.ctx = opt.Ctx
	return nil
}

// WithContext returns a LookupOption that specifies a context.Context
// for the lookup.  This option is only useful for the Lookup method;
// the LookupFuture method merely returns a Future and ignores any
// passed in context.
func WithContext(ctx context.Context) LookupOption {
	return withContextOption{
		Ctx: ctx,
	}
}

// CleanOption identifies an option that may be passed to the
// FCache.Clean method.
type CleanOption interface {
	// apply simply applies the option.
	apply(o *cleanOptions)
}

// cleanOptions contains the consolidated options for a cache cleanup
// operation.
type cleanOptions struct {
	objects bool // Clean objects from the cache
	errors  bool // Clean errors from the cache
	pending bool // Clean pending operations from the cache
}

// procCleanOpts processes a list of options and returns a constructed
// options structure.
func procCleanOpts(opts []CleanOption) cleanOptions {
	// If no options were passed, clean everything
	if len(opts) <= 0 {
		return cleanOptions{
			objects: true,
			errors:  true,
			pending: true,
		}
	}

	result := cleanOptions{}

	// Apply the options
	for _, opt := range opts {
		opt.apply(&result)
	}

	return result
}

// objectsOption is a CleanOption that specifies that only the objects
// should be cleaned from the cache.
type objectsOption bool

// apply simply applies the option.
func (opt objectsOption) apply(o *cleanOptions) {
	o.objects = bool(opt)
}

// errorsOption is a CleanOption that specifies that only the errors
// should be cleaned from the cache.
type errorsOption bool

// apply simply applies the option.
func (opt errorsOption) apply(o *cleanOptions) {
	o.errors = bool(opt)
}

// pendingOption is a CleanOption that specifies that only the pending
// operations should be cleaned from the cache.
type pendingOption bool

// apply simply applies the option.
func (opt pendingOption) apply(o *cleanOptions) {
	o.pending = bool(opt)
}

// CleanOptions that may be passed to the FCache.Clean method.
var (
	Objects objectsOption = true // Clean objects from the cache
	Errors  errorsOption  = true // Clean errors from the cache
	Pending pendingOption = true // Clean pending operations from the cache
)
