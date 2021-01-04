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

func TestFCacheContentsBase(t *testing.T) {
	obj := &FCache{
		indexes: map[interface{}]index{
			"idx": {
				entries: map[interface{}]*entry{
					"o1": {
						content: &Entry{
							Object: "o1",
						},
					},
					"o2": {
						content: &Entry{
							Error: assert.AnError,
						},
					},
					"o3": {
						content: &Entry{
							Object: "o3",
						},
					},
					"o4": {},
				},
			},
		},
	}

	result, err := obj.Contents("idx")

	assert.NoError(t, err)
	assert.Len(t, result, 3)
	assert.ElementsMatch(t, []Entry{
		{
			Object: "o1",
		},
		{
			Error: assert.AnError,
		},
		{
			Object: "o3",
		},
	}, result)
}

func TestFCacheContentsBadIndex(t *testing.T) {
	obj := &FCache{
		indexes: map[interface{}]index{},
	}

	result, err := obj.Contents("idx")

	assert.Same(t, ErrBadIndex, err)
	assert.Nil(t, result)
}

func TestFCacheContentsFutureBase(t *testing.T) {
	obj := &FCache{
		indexes: map[interface{}]index{
			"idx": {
				entries: map[interface{}]*entry{
					"o1": {
						content: &Entry{
							Object: "o1",
						},
					},
					"o2": {
						content: &Entry{
							Error: assert.AnError,
						},
					},
					"o3": {
						content: &Entry{
							Object: "o3",
						},
					},
					"o4": {},
				},
			},
		},
	}

	result, err := obj.ContentsFuture("idx")

	assert.NoError(t, err)
	assert.Len(t, result, 4)
	for _, future := range result {
		if future.ent.content == nil {
			assert.Same(t, obj.indexes["idx"].entries["o4"], future.ent)
		} else if future.ent.content.Error != nil {
			assert.Same(t, obj.indexes["idx"].entries["o2"], future.ent)
		} else {
			assert.Same(t, obj.indexes["idx"].entries[future.ent.content.Object.(string)], future.ent)
		}
	}
}

func TestFCacheContentsFutureBadIndex(t *testing.T) {
	obj := &FCache{
		indexes: map[interface{}]index{},
	}

	result, err := obj.ContentsFuture("idx")

	assert.Same(t, ErrBadIndex, err)
	assert.Nil(t, result)
}
