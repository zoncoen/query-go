package query

import (
	"errors"
	"fmt"
)

// ErrNotFound is the sentinel error that reports the queried element as
// absent. Extractor implementations return it (optionally wrapped) when the
// key or index does not match anything; any other error is treated as an
// extraction failure and aborts the query.
//
// Callers can test a *Query.Extract error with errors.Is(err, ErrNotFound)
// to tell a genuinely absent value apart from a failure such as a context
// cancellation that interrupted a blocking extractor.
var ErrNotFound = errors.New("not found")

// NotFoundError is the error returned by Query.Extract when the queried
// element is absent. It matches ErrNotFound with errors.Is and carries the
// position information of the failure.
type NotFoundError struct {
	// Query is the string representation of the whole query.
	Query string
	// FailedAt is the prefix of the query up to and including the extractor
	// that did not match, e.g. ".a.b" when ".a.b.c" failed at "b".
	FailedAt string
	// Err is the error reported by the extractor that did not match. It
	// preserves the diagnostic of an extractor that wrapped ErrNotFound
	// (e.g. "stream ended after 3 messages: not found") and, when the
	// extractor ran a sub-query, the inner *NotFoundError. It is exposed
	// via Unwrap, not via Error, so the message stays stable.
	Err error
}

// Error implements the error interface.
func (e *NotFoundError) Error() string {
	return fmt.Sprintf(`"%s" not found`, e.Query)
}

// Is reports whether target is ErrNotFound, so that
// errors.Is(err, ErrNotFound) matches a *NotFoundError.
func (e *NotFoundError) Is(target error) bool {
	return target == ErrNotFound
}

// Unwrap returns the extractor's original error.
func (e *NotFoundError) Unwrap() error {
	return e.Err
}
