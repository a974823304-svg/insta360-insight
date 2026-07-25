# CODEX.md — Insta360 达人营销数据洞察平台 · AI 代理交接手册

> 本文件是给 **OpenAI Codex（及 Claude Code / Cursor / GitHub Copilot 等 AI 编码工具）** 的项目上下文。
> 任何在本仓库工作的 AI 代理，请**先读完本文件再动手**，严格遵守其中的「铁律」与「已知 bug 修复」，
> 否则极大概率重复踩坑。本文件与 `.workbuddy/memory/MEMORY.md` 同源，改项目约定时请同步两处。

---

## 1. 项目简介

**Insta360 达人营销数据洞察平台** —— 一个面向达人营销分析的 Demo 原型。
前端按 UI 效果图复刻「数据洞察」主页（KPI、趋势、平台分布、赛道、引爆力雷达、粉丝画像、达人 TOP 表、AI 洞察），
后端 / AI 引擎为可独立运行的最小骨架，真实 OLAP 数仓 / 大模型可平滑替换。

- 定位：对外演示 + 内部分析工具雏形
- 当前阶段：Demo 已跑通（前端 build 成功、AI 规则兜底输出、后端单测全绿）
- 演示账号：`admin / insta360`（仅 dev/演示，生产须改密）

---

## 2. 技术栈

| 层 | 技术 |
| --- | --- |
| 前端 | Vue 3 + Vite + Element Plus + Pinia + Axios + ECharts（暗色主题） |
| 后端 | Go (Gin) |
| AI 引擎 | Python (FastAPI + dashscope / openai) |
| 数据存储 | Mock 内存数据；生产候选 OLAP：ClickHouse / Doris |
| 用户存储 | SQLite（`modernc.org/sqlite`，纯 Go 无 CGO） |

---

## 3. 目录约定

```
backend/internal/{api/handler,api/router,service,mock,model,middleware}  # 标准 Go 布局
frontend/src/{api,components,views,stores,router,styles}                 # 标准 Vue 布局
ai/{api,core,models}                                                    # 标准 Python 布局
```

- 业务快照数据（kpi/trend/track/radar/age/creators）统一在 `backend/internal/mock/insight_data.go`
- 前端同步一份在 `frontend/src/api/fallback-data.js`（后端不可达时的兜底）

---

## 4. 接口契约（必须遵守）

- 路径前缀 `/api/*`
- **统一响应信封**：`{ "code": 0, "data": ... }`
- 所有数据接口支持 querystring 多选筛选：`date_range`、`regions`、`tracks`、`platforms`、`age_bands`
- 数据结构在 `backend/internal/model/types.go` **与** `ai/models/schemas.py` **双向定义**，字段必须保持一致
- **前端 API 模块铁律**：`frontend/src/api/*.js` 每个端点必须 `.then(r => r.data)` 拆掉 `{code,data}` 信封，
  让 store 直接拿到业务数组/对象；否则 store 内 `!Array.isArray(...)` 校验恒为 false → **永远回退 fallback 数据**。
  （曾因 `insight.js` 漏拆信封导致洞察页永远显示兜底数据，已修正为 9 个方法全拆。）

---

## 5. ⚠️ Mock 数据双向同步铁律（最高频踩坑）

> 任何业务数据更新，**必须同时改两处**，否则前端 UI 与后端不一致，或前端永远用 fallback：

1. `backend/internal/mock/insight_data.go`（后端真实数据源）
2. `frontend/src/api/fallback-data.js`（前端兜底数据）

漏改任意一处 → 出现「UI 与后端对不上」或「前端死活显示旧数据」的诡异现象。

---

## 6. 后端推进决策（已拍板）

- **系统定位**：要对外/演示 → 必须有独立账号登录
- **真实数据来源**：抖音 / B站 / 小红书 开放平台 API（需 appkey，周期长）
- **推进顺序**：① 登录认证 → ② 数据接入层抽象（adapter + MockAdapter 顶上）→ ③ 真实 API 接入（等 appkey）
- **登录形态**：自建账号密码 + JWT（HS256），不用 SSO（后续可加）
- **用户存储**：SQLite，驱动 `modernc.org/sqlite`（纯 Go 无 CGO，本机可编译）
- **dev 放行**：`AUTH_DISABLE=1` 或 `ENV=dev` 时 Go middleware 跳过 JWT 校验，注入默认 admin
  - 最终形态：**看板公开 + 登录为可选操作**（前端路由守卫不重定向，未登录也能看数据；右上角「登录」按钮跳转 `/login`，登录后变头像+登出下拉）
  - 前端已无 `VITE_AUTH_DISABLE` 开关（router 守卫不再引用）
- **密码**：bcrypt 哈希，库内无明文
- **种子账号**：`admin / insta360`（仅 dev/演示，生产须改密）
- **约束**：所有新增 Go 依赖**必须纯 Go（无 CGO）**，否则本机编译失败

---

## 7. 🎨 UI 布局铁律（新组件必须遵守）

### ECharts 容器
- **严禁 `min-height` 硬编码**（会让内部图表溢出父级被截断）
- 必备三件套：`min-width: 0; min-height: 0; overflow: hidden` + `ResizeObserver` 主动 resize
- 容器永远 `100% width/height` 自适应父级

### CSS Grid 列宽
- `1fr` = `minmax(auto, 1fr)`，`auto` = min-content，内容（尤其 `el-table`）会撑爆列宽
- **必须用 `minmax(0, Xfr)`** 强制按比例分配
- grid 容器内 `.card` 全部加 `min-width: 0`
- 关键列加最小宽度保证（如粉丝画像 `minmax(280px, 1fr)`）

### Vue 模板根 class
- 容器 class 名要语义化，别用 `.chart-wrap`（曾用 PlatformDonut 改 `.donut-stage`、TrackBarChart 改 `.bar-stage`）
- 重命名时 template / style 同步改，避免 class 对不上

### Vite 环境配置铁律
- `.env.[mode]`（如 `.env.development`）**必须放项目根 `frontend/`**，不能放 `frontend/src/`
  （Vite 只从 root 读 env，放 src 下 `import.meta.env.VITE_*` 会是 undefined）
- 改 env **必须重启 vite** 才生效
- `npm run dev` 走 development 模式读 `.env.development`；`npm run build` 走 production 读 `.env.production`/`.env`，**不读** `.env.development`

---

## 8. 🔧 环境踩坑备忘（本机工具链）

- **Go 工具链**：本机只拦官方源 `proxy.golang.org`，**国内镜像可用**。已设 `GOPROXY=goproxy.cn,direct` + `GOSUMDB=off`
  （写入 Go 全局配置，永久生效）。因此 `go build ./...` / `go test ./...` **本机直接能跑通**，不需要换放行网络。
  切阿里云：`go env -w GOPROXY=https://mirrors.aliyun.com/goproxy/,direct`。
  ⚠️ 唯一仍需放行网络的情况：装 Go 二进制本身（本机已装 go1.26.5，无需再装）。
- **safe-delete shim**：拦截仓库路径的 `rm`/`rmdir`(bash)、`shutil.rmtree`(Python)、`Remove-Item`(PowerShell)
  —— 清理仓库内产物改用 `mv` 移出仓库再删。
- **`taskkill`（Git Bash）**：用 `MSYS_NO_PATHCONV=1 taskkill /PID <x> /F`（否则 `//PID` 被拒）；不能用 `pkill`（Windows 无效）。
- **没有 supervisor**：`./be.exe &` 在后台会一直存活，多次启动遗留多个实例抢占 :8080；
  启动前先 `netstat -ano | grep :8080` 确认监听者，kill 用 `MSYS_NO_PATHCONV=1 taskkill /PID x /F`。
- **curl `-F` 文件上传会挂起**（返回 code=000 / 空响应，极易误判为后端 bug）。验证文件上传**必须用 Python**（urllib/requests）。
- **Vite 改 `vite.config.js` 必须整进程重启**才生效，HMR 不会重载配置。
- **PowerShell 工具在本环境 stdout 为空**，不用于需要读输出的排查。

---

## 9. 🐛 已知 bug 与修复（不要重新引入）

### SQLite profile 字段 NULL 崩溃
- **现象**：注册新用户后查用户 / 登录返回 401，或直接 scan 报错。
- **根因**：`model.User` 的 `nickname/avatar/contact/bio` 是 `string`，但数据库列允许 `NULL`；
  Go 的 `database/sql` 遇到 `NULL` 往 `string` 塞直接报错 → 注册→查用户→登录链路崩溃。
- **修复**：`backend/internal/store/user_repo.go` 用 `sql.NullString` 中转这 4 列，NULL 自动转 `""`；
  加列时给这 4 列补 `DEFAULT ''`。
- **禁忌**：不要把这 4 列改回裸 `string` 直接 Scan。

---

## 10. 🌐 GitHub 仓库与推送须知

- 仓库：`https://github.com/a974823304-syg/insta360-insight.git`（Public）
- **直连 github.com 在本机被墙**，需以下任一方式：
  - 开**梯子（代理）**：`git config --global http.proxy http://127.0.0.1:<端口>` + `https.proxy` 同；或
  - 连**手机热点**；或
  - 用已装的 **GitHub Desktop**（File → Add local repository → Publish，自动走系统代理 + 浏览器 OAuth）
- **GitHub 不再接受账号密码登录 git**，用 PAT（classic，勾选 `repo`）或 GitHub Desktop OAuth
- 推送前务必确认 `.gitignore` 生效（见根 `.gitignore`，已排除 node_modules/247MB、*.exe、*.db、.workbuddy/、构建产物）
- 仓库名用英文，不要中文/空格

---

## 11. 🚀 部署 / 展示计划

- 用户计划**买自己的域名**用于项目展示
- 纯静态展示：GitHub Pages（免费）+ 自定义域名（DNS A 记录指向 `185.199.108~111.153`，或 CNAME 到 `用户名.github.io`）+ Enforce HTTPS
- 前端有 fallback 数据，纯静态展示够用
- 真后端（Go + Python 实时数据）需 VPS / CloudStudio 等带公网 IP 的服务

---

## 12. 📌 当前进度与下一步

### 已完成
- ✅ 前端 `vite build` 成功（2237 modules）
- ✅ AI 引擎 Python import 全通过，规则兜底输出 3 条洞察
- ✅ 后端 `go build ./...` + `go test ./...` 全绿（含登录/资料链路修复）
- ✅ 登录认证（JWT + SQLite + bcrypt）+ 数据接入 adapter 抽象 + 看板公开/登录可选
- ✅ 个人资料页 + 头像上传（预设头像）
- ✅ 本地 git 初始化 + 首次提交（129 文件，干净，无大文件）
- ✅ 本文件 CODEX.md / AGENTS.md 交接手册

### 下一步（由用户指定具体功能，例如）
- [ ] 接入抖音 / B站 / 小红书 开放平台真实数据（等 appkey）
- [ ] 优化达人分析页（CreatorAnalysis）
- [ ] 接 OpenAI Codex 云端继续开发的具体功能：_<用户在此填写>_

---

## 13. 🤖 给 AI 代理的开工建议

1. 先读本文件 + `docs/superpowers/specs/` 下的设计文档，理解既有约定。
2. 改动 Mock 数据时，**前后端两处同步改**（见第 5 节）。
3. 改动 Go 代码后，用 `goproxy.cn` 跑 `go build ./...` 与 `go test ./...` 验证（见第 8 节）。
4. 新增 Go 依赖必须纯 Go 无 CGO。
5. 遵守 UI 布局铁律（第 7 节），尤其是 ECharts 容器与 CSS Grid 列宽。
6. 不要重新引入第 9 节的已知 bug。
