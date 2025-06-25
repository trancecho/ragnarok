package b_util

import (
	"github.com/gin-gonic/gin"
)

const (
	AdminRole        = "admin"             // 管理员角色
	DepartmentMaster = "department_master" // 部门主管角色
	ManagerRole      = "manager"           // 经理角色
)

func GetRole(c *gin.Context) string {
	s := c.GetString("role")
	if s == "" {
		return "" // 默认角色
	}
	return s
}
