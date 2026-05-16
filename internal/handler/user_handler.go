package handler

import (
	"gin-quickstart/internal/service"
	"gin-quickstart/pkg/logger"
	"gin-quickstart/pkg/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	log     logger.Logger
	Service *service.UserService
}

func NewUserHandler(log *logger.Logger, service *service.UserService) *UserHandler {
	return &UserHandler{
		log:     *log,
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
	h.log.Debug(c, "GetAllUsers called")
	users, err := h.Service.GetAllUsers(c)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
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
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "user ID is required",
			})
			return
		}

		id = uint64(paramUid.(uint))
	}

	if id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid user ID",
		})
		return
	}

	user, err := h.Service.GetUserByID(c, id)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "ok",
		"user":    user,
	})
}

func (h UserHandler) GetUserByUsername(c *gin.Context) {
	username := c.Param("username")

	user, err := h.Service.GetUserByUsername(c, username)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "ok",
		"user":    user,
	})
}

func (h UserHandler) GetUserByEmail(c *gin.Context) {
	email := c.Param("email")

	user, err := h.Service.GetUserByEmail(c, email)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "ok",
		"user":    user,
	})
}

func (h UserHandler) GetFollowers(c *gin.Context) {
	var userID uint64
	userIDParam := c.Param("id")

	userID, err := strconv.ParseUint(userIDParam, 10, 64)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid user ID",
		})
		return
	}

	followers, err := h.Service.GetFollowers(c, userID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    followers,
	})
}

func (h UserHandler) GetFollowing(c *gin.Context) {
	var userID uint64
	userIDParam := c.Param("id")

	userID, err := strconv.ParseUint(userIDParam, 10, 64)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid user ID",
		})
		return
	}

	following, err := h.Service.GetFollowing(c, userID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    following,
	})
}

// SETTER
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req CreateUserRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"data":    utils.BuildValidationErrors(err, &req),
			"error":   err.Error(),
		})
		return
	}

	user, err := h.Service.CreateUser(
		c,
		req.Name,
		req.Username,
		req.Email,
		req.Password,
		req.Avatar,
		req.Bio,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "ok",
		"user":    user,
	})

}

func (h *UserHandler) UpdateUser(c *gin.Context) {
	var userID uint64
	userIDParam := c.Param("id")

	userID, err := strconv.ParseUint(userIDParam, 10, 64)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "user ID is required",
		})
		return
	}

	var req UpdateUserRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"data":    utils.BuildValidationErrors(err, &req),
			"error":   err.Error(),
		})
		return
	}

	updatedUser, err := h.Service.UpdateUser(
		c,
		userID,
		req.Name,
		req.Username,
		req.Email,
		req.Password,
		req.Avatar,
		req.Bio,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "ok",
		"user":    updatedUser,
	})
}

func (h *UserHandler) DeleteUser(c *gin.Context) {
	param := c.Param("id")

	id, err := strconv.ParseUint(param, 10, 64)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid user ID",
		})
		return
	}

	err = h.Service.DeleteUser(c, id)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "ok",
	})
}

func (h *UserHandler) Follow(c *gin.Context) {
	param := c.Param("id")
	var userID uint64

	userIDParam, pErr := c.Get("user_id")

	if !pErr {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "user ID is required",
		})
		return
	}

	userID = uint64(userIDParam.(uint))
	id, err := strconv.ParseUint(param, 10, 64)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid user ID",
		})
		return
	}

	err = h.Service.FollowUser(c, userID, id)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "ok",
	})
}

func (h *UserHandler) Unfollow(c *gin.Context) {
	param := c.Param("id")
	var userID uint64

	userIDParam, pErr := c.Get("user_id")

	if !pErr {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "user ID is required",
		})
		return
	}

	userID = uint64(userIDParam.(uint))
	id, err := strconv.ParseUint(param, 10, 64)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid user ID",
		})
		return
	}

	err = h.Service.UnfollowUser(c, userID, id)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "ok",
	})
}
