package apperror

func NewRabbitMQFailPublishError(err error) *AppError {
	return &AppError{
		Code:       "RABBITMQ_ERROR",
		StatusCode: 500,
		Message:    "Failed to publish plan to RabbitMQ",
		Details:    err.Error(),
	}
}
