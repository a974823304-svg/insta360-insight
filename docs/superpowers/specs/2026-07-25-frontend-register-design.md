# 设计：前端用户注册界面（登录页 Tab 切换）

- 日期：2026-07-25
- 背景：用户希望「接入数据库 + 登录用户注册」。经排查，后端注册链路（阶段一）已完整实现：`POST /api/auth/register` 路由、service（bcrypt 哈希 + 唯一约束处理）、SQLite 用户表、种子 admin 均就绪。本次**仅需补齐前端注册界面**，并在动手前实测后端注册确实落库。
- 范围：纯前端（LoginView + request.js + auth store）。数据库保持 SQLite（阶段一已搭，纯 Go 无 CGO，本机可编译）。后端代码零改动，仅做端到端验证。

## 目标

1. 登录页提供「登录 / 注册」Tab 切换。
2. 注册表单：账号 + 密码 + 确认密码；前端校验一致性 + 密码≥6 位。
3. 提交后调用 `POST /api/auth/register`，成功后切回登录 Tab 并预填账号，用户再点登录进入看板。
4. 验证后端注册链路端到端落 SQLite（此前未实测）。

## 非目标（YAGNI）

- 不改后端（endpoint / service / store 已完备）。
- 不引入 email、验证码、角色选择（后端 `role` 默认 `viewer`，前端不暴露）。
- 不做「注册后自动登录」（改为切回登录 Tab 预填账号，行为可预测、易调试）。
- 不接真实 OLAP 数据库（insight 数据仍走 mock，与本次注册无关）。

## 改动清单

### 1. 前端注册 API 封装
文件：`frontend/src/api/request.js`

- 在 `login()` 旁新增：
  ```js
  // 注册:POST /api/auth/register,返回原始 APIResponse(含 code/data)
  export function register(username, password) {
    return request.post('/auth/register', { username, password })
  }
  ```

### 2. 前端 auth store 加 register action
文件：`frontend/src/stores/auth.js`

- `import` 改为：
  ```js
  import { login as apiLogin, register as apiRegister } from '../api/request'
  ```
- 在 `actions` 内新增（与 `login` 同错误处理风格）：
  ```js
  async register(username, password) {
    const res = await apiRegister(username, password)
    if (!res || res.code !== 0) {
      throw new Error((res && res.message) || '注册失败')
    }
    return res.data.user
  }
  ```

### 3. 登录页 Tab 切换 + 注册表单
文件：`frontend/src/views/LoginView.vue`

- `script` 新增：
  - `const mode = ref('login')`（'login' | 'register'）
  - 注册用 ref：`const confirm = ref('')`
  - 复用现有 `username` / `password`
  - `onSubmit()` 按 mode 分流：
    ```js
    async function onSubmit() {
      err.value = ''
      if (!username.value || !password.value) {
        err.value = '请输入账号和密码'
        return
      }
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
      // 登录分支(原逻辑)
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
    ```
- `template` 调整：
  - 在 `.login-card` 内、`.brand` 下方加 Tab 切换（青色高亮，对齐 v3 风格）：
    ```html
    <div class="tabs">
      <button :class="{ active: mode === 'login' }" @click="mode = 'login'">登录</button>
      <button :class="{ active: mode === 'register' }" @click="mode = 'register'">注册</button>
    </div>
    ```
  - 表单内：注册模式下多一个「确认密码」`el-form-item`（用 `v-if="mode === 'register'"`），`@keyup.enter="onSubmit"` 保留
  - 提交按钮文字随 mode 变：`{{ mode === 'login' ? '登录' : '注册' }}`
- `style` 新增 `.tabs` / `.tabs button` 样式（active 态青色下划线/底色，复用 `--brand`）

### 4. 后端注册链路端到端验证（不改代码）
- 起 Go 后端（`AUTH_DISABLE=1 SOURCE=mock`，阶段一约定）
- `curl -s -X POST localhost:8080/api/auth/register -H 'Content-Type: application/json' -d '{"username":"alice","password":"secret123"}'` → 期望 `{"code":0,"data":{"user":{...}}}`
- 重复注册同名 → 期望 `{"code":409,"message":"用户名已存在"}`
- 用新账号 `POST /api/auth/login` → 期望返回 `token`
- （可选）查 `backend/*.db` 确认行已插入（或信后端唯一约束 + 登录成功即可）
- 若发现后端 bug（未编译/未落库），单独记录并修复——但预期无需改

## 数据流

```
登录页 Tab=注册 → 填 账号/密码/确认密码 → onSubmit(register 分支)
  → 前端校验(一致 + ≥6 位)
  → auth.register() → request.post('/api/auth/register')
  → 后端 service.Register() → bcrypt 哈希 → repo.CreateUser() → SQLite users 表
  → 返回 {user}
  → ElMessage 成功 → mode='login' + 预填账号
  → 用户点登录 → auth.login() → JWT → 进 /insight
```

## 错误处理

- 前端：两次密码不一致、密码<6 位 → 行内 `err` 提示，不发请求
- 后端 409（用户名已存在）→ `auth.register` 抛错 → `err` 显示「用户名已存在」
- 后端 400（参数缺）→ 显示后端 message
- 网络/超时 → axios 拦截器已有兜底（非 JSON 响应 reject），`err` 显示「注册失败」
- 注册成功但登录失败 → 已切回登录 Tab 预填账号，用户重试即可

## 验证

1. 后端实测：起后端 → curl 注册新账号成功 → 重复注册 409 → 新账号登录拿 token
2. `npm run dev`（重启，因改了 LoginView）→ 登录页点「注册」Tab
3. 填 账号/密码/确认密码（不一致 / <6 位 应有行内提示）→ 提交
4. 成功提示「注册成功，请登录」→ 自动切回登录 Tab 且账号已预填
5. 点登录 → 进看板；右上角变头像首字母
6. 退出登录 → 回「登录」按钮（v15 已实现的闭环不受影响）

## 备注

- 后端注册实现来自阶段一 `docs/superpowers/specs/2026-07-25-backend-auth-adapter-design.md` 与 `auth_service.go`，本次不重复造轮子。
- 生产数据接口公开性（v15 遗留议题）不在本次范围。
