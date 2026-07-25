<template>
  <div class="login-stage">
    <div class="login-card">
      <div class="brand">
        <Insta360Logo />
        <h1>Insta360 达人营销洞察</h1>
        <p class="sub">{{ mode === 'login' ? '数据洞察平台 · 请登录后继续' : '创建账号 · 立即开启洞察' }}</p>
      </div>
      <div class="tabs">
        <button :class="['tab', { active: mode === 'login' }]" type="button" @click="mode = 'login'">登录</button>
        <button :class="['tab', { active: mode === 'register' }]" type="button" @click="mode = 'register'">注册</button>
      </div>
      <el-form @submit.prevent="onSubmit" label-position="top">
        <el-form-item label="账号">
          <el-input v-model="username" placeholder="请输入账号" size="large" />
        </el-form-item>
        <el-form-item label="密码">
          <el-input v-model="password" type="password" show-password placeholder="请输入密码(至少 6 位)" size="large" />
        </el-form-item>
        <el-form-item v-if="mode === 'register'" label="确认密码">
          <el-input v-model="confirm" type="password" show-password placeholder="请再次输入密码" size="large" @keyup.enter="onSubmit" />
        </el-form-item>
        <el-button type="primary" size="large" class="submit" :loading="loading" @click="onSubmit">
          {{ mode === 'login' ? '登录' : '注册' }}
        </el-button>
        <p v-if="err" class="err">{{ err }}</p>
      </el-form>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { useAuthStore } from '../stores/auth'
import Insta360Logo from '../components/Insta360Logo.vue'

const router = useRouter()
const auth = useAuthStore()
const mode = ref('login')
const username = ref('')
const password = ref('')
const confirm = ref('')
const loading = ref(false)
const err = ref('')

async function onSubmit() {
  err.value = ''
  if (!username.value || !password.value) {
    err.value = '请输入账号和密码'
    return
  }
  if (mode.value === 'register') {
    if (password.value.length < 6) {
      err.value = '密码至少 6 位'
      return
    }
    if (password.value !== confirm.value) {
      err.value = '两次输入的密码不一致'
      return
    }
  }
  loading.value = true
  try {
    if (mode.value === 'register') {
      await auth.register(username.value, password.value)
      ElMessage.success('注册成功，请登录')
      // 注册成功后自动切回登录 tab，并预填账号方便直接登录
      mode.value = 'login'
      confirm.value = ''
    } else {
      await auth.login(username.value, password.value)
      ElMessage.success('登录成功')
      router.push('/insight')
    }
  } catch (e) {
    err.value = (e && e.message) || (mode.value === 'register' ? '注册失败' : '登录失败')
  } finally {
    loading.value = false
  }
}
</script>

<style scoped lang="scss">
.login-stage {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: radial-gradient(1200px 600px at 50% -10%, #16203c 0%, #0b1020 60%);
}
.login-card {
  width: 380px;
  padding: 36px 32px;
  border-radius: 16px;
  background: rgba(19, 26, 48, 0.72);
  border: 1px solid rgba(61, 217, 235, 0.18);
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.45);
  backdrop-filter: blur(14px);
}
.brand { text-align: center; margin-bottom: 20px; }
.brand h1 { font-size: 18px; margin: 12px 0 4px; color: #eaf2ff; }
.brand .sub { font-size: 13px; color: #8fa3c8; margin: 0; }
.tabs {
  display: flex;
  gap: 8px;
  margin-bottom: 20px;
  padding: 4px;
  border-radius: 12px;
  background: rgba(11, 16, 32, 0.6);
  border: 1px solid rgba(61, 217, 235, 0.1);
}
.tab {
  flex: 1;
  padding: 10px 0;
  border: none;
  border-radius: 9px;
  background: transparent;
  color: #8fa3c8;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.25s cubic-bezier(0.16, 1, 0.3, 1);
}
.tab:hover { color: #eaf2ff; }
.tab.active {
  color: #0b1020;
  background: linear-gradient(135deg, #3dd9eb 0%, #5ea1ff 100%);
  box-shadow: 0 6px 18px rgba(61, 217, 235, 0.35);
}
.submit { width: 100%; margin-top: 8px; }
.err { color: #ff6b6b; font-size: 13px; text-align: center; margin: 12px 0 0; }
</style>
