import{h as c}from"./api-CPwamEKB.js";async function s(e){const t={page:1,page_size:200};e!=null&&(t.status=e);const n=(await c.get("/admin/platform/list",{params:t})).list||[],a={};for(const o of n)a[o.id]=o.platform_name;return{list:n,map:a}}function u(e,t){return t==null||t===""||String(t)==="0"?"未分配":e[t]||e[String(t)]||`平台#${t}`}function i(e,t,r="text/markdown;charset=utf-8"){const n=new Blob([t],{type:r}),a=URL.createObjectURL(n),o=document.createElement("a");o.href=a,o.download=e,o.click(),URL.revokeObjectURL(a)}function y(e,t=window.location.origin){const r=(e==null?void 0:e.platform_name)||"对接方",n=(e==null?void 0:e.app_key)||"pk_your_app_key",a=(e==null?void 0:e.allowed_ips)||"（由我方配置）";return`# ${r} — 收银 API 对接说明

> 生成时间：${new Date().toLocaleString("zh-CN")}

## 基本信息

| 项目 | 值 |
|------|-----|
| 接口基址 | \`${t}\` |
| App Key | \`${n}\` |
| 对接 IP 白名单 | ${a} |

## 鉴权

请求头携带 \`X-App-Key: ${n}\`，来源 IP 须在白名单内。

## 1. 创建支付订单

\`POST ${t}/api/pay/create\`

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
curl -X POST ${t}/api/pay/create \\
  -H "Content-Type: application/json" \\
  -H "X-App-Key: ${n}" \\
  -d '{"amount":100,"pay_type":1,"pay_scene":"h5","return_url":"https://your-site.com/done","subject":"test"}'
\`\`\`

## 2. 查询订单

\`GET ${t}/api/pay/query?order_no={order_id}\`

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
`}export{y as b,i as d,s as l,u as p};
