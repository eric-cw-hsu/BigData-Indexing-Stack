package apperror

func NewRedisError(message string, err error) *AppError {
	return &AppError{
		Code:       "REDIS_ERROR",
		StatusCode: 500,
		Message:    "Redis error",
		Details:    err.Error(),
	}
}

func NewStorageError(message string, err error) *AppError {
	return &AppError{
		Code:       "STORAGE_ERROR",
		StatusCode: 500,
		Message:    message,
		Details:    err.Error(),
	}
}
