package controller

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"wpay/common"
	"wpay/model"
	"wpay/service"

	"github.com/gin-gonic/gin"
)

type ReconcileController struct {
	reconcileSvc *service.ReconcileService
}

func NewReconcileController() *ReconcileController {
	return &ReconcileController{reconcileSvc: service.NewReconcileService()}
}

func (c *ReconcileController) Stat(ctx *gin.Context) {
	q, err := buildReconcileQuery(ctx)
	if err != nil {
		common.BadRequest(ctx, err.Error())
		return
	}
	resp, err := c.reconcileSvc.Stat(q)
	if err != nil {
		common.ServerError(ctx, "统计失败")
		return
	}
	common.OK(ctx, resp)
}

func (c *ReconcileController) ExportOrders(ctx *gin.Context) {
	oq, err := buildOrderQuery(ctx)
	if err != nil {
		common.BadRequest(ctx, err.Error())
		return
	}
	f, err := c.reconcileSvc.ExportExcel(oq)
	if err != nil {
		common.ServerError(ctx, "导出失败")
		return
	}
	filename := fmt.Sprintf("orders_%s.xlsx", time.Now().Format("20060102150405"))
	ctx.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	if err := f.Write(ctx.Writer); err != nil {
		common.ServerError(ctx, "写入失败")
	}
}

func (c *ReconcileController) ExportStat(ctx *gin.Context) {
	q, err := buildReconcileQuery(ctx)
	if err != nil {
		common.BadRequest(ctx, err.Error())
		return
	}
	f, err := c.reconcileSvc.ExportStatExcel(q)
	if err != nil {
		common.ServerError(ctx, "导出失败")
		return
	}
	filename := fmt.Sprintf("stat_%s.xlsx", time.Now().Format("20060102150405"))
	ctx.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	if err := f.Write(ctx.Writer); err != nil {
		common.ServerError(ctx, "写入失败")
	}
}

func buildReconcileQuery(ctx *gin.Context) (service.ReconcileQuery, error) {
	q := service.ReconcileQuery{
		GroupByDay: ctx.DefaultQuery("group_by", "day") != "month",
		Dimension:  ctx.DefaultQuery("dimension", "store"),
	}

	if ids := ctx.Query("store_ids"); ids != "" {
		for _, part := range strings.Split(ids, ",") {
			id, err := strconv.ParseUint(strings.TrimSpace(part), 10, 64)
			if err == nil {
				q.StoreIDs = append(q.StoreIDs, id)
			}
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
	if s := ctx.Query("status"); s != "" {
		v, err := strconv.ParseInt(s, 10, 8)
		if err == nil {
			st := model.OrderStatus(v)
			q.Status = &st
		}
	}

	startStr := ctx.DefaultQuery("start_time", time.Now().AddDate(0, -1, 0).Format("2006-01-02"))
	endStr := ctx.DefaultQuery("end_time", time.Now().Format("2006-01-02"))

	start, err := time.Parse("2006-01-02", startStr)
	if err != nil {
		return q, common.ErrInvalidInput("start_time 格式错误")
	}
	end, err := time.Parse("2006-01-02", endStr)
	if err != nil {
		return q, common.ErrInvalidInput("end_time 格式错误")
	}
	q.StartTime = start
	q.EndTime = end.Add(24*time.Hour - time.Second)
	return q, nil
}

// Health 健康检查
func Health(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"status": "ok", "service": common.ServiceName, "name": common.ServiceDisplayName})
}

// Index 根路径欢迎页
func Index(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{
		"service": common.ServiceName,
		"name":    common.ServiceDisplayName,
		"status":  "ok",
		"docs":    "docs/API.md",
		"health":  "/health",
		"api":     []string{"/api/pay/create", "/api/pay/query", "/api/admin/login", "/api/admin/..."},
	})
}
