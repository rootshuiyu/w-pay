package controller

import (
	"wpay/common"
	"wpay/service"

	"github.com/gin-gonic/gin"
)

func handleBizError(ctx *gin.Context, err error) {
	if be, ok := common.IsBizError(err); ok {
		common.BadRequest(ctx, be.Msg)
		return
	}
	common.ServerError(ctx, "操作失败")
}

func handlePayBizError(ctx *gin.Context, err error) {
	if be, ok := common.IsBizError(err); ok {
		common.BadRequest(ctx, service.SanitizePayBizMessage(be.Msg))
		return
	}
	common.ServerError(ctx, "操作失败")
}
