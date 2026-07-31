<template>
  <div>
    <p class="page-tip">
      流程：<strong>系统设置录入商户码进公共池</strong> → 新建代收平台并填 IP → <strong>从码池绑定商户码</strong> → 导出对接文档发给对方
    </p>

    <div class="page-toolbar">
      <el-input v-model="keyword" placeholder="搜索平台" clearable style="width: 200px" @keyup.enter="load" />
      <el-button @click="load">刷新</el-button>
      <div class="page-toolbar-spacer" />
      <el-button type="primary" @click="openWizard"><el-icon><Plus /></el-icon>新建平台</el-button>
    </div>

    <el-table :data="rows" v-loading="loading" stripe empty-text=" ">
      <el-table-column prop="platform_name" label="平台名称" min-width="130" />
      <el-table-column label="商户码" width="72" align="center">
        <template #default="{ row }">{{ row.channel_count }}</template>
      </el-table-column>
      <el-table-column label="对接 IP" min-width="160" show-overflow-tooltip>
        <template #default="{ row }">
          <span v-if="row.allowed_ips">{{ row.allowed_ips }}</span>
          <el-tag v-else size="small" type="danger">未配置</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="App Key" min-width="200">
        <template #default="{ row }">
          <span class="mono">{{ row.app_key }}</span>
          <el-button link type="primary" size="small" @click="copy(row.app_key)">复制</el-button>
        </template>
      </el-table-column>
      <el-table-column label="状态" width="80" align="center">
        <template #default="{ row }">
          <el-switch :model-value="row.status === 1" @change="(v) => toggleStatus(row, v)" />
        </template>
      </el-table-column>
      <el-table-column label="操作" width="220" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" @click="openManage(row)">管理</el-button>
          <el-button link @click="exportDoc(row)">导出文档</el-button>
          <el-popconfirm title="删除后商户码将解绑回公共池" @confirm="del(row)">
            <template #reference>
              <el-button link type="danger">删除</el-button>
            </template>
          </el-popconfirm>
        </template>
      </el-table-column>
    </el-table>

    <EmptyState v-if="!loading && rows.length === 0" title="暂无代收平台" description="点击「新建平台」开始接入外部代收方">
      <el-button type="primary" @click="openWizard">新建平台</el-button>
    </EmptyState>

    <!-- 新建平台 -->
    <el-dialog v-model="wizard.visible" title="新建代收平台" width="560px" destroy-on-close @closed="resetWizard">
      <el-form label-width="96px">
        <el-form-item label="平台名称" required>
          <el-input v-model="wizard.name" placeholder="如：XX商城" />
        </el-form-item>
        <el-form-item label="对接 IP" required>
          <el-input v-model="wizard.allowed_ips" type="textarea" :rows="2" placeholder="对方出口 IP，多个用英文逗号分隔，如：203.0.113.10,192.168.1.0/24" />
          <p class="hint">对方 API 请求必须来自这些 IP，否则拒绝</p>
        </el-form-item>
        <el-form-item label="绑定码池">
          <el-select v-model="wizard.channelIds" multiple filterable collapse-tags placeholder="可选：从公共码池勾选商户码" style="width: 100%">
            <el-option v-for="c in unassignedChannels" :key="c.id" :label="channelLabel(c)" :value="c.id" />
          </el-select>
          <p class="hint">也可创建后在「管理」里随时绑定/解绑</p>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="wizard.visible = false">取消</el-button>
        <el-button type="primary" :loading="wizard.saving" @click="submitWizard">创建</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="done.visible" title="创建成功" width="500px">
      <p>平台 <strong>{{ done.name }}</strong> 已创建，请将以下信息发给对方：</p>
      <div class="info-block"><span class="lbl">App Key</span><code>{{ done.appKey }}</code></div>
      <div class="info-block"><span class="lbl">对接 IP</span><code>{{ done.ips }}</code></div>
      <template #footer>
        <el-button type="primary" @click="exportDoc(done.row)">导出对接文档</el-button>
        <el-button @click="copy(done.appKey)">复制 App Key</el-button>
        <el-button @click="done.visible = false">关闭</el-button>
      </template>
    </el-dialog>

    <!-- 平台管理 -->
    <el-drawer v-model="manage.visible" :title="'管理 · ' + manage.name" size="620px">
      <el-form label-width="88px" size="small">
        <el-form-item label="App Key">
          <code class="mono">{{ manage.appKey }}</code>
          <el-button link type="primary" @click="copy(manage.appKey)">复制</el-button>
        </el-form-item>
        <el-form-item label="对接 IP" required>
          <el-input v-model="manage.allowed_ips" type="textarea" :rows="2" />
          <el-button type="primary" size="small" :loading="manage.savingIp" style="margin-top: 8px" @click="saveIp">保存 IP</el-button>
        </el-form-item>
      </el-form>

      <div class="section-head">
        <span>已绑定商户码（{{ manage.channels.length }}）</span>
        <el-button type="primary" size="small" @click="openBind">从码池绑定</el-button>
      </div>
      <el-table :data="manage.channels" v-loading="manage.loading" size="small" stripe>
        <el-table-column prop="mch_no" label="商户号" min-width="120" />
        <el-table-column label="支付" width="72">
          <template #default="{ row }">{{ payTypeText(row.pay_type) }}</template>
        </el-table-column>
        <el-table-column label="今日/限额" width="130">
          <template #default="{ row }">
            ￥{{ fen2yuan(row.daily_used_fen) }}/{{ row.daily_limit_fen > 0 ? fen2yuan(row.daily_limit_fen) : '∞' }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="72" fixed="right">
          <template #default="{ row }">
            <el-popconfirm title="解绑后该码回到公共码池" @confirm="unbind(row)">
              <template #reference>
                <el-button link type="danger" size="small">解绑</el-button>
              </template>
            </el-popconfirm>
          </template>
        </el-table-column>
      </el-table>

      <div class="drawer-footer">
        <el-button @click="exportDoc(manage.row)">导出对接文档</el-button>
        <el-button link type="primary" @click="$router.push('/settings?tab=pool')">到码池入库补填密钥 →</el-button>
      </div>
    </el-drawer>

    <!-- 从码池绑定 -->
    <el-dialog v-model="bind.visible" title="从公共码池绑定商户码" width="520px" append-to-body>
      <el-empty v-if="!bind.options.length" description="公共码池暂无可绑定的商户码，请先到「系统设置 → 码池入库」录入" />
      <el-checkbox-group v-else v-model="bind.selected">
        <div v-for="c in bind.options" :key="c.id" class="bind-line">
          <el-checkbox :value="c.id">{{ channelLabel(c) }}</el-checkbox>
        </div>
      </el-checkbox-group>
      <template #footer>
        <el-button @click="bind.visible = false">取消</el-button>
        <el-button type="primary" :loading="bind.saving" :disabled="!bind.selected.length" @click="submitBind">绑定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import http, { fen2yuan, payTypeText } from '../api'
import EmptyState from './EmptyState.vue'
import { buildPayApiDoc, downloadText } from '../utils/platforms.js'

const rows = ref([])
const loading = ref(false)
const keyword = ref('')
const unassignedChannels = ref([])
const wizard = reactive({ visible: false, saving: false, name: '', allowed_ips: '', channelIds: [] })
const done = reactive({ visible: false, name: '', appKey: '', ips: '', row: null })
const manage = reactive({
  visible: false, loading: false, savingIp: false, id: '', name: '', appKey: '',
  allowed_ips: '', channels: [], row: null,
})
const bind = reactive({ visible: false, saving: false, options: [], selected: [] })

function channelLabel(c) {
  const pt = payTypeText(c.pay_type)
  const lim = c.daily_limit_fen > 0 ? `日限￥${fen2yuan(c.daily_limit_fen)}` : '不限额'
  return `${pt} · ${c.mch_no} · ${lim}`
}

async function loadUnassigned() {
  const data = await http.get('/admin/platform/available-channels', { params: { unassigned: 1 } })
  unassignedChannels.value = data.list || []
}

async function load() {
  loading.value = true
  try {
    const params = { page: 1, page_size: 100 }
    if (keyword.value) params.keyword = keyword.value
    const data = await http.get('/admin/platform/list', { params })
    rows.value = data.list || []
  } finally {
    loading.value = false
  }
}

function openWizard() {
  wizard.name = ''
  wizard.allowed_ips = ''
  wizard.channelIds = []
  loadUnassigned()
  wizard.visible = true
}
function resetWizard() {
  wizard.name = ''
  wizard.allowed_ips = ''
  wizard.channelIds = []
}

async function submitWizard() {
  const name = wizard.name.trim()
  const ips = wizard.allowed_ips.trim()
  if (!name) return ElMessage.warning('请填写平台名称')
  if (!ips) return ElMessage.warning('请填写对接 IP 白名单')
  wizard.saving = true
  try {
    const p = await http.post('/admin/platform/add', { platform_name: name, allowed_ips: ips })
    if (wizard.channelIds.length) {
      await http.put('/admin/platform/bind-channels', {
        platform_id: p.id, channel_ids: wizard.channelIds,
      })
    }
    wizard.visible = false
    done.name = name
    done.appKey = p.app_key
    done.ips = ips
    done.row = { ...p, platform_name: name, allowed_ips: ips }
    done.visible = true
    load()
    loadUnassigned()
  } finally {
    wizard.saving = false
  }
}

async function openManage(row) {
  manage.id = row.id
  manage.name = row.platform_name
  manage.appKey = row.app_key
  manage.allowed_ips = row.allowed_ips || ''
  manage.row = row
  manage.visible = true
  manage.loading = true
  try {
    const data = await http.get('/admin/platform/detail', { params: { platform_id: row.id } })
    manage.channels = data.channels || []
    manage.allowed_ips = data.allowed_ips || ''
  } finally {
    manage.loading = false
  }
}

async function saveIp() {
  if (!manage.allowed_ips.trim()) return ElMessage.warning('请填写对接 IP')
  manage.savingIp = true
  try {
    await http.put('/admin/platform/edit', {
      id: manage.id, platform_name: manage.name, allowed_ips: manage.allowed_ips.trim(),
    })
    ElMessage.success('IP 白名单已保存')
    load()
  } finally {
    manage.savingIp = false
  }
}

async function openBind() {
  bind.selected = []
  const data = await http.get('/admin/platform/available-channels', { params: { unassigned: 1 } })
  bind.options = data.list || []
  bind.visible = true
}

async function submitBind() {
  bind.saving = true
  try {
    await http.put('/admin/platform/bind-channels', {
      platform_id: manage.id, channel_ids: bind.selected,
    })
    ElMessage.success('绑定成功')
    bind.visible = false
    openManage({ id: manage.id, platform_name: manage.name, app_key: manage.appKey })
    load()
    loadUnassigned()
  } finally {
    bind.saving = false
  }
}

async function unbind(row) {
  await http.put('/admin/platform/unbind-channel', { platform_id: manage.id, channel_id: row.id })
  ElMessage.success('已解绑')
  openManage({ id: manage.id, platform_name: manage.name, app_key: manage.appKey })
  load()
  loadUnassigned()
}

async function toggleStatus(row, on) {
  try {
    await http.put('/admin/platform/status', { id: row.id, status: on ? 1 : 0 })
    row.status = on ? 1 : 0
  } catch {
    /* 后端会提示未配 IP */
  }
}

async function del(row) {
  await http.delete('/admin/platform/del', { data: { id: row.id } })
  ElMessage.success('已删除')
  load()
  loadUnassigned()
}

function exportDoc(row) {
  if (!row) return
  const md = buildPayApiDoc(row)
  const safe = (row.platform_name || 'platform').replace(/[^\w\u4e00-\u9fa5-]+/g, '_')
  downloadText(`${safe}-收银API对接说明.md`, md)
  ElMessage.success('文档已下载')
}

function copy(text) {
  navigator.clipboard?.writeText(String(text))
  ElMessage.success('已复制')
}

load()
loadUnassigned()
</script>

<style scoped>
.page-tip { margin: 0 0 16px; font-size: 13px; color: var(--text-secondary); line-height: 1.6; }
.mono { font-family: Consolas, monospace; font-size: 12px; }
.hint { margin: 4px 0 0; font-size: 12px; color: var(--text-muted); }
.info-block { margin: 10px 0; padding: 10px; background: #f5f7fa; border-radius: 6px; }
.info-block .lbl { display: block; font-size: 12px; color: var(--text-muted); margin-bottom: 4px; }
.info-block code { word-break: break-all; font-size: 13px; }
.section-head { display: flex; justify-content: space-between; align-items: center; margin: 16px 0 8px; font-weight: 600; font-size: 14px; }
.bind-line { padding: 6px 0; border-bottom: 1px solid #f0f0f0; }
.drawer-footer { margin-top: 20px; display: flex; gap: 12px; align-items: center; }
</style>
