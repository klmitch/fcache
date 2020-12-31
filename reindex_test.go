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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFCacheFillKeyMapBase(t *testing.T) {
	ent := &entry{
		content: &Entry{
			Object: "object",
			Keys: []Key{
				{"one", 1},
				{"two", 2},
				{"three", 3},
				{"four", 4},
				{"one", 1},
			},
		},
	}
	obj := &FCache{
		indexes: map[interface{}]index{
			"one": {
				entries: map[interface{}]*entry{
					1: ent,
				},
			},
			"two": {
				entries: map[interface{}]*entry{
					2: ent,
				},
			},
			"three": {
				entries: map[interface{}]*entry{
					3: {
						content: ent.content,
					},
				},
			},
		},
	}

	result, err := obj.fillKeyMap(ent)

	assert.NoError(t, err)
	assert.Equal(t, map[interface{}]*keyMap{
		"one": {
			idx:    obj.indexes["one"],
			old:    1,
			detect: true,
		},
		"two": {
			idx:    obj.indexes["two"],
			old:    2,
			detect: true,
		},
		"three": {
			idx:    obj.indexes["three"],
			old:    3,
			detect: true,
		},
	}, result)
}

func TestFCacheFillKeyMapMissingEntry(t *testing.T) {
	ent := &entry{
		content: &Entry{
			Object: "object",
			Keys: []Key{
				{"one", 1},
				{"two", 2},
				{"three", 3},
				{"four", 4},
				{"one", 1},
			},
		},
	}
	obj := &FCache{
		indexes: map[interface{}]index{
			"one": {
				entries: map[interface{}]*entry{
					1: ent,
				},
			},
			"two": {
				entries: map[interface{}]*entry{
					2: ent,
				},
			},
			"three": {
				entries: map[interface{}]*entry{},
			},
		},
	}

	result, err := obj.fillKeyMap(ent)

	assert.Same(t, ErrEntryNotFound, err)
	assert.Nil(t, result)
}

func TestFCacheFillKeyMapPendingEntry(t *testing.T) {
	ent := &entry{
		content: &Entry{
			Object: "object",
			Keys: []Key{
				{"one", 1},
				{"two", 2},
				{"three", 3},
				{"four", 4},
				{"one", 1},
			},
		},
	}
	obj := &FCache{
		indexes: map[interface{}]index{
			"one": {
				entries: map[interface{}]*entry{
					1: ent,
				},
			},
			"two": {
				entries: map[interface{}]*entry{
					2: ent,
				},
			},
			"three": {
				entries: map[interface{}]*entry{
					3: {},
				},
			},
		},
	}

	result, err := obj.fillKeyMap(ent)

	assert.Same(t, ErrEntryNotFound, err)
	assert.Nil(t, result)
}

func TestFCacheFillKeyMapMissmatchedEntry(t *testing.T) {
	ent := &entry{
		content: &Entry{
			Object: "object",
			Keys: []Key{
				{"one", 1},
				{"two", 2},
				{"three", 3},
				{"four", 4},
				{"one", 1},
			},
		},
	}
	obj := &FCache{
		indexes: map[interface{}]index{
			"one": {
				entries: map[interface{}]*entry{
					1: ent,
				},
			},
			"two": {
				entries: map[interface{}]*entry{
					2: ent,
				},
			},
			"three": {
				entries: map[interface{}]*entry{
					3: {
						content: &Entry{
							Error: assert.AnError,
							Keys:  []Key{{"three", 3}},
						},
					},
				},
			},
		},
	}

	result, err := obj.fillKeyMap(ent)

	assert.Same(t, ErrEntryNotFound, err)
	assert.Nil(t, result)
}

func TestFCacheFinishKeyMapBase(t *testing.T) {
	obj := &FCache{}
	indexes := map[interface{}]*keyMap{
		"one": {
			old:    1,
			detect: true,
		},
		"two": {
			old:    2,
			detect: true,
		},
		"three": {
			old:    3,
			detect: true,
		},
	}

	err := obj.finishKeyMap(indexes, []Key{
		{"one", 3},
		{"two", 2},
		{"three", 1},
	})

	assert.NoError(t, err)
	assert.Equal(t, map[interface{}]*keyMap{
		"one": {
			old: 1,
			new: 3,
		},
		"three": {
			old: 3,
			new: 1,
		},
	}, indexes)
}

func TestFCacheFinishKeyMapIncongruentIndex(t *testing.T) {
	obj := &FCache{}
	indexes := map[interface{}]*keyMap{
		"one": {
			old:    1,
			detect: true,
		},
		"two": {
			old:    2,
			detect: true,
		},
		"three": {
			old:    3,
			detect: true,
		},
	}

	err := obj.finishKeyMap(indexes, []Key{
		{"one", 3},
		{"two", 2},
		{"four", 4},
	})

	assert.Same(t, ErrIncongruentKeys, err)
}

func TestFCacheFinishKeyMapDuplicateKey(t *testing.T) {
	obj := &FCache{}
	indexes := map[interface{}]*keyMap{
		"one": {
			old:    1,
			detect: true,
		},
		"two": {
			old:    2,
			detect: true,
		},
		"three": {
			old:    3,
			detect: true,
		},
	}

	err := obj.finishKeyMap(indexes, []Key{
		{"one", 3},
		{"one", 2},
		{"three", 4},
	})

	assert.Same(t, ErrIncongruentKeys, err)
}

func TestFCacheRemap(t *testing.T) {
	ent := &entry{
		content: &Entry{
			Object: "object",
			Keys: []Key{
				{"one", 8},
				{"two", 7},
				{"three", 6},
				{"four", 5},
				{"five", 4},
			},
		},
	}
	existing := &entry{
		content: &Entry{
			Object: "existing",
			Keys: []Key{
				{"one", 1},
				{"two", 2},
			},
		},
	}
	obj := &FCache{
		indexes: map[interface{}]index{
			"one": {
				entries: map[interface{}]*entry{
					1: existing,
					8: ent,
				},
			},
			"two": {
				entries: map[interface{}]*entry{
					2: existing,
					7: ent,
				},
			},
			"three": {
				entries: map[interface{}]*entry{
					3: {},
					6: ent,
				},
			},
			"four": {
				entries: map[interface{}]*entry{
					5: ent,
				},
			},
			"five": {
				entries: map[interface{}]*entry{
					4: ent,
				},
			},
		},
	}
	indexes := map[interface{}]*keyMap{
		"one": {
			idx: obj.indexes["one"],
			old: 8,
			new: 1,
		},
		"two": {
			idx: obj.indexes["two"],
			old: 7,
			new: 2,
		},
		"three": {
			idx: obj.indexes["three"],
			old: 6,
			new: 3,
		},
		"four": {
			idx: obj.indexes["four"],
			old: 5,
			new: 4,
		},
	}

	obj.remap(indexes, ent)

	assert.Equal(t, &FCache{
		indexes: map[interface{}]index{
			"one": {
				entries: map[interface{}]*entry{
					1: ent,
				},
			},
			"two": {
				entries: map[interface{}]*entry{
					2: ent,
				},
			},
			"three": {
				entries: map[interface{}]*entry{
					3: ent,
				},
			},
			"four": {
				entries: map[interface{}]*entry{
					4: ent,
				},
			},
			"five": {
				entries: map[interface{}]*entry{
					4: ent,
				},
			},
		},
	}, obj)
}

func TestFCacheReindexBase(t *testing.T) {
	object := &entry{
		content: &Entry{
			Object: "object",
			Keys: []Key{
				{"one", 4},
				{"two", 3},
				{"three", 2},
				{"four", 1},
			},
		},
	}
	newKeys := []Key{
		{"one", 1},
		{"two", 2},
		{"three", 3},
		{"four", 4},
	}
	obj := &FCache{
		indexes: map[interface{}]index{
			"one": {
				entries: map[interface{}]*entry{
					4: object,
				},
			},
			"two": {
				entries: map[interface{}]*entry{
					3: object,
				},
			},
			"three": {
				entries: map[interface{}]*entry{
					2: object,
				},
			},
			"four": {
				entries: map[interface{}]*entry{
					1: object,
				},
			},
		},
	}

	err := obj.Reindex(newKeys, ByKey(Key{"one", 4}))

	assert.NoError(t, err)
	assert.Equal(t, &FCache{
		indexes: map[interface{}]index{
			"one": {
				entries: map[interface{}]*entry{
					1: object,
				},
			},
			"two": {
				entries: map[interface{}]*entry{
					2: object,
				},
			},
			"three": {
				entries: map[interface{}]*entry{
					3: object,
				},
			},
			"four": {
				entries: map[interface{}]*entry{
					4: object,
				},
			},
		},
	}, obj)
}

func TestFCacheReindexBadOption(t *testing.T) {
	object := &entry{
		content: &Entry{
			Object: "object",
			Keys: []Key{
				{"one", 4},
				{"two", 3},
				{"three", 2},
				{"four", 1},
			},
		},
	}
	newKeys := []Key{
		{"one", 1},
		{"two", 2},
		{"three", 3},
		{"four", 4},
	}
	obj := &FCache{
		indexes: map[interface{}]index{
			"one": {
				entries: map[interface{}]*entry{
					4: object,
				},
			},
			"two": {
				entries: map[interface{}]*entry{
					3: object,
				},
			},
			"three": {
				entries: map[interface{}]*entry{
					2: object,
				},
			},
			"four": {
				entries: map[interface{}]*entry{
					1: object,
				},
			},
		},
	}

	err := obj.Reindex(newKeys)

	assert.Same(t, ErrNoKey, err)
	assert.Equal(t, &FCache{
		indexes: map[interface{}]index{
			"one": {
				entries: map[interface{}]*entry{
					4: object,
				},
			},
			"two": {
				entries: map[interface{}]*entry{
					3: object,
				},
			},
			"three": {
				entries: map[interface{}]*entry{
					2: object,
				},
			},
			"four": {
				entries: map[interface{}]*entry{
					1: object,
				},
			},
		},
	}, obj)
}

func TestFCacheReindexBadIndex(t *testing.T) {
	object := &entry{
		content: &Entry{
			Object: "object",
			Keys: []Key{
				{"one", 4},
				{"two", 3},
				{"three", 2},
				{"four", 1},
			},
		},
	}
	newKeys := []Key{
		{"one", 1},
		{"two", 2},
		{"three", 3},
		{"four", 4},
	}
	obj := &FCache{
		indexes: map[interface{}]index{
			"one": {
				entries: map[interface{}]*entry{
					4: object,
				},
			},
			"two": {
				entries: map[interface{}]*entry{
					3: object,
				},
			},
			"three": {
				entries: map[interface{}]*entry{
					2: object,
				},
			},
			"four": {
				entries: map[interface{}]*entry{
					1: object,
				},
			},
		},
	}

	err := obj.Reindex(newKeys, ByKey(Key{"five", 5}))

	assert.Same(t, ErrBadIndex, err)
	assert.Equal(t, &FCache{
		indexes: map[interface{}]index{
			"one": {
				entries: map[interface{}]*entry{
					4: object,
				},
			},
			"two": {
				entries: map[interface{}]*entry{
					3: object,
				},
			},
			"three": {
				entries: map[interface{}]*entry{
					2: object,
				},
			},
			"four": {
				entries: map[interface{}]*entry{
					1: object,
				},
			},
		},
	}, obj)
}

func TestFCacheReindexNotCached(t *testing.T) {
	object := &entry{
		content: &Entry{
			Object: "object",
			Keys: []Key{
				{"one", 4},
				{"two", 3},
				{"three", 2},
				{"four", 1},
			},
		},
	}
	newKeys := []Key{
		{"one", 1},
		{"two", 2},
		{"three", 3},
		{"four", 4},
	}
	obj := &FCache{
		indexes: map[interface{}]index{
			"one": {
				entries: map[interface{}]*entry{
					4: object,
				},
			},
			"two": {
				entries: map[interface{}]*entry{
					3: object,
				},
			},
			"three": {
				entries: map[interface{}]*entry{
					2: object,
				},
			},
			"four": {
				entries: map[interface{}]*entry{
					1: object,
				},
			},
		},
	}

	err := obj.Reindex(newKeys, ByKey(Key{"one", 1}))

	assert.Same(t, ErrNotCached, err)
	assert.Equal(t, &FCache{
		indexes: map[interface{}]index{
			"one": {
				entries: map[interface{}]*entry{
					4: object,
				},
			},
			"two": {
				entries: map[interface{}]*entry{
					3: object,
				},
			},
			"three": {
				entries: map[interface{}]*entry{
					2: object,
				},
			},
			"four": {
				entries: map[interface{}]*entry{
					1: object,
				},
			},
		},
	}, obj)
}

func TestFCacheReindexMissingEntry(t *testing.T) {
	object := &entry{
		content: &Entry{
			Object: "object",
			Keys: []Key{
				{"one", 4},
				{"two", 3},
				{"three", 2},
				{"four", 1},
			},
		},
	}
	newKeys := []Key{
		{"one", 1},
		{"two", 2},
		{"three", 3},
		{"four", 4},
	}
	obj := &FCache{
		indexes: map[interface{}]index{
			"one": {
				entries: map[interface{}]*entry{
					4: object,
				},
			},
			"two": {
				entries: map[interface{}]*entry{
					3: object,
				},
			},
			"three": {
				entries: map[interface{}]*entry{
					2: object,
				},
			},
			"four": {
				entries: map[interface{}]*entry{},
			},
		},
	}

	err := obj.Reindex(newKeys, ByKey(Key{"one", 4}))

	assert.Same(t, ErrEntryNotFound, err)
	assert.Equal(t, &FCache{
		indexes: map[interface{}]index{
			"one": {
				entries: map[interface{}]*entry{
					4: object,
				},
			},
			"two": {
				entries: map[interface{}]*entry{
					3: object,
				},
			},
			"three": {
				entries: map[interface{}]*entry{
					2: object,
				},
			},
			"four": {
				entries: map[interface{}]*entry{},
			},
		},
	}, obj)
}

func TestFCacheReindexIncongruentKeys(t *testing.T) {
	object := &entry{
		content: &Entry{
			Object: "object",
			Keys: []Key{
				{"one", 4},
				{"two", 3},
				{"three", 2},
				{"four", 1},
			},
		},
	}
	newKeys := []Key{
		{"one", 1},
		{"two", 2},
		{"three", 3},
		{"four", 4},
		{"five", 5},
	}
	obj := &FCache{
		indexes: map[interface{}]index{
			"one": {
				entries: map[interface{}]*entry{
					4: object,
				},
			},
			"two": {
				entries: map[interface{}]*entry{
					3: object,
				},
			},
			"three": {
				entries: map[interface{}]*entry{
					2: object,
				},
			},
			"four": {
				entries: map[interface{}]*entry{
					1: object,
				},
			},
		},
	}

	err := obj.Reindex(newKeys, ByKey(Key{"one", 4}))

	assert.Same(t, ErrIncongruentKeys, err)
	assert.Equal(t, &FCache{
		indexes: map[interface{}]index{
			"one": {
				entries: map[interface{}]*entry{
					4: object,
				},
			},
			"two": {
				entries: map[interface{}]*entry{
					3: object,
				},
			},
			"three": {
				entries: map[interface{}]*entry{
					2: object,
				},
			},
			"four": {
				entries: map[interface{}]*entry{
					1: object,
				},
			},
		},
	}, obj)
}
