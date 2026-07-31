package controller

import (
	"strconv"

	"wpay/common"
	"wpay/model"
	"wpay/service"

	"github.com/gin-gonic/gin"
)

type PayPlatformController struct {
	platformSvc *service.PayPlatformService
}

func NewPayPlatformController() *PayPlatformController {
	return &PayPlatformController{platformSvc: service.NewPayPlatformService()}
}

func (c *PayPlatformController) Add(ctx *gin.Context) {
	var req service.PlatformCreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		common.BadRequest(ctx, "参数错误")
		return
	}
	p, err := c.platformSvc.Create(req)
	if err != nil {
		handleBizError(ctx, err)
		return
	}
	common.OK(ctx, p)
}

func (c *PayPlatformController) Edit(ctx *gin.Context) {
	var req service.PlatformEditRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		common.BadRequest(ctx, "参数错误")
		return
	}
	if req.ID == 0 {
		common.BadRequest(ctx, "缺少平台 id")
		return
	}
	p, err := c.platformSvc.Update(req.ID.Uint64(), req)
	if err != nil {
		handleBizError(ctx, err)
		return
	}
	common.OK(ctx, p)
}

func (c *PayPlatformController) UpdateStatus(ctx *gin.Context) {
	var req service.PlatformStatusRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		common.BadRequest(ctx, "参数错误")
		return
	}
	if req.ID == 0 {
		common.BadRequest(ctx, "缺少平台 id")
		return
	}
	p, err := c.platformSvc.UpdateStatus(req.ID.Uint64(), req.Status)
	if err != nil {
		handleBizError(ctx, err)
		return
	}
	common.OK(ctx, p)
}

func (c *PayPlatformController) Del(ctx *gin.Context) {
	var req service.PlatformDeleteRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		if idStr := ctx.Query("id"); idStr != "" {
			id, err := strconv.ParseUint(idStr, 10, 64)
			if err != nil {
				common.BadRequest(ctx, "无效ID")
				return
			}
			req.ID = common.FlexUint64(id)
		} else {
			common.BadRequest(ctx, "参数错误")
			return
		}
	}
	if req.ID == 0 {
		common.BadRequest(ctx, "缺少平台 id")
		return
	}
	if err := c.platformSvc.Delete(req.ID.Uint64()); err != nil {
		handleBizError(ctx, err)
		return
	}
	common.OKMessage(ctx, "删除成功")
}

func (c *PayPlatformController) List(ctx *gin.Context) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("page_size", "20"))
	keyword := ctx.Query("keyword")
	var status *int8
	if s := ctx.Query("status"); s != "" {
		v, err := strconv.ParseInt(s, 10, 8)
		if err == nil {
			st := int8(v)
			status = &st
		}
	}
	list, total, err := c.platformSvc.List(keyword, status, page, pageSize)
	if err != nil {
		common.ServerError(ctx, "查询失败")
		return
	}
	common.OK(ctx, gin.H{"list": list, "total": total, "page": page, "page_size": pageSize})
}

func (c *PayPlatformController) Channels(ctx *gin.Context) {
	platformIDStr := ctx.Query("platform_id")
	if platformIDStr == "" {
		common.BadRequest(ctx, "缺少 platform_id")
		return
	}
	platformID, err := strconv.ParseUint(platformIDStr, 10, 64)
	if err != nil {
		common.BadRequest(ctx, "无效 platform_id")
		return
	}
	list, err := c.platformSvc.ListChannels(platformID)
	if err != nil {
		common.ServerError(ctx, "查询失败")
		return
	}
	common.OK(ctx, gin.H{"list": list, "platform_id": platformIDStr})
}

func (c *PayPlatformController) AvailableChannels(ctx *gin.Context) {
	unassigned := ctx.Query("unassigned") == "1" || ctx.Query("unassigned") == "true"
	list, err := c.platformSvc.ListAvailableChannels(unassigned)
	if err != nil {
		common.ServerError(ctx, "查询失败")
		return
	}
	common.OK(ctx, gin.H{"list": list})
}

func (c *PayPlatformController) BindChannels(ctx *gin.Context) {
	var req service.PlatformBindChannelsRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		common.BadRequest(ctx, "参数错误")
		return
	}
	ids := make([]uint64, 0, len(req.ChannelIDs))
	for _, id := range req.ChannelIDs {
		ids = append(ids, id.Uint64())
	}
	if err := c.platformSvc.BindChannels(req.PlatformID.Uint64(), ids); err != nil {
		handleBizError(ctx, err)
		return
	}
	common.OKMessage(ctx, "绑定成功")
}

func (c *PayPlatformController) UnbindChannel(ctx *gin.Context) {
	var req service.PlatformUnbindChannelRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		common.BadRequest(ctx, "参数错误")
		return
	}
	if err := c.platformSvc.UnbindChannel(req.PlatformID.Uint64(), req.ChannelID.Uint64()); err != nil {
		handleBizError(ctx, err)
		return
	}
	common.OKMessage(ctx, "已解绑")
}

func (c *PayPlatformController) SetChannels(ctx *gin.Context) {
	var req service.PlatformSetChannelsRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		common.BadRequest(ctx, "参数错误")
		return
	}
	if req.PlatformID == 0 {
		common.BadRequest(ctx, "缺少 platform_id")
		return
	}
	ids := make([]uint64, 0, len(req.ChannelIDs))
	for _, id := range req.ChannelIDs {
		ids = append(ids, id.Uint64())
	}
	if err := c.platformSvc.SetChannels(req.PlatformID.Uint64(), ids); err != nil {
		handleBizError(ctx, err)
		return
	}
	common.OKMessage(ctx, "绑定已更新")
}

func (c *PayPlatformController) Pool(ctx *gin.Context) {
	platformIDStr := ctx.Query("platform_id")
	if platformIDStr == "" {
		common.BadRequest(ctx, "缺少 platform_id")
		return
	}
	platformID, err := strconv.ParseUint(platformIDStr, 10, 64)
	if err != nil {
		common.BadRequest(ctx, "无效 platform_id")
		return
	}
	var payType *model.PayType
	if pt := ctx.Query("pay_type"); pt != "" {
		v, err := strconv.ParseInt(pt, 10, 8)
		if err == nil {
			p := model.PayType(v)
			payType = &p
		}
	}
	list, err := c.platformSvc.PoolStats(platformID, payType)
	if err != nil {
		common.ServerError(ctx, "查询失败")
		return
	}
	masked := make([]map[string]interface{}, 0, len(list))
	for i := range list {
		masked = append(masked, service.MaskChannel(&list[i]))
	}
	common.OK(ctx, gin.H{"list": masked, "platform_id": platformIDStr})
}

func (c *PayPlatformController) RegenerateKey(ctx *gin.Context) {
	var req struct {
		ID common.FlexUint64 `json:"id" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		common.BadRequest(ctx, "参数错误")
		return
	}
	p, err := c.platformSvc.RegenerateAppKey(req.ID.Uint64())
	if err != nil {
		handleBizError(ctx, err)
		return
	}
	common.OK(ctx, p)
}

func (c *PayPlatformController) QuickSetup(ctx *gin.Context) {
	var req service.PlatformQuickSetupRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		common.BadRequest(ctx, "参数错误")
		return
	}
	data, err := c.platformSvc.QuickSetup(req)
	if err != nil {
		handleBizError(ctx, err)
		return
	}
	common.OK(ctx, data)
}

func (c *PayPlatformController) AddChannel(ctx *gin.Context) {
	var req service.PlatformAddChannelRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		common.BadRequest(ctx, "参数错误")
		return
	}
	data, err := c.platformSvc.AddChannel(req)
	if err != nil {
		handleBizError(ctx, err)
		return
	}
	common.OK(ctx, data)
}

func (c *PayPlatformController) Detail(ctx *gin.Context) {
	platformIDStr := ctx.Query("platform_id")
	if platformIDStr == "" {
		platformIDStr = ctx.Query("id")
	}
	if platformIDStr == "" {
		common.BadRequest(ctx, "缺少 platform_id")
		return
	}
	platformID, err := strconv.ParseUint(platformIDStr, 10, 64)
	if err != nil {
		common.BadRequest(ctx, "无效 platform_id")
		return
	}
	data, err := c.platformSvc.GetDetail(platformID)
	if err != nil {
		handleBizError(ctx, err)
		return
	}
	common.OK(ctx, data)
}
