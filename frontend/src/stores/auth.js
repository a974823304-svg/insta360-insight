import { defineStore } from 'pinia'
import { login as apiLogin, register as apiRegister, getProfile as apiGetProfile, updateProfile as apiUpdateProfile } from '../api/request'

const TOKEN_KEY = 'insta_token'
const USER_KEY = 'insta_user'

function loadToken() {
  return localStorage.getItem(TOKEN_KEY) || ''
}
function loadUser() {
  const raw = localStorage.getItem(USER_KEY)
  if (!raw) return null
  try { return JSON.parse(raw) } catch { return null }
}

export const useAuthStore = defineStore('auth', {
  state: () => ({
    token: loadToken(),
    user: loadUser()
  }),
  getters: {
    isLoggedIn: (s) => !!s.token
  },
  actions: {
    async login(username, password) {
      const res = await apiLogin(username, password)
      if (!res || res.code !== 0) {
        throw new Error((res && res.message) || '登录失败')
      }
      this.token = res.data.token
      this.user = res.data.user
      localStorage.setItem(TOKEN_KEY, this.token)
      localStorage.setItem(USER_KEY, JSON.stringify(this.user))
      await this.fetchProfile() // 补全 nickname/avatar/contact/bio(登录响应可能未带全)
    },
    async register(username, password) {
      const res = await apiRegister(username, password)
      if (!res || res.code !== 0) {
        throw new Error((res && res.message) || '注册失败')
      }
      // 注册成功仅返回 user(无 token),需用户再登录;防御性处理 res.data 为 null
      return res.data ? res.data.user : null
    },
    async fetchProfile() {
      // 资料失败不阻塞主流程,保留已有 user
      try {
        const res = await apiGetProfile()
        if (res && res.code === 0 && res.data) {
          this.user = { ...this.user, ...res.data }
          localStorage.setItem(USER_KEY, JSON.stringify(this.user))
        }
      } catch (e) {
        /* 忽略:无 token / 网络异常时静默 */
      }
    },
    async saveProfile(payload) {
      const res = await apiUpdateProfile(payload)
      if (!res || res.code !== 0) {
        throw new Error((res && res.message) || '保存失败')
      }
      this.user = { ...this.user, ...(res.data || {}) }
      localStorage.setItem(USER_KEY, JSON.stringify(this.user))
      return this.user
    },
    logout() {
      this.token = ''
      this.user = null
      localStorage.removeItem(TOKEN_KEY)
      localStorage.removeItem(USER_KEY)
    }
  }
})
