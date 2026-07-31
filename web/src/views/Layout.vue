<template>

  <el-container class="layout">

    <el-aside :width="collapsed ? '64px' : '220px'" class="aside">

      <div class="logo" :class="{ collapsed }">

        <el-icon size="22"><Connection /></el-icon>

        <div v-if="!collapsed" class="logo-text">

          <span class="logo-title">聚合收银</span>

          <span class="logo-sub">多平台代收调度</span>

        </div>

      </div>

      <el-menu
        :default-active="activeMenu"
        :collapse="collapsed"
        router
        background-color="#1d2b53"
        text-color="#c8cee8"
        active-text-color="#ffffff"
      >
        <el-menu-item v-if="can('dashboard')" index="/dashboard">
          <el-icon><Odometer /></el-icon><template #title>工作台</template>
        </el-menu-item>

        <el-menu-item v-if="can('platforms')" index="/platforms">
          <el-icon><Grid /></el-icon><template #title>代收平台</template>
        </el-menu-item>

        <el-menu-item v-if="can('orders')" index="/orders">
          <el-icon><Tickets /></el-icon><template #title>订单</template>
        </el-menu-item>

        <el-menu-item v-if="can('stats')" index="/stats">
          <el-icon><DataAnalysis /></el-icon><template #title>对账</template>
        </el-menu-item>

        <el-menu-item v-if="can('pay-api')" index="/pay-api">
          <el-icon><Link /></el-icon><template #title>收银 API</template>
        </el-menu-item>

        <el-menu-item v-if="can('settings')" index="/settings">
          <el-icon><Setting /></el-icon><template #title>系统设置</template>
        </el-menu-item>
      </el-menu>

      <div v-if="!collapsed" class="aside-footer">v1.0</div>

    </el-aside>

    <el-container>

      <el-header class="header">

        <div class="header-left">

          <el-button link class="collapse-btn" @click="collapsed = !collapsed">

            <el-icon size="18"><Fold v-if="!collapsed" /><Expand v-else /></el-icon>

          </el-button>

          <div>

            <div class="page-title">{{ route.meta.title }}</div>

            <div v-if="route.meta.subtitle" class="page-subtitle">{{ route.meta.subtitle }}</div>

          </div>

        </div>

        <div class="user-box">

          <el-tag size="small" effect="plain">{{ roleText }}</el-tag>

          <el-dropdown trigger="click" @command="onUserCmd">

            <span class="user-trigger">

              <el-icon><UserFilled /></el-icon>

              {{ user.username }}

              <el-icon><ArrowDown /></el-icon>

            </span>

            <template #dropdown>

              <el-dropdown-menu>

                <el-dropdown-item command="logout">

                  <el-icon><SwitchButton /></el-icon>退出登录

                </el-dropdown-item>

              </el-dropdown-menu>

            </template>

          </el-dropdown>

        </div>

      </el-header>

      <el-main class="main">

        <div class="page-wrap">

          <router-view />

        </div>

      </el-main>

    </el-container>

  </el-container>

</template>



<script setup>

import { computed, ref } from 'vue'

import { useRoute, useRouter } from 'vue-router'

import http from '../api'

import { clearAuth, getUser } from '../utils/auth'



const route = useRoute()

const router = useRouter()

const user = getUser()

const collapsed = ref(false)



const activeMenu = computed(() => {
  const p = route.path
  if (p.startsWith('/platforms')) return '/platforms'
  if (p.startsWith('/settings')) return '/settings'
  return p
})



const roleText = computed(

  () => ({ super_admin: '超级管理员', finance: '财务', operator: '运营' })[user.role] || user.role

)



const roleMenu = {

  super_admin: ['dashboard', 'platforms', 'orders', 'stats', 'pay-api', 'settings'],

  finance: ['dashboard', 'orders', 'stats'],

  operator: ['dashboard', 'orders', 'pay-api', 'settings'],

}

const can = (key) => (roleMenu[user.role] || roleMenu.super_admin).includes(key)



async function onLogout() {

  try { await http.post('/admin/logout') } catch (e) { /* ignore */ }

  clearAuth()

  router.push('/login')

}



function onUserCmd(cmd) {

  if (cmd === 'logout') onLogout()

}

</script>



<style scoped>

.layout { height: 100%; }

.aside {

  background-color: var(--brand-primary);

  display: flex;

  flex-direction: column;

  transition: width 0.2s;

}

.logo {

  height: 64px;

  display: flex;

  align-items: center;

  justify-content: center;

  gap: 10px;

  color: #fff;

  border-bottom: 1px solid rgba(255, 255, 255, 0.08);

  padding: 0 12px;

}

.logo.collapsed { padding: 0; }

.logo-text { display: flex; flex-direction: column; line-height: 1.2; }

.logo-title { font-size: 15px; font-weight: 700; letter-spacing: 0.5px; }

.logo-sub { font-size: 11px; opacity: 0.72; }

.aside :deep(.el-menu) { border-right: none; flex: 1; }

.aside :deep(.el-menu-item.is-active) { background-color: var(--brand-primary-light); }

.aside-footer {

  padding: 12px;

  text-align: center;

  font-size: 11px;

  color: rgba(255, 255, 255, 0.4);

  border-top: 1px solid rgba(255, 255, 255, 0.08);

}

.header {

  display: flex;

  align-items: center;

  justify-content: space-between;

  border-bottom: 1px solid var(--card-border);

  background: #fff;

  height: 56px;

  padding: 0 20px;

}

.header-left { display: flex; align-items: center; gap: 8px; }

.collapse-btn { color: var(--text-secondary); }

.page-title { font-size: 16px; font-weight: 600; color: var(--text-primary); line-height: 1.3; }

.page-subtitle { font-size: 12px; color: var(--text-secondary); margin-top: 2px; }

.user-box { display: flex; align-items: center; gap: 12px; }

.user-trigger {

  display: inline-flex;

  align-items: center;

  gap: 4px;

  cursor: pointer;

  font-size: 14px;

  color: var(--text-primary);

}

.main { background: var(--page-bg); padding: var(--space-md); }

</style>
