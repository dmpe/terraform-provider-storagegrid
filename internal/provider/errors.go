// Copyright (c) github.com/dmpe
// SPDX-License-Identifier: MIT

package provider

import (
	"errors"
	"fmt"
)

var ErrBucketNotFound = fmt.Errorf("bucket not found")

type GenericError struct {
	Summary string
	Details string
	Err     error
}

func (e *GenericError) Error() string {
	return fmt.Sprintf("%s: %s", e.Summary, e.Details)
}

func (e *GenericError) Unwrap() error {
	return e.Err
}

func (e *GenericError) Is(target error) bool {
	var t *GenericError
	ok := errors.As(target, &t)
	return ok && e.Summary == t.Summary
}
