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

func TestFutureWaitInternalBase(t *testing.T) {
	resultChan := make(chan Entry, 1)
	resultChan <- Entry{
		Error: assert.AnError,
	}
	ctx := context.Background()
	obj := &Future{
		result: resultChan,
	}

	result := obj.wait(ctx)

	assert.Equal(t, Entry{
		Error: assert.AnError,
	}, result)
}

func TestFutureWaitInternalCanceled(t *testing.T) {
	resultChan := make(chan Entry, 1)
	ctx, cancelFunc := context.WithCancel(context.Background())
	cancelFunc()
	obj := &Future{
		result: resultChan,
	}

	result := obj.wait(ctx)

	assert.Equal(t, Entry{
		Error: context.Canceled,
	}, result)
}

func TestFutureWaitWithContextBase(t *testing.T) {
	resultChan := make(chan Entry, 1)
	resultChan <- Entry{
		Object: "object",
		Error:  assert.AnError,
	}
	ctx := context.Background()
	obj := &Future{
		result: resultChan,
	}

	result, err := obj.WaitWithContext(ctx)

	assert.Same(t, assert.AnError, err)
	assert.Equal(t, "object", result)
	assert.Nil(t, obj.result)
}

func TestFutureWaitWithContextComplete(t *testing.T) {
	ctx := context.Background()
	obj := &Future{
		fc: &FCache{},
		ent: &entry{
			content: &Entry{
				Object: "object",
				Error:  assert.AnError,
			},
		},
	}

	result, err := obj.WaitWithContext(ctx)

	assert.Same(t, assert.AnError, err)
	assert.Equal(t, "object", result)
}

func TestFutureWaitWithContextCanceled(t *testing.T) {
	ctx := context.Background()
	obj := &Future{
		canceled: true,
	}

	result, err := obj.WaitWithContext(ctx)

	assert.Same(t, ErrFutureCanceled, err)
	assert.Nil(t, result)
}

func TestFutureWaitWithContextUncached(t *testing.T) {
	ctx := context.Background()
	obj := &Future{
		fc:  &FCache{},
		ent: &entry{},
	}

	result, err := obj.WaitWithContext(ctx)

	assert.Same(t, ErrNotCached, err)
	assert.Nil(t, result)
}

func TestFutureWait(t *testing.T) {
	resultChan := make(chan Entry, 1)
	resultChan <- Entry{
		Object: "object",
		Error:  assert.AnError,
	}
	obj := &Future{
		result: resultChan,
	}

	result, err := obj.Wait()

	assert.Same(t, assert.AnError, err)
	assert.Equal(t, "object", result)
	assert.Nil(t, obj.result)
}

func TestFutureCancelBase(t *testing.T) {
	resultChan := make(chan Entry, 1)
	obj := &Future{
		fc: &FCache{},
		ent: &entry{
			reqs: map[uint64]chan<- Entry{
				42: resultChan,
			},
		},
		result: resultChan,
		cookie: 42,
	}

	obj.Cancel()

	assert.Equal(t, &Future{
		fc: &FCache{},
		ent: &entry{
			reqs: map[uint64]chan<- Entry{},
		},
		cookie:   42,
		canceled: true,
	}, obj)
}

func TestFutureCancelCanceled(t *testing.T) {
	resultChan := make(chan Entry, 1)
	obj := &Future{
		fc: &FCache{},
		ent: &entry{
			reqs: map[uint64]chan<- Entry{
				42: resultChan,
			},
		},
		result:   resultChan,
		cookie:   42,
		canceled: true,
	}

	obj.Cancel()

	assert.Equal(t, &Future{
		fc: &FCache{},
		ent: &entry{
			reqs: map[uint64]chan<- Entry{
				42: resultChan,
			},
		},
		result:   resultChan,
		cookie:   42,
		canceled: true,
	}, obj)
}

func TestFutureChannelBase(t *testing.T) {
	resultChan := make(chan Entry, 1)
	resultChan <- Entry{
		Error: assert.AnError,
	}
	obj := &Future{
		fc:     &FCache{},
		result: resultChan,
	}

	result := obj.Channel()

	data, ok := <-result
	assert.True(t, ok)
	assert.Equal(t, Entry{
		Error: assert.AnError,
	}, data)
}

func TestFutureChannelSynthesized(t *testing.T) {
	obj := &Future{
		fc: &FCache{},
		ent: &entry{
			content: &Entry{
				Error: assert.AnError,
			},
		},
	}

	result := obj.Channel()

	data, ok := <-result
	assert.True(t, ok)
	assert.Equal(t, Entry{
		Error: assert.AnError,
	}, data)
}

func TestFutureChannelCanceled(t *testing.T) {
	resultChan := make(chan Entry, 1)
	resultChan <- Entry{
		Error: assert.AnError,
	}
	obj := &Future{
		fc:       &FCache{},
		result:   resultChan,
		canceled: true,
	}

	result := obj.Channel()

	assert.Nil(t, result)
}

func TestFutureChannelNoEntry(t *testing.T) {
	obj := &Future{
		fc:  &FCache{},
		ent: &entry{},
	}

	result := obj.Channel()

	assert.Nil(t, result)
}
