package types

type ValidationError struct {
	Message string
}

func NewValidationError(msg string) *ValidationError {
	return &ValidationError{
		Message: msg,
	}
}

func (ve *ValidationError) Error() string {
	return ve.Message
}

type NotFoundError struct {
	Message string
}

func NewNotFoundError(msg string) *NotFoundError {
	return &NotFoundError{
		Message: msg,
	}
}

func (ne *NotFoundError) Error() string {
	return ne.Message
}
