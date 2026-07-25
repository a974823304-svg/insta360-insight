# 个人资料编辑页 — 设计文档 (Spec)

- **日期**: 2026-07-25
- **项目**: Insta360 达人营销数据洞察平台
- **关联**: 阶段一登录认证（账号密码 + JWT HS256 + SQLite）、v17 前端注册 UI
- **状态**: 已与用户确认设计，待写实现计划

---

## 1. 目标与范围

为**已登录用户**提供一个个人资料编辑页，支持：

- 填写 / 更新个人信息：昵称、头像、联系方式、个人简介
- 表单校验，确保必填字段完整
- 资料持久化存储（后端 SQLite，与登录账号绑定）
- 仅对认证用户开放，与现有登录系统无缝集成
- 顶栏头像下拉提供清晰导航入口

**不做（Out of Scope）**：真实头像文件上传/对象存储、第三方账号绑定、密码修改、资料对外展示页。

---

## 2. 已确认的关键决策（来自澄清）

| 项 | 决策 |
|----|------|
| 持久化方式 | 后端 SQLite 落库（扩展 `users` 表） |
| 存储架构 | 在 `users` 表直接加 4 列（非独立表、非 JSON 列） |
| 头像 | 存一个字符串：URL 链接 或 预设标识 `preset:xxx`（无文件上传） |
| 联系方式 | 单一自由文本字段 `contact` |
| 必填项 | **昵称 + 联系方式** 必填；头像、个人简介选填 |
| 用户名 | 登录账号，不可改（表单中禁用展示） |

---

## 3. 数据模型变更（后端）

### 3.1 表结构（SQLite）
在现有 `users` 表基础上新增 4 个可空列：

```sql
ALTER TABLE users ADD COLUMN nickname TEXT;
ALTER TABLE users ADD COLUMN avatar   TEXT;
ALTER TABLE users ADD COLUMN contact  TEXT;
ALTER TABLE users ADD COLUMN bio      TEXT;
```

- 迁移幂等：`migrate()` 中先用 `PRAGMA table_info(users)` 读取已有列，仅对缺失列执行 `ADD COLUMN`，避免重复加列报错。
- 兼容已有数据：旧账号这 4 列为 NULL，前端用占位（首字母 / 空简介）。

### 3.2 Go 模型 `model/user.go`
```go
type User struct {
    ID           int64     `db:"id" json:"id"`
    Username     string    `db:"username" json:"username"`
    PasswordHash string    `db:"password_hash" json:"-"`
    Role         string    `db:"role" json:"role"`
    CreatedAt    time.Time `db:"created_at" json:"created_at"`
    // 新增资料字段
    Nickname string `db:"nickname" json:"nickname"`
    Avatar   string `db:"avatar"   json:"avatar"`
    Contact  string `db:"contact"  json:"contact"`
    Bio      string `db:"bio"      json:"bio"`
}
```

### 3.3 `store/user_repo.go` 改动
- `GetByUsername` 的 SELECT 列表补充 `nickname, avatar, contact, bio`（使登录返回自带资料）。
- 新增 `GetByID(id int64) (*model.User, error)`：按主键查全字段。
- 新增 `UpdateProfile(id int64, u *model.User) error`：`UPDATE users SET nickname=?, avatar=?, contact=?, bio=? WHERE id=?`（只改资料列，不动账号/密码/角色）。

---

## 4. 后端接口契约

两个接口挂在受 `JWTAuth` 保护的 `/api` 组下。从 `c.MustGet("claims")` 取 `service.Claims.UserID` 作为当前用户。

### 4.1 `GET /api/user/profile`
响应：
```json
{ "code": 0, "data": { "id": 3, "username": "alice", "role": "viewer",
  "nickname": "Alice", "avatar": "preset:blue", "contact": "wx: alice01", "bio": "..." } }
```
（无 token → 中间件已返回 401，不会到此。）

### 4.2 `PUT /api/user/profile`
请求体：
```json
{ "nickname": "Alice", "avatar": "preset:blue", "contact": "wx: alice01", "bio": "热爱运动影像" }
```
校验规则（handler/service 层）：
- `nickname` 非空，否则 `model.Fail(400, "昵称为必填项")`
- `contact` 非空，否则 `model.Fail(400, "联系方式为必填项")`
- `avatar` / `bio` 选填；`bio` 服务端可加长度上限（如 ≤ 500 字符）防滥用
成功响应：`{ "code": 0, "data": <更新后的 User> }`
失败（如校验不通过）：`{ "code": 400, "message": "..." }`，HTTP 仍为 200（沿用项目统一响应约定）

### 4.3 服务与处理器
- `service.AuthService`：新增 `GetProfile(userID int64) (*model.User, error)`、`UpdateProfile(userID int64, nickname, avatar, contact, bio string) (*model.User, error)`
- `handler.Auth`：新增 `ProfileGet`、`ProfileUpdate` 两个方法
- `router`：`g.GET("/user/profile", auth.ProfileGet)`、`g.PUT("/user/profile", auth.ProfileUpdate)`

---

## 5. 前端数据层

### 5.1 `src/api/request.js`
```js
export function getProfile() { return request.get('/user/profile') }
export function updateProfile(payload) { return request.put('/user/profile', payload) }
```
（注意：`PUT` 经 Vite 代理到 `localhost:8080`，与现有 `/auth/*` 一致。）

### 5.2 `src/stores/auth.js`
把资料字段合并进单一 `user` 对象（显示唯一数据源）：
- 状态已含 `user`（含 nickname/avatar/contact/bio，登录即返回）
- 新增 `async fetchProfile()`：调 `getProfile()`，把返回 `data` 合并进 `this.user` 并写 `localStorage`（防止登录未带全时补全）
- 新增 `async saveProfile(payload)`：调 `updateProfile(payload)`，把返回 `data` 合并进 `this.user` + 写 `localStorage`，供 TopBar 即时反映
- `login()` / `register()` 成功后主动 `fetchProfile()`（注册用户资料为空，fetch 拿默认 NULL 也无妨）

---

## 6. 路由与守卫

### 6.1 `src/router/index.js`
新增：
```js
{ path: '/profile', component: () => import('../views/ProfileView.vue'), meta: { title: '个人资料', requiresAuth: true } }
```

### 6.2 守卫增强
现有守卫只处理"已登录访问登录页弹回"。补充：`requiresAuth` 且无 token → `return '/login'`。
看板 `/insight` 等仍保持公开（不要求登录）。

```js
router.beforeEach((to) => {
  const token = localStorage.getItem('insta_token')
  if (to.meta && to.meta.public && token) return '/insight'   // 已登录别再看登录页
  if (to.meta && to.meta.requiresAuth && !token) return '/login' // 资料页必须登录
  return true
})
```

---

## 7. 个人资料页 UI（`src/views/ProfileView.vue`）

布局：居中卡片（沿用登录页 `.login-stage`/`.login-card` 玻璃拟态 + 品牌色），表单分区。

字段与校验：
| 字段 | 控件 | 规则 |
|------|------|------|
| 账号 | `el-input` 禁用 | 显示 `user.username`，不可编辑 |
| 昵称 | `el-input` | **必填**（el-form rule required） |
| 头像 | 预设选择（几个渐变/emoji 预设）+ `el-input` 填 URL；右侧头像预览 | 选填；选中预设 → `avatar="preset:xxx"`；填 URL → 原样存 |
| 联系方式 | `el-input`（placeholder 电话/微信/邮箱） | **必填** |
| 个人简介 | `el-input type=textarea` + `show-word-limit`，maxlength 500 | 选填 |

交互：
- `onMounted`：`await auth.fetchProfile()` 回填表单（nickname/avatar/contact/bio）
- 提交：`auth.saveProfile({nickname, avatar, contact, bio})` → 成功 `ElMessage.success('资料已保存')`；失败显示错误信息
- 头像预览：URL → `<img>`；`preset:xxx` → 对应渐变圆；否则首字母

样式：复用 `--brand`/`--brand-2`/`--bg-elev`/`--glass-blur` 等主题变量，与 TopBar/Login 一致。

---

## 8. 导航入口（TopBar）

`src/components/TopBar.vue` 头像下拉：
- 新增 `el-dropdown-item command="profile"` → `onCommand` 中 `router.push('/profile')`（放在用户名与退出登录之间）
- 头像渲染升级：有 `auth.user?.avatar` 时
  - 以 `http` 开头 → 显示 `<img>` 头像
  - 以 `preset:` 开头 → 应用对应预设渐变样式（映射 class）
  - 否则回退首字母（现有逻辑）

---

## 9. 错误处理与边界

- 资料接口 401 → `request.js` 响应拦截器已处理（清 token + 跳登录页）
- 网络不可用 / 后端未起：保存失败 → Promise reject → 页面显示错误提示（资料属用户私有，无 fallback 数据，属预期）
- 迁移安全：`PRAGMA` 幂等检查，旧库升级不丢数据
- `UpdateProfile` 严格只改 4 列，绝不触碰 `username`/`password_hash`/`role`

---

## 10. 验证计划

### 前端（本机可验证）
1. `npx vite build --outDir .build-tmp` 编译通过（LoginView/ProfileView 各自 chunk）
2. Edge 无头 `--dump-dom`：
   - 无 token 访问 `/#/profile` → 重定向到 `/#/login`（验证 requiresAuth）
   - 带 token 访问 `/#/profile` → 表单含 昵称/头像/联系方式/个人简介 字段，无控制台报错
3. 登录态下顶栏下拉出现「个人资料」项

### 后端（需用户在放行网络 rebuild 后验收）
本机 Go 工具链网络受限、运行中的二进制为旧版（无 profile 接口），**无法在此环境直接跑新接口**。提供 curl 测试清单：
```bash
# 1) 注册
curl -s -X POST localhost:5173/api/auth/register -H 'Content-Type: application/json' -d '{"username":"p1","password":"abc12345"}'
# 2) 登录拿 token
TOKEN=$(curl -s -X POST localhost:5173/api/auth/login -H 'Content-Type: application/json' -d '{"username":"p1","password":"abc12345"}' | python -c 'import sys,json;print(json.load(sys.stdin)["data"]["token"])')
# 3) GET 资料（初始应带 nickname 等空）
curl -s localhost:5173/api/user/profile -H "Authorization: Bearer $TOKEN"
# 4) PUT 资料（必填齐全）
curl -s -X PUT localhost:5173/api/user/profile -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' -d '{"nickname":"P1","contact":"wx:p1","avatar":"preset:blue","bio":"hi"}'
# 5) PUT 缺必填（应 400 昵称/联系方式必填）
curl -s -X PUT localhost:5173/api/user/profile -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' -d '{"nickname":"","contact":""}'
# 6) 401 无 token
curl -s localhost:5173/api/user/profile
```
预期：1 注册成功；3 返回含资料字段；4 返回更新后资料；5 返回 400 校验失败；6 返回 401。

---

## 11. 改动文件清单

后端（Go）：
- `backend/internal/model/user.go` — 加 4 字段
- `backend/internal/store/user_repo.go` — 幂等迁移 + GetByID + UpdateProfile + 改 SELECT
- `backend/internal/service/auth_service.go` — GetProfile / UpdateProfile
- `backend/internal/api/handler/auth.go` — ProfileGet / ProfileUpdate
- `backend/internal/api/router/router.go` — 挂路由

前端（Vue）：
- `frontend/src/api/request.js` — getProfile / updateProfile
- `frontend/src/stores/auth.js` — fetchProfile / saveProfile + login/register 后调用
- `frontend/src/router/index.js` — /profile 路由 + 守卫增强
- `frontend/src/views/ProfileView.vue` — 新建
- `frontend/src/components/TopBar.vue` — 下拉加「个人资料」+ 头像渲染升级
