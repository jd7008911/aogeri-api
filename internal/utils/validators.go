// internal/utils/validators.go
package utils

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/go-playground/validator/v10"
)

var (
	ethAddressRe    = regexp.MustCompile(`^0x[0-9a-fA-F]{40}$`)
	tokenSymbolRe   = regexp.MustCompile(`^[A-Z0-9_\-]{1,16}$`)
	decimalNumberRe = regexp.MustCompile(`^\d+(\.\d+)?$`)
)

// NewValidator returns a *validator.Validate with project-specific custom validations registered.
func NewValidator() *validator.Validate {
	v := validator.New()

	_ = v.RegisterValidation("wallet", func(fl validator.FieldLevel) bool {
		s := fl.Field().String()
		if s == "" {
			return true
		}
		return ethAddressRe.MatchString(s)
	})

	_ = v.RegisterValidation("symbol", func(fl validator.FieldLevel) bool {
		s := fl.Field().String()
		return tokenSymbolRe.MatchString(s)
	})

	_ = v.RegisterValidation("decimal", func(fl validator.FieldLevel) bool {
		s := fl.Field().String()
		if s == "" {
			return false
		}
		if !decimalNumberRe.MatchString(s) {
			return false
		}
		// ensure it's a non-negative number
		if f, err := strconv.ParseFloat(s, 64); err == nil {
			return f >= 0
		}
		return false
	})

	return v
}

// FormatValidationError converts validator errors into a readable string.
func FormatValidationError(err error) string {
	if err == nil {
		return ""
	}
	if ve, ok := err.(validator.ValidationErrors); ok {
		out := ""
		for _, e := range ve {
			out += fmt.Sprintf("%s: %s; ", e.Field(), e.Tag())
		}
		return out
	}
	return err.Error()
}
