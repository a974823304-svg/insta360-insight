<template>
  <!--
    顶部栏 (v3.png 风格)
    --------------------
    左: Insta360 真品牌标(彩色球 + 字标)
    中: 5 个导航 tab(数据洞察 / 达人分析 / 内容分析 / 市场洞察 / 自定义看板), 激活态有底部青色细线
    右: 时间范围 + 头像
    高度紧凑(52px), 让首屏能装下完整看板
  -->
  <header class="topbar">
    <!-- 品牌标识 -->
    <div class="brand">
      <Insta360Logo :size="118" />
    </div>

    <!-- 中部导航 -->
    <nav class="tabs">
      <router-link
        v-for="t in tabs"
        :key="t.path"
        :to="t.path"
        class="tab"
        :class="{ active: route.path === t.path }"
      >
        <el-icon class="tab-icon"><component :is="t.icon" /></el-icon>
        <span class="label">{{ t.label }}</span>
      </router-link>
    </nav>

    <!-- 右侧操作区 -->
    <div class="right">
      <el-date-picker
        v-model="filter.dateRange"
        type="daterange"
        range-separator="→"
        start-placeholder="开始"
        end-placeholder="结束"
        format="YYYY/MM/DD"
        value-format="YYYY-MM-DD"
        size="small"
        class="date"
      >
        <template #prefix>
          <el-icon><Calendar /></el-icon>
        </template>
      </el-date-picker>

      <!-- 未登录: 登录按钮 -->
      <button v-if="!auth.isLoggedIn" class="btn-login" @click="goLogin">
        <el-icon><User /></el-icon>
        <span>登录</span>
      </button>

      <!-- 已登录: 头像 + 下拉 -->
      <el-dropdown v-else trigger="click" class="avatar-drop" @command="onCommand">
        <div class="avatar" :class="{ 'has-img': avatarUrl }" :style="avatarStyle" :title="auth.user?.username || '账号'">
          <img v-if="avatarUrl" :src="avatarUrl" alt="avatar" />
          <span v-else>{{ initial }}</span>
        </div>
        <template #dropdown>
          <el-dropdown-menu>
            <el-dropdown-item disabled>{{ auth.user?.username || '账号' }}</el-dropdown-item>
            <el-dropdown-item command="profile">
              <el-icon><EditPen /></el-icon>
              <span>个人资料</span>
            </el-dropdown-item>
            <el-dropdown-item command="logout" divided>
              <el-icon><SwitchButton /></el-icon>
              <span>退出登录</span>
            </el-dropdown-item>
          </el-dropdown-menu>
        </template>
      </el-dropdown>
    </div>
  </header>
</template>

<script setup>
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { Calendar, User, SwitchButton, EditPen, DataLine, Film, Position, Tickets } from '@element-plus/icons-vue'
import { useFilterStore } from '../stores/filter'
import { useAuthStore } from '../stores/auth'
import { ElMessage } from 'element-plus'
import Insta360Logo from './Insta360Logo.vue'

const route = useRoute()
const router = useRouter()
const filter = useFilterStore()
const auth = useAuthStore()

// 顶栏 Tab(与路由对应, 当前只接 /insight, 其余占位跳转占位路由)
const tabs = [
  { path: '/insight', label: '数据洞察', icon: DataLine },
  { path: '/creator', label: '达人分析', icon: User },
  { path: '/content', label: '内容分析', icon: Film },
  { path: '/market',  label: '市场洞察', icon: Position },
  { path: '/custom',  label: '自定义看板', icon: Tickets }
]

// 头像首字母(用户名首字大写; 无则 ?)
const initial = computed(() => {
  const u = auth.user?.nickname || auth.user?.username || ''
  return u ? u.charAt(0).toUpperCase() : '?'
})

// 预设头像渐变映射(preset:xxx → 背景渐变)
const presetMap = {
  'preset:blue': 'linear-gradient(135deg,#5EA1FF 0%,#A07BFF 100%)',
  'preset:cyan': 'linear-gradient(135deg,#3DD9EB 0%,#5EA1FF 100%)',
  'preset:sun': 'linear-gradient(135deg,#FFB547 0%,#FF6B6B 100%)'
}
// 头像图片地址:识别 http(s) / /avatars/ / data: 三种形式
const avatarUrl = computed(() => {
  const a = auth.user?.avatar || ''
  if (/^https?:\/\//.test(a) || a.startsWith('/avatars/') || a.startsWith('data:')) return a
  return ''
})
// 头像背景:预设 → 渐变;URL/空 → 透明(由默认或图片覆盖)
const avatarStyle = computed(() => {
  const a = auth.user?.avatar || ''
  if (presetMap[a]) return { background: presetMap[a] }
  return {}
})

function goLogin() {
  router.push('/login')
}

function onCommand(cmd) {
  if (cmd === 'profile') {
    router.push('/profile')
  } else if (cmd === 'logout') {
    auth.logout()
    ElMessage.success('已退出登录')
  }
}
</script>

<style lang="scss" scoped>
.topbar {
  display: flex;
  align-items: center;
  height: 52px;                       // 紧凑: 56 → 52
  padding: 0 18px;
  background: rgba(11, 16, 32, 0.78);
  backdrop-filter: var(--glass-blur);
  -webkit-backdrop-filter: var(--glass-blur);
  border-bottom: 1px solid var(--border);
  z-index: 10;
  flex-shrink: 0;
}

.brand {
  margin-right: 28px;
  display: flex;
  align-items: center;
}

.tabs {
  display: flex;
  gap: 2px;
  flex: 1;
  height: 100%;
}
.tab {
  position: relative;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 0 14px;
  height: 100%;
  color: var(--text-secondary);
  font-size: 13px;
  text-decoration: none;
  transition: color 0.2s ease;
  .tab-icon { font-size: 14px; }
  &:hover { color: var(--text-primary); }
  &.active {
    color: var(--brand);
    font-weight: 600;
    // 底部青色细线(对齐 v3.png 的 tab 强调样式)
    &::after {
      content: '';
      position: absolute;
      left: 12px;
      right: 12px;
      bottom: 8px;
      height: 2px;
      border-radius: 2px;
      background: linear-gradient(90deg, var(--brand), var(--brand-2));
    }
  }
}

.right {
  display: flex;
  align-items: center;
  gap: 10px;
}
.date {
  width: 240px;
  --el-text-color-regular: var(--text-primary);
  --el-text-color-placeholder: var(--text-muted);
  --el-border-color: var(--border);
  --el-fill-color-blank: rgba(19, 26, 48, 0.6);
}
.btn-login {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 6px 14px;
  border-radius: 8px;
  background: linear-gradient(135deg, var(--brand) 0%, var(--brand-2) 100%);
  color: #0B1020;
  font-weight: 600;
  font-size: 13px;
  border: none;
  cursor: pointer;
  box-shadow: 0 4px 14px rgba(61, 217, 235, 0.32);
  transition: transform 0.2s ease, box-shadow 0.2s ease;
  .el-icon { font-size: 13px; }
  &:hover {
    transform: translateY(-1px);
    box-shadow: 0 6px 22px rgba(61, 217, 235, 0.45);
  }
}
.avatar-drop { display: inline-flex; }

.avatar {
  width: 30px;
  height: 30px;
  border-radius: 50%;
  background: linear-gradient(135deg, #5EA1FF 0%, #A07BFF 100%);
  color: #fff;
  display: grid;
  place-items: center;
  font-weight: 700;
  font-size: 12px;
  cursor: pointer;
  border: 2px solid var(--bg-base);
}
.avatar img {
  width: 100%;
  height: 100%;
  border-radius: 50%;
  object-fit: cover;
  display: block;
}
</style>