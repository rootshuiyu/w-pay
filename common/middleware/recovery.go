package middleware

import (
	"net/http"
	"runtime/debug"

	"wpay/common"

	"github.com/gin-gonic/gin"
)

// Recovery 全局异常捕获
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				common.Log.Error("panic recovered: %v\n%s", r, debug.Stack())
				c.JSON(http.StatusInternalServerError, common.Response{
					Code:    common.CodeServerError,
					Message: "服务器内部错误",
				})
				c.Abort()
			}
		}()
		c.Next()
	}
}
