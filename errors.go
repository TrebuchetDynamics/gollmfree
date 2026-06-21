package gollmfree

import (
	"errors"
	"fmt"
	"strings"
)

// AttemptError records one failed provider attempt.
type AttemptError struct {
	Provider string
	Attempt  int
	Err      error
}

// Error formats the attempt failure with provider context.
func (e AttemptError) Error() string {
	if e.Err == nil {
		return fmt.Sprintf("%s attempt %d failed", e.Provider, e.Attempt)
	}
	return fmt.Sprintf("%s attempt %d failed: %v", e.Provider, e.Attempt, e.Err)
}

// Unwrap returns the underlying provider error.
func (e AttemptError) Unwrap() error { return e.Err }

// CombinedError contains every provider failure from a fallback attempt loop.
type CombinedError struct {
	Attempts []AttemptError
}

// Error formats every failed provider attempt in order.
func (e CombinedError) Error() string {
	if len(e.Attempts) == 0 {
		return "gollmfree: no provider attempts"
	}
	parts := make([]string, len(e.Attempts))
	for i, attempt := range e.Attempts {
		parts[i] = attempt.Error()
	}
	return "gollmfree: all provider attempts failed: " + strings.Join(parts, "; ")
}

// Unwrap returns the underlying attempt errors for errors.Is/errors.As traversal.
func (e CombinedError) Unwrap() []error {
	out := make([]error, len(e.Attempts))
	for i, attempt := range e.Attempts {
		out[i] = attempt
	}
	return out
}

func combinedErrorHasContext(err error, target error) bool {
	return target != nil && errors.Is(err, target)
}
