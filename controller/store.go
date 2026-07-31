package controller

import (
	"strconv"

	"wpay/common"
	"wpay/service"

	"github.com/gin-gonic/gin"
)

type StoreController struct {
	storeSvc *service.StoreService
}

func NewStoreController() *StoreController {
	return &StoreController{storeSvc: service.NewStoreService()}
}

// Add POST /api/admin/store/add
func (c *StoreController) Add(ctx *gin.Context) {
	c.Create(ctx)
}

func (c *StoreController) Create(ctx *gin.Context) {
	var req service.StoreCreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		common.BadRequest(ctx, "参数错误")
		return
	}
	store, err := c.storeSvc.Create(req)
	if err != nil {
		handleBizError(ctx, err)
		return
	}
	common.OK(ctx, store)
}

// Edit PUT /api/admin/store/edit
func (c *StoreController) Edit(ctx *gin.Context) {
	var req service.StoreEditRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		common.BadRequest(ctx, "参数错误")
		return
	}
	if req.ID == 0 {
		common.BadRequest(ctx, "缺少门店 id")
		return
	}
	update := service.StoreUpdateRequest{
		StoreName:  req.StoreName,
		Address:    req.Address,
		TaxSubject: req.TaxSubject,
		Remark:     req.Remark,
	}
	store, err := c.storeSvc.Update(req.ID.Uint64(), update)
	if err != nil {
		handleBizError(ctx, err)
		return
	}
	common.OK(ctx, store)
}

// UpdateStatus PUT /api/admin/store/status
func (c *StoreController) UpdateStatus(ctx *gin.Context) {
	var req service.StoreStatusRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		common.BadRequest(ctx, "参数错误")
		return
	}
	if req.ID == 0 {
		common.BadRequest(ctx, "缺少门店 id")
		return
	}
	status := req.Status
	store, err := c.storeSvc.Update(req.ID.Uint64(), service.StoreUpdateRequest{Status: &status})
	if err != nil {
		handleBizError(ctx, err)
		return
	}
	common.OK(ctx, store)
}

// Del DELETE /api/admin/store/del
func (c *StoreController) Del(ctx *gin.Context) {
	var req service.StoreDeleteRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		// 兼容 query 传参
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
		common.BadRequest(ctx, "缺少门店 id")
		return
	}
	if err := c.storeSvc.Delete(req.ID.Uint64()); err != nil {
		handleBizError(ctx, err)
		return
	}
	common.OKMessage(ctx, "删除成功")
}

// List GET /api/admin/store/list
func (c *StoreController) List(ctx *gin.Context) {
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
	hideSystem := ctx.DefaultQuery("hide_system", "1") != "0"
	list, total, err := c.storeSvc.List(keyword, status, hideSystem, page, pageSize)
	if err != nil {
		common.ServerError(ctx, "查询失败")
		return
	}
	common.OK(ctx, gin.H{"list": list, "total": total, "page": page, "page_size": pageSize})
}
