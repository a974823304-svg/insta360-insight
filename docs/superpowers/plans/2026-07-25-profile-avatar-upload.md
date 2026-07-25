# 资料保存 500 修复 + 头像文件上传 实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 修复资料保存报 500 的两个根因（dev 模式 id 未回填、后端不可达时迷惑文案），并新增后端头像文件上传能力（POST /api/user/avatar，图片存磁盘，avatar 字段存 /avatars/ URL）。

**Architecture:** 后端在受 JWT 保护的 `/api/user/avatar` 接收 multipart 图片，校验类型/大小后以服务端生成的安全文件名落盘到 `data/avatars`，并通过 `engine.Static("/avatars", …)` 公开托管；返回的 `/avatars/<file>` 写回现有 `avatar` 字段。前端在资料页加「上传图片」按钮，上传成功后回填 `form.avatar` 并预览；`TopBar` 与编辑页的预览逻辑扩展为识别 `http(s)://`、`/avatars/`、`data:` 三种形式。

**Tech Stack:** Go + Gin + SQLite（modernc.org/sqlite，纯 Go 无 CGO）；前端 Vue 3 + Vite + Element Plus + Axios。

## Global Constraints

- 纯 Go 依赖、无 CGO：文件上传只用标准库（`mime/multipart`、`os`、`path/filepath`），不引入新依赖。
- 前端 API 模块约定：`frontend/src/api/*.js` 返回完整 `{code,data}` 信封，调用方自行检查 `res.code` 并取 `res.data`（本次 `uploadAvatar` 返回原始信封，组件内拆 `res.data.url`）。
- 头像上传白名单：Content-Type 限 `image/png`/`image/jpeg`/`image/webp`；大小上限 2MB。
- `avatar` 字段取值：`preset:*` / `https://...` / `/avatars/<file>` / 空串（回退首字母）—— 向后兼容，不改动 model。
- dev 模式（`AUTH_DISABLE=1` 或 `ENV=dev`）跳过 JWT 并注入 devUser；`ENV=dev` 下 `go run .` 直接可用。
- 响应统一用 `model.OK` / `model.Fail(code, msg)`；业务失败返回 `code:400`，用户不存在用 `code:404`（不再用 500 表达业务错误）。

---

### Task 1: 后端 store — CreateUser 回填自增 id（修复 dev 模式 500 根因）

**Files:**
- Modify: `backend/internal/store/user_repo.go`（`CreateUser` 函数）
- Test: `backend/internal/store/user_repo_test.go`（新增 `TestCreateUserSetsID`）

**Interfaces:**
- Consumes: `model.User`、`r.db.Exec`
- Produces: `CreateUser(u *model.User) error` —— 调用后 `u.ID` 被回填为自增主键（后续 `Register`/`SeedAdmin`/`devUser.UserID` 依赖此值）

- [ ] **Step 1: 写失败测试**

在 `user_repo_test.go` 末尾追加：

```go
func TestCreateUserSetsID(t *testing.T) {
	repo := newTestRepo(t)
	u := &model.User{Username: "idtest", PasswordHash: "x", Role: "viewer"}
	if err := repo.CreateUser(u); err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	if u.ID <= 0 {
		t.Fatalf("expected positive ID after CreateUser, got %d", u.ID)
	}
	got, err := repo.GetByID(u.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.Username != "idtest" {
		t.Fatalf("unexpected user %+v", got)
	}
}
```

- [ ] **Step 2: 运行测试确认失败**

Run: `cd backend && go test ./internal/store/ -run TestCreateUserSetsID -v`
Expected: FAIL（`u.ID <= 0`）

- [ ] **Step 3: 最小实现**

将 `user_repo.go` 的 `CreateUser` 改为：

```go
// CreateUser 插入新账号,并回填自增主键到 u.ID。
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

- [ ] **Step 4: 运行测试确认通过**

Run: `cd backend && go test ./internal/store/ -run TestCreateUserSetsID -v`
Expected: PASS

---

### Task 2: 后端 handler — AvatarUpload + NewAuth 签名 + ProfileGet 改 404

**Files:**
- Modify: `backend/internal/api/handler/auth.go`（`NewAuth` 增加 `avatarDir` 字段/参数；新增 `AvatarUpload`、`allowedAvatarExt`；`ProfileGet` 的 `Fail(500)` → `Fail(404)`）
- Modify: `backend/internal/api/handler/auth_test.go`（更新 `setupAuthEngine` 的 `NewAuth` 调用；新增 `setupAvatarEngine`、`TestAvatarUploadSuccess`、`TestAvatarUploadRejectsNonImage`、`TestAvatarUploadMissingFile`）
- Modify: `backend/internal/api/router/router.go`（`NewAuth(authSvc)` → `NewAuth(authSvc, "data/avatars")`，仅保编译通过，路由/Static 在 Task 3 补）

**Interfaces:**
- Consumes: `service.AuthService.GetProfile(userID)`、`c.FormFile`、`c.SaveUploadedFile`、`model.OK`/`model.Fail`
- Produces:
  - `func NewAuth(svc *service.AuthService, avatarDir string) *Auth`
  - `func (h *Auth) AvatarUpload(c *gin.Context)` —— 成功返回 `model.OK(gin.H{"url": "/avatars/<file>"})`，失败 `model.Fail(400, "…")`

- [ ] **Step 1: 写失败测试**

在 `auth_test.go` 顶部 import 增加：`"mime/multipart"`、`"os"`、`"path/filepath"`、`"strings"`。
在 `auth_test.go` 末尾追加（含最小合法 PNG 字节与 helper）：

```go
// 最小 1x1 PNG,用于构造合法上传体
var minPNG = []byte{
	0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x00, 0x00, 0x0D, 0x49, 0x48, 0x44, 0x52,
	0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01, 0x08, 0x06, 0x00, 0x00, 0x00, 0x1F, 0x15, 0xC4, 0x89,
	0x00, 0x00, 0x00, 0x0A, 0x49, 0x44, 0x41, 0x54, 0x78, 0x9C, 0x63, 0x00, 0x01, 0x00, 0x00, 0x05, 0x00,
	0x01, 0x0D, 0x0A, 0x2D, 0xB4, 0x00, 0x00, 0x00, 0x00, 0x49, 0x45, 0x4E, 0x44, 0xAE, 0x42, 0x60, 0x82,
}

func setupAvatarEngine(t *testing.T) (*gin.Engine, string) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	repo, err := store.NewUserRepo(t.TempDir() + "/av.db")
	if err != nil {
		t.Fatalf("repo: %v", err)
	}
	t.Cleanup(func() { repo.Close() })
	svc := service.NewAuthService(repo, "test-secret")
	u, err := svc.Register("tester", "secret123", "admin")
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	avatarDir := t.TempDir()
	if err := os.MkdirAll(avatarDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	r := gin.New()
	a := NewAuth(svc, avatarDir)
	r.Use(func(c *gin.Context) {
		c.Set("claims", service.Claims{UserID: u.ID, Username: u.Username, Role: u.Role})
		c.Next()
	})
	r.POST("/api/user/avatar", a.AvatarUpload)
	return r, avatarDir
}

func TestAvatarUploadSuccess(t *testing.T) {
	r, avatarDir := setupAvatarEngine(t)
	body := &bytes.Buffer{}
	w := multipart.NewWriter(body)
	fw, _ := w.CreateFormFile("file", "a.png")
	fw.Write(minPNG)
	w.Close()
	req := httptest.NewRequest(http.MethodPost, "/api/user/avatar", body)
	req.Header.Set("Content-Type", w.FormDataContentType())
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	var resp model.APIResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp.Code != 0 {
		t.Fatalf("expected code 0, got %d msg=%s", resp.Code, resp.Message)
	}
	data := resp.Data.(map[string]interface{})
	url, _ := data["url"].(string)
	if !strings.HasPrefix(url, "/avatars/") {
		t.Fatalf("expected /avatars/ url, got %q", url)
	}
	if _, err := os.Stat(filepath.Join(avatarDir, filepath.Base(url))); err != nil {
		t.Fatalf("avatar file not saved: %v", err)
	}
}

func TestAvatarUploadRejectsNonImage(t *testing.T) {
	r, _ := setupAvatarEngine(t)
	body := &bytes.Buffer{}
	w := multipart.NewWriter(body)
	fw, _ := w.CreateFormFile("file", "a.txt")
	fw.Write([]byte("not an image"))
	w.Close()
	req := httptest.NewRequest(http.MethodPost, "/api/user/avatar", body)
	req.Header.Set("Content-Type", w.FormDataContentType())
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	var resp model.APIResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp.Code == 0 {
		t.Fatal("expected non-zero code for non-image upload")
	}
}

func TestAvatarUploadMissingFile(t *testing.T) {
	r, _ := setupAvatarEngine(t)
	req := httptest.NewRequest(http.MethodPost, "/api/user/avatar", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	var resp model.APIResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp.Code == 0 {
		t.Fatal("expected non-zero code when file missing")
	}
}
```

同时把 `auth_test.go` 中已有 `setupAuthEngine` 里的 `a := NewAuth(svc)` 改为 `a := NewAuth(svc, t.TempDir())`（保持编译通过）。

- [ ] **Step 2: 运行测试确认失败**

Run: `cd backend && go test ./internal/api/handler/ -run 'TestAvatarUpload' -v`
Expected: 编译失败（`NewAuth` 参数数量不符 / `AvatarUpload` 未定义）

- [ ] **Step 3: 最小实现**

`auth.go` 修改 import 块增加：`"fmt"`、`"os"`、`"path/filepath"`、`"strings"`、`"time"`。
结构体与构造函数：

```go
// Auth 账号相关 HTTP 处理器(瘦层)。
type Auth struct {
	svc       *service.AuthService
	avatarDir string
}

func NewAuth(svc *service.AuthService, avatarDir string) *Auth {
	return &Auth{svc: svc, avatarDir: avatarDir}
}
```

`ProfileGet` 改错误码：

```go
func (h *Auth) ProfileGet(c *gin.Context) {
	claims := c.MustGet("claims").(service.Claims)
	u, err := h.svc.GetProfile(claims.UserID)
	if err != nil {
		c.JSON(http.StatusOK, model.Fail(404, err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.OK(u))
}
```

新增上传逻辑：

```go
const avatarMaxBytes = 2 << 20 // 2MB

// allowedAvatarExt 按 Content-Type 返回允许的扩展名,不允许返回 ""。
func allowedAvatarExt(contentType string) string {
	switch contentType {
	case "image/png":
		return ".png"
	case "image/jpeg":
		return ".jpg"
	case "image/webp":
		return ".webp"
	default:
		return ""
	}
}

// AvatarUpload POST /api/user/avatar —— 接收图片文件,落盘后返回可访问 URL。
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
	ext := allowedAvatarExt(fileHeader.Header.Get("Content-Type"))
	if ext == "" {
		c.JSON(http.StatusOK, model.Fail(400, "仅支持 PNG / JPG / WEBP 格式"))
		return
	}
	// 删除旧本地头像(若当前 avatar 为 /avatars/ 开头),避免磁盘膨胀
	if cur, gerr := h.svc.GetProfile(claims.UserID); gerr == nil && strings.HasPrefix(cur.Avatar, "/avatars/") {
		_ = os.Remove(filepath.Join(h.avatarDir, filepath.Base(cur.Avatar)))
	}
	filename := fmt.Sprintf("%d_%d%s", claims.UserID, time.Now().UnixNano(), ext)
	if err := c.SaveUploadedFile(fileHeader, filepath.Join(h.avatarDir, filename)); err != nil {
		c.JSON(http.StatusOK, model.Fail(500, "头像保存失败"))
		return
	}
	c.JSON(http.StatusOK, model.OK(gin.H{"url": "/avatars/" + filename}))
}
```

`router.go` 中 `auth := handler.NewAuth(authSvc)` 改为 `auth := handler.NewAuth(authSvc, "data/avatars")`（Task 3 会替换为参数化目录）。

- [ ] **Step 4: 运行测试确认通过**

Run: `cd backend && go test ./internal/api/handler/ -run 'TestAvatarUpload|TestLogin|TestRegister' -v`
Expected: PASS

---

### Task 3: 后端 wiring — 路由注册 + 静态托管 + main.go 目录参数

**Files:**
- Modify: `backend/internal/api/router/router.go`（`New` 增加 `avatarDir string` 参数；注册 `g.POST("/user/avatar", …)`；`r.Static("/avatars", avatarDir)`）
- Modify: `backend/main.go`（读取 `AVATAR_DIR`、建目录、传给 `router.New`；`NewAuth` 改用参数化目录）

**Interfaces:**
- Consumes: `handler.NewAuth(svc, avatarDir)`、`engine.Static`
- Produces: 运行时 `POST /api/user/avatar` 可达；`GET /avatars/<file>` 可公开访问

- [ ] **Step 1: 改 router.go**

`New` 签名增加 `avatarDir string`：

```go
func New(insightSvc *service.InsightService, aiSvc *service.AIService, authSvc *service.AuthService, creatorSvc *service.CreatorService, disableAuth bool, devUser service.Claims, avatarDir string) *gin.Engine {
```

组内新增上传路由，并把 `NewAuth` 调用改为参数化：

```go
	auth := handler.NewAuth(authSvc, avatarDir)
```

在受保护组 `g` 内（`g.PUT("/user/profile", auth.ProfileUpdate)` 之后）追加：

```go
		g.POST("/user/avatar", auth.AvatarUpload)
```

在 `return r` 之前（根 engine 上，公开托管）：

```go
	r.Static("/avatars", avatarDir)
```

- [ ] **Step 2: 改 main.go**

imports 已含 `os`。在 `dbPath` 定义附近增加：

```go
	avatarDir := envOrDefault("AVATAR_DIR", "data/avatars")
	if err := os.MkdirAll(avatarDir, 0o755); err != nil {
		log.Fatalf("创建头像目录失败: %v", err)
	}
```

并把启动调用改为：

```go
	engine := router.New(insightSvc, aiSvc, authSvc, creatorSvc, disableAuth, devUser, avatarDir)
```

- [ ] **Step 3: 构建 + 冒烟验证**

Run:
```bash
cd backend && go build ./... && go test ./... 
```
Expected: 全部编译通过、测试 PASS。

启动（dev 模式，临时库）：
```bash
cd backend && DB_PATH=/tmp/be_avatar_demo/app.db AUTH_DISABLE=1 ENV=dev JWT_SECRET=testsecret ./be.exe
```
（先 `go build -o be.exe .`；Windows 用 `be.exe`，Linux/macOS 用 `./be`）

上传冒烟（另开终端，准备一张小图 `sample.png`）：
```bash
curl -F "file=@sample.png" http://localhost:8080/api/user/avatar
# 期望 {"code":0,"data":{"url":"/avatars/1_....png"}}
curl -I http://localhost:8080/avatars/<上一步的 file>
# 期望 HTTP/1.1 200 且 Content-Type: image/png
```

- [ ] **Step 4: 收尾**

删除临时构建产物 `be.exe`（在 backend 目录下，避免误入库；可用 `cmd /c del` 或移出仓库）。

---

### Task 4: 前端 — uploadAvatar API + 资料页上传 UI + 友好报错

**Files:**
- Modify: `frontend/src/api/request.js`（新增 `uploadAvatar`）
- Modify: `frontend/src/views/ProfileView.vue`（文件输入 + 上传按钮 + `isImageRef` + 保存失败友好文案）

**Interfaces:**
- Consumes: `request` 实例（axios 封装）
- Produces: `uploadAvatar(file)` 返回原始信封 `{code,data:{url}}`；`ProfileView` 将 `res.data.url` 写入 `form.avatar`

- [ ] **Step 1: request.js 新增上传方法**

在 `updateProfile` 之后追加：

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

- [ ] **Step 2: ProfileView.vue 模板增加上传控件**

将 `<el-form-item label="头像">` 内部改为：

```html
        <el-form-item label="头像">
          <div class="avatar-row">
            <div
              v-for="p in presets"
              :key="p.id"
              class="preset"
              :class="{ active: form.avatar === p.id }"
              :style="{ background: `linear-gradient(135deg, ${p.from}, ${p.to})` }"
              @click="form.avatar = p.id"
            >{{ initial }}</div>
            <div v-if="isImageRef(form.avatar)" class="preset url">
              <img :src="form.avatar" alt="avatar" />
            </div>
          </div>
          <div class="avatar-actions">
            <el-button size="small" :loading="uploadLoading" @click="pickFile">上传图片</el-button>
            <input ref="fileInput" type="file" accept="image/png,image/jpeg,image/webp" hidden @change="onFile" />
            <span v-if="uploadErr" class="upload-err">{{ uploadErr }}</span>
          </div>
          <el-input v-model="form.avatar" placeholder="或粘贴头像图片 URL（留空则用上方预设 / 首字母）" size="large" class="avatar-input" />
        </el-form-item>
```

- [ ] **Step 3: ProfileView.vue 脚本逻辑**

`<script setup>` 顶部 import 增加：
```js
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { useAuthStore } from '../stores/auth'
import { uploadAvatar } from '../api/request'
```

在 `const isUrl = ...` 处替换为：
```js
const isImageRef = (v) =>
  typeof v === 'string' &&
  (/^https?:\/\//.test(v) || v.startsWith('/avatars/') || v.startsWith('data:'))
```

在 `loading`/`err`/`ok` 附近新增：
```js
const fileInput = ref(null)
const uploadLoading = ref(false)
const uploadErr = ref('')

function pickFile() {
  fileInput.value?.click()
}

async function onFile(e) {
  const f = e.target.files?.[0]
  if (!f) return
  uploadErr.value = ''
  if (!/image\/(png|jpeg|webp)/.test(f.type)) {
    uploadErr.value = '仅支持 PNG / JPG / WEBP'
    e.target.value = ''
    return
  }
  if (f.size > 2 << 20) {
    uploadErr.value = '图片不能超过 2MB'
    e.target.value = ''
    return
  }
  uploadLoading.value = true
  try {
    const res = await uploadAvatar(f)
    if (!res || res.code !== 0) throw new Error((res && res.message) || '上传失败')
    form.avatar = res.data.url
  } catch (err) {
    uploadErr.value = (err && err.message) || '上传失败'
  } finally {
    uploadLoading.value = false
    e.target.value = ''
  }
}
```

`onSubmit` 的 `catch` 改为友好文案：
```js
  } catch (e) {
    const msg = (e && e.message) || '保存失败'
    err.value = /status code 500|Network Error|timeout/i.test(msg)
      ? '保存失败：无法连接服务器，请确认 Go 后端(:8080)已启动'
      : msg
  } finally {
    loading.value = false
  }
```

- [ ] **Step 4: 样式补充**

在 `<style scoped>` 中 `.avatar-input` 附近追加：
```scss
.avatar-actions { display: flex; align-items: center; gap: 10px; margin-bottom: 10px; }
.upload-err { color: #ff6b6b; font-size: 12px; }
```

- [ ] **Step 5: 构建验证**

Run: `cd frontend && npm run build`
Expected: 构建通过（无类型/编译错误）

---

### Task 5: 前端 — TopBar 头像识别 + Vite 代理 /avatars

**Files:**
- Modify: `frontend/src/components/TopBar.vue`（`avatarUrl` computed 扩展）
- Modify: `frontend/vite.config.js`（proxy 增加 `/avatars`）

**Interfaces:**
- Consumes: `auth.user.avatar` 字符串
- Produces: 右上角头像在下拉中正确渲染 `/avatars/` 与 `data:` 图片

- [ ] **Step 1: TopBar.vue avatarUrl 扩展**

将 `avatarUrl` computed 改为：
```js
// 头像图片地址:识别 http(s) / /avatars/ / data: 三种形式
const avatarUrl = computed(() => {
  const a = auth.user?.avatar || ''
  if (/^https?:\/\//.test(a) || a.startsWith('/avatars/') || a.startsWith('data:')) return a
  return ''
})
```

- [ ] **Step 2: vite.config.js 代理补全**

`server.proxy` 增加 `/avatars`：
```js
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true
      },
      '/avatars': {
        target: 'http://localhost:8080',
        changeOrigin: true
      }
    }
```

- [ ] **Step 3: 构建验证**

Run: `cd frontend && npm run build`
Expected: 构建通过

---

### Task 6: 端到端验证

**Files:** 无新增，仅验证

- [ ] **Step 1: 后端全量测试 + 构建**

Run: `cd backend && go build ./... && go test ./...`
Expected: 全绿

- [ ] **Step 2: 前端全量构建**

Run: `cd frontend && npm run build`
Expected: 构建通过

- [ ] **Step 3: 运行后端并冒烟上传 + 资料保存**

```bash
cd backend
go build -o be.exe .
DB_PATH=data/app.db AUTH_DISABLE=1 ENV=dev JWT_SECRET=testsecret ./be.exe &
curl -F "file=@sample.png" http://localhost:8080/api/user/avatar
# 期望 {"code":0,"data":{"url":"/avatars/1_....png"}}
curl -X PUT http://localhost:8080/api/user/profile -H 'Content-Type: application/json' \
  -d '{"nickname":"测试","avatar":"/avatars/1_....png","contact":"wx:test","bio":"hi"}'
# 期望 {"code":0,"data":{...}}  (dev 模式不再 no rows)
```
验证后停掉后端进程并清理 `be.exe`。

- [ ] **Step 4: 可选（用户侧）前端 dev 验证**

用户在本机 `cd frontend && npm run dev`，同时后端在 8080 运行；打开资料页 → 点「上传图片」→ 选 PNG/JPG/WEBP → 预览出现 → 保存 → 右上角头像更新；断后端后再点保存应显示「请确认 Go 后端(:8080)已启动」而非原始 500。

---

## Self-Review

1. **Spec 覆盖**：dev-ID bug（Task 1）✓；上传接口 + 校验 + 落盘 + 删旧（Task 2/3）✓；静态托管（Task 3 `r.Static`）✓；ProfileGet 404（Task 2）✓；前端上传 UI + isImageRef + 友好报错（Task 4）✓；TopBar 识别 /avatars/（Task 5）✓；vite 代理 /avatars（Task 5）✓；TDD（Task 1/2 含失败测试→实现→通过）✓；验证（Task 3/6 含 curl 冒烟）✓。
2. **Placeholder 扫描**：无 TBD/TODO；所有代码步均给出实际代码片段；前端交互无单测框架，已用 `npm run build` + curl 冒烟替代并标明。
3. **类型一致性**：`NewAuth(svc, avatarDir)` 在 Task 2（auth_test、router.go 字面量）、Task 3（router.New 参数、main.go）保持一致；`AvatarUpload` 返回 `model.OK(gin.H{"url":...})` 与前端 `res.data.url` 对应；`avatarDir` 类型 `string` 贯穿；`isImageRef` 在 ProfileView 与 TopBar 的判定表达式一致。
