package service



import (

	"strings"

	"time"



	"wpay/common"

	"wpay/dao"

	"wpay/model"

)



type OrderService struct {
	orderDAO    *dao.OrderDAO
	paymentSvc  *PaymentService
	poolSvc     *ChannelPoolService
	platformSvc *PayPlatformService
}

func NewOrderService() *OrderService {
	return &OrderService{
		orderDAO:    dao.NewOrderDAO(),
		paymentSvc:  NewPaymentService(),
		poolSvc:     NewChannelPoolService(),
		platformSvc: NewPayPlatformService(),
	}
}



type CreateOrderRequest struct {

	StoreID     common.FlexUint64 `json:"store_id"`

	Amount      int64             `json:"amount" binding:"required,gt=0"`

	PayType     model.PayType     `json:"pay_type"`

	ChannelType string            `json:"channel_type"`

	PayScene    string            `json:"pay_scene"`    // h5(默认) | native 扫码兜底

	ReturnURL   string            `json:"return_url"`   // 支付完成回跳（建议传收银页地址）

	QuitURL     string            `json:"quit_url"`     // 支付宝取消支付回跳

	DeviceSN    string            `json:"device_sn"`

	Subject     string            `json:"subject"`

	BizRemark   string            `json:"biz_remark"`

	UsePool     *bool             `json:"use_pool"`

	AppKey      string            `json:"app_key"`
}



type CreateOrderResponse struct {

	OrderID   string        `json:"order_id"`

	PayScene  model.PayScene `json:"pay_scene"`

	PayURL    string        `json:"pay_url"`     // H5/WAP 跳转链接（pay_scene=h5）

	QRCodeURL string        `json:"qr_code_url"` // 扫码链接（pay_scene=native 或兜底）

	Amount    int64         `json:"amount"`

	StoreID   uint64        `json:"store_id,string"`

	ChannelID uint64        `json:"channel_id,string"`

	MchNo     string        `json:"mch_no"`

	PayType   model.PayType `json:"pay_type"`

}



func (s *OrderService) Create(req CreateOrderRequest, clientIP string, appKey string) (*CreateOrderResponse, error) {
	payType := s.resolvePayType(req)
	if !payType.Valid() {
		return nil, common.ErrInvalidInput("pay_type 无效，1=微信 2=支付宝")
	}

	platformID := uint64(0)
	if appKey == "" {
		appKey = strings.TrimSpace(req.AppKey)
	}
	platform, err := s.platformSvc.Resolve(appKey, clientIP)
	if err != nil {
		return nil, err
	}
	if platform != nil {
		platformID = platform.ID
	}

	subject := req.Subject

	if subject == "" {

		subject = req.BizRemark

	}

	for _, field := range []string{req.DeviceSN, subject, req.ReturnURL, req.QuitURL} {

		if err := common.SanitizeString(field); err != nil {

			return nil, err

		}

	}

	if req.Amount <= 0 {

		return nil, common.ErrInvalidInput("订单金额必须大于0")

	}



	usePool := req.UsePool == nil || *req.UsePool

	store, ch, err := s.pickChannel(req, payType, usePool, platformID)

	if err != nil {

		return nil, err

	}



	orderNo, err := common.GenerateOrderNo()

	if err != nil {

		return nil, err

	}

	if subject == "" {

		subject = store.StoreName + "收银"

	}



	payScene := model.ParsePayScene(strings.TrimSpace(req.PayScene))

	mobile := MobilePayContext{

		ClientIP:  clientIP,

		ReturnURL: req.ReturnURL,

		QuitURL:   req.QuitURL,

	}



	payURL, qrURL, finalScene, err := s.createPayment(payType, payScene, store, ch, orderNo, req.Amount, subject, mobile)

	if err != nil {

		return nil, common.ErrInvalidInput("创建支付订单失败: " + err.Error())

	}



	payLink := payURL

	if payLink == "" {

		payLink = qrURL

	}



	order := &model.Order{
		OrderNo:     orderNo,
		PlatformID:  platformID,
		StoreID:     store.ID,

		ChannelID:   ch.ID,

		PayType:     payType,

		TotalAmount: req.Amount,

		OrderStatus: model.OrderStatusPending,

		DeviceSN:    req.DeviceSN,

		Subject:     subject,

		QRCodeURL:   payLink,

		PayScene:    string(finalScene),

	}

	if err := s.orderDAO.Create(order); err != nil {

		return nil, err

	}

	return &CreateOrderResponse{

		OrderID:   orderNo,

		PayScene:  finalScene,

		PayURL:    payURL,

		QRCodeURL: qrURL,

		Amount:    req.Amount,

		StoreID:   store.ID,

		ChannelID: ch.ID,

		MchNo:     ch.MchNo,

		PayType:   payType,

	}, nil

}



func (s *OrderService) pickChannel(req CreateOrderRequest, payType model.PayType, usePool bool, platformID uint64) (*model.Store, *model.PayChannel, error) {
	var store *model.Store
	var ch *model.PayChannel
	var err error

	if usePool {
		store, ch, err = s.poolSvc.Select(platformID, req.StoreID.Uint64(), payType, req.Amount)
		// 已识别代收平台时禁止回退到任意门店渠道，避免跨平台串池
		// platformID > 0 时，即使平台池无可用码，也不允许回退到公共池或门店渠道
		if err != nil && platformID == 0 && req.StoreID != 0 {
			store, ch, err = s.paymentSvc.GetStoreChannel(req.StoreID.Uint64(), payType)
			if err == nil && ch != nil && !ch.CanAccept(req.Amount) {
				return nil, nil, common.ErrInvalidInput("该商户码已达额度上限或单笔超限")
			}
		}
	} else {
		if platformID != 0 {
			return nil, nil, common.ErrInvalidInput("代收平台订单须走码池轮询，不可指定固定单码")
		}
		if req.StoreID == 0 {
			return nil, nil, common.ErrInvalidInput("固定单码模式需指定 store_id")
		}
		store, ch, err = s.paymentSvc.GetStoreChannel(req.StoreID.Uint64(), payType)
		if err == nil && ch != nil && !ch.CanAccept(req.Amount) {
			return nil, nil, common.ErrInvalidInput("该商户码已达额度上限或单笔超限")
		}
	}

	if err != nil {

		return nil, nil, err

	}

	return store, ch, nil

}



func (s *OrderService) createPayment(

	payType model.PayType,

	scene model.PayScene,

	store *model.Store,

	ch *model.PayChannel,

	orderNo string,

	amount int64,

	subject string,

	mobile MobilePayContext,

) (payURL, qrURL string, finalScene model.PayScene, err error) {

	if scene == model.PaySceneH5 {

		payURL, err = s.invokeH5OrWap(payType, store, ch, orderNo, amount, subject, mobile)

		if err == nil && payURL != "" {

			return payURL, "", model.PaySceneH5, nil

		}

		common.Log.Warn("h5/wap pay failed, fallback native: %v", err)

	}



	qrURL, err = s.invokeNative(payType, store, ch, orderNo, amount, subject)

	if err != nil {

		return "", "", model.PaySceneNative, err

	}

	return "", qrURL, model.PaySceneNative, nil

}



func (s *OrderService) invokeH5OrWap(payType model.PayType, store *model.Store, ch *model.PayChannel, orderNo string, amount int64, subject string, mobile MobilePayContext) (string, error) {

	switch payType {

	case model.PayTypeWechat:

		return s.paymentSvc.CreateWechatH5Order(store, ch, orderNo, amount, subject, mobile)

	case model.PayTypeAlipay:

		return s.paymentSvc.CreateAlipayWapOrder(store, ch, orderNo, amount, subject, mobile)

	default:

		return "", common.ErrInvalidInput("unsupported pay type")

	}

}



func (s *OrderService) invokeNative(payType model.PayType, store *model.Store, ch *model.PayChannel, orderNo string, amount int64, subject string) (string, error) {

	switch payType {

	case model.PayTypeWechat:

		return s.paymentSvc.CreateWechatNativeOrder(store, ch, orderNo, amount, subject)

	case model.PayTypeAlipay:

		return s.paymentSvc.CreateAlipayPrecreateOrder(store, ch, orderNo, amount, subject)

	default:

		return "", common.ErrInvalidInput("unsupported pay type")

	}

}



func (s *OrderService) resolvePayType(req CreateOrderRequest) model.PayType {

	if req.PayType.Valid() {

		return req.PayType

	}

	if req.ChannelType != "" {

		return model.PayTypeFromString(req.ChannelType)

	}

	return 0

}



func (s *OrderService) GetByOrderNo(orderNo string) (*model.Order, error) {

	if err := common.SanitizeString(orderNo); err != nil {

		return nil, err

	}

	return s.orderDAO.FindByOrderNo(orderNo)

}



func (s *OrderService) Query(q dao.OrderQuery) ([]model.Order, int64, error) {

	if q.Page < 1 {

		q.Page = 1

	}

	if q.PageSize < 1 || q.PageSize > 200 {

		q.PageSize = 20

	}

	return s.orderDAO.Query(q)

}



func (s *OrderService) CloseTimeoutOrders(timeoutMinutes int) (int64, error) {

	before := time.Now().Add(-time.Duration(timeoutMinutes) * time.Minute)

	return s.orderDAO.CloseTimeoutOrders(before)

}

