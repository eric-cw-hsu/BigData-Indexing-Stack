package apperror

func NewJSONMergeError(err error) *AppError {
	return &AppError{
		Code:       "JSON_MERGE_ERROR",
		StatusCode: 400,
		Message:    "Failed to merge JSON",
		Details:    err.Error(),
	}
}
