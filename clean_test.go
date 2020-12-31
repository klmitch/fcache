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

func TestFCacheCleanBase(t *testing.T) {
	cancel1Called := false
	cancel4Called := false
	obj := &FCache{
		indexes: map[interface{}]index{
			"one": {
				entries: map[interface{}]*entry{
					1: {
						cancel: func() {
							cancel1Called = true
						},
					},
					2: {
						content: &Entry{
							Object: "object",
						},
					},
					3: {
						content: &Entry{
							Error: assert.AnError,
						},
					},
				},
			},
			"two": {
				entries: map[interface{}]*entry{
					4: {
						cancel: func() {
							cancel4Called = true
						},
					},
					5: {
						content: &Entry{
							Object: "object",
						},
					},
					6: {
						content: &Entry{
							Error: assert.AnError,
						},
					},
				},
			},
		},
	}

	obj.Clean()

	assert.NotContains(t, obj.indexes["one"].entries, 1)
	assert.NotContains(t, obj.indexes["one"].entries, 2)
	assert.NotContains(t, obj.indexes["one"].entries, 3)
	assert.NotContains(t, obj.indexes["two"].entries, 4)
	assert.NotContains(t, obj.indexes["two"].entries, 5)
	assert.NotContains(t, obj.indexes["two"].entries, 6)
	assert.True(t, cancel1Called)
	assert.True(t, cancel4Called)
}

func TestFCacheCleanObjects(t *testing.T) {
	cancel1Called := false
	cancel4Called := false
	obj := &FCache{
		indexes: map[interface{}]index{
			"one": {
				entries: map[interface{}]*entry{
					1: {
						cancel: func() {
							cancel1Called = true
						},
					},
					2: {
						content: &Entry{
							Object: "object",
						},
					},
					3: {
						content: &Entry{
							Error: assert.AnError,
						},
					},
				},
			},
			"two": {
				entries: map[interface{}]*entry{
					4: {
						cancel: func() {
							cancel4Called = true
						},
					},
					5: {
						content: &Entry{
							Object: "object",
						},
					},
					6: {
						content: &Entry{
							Error: assert.AnError,
						},
					},
				},
			},
		},
	}

	obj.Clean(Objects)

	assert.Contains(t, obj.indexes["one"].entries, 1)
	assert.NotContains(t, obj.indexes["one"].entries, 2)
	assert.Contains(t, obj.indexes["one"].entries, 3)
	assert.Contains(t, obj.indexes["two"].entries, 4)
	assert.NotContains(t, obj.indexes["two"].entries, 5)
	assert.Contains(t, obj.indexes["two"].entries, 6)
	assert.False(t, cancel1Called)
	assert.False(t, cancel4Called)
}

func TestFCacheCleanErrors(t *testing.T) {
	cancel1Called := false
	cancel4Called := false
	obj := &FCache{
		indexes: map[interface{}]index{
			"one": {
				entries: map[interface{}]*entry{
					1: {
						cancel: func() {
							cancel1Called = true
						},
					},
					2: {
						content: &Entry{
							Object: "object",
						},
					},
					3: {
						content: &Entry{
							Error: assert.AnError,
						},
					},
				},
			},
			"two": {
				entries: map[interface{}]*entry{
					4: {
						cancel: func() {
							cancel4Called = true
						},
					},
					5: {
						content: &Entry{
							Object: "object",
						},
					},
					6: {
						content: &Entry{
							Error: assert.AnError,
						},
					},
				},
			},
		},
	}

	obj.Clean(Errors)

	assert.Contains(t, obj.indexes["one"].entries, 1)
	assert.Contains(t, obj.indexes["one"].entries, 2)
	assert.NotContains(t, obj.indexes["one"].entries, 3)
	assert.Contains(t, obj.indexes["two"].entries, 4)
	assert.Contains(t, obj.indexes["two"].entries, 5)
	assert.NotContains(t, obj.indexes["two"].entries, 6)
	assert.False(t, cancel1Called)
	assert.False(t, cancel4Called)
}

func TestFCacheCleanPending(t *testing.T) {
	cancel1Called := false
	cancel4Called := false
	obj := &FCache{
		indexes: map[interface{}]index{
			"one": {
				entries: map[interface{}]*entry{
					1: {
						cancel: func() {
							cancel1Called = true
						},
					},
					2: {
						content: &Entry{
							Object: "object",
						},
					},
					3: {
						content: &Entry{
							Error: assert.AnError,
						},
					},
				},
			},
			"two": {
				entries: map[interface{}]*entry{
					4: {
						cancel: func() {
							cancel4Called = true
						},
					},
					5: {
						content: &Entry{
							Object: "object",
						},
					},
					6: {
						content: &Entry{
							Error: assert.AnError,
						},
					},
				},
			},
		},
	}

	obj.Clean(Pending)

	assert.NotContains(t, obj.indexes["one"].entries, 1)
	assert.Contains(t, obj.indexes["one"].entries, 2)
	assert.Contains(t, obj.indexes["one"].entries, 3)
	assert.NotContains(t, obj.indexes["two"].entries, 4)
	assert.Contains(t, obj.indexes["two"].entries, 5)
	assert.Contains(t, obj.indexes["two"].entries, 6)
	assert.True(t, cancel1Called)
	assert.True(t, cancel4Called)
}
