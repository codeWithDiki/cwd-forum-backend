package dto

import (
	"github.com/go-playground/validator/v10"
)

type RegisterRequest struct {
	Name                 string `json:"name" form:"name" binding:"required"`
	Username             string `json:"username" form:"username" binding:"required,Alphanumdash"`
	Email                string `json:"email" form:"email" binding:"required,email"`
	Password             string `json:"password" form:"password" binding:"required,min=6"`
	PasswordConfirmation string `json:"password_confirmation" form:"password_confirmation" binding:"required,min=6,eqcsfield=Password"`
}

type LoginRequest struct {
	Username string `json:"username" form:"username" binding:"required"`
	Password string `json:"password" form:"password" binding:"required"`
}

var Alphanumdash validator.Func = func(fl validator.FieldLevel) bool {
	value := fl.Field().String()

	for _, char := range value {
		if !(char >= 'a' && char <= 'z') &&
			!(char >= 'A' && char <= 'Z') &&
			!(char >= '0' && char <= '9') &&
			!(char == '_') {
			return false
		}
	}

	return true
}
