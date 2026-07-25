# 设计：数据公开 + 右上角「登录」按钮 → 登录页 → 头像/登出

- 日期：2026-07-25
- 背景：用户希望打开网址即可看到看板数据（初始未登录），右上角显示「登录」按钮，点击后跳转到登录页；登录成功后右上角变为头像，点头像可「退出登录」。与 2026-07-25 早些时候实现的「dev 强制登录墙」相反，本次改为「数据公开 + 登录为可选操作」。
- 范围：仅前端（Vue 路由守卫 + TopBar + env）。后端数据接口公开性不在本次范围（dev 已用 `AUTH_DISABLE=1` 放行，生产数据接口开放待后续决策）。

## 目标

1. 打开 `/` 或 `/insight` 无需登录即可看到看板数据。
2. 未登录时右上角显示「登录」按钮；点击跳转 `/login`。
3. 登录成功后右上角显示头像（用户名首字母），点头像弹出下拉，含用户名与「退出登录」。
4. 退出登录后回到未登录态，右上角重新显示「登录」按钮，页面不跳转。

## 非目标（YAGNI）

- 不实现登录弹窗 / 抽屉（用户已选「跳转登录页」）。
- 不实现后端数据接口公开化（生产部署另议）。
- 不做注册、找回密码、改密。
- 不做基于登录态的内容权限裁剪（看板数据对所有访客一致公开）。

## 改动清单

### 1. 路由守卫：看板公开
文件：`frontend/src/router/index.js`

- 删除 `const AUTH_DISABLE = import.meta.env.VITE_AUTH_DISABLE === 'true'` 及基于它的分支。
- 新守卫逻辑：
  - 已登录（`localStorage` 有 `insta_token`）且访问 `meta.public` 路由（`/login`）→ 重定向 `/insight`（避免重复登录）。
  - 其余一律 `return true`（看板公开，未登录也能看）。
- `/insight` 不设 `meta.public`，靠「无 token 不重定向」自然公开。

### 2. TopBar：登录按钮 ↔ 头像下拉
文件：`frontend/src/components/TopBar.vue`

- 新增依赖：`useAuthStore`（`../stores/auth`）、`useRouter`、`computed`、`ElMessage`、`@element-plus/icons-vue` 的 `User` 与 `SwitchButton`。
- 右侧 `.right` 区：
  - 未登录（`!auth.isLoggedIn`）：渲染「登录」按钮（`<el-icon><User/></el-icon>` + 文字「登录」），`@click` 调 `goLogin()` → `router.push('/login')`。
  - 已登录：渲染 `<el-dropdown trigger="click">`，默认插槽为头像 `div.avatar`（内容 = `initial`，`title` = 用户名）；下拉插槽含「用户名（disabled 项）」与「退出登录」项（`command="logout"`，`@command` 调 `onCommand`）。
- `initial` computed：`auth.user?.username` 首字母大写，无则 `?`。
- `goLogin()`：`router.push('/login')`。
- `onCommand(cmd)`：`cmd === 'logout'` 时 `auth.logout()` + `ElMessage.success('已退出登录')`，停留在当前页（不跳路由）。
- 样式：
  - 新增 `.btn-login`（与历史 `.btn-export` 风格一致：青→蓝渐变、圆角 8px、白字、hover 上浮 + 阴影）。
  - `.avatar` 现有样式可复用；下拉项内图标/文字间距微调。

### 3. 环境变量
文件：`frontend/.env.development`

- 删除 `VITE_AUTH_DISABLE` 整行（前端守卫不再依赖该开关，避免死配置误导）。

### 4. LoginView（无需改动）
文件：`frontend/src/views/LoginView.vue`

- 现有登录成功 `router.push('/insight')` 已正确衔接新流程，不改。

## 数据流

```
打开 URL
  └─ router.beforeEach → 无 token + 访问 /insight → return true → 渲染看板（公开）
       └─ TopBar 读 auth.isLoggedIn === false → 显示「登录」按钮
            └─ 点击 → router.push('/login')
                 └─ LoginView 调 auth.login() → 写 token + user 到 localStorage + store
                      └─ router.push('/insight')
                           └─ TopBar 读 auth.isLoggedIn === true → 显示头像 A
                                └─ 点头像 → 下拉 → 「退出登录」
                                     └─ auth.logout() 清 localStorage + store
                                          └─ TopBar 响应式变回「登录」按钮
```

## 错误处理

- 登录失败：`LoginView` 已有 `try/catch` 显示 `err`，不受本次改动影响。
- 退出登录后若当前页依赖登录态数据：本项目看板数据公开，无需重载；保持当前页即可。
- `auth.user` 为 null 时 `initial` 回退 `?`，避免渲染空头像。

## 验证

1. `npm run dev`（改过 env/router，需重启 Vite）。
2. 打开 `http://localhost:5173/` → 直接看到看板；右上角为「登录」按钮（无头像）。
3. 点头像区「登录」→ URL 变 `/#/login` → 输 `admin` / `insta360` → 回 `/insight`，右上角变头像 `A`。
4. 点头像 → 下拉显示 `admin` + 「退出登录」→ 点退出 → 回到「登录」按钮，页面不跳。
5. 已登录状态下手动访问 `/#/login` → 应自动跳回 `/insight`（守卫逻辑验证）。

## 备注

- 生产部署：若后端开启 JWT 校验，公开前端会因取不到数据走 fallback 假数据。是否开放生产数据接口另议，不在本次范围。
- 历史记录：本次推翻 2026-07-25 早些时候「dev 强制登录（VITE_AUTH_DISABLE=false）」的决定，改为公开 + 可选登录。
