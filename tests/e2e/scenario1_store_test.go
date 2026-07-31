//go:build e2e

package e2e

import (
	"fmt"
	"testing"

	"wpay/model"
	"wpay/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// 场景1：连续新增 10+ 门店，各自商户号不串号
func TestScenario1_StoreExpansion_NoMchCrossTalk(t *testing.T) {
	truncateTestData(t)

	const n = 12
	storeIDs := make([]uint64, n)
	mchNos := make([]string, n)

	for i := 0; i < n; i++ {
		name := fmt.Sprintf("扩容测试店-%02d", i+1)
		storeIDs[i] = createStore(t, name)
		mchNos[i] = fmt.Sprintf("MCH_EXP_%02d", i+1)
		addChannel(t, storeIDs[i], model.PayTypeWechat, mchNos[i])
	}

	paySvc := service.NewPaymentService()
	for i := 0; i < n; i++ {
		orderID, qrURL := createPayOrder(t, storeIDs[i], model.PayTypeWechat, int64((i+1)*100))
		assert.Contains(t, qrURL, mchNos[i], "订单 QR 应包含本店商户号 order=%s", orderID)

		// 直接从缓存/DB 读取渠道，确认不串号
		_, ch, err := paySvc.GetStoreChannel(storeIDs[i], model.PayTypeWechat)
		require.NoError(t, err)
		assert.Equal(t, mchNos[i], ch.MchNo)

		// 交叉验证：其他门店商户号不应出现在本店 QR
		for j := 0; j < n; j++ {
			if i == j {
				continue
			}
			assert.NotContains(t, qrURL, mchNos[j], "门店%d QR 不应含门店%d商户号", i, j)
		}
	}

	// 门店列表分页
	r := apiGET(t, "/api/admin/store/list?page=1&page_size=20&status=1", true)
	require.Equal(t, 200, r.Code)
}

// 场景1 补充：接口路径逐项 smoke
func TestScenario1_StoreCRUD_APIPaths(t *testing.T) {
	truncateTestData(t)

	storeID := createStore(t, "CRUD测试店")

	r := apiRequest(t, "PUT", "/api/admin/store/edit", map[string]interface{}{
		"id": storeID, "address": "新地址", "tax_subject": "个体户更新",
	}, true)
	require.Equal(t, 200, r.Code)

	r = apiRequest(t, "PUT", "/api/admin/store/status", map[string]interface{}{
		"id": storeID, "status": 0,
	}, true)
	require.Equal(t, 200, r.Code)

	r = apiGET(t, "/api/admin/store/list?status=0", true)
	require.Equal(t, 200, r.Code)

	r = apiRequest(t, "DELETE", "/api/admin/store/del", map[string]interface{}{"id": storeID}, true)
	require.Equal(t, 200, r.Code)
}
