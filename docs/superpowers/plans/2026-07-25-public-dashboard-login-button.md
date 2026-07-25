# 数据公开 + 右上角登录按钮 实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 打开网址即可看到看板数据（未登录），右上角显示「登录」按钮，点击跳转登录页；登录后右上角变头像首字母，可下拉「退出登录」。

**Architecture:** 纯前端改动。路由守卫从「未登录重定向登录页」改为「看板公开、仅已登录访问登录页才弹回」；TopBar 根据 `auth.isLoggedIn` 在「登录按钮」与「头像+下拉」间切换；删除已无用的 `VITE_AUTH_DISABLE` 开关。

**Tech Stack:** Vue 3 + Vue Router 4 + Pinia + Element Plus（自动按需引入）+ SCSS。

## Global Constraints

- 项目**非 git 仓库**：所有 commit 步骤改为「若当前目录是 git 仓库则提交，否则跳过仅保留改动」，绝不擅自 `git init`。
- 改 `.env.*` 或 `router/index.js` 后**必须重启 Vite** 才生效（Vite 不热重载 env / 路由守卫逻辑）。
- 路由用 `createWebHashHistory()`，URL 形如 `/#/login`、`/#/insight`。
- Element Plus 组件（`el-dropdown` / `el-dropdown-menu` / `el-dropdown-item` / `el-icon`）已自动按需引入，无需手动 import；图标需从 `@element-plus/icons-vue` 显式导入。
- 前端无单元测试 harness（项目未配 vitest/jest），验证以 spec 的**手动浏览器步骤**为准，不强行补单测（YAGNI）。
- 品牌色变量：`--brand: #3DD9EB`、`--brand-2: #5EA1FF`、`--bg-base: #0B1020`。

---

### Task 1: 路由守卫改为看板公开

**Files:**
- Modify: `frontend/src/router/index.js`（删除 `AUTH_DISABLE` 判断，重写 `beforeEach`）

**Interfaces:**
- Consumes: `localStorage` 键 `insta_token`（由 `stores/auth.js` 写入/清除）
- Produces: 公开路由行为 —— `/insight` 等无需 token 即可访问；已登录访问 `/login` 重定向 `/insight`

- [ ] **Step 1: 重写路由守卫**

将文件末尾的守卫逻辑整体替换为以下代码（删掉 `const AUTH_DISABLE = ...` 整行及其分支）：

```js
// 路由配置
// 当前 Demo 只展示「数据洞察」页面, 预留其它 Tab 的扩展点
import { createRouter, createWebHashHistory } from 'vue-router'

const InsightDashboard = () => import('../views/InsightDashboard.vue')

const routes = [
  { path: '/login', component: () => import('../views/LoginView.vue'), meta: { public: true, title: '登录' } },
  { path: '/', redirect: '/insight' },
  { path: '/insight', component: InsightDashboard, meta: { title: '数据洞察' } },
  { path: '/creator', component: () => import('../views/PlaceholderView.vue'), meta: { title: '达人分析' } },
  { path: '/content', component: () => import('../views/PlaceholderView.vue'), meta: { title: '内容分析' } },
  { path: '/market', component: () => import('../views/PlaceholderView.vue'), meta: { title: '市场洞察' } },
  { path: '/brand', component: () => import('../views/PlaceholderView.vue'), meta: { title: '品牌分析' } },
  { path: '/custom', component: () => import('../views/PlaceholderView.vue'), meta: { title: '自定义看板' } }
]

const router = createRouter({
  history: createWebHashHistory(),
  routes
})

// 看板公开: 未登录也能访问; 仅已登录访问登录页时弹回看板(避免重复登录)
router.beforeEach((to) => {
  const token = localStorage.getItem('insta_token')
  const isPublic = to.meta && to.meta.public
  if (token && isPublic) {
    return '/insight'
  }
  return true
})

export default router
```

> 说明：保留 `routes` 数组原样（仅守卫段变化）。请确认文件里已无 `import.meta.env.VITE_AUTH_DISABLE` 引用。

- [ ] **Step 2: 重启 Vite 并验证守卫**

```bash
# 在 frontend/ 目录, 先停掉当前 dev server (Ctrl+C), 再
npm run dev
```

用浏览器 / Edge headless 验证：
- 打开 `http://localhost:5173/`（无 token）→ 直接渲染看板，**不跳登录页**（kpi-row 存在、login-card 不存在）
- 手动在地址栏访问 `http://localhost:5173/#/login` 且 `localStorage` 有 `insta_token` 时 → 应自动跳回 `/#/insight`
- 无 token 访问 `/#/login` → 正常显示登录页

- [ ] **Step 3: 提交（若适用）**

```bash
# 本项目非 git 仓库则跳过; 若是 git 仓库:
git add frontend/src/router/index.js
git commit -m "fix(router): 看板公开, 移除 AUTH_DISABLE 强制登录"
```

---

### Task 2: TopBar 登录按钮 ↔ 头像下拉

**Files:**
- Modify: `frontend/src/components/TopBar.vue`（template 右侧区 + script 导入/逻辑 + style 加 `.btn-login`）

**Interfaces:**
- Consumes: `useAuthStore`（`../stores/auth`，提供 `isLoggedIn` getter、`user.username`、`logout()`）；`useRouter`（`vue-router`）；`ElMessage`（`element-plus`）；图标 `User` / `SwitchButton`（`@element-plus/icons-vue`）
- Produces: 右上角未登录态「登录」按钮（跳转 `/login`）+ 已登录态头像首字母 + 下拉「退出登录」

- [ ] **Step 1: 替换 template 右侧操作区**

将 `.right` 容器内「日期选择器 + 旧导出按钮 + 头像」整段替换为：

```html
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
        <div class="avatar" :title="auth.user?.username || '账号'">{{ initial }}</div>
        <template #dropdown>
          <el-dropdown-menu>
            <el-dropdown-item disabled>{{ auth.user?.username || '账号' }}</el-dropdown-item>
            <el-dropdown-item command="logout" divided>
              <el-icon><SwitchButton /></el-icon>
              <span>退出登录</span>
            </el-dropdown-item>
          </el-dropdown-menu>
        </template>
      </el-dropdown>
    </div>
```

- [ ] **Step 2: 更新 script（导入 + 逻辑）**

将 `<script setup>` 整段替换为：

```js
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { Calendar, User, SwitchButton, DataLine, Film, Position, Tickets } from '@element-plus/icons-vue'
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
  const u = auth.user?.username || ''
  return u ? u.charAt(0).toUpperCase() : '?'
})

function goLogin() {
  router.push('/login')
}

function onCommand(cmd) {
  if (cmd === 'logout') {
    auth.logout()
    ElMessage.success('已退出登录')
  }
}
```

> 注意：删除了旧的 `Download` 图标导入与 `onExport` 函数（已在上一轮「删导出报告」中移除，此处确保无残留引用）。

- [ ] **Step 3: 新增 `.btn-login` 样式**

在 `<style lang="scss" scoped>` 中，于 `.avatar { ... }` 规则**之前**插入：

```scss
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
```

- [ ] **Step 4: 重启 Vite 并手动验证（Task 1 已重启则 HMR 即可）**

浏览器验证：
- 未登录刷新 `/` → 右上角出现「登录」按钮（青→蓝渐变），**无头像**
- 点「登录」→ URL 变 `/#/login`，登录页出现；输 `admin` / `insta360` → 回 `/#/insight`，右上角变头像 `A`
- 点头像 → 下拉显示 `admin` + 「退出登录」→ 点退出 → 回到「登录」按钮，页面不跳转
- 退出后当前页数据仍在（看板公开）

- [ ] **Step 5: 提交（若适用）**

```bash
# 本项目非 git 仓库则跳过; 若是 git 仓库:
git add frontend/src/components/TopBar.vue
git commit -m "feat(topbar): 未登录显示登录按钮, 已登录显示头像+登出下拉"
```

---

### Task 3: 清理 .env.development 死配置

**Files:**
- Modify: `frontend/.env.development`（删除 `VITE_AUTH_DISABLE` 行）

**Interfaces:**
- Consumes: 无（仅删配置）
- Produces: 前端不再引用 `VITE_AUTH_DISABLE`，避免误导

- [ ] **Step 1: 删除开关行**

将 `frontend/.env.development` 内容由：

```
VITE_AUTH_DISABLE=false
```

改为空文件（或仅保留注释说明看板公开）。推荐直接清空该文件内容。

- [ ] **Step 2: 全局搜索确认无残留引用**

```bash
# 在 frontend/ 目录
grep -rn "VITE_AUTH_DISABLE" src/ 2>/dev/null || echo "OK: 无残留"
```

Expected: 输出 `OK: 无残留`（Task 1 已移除 router 中的引用）。

- [ ] **Step 3: 提交（若适用）**

```bash
# 本项目非 git 仓库则跳过; 若是 git 仓库:
git add frontend/.env.development
git commit -m "chore(env): 移除已无用的 VITE_AUTH_DISABLE 开关"
```

---

### Task 4: 端到端手动验收

**Files:**
- 无文件改动，纯验收

**Interfaces:**
- Consumes: Task 1-3 的全部改动

- [ ] **Step 1: 完整走查（未登录 → 登录 → 登出）**

1. `npm run dev` 确保已重启（Task 1/2/3 任意改了 router/env 都需重启）
2. 浏览器打开 `http://localhost:5173/`
   - 断言：直接看到看板数据（KPI、图表、表格齐全），右上角「登录」按钮，无登录墙
3. 点「登录」→ `/#/login` → 输 `admin` / `insta360` → 登录成功跳回 `/#/insight`
   - 断言：右上角变为头像 `A`
4. 点头像 → 下拉 → 点「退出登录」
   - 断言：回到「登录」按钮，当前页仍在（数据不丢）
5. 已登录状态下手动访问 `/#/login`
   - 断言：自动跳回 `/#/insight`

- [ ] **Step 2: 记录验收结论**

在 `F:\workbuddy\影石\.workbuddy\memory\2026-07-25.md` 追加验收结论（公开看板 + 登录按钮链路闭环通过）。本项目非 git，无需 commit 源码。

---

## 自审小结（spec 覆盖）

- [x] 路由公开（Task 1）— 对应 spec 改动 1
- [x] TopBar 登录按钮 + 头像下拉（Task 2）— 对应 spec 改动 2
- [x] 删除 VITE_AUTH_DISABLE（Task 3）— 对应 spec 改动 3
- [x] LoginView 不改（spec 明确，无需任务）
- [x] 验证路径（Task 4）— 对应 spec 验证
- [x] 无占位符 / 无 "TBD" / 代码块均为可直接粘贴内容
- [x] 类型一致：`auth.isLoggedIn` / `auth.user.username` / `auth.logout()` 与 `stores/auth.js` 现有定义一致；`goLogin` / `onCommand` / `initial` 命名跨 template/script 一致
