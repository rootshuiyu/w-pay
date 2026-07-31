<template>
  <el-card shadow="never" class="page-card">
    <div class="page-toolbar">
      <el-select v-model="query.platformIds" multiple collapse-tags placeholder="全部平台" style="width: 220px" filterable clearable>
        <el-option v-for="p in platformOptions" :key="p.id" :label="p.platform_name" :value="p.id" />
      </el-select>
      <el-select v-model="query.storeIds" multiple collapse-tags placeholder="全部门店" style="width: 200px" filterable clearable>
        <el-option v-for="s in storeOptions" :key="s.id" :label="s.store_name" :value="s.id" />
      </el-select>
      <el-select v-model="query.pay_type" placeholder="支付方式" clearable style="width: 110px">
        <el-option label="微信" :value="1" />
        <el-option label="支付宝" :value="2" />
      </el-select>
      <el-select v-model="query.status" placeholder="状态" clearable style="width: 110px">
        <el-option label="待支付" :value="0" />
        <el-option label="已支付" :value="1" />
        <el-option label="已关闭" :value="2" />
        <el-option label="退款" :value="3" />
      </el-select>
      <el-date-picker
        v-model="query.range"
        type="daterange"
        value-format="YYYY-MM-DD"
        start-placeholder="开始"
        end-placeholder="结束"
        style="width: 240px"
      />
      <el-input v-model="query.order_no" placeholder="订单号" clearable style="width: 180px" @keyup.enter="onSearch" />
      <el-button type="primary" @click="onSearch"><el-icon><Search /></el-icon>查询</el-button>
      <el-button @click="resetQuery">重置</el-button>
    </div>

    <el-table :data="rows" v-loading="loading" stripe empty-text=" ">
      <el-table-column prop="order_id" label="订单号" width="190">
        <template #default="{ row }"><span class="mono">{{ row.order_id }}</span></template>
      </el-table-column>
      <el-table-column label="平台" min-width="110">
        <template #default="{ row }">{{ platformName(row.platform_id) }}</template>
      </el-table-column>
      <el-table-column label="门店" min-width="110">
        <template #default="{ row }">{{ storeName(row.store_id) }}</template>
      </el-table-column>
      <el-table-column label="方式" width="80">
        <template #default="{ row }">
          <el-tag size="small" :type="row.pay_type === 1 ? 'success' : 'primary'">{{ payTypeText(row.pay_type) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="金额" width="90" align="right">
        <template #default="{ row }">￥{{ fen2yuan(row.total_amount) }}</template>
      </el-table-column>
      <el-table-column label="状态" width="88">
        <template #default="{ row }">
          <el-tag size="small" :type="statusType(row.order_status)">{{ orderStatusText(row.order_status) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="subject" label="备注" min-width="100" show-overflow-tooltip />
      <el-table-column label="创建时间" width="160">
        <template #default="{ row }">{{ fmtTime(row.created_at) }}</template>
      </el-table-column>
    </el-table>

    <EmptyState v-if="!loading && rows.length === 0" title="无匹配订单" description="调整筛选条件或扩大日期范围" />

    <el-pagination
      class="page-pager"
      v-model:current-page="query.page"
      v-model:page-size="query.page_size"
      :total="total"
      layout="total, sizes, prev, pager, next"
      :page-sizes="[10, 20, 50, 100]"
      @change="load"
    />
  </el-card>
</template>

<script setup>
import { onMounted, reactive, ref } from 'vue'
import http, { payTypeText, orderStatusText, fen2yuan } from '../api'
import EmptyState from '../components/EmptyState.vue'
import { fmtTime, statusType } from '../utils/format.js'
import { loadStoreOptions, storeName as mapName } from '../utils/stores.js'
import { loadPlatformOptions, platformName as mapPlatformName } from '../utils/platforms.js'

const rows = ref([])
const total = ref(0)
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
  platformIds: [],
  storeIds: [],
  pay_type: null,
  status: null,
  range: [dayStr(-6), dayStr()],
  order_no: '',
  page: 1,
  page_size: 20,
})

const storeName = (id) => mapName(storeMap.value, id)
const platformName = (id) => mapPlatformName(platformMap.value, id)

function onSearch() {
  query.page = 1
  load()
}

function resetQuery() {
  Object.assign(query, {
    platformIds: [], storeIds: [], pay_type: null, status: null,
    range: [dayStr(-6), dayStr()], order_no: '', page: 1,
  })
  load()
}

async function load() {
  loading.value = true
  try {
    const params = { page: query.page, page_size: query.page_size }
    if (query.platformIds.length) params.platform_ids = query.platformIds.join(',')
    if (query.storeIds.length) params.store_ids = query.storeIds.join(',')
    if (query.pay_type !== null && query.pay_type !== '') params.pay_type = query.pay_type
    if (query.status !== null && query.status !== '') params.order_status = query.status
    if (query.range?.length === 2) {
      params.start_time = query.range[0]
      params.end_time = query.range[1]
    }
    if (query.order_no) params.order_no = query.order_no
    const data = await http.get('/admin/order/list', { params })
    rows.value = data.list || []
    total.value = data.total || 0
  } finally {
    loading.value = false
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
