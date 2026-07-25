# 规格文档：资料保存修复 + 头像文件上传（后端上传接口方案）

- 日期：2026-07-25
- 关联设计决策：用户选择「后端上传接口」方案（头像存服务器磁盘，avatar 字段存访问 URL）
- 状态：已批准（待进入实现计划）

---

## 1. 背景与问题

用户在使用「个人资料」页时遇到两个问题：

1. **保存资料报 500**：点击「保存资料」后前端弹出红色错误 `Request failed with status code 500`。
   - 经定位，`PUT /api/user/profile` 本身只返回 HTTP 200（业务码 400/0）。HTTP 500 是 **Vite dev 代理**在后端不可达时返回的（`baseURL:'/api'` → 代理 → 死掉的 `:8080`）。
   - 另发现一个 **dev 模式真实 bug**：`store.CreateUser` 插入后未回填自增 `id`，导致 `SeedAdmin` 返回的 `admin.ID == 0`，dev 模式下注入的 `devUser.UserID = 0` → 资料 GET/PUT 命中 `GetByID(0)` → `sql: no rows in result set`。该 bug 在正常登录模式下不触发（登录返回的 user 带正确 id）。
2. **头像只能选预设渐变球 + 粘贴 URL**，没有「上传图片文件」的入口，用户希望直接上传本地图片。

本规格在修复 500 根因的同时，新增后端头像上传接口，使头像支持真实图片文件。

---

## 2. 目标

- 资料保存端到端可用（dev 模式 + 正常登录模式都不再 `no rows` / 不再因后端可达性给出迷惑的 500）。
- 新增「上传图片」能力：选本地图片 → 上传到服务器 → URL 写回 `avatar` 字段 → 编辑页与右上角头像都能正常显示。
- 前端对保存失败给出清晰、可操作的中文提示，而非原始 axios 错误。

## 3. 非目标（本次不做）

- 不做头像裁剪/编辑 UI（仅做前端大小校验与可选压缩，不做视觉裁剪）。
- 不接入对象存储（OSS/COS），本次为本地磁盘存储。
- 不改动既有 preset / URL 两种头像方式（保留）。

---

## 4. 后端设计

### 4.1 修复 dev 模式 id bug

文件：`backend/internal/store/user_repo.go` 的 `CreateUser`

```go
func (r *UserRepo) CreateUser(u *model.User) error {
	if u.CreatedAt.IsZero() {
		u.CreatedAt = timeNow()
	}
	res, err := r.db.Exec(
		`INSERT INTO users (username, password_hash, role, created_at) VALUES (?, ?, ?, ?)`,
		u.Username, u.PasswordHash, u.Role, u.CreatedAt,
	)
	if err != nil {
		return err
	}
	if id, err := res.LastInsertId(); err == nil {
		u.ID = id
	}
	return nil
}
```

`service.Register` / `service.SeedAdmin` 已返回 `*model.User`，修复后 `u.ID` 被正确回填，`main.go` 中 `devUser.UserID = admin.ID` 即正确。

### 4.2 新增上传接口 `POST /api/user/avatar`

- 归属：受 JWT 保护的 `/api` 组（与 `PUT /api/user/profile` 同组）。
- 请求：`multipart/form-data`，字段名 `file`。
- 校验：
  - Content-Type 须在白名单：`image/png`、`image/jpeg`、`image/webp`；
  - 文件大小 ≤ 2MB（`file.Size` 校验，超限返回 400）；
  - 必须有文件，否则 400。
- 存储：
  - 目录 `avatarDir` 由环境变量 `AVATAR_DIR` 控制，默认 `data/avatars`；启动时 `os.MkdirAll(avatarDir, 0o755)` 建目录。
  - 文件名由服务端生成，格式 `<userID>_<纳秒时间戳>.<ext>`，**不使用用户原始文件名**（防路径穿越 / 覆盖）。
  - 扩展名仅取白名单后缀（png/jpg/webp）。
  - 写入 `avatarDir/<filename>`。
  - 若用户当前 `avatar` 为本地头像（`/avatars/` 开头），先删除旧文件，避免磁盘膨胀。
- 响应：
  - 成功：`{ "code": 0, "data": { "url": "/avatars/<filename>" } }`
  - 失败：`{ "code": 400, "message": "..." }`

处理器实现要点（`backend/internal/api/handler/auth.go`）：

```go
const (
	avatarMaxBytes = 2 << 20
	avatarAllowExt = ".png .jpg .jpeg .webp"
)

func (h *Auth) AvatarUpload(c *gin.Context) {
	claims := c.MustGet("claims").(service.Claims)
	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusOK, model.Fail(400, "请选择要上传的图片"))
		return
	}
	if fileHeader.Size > avatarMaxBytes {
		c.JSON(http.StatusOK, model.Fail(400, "图片不能超过 2MB"))
		return
	}
	ct := fileHeader.Header.Get("Content-Type")
	ext := allowedAvatarExt(ct)
	if ext == "" {
		c.JSON(http.StatusOK, model.Fail(400, "仅支持 PNG / JPG / WEBP 格式"))
		return
	}
	// 生成安全文件名并保存 ...
	// 删除旧本地头像 ...
	url := "/avatars/" + filename
	c.JSON(http.StatusOK, model.OK(gin.H{"url": url}))
}
```

`NewAuth` 构造器新增 `avatarDir string` 参数。

### 4.3 静态托管头像

`main.go` 在 `engine := router.New(...)` 之后（或 router 内部）注册：

```go
engine.Static("/avatars", avatarDir)
```

`/avatars` 位于根路由（不在 `/api` 组），为公开可读，符合头像图片的公开属性。

### 4.4 `ProfileGet` 错误码修正

`backend/internal/api/handler/auth.go` 的 `ProfileGet`：将 `model.Fail(500, err.Error())` 改为 `model.Fail(404, err.Error())`，使错误信息更准确（用户不存在而非服务器错误）。

### 4.5 路由与依赖注入改动

- `backend/internal/api/router/router.go`：`New(...)` 新增 `avatarDir string` 参数；`auth := handler.NewAuth(authSvc, avatarDir)`；新增 `g.POST("/user/avatar", auth.AvatarUpload)`；并在根 engine 上 `r.Static("/avatars", avatarDir)`（或在 main 中注册，二选一，推荐在 main 中）。
- `backend/main.go`：读取 `AVATAR_DIR`（默认 `data/avatars`），`os.MkdirAll`，传入 `router.New(...)` 与 `handler.NewAuth`。

---

## 5. 前端设计

### 5.1 `frontend/src/api/request.js` 新增上传方法

```js
// 上传头像:POST /api/user/avatar (multipart),返回原始 APIResponse(含 code/data)
export function uploadAvatar(file) {
  const fd = new FormData()
  fd.append('file', file)
  return request.post('/user/avatar', fd, {
    headers: { 'Content-Type': 'multipart/form-data' }
  })
}
```

### 5.2 `frontend/src/views/ProfileView.vue`

- 头像区新增「上传图片」按钮 + 隐藏 `<input type="file" accept="image/png,image/jpeg,image/webp">`。
- 新增 `onFile(e)`：
  - 取 `e.target.files[0]`；前端校验类型（png/jpg/webp）与大小（≤2MB），不合规给出 `ElMessage` 提示并 return；
  - `loadingUpload` 中调用 `uploadAvatar(file)`；
  - 成功：`form.avatar = res.data.url`，并显示预览；
  - 失败：`uploadErr` 提示（含 500/网络不可达的友好文案）。
- 预览判定从 `isUrl` 升级为 `isImageRef`：识别 `http(s)://`、`/avatars/`、`data:` 三种形式：
  ```js
  const isImageRef = (v) =>
    typeof v === 'string' &&
    (/^https?:\/\//.test(v) || v.startsWith('/avatars/') || v.startsWith('data:'))
  ```
  模板 `<div v-if="isImageRef(form.avatar)" class="preset url"><img :src="form.avatar" /></div>`。
- 保存失败的友好文案（`onSubmit` 的 catch）：
  ```js
  const msg = (e && e.message) || '保存失败'
  err.value = /status code 500|Network Error|timeout/i.test(msg)
    ? '保存失败：无法连接服务器，请确认 Go 后端(:8080)已启动'
    : msg
  ```
- 预设渐变球 + URL 输入框保留不变。

### 5.3 `frontend/src/components/TopBar.vue`

`avatarUrl` computed 扩展为同时识别 `/avatars/` 与 `data:`：

```js
const avatarUrl = computed(() => {
  const a = auth.user?.avatar || ''
  if (/^https?:\/\//.test(a) || a.startsWith('/avatars/') || a.startsWith('data:')) return a
  return ''
})
```

使上传后的头像在右上角下拉中正常显示。

### 5.4 `frontend/vite.config.js` 代理补全

dev 模式下前端运行在 5173，头像相对路径 `/avatars/...` 需代理到 8080：

```js
proxy: {
  '/api': { target: 'http://localhost:8080', changeOrigin: true },
  '/avatars': { target: 'http://localhost:8080', changeOrigin: true }
}
```

---

## 6. 数据契约

### 上传请求
```
POST /api/user/avatar
Content-Type: multipart/form-data
Authorization: Bearer <token>
Body: file=<二进制图片>
```

### 上传响应（成功）
```json
{ "code": 0, "data": { "url": "/avatars/1_1719312345000000000.png" } }
```

### 上传响应（失败）
```json
{ "code": 400, "message": "仅支持 PNG / JPG / WEBP 格式" }
```

### avatar 字段取值（不变，向后兼容）
- `preset:blue | preset:cyan | preset:sun`（预设渐变球）
- `https://...`（外链图片 URL）
- `/avatars/<file>`（本次新增：服务器上传的头像）
- 空串（回退到首字母）

---

## 7. 涉及文件清单

后端：
- `backend/internal/store/user_repo.go` — `CreateUser` 回填 id（4.1）
- `backend/internal/service/auth_service.go` — 无需改逻辑，`Register` 自动获得正确 id
- `backend/internal/api/handler/auth.go` — `NewAuth` 加 `avatarDir`；新增 `AvatarUpload`；`ProfileGet` 错误码 404（4.2 / 4.4）
- `backend/internal/api/router/router.go` — `New` 加 `avatarDir`；注册 `POST /user/avatar` 与 `Static("/avatars", ...)`（4.5）
- `backend/main.go` — 读取 `AVATAR_DIR`、建目录、传参（4.3 / 4.5）
- 测试：`backend/internal/api/handler/auth_test.go`（或新增 `auth_upload_test.go`）、`backend/internal/service/auth_service_test.go`、`backend/internal/store/user_repo_test.go`（随 4.1 更新）

前端：
- `frontend/src/api/request.js` — `uploadAvatar`（5.1）
- `frontend/src/views/ProfileView.vue` — 文件按钮 + 上传 + `isImageRef` + 友好报错（5.2）
- `frontend/src/components/TopBar.vue` — `avatarUrl` 扩展（5.3）
- `frontend/vite.config.js` — 代理 `/avatars`（5.4）

---

## 8. 测试策略（TDD）

后端：
- `AvatarUpload` handler 测试：
  - 合法 png → `code:0` 且 `data.url` 以 `/avatars/` 开头，磁盘文件存在；
  - 非图片类型（如 `.txt` / 错误 Content-Type）→ `code:400`；
  - 缺 `file` 字段 → `code:400`；
  - 替换上传时旧 `/avatars/` 文件被删除（可选断言）。
- `Register` 单测：新建用户后 `u.ID > 0`（验证 4.1 修复）。

前端：
- 以构建通过 + 手动/curl 验证为主（无单测框架约束此交互）。

---

## 9. 验证步骤

1. 后端：`go build ./...` 与 `go test ./...` 全绿。
2. 启动后端（dev 模式）：
   ```
   cd backend
   DB_PATH=data/app.db AUTH_DISABLE=1 ENV=dev JWT_SECRET=testsecret go run .
   ```
3. 上传验证：
   ```
   curl -F "file=@/path/to/a.png" http://localhost:8080/api/user/avatar
   # 期望 {"code":0,"data":{"url":"/avatars/1_....png"}}
   curl -I http://localhost:8080/avatars/<file>   # 期望 200 image/png
   ```
4. 资料保存验证：`PUT /api/user/profile` 任意合法 payload → `code:0`（dev 模式不再 `no rows`）。
5. 前端：`npm run build` 通过；`npm run dev` 下打开资料页，上传图片 → 预览出现 → 保存 → 右上角头像更新。

---

## 10. 风险与注意

- **后端必须运行**：本方案依赖 Go 后端。若用户仅部署静态前端（CloudStudio），头像与资料保存仍不可用——这是「后端上传接口」方案的固有前提，已与用户确认。
- 文件落盘：`data/avatars` 需可写权限；容器/静态托管环境需挂载该目录。
- 头像 URL 为相对路径 `/avatars/...`：dev 靠 Vite 代理；若由 Go 直接托管前端静态资源则天然可用；若前端与后端跨域部署，需补充 CORS/反向代理（图片 `<img>` 本身不受 CORS 限制，可直接用绝对地址，本方案返回相对路径，跨域场景由部署侧处理）。
