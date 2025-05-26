package apperror

type AppError struct {
	Code       string `json:"error_code"`
	StatusCode int    `json:"status_code"`
	Message    string `json:"message"`
	Details    any    `json:"details,omitempty"`
}

func (e *AppError) Error() string {
	return e.Message
}
