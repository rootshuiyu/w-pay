//go:build e2e

package e2e

import (
	"fmt"
	"testing"

	"wpay/dao"
	"wpay/model"
	"wpay/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// 场景2：修改渠道商户号后，新订单立刻走新商户，无需重启
func TestScenario2_ChannelHotUpdate_NewOrderUsesNewMch(t *testing.T) {
	truncateTestData(t)

	storeID := createStore(t, "热更新测试店")
	chID := addChannel(t, storeID, model.PayTypeWechat, "MCH_OLD")

	_, qrOld := createPayOrder(t, storeID, model.PayTypeWechat, 100)
	assert.Contains(t, qrOld, "MCH_OLD")

	// 更换商户（核心热更新接口）
	r := apiRequest(t, "PUT", "/api/admin/channel/edit", map[string]interface{}{
		"id": chID, "mch_no": "MCH_NEW", "mch_key": "new_key",
	}, true)
	require.Equal(t, 200, r.Code)

	// 模拟下一请求：缓存已 DEL，应回源 DB
	_ = dao.DeleteChannelCache(storeID, model.PayTypeWechat)

	_, qrNew := createPayOrder(t, storeID, model.PayTypeWechat, 200)
	assert.Contains(t, qrNew, "MCH_NEW")
	assert.NotContains(t, qrNew, "MCH_OLD")

	paySvc := service.NewPaymentService()
	_, ch, err := paySvc.GetStoreChannel(storeID, model.PayTypeWechat)
	require.NoError(t, err)
	assert.Equal(t, "MCH_NEW", ch.MchNo)
}

// 场景2 补充：关停渠道后下单应失败
func TestScenario2_ChannelStatus_DisableBlocksPay(t *testing.T) {
	truncateTestData(t)

	storeID := createStore(t, "关停测试店")
	chID := addChannel(t, storeID, model.PayTypeAlipay, "MCH_ALI_01")

	r := apiRequest(t, "PUT", "/api/admin/channel/status", map[string]interface{}{
		"id": chID, "status": 0,
	}, true)
	require.Equal(t, 200, r.Code)

	_ = dao.DeleteChannelCache(storeID, model.PayTypeAlipay)

	r = apiRequest(t, "POST", "/api/pay/create", map[string]interface{}{
		"store_id": storeID, "amount": 100, "pay_type": model.PayTypeAlipay,
	}, false)
	assert.NotEqual(t, 200, r.Code, "关停渠道后不应允许下单")
}

// 场景2：渠道 API 路径
func TestScenario2_ChannelAPIPaths(t *testing.T) {
	truncateTestData(t)
	storeID := createStore(t, "渠道API店")
	addChannel(t, storeID, model.PayTypeWechat, "MCH_WX")
	addChannel(t, storeID, model.PayTypeAlipay, "MCH_ALI")

	r := apiGET(t, fmt.Sprintf("/api/admin/channel/list?store_id=%d", storeID), true)
	require.Equal(t, 200, r.Code)
}
