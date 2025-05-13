package utils

type AppError struct {
	Code       string `json:"error_code"`
	StatusCode int    `json:"status_code"`
	Message    string `json:"message"`
	Details    any    `json:"details,omitempty"`
}

func NewPlanExistsError() *AppError {
	return &AppError{
		Code:       "PLAN_EXISTS",
		StatusCode: 409,
		Message:    "Plan already exists",
		Details:    "Plan with the same objectId already exists.",
	}
}

func NewPlanNotFoundError(err error) *AppError {
	return &AppError{
		Code:       "PLAN_NOT_FOUND",
		StatusCode: 404,
		Message:    "Plan not found",
		Details:    err.Error(),
	}
}

func NewRabbitMQFailPublishError(err error) *AppError {
	return &AppError{
		Code:       "RABBITMQ_ERROR",
		StatusCode: 500,
		Message:    "Failed to publish plan to RabbitMQ",
		Details:    err.Error(),
	}
}

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

func NewStorageError(message string, err error) *AppError {
	return &AppError{
		Code:       "STORAGE_ERROR",
		StatusCode: 500,
		Message:    message,
		Details:    err.Error(),
	}
}

func NewJSONMergeError(err error) *AppError {
	return &AppError{
		Code:       "JSON_MERGE_ERROR",
		StatusCode: 400,
		Message:    "Failed to merge JSON",
		Details:    err.Error(),
	}
}

func NewRedisError(message string, err error) *AppError {
	return &AppError{
		Code:       "REDIS_ERROR",
		StatusCode: 500,
		Message:    "Redis error",
		Details:    err.Error(),
	}
}
