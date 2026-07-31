import { createRouter, createWebHashHistory } from 'vue-router'

import { getToken, getUser } from './utils/auth'

const routes = [
  { path: '/login', component: () => import('./views/Login.vue') },
  {
    path: '/',
    component: () => import('./views/Layout.vue'),
    redirect: '/dashboard',
    children: [
      { path: 'dashboard', component: () => import('./views/Dashboard.vue'), meta: { title: '工作台', subtitle: '业务概览与功能入口', roles: ['super_admin', 'finance', 'operator'] } },

      { path: 'platforms', component: () => import('./views/Platforms.vue'), meta: { title: '代收平台', subtitle: '平台管理 · 码池监控', roles: ['super_admin'] } },

      { path: 'pay-api', component: () => import('./views/PayApi.vue'), meta: { title: '收银 API', subtitle: '对接方接口文档', roles: ['super_admin', 'operator'] } },

      { path: 'orders', component: () => import('./views/Orders.vue'), meta: { title: '订单', subtitle: '全平台订单查询', roles: ['super_admin', 'finance', 'operator'] } },

      { path: 'stats', component: () => import('./views/Stats.vue'), meta: { title: '对账', subtitle: '营收统计与导出', roles: ['super_admin', 'finance'] } },

      { path: 'settings', component: () => import('./views/Settings.vue'), meta: { title: '系统设置', subtitle: '码池 · 安全', roles: ['super_admin', 'operator'] } },

      // 旧路径兼容
      { path: 'channel-pool', redirect: { path: '/platforms', query: { tab: 'pool' } } },
      { path: 'channels', redirect: { path: '/settings', query: { tab: 'pool' } } },
      { path: 'stores', redirect: { path: '/settings', query: { tab: 'stores' } } },
      { path: 'whitelist', redirect: { path: '/settings', query: { tab: 'security' } } },
    ],
  },
]

const router = createRouter({
  history: createWebHashHistory(),
  routes,
})

function currentRole() {
  try {
    return getUser().role || ''
  } catch {
    return ''
  }
}

router.beforeEach((to) => {
  if (to.path !== '/login' && !getToken()) {
    return '/login'
  }
  const roles = to.meta.roles
  if (roles && roles.length && !roles.includes(currentRole())) {
    return '/dashboard'
  }
})

export default router
