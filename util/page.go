package util

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetPageParams(c *gin.Context) (int, int) {
	page := c.Param("page")
	pageSize := c.Param("page_size")
	p, err := strconv.Atoi(page)
	if err != nil || p <= 0 {
		p = 1
	}
	ps, err := strconv.Atoi(pageSize)
	if err != nil || ps <= 0 {
		ps = 10
	}
	return p, ps
}
