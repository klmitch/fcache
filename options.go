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

// Option identifies an option that may be passed to the FCache.Lookup
// and FCache.Evict methods.
type Option interface {
	// apply simply applies the option.
	apply(o *options) error
}

// options contains the consolidated options for a cache lookup or
// invalidation operation.
type options struct {
	ent  *Entry          // Specific entry to look up or cache
	key  *Key            // Key to look up
	only bool            // Flag to allow the miss and return an error
	ctx  context.Context // Context to monitor for cancellation
}

// procOpts processes a list of options and returns a constructed
// options structure.
func procOpts(opts []Option) (options, error) {
	result := options{}

	// Apply the options
	for _, opt := range opts {
		if err := opt.apply(&result); err != nil {
			return options{}, err
		}
	}

	// Make sure we have a key
	if result.key == nil {
		return options{}, ErrNoKey
	}

	return result, nil
}

// ByEntryOption is an Option that specifies an object to look up.
// If the object is in the cache, the cached version (rather than the
// passed version) will be returned; otherwise, the specified object
// will be added to the cache and returned.
type ByEntryOption struct {
	Ent Entry // The entry
}

// apply simply applies the option.
func (opt ByEntryOption) apply(o *options) error {
	if o.key != nil {
		return ErrDuplicateOption
	}
	o.ent = &opt.Ent
	o.key = &o.ent.Keys[0]
	return nil
}

// ByEntry returns an Option that specifies an object to look up.  If
// the object is in the cache, the cached version (rather than the
// passed version) will be returned; otherwise, the specified object
// will be added to the cache and returned.
func ByEntry(ent Entry) ByEntryOption {
	return ByEntryOption{
		Ent: ent,
	}
}

// ByKeyOption is an Option that specifies a cache key to look up.  If
// the object exists in the cache, the cached version will be
// returned; otherwise, the index's factory function will be called
// (unless SearchCache is also provided).
type ByKeyOption struct {
	Key Key // The key
}

// apply simply applies the option.
func (opt ByKeyOption) apply(o *options) error {
	if o.key != nil {
		return ErrDuplicateOption
	}
	o.key = &opt.Key
	return nil
}

// ByKey returns an Option that specifies a cache key to look up.  If
// the object exists in the cache, the cached version will be
// returned; otherwise, the index's factory function will be called
// (unless SearchCache is also provided).
func ByKey(key Key) ByKeyOption {
	return ByKeyOption{
		Key: key,
	}
}

// SearchCache is an Option that specifies that only the cache should
// be searched.  If specified, the index factory function will not be
// called, even if the key is not found.
type SearchCache bool

// apply simply applies the option.
func (opt SearchCache) apply(o *options) error {
	o.only = bool(opt)
	return nil
}

// WithContextOption is an Option that specifies a context.Context for
// the lookup.
type WithContextOption struct {
	Ctx context.Context // The context
}

// apply simply applies the option.
func (opt WithContextOption) apply(o *options) error {
	if o.ctx != nil {
		return ErrDuplicateOption
	}
	o.ctx = opt.Ctx
	return nil
}

// WithContext returns an Option that specifies a context.Context for
// the lookup.
func WithContext(ctx context.Context) WithContextOption {
	return WithContextOption{
		Ctx: ctx,
	}
}
