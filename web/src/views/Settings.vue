<template>
  <el-card shadow="never" class="page-card">
    <el-tabs v-model="activeTab" @tab-change="onTabChange">
      <el-tab-pane v-if="showTab('pool')" label="码池入库" name="pool">
        <CodePoolIntake />
      </el-tab-pane>
      <el-tab-pane v-if="showTab('security')" label="IP 白名单" name="security">
        <Whitelist embedded />
      </el-tab-pane>
      <el-tab-pane v-if="showTab('stores')" label="门店档案" name="stores">
        <Stores embedded />
      </el-tab-pane>
    </el-tabs>
  </el-card>
</template>

<script setup>
import { ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { getUser } from '../utils/auth'
import CodePoolIntake from '../components/CodePoolIntake.vue'
import Stores from './Stores.vue'
import Whitelist from './Whitelist.vue'

const route = useRoute()
const router = useRouter()
const user = getUser()

const tabsForRole = {
  super_admin: ['pool', 'security', 'stores'],
  operator: ['stores'],
  finance: [],
}

const allowed = tabsForRole[user.role] || tabsForRole.super_admin
const defaultTab = route.query.tab === 'channels' ? 'pool' : (allowed.includes(route.query.tab) ? route.query.tab : allowed[0])
const activeTab = ref(defaultTab)

function showTab(name) {
  return allowed.includes(name)
}

function onTabChange(name) {
  router.replace({ query: { tab: name } })
}

watch(() => route.query.tab, (tab) => {
  const t = tab === 'channels' ? 'pool' : tab
  if (t && allowed.includes(t)) activeTab.value = t
})
</script>
