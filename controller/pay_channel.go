package controller

import (
	"strconv"

	"wpay/common"
	"wpay/model"
	"wpay/service"

	"github.com/gin-gonic/gin"
)

type PayChannelController struct {
	channelSvc *service.PayChannelService
	poolSvc    *service.ChannelPoolService
}

func NewPayChannelController() *PayChannelController {
	return &PayChannelController{
		channelSvc: service.NewPayChannelService(),
		poolSvc:    service.NewChannelPoolService(),
	}
}

// Add POST /api/admin/channel/add
func (c *PayChannelController) Add(ctx *gin.Context) {
	var req service.ChannelCreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		common.BadRequest(ctx, "参数错误")
		return
	}
	ch, err := c.channelSvc.Create(req)
	if err != nil {
		handleBizError(ctx, err)
		return
	}
	common.OK(ctx, service.MaskChannel(ch))
}

// PoolAdd POST /api/admin/channel/pool-add — 公共码池入库
func (c *PayChannelController) PoolAdd(ctx *gin.Context) {
	var req service.PoolAddRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		common.BadRequest(ctx, "参数错误")
		return
	}
	ch, err := c.channelSvc.PoolAdd(req)
	if err != nil {
		handleBizError(ctx, err)
		return
	}
	common.OK(ctx, service.MaskChannel(ch))
}

// Edit PUT /api/admin/channel/edit — 更换商户号/密钥核心接口
func (c *PayChannelController) Edit(ctx *gin.Context) {
	var req service.ChannelEditRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		common.BadRequest(ctx, "参数错误")
		return
	}
	if req.ID == 0 {
		common.BadRequest(ctx, "缺少渠道 id")
		return
	}
	update := service.ChannelUpdateRequest{
		PoolEnabled:    req.PoolEnabled,
		DailyLimitFen:  req.DailyLimitFen,
		SingleLimitFen: req.SingleLimitFen,
		MchNo:          req.MchNo,
		MchKey:     req.MchKey,
		AppID:      req.AppID,
		SerialNo:   req.SerialNo,
		PrivateKey: req.PrivateKey,
		PublicKey:  req.PublicKey,
		NotifyURL:  req.NotifyURL,
		CertFile:   req.CertFile,
		Remark:     req.Remark,
	}
	ch, err := c.channelSvc.Update(req.ID.Uint64(), update)
	if err != nil {
		handleBizError(ctx, err)
		return
	}
	common.OK(ctx, service.MaskChannel(ch))
}

// UpdateStatus PUT /api/admin/channel/status — 临时关停微信/支付宝通道
func (c *PayChannelController) UpdateStatus(ctx *gin.Context) {
	var req service.ChannelStatusRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		common.BadRequest(ctx, "参数错误")
		return
	}
	if req.ID == 0 {
		common.BadRequest(ctx, "缺少渠道 id")
		return
	}
	status := req.Status
	ch, err := c.channelSvc.Update(req.ID.Uint64(), service.ChannelUpdateRequest{Status: &status})
	if err != nil {
		handleBizError(ctx, err)
		return
	}
	common.OK(ctx, service.MaskChannel(ch))
}

// List GET /api/admin/channel/list?store_id=
func (c *PayChannelController) List(ctx *gin.Context) {
	storeIDStr := ctx.Query("store_id")
	if storeIDStr == "" {
		common.BadRequest(ctx, "缺少 store_id")
		return
	}
	storeID, err := strconv.ParseUint(storeIDStr, 10, 64)
	if err != nil {
		common.BadRequest(ctx, "无效 store_id")
		return
	}
	list, err := c.channelSvc.ListByStore(storeID)
	if err != nil {
		common.ServerError(ctx, "查询失败")
		return
	}
	masked := make([]map[string]interface{}, 0, len(list))
	for i := range list {
		masked = append(masked, service.MaskChannel(&list[i]))
	}
	common.OK(ctx, gin.H{"list": masked, "store_id": storeID})
}

// Pool GET /api/admin/channel/pool — 轮询池商户码及额度用量
func (c *PayChannelController) Pool(ctx *gin.Context) {
	var payType *model.PayType
	if pt := ctx.Query("pay_type"); pt != "" {
		v, err := strconv.ParseInt(pt, 10, 8)
		if err == nil {
			p := model.PayType(v)
			payType = &p
		}
	}
	var platformID *uint64
	if pid := ctx.Query("platform_id"); pid != "" {
		v, err := strconv.ParseUint(pid, 10, 64)
		if err == nil {
			platformID = &v
		}
	}
	list, err := c.poolSvc.PoolStatsForPlatform(payType, platformID)
	if err != nil {
		common.ServerError(ctx, "查询失败")
		return
	}
	masked := make([]map[string]interface{}, 0, len(list))
	for i := range list {
		masked = append(masked, service.MaskChannel(&list[i]))
	}
	common.OK(ctx, gin.H{"list": masked})
}
