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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFCacheManufacture(t *testing.T) {
	ctx := context.Background()
	key := Key{"one", 1}
	ent := &Entry{
		Keys: []Key{
			{"one", 1},
			{"two", 2},
			{"three", 3},
		},
	}
	factoryCalled := false
	factory := func(tCtx context.Context, tKey Key) *Entry {
		assert.Same(t, ctx, tCtx)
		assert.Equal(t, key, tKey)
		factoryCalled = true
		return ent
	}
	obj := &FCache{
		indexes: map[interface{}]index{
			"one": {
				entries: map[interface{}]*entry{},
			},
			"two": {
				entries: map[interface{}]*entry{},
			},
			"three": {
				entries: map[interface{}]*entry{},
			},
		},
	}

	obj.manufacture(ctx, key, factory)

	assert.Equal(t, &FCache{
		indexes: map[interface{}]index{
			"one": {
				entries: map[interface{}]*entry{
					1: {
						content: ent,
					},
				},
			},
			"two": {
				entries: map[interface{}]*entry{
					2: {
						content: ent,
					},
				},
			},
			"three": {
				entries: map[interface{}]*entry{
					3: {
						content: ent,
					},
				},
			},
		},
	}, obj)
	assert.True(t, factoryCalled)
}

func TestFCacheInsertBase(t *testing.T) {
	ent := &Entry{
		Object: "object",
		Keys: []Key{
			{"one", 1},
			{"two", 2},
			{"three", 3},
			{"four", 4},
			{"five", 5},
		},
	}
	obj := &FCache{
		indexes: map[interface{}]index{
			"one": {
				entries: map[interface{}]*entry{
					1: {},
				},
			},
			"two": {
				entries: map[interface{}]*entry{},
			},
			"three": {
				entries: map[interface{}]*entry{
					3: {},
				},
			},
			"five": {
				entries: map[interface{}]*entry{},
			},
		},
	}

	result := obj.insert(ent)

	assert.Equal(t, &entry{
		content: ent,
	}, result)
	assert.Equal(t, &FCache{
		indexes: map[interface{}]index{
			"one": {
				entries: map[interface{}]*entry{
					1: result,
				},
			},
			"two": {
				entries: map[interface{}]*entry{
					2: result,
				},
			},
			"three": {
				entries: map[interface{}]*entry{
					3: result,
				},
			},
			"five": {
				entries: map[interface{}]*entry{
					5: result,
				},
			},
		},
	}, obj)
}

func TestFCacheInsertError(t *testing.T) {
	ent := &Entry{
		Error: assert.AnError,
		Keys: []Key{
			{"one", 1},
			{"two", 2},
			{"three", 3},
			{"four", 4},
			{"five", 5},
		},
	}
	obj := &FCache{
		indexes: map[interface{}]index{
			"one": {
				entries: map[interface{}]*entry{
					1: {},
				},
			},
			"two": {
				entries: map[interface{}]*entry{},
			},
			"three": {
				entries: map[interface{}]*entry{
					3: {},
				},
			},
			"five": {
				entries: map[interface{}]*entry{},
			},
		},
	}

	result := obj.insert(ent)

	assert.Nil(t, result)
	assert.Equal(t, &FCache{
		indexes: map[interface{}]index{
			"one": {
				entries: map[interface{}]*entry{},
			},
			"two": {
				entries: map[interface{}]*entry{},
			},
			"three": {
				entries: map[interface{}]*entry{},
			},
			"five": {
				entries: map[interface{}]*entry{},
			},
		},
	}, obj)
}

func TestFCacheInsertPermanentError(t *testing.T) {
	ent := &Entry{
		Error: &PermanentError{assert.AnError},
		Keys: []Key{
			{"one", 1},
			{"two", 2},
			{"three", 3},
			{"four", 4},
			{"five", 5},
		},
	}
	obj := &FCache{
		indexes: map[interface{}]index{
			"one": {
				entries: map[interface{}]*entry{
					1: {},
				},
			},
			"two": {
				entries: map[interface{}]*entry{},
			},
			"three": {
				entries: map[interface{}]*entry{
					3: {},
				},
			},
			"five": {
				entries: map[interface{}]*entry{},
			},
		},
	}

	result := obj.insert(ent)

	assert.Equal(t, &entry{
		content: ent,
	}, result)
	assert.Equal(t, &FCache{
		indexes: map[interface{}]index{
			"one": {
				entries: map[interface{}]*entry{
					1: result,
				},
			},
			"two": {
				entries: map[interface{}]*entry{
					2: result,
				},
			},
			"three": {
				entries: map[interface{}]*entry{
					3: result,
				},
			},
			"five": {
				entries: map[interface{}]*entry{
					5: result,
				},
			},
		},
	}, obj)
}

func TestFCacheLookupInternalBase(t *testing.T) {
	obj := &FCache{
		indexes: map[interface{}]index{
			"one": {
				entries: map[interface{}]*entry{
					1: {
						content: &Entry{
							Object: "object",
						},
					},
				},
			},
		},
	}

	result, err := obj.lookup(lookupOptions{
		key: &Key{"one", 1},
	})

	assert.NoError(t, err)
	assert.Equal(t, &Future{
		fc:  obj,
		ent: obj.indexes["one"].entries[1],
	}, result)
}

func TestFCacheLookupInternalMissBase(t *testing.T) {
	obj := &FCache{
		indexes: map[interface{}]index{
			"one": {
				entries: map[interface{}]*entry{},
				factory: func(tCtx context.Context, tKey Key) *Entry {
					return &Entry{
						Object: "object",
						Keys:   []Key{{"one", 1}},
					}
				},
			},
		},
	}

	result, err := obj.lookup(lookupOptions{
		key: &Key{"one", 1},
	})

	assert.NoError(t, err)
	assert.NotNil(t, result)
	object, err := result.Wait()
	assert.NoError(t, err)
	assert.Equal(t, "object", object)
}

func TestFCacheLookupInternalMissWithObject(t *testing.T) {
	obj := &FCache{
		indexes: map[interface{}]index{
			"one": {
				entries: map[interface{}]*entry{},
				factory: func(tCtx context.Context, tKey Key) *Entry {
					return &Entry{
						Error: assert.AnError,
						Keys:  []Key{{"one", 1}},
					}
				},
			},
		},
	}

	result, err := obj.lookup(lookupOptions{
		ent: &Entry{
			Object: "object",
			Keys:   []Key{{"one", 1}},
		},
		key: &Key{"one", 1},
	})

	assert.NoError(t, err)
	assert.Equal(t, &Future{
		fc:  obj,
		ent: obj.indexes["one"].entries[1],
	}, result)
}

func TestFCacheLookupInternalPendingSearchOnly(t *testing.T) {
	obj := &FCache{
		indexes: map[interface{}]index{
			"one": {
				entries: map[interface{}]*entry{
					1: {},
				},
			},
		},
	}

	result, err := obj.lookup(lookupOptions{
		key:  &Key{"one", 1},
		only: true,
	})

	assert.Same(t, ErrNotCached, err)
	assert.Nil(t, result)
}

func TestFCacheLookupInternalMissSearchCache(t *testing.T) {
	obj := &FCache{
		indexes: map[interface{}]index{
			"one": {
				entries: map[interface{}]*entry{},
				factory: func(tCtx context.Context, tKey Key) *Entry {
					return &Entry{
						Object: "object",
						Keys:   []Key{{"one", 1}},
					}
				},
			},
		},
	}

	result, err := obj.lookup(lookupOptions{
		key:  &Key{"one", 1},
		only: true,
	})

	assert.Same(t, ErrNotCached, err)
	assert.Nil(t, result)
}

func TestFCacheLookupInternalBadIndex(t *testing.T) {
	obj := &FCache{
		indexes: map[interface{}]index{},
	}

	result, err := obj.lookup(lookupOptions{
		key: &Key{"one", 1},
	})

	assert.Same(t, ErrBadIndex, err)
	assert.Nil(t, result)
}

func TestFCacheLookupBase(t *testing.T) {
	obj := &FCache{
		indexes: map[interface{}]index{
			"one": {
				entries: map[interface{}]*entry{
					1: {
						content: &Entry{
							Object: "object",
						},
					},
				},
			},
		},
	}

	result, err := obj.Lookup(ByKey(Key{"one", 1}))

	assert.NoError(t, err)
	assert.Equal(t, "object", result)
}

func TestFCacheLookupBadOption(t *testing.T) {
	obj := &FCache{
		indexes: map[interface{}]index{
			"one": {
				entries: map[interface{}]*entry{
					1: {
						content: &Entry{
							Object: "object",
						},
					},
				},
			},
		},
	}

	result, err := obj.Lookup()

	assert.Same(t, ErrNoKey, err)
	assert.Nil(t, result)
}

func TestFCacheLookupLookupError(t *testing.T) {
	obj := &FCache{
		indexes: map[interface{}]index{},
	}

	result, err := obj.Lookup(ByKey(Key{"one", 1}))

	assert.Same(t, ErrBadIndex, err)
	assert.Nil(t, result)
}

func TestFCacheLookupFutureBase(t *testing.T) {
	obj := &FCache{
		indexes: map[interface{}]index{
			"one": {
				entries: map[interface{}]*entry{
					1: {
						content: &Entry{
							Object: "object",
						},
					},
				},
			},
		},
	}

	result, err := obj.LookupFuture(ByKey(Key{"one", 1}))

	assert.NoError(t, err)
	assert.Equal(t, &Future{
		fc:  obj,
		ent: obj.indexes["one"].entries[1],
	}, result)
}

func TestFCacheLookupFutureBadOption(t *testing.T) {
	obj := &FCache{
		indexes: map[interface{}]index{
			"one": {
				entries: map[interface{}]*entry{
					1: {
						content: &Entry{
							Object: "object",
						},
					},
				},
			},
		},
	}

	result, err := obj.LookupFuture()

	assert.Same(t, ErrNoKey, err)
	assert.Nil(t, result)
}
