<template>
  <component :is="hidePool ? 'div' : 'el-card'" shadow="never" :class="hidePool ? '' : 'page-card'">
    <el-alert
      v-if="hidePool"
      type="info"
      :closable="false"
      show-icon
      title="在此录入微信/支付宝官方申请的商户号，默认进入公共码池（未分配）。设置额度后，到「代收平台 → 平台管理」绑定给外部平台轮询收款。"
      style="margin-bottom: 16px"
    />
    <el-tabs v-model="activeTab" @tab-change="onTabChange">
      <el-tab-pane label="按门店查看" name="store">
        <div class="page-toolbar">
          <el-select
            v-model="storeId"
            placeholder="选择门店"
            filterable
            style="width: 280px"
            @change="loadChannels"
          >
            <el-option
              v-for="s in storeOptions"
              :key="s.id"
              :label="`${s.store_name}（${s.id}）`"
              :value="s.id"
            />
          </el-select>
          <div class="page-toolbar-spacer" />
          <el-button type="primary" :disabled="!storeId" @click="openAdd">
            <el-icon><Plus /></el-icon>新增渠道
          </el-button>
        </div>

        <el-empty v-if="!storeId" description="请先选择门店，或到「系统设置 → 门店档案」新建收款主体" />

        <template v-else>
          <el-table :data="rows" v-loading="loading" stripe empty-text=" ">
            <el-table-column label="渠道" width="96">
              <template #default="{ row }">
                <el-tag :type="row.pay_type === 1 ? 'success' : 'primary'" size="small">
                  {{ payTypeText(row.pay_type) }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="mch_no" label="商户号" min-width="130" />
            <el-table-column v-if="hidePool" label="所属平台" width="110">
              <template #default="{ row }">
                <el-tag v-if="!row.platform_id || row.platform_id === '0'" size="small" type="info">未分配</el-tag>
                <span v-else>{{ mapPlatformName(platformMap, row.platform_id) }}</span>
              </template>
            </el-table-column>
            <el-table-column label="轮询" width="72" align="center">
              <template #default="{ row }">
                <el-tag :type="row.pool_enabled === 1 ? 'success' : 'info'" size="small">
                  {{ row.pool_enabled === 1 ? '是' : '否' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column label="日额度" min-width="200">
              <template #default="{ row }">
                <el-progress
                  v-if="row.daily_limit_fen > 0"
                  :percentage="quotaPercent(row.daily_used_fen, row.daily_limit_fen)"
                  :status="quotaStatus(row.daily_used_fen, row.daily_limit_fen)"
                  :stroke-width="8"
                />
                <span v-else class="hint-text">不限 · 已收 ￥{{ fen2yuan(row.daily_used_fen) }}</span>
              </template>
            </el-table-column>
            <el-table-column label="单笔上限" width="96" align="right">
              <template #default="{ row }">
                {{ row.single_limit_fen > 0 ? `￥${fen2yuan(row.single_limit_fen)}` : '不限' }}
              </template>
            </el-table-column>
            <el-table-column label="状态" width="88" align="center">
              <template #default="{ row }">
                <el-switch
                  :model-value="row.status === 1"
                  inline-prompt
                  active-text="启"
                  inactive-text="停"
                  @change="(v) => toggleStatus(row, v)"
                />
              </template>
            </el-table-column>
            <el-table-column label="操作" width="150" fixed="right">
              <template #default="{ row }">
                <el-button link type="primary" @click="openDetail(row)">详情</el-button>
                <el-button link type="primary" @click="openEdit(row)">编辑</el-button>
              </template>
            </el-table-column>
          </el-table>
          <EmptyState
            v-if="!loading && rows.length === 0"
            title="该门店尚未配置支付渠道"
          >
            <el-button type="primary" @click="openAdd">新增渠道</el-button>
          </EmptyState>
        </template>
      </el-tab-pane>

      <el-tab-pane v-if="!hidePool" name="pool" lazy>
        <template #label>
          额度总览
          <el-badge v-if="poolWarnCount > 0" :value="poolWarnCount" class="tab-badge" />
        </template>
        <ChannelPoolPanel v-if="!hidePool" ref="poolPanel" />
      </el-tab-pane>
    </el-tabs>
  </component>

  <el-drawer v-model="detail.visible" title="渠道详情" size="420px">
    <el-descriptions v-if="detail.row" :column="1" border size="small">
      <el-descriptions-item label="渠道 ID"><span class="mono">{{ detail.row.id }}</span></el-descriptions-item>
      <el-descriptions-item label="AppID">{{ detail.row.app_id || '—' }}</el-descriptions-item>
      <el-descriptions-item label="证书序列号">{{ detail.row.serial_no || '—' }}</el-descriptions-item>
      <el-descriptions-item label="密钥">{{ detail.row.mch_key || '—' }}</el-descriptions-item>
      <el-descriptions-item label="回调地址">{{ detail.row.notify_url || defaultNotifyUrl(detail.row) }}</el-descriptions-item>
      <el-descriptions-item label="备注">{{ detail.row.remark || '—' }}</el-descriptions-item>
    </el-descriptions>
  </el-drawer>

  <el-dialog
    v-model="dialog.visible"
    :title="dialog.isEdit ? '编辑支付渠道' : '新增支付渠道'"
    width="620px"
    destroy-on-close
  >
    <el-alert
      v-if="dialog.isEdit"
      type="warning"
      :closable="false"
      title="留空字段表示不修改；更换密钥后旧密钥归档 7 天供延迟回调验签"
      style="margin-bottom: 16px"
    />
    <el-form :model="form" label-width="120px">
      <el-tabs v-model="formTab">
        <el-tab-pane label="基础" name="basic">
          <el-form-item v-if="!dialog.isEdit" label="渠道类型" required>
            <el-radio-group v-model="form.pay_type">
              <el-radio :value="1">微信支付</el-radio>
              <el-radio :value="2">支付宝</el-radio>
            </el-radio-group>
          </el-form-item>
          <el-form-item label="商户号" :required="!dialog.isEdit">
            <el-input v-model="form.mch_no" placeholder="微信商户号 / 支付宝 PID" />
          </el-form-item>
          <el-form-item label="AppID">
            <el-input v-model="form.app_id" placeholder="wx 开头 / 支付宝应用 ID" />
          </el-form-item>
          <el-form-item label="回调地址">
            <el-input v-model="form.notify_url" :placeholder="notifyPlaceholder" />
          </el-form-item>
          <el-form-item label="备注">
            <el-input v-model="form.remark" />
          </el-form-item>
        </el-tab-pane>
        <el-tab-pane label="额度与轮询" name="quota">
          <el-form-item label="参与轮询">
            <el-switch v-model="form.pool_enabled" />
            <span class="hint-inline">开启后加入全池轮询代收</span>
          </el-form-item>
          <el-form-item label="日收款上限">
            <el-input-number v-model="form.daily_limit_yuan" :min="0" :precision="2" :step="100" style="width: 160px" />
            <span class="hint-inline">0 = 不限</span>
          </el-form-item>
          <el-form-item label="单笔上限">
            <el-input-number v-model="form.single_limit_yuan" :min="0" :precision="2" :step="10" style="width: 160px" />
            <span class="hint-inline">0 = 不限</span>
          </el-form-item>
        </el-tab-pane>
        <el-tab-pane label="密钥证书" name="keys">
          <el-form-item label="商户密钥">
            <el-input v-model="form.mch_key" type="password" show-password placeholder="微信 APIv3 密钥" />
          </el-form-item>
          <el-form-item v-if="form.pay_type === 1" label="证书序列号">
            <el-input v-model="form.serial_no" placeholder="微信商户证书序列号" />
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
</template>

<script setup>
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import http, { fen2yuan, payTypeText } from '../api'
import EmptyState from '../components/EmptyState.vue'
import ChannelPoolPanel from '../components/ChannelPoolPanel.vue'
import { loadStoreOptions } from '../utils/stores.js'
import { loadPlatformOptions, platformName as mapPlatformName } from '../utils/platforms.js'
import { quotaPercent, quotaStatus } from '../utils/format.js'

const props = defineProps({ hidePool: { type: Boolean, default: false } })

const route = useRoute()
const activeTab = ref('store')
const poolPanel = ref(null)
const storeOptions = ref([])
const storeId = ref('')
const platformMap = ref({})
const rows = ref([])
const loading = ref(false)
const formTab = ref('basic')
const dialog = reactive({ visible: false, isEdit: false, saving: false })
const detail = reactive({ visible: false, row: null })
const form = reactive({
  id: '', pay_type: 1, pool_enabled: true, daily_limit_yuan: 0, single_limit_yuan: 0,
  mch_no: '', mch_key: '', app_id: '', serial_no: '', private_key: '', public_key: '',
  notify_url: '', remark: '',
})

const poolWarnCount = computed(() => {
  const k = poolPanel.value?.kpi
  return k ? k.nearLimit + k.exhausted : 0
})

const origin = window.location.origin
const notifyPlaceholder = computed(() =>
  form.pay_type === 1
    ? `${origin}/api/notify/wx?store_id=${storeId.value || '{门店ID}'}&channel_id={渠道ID}`
    : `${origin}/api/notify/alipay`
)

function defaultNotifyUrl(row) {
  if (row.pay_type === 1) {
    return `${origin}/api/notify/wx?store_id=${row.store_id}&channel_id=${row.id}`
  }
  return `${origin}/api/notify/alipay`
}

function onTabChange(name) {
  if (name === 'pool') poolPanel.value?.load()
}

function focusStore(id) {
  activeTab.value = 'store'
  storeId.value = String(id)
  loadChannels()
}

function openEditById(id) {
  const row = rows.value.find((r) => String(r.id) === String(id))
  if (row) openEdit(row)
}

async function applyRouteChannelQuery() {
  if (!route.query.store_id) return
  storeId.value = String(route.query.store_id)
  activeTab.value = 'store'
  await loadChannels()
  if (route.query.channel_id) openEditById(route.query.channel_id)
}

async function loadChannels() {
  if (!storeId.value) return
  loading.value = true
  try {
    const data = await http.get('/admin/channel/list', { params: { store_id: storeId.value } })
    rows.value = data.list || []
  } finally {
    loading.value = false
  }
}

function yuan2fen(v) {
  return Math.round((Number(v) || 0) * 100)
}

function poolPayload() {
  return {
    pool_enabled: form.pool_enabled ? 1 : 0,
    daily_limit_fen: yuan2fen(form.daily_limit_yuan),
    single_limit_fen: yuan2fen(form.single_limit_yuan),
  }
}

function openAdd() {
  dialog.isEdit = false
  formTab.value = 'basic'
  Object.assign(form, {
    id: '', pay_type: 1, pool_enabled: true, daily_limit_yuan: 0, single_limit_yuan: 0,
    mch_no: '', mch_key: '', app_id: '', serial_no: '', private_key: '', public_key: '',
    notify_url: '', remark: '',
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
    mch_no: row.mch_no || '', mch_key: '', app_id: '', serial_no: '', private_key: '', public_key: '',
    notify_url: '', remark: '',
  })
  dialog.visible = true
}

function openDetail(row) {
  detail.row = row
  detail.visible = true
}

async function save() {
  dialog.saving = true
  try {
    if (dialog.isEdit) {
      const body = { id: form.id, ...poolPayload() }
      for (const k of ['mch_no', 'mch_key', 'app_id', 'serial_no', 'private_key', 'public_key', 'notify_url', 'remark']) {
        if (form[k]) body[k] = form[k]
      }
      await http.put('/admin/channel/edit', body)
      ElMessage.success('配置已更新')
    } else {
      if (!form.mch_no) {
        ElMessage.warning('请填写商户号')
        return
      }
      await http.post('/admin/channel/add', {
        store_id: storeId.value, pay_type: form.pay_type, ...poolPayload(),
        mch_no: form.mch_no, mch_key: form.mch_key, app_id: form.app_id,
        serial_no: form.serial_no, private_key: form.private_key, public_key: form.public_key,
        notify_url: form.notify_url, remark: form.remark,
      })
      ElMessage.success('渠道已创建')
    }
    dialog.visible = false
    loadChannels()
    poolPanel.value?.load()
  } finally {
    dialog.saving = false
  }
}

async function toggleStatus(row, enabled) {
  if (!enabled) {
    try {
      await ElMessageBox.confirm('关停后该商户码将不再参与轮询', '确认关停', { type: 'warning' })
    } catch {
      return
    }
  }
  await http.put('/admin/channel/status', { id: row.id, status: enabled ? 1 : 0 })
  row.status = enabled ? 1 : 0
  ElMessage.success(enabled ? '已启用' : '已关停')
  poolPanel.value?.load()
}

watch(() => route.query.tab, (tab) => {
  if (tab === 'pool') activeTab.value = 'pool'
})

watch(
  () => [route.query.store_id, route.query.channel_id],
  () => applyRouteChannelQuery(),
)

onMounted(async () => {
  const [{ list }, plat] = await Promise.all([loadStoreOptions(), loadPlatformOptions()])
  storeOptions.value = list
  platformMap.value = plat.map
  if (route.query.tab === 'pool' && !props.hidePool) activeTab.value = 'pool'
  await applyRouteChannelQuery()
})
</script>

<style scoped>
.hint-inline { margin-left: 8px; color: var(--text-secondary); font-size: 12px; }
.tab-badge { margin-left: 6px; vertical-align: middle; }
</style>
