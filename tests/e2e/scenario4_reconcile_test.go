//go:build e2e

package e2e

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"wpay/dao"
	"wpay/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xuri/excelize/v2"
)

// 场景4：多门店汇总金额与订单明细一致
func TestScenario4_Reconcile_StatMatchesOrders(t *testing.T) {
	truncateTestData(t)

	store1 := createStore(t, "对账店A")
	store2 := createStore(t, "对账店B")
	store3 := createStore(t, "对账店C")

	addChannel(t, store1, model.PayTypeWechat, "MCH_A")
	addChannel(t, store2, model.PayTypeWechat, "MCH_B")
	addChannel(t, store3, model.PayTypeWechat, "MCH_C")

	amounts := []int64{100, 200, 300, 150, 250}
	stores := []uint64{store1, store1, store2, store3, store2}

	var orderNos []string
	var totalPaid int64
	for i, amt := range amounts {
		no, _ := createPayOrder(t, stores[i], model.PayTypeWechat, amt)
		markOrderPaid(t, no, amt)
		orderNos = append(orderNos, no)
		totalPaid += amt
	}
	_ = orderNos

	start := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	end := time.Now().AddDate(0, 0, 1).Format("2006-01-02")
	ids := fmt.Sprintf("%d,%d,%d", store1, store2, store3)

	// 汇总统计
	r := apiGET(t, fmt.Sprintf("/api/admin/stat/summary?store_ids=%s&start_time=%s&end_time=%s&group_by=day", ids, start, end), true)
	require.Equal(t, 200, r.Code)

	var statResp struct {
		Stats []dao.OrderStat `json:"stats"`
	}
	require.NoError(t, json.Unmarshal(r.Data, &statResp))

	var statPaid int64
	var statPaidCount int64
	for _, s := range statResp.Stats {
		statPaid += s.PaidAmount
		statPaidCount += s.PaidCount
	}
	assert.Equal(t, totalPaid, statPaid, "汇总已支付金额应等于明细合计")
	assert.Equal(t, int64(len(amounts)), statPaidCount)

	// 订单列表
	r = apiGET(t, fmt.Sprintf("/api/admin/order/list?store_ids=%s&order_status=1&start_time=%s&end_time=%s&page_size=100", ids, start, end), true)
	require.Equal(t, 200, r.Code)

	var listResp struct {
		List []model.Order `json:"list"`
	}
	require.NoError(t, json.Unmarshal(r.Data, &listResp))
	var listPaid int64
	for _, o := range listResp.List {
		listPaid += o.PayAmount
	}
	assert.Equal(t, totalPaid, listPaid)
}

// 场景4：Excel 导出可正常生成
func TestScenario4_Export_ExcelDownload(t *testing.T) {
	truncateTestData(t)

	storeID := createStore(t, "导出测试店")
	addChannel(t, storeID, model.PayTypeWechat, "MCH_EXP")
	no, _ := createPayOrder(t, storeID, model.PayTypeWechat, 999)
	markOrderPaid(t, no, 999)

	start := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	end := time.Now().AddDate(0, 0, 1).Format("2006-01-02")

	req, _ := http.NewRequest(http.MethodGet,
		fmt.Sprintf("%s/api/admin/export/orders?store_ids=%d&start_time=%s&end_time=%s", testServer.URL, storeID, start, end), nil)
	req.Header.Set("Authorization", "Bearer "+testToken)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, resp.Header.Get("Content-Type"), "spreadsheetml")

	f, err := excelize.OpenReader(resp.Body)
	require.NoError(t, err)
	rows, err := f.GetRows("对账明细")
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(rows), 2, "至少含表头+1行数据")

	// 汇总导出
	req2, _ := http.NewRequest(http.MethodGet,
		fmt.Sprintf("%s/api/admin/export/stat?store_ids=%d&start_time=%s&end_time=%s", testServer.URL, storeID, start, end), nil)
	req2.Header.Set("Authorization", "Bearer "+testToken)
	resp2, err := http.DefaultClient.Do(req2)
	require.NoError(t, err)
	defer resp2.Body.Close()
	assert.Equal(t, http.StatusOK, resp2.StatusCode)
}

// 场景4：订单查询 API 路径
func TestScenario4_OrderList_APIPath(t *testing.T) {
	truncateTestData(t)
	storeID := createStore(t, "订单列表店")
	addChannel(t, storeID, model.PayTypeWechat, "MCH_OL")
	no, _ := createPayOrder(t, storeID, model.PayTypeWechat, 50)

	r := apiGET(t, fmt.Sprintf("/api/admin/order/list?store_ids=%d", storeID), true)
	require.Equal(t, 200, r.Code)

	r = apiGET(t, "/api/pay/query?order_no="+no, false)
	require.Equal(t, 200, r.Code)
}
