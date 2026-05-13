package handler

import (
	"gin-quickstart/internal/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	Service *service.UserService
}

func NewUserHandler(service *service.UserService) *UserHandler {
	return &UserHandler{
		Service: service,
	}
}

type CreateUserRequest struct {
	Name     string `json:"name" binding:"required"`
	Username string `json:"username" binding:"required,alphanum"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Avatar   string `json:"avatar" binding:"omitempty,url"`
	Bio      string `json:"bio" binding:"omitempty,max=500"`
}

type UpdateUserRequest struct {
	Name     *string `json:"name,omitempty"`
	Username *string `json:"username,omitempty" binding:"omitempty,alphanum"`
	Email    *string `json:"email,omitempty" binding:"omitempty,email"`
	Password *string `json:"password,omitempty" binding:"omitempty,min=8"`
	Avatar   *string `json:"avatar,omitempty" binding:"omitempty,url"`
	Bio      *string `json:"bio,omitempty" binding:"omitempty,max=500"`
}

// GETTER
func (h UserHandler) GetAllUsers(c *gin.Context) {
	users, err := h.Service.GetAllUsers(c)

	if err != nil {
		c.JSON(500, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"data":    users,
	})
}

func (h UserHandler) GetUserByID(c *gin.Context) {
	var param string

	param = c.Param("id")
	id, err := strconv.ParseUint(param, 10, 64)

	if id == 0 {
		paramUid, err := c.Get("user_id")

		if !err {
			c.JSON(400, gin.H{
				"success": false,
				"error":   "user ID is required",
			})
			return
		}

		id = uint64(paramUid.(uint))
	}

	if id == 0 {
		c.JSON(400, gin.H{
			"success": false,
			"error":   "invalid user ID",
		})
		return
	}

	user, err := h.Service.GetUserByID(id, c)

	if err != nil {
		c.JSON(500, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"message": "ok",
		"user":    user,
	})
}

func (h UserHandler) GetUserByUsername(c *gin.Context) {
	username := c.Param("username")

	user, err := h.Service.GetUserByUsername(username, c)

	if err != nil {
		c.JSON(500, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"message": "ok",
		"user":    user,
	})
}

func (h UserHandler) GetUserByEmail(c *gin.Context) {
	email := c.Param("email")

	user, err := h.Service.GetUserByEmail(email, c)

	if err != nil {
		c.JSON(500, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"message": "ok",
		"user":    user,
	})
}

func (h UserHandler) GetFollowers(c *gin.Context) {

	userID := c.GetUint64("user_id")

	if userID != 0 {
		c.JSON(400, gin.H{
			"success": false,
			"error":   "invalid user ID",
		})
		return
	}

	followers, err := h.Service.GetFollowers(userID, c)

	if err != nil {
		c.JSON(500, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"data":    followers,
	})
}

func (h UserHandler) GetFollowing(c *gin.Context) {
	userID := c.GetUint64("user_id")

	if userID != 0 {
		c.JSON(400, gin.H{
			"success": false,
			"error":   "invalid user ID",
		})
		return
	}

	following, err := h.Service.GetFollowing(userID, c)

	if err != nil {
		c.JSON(500, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"data":    following,
	})
}

// SETTER
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req CreateUserRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	user, err := h.Service.CreateUser(
		req.Name,
		req.Username,
		req.Email,
		req.Password,
		req.Avatar,
		req.Bio,
		c,
	)

	if err != nil {
		c.JSON(500, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"message": "ok",
		"user":    user,
	})

}

func (h *UserHandler) UpdateUser(c *gin.Context) {
	var id uint64
	uidParam, pErr := c.Get("user_id")

	if !pErr {
		c.JSON(400, gin.H{
			"success": false,
			"error":   "user ID is required",
		})
	}

	id = uint64(uidParam.(uint))

	var req UpdateUserRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	updatedUser, err := h.Service.UpdateUser(
		id,
		req.Name,
		req.Username,
		req.Email,
		req.Password,
		req.Avatar,
		req.Bio,
		c,
	)

	if err != nil {
		c.JSON(500, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"message": "ok",
		"user":    updatedUser,
	})
}

func (h *UserHandler) DeleteUser(c *gin.Context) {
	param := c.Param("id")

	id, err := strconv.ParseUint(param, 10, 64)

	if err != nil {
		c.JSON(400, gin.H{
			"success": false,
			"error":   "invalid user ID",
		})
		return
	}

	err = h.Service.DeleteUser(id, c)

	if err != nil {
		c.JSON(500, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"message": "ok",
	})
}

func (h *UserHandler) Follow(c *gin.Context) {
	param := c.Param("id")
	var userID uint64

	userIDParam, pErr := c.Get("user_id")

	if !pErr {
		c.JSON(400, gin.H{
			"success": false,
			"error":   "user ID is required",
		})
		return
	}

	userID = uint64(userIDParam.(uint))
	id, err := strconv.ParseUint(param, 10, 64)

	if err != nil {
		c.JSON(400, gin.H{
			"success": false,
			"error":   "invalid user ID",
		})
		return
	}

	err = h.Service.FollowUser(userID, id, c)

	if err != nil {
		c.JSON(500, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"message": "ok",
	})
}

func (h *UserHandler) Unfollow(c *gin.Context) {
	param := c.Param("id")
	var userID uint64

	userIDParam, pErr := c.Get("user_id")

	if !pErr {
		c.JSON(400, gin.H{
			"success": false,
			"error":   "user ID is required",
		})
		return
	}

	userID = uint64(userIDParam.(uint))
	id, err := strconv.ParseUint(param, 10, 64)

	if err != nil {
		c.JSON(400, gin.H{
			"success": false,
			"error":   "invalid user ID",
		})
		return
	}

	err = h.Service.UnfollowUser(userID, id, c)

	if err != nil {
		c.JSON(500, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"message": "ok",
	})
}
