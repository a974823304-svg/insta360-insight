# 前端用户注册界面 实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在登录页增加「登录 / 注册」Tab 切换与注册表单，让访客能注册账号并落 SQLite 用户库。

**Architecture:** 后端注册链路（阶段一）已完备且零改动；本次仅在前端三层补注册能力：API 封装 `request.register()` → store action `auth.register()` → LoginView 双 Tab 表单。注册成功切回登录 Tab 并预填账号，用户再登录进看板。

**Tech Stack:** Vue 3 + Vue Router 4 + Pinia + Element Plus + SCSS（前端）；Go/Gin + SQLite（后端，仅验证不动）。

## Global Constraints

- 项目**非 git 仓库**：所有 commit 步骤改为「若当前目录是 git 仓库则提交，否则跳过仅保留改动」，绝不擅自 `git init`。
- 前端无单元测试 harness（未配 vitest/jest），验证以 spec 的**手动浏览器/接口步骤**为准，不强行补单测。
- 后端注册接口契约：`POST /api/auth/register`，body `{username,password}`，成功 `{"code":0,"data":{"user":{...}}}`；用户名重复 `{"code":409,"message":"用户名已存在"}`；密码<6 位 `{"code":400,...}`。
- 后端 `role` 默认 `viewer`，前端不暴露角色选择。
- 品牌色：`--brand: #3DD9EB`、`--brand-2: #5EA1FF`。
- 路由 hash 模式，`/login` 为公开路由（已 `meta.public`）。
- 改 `LoginView.vue` 后**必须重启 Vite** 才生效。

---

### Task 1: 后端注册链路端到端验证（不改代码）

**Files:**
- 验证对象：`backend/internal/api/handler/auth.go`、`backend/internal/service/auth_service.go`、`backend/internal/store`（SQLite `users` 表）
- 不改任何文件

**Interfaces:**
- Consumes: 已运行的 Go 后端（`AUTH_DISABLE=1 SOURCE=mock`，阶段一约定）
- Produces: 确认 `/api/auth/register` 真落库，为后续前端注册提供可信后端

- [ ] **Step 1: 确认后端进程在跑**

若未运行，在 `backend/` 目录启动：
```bash
# 本机网络受限需放行环境; 纯 Go 无 CGO 可直接 build/run
AUTH_DISABLE=1 SOURCE=mock go run . 2>&1 | head -5
```
预期：监听 `:8080`，无编译错误。

- [ ] **Step 2: 注册新账号**

```bash
curl -s -X POST http://localhost:8080/api/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"username":"alice","password":"secret123"}'
```
预期：`{"code":0,"data":{"user":{...}}}`

- [ ] **Step 3: 重复注册同名**

```bash
curl -s -X POST http://localhost:8080/api/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"username":"alice","password":"secret123"}'
```
预期：`{"code":409,"message":"用户名已存在"}`

- [ ] **Step 4: 用新账号登录拿 token**

```bash
curl -s -X POST http://localhost:8080/api/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"alice","password":"secret123"}'
```
预期：`{"code":0,"data":{"token":"...","user":{...}}}`

- [ ] **Step 5: 收尾说明**

后端链路确认 OK。若任一步返回预期外结果（如编译失败、未落库），**单独记录并修复后端 bug**——但按阶段一实现，预期无需改。本项目非 git，不提交。

---

### Task 2: request.js 增加 register 封装

**Files:**
- Modify: `frontend/src/api/request.js`（在 `login()` 之后追加 `register()`）

**Interfaces:**
- Consumes: `request` 实例（`axios.create`，baseURL `/api`）
- Produces: `register(username, password)` 导出函数，供 `stores/auth.js` 调用

- [ ] **Step 1: 在 login 函数后追加 register**

将文件末尾 `export function login(...){...}` 段改为：

```js
// 登录:POST /api/auth/login,返回原始 APIResponse(含 code/data)
export function login(username, password) {
  return request.post('/auth/login', { username, password })
}

// 注册:POST /api/auth/register,返回原始 APIResponse(含 code/data)
export function register(username, password) {
  return request.post('/auth/register', { username, password })
}

export default request
```

- [ ] **Step 2: 构建校验（前端无测试 harness，以 build 验证编译）**

```bash
cd frontend && npx vite build --outDir .build-tmp 2>&1 | tail -3 && rm -rf .build-tmp
```
预期：`built in Xs`（无语法错误）。

- [ ] **Step 3: 提交（若适用）**

```bash
# 本项目非 git 仓库则跳过; 若是 git 仓库:
git add frontend/src/api/request.js
git commit -m "feat(api): 增加 register 注册接口封装"
```

---

### Task 3: auth store 增加 register action

**Files:**
- Modify: `frontend/src/stores/auth.js`（`import` 行 + `actions` 内新增）

**Interfaces:**
- Consumes: `register as apiRegister` from `../api/request`
- Produces: `auth.register(username, password)` → 返回 `res.data.user`；供 `LoginView.vue` 调用

- [ ] **Step 1: 修改 import**

将首部 `import { login as apiLogin } from '../api/request'` 改为：

```js
import { login as apiLogin, register as apiRegister } from '../api/request'
```

- [ ] **Step 2: 在 actions 内新增 register**

在 `login(...)` action 之后、`logout()` 之前插入：

```js
    async register(username, password) {
      const res = await apiRegister(username, password)
      if (!res || res.code !== 0) {
        throw new Error((res && res.message) || '注册失败')
      }
      return res.data.user
    },
```

- [ ] **Step 3: 构建校验**

```bash
cd frontend && npx vite build --outDir .build-tmp 2>&1 | tail -3 && rm -rf .build-tmp
```
预期：`built in Xs`。

- [ ] **Step 4: 提交（若适用）**

```bash
# 本项目非 git 仓库则跳过; 若是 git 仓库:
git add frontend/src/stores/auth.js
git commit -m "feat(auth): store 增加 register action"
```

---

### Task 4: LoginView 双 Tab + 注册表单

**Files:**
- Modify: `frontend/src/views/LoginView.vue`（template 加 tabs + 确认密码字段 + 提交按钮文案；script 加 `mode`/`confirm` + 注册分支；style 加 `.tabs`）

**Interfaces:**
- Consumes: `auth.register(username, password)`（Task 3）、`auth.login`（已有）、`ElMessage`、`useRouter`
- Produces: 登录页「登录/注册」切换 + 注册表单；注册成功切回登录 Tab 并预填账号

- [ ] **Step 1: 改 template**

将整个 `<template>` 段替换为：

```html
<template>
  <div class="login-stage">
    <div class="login-card">
      <div class="brand">
        <Insta360Logo />
        <h1>Insta360 达人营销洞察</h1>
        <p class="sub">数据洞察平台 · 请登录后继续</p>
      </div>

      <div class="tabs">
        <button type="button" :class="{ active: mode === 'login' }" @click="mode = 'login'">登录</button>
        <button type="button" :class="{ active: mode === 'register' }" @click="mode = 'register'">注册</button>
      </div>

      <el-form @submit.prevent="onSubmit" label-position="top">
        <el-form-item label="账号">
          <el-input v-model="username" placeholder="请输入账号" size="large" />
        </el-form-item>
        <el-form-item label="密码">
          <el-input v-model="password" type="password" show-password placeholder="请输入密码" size="large" @keyup.enter="onSubmit" />
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
```

- [ ] **Step 2: 改 script**

将 `<script setup>` 段替换为：

```js
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
  // 注册分支
  if (mode.value === 'register') {
    if (password.value !== confirm.value) {
      err.value = '两次输入的密码不一致'
      return
    }
    if (password.value.length < 6) {
      err.value = '密码至少 6 位'
      return
    }
    loading.value = true
    try {
      await auth.register(username.value, password.value)
      ElMessage.success('注册成功，请登录')
      mode.value = 'login'
      confirm.value = ''
      // username 保留, 方便直接登录
    } catch (e) {
      err.value = (e && e.message) || '注册失败'
    } finally {
      loading.value = false
    }
    return
  }
  // 登录分支
  loading.value = true
  try {
    await auth.login(username.value, password.value)
    ElMessage.success('登录成功')
    router.push('/insight')
  } catch (e) {
    err.value = (e && e.message) || '登录失败'
  } finally {
    loading.value = false
  }
}
</script>
```

- [ ] **Step 3: 加 .tabs 样式**

在 `<style scoped lang="scss">` 内、`.brand .sub` 规则之后、`.submit` 之前插入：

```scss
.tabs {
  display: flex;
  gap: 8px;
  margin-bottom: 20px;
  button {
    flex: 1;
    padding: 8px 0;
    border: none;
    border-radius: 8px;
    background: rgba(255, 255, 255, 0.04);
    color: #8fa3c8;
    font-size: 14px;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.2s ease;
    &.active {
      background: linear-gradient(135deg, var(--brand) 0%, var(--brand-2) 100%);
      color: #0B1020;
      box-shadow: 0 4px 14px rgba(61, 217, 235, 0.3);
    }
    &:hover:not(.active) { color: #eaf2ff; }
  }
}
```

- [ ] **Step 4: 构建校验**

```bash
cd frontend && npx vite build --outDir .build-tmp 2>&1 | tail -3 && rm -rf .build-tmp
```
预期：`built in Xs`，无错误。

- [ ] **Step 5: 提交（若适用）**

```bash
# 本项目非 git 仓库则跳过; 若是 git 仓库:
git add frontend/src/views/LoginView.vue
git commit -m "feat(login): 登录页增加注册 Tab 与注册表单"
```

---

### Task 5: 前端注册闭环手动验收

**Files:**
- 无文件改动，纯验收

**Interfaces:**
- Consumes: Task 1-4 全部改动 + 后端（Task 1 验证通过的注册链路）

- [ ] **Step 1: 重启 Vite（改了 LoginView）**

```bash
# frontend/ 下先停旧 dev server, 再
npm run dev
```

- [ ] **Step 2: 走查注册闭环**

1. 浏览器打开 `http://localhost:5173/`（或你实际端口），右上角「登录」按钮 → `/#/login`
2. 点「注册」Tab → 表单出现「确认密码」字段，提交按钮变「注册」
3. 故意不一致（密码 `aaa` / 确认 `bbb`）→ 行内提示「两次输入的密码不一致」，不发请求
4. 密码 <6 位（如 `abc`）→ 「密码至少 6 位」
5. 填正确（账号 `bob` / 密码 `secret123` / 确认 `secret123`）→ 提交
6. 预期：提示「注册成功，请登录」→ 自动切回「登录」Tab 且账号 `bob` 已预填
7. 点「登录」→ 进 `/insight`，右上角变头像首字母 `B`
8. 点头像 → 退出登录 → 回「登录」按钮（v15 闭环不受影响）

- [ ] **Step 3: 记录验收结论**

在 `F:\workbuddy\影石\.workbuddy\memory\2026-07-25.md` 追加验收结论（前端注册闭环通过 + 后端注册落库确认）。本项目非 git，无需提交源码。

---

## 自审小结（spec 覆盖）

- [x] 后端端到端验证（Task 1）— 对应 spec「后端注册链路端到端验证」
- [x] request.js register 封装（Task 2）— 对应 spec 改动 1
- [x] auth store register action（Task 3）— 对应 spec 改动 2
- [x] LoginView 双 Tab + 注册表单（Task 4）— 对应 spec 改动 3
- [x] 前端注册闭环验收（Task 5）— 对应 spec 验证
- [x] 无占位符 / 无 "TBD" / 代码块均为可直接粘贴内容
- [x] 类型一致：`register(username, password)` 在 request.js / auth.js / LoginView 三处签名一致；`auth.register` 返回 `res.data.user` 与 Task 1 后端返回结构一致
