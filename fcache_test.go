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
	"github.com/stretchr/testify/require"
)

func factory(ctx context.Context, key Key) *Entry {
	return nil
}

func TestNewBase(t *testing.T) {
	result, err := New(
		Index{"one", factory},
		Index{"two", factory},
	)

	assert.NoError(t, err)
	assert.Len(t, result.indexes, 2)
	require.Contains(t, result.indexes, "one")
	assert.NotNil(t, result.indexes["one"].factory)
	assert.Contains(t, result.indexes, "two")
	assert.NotNil(t, result.indexes["two"].factory)
}

func TestNewOneIndex(t *testing.T) {
	result, err := New(
		Index{"one", factory},
	)

	assert.NoError(t, err)
	assert.Len(t, result.indexes, 1)
	require.Contains(t, result.indexes, "one")
	assert.NotNil(t, result.indexes["one"].factory)
}

func TestNewMissingIndex(t *testing.T) {
	result, err := New()

	assert.Same(t, ErrMissingIndex, err)
	assert.Nil(t, result)
}

func TestNewDuplicateOption(t *testing.T) {
	result, err := New(
		Index{"one", factory},
		Index{"one", factory},
	)

	assert.Same(t, ErrDuplicateOption, err)
	assert.Nil(t, result)
}

func TestNewMissingFactory(t *testing.T) {
	result, err := New(
		Index{"one", factory},
		Index{"two", nil},
	)

	assert.Same(t, ErrMissingFactory, err)
	assert.Nil(t, result)
}
