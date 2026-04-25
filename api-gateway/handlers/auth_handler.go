package handlers

import (
	"github.com/dapr-oms/api-gateway/utils"
	"github.com/gin-gonic/gin"
)

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// Login 用户登录
func Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, utils.CodeBadRequest, "invalid request parameters")
		return
	}

	// 简单验证（生产环境应从数据库验证）
	// 默认账号: admin/admin123
	if req.Username != "admin" || req.Password != "admin123" {
		utils.Error(c, utils.CodeUnauthorized, "invalid username or password")
		return
	}

	token, err := utils.GenerateToken(1, req.Username)
	if err != nil {
		utils.Error(c, utils.CodeInternalError, "failed to generate token")
		return
	}

	utils.Success(c, gin.H{
		"token":       token,
		"expires_in":  86400,
		"token_type":  "Bearer",
	})
}

// Logout 用户登出
func Logout(c *gin.Context) {
	utils.Success(c, nil)
}
