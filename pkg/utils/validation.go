package utils

import (
	"errors"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Value   any    `json:"value"`
}

type ValidationErrors []ValidationError

func BuildValidationErrors(err error, req any) ValidationErrors {
	var ve validator.ValidationErrors
	if !errors.As(err, &ve) {
		return nil
	}

	t := reflect.TypeOf(req)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	out := make(ValidationErrors, 0, len(ve))
	for _, fe := range ve {
		field := fe.Field()
		if sf, ok := t.FieldByName(fe.StructField()); ok {
			if tag := sf.Tag.Get("form"); tag != "" {
				field = strings.Split(tag, ",")[0]
			} else if tag := sf.Tag.Get("json"); tag != "" {
				field = strings.Split(tag, ",")[0]
			}
		}

		msg := buildMessage(field, fe)

		out = append(out, ValidationError{
			Field:   field,
			Message: msg,
			Value:   fe.Value(),
		})
	}
	return out
}

func buildMessage(field string, fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return field + " is required"
	case "email":
		return field + " must be a valid email"
	case "url":
		return field + " must be a valid URL"
	case "min":
		return field + " must be at least " + fe.Param() + " characters"
	case "max":
		return field + " must be at most " + fe.Param() + " characters"
	case "len":
		return field + " must be exactly " + fe.Param() + " characters"
	case "eq":
		return field + " must be equal to " + fe.Param()
	case "ne":
		return field + " must not be equal to " + fe.Param()
	case "gt":
		return field + " must be greater than " + fe.Param()
	case "gte":
		return field + " must be greater than or equal to " + fe.Param()
	case "lt":
		return field + " must be less than " + fe.Param()
	case "lte":
		return field + " must be less than or equal to " + fe.Param()
	case "oneof":
		return field + " must be one of [" + fe.Param() + "]"
	case "alpha":
		return field + " must contain only letters"
	case "alphanum":
		return field + " must contain only letters and digits"
	case "numeric":
		return field + " must be numeric"
	case "uuid":
		return field + " must be a valid UUID"
	case "no_spaces":
		return field + " cannot contain whitespace"
	case "strong_password":
		return field + " must be at least 8 characters and contain uppercase, lowercase, and a digit"
	case "slug":
		return field + " must be a valid slug (lowercase, digits, hyphens)"
	case "hex_color":
		return field + " must be a valid hex color (e.g. #fff or #ffffff)"
	case "phone":
		return field + " must be a valid phone number"
	case "not_blank":
		return field + " must not be blank"
	case "username":
		return field + " must be 3-30 chars of letters, digits, underscore, or dot"
	default:
		return field + " failed on the '" + fe.Tag() + "' rule"
	}
}
