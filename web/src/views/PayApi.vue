<template>
  <div class="api-page">
    <el-card shadow="never" class="page-card" style="margin-bottom: 16px">
      <div class="doc-toolbar">
        <el-select v-model="selectedPlatformId" filterable placeholder="选择代收平台（导出专属文档）" style="width: 280px" @change="onPlatformChange">
          <el-option v-for="p in platforms" :key="p.id" :label="p.platform_name" :value="p.id" />
        </el-select>
        <el-button type="primary" :disabled="!currentPlatform" @click="exportDoc">
          <el-icon><Download /></el-icon>导出对接文档 (.md)
        </el-button>
        <el-button v-if="currentPlatform" @click="copy(currentPlatform.app_key)">复制 App Key</el-button>
      </div>
      <p v-if="currentPlatform" class="doc-meta">
        当前：<strong>{{ currentPlatform.platform_name }}</strong>
        · App Key <code>{{ currentPlatform.app_key }}</code>
        · IP <code>{{ currentPlatform.allowed_ips || '未配置' }}</code>
      </p>
    </el-card>

    <el-alert
      type="success"
      :closable="false"
      show-icon
      title="标准聚合支付接口：创建订单 → 跳转 pay_url 或展示 qr_code_url → 查单确认。须携带 X-App-Key，且来源 IP 须在该平台对接白名单内。手机浏览器推荐 pay_scene=h5；PC/收款码屏用 native。"
      style="margin-bottom: 16px"
    />

    <el-row :gutter="16">
      <el-col :xs="24" :lg="14">
        <el-card shadow="never" class="page-card">
          <template #header><span class="endpoint-title">POST /api/pay/create</span> — 创建支付订单</template>
          <p class="desc">
            传入金额与支付方式即可出单。系统返回 <code>pay_url</code>（H5/WAP 跳转）或 <code>qr_code_url</code>（扫码），对接方按 <code>pay_scene</code> 处理即可。
          </p>
          <div class="label">请求体 JSON</div>
          <pre class="code">{{ createReq }}</pre>
          <div class="label">成功响应（外层统一格式）</div>
          <pre class="code">{{ createResp }}</pre>
          <div class="label">失败响应示例</div>
          <pre class="code">{{ errResp }}</pre>
          <div class="label">curl</div>
          <pre class="code">{{ createCurl }}</pre>
          <el-button size="small" @click="copy(createCurl)">复制 curl</el-button>
        </el-card>

        <el-card shadow="never" class="page-card" style="margin-top: 16px">
          <template #header><span class="endpoint-title">GET /api/pay/query</span> — 查询订单</template>
          <p class="desc">参数 <code>order_no</code> 或 <code>order_id</code>（二选一）。支付完成后建议定时查单，直至 <code>order_status=1</code>。</p>
          <pre class="code">{{ queryExample }}</pre>
          <div class="label">成功响应 data</div>
          <pre class="code">{{ queryResp }}</pre>
          <el-button size="small" @click="copy(queryExample)">复制 curl</el-button>
        </el-card>

        <el-card shadow="never" class="page-card" style="margin-top: 16px">
          <template #header>收银端接入示例（JavaScript）</template>
          <pre class="code">{{ jsExample }}</pre>
          <el-button size="small" @click="copy(jsExample)">复制代码</el-button>
        </el-card>
      </el-col>

      <el-col :xs="24" :lg="10">
        <el-card shadow="never" class="page-card">
          <template #header>create 请求字段</template>
          <el-table :data="reqFields" size="small" stripe>
            <el-table-column prop="name" label="字段" width="110" />
            <el-table-column prop="req" label="必填" width="52" />
            <el-table-column prop="desc" label="说明" />
          </el-table>
        </el-card>

        <el-card shadow="never" class="page-card" style="margin-top: 16px">
          <template #header>create / query 响应字段</template>
          <el-table :data="respFields" size="small" stripe>
            <el-table-column prop="name" label="字段" width="120" />
            <el-table-column prop="desc" label="说明" />
          </el-table>
        </el-card>

        <el-card shadow="never" class="page-card" style="margin-top: 16px">
          <template #header>流程说明</template>
          <ul class="flow">
            <li>POST <code>/api/pay/create</code> → 返回支付链接或二维码</li>
            <li><code>pay_scene=h5</code> → 使用 <code>pay_url</code> 跳转唤起支付</li>
            <li><code>pay_scene=native</code> → 使用 <code>qr_code_url</code> 展示二维码</li>
            <li>用户完成支付 → 平台异步通知 → 订单状态更新为已支付</li>
            <li>GET <code>/api/pay/query</code> 查单确认（勿仅依赖 return_url）</li>
          </ul>
        </el-card>

        <el-card shadow="never" class="page-card" style="margin-top: 16px">
          <template #header>对接提示</template>
          <ul class="flow">
            <li>金额单位均为<strong>分</strong>（100 = 1.00 元）</li>
            <li>兼容别名：<code>channel_type</code>=wechat/alipay；<code>biz_remark</code>=subject</li>
            <li>每个代收平台在「平台管理」配置<strong>专属 IP 白名单</strong>，请求须携带 <code>X-App-Key</code></li>
            <li>无需传 store_id、商户号，系统自动从该平台码池轮询收款</li>
          </ul>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { buildPayApiDoc, downloadText, loadPlatformOptions } from '../utils/platforms.js'

const base = window.location.origin
const platforms = ref([])
const selectedPlatformId = ref('')

const currentPlatform = computed(() =>
  platforms.value.find((p) => String(p.id) === String(selectedPlatformId.value))
)

function onPlatformChange() {
  if (currentPlatform.value) {
    const key = currentPlatform.value.app_key
    createCurl.value = buildCurl(key)
    queryExample.value = buildQueryCurl(key)
    jsExample.value = buildJsExample(key)
  }
}

function buildCurl(appKey) {
  return `curl -X POST ${base}/api/pay/create \\
  -H "Content-Type: application/json" \\
  -H "X-App-Key: ${appKey || 'pk_your_app_key'}" \\
  -d '{"amount":100,"pay_type":1,"pay_scene":"h5","return_url":"https://cashier.example.com/done","subject":"test"}'`
}

function buildQueryCurl(appKey) {
  return `curl "${base}/api/pay/query?order_no=ORDER_ID" \\
  -H "X-App-Key: ${appKey || 'pk_your_app_key'}"
# 或 order_id=ORDER_ID`
}

function buildJsExample(appKey) {
  const key = appKey || 'pk_your_app_key'
  return `const base = '${base}';
const appKey = '${key}';
const r = await fetch(base + '/api/pay/create', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    'X-App-Key': appKey,
  },
  body: JSON.stringify({
    amount: 100,           // 分
    pay_type: 1,           // 1微信 2支付宝
    pay_scene: 'h5',
    return_url: location.href,
    subject: '收款',
  }),
}).then(res => res.json());
if (r.code !== 200) throw new Error(r.message);
const d = r.data;
if (d.pay_url) location.href = d.pay_url;
else if (d.qr_code_url) showQRCode(d.qr_code_url);

// 查单确认（同样携带 X-App-Key）
const timer = setInterval(async () => {
  const q = await fetch(base + '/api/pay/query?order_no=' + d.order_id, {
    headers: { 'X-App-Key': appKey },
  }).then(x => x.json());
  if (q.data?.order_status === 1) { clearInterval(timer); onPaid(); }
}, 2000);`
}

const createCurl = ref(buildCurl('pk_your_app_key'))
const queryExample = ref(buildQueryCurl('pk_your_app_key'))
const jsExample = ref(buildJsExample('pk_your_app_key'))

const createReq = `{
  "amount": 100,
  "pay_type": 1,
  "pay_scene": "h5",
  "return_url": "https://cashier.example.com/done",
  "subject": "堂食消费"
}
// Header: X-App-Key: pk_xxxxxxxx（由我方分配）`

const createResp = `{
  "code": 200,
  "message": "success",
  "data": {
    "order_id": "2082511055136755712",
    "pay_scene": "h5",
    "pay_url": "https://wx.tenpay.com/...",
    "qr_code_url": "",
    "amount": 100,
    "pay_type": 1
  }
}`

const errResp = `{
  "code": 400,
  "message": "支付通道暂不可用，请稍后重试"
}`

const queryResp = `{
  "code": 200,
  "message": "success",
  "data": {
    "order_id": "2082511055136755712",
    "order_status": 0,
    "total_amount": 100,
    "pay_amount": 0,
    "pay_type": 1,
    "pay_scene": "h5",
    "qr_code_url": "...",
    "pay_time": null,
    "transaction_id": ""
  }
}`

const reqFields = [
  { name: 'amount', req: '是', desc: '金额（分），须 > 0' },
  { name: 'pay_type', req: '是', desc: '1=微信，2=支付宝' },
  { name: 'pay_scene', req: '否', desc: 'h5（默认）| native；H5 失败自动降级 native' },
  { name: 'return_url', req: '建议', desc: 'H5/WAP 支付完成回跳' },
  { name: 'quit_url', req: '否', desc: '支付宝取消支付回跳' },
  { name: 'subject', req: '否', desc: '备注；也可用 biz_remark' },
  { name: 'device_sn', req: '否', desc: '收银设备号' },
  { name: 'store_id', req: '否', desc: '可选业务门店标识' },
  { name: 'channel_type', req: '否', desc: '兼容：wechat / alipay' },
]

const respFields = [
  { name: 'order_id', desc: '订单号（create/query 通用）' },
  { name: 'order_status', desc: 'query：0待支付 1已支付 2已关闭 3退款' },
  { name: 'total_amount', desc: 'query：下单金额（分）' },
  { name: 'pay_amount', desc: 'query：实付金额（分），支付后更新' },
  { name: 'pay_scene', desc: '实际支付方式 h5|native' },
  { name: 'pay_url', desc: 'create：H5/WAP 跳转链接' },
  { name: 'qr_code_url', desc: 'create/query：扫码链接（native 或兜底）' },
  { name: 'amount', desc: 'create：下单金额（分）' },
  { name: 'pay_type', desc: '1=微信 2=支付宝' },
  { name: 'pay_time', desc: 'query：支付完成时间' },
  { name: 'transaction_id', desc: 'query：第三方支付流水号' },
]

function copy(text) {
  const v = typeof text === 'string' ? text : text?.value ?? String(text)
  navigator.clipboard?.writeText(v)
  ElMessage.success('已复制')
}

function exportDoc() {
  if (!currentPlatform.value) return
  const md = buildPayApiDoc(currentPlatform.value, base)
  const safe = (currentPlatform.value.platform_name || 'platform').replace(/[^\w\u4e00-\u9fa5-]+/g, '_')
  downloadText(`${safe}-收银API对接说明.md`, md)
  ElMessage.success('文档已下载')
}

onMounted(async () => {
  const { list } = await loadPlatformOptions(1)
  platforms.value = list
  if (list.length) {
    selectedPlatformId.value = list[0].id
    onPlatformChange()
  }
})
</script>

<style scoped>
.doc-toolbar { display: flex; flex-wrap: wrap; gap: 12px; align-items: center; }
.doc-meta { margin: 12px 0 0; font-size: 13px; color: var(--text-secondary); }
.doc-meta code { background: #f0f2f5; padding: 2px 6px; border-radius: 4px; font-size: 12px; }
.api-page :deep(.el-card) { border: 1px solid var(--card-border); }
.endpoint-title { font-weight: 700; font-family: Consolas, monospace; font-size: 14px; }
.desc { color: var(--text-secondary); font-size: 14px; margin-bottom: 12px; line-height: 1.6; }
.label { font-size: 13px; color: var(--text-muted); margin: 12px 0 6px; font-weight: 600; }
.code {
  background: #1e1e1e;
  color: #d4d4d4;
  padding: 12px;
  border-radius: 6px;
  font-size: 12px;
  line-height: 1.5;
  overflow-x: auto;
  white-space: pre-wrap;
  word-break: break-all;
}
.flow { padding-left: 20px; line-height: 2; color: #555; font-size: 14px; }
.flow :deep(a) { color: var(--brand-primary-light); text-decoration: none; }
.flow :deep(a:hover) { text-decoration: underline; }
.flow code { background: #f0f2f5; padding: 2px 6px; border-radius: 4px; }
</style>
