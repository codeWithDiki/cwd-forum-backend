package handler

import (
	"gin-quickstart/internal/dto"
	"gin-quickstart/internal/enum"
	"gin-quickstart/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
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

	token, err := h.s.Login(req.Username, req.Password, c)

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

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	err := h.s.Register(req.Name, req.Username, req.Email, req.Password, enum.RoleUser.String(), c)

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

	err := h.s.Logout(uint64(userID.(uint)), token.(string))

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

	user, err := h.s.GetUserByID(uint64(userID.(uint)), c)

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
