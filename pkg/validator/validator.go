package validator

import "github.com/gin-gonic/gin/binding"

type Validator struct {
	binding.StructValidator
}

func NewValidator() *Validator {
	return &Validator{
		StructValidator: binding.Validator,
	}
}
