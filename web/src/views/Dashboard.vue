<template>
  <div class="dashboard">
    <!-- 业务流程 -->
    <el-card shadow="never" class="page-card flow-card">
      <div class="flow-title">系统怎么用</div>
      <div class="flow-steps">
        <div class="flow-step" @click="go('/settings', { tab: 'pool' })">
          <span class="num">1</span>
          <span class="text">录入商户码<br><small>进入公共码池</small></span>
        </div>
        <el-icon class="flow-arrow"><ArrowRight /></el-icon>
        <div class="flow-step" @click="go('/platforms')">
          <span class="num">2</span>
          <span class="text">新建代收平台<br><small>绑定码 + 配 IP</small></span>
        </div>
        <el-icon class="flow-arrow"><ArrowRight /></el-icon>
        <div class="flow-step" @click="go('/platforms', { tab: 'pool' })">
          <span class="num">3</span>
          <span class="text">监控码池额度<br><small>轮询收款</small></span>
        </div>
        <el-icon class="flow-arrow"><ArrowRight /></el-icon>
        <div class="flow-step" @click="go('/pay-api')">
          <span class="num">4</span>
          <span class="text">导出文档<br><small>发给对方对接</small></span>
        </div>
        <el-icon class="flow-arrow"><ArrowRight /></el-icon>
        <div class="flow-step" @click="go('/orders')">
          <span class="num">5</span>
          <span class="text">查订单 / 对账<br><small>财务核对</small></span>
        </div>
      </div>
    </el-card>

    <!-- KPI -->
    <el-row :gutter="16" class="kpi-grid">
      <el-col v-for="item in kpiItems" :key="item.label" :xs="12" :sm="6">
        <el-card shadow="hover" class="kpi-card page-card" :class="{ clickable: item.to }" @click="item.to && go(item.to, item.query)">
          <div class="kpi-num" :class="item.cls">{{ item.value }}</div>
          <div class="kpi-label">{{ item.label }}</div>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="16">
      <!-- 功能模块 -->
      <el-col :xs="24" :lg="10">
        <el-card shadow="never" class="page-card">
          <template #header><span>功能模块</span></template>
          <div class="modules">
            <div v-for="m in modules" :key="m.title" class="module-item" @click="go(m.to, m.query)">
              <el-icon :size="22" :color="m.color"><component :is="m.icon" /></el-icon>
              <div class="module-body">
                <div class="module-title">{{ m.title }}</div>
                <div class="module-desc">{{ m.desc }}</div>
              </div>
              <el-badge v-if="m.badge" :value="m.badge" type="warning" />
              <el-icon class="arrow"><ArrowRight /></el-icon>
            </div>
          </div>
        </el-card>
      </el-col>

      <!-- 最近订单 -->
      <el-col :xs="24" :lg="14">
        <el-card shadow="never" class="page-card">
          <template #header>
            <div class="card-head">
              <span>今日订单</span>
              <el-button link type="primary" @click="go('/orders')">全部</el-button>
            </div>
          </template>
          <el-table :data="recentOrders" v-loading="loading" size="small" stripe max-height="360" empty-text=" ">
            <el-table-column prop="order_id" label="订单号" min-width="160">
              <template #default="{ row }"><span class="mono">{{ row.order_id }}</span></template>
            </el-table-column>
            <el-table-column label="金额" width="80" align="right">
              <template #default="{ row }">￥{{ fen2yuan(row.total_amount) }}</template>
            </el-table-column>
            <el-table-column label="状态" width="76">
              <template #default="{ row }">
                <el-tag size="small" :type="statusType(row.order_status)">{{ orderStatusText(row.order_status) }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="时间" width="140">
              <template #default="{ row }">{{ fmtTime(row.created_at) }}</template>
            </el-table-column>
          </el-table>
          <EmptyState v-if="!loading && recentOrders.length === 0" title="今日暂无订单" description="对接方调用收银 API 后订单会出现在这里" />
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import http, { fen2yuan, orderStatusText } from '../api'
import EmptyState from '../components/EmptyState.vue'
import { fmtTime, statusType } from '../utils/format.js'
import { getUser } from '../utils/auth'

const router = useRouter()
const user = getUser()
const loading = ref(false)
const recentOrders = ref([])
const summary = ref({
  platformCount: 0,
  channelCount: 0,
  poolWarn: 0,
  unassignedCount: 0,
  todayOrders: 0,
  todayPaid: 0,
  pending: 0,
})

const today = () => new Date().toISOString().slice(0, 10)

function go(path, query) {
  router.push(query ? { path, query } : path)
}

const kpiItems = computed(() => {
  const items = [
    { label: '今日订单', value: summary.value.todayOrders, cls: '', to: '/orders' },
    { label: '今日实收', value: `￥${fen2yuan(summary.value.todayPaid)}`, cls: 'success', to: '/stats' },
    { label: '待支付', value: summary.value.pending, cls: 'warn', to: '/orders' },
  ]
  if (user.role === 'super_admin') {
    items.unshift(
      { label: '未分配码', value: summary.value.unassignedCount ?? '—', cls: 'warn', to: '/platforms', query: { tab: 'pool' } },
      { label: '代收平台', value: summary.value.platformCount, cls: '', to: '/platforms' },
    )
  } else if (user.role === 'operator') {
    items.unshift({ label: '门店', value: summary.value.storeCount || '—', cls: '', to: '/settings', query: { tab: 'stores' } })
  }
  return items
})

const modules = computed(() => {
  const all = [
    { title: '代收平台', desc: '新建平台、绑定码池、配 IP、导出文档', icon: 'Grid', color: '#1d2b53', to: '/platforms', roles: ['super_admin'] },
    { title: '码池监控', desc: '商户码额度、分配状态', icon: 'Monitor', color: '#14805e', to: '/platforms', query: { tab: 'pool' }, badge: summary.value.poolWarn || null, roles: ['super_admin'] },
    { title: '商户码入库', desc: '录入微信/支付宝官方商户号', icon: 'Key', color: '#626aef', to: '/settings', query: { tab: 'pool' }, roles: ['super_admin'] },
    { title: '收银 API', desc: '对接文档、请求示例', icon: 'Link', color: '#409eff', to: '/pay-api', roles: ['super_admin', 'operator'] },
    { title: '订单查询', desc: '全平台订单筛选', icon: 'Tickets', color: '#e6a23c', to: '/orders', roles: ['super_admin', 'finance', 'operator'] },
    { title: '对账导出', desc: '营收统计、Excel', icon: 'DataAnalysis', color: '#67c23a', to: '/stats', roles: ['super_admin', 'finance'] },
    { title: '系统设置', desc: '码池入库、IP 白名单', icon: 'Setting', color: '#909399', to: '/settings', query: { tab: 'pool' }, roles: ['super_admin'] },
    { title: '门店档案', desc: '收款主体信息', icon: 'Shop', color: '#909399', to: '/settings', query: { tab: 'stores' }, roles: ['operator'] },
  ]
  return all.filter((m) => m.roles.includes(user.role))
})

async function load() {
  loading.value = true
  const d = today()
  try {
    const tasks = [
      http.get('/admin/order/list', { params: { page: 1, page_size: 8, start_time: d, end_time: d } }),
      http.get('/admin/order/list', { params: { page: 1, page_size: 1, order_status: 0, start_time: d, end_time: d } }),
    ]
    if (user.role === 'super_admin') {
      tasks.push(http.get('/admin/platform/list', { params: { page: 1, page_size: 1 } }))
      tasks.push(http.get('/admin/channel/pool'))
    }
    if (user.role === 'operator') {
      tasks.push(http.get('/admin/store/list', { params: { page: 1, page_size: 1 } }))
    }
    const results = await Promise.all(tasks)
    let i = 0
    recentOrders.value = results[i++].list || []
    summary.value.todayOrders = results[0].total || 0
    summary.value.pending = results[i++].total || 0

    if (user.role === 'super_admin') {
      summary.value.platformCount = results[i++].total || 0
      const pool = results[i++].list || []
      summary.value.channelCount = pool.length
      summary.value.unassignedCount = pool.filter((r) => !r.platform_id || r.platform_id === '0').length
      summary.value.poolWarn = pool.filter((r) => {
        if (r.status !== 1) return true
        if (r.daily_limit_fen > 0 && r.daily_used_fen >= r.daily_limit_fen) return true
        if (r.daily_limit_fen > 0 && r.daily_used_fen / r.daily_limit_fen >= 0.8) return true
        return false
      }).length
    }
    if (user.role === 'operator') {
      summary.value.storeCount = results[i++].total || 0
    }

    if (['super_admin', 'finance'].includes(user.role)) {
      const stats = await http.get('/admin/stat/summary', { params: { start_time: d, end_time: d, group_by: 'day' } })
      summary.value.todayPaid = (stats.stats || []).reduce((s, r) => s + Number(r.paid_amount || 0), 0)
    }
  } finally {
    loading.value = false
  }
}

onMounted(load)
</script>

<style scoped>
.flow-card { margin-bottom: 16px; }
.flow-title { font-size: 14px; font-weight: 600; margin-bottom: 16px; color: var(--text-primary); }
.flow-steps { display: flex; align-items: center; flex-wrap: wrap; gap: 8px; }
.flow-step {
  display: flex; align-items: center; gap: 10px; padding: 12px 16px;
  background: #f8f9fb; border-radius: 8px; cursor: pointer; flex: 1; min-width: 140px;
  transition: background 0.15s;
}
.flow-step:hover { background: #eef1f6; }
.flow-step .num {
  width: 28px; height: 28px; border-radius: 50%; background: var(--brand-primary); color: #fff;
  display: flex; align-items: center; justify-content: center; font-size: 13px; font-weight: 700; flex-shrink: 0;
}
.flow-step .text { font-size: 13px; line-height: 1.4; color: var(--text-primary); }
.flow-step small { color: var(--text-muted); }
.flow-arrow { color: var(--text-muted); flex-shrink: 0; }
.kpi-card.clickable { cursor: pointer; transition: box-shadow 0.2s; }
.kpi-card.clickable:hover { box-shadow: 0 4px 12px rgba(0,0,0,0.08); }
.modules { display: flex; flex-direction: column; gap: 4px; }
.module-item {
  display: flex; align-items: center; gap: 12px; padding: 12px; border-radius: 8px; cursor: pointer;
  transition: background 0.15s;
}
.module-item:hover { background: #f5f7fa; }
.module-body { flex: 1; min-width: 0; }
.module-title { font-size: 14px; font-weight: 600; }
.module-desc { font-size: 12px; color: var(--text-secondary); margin-top: 2px; }
.module-item .arrow { color: var(--text-muted); }
.card-head { display: flex; justify-content: space-between; align-items: center; }
.mono { font-family: Consolas, monospace; font-size: 12px; }
@media (max-width: 992px) {
  .flow-arrow { display: none; }
  .flow-step { min-width: calc(50% - 8px); }
}
</style>
