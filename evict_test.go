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

func TestFCacheEvictInternal(t *testing.T) {
	keys := []Key{
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
					1: {
						content: &Entry{},
					},
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
				entries: map[interface{}]*entry{
					5: {
						content: &Entry{},
					},
				},
			},
		},
	}

	obj.evict(keys)

	assert.Equal(t, &FCache{
		indexes: map[interface{}]index{
			"one": {
				entries: map[interface{}]*entry{},
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
	}, obj)
}

func TestFCacheEvictBase(t *testing.T) {
	obj := &FCache{
		indexes: map[interface{}]index{
			"one": {
				entries: map[interface{}]*entry{},
			},
		},
	}

	err := obj.Evict(ByKey(Key{"one", 1}))

	assert.NoError(t, err)
	assert.Equal(t, map[interface{}]*entry{}, obj.indexes["one"].entries)
}

func TestFCacheEvictPending(t *testing.T) {
	obj := &FCache{
		indexes: map[interface{}]index{
			"one": {
				entries: map[interface{}]*entry{
					1: {},
				},
			},
		},
	}

	err := obj.Evict(ByKey(Key{"one", 1}))

	assert.NoError(t, err)
	assert.Equal(t, map[interface{}]*entry{
		1: {},
	}, obj.indexes["one"].entries)
}

func TestFCacheEvictCached(t *testing.T) {
	obj := &FCache{
		indexes: map[interface{}]index{
			"one": {
				entries: map[interface{}]*entry{
					1: {
						content: &Entry{
							Keys: []Key{{"one", 1}},
						},
					},
				},
			},
		},
	}

	err := obj.Evict(ByKey(Key{"one", 1}))

	assert.NoError(t, err)
	assert.Equal(t, map[interface{}]*entry{}, obj.indexes["one"].entries)
}

func TestFCacheEvictBadOption(t *testing.T) {
	obj := &FCache{
		indexes: map[interface{}]index{
			"one": {
				entries: map[interface{}]*entry{},
			},
		},
	}

	err := obj.Evict()

	assert.Same(t, ErrNoKey, err)
	assert.Equal(t, map[interface{}]*entry{}, obj.indexes["one"].entries)
}

func TestFCacheEvictBadIndex(t *testing.T) {
	obj := &FCache{
		indexes: map[interface{}]index{},
	}

	err := obj.Evict(ByKey(Key{"one", 1}))

	assert.Same(t, ErrBadIndex, err)
}
