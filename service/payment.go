package service



import (

	"context"

	"fmt"

	"net/url"

	"strings"



	"wpay/common"

	"wpay/config"

	"wpay/dao"

	"wpay/model"



	"github.com/go-pay/gopay"

	"github.com/go-pay/gopay/alipay"

	"github.com/go-pay/gopay/wechat/v3"

)



// PaymentService 支付下单（go-pay/gopay）

type PaymentService struct{}



func NewPaymentService() *PaymentService { return &PaymentService{} }



// GetStoreChannel 读取门店最新渠道：优先 Redis，miss 回源 DB

func (s *PaymentService) GetStoreChannel(storeID uint64, payType model.PayType) (*model.Store, *model.PayChannel, error) {

	item, err := dao.GetChannelCache(storeID, payType)

	if err != nil {

		return nil, nil, err

	}

	if item == nil {

		item, err = dao.LoadAndCacheChannel(storeID, payType)

		if err != nil || item == nil {

			return nil, nil, common.ErrInvalidInput("门店渠道配置不存在")

		}

	}

	if item.Store.Status != model.StoreStatusNormal {

		return nil, nil, common.ErrInvalidInput("门店已停用")

	}

	if !item.Channel.Enabled() {

		return nil, nil, common.ErrInvalidInput("该门店支付渠道已关停")

	}

	return &item.Store, &item.Channel, nil

}



type MobilePayContext struct {

	ClientIP  string

	ReturnURL string

	QuitURL   string

}



func (s *PaymentService) CreateWechatNativeOrder(store *model.Store, ch *model.PayChannel, orderNo string, amount int64, remark string) (string, error) {

	if common.TestMode() {

		return fmt.Sprintf("mock://wechat/native/%s/%s/%d", ch.MchNo, orderNo, amount), nil

	}

	client, err := s.newWechatClient(ch)

	if err != nil {

		return "", err

	}

	bm := s.wechatBaseBody(ch, orderNo, amount, remark)

	wxRsp, err := client.V3TransactionNative(context.Background(), bm)

	if err != nil {

		return "", fmt.Errorf("wechat native: %w", err)

	}

	if wxRsp.Code != wechat.Success || wxRsp.Response == nil || wxRsp.Response.CodeUrl == "" {

		return "", fmt.Errorf("wechat error: %s", wxRsp.Error)

	}

	common.Log.Info("wechat native store=%d order=%s mch=%s", store.ID, orderNo, ch.MchNo)

	return wxRsp.Response.CodeUrl, nil

}



func (s *PaymentService) CreateWechatH5Order(store *model.Store, ch *model.PayChannel, orderNo string, amount int64, remark string, mobile MobilePayContext) (string, error) {

	if common.TestMode() {

		return fmt.Sprintf("mock://wechat/h5/%s/%s/%d", ch.MchNo, orderNo, amount), nil

	}

	client, err := s.newWechatClient(ch)

	if err != nil {

		return "", err

	}

	appName, appURL := h5SiteInfo(mobile.ReturnURL)

	clientIP := mobile.ClientIP

	if clientIP == "" {

		clientIP = "127.0.0.1"

	}

	bm := s.wechatBaseBody(ch, orderNo, amount, remark)

	bm.SetBodyMap("scene_info", func(b gopay.BodyMap) {

		b.Set("payer_client_ip", clientIP).

			SetBodyMap("h5_info", func(h gopay.BodyMap) {

				h.Set("type", "Wap").

					Set("app_name", appName).

					Set("app_url", appURL)

			})

	})

	wxRsp, err := client.V3TransactionH5(context.Background(), bm)

	if err != nil {

		return "", fmt.Errorf("wechat h5: %w", err)

	}

	if wxRsp.Code != wechat.Success || wxRsp.Response == nil || wxRsp.Response.H5Url == "" {

		return "", fmt.Errorf("wechat h5 error: %s", wxRsp.Error)

	}

	common.Log.Info("wechat h5 store=%d order=%s mch=%s", store.ID, orderNo, ch.MchNo)

	return wxRsp.Response.H5Url, nil

}



func (s *PaymentService) CreateAlipayPrecreateOrder(store *model.Store, ch *model.PayChannel, orderNo string, amount int64, remark string) (string, error) {

	if common.TestMode() {

		return fmt.Sprintf("mock://alipay/native/%s/%s/%d", ch.MchNo, orderNo, amount), nil

	}

	client, err := s.newAlipayClient(ch)

	if err != nil {

		return "", err

	}

	bm := s.alipayBaseBody(orderNo, amount, remark)

	aliRsp, err := client.TradePrecreate(context.Background(), bm)

	if err != nil {

		return "", fmt.Errorf("alipay precreate: %w", err)

	}

	if aliRsp.Response.Code != "10000" || aliRsp.Response.QrCode == "" {

		return "", fmt.Errorf("alipay error: %s", aliRsp.Response.SubMsg)

	}

	common.Log.Info("alipay precreate store=%d order=%s mch=%s", store.ID, orderNo, ch.MchNo)

	return aliRsp.Response.QrCode, nil

}



func (s *PaymentService) CreateAlipayWapOrder(store *model.Store, ch *model.PayChannel, orderNo string, amount int64, remark string, mobile MobilePayContext) (string, error) {

	if common.TestMode() {

		return fmt.Sprintf("mock://alipay/wap/%s/%s/%d", ch.MchNo, orderNo, amount), nil

	}

	client, err := s.newAlipayClient(ch)

	if err != nil {

		return "", err

	}

	bm := s.alipayBaseBody(orderNo, amount, remark)

	if u := pickURL(mobile.ReturnURL); u != "" {

		bm.Set("return_url", u)

	}

	if u := pickURL(mobile.QuitURL, mobile.ReturnURL); u != "" {

		bm.Set("quit_url", u)

	}

	payURL, err := client.TradeWapPay(context.Background(), bm)

	if err != nil {

		return "", fmt.Errorf("alipay wap: %w", err)

	}

	if payURL == "" {

		return "", fmt.Errorf("alipay wap: empty pay url")

	}

	common.Log.Info("alipay wap store=%d order=%s mch=%s", store.ID, orderNo, ch.MchNo)

	return payURL, nil

}



func (s *PaymentService) BuildWechatNotifyClient(ch *model.PayChannel) (*wechat.ClientV3, error) {

	return s.newWechatClient(ch)

}



func (s *PaymentService) newWechatClient(ch *model.PayChannel) (*wechat.ClientV3, error) {

	client, err := wechat.NewClientV3(ch.MchNo, ch.SerialNo, ch.MchKey, ch.PrivateKey)

	if err != nil {

		return nil, fmt.Errorf("init wechat client: %w", err)

	}

	_ = client.AutoVerifySign()

	return client, nil

}



func (s *PaymentService) newAlipayClient(ch *model.PayChannel) (*alipay.Client, error) {

	client, err := alipay.NewClient(ch.AppID, ch.PrivateKey, true)

	if err != nil {

		return nil, fmt.Errorf("init alipay: %w", err)

	}

	client.SetCharset("utf-8").SetSignType(alipay.RSA2).SetNotifyUrl(ch.NotifyURL)

	if ch.PublicKey != "" {

		client.AutoVerifySign([]byte(ch.PublicKey))

	}

	return client, nil

}



func (s *PaymentService) wechatBaseBody(ch *model.PayChannel, orderNo string, amount int64, remark string) gopay.BodyMap {

	bm := make(gopay.BodyMap)

	bm.Set("appid", ch.AppID).Set("mchid", ch.MchNo).

		Set("description", remark).Set("out_trade_no", orderNo).

		Set("notify_url", ch.NotifyURL).

		SetBodyMap("amount", func(b gopay.BodyMap) {

			b.Set("total", amount).Set("currency", "CNY")

		})

	return bm

}



func (s *PaymentService) alipayBaseBody(orderNo string, amount int64, remark string) gopay.BodyMap {

	amountYuan := fmt.Sprintf("%.2f", float64(amount)/100.0)

	bm := make(gopay.BodyMap)

	bm.Set("out_trade_no", orderNo).Set("total_amount", amountYuan).Set("subject", remark)

	return bm

}



func h5SiteInfo(returnURL string) (appName, appURL string) {

	appName = "聚合收银"

	appURL = "https://localhost"

	if config.Global != nil {

		if config.Global.Pay.H5AppName != "" {

			appName = config.Global.Pay.H5AppName

		}

		if config.Global.Pay.H5AppURL != "" {

			appURL = config.Global.Pay.H5AppURL

		}

	}

	if u := pickURL(returnURL); u != "" {

		appURL = u

		if parsed, err := url.Parse(u); err == nil && parsed.Host != "" {

			appURL = parsed.Scheme + "://" + parsed.Host

		}

	}

	return appName, appURL

}



func pickURL(candidates ...string) string {

	for _, c := range candidates {

		c = strings.TrimSpace(c)

		if c != "" {

			return c

		}

	}

	return ""

}



func ParseAlipayPaidAmount(totalAmount string) int64 {

	var yuan float64

	fmt.Sscanf(totalAmount, "%f", &yuan)

	return int64(yuan * 100)

}

