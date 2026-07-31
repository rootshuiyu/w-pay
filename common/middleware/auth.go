package middleware

import (
	"strings"

	"wpay/common"
	"wpay/config"
	"wpay/dao"

	"github.com/gin-gonic/gin"
)

const ContextKeyClaims = "admin_claims"

// AuthToken 后台 token 鉴权中间件
// 合规：后台管理系统仅内网/员工账号 token 登录
func AuthToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractToken(c)
		if token == "" {
			common.Unauthorized(c, "请先登录")
			c.Abort()
			return
		}

		claims, err := common.ParseToken(token, config.Global.JWT.Secret)
		if err != nil {
			common.Unauthorized(c, "token 无效或已过期")
			c.Abort()
			return
		}

		cached, err := dao.GetAdminToken(claims.AdminID)
		if err != nil || cached != token {
			common.Unauthorized(c, "登录已失效，请重新登录")
			c.Abort()
			return
		}

		c.Set(ContextKeyClaims, claims)
		c.Next()
	}
}

// RequireRole 角色权限校验
func RequireRole(roles ...common.AdminRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		val, exists := c.Get(ContextKeyClaims)
		if !exists {
			common.Forbidden(c, "无权访问")
			c.Abort()
			return
		}
		claims := val.(*common.AdminClaims)
		if !common.HasRole(claims.Role, roles...) {
			common.Forbidden(c, "权限不足")
			c.Abort()
			return
		}
		c.Next()
	}
}

func extractToken(c *gin.Context) string {
	auth := c.GetHeader("Authorization")
	if auth != "" {
		parts := strings.SplitN(auth, " ", 2)
		if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
			return parts[1]
		}
	}
	return c.GetHeader("X-Token")
}

// GetClaims 从上下文获取管理员 claims
func GetClaims(c *gin.Context) *common.AdminClaims {
	val, _ := c.Get(ContextKeyClaims)
	if val == nil {
		return nil
	}
	return val.(*common.AdminClaims)
}
