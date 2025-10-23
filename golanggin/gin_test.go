package golanggin

import (
	"testing"

	"github.com/gin-gonic/gin"
)

func Test_PathRegister(t *testing.T) {
	engine := gin.New()
	// Gin 的路由匹配遵循 “固定路径优先级 > 参数路径” 的规则
	engine.POST("/users/me", func(c *gin.Context) {})
	engine.POST("/users/:id", func(c *gin.Context) {})
	engine.POST("/users/me/abc", func(c *gin.Context) {})

	// will panic when register second path
	engine.POST("/student/xx", func(c *gin.Context) {})
	// engine.POST("/student/*all", func(c *gin.Context) {})

	// panic: ':name' in new path '/bcd/:name/test' conflicts with existing wildcard ':id' in existing prefix '/bcd/:id'
	// engine.POST("/bcd/:id/me", func(c *gin.Context) {})
	// engine.POST("/bcd/:name/test", func(c *gin.Context) {})

	engine.POST("/students/x:x", func(c *gin.Context) {})
	// panic: not valid param name
	// engine.POST("/students/:x:x", func(c *gin.Context) {})
	// panic: not valid
	// engine.POST("/flower/*/abc", func(c *gin.Context) {})
	// engine.POST("/flower/:/abc", func(c *gin.Context) {})
}
