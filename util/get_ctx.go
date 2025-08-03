package util

import (
	"github.com/gin-gonic/gin"
)

func Uid(c *gin.Context) uint {
	uid := c.MustGet("uid").(uint)
	uidInt := uid
	return uidInt
}

func GetUsername(c *gin.Context) string {
	username := c.MustGet("username").(string)
	return username
}

func GetRole(c *gin.Context) string {
	role := c.MustGet("role").(string)
	return role
}
