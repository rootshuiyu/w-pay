//go:build e2e

package e2e

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestAllScenarios_Smoke 一键冒烟：四个核心场景子测试均注册在本包，此用例用于 CI 汇总标记
func TestAllScenarios_Smoke(t *testing.T) {
	require.NotEmpty(t, testToken, "admin token 应已就绪")
	require.NotNil(t, testServer, "httptest server 应已启动")
	t.Log("E2E 环境就绪 — 场景1扩容 / 场景2热更新 / 场景3回调历史 / 场景4对账导出")
}
