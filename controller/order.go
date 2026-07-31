package controller

import (
	"strconv"
	"strings"
	"time"

	"wpay/common"
	"wpay/dao"
	"wpay/model"
	"wpay/service"

	"github.com/gin-gonic/gin"
)

type OrderController struct {
	orderSvc *service.OrderService
}

func NewOrderController() *OrderController {
	return &OrderController{orderSvc: service.NewOrderService()}
}

func (c *OrderController) Create(ctx *gin.Context) {
	var req service.CreateOrderRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		common.BadRequest(ctx, "参数错误")
		return
	}
	resp, err := c.orderSvc.Create(req, ctx.ClientIP(), "")
	if err != nil {
		handleBizError(ctx, err)
		return
	}
	common.OK(ctx, resp)
}

// PayCreate POST /api/pay/create — 对外收银，响应不含内部渠道路由字段
func (c *OrderController) PayCreate(ctx *gin.Context) {
	var req service.CreateOrderRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		common.BadRequest(ctx, "参数错误")
		return
	}
	appKey := strings.TrimSpace(ctx.GetHeader("X-App-Key"))
	if appKey == "" {
		appKey = ctx.Query("app_key")
	}
	if appKey == "" {
		appKey = strings.TrimSpace(req.AppKey)
	}
	resp, err := c.orderSvc.Create(req, ctx.ClientIP(), appKey)
	if err != nil {
		handlePayBizError(ctx, err)
		return
	}
	common.OK(ctx, service.ToPayCreateResponse(resp))
}

// PayQuery GET /api/pay/query — 对外收银查单
func (c *OrderController) PayQuery(ctx *gin.Context) {
	orderNo := ctx.Query("order_no")
	if orderNo == "" {
		orderNo = ctx.Query("order_id")
	}
	if orderNo == "" {
		common.BadRequest(ctx, "缺少 order_no")
		return
	}

	// 获取客户端IP和app_key
	clientIP := ctx.ClientIP()
	appKey := ctx.GetHeader("X-App-Key")
	if appKey == "" {
		appKey = ctx.Query("app_key")
	}

	// 识别平台
	platform, err := service.NewPayPlatformService().Resolve(appKey, clientIP)
	if err != nil && appKey != "" {
		// 如果明确传入 app_key，则按平台鉴权处理；否则继续查询公共订单
		handlePayBizError(ctx, err)
		return
	}

	order, err := c.orderSvc.GetByOrderNo(orderNo)
	if err != nil {
		handlePayBizError(ctx, err)
		return
	}
	if order == nil {
		common.NotFound(ctx, "订单不存在")
		return
	}

	// 如果订单属于平台码池订单，必须使用对应平台 app_key/IP 进行查询
	if order.PlatformID != 0 {
		if platform == nil || order.PlatformID != platform.ID {
			common.Forbidden(ctx, "无权查询此订单")
			return
		}
	}

	common.OK(ctx, service.ToPayQueryResponse(order))
}

func (c *OrderController) Get(ctx *gin.Context) {
	orderNo := ctx.Param("order_no")
	if orderNo == "" {
		orderNo = ctx.Param("order_id")
	}
	c.getByOrderNo(ctx, orderNo)
}

// QueryByParam GET /api/pay/query 或 /api/admin/order/detail?order_no=
func (c *OrderController) QueryByParam(ctx *gin.Context) {
	orderNo := ctx.Query("order_no")
	if orderNo == "" {
		orderNo = ctx.Query("order_id")
	}
	if orderNo == "" {
		common.BadRequest(ctx, "缺少 order_no")
		return
	}
	c.getByOrderNo(ctx, orderNo)
}

func (c *OrderController) getByOrderNo(ctx *gin.Context, orderNo string) {
	order, err := c.orderSvc.GetByOrderNo(orderNo)
	if err != nil {
		handleBizError(ctx, err)
		return
	}
	if order == nil {
		common.NotFound(ctx, "订单不存在")
		return
	}
	common.OK(ctx, order)
}

func (c *OrderController) List(ctx *gin.Context) {
	q, err := buildOrderQuery(ctx)
	if err != nil {
		common.BadRequest(ctx, err.Error())
		return
	}
	list, total, err := c.orderSvc.Query(q)
	if err != nil {
		common.ServerError(ctx, "查询失败")
		return
	}
	common.OK(ctx, gin.H{"list": list, "total": total, "page": q.Page, "page_size": q.PageSize})
}

func buildOrderQuery(ctx *gin.Context) (dao.OrderQuery, error) {
	q := dao.OrderQuery{
		OrderNo:  ctx.Query("order_no"),
		Page:     1,
		PageSize: 20,
	}
	if p, err := strconv.Atoi(ctx.DefaultQuery("page", "1")); err == nil {
		q.Page = p
	}
	if ps, err := strconv.Atoi(ctx.DefaultQuery("page_size", "20")); err == nil {
		q.PageSize = ps
	}
	if s := ctx.Query("order_status"); s != "" {
		v, err := strconv.ParseInt(s, 10, 8)
		if err == nil {
			st := model.OrderStatus(v)
			q.Status = &st
		}
	} else if s := ctx.Query("status"); s != "" {
		v, err := strconv.ParseInt(s, 10, 8)
		if err == nil {
			st := model.OrderStatus(v)
			q.Status = &st
		}
	}
	if pt := ctx.Query("pay_type"); pt != "" {
		v, err := strconv.ParseInt(pt, 10, 8)
		if err == nil {
			p := model.PayType(v)
			q.PayType = &p
		}
	} else if ct := ctx.Query("channel_type"); ct != "" {
		p := model.PayTypeFromString(ct)
		if p.Valid() {
			q.PayType = &p
		}
	}
	if ids := ctx.Query("platform_ids"); ids != "" {
		for _, part := range strings.Split(ids, ",") {
			id, err := strconv.ParseUint(strings.TrimSpace(part), 10, 64)
			if err == nil {
				q.PlatformIDs = append(q.PlatformIDs, id)
			}
		}
	}
	if ids := ctx.Query("store_ids"); ids != "" {
		for _, part := range strings.Split(ids, ",") {
			id, err := strconv.ParseUint(strings.TrimSpace(part), 10, 64)
			if err == nil {
				q.StoreIDs = append(q.StoreIDs, id)
			}
		}
	}
	if start := ctx.Query("start_time"); start != "" {
		t, err := time.Parse("2006-01-02", start)
		if err != nil {
			return q, common.ErrInvalidInput("start_time 格式应为 YYYY-MM-DD")
		}
		q.StartTime = &t
	}
	if end := ctx.Query("end_time"); end != "" {
		t, err := time.Parse("2006-01-02", end)
		if err != nil {
			return q, common.ErrInvalidInput("end_time 格式应为 YYYY-MM-DD")
		}
		endDay := t.Add(24*time.Hour - time.Second)
		q.EndTime = &endDay
	}
	return q, nil
}
