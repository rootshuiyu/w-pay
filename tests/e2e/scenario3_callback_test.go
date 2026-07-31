//go:build e2e

package e2e

import (
	"testing"
	"time"

	"wpay/dao"
	"wpay/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// 场景3：更换商户后，旧密钥归档到 history，7 天内可供回调验签
func TestScenario3_CallbackHistory_OnChannelEdit(t *testing.T) {
	truncateTestData(t)

	storeID := createStore(t, "回调兼容店")
	chID := addChannel(t, storeID, model.PayTypeWechat, "MCH_CB_OLD")

	orderNo, _ := createPayOrder(t, storeID, model.PayTypeWechat, 500)

	order, err := dao.NewOrderDAO().FindByOrderNo(orderNo)
	require.NoError(t, err)
	require.NotNil(t, order)
	assert.Equal(t, chID, order.ChannelID)

	// 更换商户密钥
	r := apiRequest(t, "PUT", "/api/admin/channel/edit", map[string]interface{}{
		"id": chID, "mch_no": "MCH_CB_NEW", "mch_key": "key_new", "private_key": "PRIV_NEW",
	}, true)
	require.Equal(t, 200, r.Code)

	histDAO := dao.NewPayChannelHistoryDAO()
	histories, err := histDAO.ListValidByStoreAndType(storeID, model.PayTypeWechat)
	require.NoError(t, err)
	require.Len(t, histories, 1)
	assert.Equal(t, "MCH_CB_OLD", histories[0].MchNo)
	assert.Equal(t, chID, histories[0].ChannelID)
	assert.True(t, histories[0].ExpiresAt.After(time.Now()))

	byChannel, err := histDAO.ListValidByChannelID(chID)
	require.NoError(t, err)
	require.Len(t, byChannel, 1)
	assert.Equal(t, "MCH_CB_OLD", byChannel[0].MchNo)

	// 当前渠道已是新商户
	ch, err := dao.NewPayChannelDAO().FindByID(chID)
	require.NoError(t, err)
	assert.Equal(t, "MCH_CB_NEW", ch.MchNo)
}

// 场景3：仅改 status 不应产生 history
func TestScenario3_NoHistory_WhenStatusOnlyChange(t *testing.T) {
	truncateTestData(t)

	storeID := createStore(t, "状态变更店")
	chID := addChannel(t, storeID, model.PayTypeAlipay, "MCH_ALI_OLD")

	r := apiRequest(t, "PUT", "/api/admin/channel/status", map[string]interface{}{
		"id": chID, "status": 0,
	}, true)
	require.Equal(t, 200, r.Code)

	histories, err := dao.NewPayChannelHistoryDAO().ListValidByChannelID(chID)
	require.NoError(t, err)
	assert.Len(t, histories, 0)
}

// 场景3：统一返回结构
func TestScenario3_UnifiedResponseFormat(t *testing.T) {
	r := apiGET(t, "/health", false)
	// health 非标准包装，测 admin login
	r = apiRequest(t, "POST", "/api/admin/login", map[string]string{
		"username": "admin", "password": "admin123",
	}, false)
	require.Equal(t, 200, r.Code)
	assert.Equal(t, "success", r.Message)
}
