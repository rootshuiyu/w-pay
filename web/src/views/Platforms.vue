<template>
  <el-card shadow="never" class="page-card">
    <el-tabs v-model="activeTab" @tab-change="onTabChange">
      <el-tab-pane label="平台管理" name="list">
        <PlatformsPanel />
      </el-tab-pane>
      <el-tab-pane label="码池监控" name="pool" lazy>
        <ChannelPoolPanel ref="poolPanel" />
      </el-tab-pane>
    </el-tabs>
  </el-card>
</template>

<script setup>
import { ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import PlatformsPanel from '../components/PlatformsPanel.vue'
import ChannelPoolPanel from '../components/ChannelPoolPanel.vue'

const route = useRoute()
const router = useRouter()
const activeTab = ref(route.query.tab === 'pool' ? 'pool' : 'list')
const poolPanel = ref(null)

function onTabChange(name) {
  router.replace({ query: name === 'list' ? {} : { tab: name } })
  if (name === 'pool') poolPanel.value?.load()
}

watch(() => route.query.tab, (tab) => {
  activeTab.value = tab === 'pool' ? 'pool' : 'list'
})
</script>
