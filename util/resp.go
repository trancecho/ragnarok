package util

import (
	"log"

	"github.com/gin-gonic/gin"
)

func Ok(c *gin.Context, message string, data any) {
	if data == nil {
		data = gin.H{}
	}
	c.JSON(200, gin.H{
		"err_code": Success,
		"message":  message,
		"data":     data,
	})
}

// 规范：错误码提供三位，从000开始
func ClientError(c *gin.Context, errCode any, message string) {
	log.Println("客户端错误", errCode, message)
	c.JSON(400, gin.H{
		"err_code": errCode,
		"message":  message,
	})
}

func ServerError(c *gin.Context, errCode any, message string) {
	log.Println("服务端错误", errCode, message)
	c.JSON(500, gin.H{
		"err_code": errCode,
		"message":  message,
	})
}

const (
	Success = "Success"
)

const (
	AuthError         = "Error.Auth"
	DefaultError      = "Error.Default"
	QueryParamError   = "Error.QueryParam"
	UnAuthorizedError = "Error.UnAuthorized"
	TokenExpiredError = "Error.TokenExpired"
	// 权限不足
	PermissionDeniedError = "Error.PermissionDenied"
)
