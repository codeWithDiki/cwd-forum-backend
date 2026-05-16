package handler

import (
	"gin-quickstart/internal/enum"
	"gin-quickstart/internal/service"
	"gin-quickstart/pkg/utils"
	"mime/multipart"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	s *service.AuthService
}

type RegisterRequest struct {
	Name                 string `json:"name" form:"name" binding:"required"`
	Username             string `json:"username" form:"username" binding:"required,username"`
	Email                string `json:"email" form:"email" binding:"required,email"`
	Password             string `json:"password" form:"password" binding:"required,min=6"`
	PasswordConfirmation string `json:"password_confirmation" form:"password_confirmation" binding:"required,min=6,eqcsfield=Password"`
}

type LoginRequest struct {
	Username string `json:"username" form:"username" binding:"required"`
	Password string `json:"password" form:"password" binding:"required"`
}

type UpdateProfileRequest struct {
	Name   string                `json:"name,omitempty" form:"name,omitempty" binding:"omitempty,min=3,max=50"`
	Email  string                `json:"email,omitempty" form:"email,omitempty" binding:"omitempty,email"`
	Bio    string                `json:"bio,omitempty" form:"bio,omitempty" binding:"omitempty,max=500"`
	Avatar *multipart.FileHeader `json:"avatar,omitempty" form:"avatar,omitempty" binding:"omitempty"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" form:"email" binding:"required,email"`
}

type ResetPasswordRequest struct {
	Token    string `json:"token" form:"token" binding:"required"`
	Password string `json:"password" form:"password" binding:"required,min=6"`
}

func NewAuthHandler(s *service.AuthService) *AuthHandler {
	return &AuthHandler{
		s: s,
	}
}

// GETTER
func (h AuthHandler) GetProfile(c *gin.Context) {
	user, err := h.s.GetLoggedUser(c)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    user,
	})
}

// SETTER
func (h AuthHandler) Login(c *gin.Context) {
	var req LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	token, err := h.s.Login(c, req.Username, req.Password)

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Invalid username or password : " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Login successful",
		"token":   token,
	})
}

func (h AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	err := h.s.Register(c, req.Name, req.Username, req.Email, req.Password, enum.RoleUser.String())

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "User registered successfully",
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	userID, exists := c.Get("user_id")
	token, tokenExists := c.Get("token")

	if !exists || !tokenExists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Unauthorized",
		})
		return
	}

	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Unauthorized",
		})
		return
	}

	err := h.s.Logout(c, uint64(userID.(uint)), token.(string))

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "User logged out successfully",
	})
}

func (h *AuthHandler) UpdateProfile(c *gin.Context) {

	userID, exists := c.Get("user_id")

	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Unauthorized",
		})
	}

	user, err := h.s.GetUserByID(c, uint64(userID.(uint)))

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	var req UpdateProfileRequest

	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"data":    utils.BuildValidationErrors(err, &req),
			"error":   "Validation Errors",
		})
		return
	}

	uErr := h.s.UpdateProfile(c, uint64(user.ID), req.Name, req.Email, req.Bio, req.Avatar)

	if uErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   uErr.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Profile updated successfully",
	})
}

func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	resetBaseURL := os.Getenv("APP_URL") + "/reset-password"
	if err := h.s.ForgotPassword(req.Email, resetBaseURL, c); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "If the email exists, a reset link has been sent"})
}

func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	if err := h.s.ResetPassword(req.Token, req.Password, c); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Password reset successfully"})
}
