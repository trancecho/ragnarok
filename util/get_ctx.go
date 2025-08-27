package util

import (
	"github.com/gin-gonic/gin"
)

func Uid(c *gin.Context) uint {
	val, ok := c.Get("uid")
	if !ok {
		// 这里可以按需要处理，比如返回 0 或直接中断请求
		return 0
	}

	uid, ok := val.(uint)
	if !ok {
		// 类型不对也要处理
		return 0
	}

	return uid
}
func GetUsername(c *gin.Context) string {
	username := c.MustGet("username").(string)
	return username
}

func GetRole(c *gin.Context) string {
	role := c.MustGet("role").(string)
	return role
}
