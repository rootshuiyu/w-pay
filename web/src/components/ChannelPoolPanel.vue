<template>
  <div>
    <el-row :gutter="16" class="kpi-grid">
      <el-col :xs="12" :sm="6">
        <el-card shadow="hover" class="kpi-card page-card">
          <div class="kpi-num">{{ kpi.total }}</div>
          <div class="kpi-label">池中商户码</div>
        </el-card>
      </el-col>
      <el-col :xs="12" :sm="6">
        <el-card shadow="hover" class="kpi-card page-card">
          <div class="kpi-num success">{{ kpi.available }}</div>
          <div class="kpi-label">当前可用</div>
        </el-card>
      </el-col>
      <el-col :xs="12" :sm="6">
        <el-card shadow="hover" class="kpi-card page-card">
          <div class="kpi-num warn">{{ kpi.nearLimit }}</div>
          <div class="kpi-label">额度 ≥80%</div>
        </el-card>
      </el-col>
      <el-col :xs="12" :sm="6">
        <el-card shadow="hover" class="kpi-card page-card">
          <div class="kpi-num danger">{{ kpi.exhausted }}</div>
          <div class="kpi-label">已满额/停用</div>
        </el-card>
      </el-col>
    </el-row>

    <div class="page-toolbar">
      <el-select v-model="payType" placeholder="支付方式" clearable style="width: 130px" @change="load">
        <el-option label="全部" :value="null" />
        <el-option label="微信" :value="1" />
        <el-option label="支付宝" :value="2" />
      </el-select>
        <el-select v-model="platformFilter" placeholder="代收平台" clearable filterable style="width: 200px" @change="load">
          <el-option label="未分配池" value="0" />
          <el-option v-for="p in platformOptions" :key="p.id" :label="p.platform_name" :value="String(p.id)" />
        </el-select>
        <el-select v-model="storeFilter" placeholder="门店" clearable filterable style="width: 220px">
        <el-option v-for="s in storeOptions" :key="s.id" :label="s.store_name" :value="s.id" />
      </el-select>
      <el-select v-model="statusFilter" placeholder="状态" style="width: 120px">
        <el-option label="全部" value="all" />
        <el-option label="可用" value="ok" />
        <el-option label="告警" value="warn" />
        <el-option label="不可用" value="bad" />
      </el-select>
      <div class="page-toolbar-spacer" />
      <el-button @click="load"><el-icon><Refresh /></el-icon>刷新</el-button>
    </div>

    <el-table :data="filteredRows" v-loading="loading" stripe>
      <el-table-column label="门店" min-width="120">
        <template #default="{ row }">{{ storeName(storeMap, row.store_id) }}</template>
      </el-table-column>
      <el-table-column label="渠道" width="90">
        <template #default="{ row }">
          <el-tag size="small" :type="row.pay_type === 1 ? 'success' : 'primary'">
            {{ payTypeText(row.pay_type) }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="mch_no" label="商户号" width="140" />
      <el-table-column label="所属平台" min-width="120">
        <template #default="{ row }">
          <el-tag v-if="!row.platform_id || row.platform_id === '0'" size="small" type="info">未分配</el-tag>
          <span v-else>{{ platformName(platformMap, row.platform_id) }}</span>
        </template>
      </el-table-column>
      <el-table-column label="日额度" min-width="180">
        <template #default="{ row }">
          <div class="quota-cell">
            <el-progress
              :percentage="quotaPercent(row.daily_used_fen, row.daily_limit_fen)"
              :status="quotaStatus(row.daily_used_fen, row.daily_limit_fen)"
              :stroke-width="8"
            />
            <span class="hint-text">
              {{ row.daily_limit_fen > 0
                ? `￥${fen2yuan(row.daily_used_fen)} / ￥${fen2yuan(row.daily_limit_fen)}`
                : `已收 ￥${fen2yuan(row.daily_used_fen)}（不限）` }}
            </span>
          </div>
        </template>
      </el-table-column>
      <el-table-column label="单笔上限" width="100" align="right">
        <template #default="{ row }">
          {{ row.single_limit_fen > 0 ? `￥${fen2yuan(row.single_limit_fen)}` : '不限' }}
        </template>
      </el-table-column>
      <el-table-column label="状态" width="100">
        <template #default="{ row }">
          <el-tag size="small" :type="rowStatusType(row)">{{ rowStatusText(row) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="140" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" @click="openConfigure(row)">配置</el-button>
          <el-button v-if="!row.platform_id || row.platform_id === '0'" link @click="openAssign(row)">分配</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog v-model="dialog.visible" title="商户码配置" width="560px" destroy-on-close>
      <el-form :model="form" label-width="110px">
        <el-form-item label="商户号">
          <el-input v-model="form.mch_no" disabled />
        </el-form-item>
        <el-form-item label="参与轮询">
          <el-switch v-model="form.pool_enabled" />
          <span class="hint-inline">开启后加入码池轮询</span>
        </el-form-item>
        <el-form-item label="日收款上限">
          <el-input-number v-model="form.daily_limit_yuan" :min="0" :precision="2" :step="100" style="width: 160px" />
          <span class="hint-inline">0 = 不限</span>
        </el-form-item>
        <el-form-item label="单笔上限">
          <el-input-number v-model="form.single_limit_yuan" :min="0" :precision="2" :step="10" style="width: 160px" />
          <span class="hint-inline">0 = 不限</span>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialog.visible = false">取消</el-button>
        <el-button type="primary" :loading="dialog.saving" @click="saveConfigure">保存</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="assign.visible" title="分配到代收平台" width="440px" destroy-on-close>
      <el-form label-width="88px">
        <el-form-item label="商户号"><span>{{ assign.mch_no }}</span></el-form-item>
        <el-form-item label="目标平台" required>
          <el-select v-model="assign.platform_id" filterable placeholder="选择平台" style="width: 100%">
            <el-option v-for="p in platformOptions" :key="p.id" :label="p.platform_name" :value="p.id" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="assign.visible = false">取消</el-button>
        <el-button type="primary" :loading="assign.saving" @click="saveAssign">确定绑定</el-button>
      </template>
    </el-dialog>

    <EmptyState
      v-if="!loading && filteredRows.length === 0"
      title="码池暂无商户码"
      description="请到「系统设置 → 码池入库」录入微信/支付宝官方商户号，录入后进入公共码池等待分配"
    />
  </div>
</template>

<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import http, { fen2yuan, payTypeText } from '../api'
import EmptyState from './EmptyState.vue'
import { loadStoreOptions, storeName } from '../utils/stores.js'
import { platformName } from '../utils/platforms.js'
import { quotaPercent, quotaStatus } from '../utils/format.js'

const emit = defineEmits(['saved'])

const loading = ref(false)
const rows = ref([])
const storeOptions = ref([])
const storeMap = ref({})
const platformMap = ref({})
const payType = ref(null)
const platformFilter = ref('')
const platformOptions = ref([])
const storeFilter = ref('')
const statusFilter = ref('all')
const dialog = reactive({ visible: false, saving: false })
const assign = reactive({ visible: false, saving: false, channel_id: '', mch_no: '', platform_id: '' })
const form = reactive({
  id: '', mch_no: '', pool_enabled: true, daily_limit_yuan: 0, single_limit_yuan: 0,
})

const kpi = computed(() => {
  let available = 0
  let nearLimit = 0
  let exhausted = 0
  for (const r of rows.value) {
    const st = rowHealth(r)
    if (st === 'ok') available++
    else if (st === 'warn') nearLimit++
    else exhausted++
  }
  return { total: rows.value.length, available, nearLimit, exhausted }
})

const filteredRows = computed(() => {
  let list = rows.value
  if (storeFilter.value) {
    list = list.filter((r) => String(r.store_id) === String(storeFilter.value))
  }
  if (statusFilter.value !== 'all') {
    list = list.filter((r) => rowHealth(r) === statusFilter.value)
  }
  return list
})

function rowHealth(row) {
  if (row.status !== 1) return 'bad'
  if (row.daily_limit_fen > 0 && row.daily_used_fen >= row.daily_limit_fen) return 'bad'
  if (row.daily_limit_fen > 0 && quotaPercent(row.daily_used_fen, row.daily_limit_fen) >= 80) return 'warn'
  return 'ok'
}

function rowStatusType(row) {
  return { ok: 'success', warn: 'warning', bad: 'danger' }[rowHealth(row)] || 'info'
}

function rowStatusText(row) {
  if (row.status !== 1) return '已停用'
  if (row.daily_limit_fen > 0 && row.daily_used_fen >= row.daily_limit_fen) return '已满额'
  if (row.daily_limit_fen > 0 && quotaPercent(row.daily_used_fen, row.daily_limit_fen) >= 80) return '即将满额'
  return '可用'
}

async function load() {
  loading.value = true
  try {
    const params = {}
    if (payType.value) params.pay_type = payType.value
    if (platformFilter.value !== '' && platformFilter.value != null) params.platform_id = platformFilter.value
    const data = await http.get('/admin/channel/pool', { params })
    rows.value = data.list || []
  } finally {
    loading.value = false
  }
}

function yuan2fen(v) {
  return Math.round((Number(v) || 0) * 100)
}

function openConfigure(row) {
  Object.assign(form, {
    id: row.id,
    mch_no: row.mch_no || '',
    pool_enabled: row.pool_enabled === 1,
    daily_limit_yuan: row.daily_limit_fen ? row.daily_limit_fen / 100 : 0,
    single_limit_yuan: row.single_limit_fen ? row.single_limit_fen / 100 : 0,
  })
  dialog.visible = true
}

async function saveConfigure() {
  dialog.saving = true
  try {
    await http.put('/admin/channel/edit', {
      id: form.id,
      pool_enabled: form.pool_enabled ? 1 : 0,
      daily_limit_fen: yuan2fen(form.daily_limit_yuan),
      single_limit_fen: yuan2fen(form.single_limit_yuan),
    })
    ElMessage.success('配置已保存')
    dialog.visible = false
    await load()
    emit('saved')
  } finally {
    dialog.saving = false
  }
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
      platform_id: assign.platform_id,
      channel_ids: [assign.channel_id],
    })
    ElMessage.success('已分配到平台')
    assign.visible = false
    await load()
    emit('saved')
  } finally {
    assign.saving = false
  }
}

onMounted(async () => {
  const [{ list, map }, plat] = await Promise.all([
    loadStoreOptions(),
    http.get('/admin/platform/list', { params: { page: 1, page_size: 200, status: 1 } }).catch(() => ({ list: [] })),
  ])
  storeOptions.value = list
  storeMap.value = map
  platformOptions.value = plat.list || []
  platformMap.value = Object.fromEntries((plat.list || []).map((p) => [String(p.id), p.platform_name]))
  await load()
})

defineExpose({ load, kpi })
</script>

<style scoped>
.quota-cell { display: flex; flex-direction: column; gap: 4px; }
.hint-inline { margin-left: 8px; color: var(--text-secondary); font-size: 12px; }
</style>
