<template>
  <component :is="embedded ? 'div' : 'el-card'" shadow="never">
    <el-alert
      v-if="!embedded"
      type="info"
      :closable="false"
      show-icon
      title="配置后台管理、收银 API、支付回调的来源 IP 限制，以及反向代理信任列表。保存后立即生效，无需重启。"
      style="margin-bottom: 14px"
    />

    <el-tabs v-model="activeScope" @tab-change="load">
      <el-tab-pane label="后台管理 IP" name="admin" />
      <el-tab-pane label="收银 API IP" name="pay" />
      <el-tab-pane label="支付回调 IP" name="callback" />
      <el-tab-pane label="可信反向代理" name="trusted_proxy" />
    </el-tabs>

    <p class="hint">{{ scopeHint }}</p>

    <div class="toolbar">
      <span class="policy-label">启用限制</span>
      <el-switch :model-value="currentPolicy.enabled === 1" @change="togglePolicy" />
      <span class="policy-tip">{{ currentPolicy.enabled === 1 ? '已启用：仅允许列表内来源' : '未启用：不限制来源' }}</span>
      <div style="flex: 1" />
      <el-button @click="load">刷新</el-button>
      <el-button type="success" @click="openAdd"><el-icon><Plus /></el-icon>新增条目</el-button>
    </div>

    <el-table :data="filteredRows" v-loading="loading" stripe>
      <el-table-column prop="cidr" label="IP / CIDR" min-width="180" />
      <el-table-column prop="remark" label="备注" min-width="160" show-overflow-tooltip />
      <el-table-column label="状态" width="90">
        <template #default="{ row }">
          <el-switch :model-value="row.status === 1" @change="(v) => toggleRow(row, v)" />
        </template>
      </el-table-column>
      <el-table-column label="操作" width="140" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" @click="openEdit(row)">编辑</el-button>
          <el-popconfirm title="确认删除该条目？" @confirm="del(row)">
            <template #reference>
              <el-button link type="danger">删除</el-button>
            </template>
          </el-popconfirm>
        </template>
      </el-table-column>
    </el-table>
  </component>

  <el-dialog v-model="dialog.visible" :title="dialog.isEdit ? '编辑备注' : '新增 IP/CIDR'" width="520px">
    <el-form :model="form" label-width="100px">
      <el-form-item label="作用域">
        <el-tag>{{ scopeLabel(activeScope) }}</el-tag>
      </el-form-item>
      <el-form-item v-if="!dialog.isEdit" label="IP / CIDR" required>
        <el-input v-model="form.cidr" placeholder="如 10.0.0.0/8 或 203.0.113.5" />
      </el-form-item>
      <el-form-item v-else label="IP / CIDR">
        <el-input v-model="form.cidr" disabled />
      </el-form-item>
      <el-form-item label="备注">
        <el-input v-model="form.remark" placeholder="如：门店收银机 / 微信回调" />
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="dialog.visible = false">取消</el-button>
      <el-button type="primary" :loading="dialog.saving" @click="save">保存</el-button>
    </template>
  </el-dialog>
</template>

<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import http from '../api'

defineProps({ embedded: { type: Boolean, default: false } })

const activeScope = ref('admin')
const loading = ref(false)
const policies = ref([])
const rows = ref([])
const dialog = reactive({ visible: false, isEdit: false, saving: false })
const form = reactive({ id: '', cidr: '', remark: '' })

const hints = {
  admin: '限制 /api/admin/* 访问来源（登录与白名单管理页不受限，便于误配后自救）',
  pay: '限制 /api/pay/create、/api/pay/query 的调用来源；添加门店收银系统或对接站点的出口 IP，启用后仅列表内 IP 可下单/查单',
  callback: '限制 /api/notify/* 仅放行微信/支付宝官方回调 IP（支付平台主动通知本系统）',
  trusted_proxy: '启用后信任下列代理 IP 的 X-Forwarded-For（同机 Nginx 填 127.0.0.1）',
}
const scopeLabels = {
  admin: '后台管理',
  pay: '收银 API',
  callback: '支付回调',
  trusted_proxy: '可信反向代理',
}

const scopeHint = computed(() => hints[activeScope.value] || '')
const scopeLabel = (s) => scopeLabels[s] || s
const filteredRows = computed(() => rows.value.filter((r) => r.scope === activeScope.value))
const currentPolicy = computed(() => policies.value.find((p) => p.scope === activeScope.value) || { enabled: 0 })

async function load() {
  loading.value = true
  try {
    const data = await http.get('/admin/whitelist/overview')
    policies.value = data.policies || []
    rows.value = data.entries || []
  } finally {
    loading.value = false
  }
}

async function togglePolicy(enabled) {
  await http.put('/admin/whitelist/policy', { scope: activeScope.value, enabled: enabled ? 1 : 0 })
  ElMessage.success(enabled ? '已启用限制' : '已关闭限制')
  await load()
}

async function toggleRow(row, enabled) {
  await http.put('/admin/whitelist/status', { id: row.id, status: enabled ? 1 : 0 })
  ElMessage.success('状态已更新')
  await load()
}

function openAdd() {
  dialog.isEdit = false
  Object.assign(form, { id: '', cidr: '', remark: '' })
  dialog.visible = true
}

function openEdit(row) {
  dialog.isEdit = true
  Object.assign(form, { id: row.id, cidr: row.cidr, remark: row.remark || '' })
  dialog.visible = true
}

async function save() {
  if (!dialog.isEdit && !form.cidr.trim()) {
    ElMessage.warning('请填写 IP 或 CIDR')
    return
  }
  dialog.saving = true
  try {
    if (dialog.isEdit) {
      await http.put('/admin/whitelist/edit', { id: form.id, remark: form.remark })
    } else {
      await http.post('/admin/whitelist/add', {
        scope: activeScope.value,
        cidr: form.cidr.trim(),
        remark: form.remark,
      })
    }
    ElMessage.success('保存成功')
    dialog.visible = false
    await load()
  } finally {
    dialog.saving = false
  }
}

async function del(row) {
  await http.delete('/admin/whitelist/del', { data: { id: row.id } })
  ElMessage.success('已删除')
  await load()
}

onMounted(load)
</script>

<style scoped>
.hint { color: #666; font-size: 13px; margin: 0 0 12px; }
.toolbar {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 14px;
}
.policy-label { font-size: 14px; color: #333; }
.policy-tip { font-size: 13px; color: #888; }
</style>
