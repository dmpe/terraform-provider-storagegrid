// Copyright (c) github.com/dmpe
// SPDX-License-Identifier: MIT

package provider

import (
	"errors"
	"fmt"
	"testing"
)

func TestGenericErrorErrorString(t *testing.T) {
	ge := &GenericError{Summary: "Operation failed", Details: "something went wrong"}
	got := ge.Error()
	want := "Operation failed: something went wrong"
	if got != want {
		t.Fatalf("Error() mismatch: got %q, want %q", got, want)
	}
}

func TestGenericErrorUnwrap(t *testing.T) {
	inner := fmt.Errorf("inner cause")
	ge := &GenericError{Summary: "failure", Details: "details", Err: inner}

	// Direct Unwrap method
	if !errors.Is(inner, ge.Unwrap()) {
		t.Fatalf("Unwrap() did not return the inner error")
	}

	// Standard library helpers should work as well
	if !errors.Is(inner, errors.Unwrap(ge)) {
		t.Fatalf("errors.Unwrap did not return the inner error")
	}

	if !errors.Is(ge, inner) {
		t.Fatalf("errors.Is should match the inner error returned by Unwrap")
	}
}

func TestGenericErrorIs(t *testing.T) {
	ge := &GenericError{Summary: "not_found", Details: "bucket missing"}

	// Match by Summary via errors.Is using a target value of the same type
	targetSame := &GenericError{Summary: "not_found"}
	if !errors.Is(ge, targetSame) {
		t.Fatalf("errors.Is should return true when GenericError summaries match")
	}

	// Different summary should not match
	targetDifferent := &GenericError{Summary: "permission_denied"}
	if errors.Is(ge, targetDifferent) {
		t.Fatalf("errors.Is should return false when GenericError summaries do not match")
	}

	// Non-GenericError target should not match
	if errors.Is(ge, fmt.Errorf("some other error")) {
		t.Fatalf("errors.Is should return false for non-GenericError targets")
	}
}

func TestGenericErrorIsThroughWrapping(t *testing.T) {
	base := &GenericError{Summary: "timeout", Details: "request timed out"}
	wrapped := fmt.Errorf("higher level context: %w", base)

	// `errors.Is` should walk the chain and invoke GenericError.Is
	if !errors.Is(wrapped, &GenericError{Summary: "timeout"}) {
		t.Fatalf("errors.Is should match GenericError through wrapping when summaries match")
	}

	if errors.Is(wrapped, &GenericError{Summary: "other"}) {
		t.Fatalf("errors.Is should not match GenericError with a different summary through wrapping")
	}
}
