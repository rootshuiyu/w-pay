package service

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"

	"wpay/common"
	"wpay/config"
	"wpay/dao"
	"wpay/model"

	"github.com/go-pay/gopay/alipay"
	"github.com/go-pay/gopay/wechat/v3"
)

type CallbackService struct {
	orderDAO    *dao.OrderDAO
	channelDAO  *dao.PayChannelDAO
	historyDAO  *dao.PayChannelHistoryDAO
	paymentSvc  *PaymentService
}

func NewCallbackService() *CallbackService {
	return &CallbackService{
		orderDAO:   dao.NewOrderDAO(),
		channelDAO: dao.NewPayChannelDAO(),
		historyDAO: dao.NewPayChannelHistoryDAO(),
		paymentSvc: NewPaymentService(),
	}
}

type wechatNotifyResult struct {
	OutTradeNo    string
	TransactionID string
	TradeState    string
	SuccessTime   string
	Total         int
}

// HandleWechat 微信回调：优先 channel_id 对应渠道，再尝试同店其他渠道 + 历史密钥
func (s *CallbackService) HandleWechat(req *http.Request, storeID, channelID uint64) error {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return err
	}

	channels, err := s.channelDAO.ListByStoreAndType(storeID, model.PayTypeWechat)
	if err != nil {
		return err
	}
	if len(channels) == 0 {
		return common.ErrInvalidInput("渠道不存在")
	}
	channels = prioritizeChannel(channels, channelID)

	var lastErr error
	for _, base := range channels {
		candidates := []model.PayChannel{base}
		histories, _ := s.historyDAO.ListValidByChannelID(base.ID)
		for i := range histories {
			candidates = append(candidates, histories[i].ToPayChannel())
		}
		for _, cand := range candidates {
			req.Body = io.NopCloser(bytes.NewReader(body))
			result, verifyErr := s.tryWechatNotify(req, &cand)
			if verifyErr != nil {
				lastErr = verifyErr
				continue
			}
			if result == nil || result.TradeState != "SUCCESS" {
				return nil
			}
			notifyData := common.DesensitizeLog(fmt.Sprintf("trade_state=%s;txn=%s;mch=%s",
				result.TradeState, result.TransactionID, cand.MchNo))
			return s.finishPaid(result.OutTradeNo, storeID, model.PayTypeWechat,
				result.TransactionID, int64(result.Total), notifyData, result.SuccessTime)
		}
	}
	if lastErr != nil {
		return lastErr
	}
	return common.ErrInvalidInput("验签失败")
}

func prioritizeChannel(channels []model.PayChannel, channelID uint64) []model.PayChannel {
	if channelID == 0 || len(channels) <= 1 {
		return channels
	}
	out := make([]model.PayChannel, 0, len(channels))
	for _, ch := range channels {
		if ch.ID == channelID {
			out = append(out, ch)
		}
	}
	for _, ch := range channels {
		if ch.ID != channelID {
			out = append(out, ch)
		}
	}
	return out
}

func (s *CallbackService) tryWechatNotify(req *http.Request, ch *model.PayChannel) (*wechatNotifyResult, error) {
	notifyReq, err := wechat.V3ParseNotify(req)
	if err != nil {
		return nil, err
	}
	client, err := s.paymentSvc.BuildWechatNotifyClient(ch)
	if err != nil {
		return nil, err
	}
	if err := notifyReq.VerifySignByPKMap(client.WxPublicKeyMap()); err != nil {
		return nil, common.ErrInvalidInput("验签失败")
	}
	result, err := notifyReq.DecryptPayCipherText(ch.MchKey)
	if err != nil {
		return nil, err
	}
	return &wechatNotifyResult{
		OutTradeNo:    result.OutTradeNo,
		TransactionID: result.TransactionId,
		TradeState:    result.TradeState,
		SuccessTime:   result.SuccessTime,
		Total:         result.Amount.Total,
	}, nil
}

// HandleAlipay 支付宝回调：优先当前密钥，失败则尝试该订单渠道的历史密钥
func (s *CallbackService) HandleAlipay(req *http.Request) error {
	notifyReq, err := alipay.ParseNotifyToBodyMap(req)
	if err != nil {
		return common.ErrInvalidInput("回调解析失败")
	}
	orderNo := notifyReq.Get("out_trade_no")
	order, err := s.orderDAO.FindByOrderNo(orderNo)
	if err != nil || order == nil {
		return common.ErrInvalidInput("订单不存在")
	}

	var ch *model.PayChannel
	if order.ChannelID > 0 {
		ch, err = s.channelDAO.FindByID(order.ChannelID)
	}
	if ch == nil {
		ch, err = s.channelDAO.FindByStoreAndType(order.StoreID, model.PayTypeAlipay)
	}
	if err != nil || ch == nil {
		return common.ErrInvalidInput("渠道不存在")
	}

	candidates := []model.PayChannel{*ch}
	histories, _ := s.historyDAO.ListValidByChannelID(ch.ID)
	for i := range histories {
		candidates = append(candidates, histories[i].ToPayChannel())
	}

	verified := false
	for _, cand := range candidates {
		ok, verr := alipay.VerifySign(cand.PublicKey, notifyReq)
		if verr == nil && ok {
			verified = true
			break
		}
	}
	if !verified {
		return common.ErrInvalidInput("验签失败")
	}

	tradeStatus := notifyReq.Get("trade_status")
	if tradeStatus != "TRADE_SUCCESS" && tradeStatus != "TRADE_FINISHED" {
		return nil
	}
	notifyData := common.DesensitizeLog(fmt.Sprintf("trade_status=%s;txn=%s", tradeStatus, notifyReq.Get("trade_no")))
	return s.finishPaid(orderNo, order.StoreID, model.PayTypeAlipay,
		notifyReq.Get("trade_no"), ParseAlipayPaidAmount(notifyReq.Get("total_amount")), notifyData, "")
}

func (s *CallbackService) finishPaid(orderNo string, storeID uint64, payType model.PayType,
	transactionID string, payAmount int64, notifyData string, successTime string) error {

	order, err := s.orderDAO.FindByOrderNo(orderNo)
	if err != nil || order == nil {
		return common.ErrInvalidInput("订单不存在")
	}
	if order.StoreID != storeID {
		return common.ErrInvalidInput("门店不匹配")
	}

	expire := time.Duration(config.Global.Callback.IdempotentExpireHours) * time.Hour
	set, err := dao.SetCallbackIdempotent(payType.String(), transactionID, expire)
	if err != nil {
		return err
	}
	if !set {
		return nil
	}

	payTime := time.Now()
	if successTime != "" {
		if t, e := time.Parse(time.RFC3339, successTime); e == nil {
			payTime = t
		}
	}
	if err := s.orderDAO.UpdateStatusPaid(orderNo, payAmount, transactionID, payTime, notifyData); err != nil {
		return err
	}
	if order.ChannelID > 0 && payAmount > 0 {
		_ = s.channelDAO.AddDailyUsed(order.ChannelID, payAmount)
	}
	return nil
}
