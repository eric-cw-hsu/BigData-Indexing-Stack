package apperror

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
