package errors

import "errors"

var (
	ErrInvalidJSON          = errors.New("invalid JSON")
	ErrKeyAlreadyExists     = errors.New("key already exists")
	ErrKeyNotFound          = errors.New("key not found")
	ErrMissingKeyQueryParam = errors.New("missing 'key' query parameter")
	ErrMissingIfNoneMatch   = errors.New("missing 'If-None-Match' header")
	ErrPreconditionFailed   = errors.New("precondition failed")
)
