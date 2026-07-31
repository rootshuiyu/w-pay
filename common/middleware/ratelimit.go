package middleware

import (
	"time"

	"wpay/common"
	"wpay/config"
	"wpay/dao"

	"github.com/gin-gonic/gin"
)

// RateLimit 接口限流中间件（按 IP）
func RateLimit() gin.HandlerFunc {
	limit := config.Global.RateLimit.RequestsPerMinute
	window := time.Minute

	return func(c *gin.Context) {
		ip := c.ClientIP()
		count, err := dao.IncrRateLimit(ip, window)
		if err != nil {
			common.Log.Warn("rate limit redis error: %v", err)
			c.Next()
			return
		}
		if count > int64(limit) {
			common.TooManyRequests(c)
			c.Abort()
			return
		}
		c.Next()
	}
}
