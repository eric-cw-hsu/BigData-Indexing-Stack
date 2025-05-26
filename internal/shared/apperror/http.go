package apperror

func NewETagRequiredError() *AppError {
	return &AppError{
		Code:       "ETAG_REQUIRED",
		StatusCode: 428,
		Message:    "ETag header required",
		Details:    "ETag header is required for this request.",
	}
}

func NewETagNotMatchError() *AppError {
	return &AppError{
		Code:       "ETAG_NOT_MATCH",
		StatusCode: 412,
		Message:    "ETag does not match",
		Details:    "ETag does not match the stored version.",
	}
}

func NewInvalidJSONError(err error) *AppError {
	return &AppError{
		Code:       "INVALID_JSON",
		StatusCode: 400,
		Message:    "Invalid JSON input",
		Details:    err.Error(),
	}
}
