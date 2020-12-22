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

import "errors"

// Errors that may be returned by the cache.
var (
	ErrNoKey           = errors.New("no key specified")
	ErrDuplicateOption = errors.New("duplicate option")
	ErrMissingIndex    = errors.New("at least one index must be provided")
	ErrMissingFactory  = errors.New("index factory is required")
	ErrBadIndex        = errors.New("unknown cache index")
	ErrNotCached       = errors.New("key does not exist in cache")
	ErrIncongruentKeys = errors.New("old keys are not congruent with new keys")
	ErrEntryNotFound   = errors.New("entry not found with specified key")
)

// PermanentError is an implementation of the error interface that
// wraps another error to signal that it is a permanent error.
// Permanent errors will be cached, as opposed to other errors.
type PermanentError struct {
	Err error // The wrapped error
}

// Error returns the error message.
func (p *PermanentError) Error() string {
	return p.Err.Error()
}

// Unwrap returns the wrapped error.
func (p *PermanentError) Unwrap() error {
	return p.Err
}

// IsPermanent is a test to see if an error is a permanent error.  It
// returns true if the error is wrapped by PermanentError, and false
// otherwise.
func IsPermanent(err error) bool {
	var tmp *PermanentError

	return errors.As(err, &tmp)
}
