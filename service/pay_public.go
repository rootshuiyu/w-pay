package service

import (
	"strings"
	"time"

	"wpay/model"
)

// PayCreateResponse 收银端对外创建订单响应（不含内部渠道路由信息）
type PayCreateResponse struct {
	OrderID   string         `json:"order_id"`
	PayScene  model.PayScene `json:"pay_scene"`
	PayURL    string         `json:"pay_url"`
	QRCodeURL string         `json:"qr_code_url"`
	Amount    int64          `json:"amount"`
	PayType   model.PayType  `json:"pay_type"`
}

func ToPayCreateResponse(r *CreateOrderResponse) *PayCreateResponse {
	if r == nil {
		return nil
	}
	return &PayCreateResponse{
		OrderID:   r.OrderID,
		PayScene:  r.PayScene,
		PayURL:    r.PayURL,
		QRCodeURL: r.QRCodeURL,
		Amount:    r.Amount,
		PayType:   r.PayType,
	}
}

// PayQueryResponse 收银端对外查单响应
type PayQueryResponse struct {
	OrderID       string            `json:"order_id"`
	OrderStatus   model.OrderStatus `json:"order_status"`
	TotalAmount   int64             `json:"total_amount"`
	PayAmount     int64             `json:"pay_amount"`
	PayType       model.PayType     `json:"pay_type"`
	PayScene      string            `json:"pay_scene"`
	QRCodeURL     string            `json:"qr_code_url"`
	PayTime       *time.Time        `json:"pay_time"`
	TransactionID string            `json:"transaction_id"`
}

func ToPayQueryResponse(o *model.Order) *PayQueryResponse {
	if o == nil {
		return nil
	}
	return &PayQueryResponse{
		OrderID:       o.OrderNo,
		OrderStatus:   o.OrderStatus,
		TotalAmount:   o.TotalAmount,
		PayAmount:     o.PayAmount,
		PayType:       o.PayType,
		PayScene:      o.PayScene,
		QRCodeURL:     o.QRCodeURL,
		PayTime:       o.PayTime,
		TransactionID: o.TransactionID,
	}
}

// SanitizePayBizMessage 对外收银 API 错误文案中性化，不暴露内部路由/轮询细节
func SanitizePayBizMessage(msg string) string {
	lower := strings.ToLower(msg)
	switch {
	case strings.Contains(msg, "轮询池"),
		strings.Contains(msg, "商户码"),
		strings.Contains(msg, "参与轮询"),
		strings.Contains(msg, "额度上限"),
		strings.Contains(msg, "单笔超限"),
		strings.Contains(msg, "该平台暂无"),
		strings.Contains(msg, "app_key"),
		strings.Contains(msg, "代收平台"):
		return "支付通道暂不可用，请稍后重试"
	case strings.Contains(lower, "store_id"):
		return "参数错误"
	default:
		return msg
	}
}
