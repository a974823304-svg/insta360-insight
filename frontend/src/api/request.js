import axios from 'axios'

// 统一网络请求封装：请求拦截器注入 Token，响应拦截器统一错误处理
const request = axios.create({
  baseURL: '/api',
  timeout: 10000
})

request.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('insta_token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => Promise.reject(error)
)

request.interceptors.response.use(
  (response) => {
    const data = response.data
    // 关键防护: 静态托管 / 后端未启动时, 网关会把 SPA 的 index.html
    // (HTTP 200, Content-Type: text/html) 当作接口响应返回。
    // axios 对 text/html 不会做 JSON 解析, 于是 data 是一大段 HTML 字符串。
    // 若直接交给 store, v-for 会把字符串按字符展开 -> 成百上千个空卡片。
    // 这里把"字符串响应"一律视为失败, 让上层回退到前端内嵌 fallback 数据。
    if (typeof data === 'string') {
      return Promise.reject(new Error('接口返回了非 JSON 内容(疑似静态托管兜底页), 已回退本地数据'))
    }
    return data
  },
  (error) => {
    // HTTP 401(后端 JWTAuth 拦截) -> 清 token 并跳登录页
    if (error.response && error.response.status === 401) {
      localStorage.removeItem('insta_token')
      localStorage.removeItem('insta_user')
      if (window.location.hash !== '#/login') {
        window.location.hash = '#/login'
      }
    }
    console.error('[request error]', error)
    return Promise.reject(error)
  }
)

// 登录:POST /api/auth/login,返回原始 APIResponse(含 code/data)
export function login(username, password) {
  return request.post('/auth/login', { username, password })
}

// 注册:POST /api/auth/register,返回原始 APIResponse(含 code/data)
export function register(username, password) {
  return request.post('/auth/register', { username, password })
}

// 获取当前用户资料:GET /api/user/profile
export function getProfile() {
  return request.get('/user/profile')
}

// 更新资料:PUT /api/user/profile
export function updateProfile(payload) {
  return request.put('/user/profile', payload)
}

// 上传头像:POST /api/user/avatar (multipart),返回原始 APIResponse(含 code/data)
export function uploadAvatar(file) {
  const fd = new FormData()
  fd.append('file', file)
  return request.post('/user/avatar', fd, {
    headers: { 'Content-Type': 'multipart/form-data' }
  })
}

export default request
