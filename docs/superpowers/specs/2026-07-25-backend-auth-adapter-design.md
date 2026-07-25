# Insta360 达人营销数据洞察平台 — 后端：登录认证 + 数据接入层设计

> 日期：2026-07-25
> 状态：待用户评审
> 关联：`backend/README.md`、`frontend`（已完成）、`MEMORY.md`（项目长期约定）

---

## 1. 背景与目标

**现状**
- 前端看板已完成（KPI / 趋势 / 平台分布 / 赛道 / 引爆力 / 热门达人 TOP 10），风格为玻璃拟态暗色。
- 后端是 Go / Gin 的 **mock 服务**：`handler / service / mock / model / middleware` 分层齐全，但数据全部来自 `internal/mock/`（README 自述"后续替换为 ClickHouse / Doris"）。
- 前端在后端不可达时由 `fallback-data.js` 兜底。
- **当前没有任何 auth / user / login 代码**。

**两个关键约束（用户已确认）**
1. 系统**要对外 / 演示** → 必须有独立账号登录，否则任何人可看数据。
2. 真实数据来源是 **抖音 / B站 / 小红书 开放平台 API** → 需申请 appkey / 企业资质 / 授权，周期长（估 1–2 周起），**不能 block 看板**。

**目标**
- 先把"登录"这个门面 + 安全边界做出来，让演示流程完整。
- 数据导入用"可插拔接入层"抽象，先用现有 mock 顶上，等 appkey 到位再填真实实现，看板全程有数据、不被外部依赖卡住。

---

## 2. 总体顺序决策

| 阶段 | 内容 | 是否阻塞看板 | 外部依赖 |
|------|------|--------------|----------|
| **阶段一** | 登录认证（自建账号密码 + JWT + SQLite） | 否（dev 可 skip） | 无 |
| **阶段二** | 数据接入层抽象（adapter 接口 + MockAdapter 顶上 + 三平台空壳） | 否 | 无 |
| **阶段三** | 真实 API 接入（OAuth2 + 字段映射 + 落库） | 否 | 三平台 appkey |

**决策依据**
- 对外/演示 → 登录必须先上，演示时"登录页 → 看板"才是完整产品。
- 三平台 API 周期最长 → 不能让它阻塞看板，故用 adapter + mock 解耦。
- 登录是"冻结性"改动（一旦加 token，所有接口要校验），但在"对外"前提下必须提前，故放在最前。

---

## 3. 阶段一：登录认证

### 3.1 后端（Go / Gin）

#### 3.1.1 技术选型
- **认证**：JWT（HS256）。Secret 从环境变量 `JWT_SECRET` 读取，缺省仅 dev 用硬编码值（启动告警）。
- **密码哈希**：`golang.org/x/crypto/bcrypt`。
- **存储**：**SQLite**，驱动用 `modernc.org/sqlite`（纯 Go 实现，**无 CGO**）。
  - 库文件路径：`backend/data/app.db`（首次启动自动建表）。
  - **理由**：本机网络受限，MySQL / Postgres 二进制可能无法安装；SQLite 文件库零运维、演示够用。用户表（少量行）本就不适合放以后的 ClickHouse / Doris（OLAP 引擎）。
- **不引入任何 CGO 依赖**，确保本机可编译。

#### 3.1.2 数据模型 `internal/model/user.go`
```go
type User struct {
    ID           int64     `db:"id" json:"id"`
    Username     string    `db:"username" json:"username"`
    PasswordHash string    `db:"password_hash" json:"-"`
    Role         string    `db:"role" json:"role"` // "admin" | "viewer"
    CreatedAt    time.Time `db:"created_at" json:"created_at"`
}
```

#### 3.1.3 存储层 `internal/mock` 同级新增 `internal/store/user_repo.go`
- 使用 `database/sql` + `modernc.org/sqlite`（驱动名 `sqlite`）。
- 方法：`CreateUser(u *User) error`、`GetByUsername(name string) (*User, error)`。
- 首次启动执行 `CREATE TABLE IF NOT EXISTS users (...)`。
- **种子用户**：若表为空，用环境变量 `SEED_ADMIN_USER` / `SEED_ADMIN_PASS`（缺省 `admin` / `insta360`）插入一个 `admin` 账号；生产环境必须改密或禁用种子。

#### 3.1.4 服务层 `internal/service/auth_service.go`
- `Register(username, password, role) error`：校验用户名唯一、密码强度（≥6 位），bcrypt 哈希后落库。
- `Login(username, password) (token string, user *User, err error)`：查用户 → bcrypt 比对 → 签发 JWT（claims：`sub=userID`、`role`、`exp=now+7d`）。
- `ParseToken(token) (*Claims, error)`：供 middleware 复用。

#### 3.1.5 处理器 `internal/api/handler/auth.go`
- `POST /api/auth/register` → `{ code:0, data:{ user } }`（dev 可放开，生产可关闭注册）。
- `POST /api/auth/login` → `{ code:0, data:{ token, user } }`。
- 复用现有统一响应包装（与 `insight.go` 一致：`{ code, data, msg }`）。

#### 3.1.6 中间件 `internal/middleware/auth.go`
- `JWTAuth()`：解析 `Authorization: Bearer <token>`，无效/过期返回 401。
- 保护组：挂载在 `/api/insight/*`（health 除外）。
- **dev 放行**：环境变量 `AUTH_DISABLE=1` 或 `ENV=dev` 时跳过校验，向 context 注入一个默认 admin（username=`dev`，role=`admin`），方便本地调试且不影响前端联调。

### 3.2 前端（Vue 3）

#### 3.2.1 `views/LoginView.vue`
- 账号 / 密码表单，玻璃拟态 + 品牌色（与看板一致）。
- 提交 → 调 `/api/auth/login` → 存 token → 跳首页。
- 错误提示（账号或密码错误）。

#### 3.2.2 `stores/auth.js`（Pinia）
- state：`token`、`user`。
- actions：`login()`、`logout()`、`loadFromStorage()`。
- 持久化：localStorage（key 如 `insta_token`）。

#### 3.2.3 路由守卫（`router/index.js`）
- 访问受保护路由（非 `/login`）且无 token → 重定向 `/login`。
- 已登录访问 `/login` → 重定向首页。

#### 3.2.4 axios 拦截器（`api/request.js` 或现有 axios 封装）
- 请求头自动带 `Authorization: Bearer <token>`。
- 响应 401 → 清 token + 跳 `/login`。

#### 3.2.5 看板改动
- 现有看板**零改动**（契约不变，仅后端加了一层校验）。

### 3.3 阶段一验收
- 启动后端 + 前端，访问 `/` → 重定向 `/login`。
- 用种子 `admin / insta360` 登录 → 进入看板，数据正常渲染。
- 无 token 直接 `curl /api/insight/kpi` → 返回 401。
- `AUTH_DISABLE=1` 启动 → 前端免登录直接进看板（本地调试）。

---

## 4. 阶段二：数据接入层抽象

### 4.1 接口定义 `internal/service/source/adapter.go`
```go
// Creator / Trend / Interaction 复用 internal/model/types.go 已有结构
type PlatformAdapter interface {
    Name() string
    FetchCreators(ctx context.Context, f Filter) ([]model.Creator, error)
    FetchPlayStats(ctx context.Context, f Filter) (model.Trend, error)
    FetchInteraction(ctx context.Context, f Filter) (model.Interaction, error)
}
```
- `Filter` 复用现有筛选结构（`date_range / regions / tracks / platforms / age_bands`）。

### 4.2 `mock_adapter.go`
- 实现 `PlatformAdapter`，内部从 `internal/mock`（现有 `insight_data.go`）读取数据并返回。
- **目的**：看板立刻有"真实结构"的数据，且后续切真实源时前端零改动。

### 4.3 三平台空壳
- `douyin_adapter.go` / `bilibili_adapter.go` / `xiaohongshu_adapter.go`：实现接口，方法体返回 `nil, ErrNotImplemented`（并注明需要 appkey / OAuth2），留 `TODO` 注释。
- 供阶段三填充。

### 4.4 `insight_service.go` 改造
- 当前直接读 `mock` → 改为**注入 `PlatformAdapter`**。
- 来源由配置选择：`SOURCE=mock`（默认）/ `douyin` / `bilibili` / `xiaohongshu`。
- dev 默认 `mock`，看板不受影响。

### 4.5 前端
- **零改动**（契约不变）。

### 4.6 阶段二验收
- 看板数据经 adapter 返回，渲染与现有一致。
- 修改 `SOURCE` 环境变量不改前端代码即可切换数据源（空壳返回错误时前端走 fallback）。

---

## 5. 阶段三：真实 API 接入（概要，appkey 到位后细化）

- 各平台 OAuth2 / 开放平台授权流程（需企业资质 + appkey/secret）。
- 字段映射：平台返回 → 现有 `model` 结构（达人 / 播放趋势 / 互动 / 引爆力）。
- 定时拉取任务 + 落 ClickHouse / Doris（OLAP），与用户表（SQLite）分离。
- 前端可加"数据来源 / 最近更新时间"标识。
- **本阶段不在本次实现范围**，仅预留接口与空壳。

---

## 6. 技术约束与环境风险

- **本机网络受限**：Go 工具链需在放行网络下由用户安装（见 `环境安装保姆级指南.md`）；`modernc.org/sqlite` 是 Go 模块，`go mod download` 走 `proxy.golang.org`（已知可达），可正常拉取。
- **禁止 CGO 依赖**：所有新增依赖必须是纯 Go（modernc.org/sqlite 满足；bcrypt、jwt 均为纯 Go）。
- **JWT secret**：生产必须从 `JWT_SECRET` 环境变量提供强随机值；dev 硬编码值仅本地用。
- **种子账号**：`admin/insta360` 仅 dev/演示；生产部署必须改密或关闭自动种子。

---

## 7. 未决 / 风险

- 三平台 appkey 申请周期不可控 → 阶段三时间不确定，但已被阶段二解耦，不影响演示。
- 用户表与业务数据分库（SQLite vs ClickHouse/Doris）→ 在阶段三落地，本次仅用户表用 SQLite。
- 注册接口生产是否开放 → 建议生产关闭公开注册，仅管理员后台建号（本期先留 `register` 接口 + dev 开关）。

---

## 8. 验收总表

| 验收项 | 阶段 | 标准 |
|--------|------|------|
| 登录门面 | 一 | `/` 跳 `/login`，登录后进看板 |
| 无 token 拦截 | 一 | 直调 `/api/insight/*` 返回 401 |
| dev 放行 | 一 | `AUTH_DISABLE=1` 免登录进看板 |
| 密码安全 | 一 | 库内仅存 bcrypt 哈希，无明文 |
| 数据经 adapter | 二 | 看板数据走 `MockAdapter`，渲染一致 |
| 可切换源 | 二 | 改 `SOURCE` 环境变量不改前端 |
| 真实接入预留 | 三 | 三平台 adapter 空壳就位，等 appkey |
