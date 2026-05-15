package handler

import (
	"gin-quickstart/internal/dto"
	"gin-quickstart/internal/enum"
	"gin-quickstart/internal/service"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

type AuthHandler struct {
	s *service.AuthService
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
		c.JSON(500, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"data":    user,
	})
}

// SETTER
func (h AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	token, err := h.s.Login(c, req.Username, req.Password)

	if err != nil {
		c.JSON(401, gin.H{
			"success": false,
			"error":   "Invalid username or password : " + err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"message": "Login successful",
		"token":   token,
	})
}

func (h AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("alphanumdash", dto.Alphanumdash)
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	err := h.s.Register(c, req.Name, req.Username, req.Email, req.Password, enum.RoleUser.String())

	if err != nil {
		c.JSON(500, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(201, gin.H{
		"success": true,
		"message": "User registered successfully",
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	userID, exists := c.Get("user_id")
	token, tokenExists := c.Get("token")

	if !exists || !tokenExists {
		c.JSON(401, gin.H{
			"success": false,
			"error":   "Unauthorized",
		})
		return
	}

	if !exists {
		c.JSON(401, gin.H{
			"success": false,
			"error":   "Unauthorized",
		})
		return
	}

	err := h.s.Logout(c, uint64(userID.(uint)), token.(string))

	if err != nil {
		c.JSON(500, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"message": "User logged out successfully",
	})
}

func (h *AuthHandler) UpdateProfile(c *gin.Context) {

	userID, exists := c.Get("user_id")

	if !exists {
		c.JSON(401, gin.H{
			"success": false,
			"error":   "Unauthorized",
		})
	}

	user, err := h.s.GetUserByID(c, uint64(userID.(uint)))

	if err != nil {
		c.JSON(500, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	file, _ := c.FormFile("avatar")

	if user == nil {
		c.JSON(404, gin.H{
			"success": false,
			"error":   "User not found",
		})
		return
	}

	var req struct {
		Name  string `json:"name" form:"name" binding:"omitempty,min=3,max=50"`
		Email string `json:"email" form:"email" binding:"omitempty,email"`
		Bio   string `json:"bio" form:"bio" binding:"omitempty,max=500"`
	}

	if err := c.ShouldBind(&req); err != nil {
		c.JSON(400, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	uErr := h.s.UpdateProfile(c, uint64(user.ID), req.Name, req.Email, req.Bio, file)

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
	var req dto.ForgotPasswordRequest
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
	var req dto.ResetPasswordRequest
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
