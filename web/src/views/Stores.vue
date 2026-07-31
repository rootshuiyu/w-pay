<template>
  <component :is="embedded ? 'div' : 'el-card'" shadow="never" :class="embedded ? '' : 'page-card'">
    <el-alert
      type="info"
      :closable="false"
      show-icon
      title="门店为系统内部档案。商户码录入请使用「码池入库」Tab，无需在此重复创建。"
      style="margin-bottom: 16px"
    />
    <div class="page-toolbar">
      <el-input v-model="query.keyword" placeholder="门店名称/编号" clearable style="width: 200px" @keyup.enter="load" />
      <el-select v-model="query.status" placeholder="状态" clearable style="width: 110px">
        <el-option label="正常" :value="1" />
        <el-option label="停用" :value="0" />
      </el-select>
      <el-checkbox v-model="query.hideSystem" @change="load">隐藏系统门店</el-checkbox>
      <el-button type="primary" @click="load"><el-icon><Search /></el-icon>查询</el-button>
      <div class="page-toolbar-spacer" />
      <el-button type="success" @click="openAdd"><el-icon><Plus /></el-icon>新增门店</el-button>
    </div>

    <el-table :data="rows" v-loading="loading" stripe>
      <el-table-column prop="store_name" label="门店名称" min-width="160" />
      <el-table-column prop="store_code" label="编号" width="100" />
      <el-table-column prop="tax_subject" label="收款主体" min-width="140" show-overflow-tooltip />
      <el-table-column prop="address" label="地址" min-width="140" show-overflow-tooltip />
      <el-table-column label="状态" width="90">
        <template #default="{ row }">
          <el-switch
            :model-value="row.status === 1"
            @change="(v) => toggleStatus(row, v)"
          />
        </template>
      </el-table-column>
      <el-table-column prop="created_at" label="创建时间" width="170">
        <template #default="{ row }">{{ fmtTime(row.created_at) }}</template>
      </el-table-column>
      <el-table-column label="操作" width="140" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" @click="openEdit(row)">编辑</el-button>
          <el-popconfirm title="确认删除？关联商户码将失效" @confirm="del(row)">
            <template #reference>
              <el-button link type="danger">删除</el-button>
            </template>
          </el-popconfirm>
        </template>
      </el-table-column>
    </el-table>

    <el-pagination
      class="page-pager"
      v-model:current-page="query.page"
      v-model:page-size="query.page_size"
      :total="total"
      layout="total, sizes, prev, pager, next"
      :page-sizes="[10, 20, 50, 100]"
      @change="load"
    />
  </component>

  <el-dialog v-model="dialog.visible" :title="dialog.isEdit ? '编辑门店' : '新增门店'" width="520px">
    <el-form :model="form" label-width="100px">
      <el-form-item label="门店名称" required>
        <el-input v-model="form.store_name" placeholder="如：人民路店" />
      </el-form-item>
      <el-form-item label="业务编号" v-if="!dialog.isEdit">
        <el-input v-model="form.store_code" placeholder="可选，如 RM001" />
      </el-form-item>
      <el-form-item label="经营地址">
        <el-input v-model="form.address" />
      </el-form-item>
      <el-form-item label="个体户主体">
        <el-input v-model="form.tax_subject" placeholder="如：个体户：张三" />
      </el-form-item>
      <el-form-item label="备注">
        <el-input v-model="form.remark" type="textarea" :rows="2" />
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="dialog.visible = false">取消</el-button>
      <el-button type="primary" :loading="dialog.saving" @click="save">保存</el-button>
    </template>
  </el-dialog>
</template>

<script setup>
import { onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import http from '../api'

defineProps({ embedded: { type: Boolean, default: false } })

const router = useRouter()
const rows = ref([])
const total = ref(0)
const loading = ref(false)
const query = reactive({ keyword: '', status: null, hideSystem: true, page: 1, page_size: 20 })
const dialog = reactive({ visible: false, isEdit: false, saving: false })
const form = reactive({ id: '', store_name: '', store_code: '', address: '', tax_subject: '', remark: '' })

const fmtTime = (t) => (t ? String(t).replace('T', ' ').slice(0, 19) : '')

async function load() {
  loading.value = true
  try {
    const params = { page: query.page, page_size: query.page_size, hide_system: query.hideSystem ? 1 : 0 }
    if (query.keyword) params.keyword = query.keyword
    if (query.status !== null && query.status !== '') params.status = query.status
    const data = await http.get('/admin/store/list', { params })
    rows.value = data.list || []
    total.value = data.total || 0
  } finally {
    loading.value = false
  }
}

function openAdd() {
  dialog.isEdit = false
  Object.assign(form, { id: '', store_name: '', store_code: '', address: '', tax_subject: '', remark: '' })
  dialog.visible = true
}

function openEdit(row) {
  dialog.isEdit = true
  Object.assign(form, {
    id: row.id,
    store_name: row.store_name,
    store_code: row.store_code,
    address: row.address,
    tax_subject: row.tax_subject,
    remark: row.remark,
  })
  dialog.visible = true
}

async function save() {
  if (!form.store_name) {
    ElMessage.warning('请填写门店名称')
    return
  }
  dialog.saving = true
  try {
    if (dialog.isEdit) {
      await http.put('/admin/store/edit', {
        id: form.id,
        store_name: form.store_name,
        address: form.address,
        tax_subject: form.tax_subject,
        remark: form.remark,
      })
      ElMessage.success('已保存')
    } else {
      const data = await http.post('/admin/store/add', {
        store_name: form.store_name,
        store_code: form.store_code,
        address: form.address,
        tax_subject: form.tax_subject,
        remark: form.remark,
      })
      ElMessage.success(`门店已创建，ID：${data.id}`)
    }
    dialog.visible = false
    load()
  } finally {
    dialog.saving = false
  }
}

async function toggleStatus(row, enabled) {
  await http.put('/admin/store/status', { id: row.id, status: enabled ? 1 : 0 })
  row.status = enabled ? 1 : 0
  ElMessage.success(enabled ? '已启用' : '已停用')
}

async function del(row) {
  await http.delete('/admin/store/del', { data: { id: row.id } })
  ElMessage.success('已删除')
  load()
}

function goChannels(row) {
  router.push({ path: '/settings', query: { tab: 'pool' } })
}

function copy(text) {
  navigator.clipboard?.writeText(String(text))
  ElMessage.success('已复制')
}

onMounted(load)
</script>

<style scoped>
.mono { font-family: Consolas, monospace; font-size: 13px; }
</style>
