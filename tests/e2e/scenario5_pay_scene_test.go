//go:build e2e

package e2e

import (
	"encoding/json"
	"net/http"
	"testing"

	"wpay/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPayCreate_H5Mode(t *testing.T) {
	truncateTestData(t)
	storeID := createStore(t, "H5测试店")
	addChannel(t, storeID, model.PayTypeWechat, "MCH_H5_WX")

	r := apiRequest(t, http.MethodPost, "/api/pay/create", map[string]interface{}{
		"store_id":   storeID,
		"amount":     100,
		"pay_type":   model.PayTypeWechat,
		"pay_scene":  "h5",
		"return_url": "https://cashier.example.com/done",
		"subject":    "H5测试",
	}, false)
	require.Equal(t, 200, r.Code)

	var data struct {
		PayScene string `json:"pay_scene"`
		PayURL   string `json:"pay_url"`
	}
	require.NoError(t, json.Unmarshal(r.Data, &data))
	assert.Equal(t, "h5", data.PayScene)
	assert.Contains(t, data.PayURL, "MCH_H5_WX")
}

func TestPayCreate_H5FallbackNative(t *testing.T) {
	truncateTestData(t)
	storeID := createStore(t, "H5兜底店")
	addChannel(t, storeID, model.PayTypeAlipay, "MCH_ALI_WAP")

	r := apiRequest(t, http.MethodPost, "/api/pay/create", map[string]interface{}{
		"store_id":  storeID,
		"amount":    200,
		"pay_type":  model.PayTypeAlipay,
		"pay_scene": "native",
		"subject":   "扫码",
	}, false)
	require.Equal(t, 200, r.Code)

	var data struct {
		PayScene  string `json:"pay_scene"`
		QRCodeURL string `json:"qr_code_url"`
	}
	require.NoError(t, json.Unmarshal(r.Data, &data))
	assert.Equal(t, "native", data.PayScene)
	assert.Contains(t, data.QRCodeURL, "MCH_ALI_WAP")
}
