package b_util

import (
	"log"

	"github.com/gin-gonic/gin"
)

func Ok(c *gin.Context, message string, data gin.H) {
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
	// 权限错误
	AuthError = "Auth.Error"
	// 参数错误
	ParamError = "Param.Error"
	// 数据库错误
	DbError = "Db.Error"
	// 业务错误
	BusinessError = "Business.Error"
	// 文件错误
	FileError = "File.Error"
	// 其他错误
	OtherError = "Other.Error"
)
