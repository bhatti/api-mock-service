package types

// ValidationError error
type ValidationError struct {
	Message string
}

// NewValidationError constructor
func NewValidationError(msg string) *ValidationError {
	return &ValidationError{
		Message: msg,
	}
}

func (ve *ValidationError) Error() string {
	return ve.Message
}

// NotFoundError error
type NotFoundError struct {
	Message string
}

// NewNotFoundError constructor
func NewNotFoundError(msg string) *NotFoundError {
	return &NotFoundError{
		Message: msg,
	}
}

func (ne *NotFoundError) Error() string {
	return ne.Message
}
