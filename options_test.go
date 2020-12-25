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
	"github.com/stretchr/testify/mock"
)

type mockLookupOption struct {
	mock.Mock
}

func (m *mockLookupOption) apply(o *lookupOptions) error {
	args := m.MethodCalled("apply", o)

	return args.Error(0)
}

func TestProcLookupOptsBase(t *testing.T) {
	opt1 := &mockLookupOption{}
	opt1.On("apply", mock.Anything).Return(nil)
	opt2 := &mockLookupOption{}
	opt2.On("apply", mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		o := args[0].(*lookupOptions)
		o.key = &Key{"one", 1}
	})

	result, err := procLookupOpts([]LookupOption{opt1, opt2})

	assert.NoError(t, err)
	assert.Nil(t, result.ent)
	assert.Equal(t, &Key{"one", 1}, result.key)
	assert.False(t, result.only)
	assert.NotNil(t, result.ctx)
	opt1.AssertExpectations(t)
	opt2.AssertExpectations(t)
}

func TestProcLookupOptsOptionError(t *testing.T) {
	opt1 := &mockLookupOption{}
	opt1.On("apply", mock.Anything).Return(assert.AnError)
	opt2 := &mockLookupOption{}

	result, err := procLookupOpts([]LookupOption{opt1, opt2})

	assert.Same(t, assert.AnError, err)
	assert.Nil(t, result.ent)
	assert.Nil(t, result.key)
	assert.False(t, result.only)
	assert.Nil(t, result.ctx)
	opt1.AssertExpectations(t)
	opt2.AssertExpectations(t)
}

func TestProcLookupOptsNoKey(t *testing.T) {
	opt1 := &mockLookupOption{}
	opt1.On("apply", mock.Anything).Return(nil)
	opt2 := &mockLookupOption{}
	opt2.On("apply", mock.Anything).Return(nil)

	result, err := procLookupOpts([]LookupOption{opt1, opt2})

	assert.Same(t, ErrNoKey, err)
	assert.Nil(t, result.ent)
	assert.Nil(t, result.key)
	assert.False(t, result.only)
	assert.Nil(t, result.ctx)
	opt1.AssertExpectations(t)
	opt2.AssertExpectations(t)
}

func TestByEntryOptionImplementsLookupOption(t *testing.T) {
	assert.Implements(t, (*LookupOption)(nil), &byEntryOption{})
}

func TestByEntryOptionApplyBase(t *testing.T) {
	o := &lookupOptions{}
	obj := byEntryOption{
		Ent: Entry{
			Error: assert.AnError,
			Keys:  []Key{{"one", 1}, {"two", 2}},
		},
	}

	err := obj.apply(o)

	assert.NoError(t, err)
	assert.Equal(t, &lookupOptions{
		ent: &Entry{
			Error: assert.AnError,
			Keys:  []Key{{"one", 1}, {"two", 2}},
		},
		key: &Key{"one", 1},
	}, o)
}

func TestByEntryOptionApplyDuplicateOption(t *testing.T) {
	o := &lookupOptions{
		key: &Key{"two", 2},
	}
	obj := byEntryOption{
		Ent: Entry{
			Error: assert.AnError,
			Keys:  []Key{{"one", 1}, {"two", 2}},
		},
	}

	err := obj.apply(o)

	assert.Same(t, ErrDuplicateOption, err)
	assert.Equal(t, &lookupOptions{
		key: &Key{"two", 2},
	}, o)
}

func TestByEntry(t *testing.T) {
	ent := Entry{
		Error: assert.AnError,
	}

	result := ByEntry(ent)

	assert.Equal(t, byEntryOption{
		Ent: Entry{
			Error: assert.AnError,
		},
	}, result)
}

func TestByKeyOptionImplementsLookupOption(t *testing.T) {
	assert.Implements(t, (*LookupOption)(nil), &byKeyOption{})
}

func TestByKeyOptionApplyBase(t *testing.T) {
	o := &lookupOptions{}
	obj := byKeyOption{
		Key: Key{"one", 1},
	}

	err := obj.apply(o)

	assert.NoError(t, err)
	assert.Equal(t, &lookupOptions{
		key: &Key{"one", 1},
	}, o)
}

func TestByKeyOptionApplyDuplicateOption(t *testing.T) {
	o := &lookupOptions{
		key: &Key{"two", 2},
	}
	obj := byKeyOption{
		Key: Key{"one", 1},
	}

	err := obj.apply(o)

	assert.Same(t, ErrDuplicateOption, err)
	assert.Equal(t, &lookupOptions{
		key: &Key{"two", 2},
	}, o)
}

func TestByKey(t *testing.T) {
	ent := Key{"one", 1}

	result := ByKey(ent)

	assert.Equal(t, byKeyOption{
		Key: Key{"one", 1},
	}, result)
}

func TestSearchCacheOptionImplementsLookupOption(t *testing.T) {
	assert.Implements(t, (*LookupOption)(nil), SearchCache)
}

func TestSearchCacheOptionApply(t *testing.T) {
	o := &lookupOptions{}

	err := SearchCache.apply(o)

	assert.NoError(t, err)
	assert.Equal(t, &lookupOptions{
		only: true,
	}, o)
}

func TestWithContextOptionImplementsLookupOption(t *testing.T) {
	assert.Implements(t, (*LookupOption)(nil), &withContextOption{})
}

func TestWithContextOptionApplyBase(t *testing.T) {
	o := &lookupOptions{}
	ctx := context.Background()
	obj := withContextOption{
		Ctx: ctx,
	}

	err := obj.apply(o)

	assert.NoError(t, err)
	assert.Same(t, ctx, o.ctx)
}

func TestWithContextOptionApplyDuplicateOption(t *testing.T) {
	ctx1 := context.Background()
	ctx2 := context.Background()
	o := &lookupOptions{
		ctx: ctx1,
	}
	obj := withContextOption{
		Ctx: ctx2,
	}

	err := obj.apply(o)

	assert.Same(t, ErrDuplicateOption, err)
	assert.Same(t, ctx1, o.ctx)
}

func TestWithContext(t *testing.T) {
	ctx := context.Background()

	result := WithContext(ctx)

	assert.Same(t, ctx, result.(withContextOption).Ctx)
}

type mockCleanOption struct {
	mock.Mock
}

func (m *mockCleanOption) apply(o *cleanOptions) {
	m.MethodCalled("apply", o)
}

func TestProcCleanOptsBase(t *testing.T) {
	opt1 := &mockCleanOption{}
	opt1.On("apply", &cleanOptions{})
	opt2 := &mockCleanOption{}
	opt2.On("apply", &cleanOptions{})

	result := procCleanOpts([]CleanOption{opt1, opt2})

	assert.Equal(t, cleanOptions{}, result)
	opt1.AssertExpectations(t)
	opt2.AssertExpectations(t)
}

func TestProcCleanOptsNoOptions(t *testing.T) {
	result := procCleanOpts([]CleanOption{})

	assert.Equal(t, cleanOptions{
		objects: true,
		errors:  true,
		pending: true,
	}, result)
}

func TestObjectsOptionImplementsCleanOption(t *testing.T) {
	assert.Implements(t, (*CleanOption)(nil), Objects)
}

func TestObjectsOptionApply(t *testing.T) {
	o := &cleanOptions{}

	Objects.apply(o)

	assert.Equal(t, &cleanOptions{
		objects: true,
	}, o)
}

func TestErrorsOptionImplementsCleanOption(t *testing.T) {
	assert.Implements(t, (*CleanOption)(nil), Errors)
}

func TestErrorsOptionApply(t *testing.T) {
	o := &cleanOptions{}

	Errors.apply(o)

	assert.Equal(t, &cleanOptions{
		errors: true,
	}, o)
}

func TestPendingOptionImplementsCleanOption(t *testing.T) {
	assert.Implements(t, (*CleanOption)(nil), Pending)
}

func TestPendingOptionApply(t *testing.T) {
	o := &cleanOptions{}

	Pending.apply(o)

	assert.Equal(t, &cleanOptions{
		pending: true,
	}, o)
}
