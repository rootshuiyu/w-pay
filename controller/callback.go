package controller

import (
	"net/http"
	"strconv"

	"wpay/common"
	"wpay/service"

	"github.com/gin-gonic/gin"
	"github.com/go-pay/gopay"
	"github.com/go-pay/gopay/wechat/v3"
)

type CallbackController struct {
	callbackSvc *service.CallbackService
}

func NewCallbackController() *CallbackController {
	return &CallbackController{callbackSvc: service.NewCallbackService()}
}

// Wechat POST /api/notify/wx
// notify_url 需配置为：https://域名/api/notify/wx?store_id={门店ID}
func (c *CallbackController) Wechat(ctx *gin.Context) {
	storeID, _ := strconv.ParseUint(ctx.Query("store_id"), 10, 64)
	channelID, _ := strconv.ParseUint(ctx.Query("channel_id"), 10, 64)
	if storeID == 0 {
		ctx.JSON(http.StatusBadRequest, &wechat.V3NotifyRsp{Code: gopay.FAIL, Message: "缺少 store_id"})
		return
	}
	if err := c.callbackSvc.HandleWechat(ctx.Request, storeID, channelID); err != nil {
		common.Log.Warn("wechat callback store=%d: %v", storeID, err)
		ctx.JSON(http.StatusBadRequest, &wechat.V3NotifyRsp{Code: gopay.FAIL, Message: "处理失败"})
		return
	}
	ctx.JSON(http.StatusOK, &wechat.V3NotifyRsp{Code: gopay.SUCCESS, Message: "成功"})
}

// Alipay POST /api/notify/alipay
func (c *CallbackController) Alipay(ctx *gin.Context) {
	if err := c.callbackSvc.HandleAlipay(ctx.Request); err != nil {
		common.Log.Warn("alipay callback: %v", err)
		ctx.String(http.StatusBadRequest, "fail")
		return
	}
	ctx.String(http.StatusOK, "success")
}
