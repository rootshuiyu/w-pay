<template>
  <div class="login-bg">
    <el-card class="login-card" shadow="always">
      <div class="login-brand">
        <el-icon size="36" color="#1d2b53"><Connection /></el-icon>
        <h1>聚合收银调度平台</h1>
        <p>多平台代收 · 统一收银 API · 订单对账</p>
      </div>
      <el-form :model="form" label-position="top" @keyup.enter="onLogin">
        <el-form-item label="用户名">
          <el-input v-model="form.username" placeholder="请输入用户名" size="large">
            <template #prefix><el-icon><User /></el-icon></template>
          </el-input>
        </el-form-item>
        <el-form-item label="密码">
          <el-input v-model="form.password" type="password" placeholder="请输入密码" size="large" show-password>
            <template #prefix><el-icon><Lock /></el-icon></template>
          </el-input>
        </el-form-item>
        <el-button type="primary" size="large" class="login-btn" :loading="loading" @click="onLogin">
          登 录
        </el-button>
      </el-form>
    </el-card>
  </div>
</template>

<script setup>
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import http from '../api'
import { setAuth } from '../utils/auth'

const router = useRouter()
const loading = ref(false)
const form = reactive({ username: '', password: '' })

async function onLogin() {
  if (!form.username || !form.password) {
    ElMessage.warning('请输入用户名和密码')
    return
  }
  loading.value = true
  try {
    const data = await http.post('/admin/login', form)
    setAuth(data.token, { username: data.username, role: data.role })
    ElMessage.success('登录成功')
    router.push('/dashboard')
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-bg {
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(145deg, #1d2b53 0%, #243566 45%, #14805e 100%);
}
.login-card {
  width: 420px;
  border-radius: 12px;
  padding: 8px 12px 16px;
  border: none;
}
.login-brand {
  text-align: center;
  margin-bottom: 28px;
}
.login-brand h1 {
  font-size: 20px;
  font-weight: 700;
  color: var(--brand-primary);
  margin: 12px 0 8px;
}
.login-brand p {
  color: var(--text-secondary);
  font-size: 13px;
  line-height: 1.5;
}
.login-btn { width: 100%; margin-top: 4px; }
</style>
