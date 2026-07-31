<template>
  <el-card shadow="never" class="page-card">
    <div class="page-toolbar">
      <el-radio-group v-model="query.dimension" @change="load">
        <el-radio-button value="store">按门店</el-radio-button>
        <el-radio-button value="platform">按平台</el-radio-button>
      </el-radio-group>
      <el-select
        v-if="query.dimension === 'store'"
        v-model="query.storeIds"
        multiple collapse-tags placeholder="全部门店" style="width: 220px" filterable clearable
      >
        <el-option v-for="s in storeOptions" :key="s.id" :label="s.store_name" :value="s.id" />
      </el-select>
      <el-select
        v-else
        v-model="query.platformIds"
        multiple collapse-tags placeholder="全部平台" style="width: 220px" filterable clearable
      >
        <el-option v-for="p in platformOptions" :key="p.id" :label="p.platform_name" :value="p.id" />
      </el-select>
      <el-date-picker
        v-model="query.range"
        type="daterange"
        value-format="YYYY-MM-DD"
        start-placeholder="开始"
        end-placeholder="结束"
        style="width: 240px"
        :clearable="false"
      />
      <el-radio-group v-model="query.group_by">
        <el-radio-button value="day">按日</el-radio-button>
        <el-radio-button value="month">按月</el-radio-button>
      </el-radio-group>
      <el-button type="primary" @click="load"><el-icon><Search /></el-icon>统计</el-button>
      <div class="page-toolbar-spacer" />
      <el-dropdown trigger="click" @command="onExport">
        <el-button><el-icon><Download /></el-icon>导出<el-icon class="el-icon--right"><ArrowDown /></el-icon></el-button>
        <template #dropdown>
          <el-dropdown-menu>
            <el-dropdown-item command="stat">汇总 Excel</el-dropdown-item>
            <el-dropdown-item command="orders">订单明细 Excel</el-dropdown-item>
          </el-dropdown-menu>
        </template>
      </el-dropdown>
    </div>

    <el-row :gutter="16" class="kpi-grid">
      <el-col v-for="item in summaryItems" :key="item.label" :xs="12" :sm="6">
        <el-card shadow="hover" class="kpi-card page-card">
          <div class="kpi-num" :class="item.cls">{{ item.value }}</div>
          <div class="kpi-label">{{ item.label }}</div>
        </el-card>
      </el-col>
    </el-row>

    <el-table :data="rows" v-loading="loading" stripe empty-text=" ">
      <el-table-column prop="stat_date" label="日期" width="120" />
      <el-table-column :label="query.dimension === 'platform' ? '代收平台' : '门店'" min-width="140">
        <template #default="{ row }">
          {{ query.dimension === 'platform' ? platformName(row.platform_id) : storeName(row.store_id) }}
        </template>
      </el-table-column>
      <el-table-column prop="total_count" label="订单数" width="90" align="right" />
      <el-table-column label="订单金额" width="120" align="right">
        <template #default="{ row }">￥{{ fen2yuan(row.total_amount) }}</template>
      </el-table-column>
      <el-table-column prop="paid_count" label="支付笔数" width="100" align="right" />
      <el-table-column label="实收金额" width="120" align="right">
        <template #default="{ row }"><b>￥{{ fen2yuan(row.paid_amount) }}</b></template>
      </el-table-column>
    </el-table>

    <EmptyState v-if="!loading && rows.length === 0" title="所选时段无数据" />
  </el-card>
</template>

<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import http, { download, fen2yuan } from '../api'
import EmptyState from '../components/EmptyState.vue'
import { loadStoreOptions, storeName as mapName } from '../utils/stores.js'
import { loadPlatformOptions, platformName as mapPlatformName } from '../utils/platforms.js'

const rows = ref([])
const loading = ref(false)
const storeOptions = ref([])
const storeMap = ref({})
const platformOptions = ref([])
const platformMap = ref({})

function dayStr(offset = 0) {
  const d = new Date(Date.now() + offset * 86400000)
  return d.toISOString().slice(0, 10)
}

const query = reactive({
  storeIds: [],
  platformIds: [],
  range: [dayStr(-29), dayStr()],
  group_by: 'day',
  dimension: 'platform',
})

const storeName = (id) => mapName(storeMap.value, id)
const platformName = (id) => mapPlatformName(platformMap.value, id)

const summary = computed(() => {
  const s = { totalCount: 0, totalAmount: 0, paidCount: 0, paidAmount: 0 }
  for (const r of rows.value) {
    s.totalCount += Number(r.total_count || 0)
    s.totalAmount += Number(r.total_amount || 0)
    s.paidCount += Number(r.paid_count || 0)
    s.paidAmount += Number(r.paid_amount || 0)
  }
  return s
})

const summaryItems = computed(() => [
  { label: '订单总数', value: summary.value.totalCount, cls: '' },
  { label: '订单总额', value: `￥${fen2yuan(summary.value.totalAmount)}`, cls: '' },
  { label: '已支付笔数', value: summary.value.paidCount, cls: '' },
  { label: '实收金额', value: `￥${fen2yuan(summary.value.paidAmount)}`, cls: 'success' },
])

function buildParams() {
  const params = {
    start_time: query.range[0],
    end_time: query.range[1],
    group_by: query.group_by,
    dimension: query.dimension,
  }
  if (query.dimension === 'store' && query.storeIds.length) params.store_ids = query.storeIds.join(',')
  if (query.dimension === 'platform' && query.platformIds.length) params.platform_ids = query.platformIds.join(',')
  return params
}

async function load() {
  loading.value = true
  try {
    const data = await http.get('/admin/stat/summary', { params: buildParams() })
    rows.value = data.stats || []
    if (data.platforms?.length) {
      for (const p of data.platforms) platformMap.value[p.id] = p.platform_name
    }
    if (data.stores?.length) {
      for (const s of data.stores) storeMap.value[s.id] = s.store_name
    }
  } finally {
    loading.value = false
  }
}

function onExport(cmd) {
  const base = buildParams()
  if (cmd === 'stat') download('/admin/export/stat', base)
  else {
    download('/admin/export/orders', {
      start_time: query.range[0],
      end_time: query.range[1],
      store_ids: query.storeIds.join(',') || undefined,
      platform_ids: query.platformIds.join(',') || undefined,
    })
  }
}

onMounted(async () => {
  const [{ list, map }, plat] = await Promise.all([loadStoreOptions(), loadPlatformOptions()])
  storeOptions.value = list
  storeMap.value = map
  platformOptions.value = plat.list
  platformMap.value = plat.map
  load()
})
</script>
