package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestParseIPList(t *testing.T) {
	cases := []struct {
		raw  string
		want int
	}{
		{"", 0},
		{"10.0.0.1", 1},
		{"10.0.0.1,192.168.1.0/24", 2},
		{" 10.0.0.1 ; 172.16.0.0/12 ,, ", 2},
	}
	for _, c := range cases {
		if got := len(ParseIPList(c.raw)); got != c.want {
			t.Errorf("ParseIPList(%q) = %d entries, want %d", c.raw, got, c.want)
		}
	}
}

func TestIPWhitelist(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cases := []struct {
		name       string
		entries    []string
		allowEmpty bool
		remoteIP   string
		want       int
	}{
		{"空白名单allowEmpty=true放行", nil, true, "203.0.113.9", http.StatusOK},
		{"空白名单allowEmpty=false拒绝", nil, false, "203.0.113.9", http.StatusForbidden},
		{"精确 IP 命中", []string{"10.1.2.3"}, true, "10.1.2.3", http.StatusOK},
		{"精确 IP 未命中", []string{"10.1.2.3"}, true, "10.1.2.4", http.StatusForbidden},
		{"CIDR 命中", []string{"192.168.0.0/16"}, true, "192.168.7.20", http.StatusOK},
		{"CIDR 未命中", []string{"192.168.0.0/16"}, true, "172.20.0.5", http.StatusForbidden},
		{"条目全非法则 fail-closed", []string{"not-an-ip"}, true, "10.0.0.1", http.StatusForbidden},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r := gin.New()
			r.Use(IPWhitelist("test", c.entries, c.allowEmpty))
			r.GET("/ping", func(ctx *gin.Context) { ctx.String(http.StatusOK, "pong") })

			req := httptest.NewRequest(http.MethodGet, "/ping", nil)
			req.RemoteAddr = c.remoteIP + ":12345"
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != c.want {
				t.Errorf("status = %d, want %d", w.Code, c.want)
			}
		})
	}
}
