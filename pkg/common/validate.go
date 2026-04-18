package common

import (
	"socrati/pkg/errors"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

// Validate validates a struct using go-playground/validator tags.
// Returns an AppError if validation fails.
func Validate(s interface{}) error {
	if err := validate.Struct(s); err != nil {
		return errors.NewValidationError(err.Error())
	}
	return nil
}
