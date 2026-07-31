package controller

import (
	"wpay/common"
	"wpay/common/middleware"
	"wpay/service"

	"github.com/gin-gonic/gin"
)

type AuthController struct {
	authSvc *service.AuthService
}

func NewAuthController() *AuthController {
	return &AuthController{authSvc: service.NewAuthService()}
}

func (c *AuthController) Login(ctx *gin.Context) {
	var req service.LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		common.BadRequest(ctx, "参数错误")
		return
	}
	resp, err := c.authSvc.Login(req)
	if err != nil {
		if be, ok := common.IsBizError(err); ok {
			common.BadRequest(ctx, be.Msg)
			return
		}
		common.ServerError(ctx, "登录失败")
		return
	}
	common.OK(ctx, resp)
}

func (c *AuthController) Logout(ctx *gin.Context) {
	claims := middleware.GetClaims(ctx)
	if claims != nil {
		_ = c.authSvc.Logout(claims.AdminID)
	}
	common.OKMessage(ctx, "已退出登录")
}

func (c *AuthController) Profile(ctx *gin.Context) {
	claims := middleware.GetClaims(ctx)
	common.OK(ctx, gin.H{
		"admin_id": claims.AdminID,
		"username": claims.Username,
		"role":     claims.Role,
	})
}
