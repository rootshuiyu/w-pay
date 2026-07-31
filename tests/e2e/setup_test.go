//go:build e2e

package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
	"time"

	"wpay/common"
	"wpay/config"
	"wpay/dao"
	"wpay/model"
	"wpay/router"
	"wpay/service"

	"github.com/gin-gonic/gin"
)

var (
	testServer *httptest.Server
	testToken  string
)

type apiResp struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

func TestMain(m *testing.M) {
	os.Setenv("APP_ENV", "test")
	os.Setenv("WPAY_TEST_MODE", "1")
	gin.SetMode(gin.TestMode)

	cfg, err := config.Load("test")
	if err != nil {
		fmt.Printf("SKIP: load test config: %v\n", err)
		os.Exit(0)
	}
	common.InitLogger("warn")
	if err := common.InitSnowflake(99); err != nil {
		fmt.Printf("FAIL init snowflake: %v\n", err)
		os.Exit(1)
	}
	if err := dao.InitDB(&cfg.Database); err != nil {
		fmt.Printf("SKIP: postgres not available: %v\n", err)
		os.Exit(0)
	}
	if err := dao.InitRedis(&cfg.Redis); err != nil {
		fmt.Printf("SKIP: redis not available: %v\n", err)
		os.Exit(0)
	}
	if err := dao.AutoMigrate(); err != nil {
		fmt.Printf("FAIL migrate: %v\n", err)
		os.Exit(1)
	}
	_ = service.InitDefaultAdmin()
	_ = dao.WarmupAllStoreCache()

	r := router.Setup("test")
	testServer = httptest.NewServer(r)

	token, err := login("admin", "admin123")
	if err != nil {
		fmt.Printf("FAIL login: %v\n", err)
		os.Exit(1)
	}
	testToken = token

	code := m.Run()
	testServer.Close()
	os.Exit(code)
}

func login(username, password string) (string, error) {
	body, _ := json.Marshal(map[string]string{"username": username, "password": password})
	resp, err := http.Post(testServer.URL+"/api/admin/login", "application/json", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var r apiResp
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return "", err
	}
	if r.Code != 200 {
		return "", fmt.Errorf("login code=%d msg=%s", r.Code, r.Message)
	}
	var data struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(r.Data, &data); err != nil {
		return "", err
	}
	return data.Token, nil
}

func apiRequest(t *testing.T, method, path string, payload interface{}, authed bool) apiResp {
	t.Helper()
	var body io.Reader
	if payload != nil {
		b, _ := json.Marshal(payload)
		body = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, testServer.URL+path, body)
	if err != nil {
		t.Fatal(err)
	}
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if authed {
		req.Header.Set("Authorization", "Bearer "+testToken)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var r apiResp
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	return r
}

func apiGET(t *testing.T, path string, authed bool) apiResp {
	return apiRequest(t, http.MethodGet, path, nil, authed)
}

func truncateTestData(t *testing.T) {
	t.Helper()
	models := []interface{}{
		&model.Order{}, &model.PayChannelHistory{}, &model.PayChannel{},
		&model.Store{}, &model.SensitiveLog{},
	}
	for _, m := range models {
		if err := dao.DB.Unscoped().Where("1 = 1").Delete(m).Error; err != nil {
			t.Fatalf("truncate: %v", err)
		}
	}
	_ = dao.WarmupAllStoreCache()
}

func createStore(t *testing.T, name string) uint64 {
	t.Helper()
	r := apiRequest(t, http.MethodPost, "/api/admin/store/add", map[string]string{
		"store_name":  name,
		"tax_subject": "个体户-" + name,
		"address":     "测试地址",
	}, true)
	if r.Code != 200 {
		t.Fatalf("create store %s: code=%d msg=%s", name, r.Code, r.Message)
	}
	var store model.Store
	if err := json.Unmarshal(r.Data, &store); err != nil {
		t.Fatal(err)
	}
	return store.ID
}

func addChannel(t *testing.T, storeID uint64, payType model.PayType, mchNo string) uint64 {
	t.Helper()
	r := apiRequest(t, http.MethodPost, "/api/admin/channel/add", map[string]interface{}{
		"store_id":       storeID,
		"pay_type":       payType,
		"pool_enabled":   1,
		"mch_no":         mchNo,
		"mch_key":    "test_key_" + mchNo,
		"app_id":     "wx_test",
		"serial_no":  "SN001",
		"private_key": "MOCK_PRIVATE_KEY",
		"notify_url": fmt.Sprintf("http://test/api/notify/wx?store_id=%d", storeID),
	}, true)
	if r.Code != 200 {
		t.Fatalf("add channel store=%d mch=%s: code=%d msg=%s", storeID, mchNo, r.Code, r.Message)
	}
	// MaskChannel 将雪花 ID 序列化为字符串，避免 JSON number 精度丢失
	var data struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(r.Data, &data); err != nil {
		t.Fatal(err)
	}
	id, err := strconv.ParseUint(data.ID, 10, 64)
	if err != nil {
		t.Fatalf("parse channel id %q: %v", data.ID, err)
	}
	return id
}

func createPayOrder(t *testing.T, storeID uint64, payType model.PayType, amount int64) (orderID, qrURL string) {
	t.Helper()
	r := apiRequest(t, http.MethodPost, "/api/pay/create", map[string]interface{}{
		"store_id":   storeID,
		"amount":     amount,
		"pay_type":   payType,
		"subject":    "E2E测试",
		"pay_scene":  "native",
	}, false)
	if r.Code != 200 {
		t.Fatalf("pay create store=%d: code=%d msg=%s", storeID, r.Code, r.Message)
	}
	var data struct {
		OrderID   string `json:"order_id"`
		QRCodeURL string `json:"qr_code_url"`
	}
	if err := json.Unmarshal(r.Data, &data); err != nil {
		t.Fatal(err)
	}
	return data.OrderID, data.QRCodeURL
}

func markOrderPaid(t *testing.T, orderNo string, payAmount int64) {
	t.Helper()
	now := time.Now()
	err := dao.NewOrderDAO().UpdateStatusPaid(orderNo, payAmount, "MOCK_TXN_"+orderNo, now, "mock_notify")
	if err != nil {
		t.Fatalf("mark paid: %v", err)
	}
}
