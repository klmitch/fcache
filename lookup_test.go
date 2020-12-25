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

func TestFCacheManufactureBase(t *testing.T) {
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
