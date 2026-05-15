package utils

import (
	"regexp"
	"strings"
	"unicode"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

var (
	slugRegex     = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
	hexColorRegex = regexp.MustCompile(`^#(?:[0-9a-fA-F]{3}|[0-9a-fA-F]{6})$`)
	phoneRegex    = regexp.MustCompile(`^\+?[0-9]{8,15}$`)
)

func RegisterCustomValidators() {
	v, ok := binding.Validator.Engine().(*validator.Validate)
	if !ok {
		return
	}

	_ = v.RegisterValidation("no_spaces", noSpaces)
	_ = v.RegisterValidation("strong_password", strongPassword)
	_ = v.RegisterValidation("slug", isSlug)
	_ = v.RegisterValidation("hex_color", isHexColor)
	_ = v.RegisterValidation("phone", isPhone)
	_ = v.RegisterValidation("not_blank", notBlank)
	_ = v.RegisterValidation("username", isUsername)
}

func noSpaces(fl validator.FieldLevel) bool {
	return !strings.ContainsAny(fl.Field().String(), " \t\n\r")
}

func strongPassword(fl validator.FieldLevel) bool {
	s := fl.Field().String()
	if len(s) < 8 {
		return false
	}
	var hasUpper, hasLower, hasDigit bool
	for _, r := range s {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasDigit = true
		}
	}
	return hasUpper && hasLower && hasDigit
}

func isSlug(fl validator.FieldLevel) bool {
	return slugRegex.MatchString(fl.Field().String())
}

func isHexColor(fl validator.FieldLevel) bool {
	return hexColorRegex.MatchString(fl.Field().String())
}

func isPhone(fl validator.FieldLevel) bool {
	return phoneRegex.MatchString(fl.Field().String())
}

func notBlank(fl validator.FieldLevel) bool {
	return strings.TrimSpace(fl.Field().String()) != ""
}

func isUsername(fl validator.FieldLevel) bool {
	s := fl.Field().String()
	if len(s) < 3 || len(s) > 30 {
		return false
	}
	for _, r := range s {
		if !(unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' || r == '.') {
			return false
		}
	}
	return true
}
