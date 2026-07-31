import http from '../api'

export async function loadPlatformOptions(status) {
  const params = { page: 1, page_size: 200 }
  if (status != null) params.status = status
  const data = await http.get('/admin/platform/list', { params })
  const list = data.list || []
  const map = {}
  for (const p of list) map[p.id] = p.platform_name
  return { list, map }
}

export function platformName(map, id) {
  if (id == null || id === '' || String(id) === '0') return '未分配'
  return map[id] || map[String(id)] || `平台#${id}`
}

export function downloadText(filename, content, mime = 'text/markdown;charset=utf-8') {
  const blob = new Blob([content], { type: mime })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = filename
  a.click()
  URL.revokeObjectURL(url)
}

export function buildPayApiDoc(platform, base = window.location.origin) {
  const name = platform?.platform_name || '对接方'
  const appKey = platform?.app_key || 'pk_your_app_key'
  const ips = platform?.allowed_ips || '（由我方配置）'
  return `# ${name} — 收银 API 对接说明

> 生成时间：${new Date().toLocaleString('zh-CN')}

## 基本信息

| 项目 | 值 |
|------|-----|
| 接口基址 | \`${base}\` |
| App Key | \`${appKey}\` |
| 对接 IP 白名单 | ${ips} |

## 鉴权

请求头携带 \`X-App-Key: ${appKey}\`，来源 IP 须在白名单内。

## 1. 创建支付订单

\`POST ${base}/api/pay/create\`

\`\`\`json
{
  "amount": 100,
  "pay_type": 1,
  "pay_scene": "h5",
  "return_url": "https://your-site.com/done",
  "subject": "商品名称"
}
\`

| 字段 | 必填 | 说明 |
|------|------|------|
| amount | 是 | 金额（分），100 = 1.00 元 |
| pay_type | 是 | 1=微信，2=支付宝 |
| pay_scene | 否 | h5（默认）或 native（扫码） |
| return_url | 建议 | H5 支付完成回跳地址 |
| subject | 否 | 订单备注 |

成功响应 \`data\` 含 \`order_id\`、\`pay_url\`（H5）或 \`qr_code_url\`（扫码）。

\`\`\`bash
curl -X POST ${base}/api/pay/create \\
  -H "Content-Type: application/json" \\
  -H "X-App-Key: ${appKey}" \\
  -d '{"amount":100,"pay_type":1,"pay_scene":"h5","return_url":"https://your-site.com/done","subject":"test"}'
\`\`\`

## 2. 查询订单

\`GET ${base}/api/pay/query?order_no={order_id}\`

\`order_status\`：0 待支付 · 1 已支付 · 2 已关闭 · 3 退款

建议支付后轮询查单，直至 \`order_status=1\`，勿仅依赖 return_url。

## 3. 接入流程

1. POST /api/pay/create 获取 pay_url 或 qr_code_url
2. 引导用户完成支付
3. GET /api/pay/query 确认订单状态
4. 我方异步接收支付平台回调并更新订单

## 注意事项

- 金额单位均为 **分**
- 无需传 store_id、商户号等内部字段，系统自动轮询收款
- 如有问题请联系我方运营
`
}
