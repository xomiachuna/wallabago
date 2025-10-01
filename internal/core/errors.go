package core

import "github.com/pkg/errors"

// RetrievalError represents an error caused by invalid user input or external factors
// beyond the system's control (e.g., invalid URLs, unreachable domains, 404 responses).
// This should result in a 400 Bad Request HTTP status code rather than a 5xx.
type RetrievalError struct {
	error
}

// NewRetrievalError creates a new RetrievalError wrapping the given error with a message
func NewRetrievalError(message string, cause error) *RetrievalError {
	return &RetrievalError{
		error: errors.Wrap(cause, message),
	}
}
