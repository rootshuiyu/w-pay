import axios from 'axios'
import { ElMessage } from 'element-plus'
import router from './router'
import { clearAuth, getToken } from './utils/auth'

const http = axios.create({ baseURL: '/api', timeout: 15000 })

http.interceptors.request.use((cfg) => {
  const token = getToken()
  if (token) cfg.headers.Authorization = `Bearer ${token}`
  return cfg
})

function toLogin() {
  clearAuth()
  router.push('/login')
}

http.interceptors.response.use(
  (resp) => {
    if (resp.config.responseType === 'blob') return resp
    const { code, message, data } = resp.data
    if (code === 200) return data
    if (code === 401) toLogin()
    ElMessage.error(message || '请求失败')
    return Promise.reject(new Error(message || '请求失败'))
  },
  (err) => {
    if (err.response?.status === 401) toLogin()
    const msg = err.response?.data?.message || err.message || '网络错误'
    ElMessage.error(msg)
    return Promise.reject(err)
  }
)

// 下载 Excel（携带 token），自动取文件名
export async function download(url, params) {
  const resp = await http.get(url, { params, responseType: 'blob' })
  const dispo = resp.headers['content-disposition'] || ''
  const match = dispo.match(/filename=([^;]+)/)
  const filename = match ? match[1].trim() : 'export.xlsx'
  const blobUrl = URL.createObjectURL(resp.data)
  const a = document.createElement('a')
  a.href = blobUrl
  a.download = filename
  a.click()
  URL.revokeObjectURL(blobUrl)
}

export const payTypeText = (t) => (t === 1 ? '微信' : t === 2 ? '支付宝' : '未知')
export const orderStatusText = (s) =>
  ({ 0: '待支付', 1: '已支付', 2: '已关闭', 3: '退款' })[s] ?? '未知'
export const fen2yuan = (v) => ((v ?? 0) / 100).toFixed(2)

export default http
