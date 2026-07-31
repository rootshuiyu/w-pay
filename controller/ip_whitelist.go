package controller

import (
	"wpay/common"
	"wpay/service"

	"github.com/gin-gonic/gin"
)

type IPWhitelistController struct {
	svc *service.IPWhitelistService
}

func NewIPWhitelistController() *IPWhitelistController {
	return &IPWhitelistController{svc: service.NewIPWhitelistService()}
}

// Overview GET /api/admin/whitelist/overview?scope=
func (c *IPWhitelistController) Overview(ctx *gin.Context) {
	data, err := c.svc.Overview(ctx.Query("scope"))
	if err != nil {
		handleBizError(ctx, err)
		return
	}
	common.OK(ctx, data)
}

// UpdatePolicy PUT /api/admin/whitelist/policy
func (c *IPWhitelistController) UpdatePolicy(ctx *gin.Context) {
	var req service.IPWhitelistPolicyRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		common.BadRequest(ctx, "参数错误")
		return
	}
	if err := c.svc.UpdatePolicy(req); err != nil {
		handleBizError(ctx, err)
		return
	}
	common.OKMessage(ctx, "策略已更新")
}

// Add POST /api/admin/whitelist/add
func (c *IPWhitelistController) Add(ctx *gin.Context) {
	var req service.IPWhitelistAddRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		common.BadRequest(ctx, "参数错误")
		return
	}
	entry, err := c.svc.Add(req)
	if err != nil {
		handleBizError(ctx, err)
		return
	}
	common.OK(ctx, entry)
}

// Edit PUT /api/admin/whitelist/edit
func (c *IPWhitelistController) Edit(ctx *gin.Context) {
	var req service.IPWhitelistEditRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		common.BadRequest(ctx, "参数错误")
		return
	}
	entry, err := c.svc.Edit(req)
	if err != nil {
		handleBizError(ctx, err)
		return
	}
	common.OK(ctx, entry)
}

// UpdateStatus PUT /api/admin/whitelist/status
func (c *IPWhitelistController) UpdateStatus(ctx *gin.Context) {
	var req service.IPWhitelistStatusRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		common.BadRequest(ctx, "参数错误")
		return
	}
	entry, err := c.svc.UpdateStatus(req)
	if err != nil {
		handleBizError(ctx, err)
		return
	}
	common.OK(ctx, entry)
}

// Del DELETE /api/admin/whitelist/del
func (c *IPWhitelistController) Del(ctx *gin.Context) {
	var req service.IPWhitelistDeleteRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		common.BadRequest(ctx, "参数错误")
		return
	}
	if req.ID == 0 {
		common.BadRequest(ctx, "缺少 id")
		return
	}
	if err := c.svc.Delete(req.ID.Uint64()); err != nil {
		handleBizError(ctx, err)
		return
	}
	common.OKMessage(ctx, "删除成功")
}
