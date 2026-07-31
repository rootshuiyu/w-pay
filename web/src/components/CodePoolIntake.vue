<template>
  <div>
    <p class="page-tip">
      在此录入从<strong>微信 / 支付宝官方</strong>申请下来的商户号，统一进入公共码池。无需选择门店，录入后可在「代收平台」分配给外部平台轮询收款。
    </p>

    <el-row :gutter="16" class="kpi-grid">
      <el-col :xs="8" :sm="8">
        <el-card shadow="hover" class="kpi-card page-card">
          <div class="kpi-num">{{ kpi.total }}</div>
          <div class="kpi-label">池中商户码</div>
        </el-card>
      </el-col>
      <el-col :xs="8" :sm="8">
        <el-card shadow="hover" class="kpi-card page-card">
          <div class="kpi-num warn">{{ kpi.unassigned }}</div>
          <div class="kpi-label">待分配</div>
        </el-card>
      </el-col>
      <el-col :xs="8" :sm="8">
        <el-card shadow="hover" class="kpi-card page-card">
          <div class="kpi-num success">{{ kpi.assigned }}</div>
          <div class="kpi-label">已绑定平台</div>
        </el-card>
      </el-col>
    </el-row>

    <div class="page-toolbar">
      <el-input v-model="keyword" placeholder="搜索商户号" clearable style="width: 160px" @keyup.enter="load" />
      <el-select v-model="payType" placeholder="支付方式" clearable style="width: 120px" @change="load">
        <el-option label="微信" :value="1" />
        <el-option label="支付宝" :value="2" />
      </el-select>
      <el-select v-model="assignFilter" placeholder="分配状态" style="width: 120px">
        <el-option label="全部" value="all" />
        <el-option label="待分配" value="free" />
        <el-option label="已绑定" value="bound" />
      </el-select>
      <el-button @click="load"><el-icon><Refresh /></el-icon>刷新</el-button>
      <div class="page-toolbar-spacer" />
      <el-button type="primary" @click="openAdd"><el-icon><Plus /></el-icon>录入商户码</el-button>
    </div>

    <el-table :data="filteredRows" v-loading="loading" stripe>
      <el-table-column label="支付" width="88">
        <template #default="{ row }">
          <el-tag size="small" :type="row.pay_type === 1 ? 'success' : 'primary'">{{ payTypeText(row.pay_type) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="mch_no" label="商户号" min-width="140" />
      <el-table-column label="所属平台" min-width="120">
        <template #default="{ row }">
          <el-tag v-if="isFree(row)" size="small" type="warning">待分配</el-tag>
          <span v-else>{{ platformName(platformMap, row.platform_id) }}</span>
        </template>
      </el-table-column>
      <el-table-column label="日额度" min-width="170">
        <template #default="{ row }">
          <el-progress
            v-if="row.daily_limit_fen > 0"
            :percentage="quotaPercent(row.daily_used_fen, row.daily_limit_fen)"
            :status="quotaStatus(row.daily_used_fen, row.daily_limit_fen)"
            :stroke-width="8"
          />
          <span class="hint-text">￥{{ fen2yuan(row.daily_used_fen) }} / {{ row.daily_limit_fen > 0 ? `￥${fen2yuan(row.daily_limit_fen)}` : '不限' }}</span>
        </template>
      </el-table-column>
      <el-table-column label="单笔上限" width="96" align="right">
        <template #default="{ row }">{{ row.single_limit_fen > 0 ? `￥${fen2yuan(row.single_limit_fen)}` : '不限' }}</template>
      </el-table-column>
      <el-table-column label="状态" width="80" align="center">
        <template #default="{ row }">
          <el-switch :model-value="row.status === 1" @change="(v) => toggleStatus(row, v)" />
        </template>
      </el-table-column>
      <el-table-column label="操作" width="180" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" @click="openEdit(row)">编辑</el-button>
          <el-button v-if="isFree(row)" link @click="openAssign(row)">分配</el-button>
        </template>
      </el-table-column>
    </el-table>

    <EmptyState v-if="!loading && filteredRows.length === 0" title="码池暂无商户码" description="点击「码池入库」添加微信/支付宝官方申请的收款商户号">
      <el-button type="primary" @click="openAdd">录入商户码</el-button>
    </EmptyState>

    <!-- 录入 / 编辑 -->
    <el-dialog v-model="dialog.visible" :title="dialog.isEdit ? '编辑码池商户码' : '码池入库'" width="600px" destroy-on-close>
      <el-form :model="form" label-width="100px">
        <el-tabs v-model="formTab">
          <el-tab-pane label="基础" name="basic">
            <el-form-item v-if="!dialog.isEdit" label="支付方式" required>
              <el-radio-group v-model="form.pay_type">
                <el-radio :value="1">微信</el-radio>
                <el-radio :value="2">支付宝</el-radio>
              </el-radio-group>
            </el-form-item>
            <el-form-item label="商户号" required>
              <el-input v-model="form.mch_no" :disabled="dialog.isEdit" placeholder="微信/支付宝商户号" />
            </el-form-item>
            <el-form-item label="AppID">
              <el-input v-model="form.app_id" placeholder="可选" />
            </el-form-item>
            <el-form-item label="备注">
              <el-input v-model="form.remark" />
            </el-form-item>
          </el-tab-pane>
          <el-tab-pane label="额度" name="quota">
            <el-form-item label="参与轮询">
              <el-switch v-model="form.pool_enabled" />
            </el-form-item>
            <el-form-item label="日收款上限">
              <el-input-number v-model="form.daily_limit_yuan" :min="0" :precision="2" :step="100" style="width: 160px" /> 元
              <span class="hint-inline">0 = 不限</span>
            </el-form-item>
            <el-form-item label="单笔上限">
              <el-input-number v-model="form.single_limit_yuan" :min="0" :precision="2" :step="10" style="width: 160px" /> 元
            </el-form-item>
          </el-tab-pane>
          <el-tab-pane label="密钥证书" name="keys">
            <el-alert type="info" :closable="false" title="编辑时留空表示不修改；密钥更换后旧密钥归档 7 天" style="margin-bottom: 12px" />
            <el-form-item label="商户密钥">
              <el-input v-model="form.mch_key" type="password" show-password />
            </el-form-item>
            <el-form-item v-if="form.pay_type === 1" label="证书序列号">
              <el-input v-model="form.serial_no" />
            </el-form-item>
            <el-form-item label="商户私钥">
              <el-input v-model="form.private_key" type="textarea" :rows="3" />
            </el-form-item>
            <el-form-item v-if="form.pay_type === 2" label="平台公钥">
              <el-input v-model="form.public_key" type="textarea" :rows="3" />
            </el-form-item>
          </el-tab-pane>
        </el-tabs>
      </el-form>
      <template #footer>
        <el-button @click="dialog.visible = false">取消</el-button>
        <el-button type="primary" :loading="dialog.saving" @click="save">保存</el-button>
      </template>
    </el-dialog>

    <!-- 分配到平台 -->
    <el-dialog v-model="assign.visible" title="分配到代收平台" width="420px">
      <p class="assign-mch">商户号：<strong>{{ assign.mch_no }}</strong></p>
      <el-select v-model="assign.platform_id" filterable placeholder="选择平台" style="width: 100%">
        <el-option v-for="p in platformOptions" :key="p.id" :label="p.platform_name" :value="p.id" />
      </el-select>
      <template #footer>
        <el-button @click="assign.visible = false">取消</el-button>
        <el-button type="primary" :loading="assign.saving" @click="saveAssign">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import http, { fen2yuan, payTypeText } from '../api'
import EmptyState from './EmptyState.vue'
import { loadPlatformOptions, platformName } from '../utils/platforms.js'
import { quotaPercent, quotaStatus } from '../utils/format.js'

const loading = ref(false)
const rows = ref([])
const keyword = ref('')
const payType = ref(null)
const assignFilter = ref('all')
const platformOptions = ref([])
const platformMap = ref({})
const formTab = ref('basic')
const dialog = reactive({ visible: false, isEdit: false, saving: false })
const assign = reactive({ visible: false, saving: false, channel_id: '', mch_no: '', platform_id: '' })
const form = reactive({
  id: '', pay_type: 1, pool_enabled: true, daily_limit_yuan: 0, single_limit_yuan: 0,
  mch_no: '', mch_key: '', app_id: '', serial_no: '', private_key: '', public_key: '', remark: '',
})

const kpi = computed(() => {
  let unassigned = 0
  for (const r of rows.value) {
    if (isFree(r)) unassigned++
  }
  return { total: rows.value.length, unassigned, assigned: rows.value.length - unassigned }
})

const filteredRows = computed(() => {
  let list = rows.value
  if (keyword.value.trim()) {
    const kw = keyword.value.trim().toLowerCase()
    list = list.filter((r) => (r.mch_no || '').toLowerCase().includes(kw))
  }
  if (assignFilter.value === 'free') list = list.filter(isFree)
  else if (assignFilter.value === 'bound') list = list.filter((r) => !isFree(r))
  return list
})

function isFree(row) {
  return !row.platform_id || row.platform_id === '0'
}

function yuan2fen(v) {
  return Math.round((Number(v) || 0) * 100)
}

async function load() {
  loading.value = true
  try {
    const params = {}
    if (payType.value) params.pay_type = payType.value
    const data = await http.get('/admin/channel/pool', { params })
    rows.value = data.list || []
  } finally {
    loading.value = false
  }
}

function openAdd() {
  dialog.isEdit = false
  formTab.value = 'basic'
  Object.assign(form, {
    id: '', pay_type: 1, pool_enabled: true, daily_limit_yuan: 0, single_limit_yuan: 0,
    mch_no: '', mch_key: '', app_id: '', serial_no: '', private_key: '', public_key: '', remark: '',
  })
  dialog.visible = true
}

function openEdit(row) {
  dialog.isEdit = true
  formTab.value = 'basic'
  Object.assign(form, {
    id: row.id, pay_type: row.pay_type, pool_enabled: row.pool_enabled === 1,
    daily_limit_yuan: row.daily_limit_fen ? row.daily_limit_fen / 100 : 0,
    single_limit_yuan: row.single_limit_fen ? row.single_limit_fen / 100 : 0,
    mch_no: row.mch_no || '', mch_key: '', app_id: row.app_id || '', serial_no: '', private_key: '', public_key: '', remark: row.remark || '',
  })
  dialog.visible = true
}

async function save() {
  if (!form.mch_no.trim() && !dialog.isEdit) return ElMessage.warning('请填写商户号')
  dialog.saving = true
  try {
    if (dialog.isEdit) {
      const body = {
        id: form.id,
        pool_enabled: form.pool_enabled ? 1 : 0,
        daily_limit_fen: yuan2fen(form.daily_limit_yuan),
        single_limit_fen: yuan2fen(form.single_limit_yuan),
      }
      for (const k of ['mch_no', 'mch_key', 'app_id', 'serial_no', 'private_key', 'public_key', 'remark']) {
        if (form[k]) body[k] = form[k]
      }
      await http.put('/admin/channel/edit', body)
      ElMessage.success('已保存')
    } else {
      await http.post('/admin/channel/pool-add', {
        pay_type: form.pay_type,
        mch_no: form.mch_no.trim(),
        pool_enabled: form.pool_enabled ? 1 : 0,
        daily_limit_fen: yuan2fen(form.daily_limit_yuan),
        single_limit_fen: yuan2fen(form.single_limit_yuan),
        mch_key: form.mch_key, app_id: form.app_id, serial_no: form.serial_no,
        private_key: form.private_key, public_key: form.public_key, remark: form.remark,
      })
      ElMessage.success('码池入库成功')
    }
    dialog.visible = false
    load()
  } finally {
    dialog.saving = false
  }
}

async function toggleStatus(row, on) {
  await http.put('/admin/channel/status', { id: row.id, status: on ? 1 : 0 })
  row.status = on ? 1 : 0
}

function openAssign(row) {
  assign.channel_id = row.id
  assign.mch_no = row.mch_no
  assign.platform_id = ''
  assign.visible = true
}

async function saveAssign() {
  if (!assign.platform_id) return ElMessage.warning('请选择平台')
  assign.saving = true
  try {
    await http.put('/admin/platform/bind-channels', {
      platform_id: assign.platform_id, channel_ids: [assign.channel_id],
    })
    ElMessage.success('已分配')
    assign.visible = false
    load()
  } finally {
    assign.saving = false
  }
}

onMounted(async () => {
  const plat = await loadPlatformOptions()
  platformOptions.value = plat.list
  platformMap.value = plat.map
  load()
})
</script>

<style scoped>
.page-tip { margin: 0 0 16px; font-size: 13px; color: var(--text-secondary); line-height: 1.6; }
.hint-text { font-size: 12px; color: var(--text-muted); }
.hint-inline { margin-left: 8px; font-size: 12px; color: var(--text-muted); }
.assign-mch { margin: 0 0 12px; font-size: 14px; }
</style>
