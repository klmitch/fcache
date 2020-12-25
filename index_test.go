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

	"github.com/klmitch/patcher"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewEntry(t *testing.T) {
	ent, ctx := newEntry()

	assert.NotNil(t, ent)
	assert.NotNil(t, ctx)
	assert.NotNil(t, ent.cancel)
}

func TestEntryMakeFutureBase(t *testing.T) {
	fc := &FCache{}
	obj := &entry{
		content: &Entry{},
	}
	defer patcher.SetVar(&reqCounter, uint64(42)).Install().Restore()

	result := obj.makeFuture(fc)

	assert.Nil(t, obj.reqs)
	assert.Same(t, fc, result.fc)
	assert.Same(t, obj, result.ent)
	assert.Nil(t, result.result)
}

func TestEntryMakeFutureIncomplete(t *testing.T) {
	fc := &FCache{}
	obj := &entry{
		reqs: map[uint64]chan<- Entry{
			17: nil,
		},
	}
	defer patcher.SetVar(&reqCounter, uint64(42)).Install().Restore()

	result := obj.makeFuture(fc)

	require.NotNil(t, obj.reqs)
	assert.Contains(t, obj.reqs, uint64(17))
	assert.Contains(t, obj.reqs, result.cookie)
	assert.Same(t, fc, result.fc)
	assert.Same(t, obj, result.ent)
	assert.NotNil(t, result.result)
}

func TestEntryMakeFutureIncompleteMakeReqs(t *testing.T) {
	fc := &FCache{}
	obj := &entry{}
	defer patcher.SetVar(&reqCounter, uint64(42)).Install().Restore()

	result := obj.makeFuture(fc)

	require.NotNil(t, obj.reqs)
	assert.Contains(t, obj.reqs, result.cookie)
	assert.Same(t, fc, result.fc)
	assert.Same(t, obj, result.ent)
	assert.NotNil(t, result.result)
}

func TestEntryCompleteBase(t *testing.T) {
	ent := &Entry{}
	obj := &entry{}

	result := obj.complete(ent)

	assert.False(t, result)
	assert.Same(t, ent, obj.content)
	assert.Nil(t, obj.cancel)
	assert.Nil(t, obj.reqs)
}

func TestEntryCompleteError(t *testing.T) {
	ent := &Entry{
		Error: assert.AnError,
	}
	obj := &entry{}

	result := obj.complete(ent)

	assert.True(t, result)
	assert.Same(t, ent, obj.content)
	assert.Nil(t, obj.cancel)
	assert.Nil(t, obj.reqs)
}

func TestEntryCompletePermanentError(t *testing.T) {
	ent := &Entry{
		Error: &PermanentError{Err: assert.AnError},
	}
	obj := &entry{}

	result := obj.complete(ent)

	assert.False(t, result)
	assert.Same(t, ent, obj.content)
	assert.Nil(t, obj.cancel)
	assert.Nil(t, obj.reqs)
}

func TestEntryCompleteCancel(t *testing.T) {
	ent := &Entry{}
	canceled := false
	obj := &entry{
		cancel: func() {
			canceled = true
		},
	}

	result := obj.complete(ent)

	assert.False(t, result)
	assert.Same(t, ent, obj.content)
	assert.Nil(t, obj.cancel)
	assert.Nil(t, obj.reqs)
	assert.True(t, canceled)
}

func TestEntryCompleteFulfillRequests(t *testing.T) {
	ent := &Entry{
		Error: assert.AnError,
	}
	reqs := map[uint64]chan Entry{
		1: make(chan Entry, 1),
		7: make(chan Entry, 1),
		9: make(chan Entry, 1),
	}
	obj := &entry{
		reqs: map[uint64]chan<- Entry{},
	}
	for cookie, req := range reqs {
		obj.reqs[cookie] = req
	}

	result := obj.complete(ent)

	assert.True(t, result)
	assert.Same(t, ent, obj.content)
	assert.Nil(t, obj.cancel)
	assert.Nil(t, obj.reqs)
	for _, req := range reqs {
		response, ok := <-req
		if !ok {
			assert.Fail(t, "Failed to receive response")
			continue
		}
		assert.Equal(t, Entry{
			Error: assert.AnError,
		}, response)
		_, ok = <-req
		if ok {
			assert.Fail(t, "Failed to close request channel")
		}
	}
}

func TestEntryCompleteCompleted(t *testing.T) {
	ent := &Entry{}
	obj := &entry{
		content: &Entry{
			Error: assert.AnError,
		},
	}

	result := obj.complete(ent)

	assert.False(t, result)
	assert.NotSame(t, ent, obj.content)
	assert.Equal(t, &Entry{
		Error: assert.AnError,
	}, obj.content)
	assert.Nil(t, obj.cancel)
	assert.Nil(t, obj.reqs)
}
