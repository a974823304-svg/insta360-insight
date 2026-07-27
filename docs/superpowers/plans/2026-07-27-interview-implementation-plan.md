# 实现计划 · 面试定位（全栈岗 / GitHub Pages 免费演示 / 抖音洞察域写实 + 测试证明）

- 日期: 2026-07-27
- 来源: `docs/superpowers/specs/2026-07-27-interview-positioning-design.md`（已与用户两批确认）
- 配套设计: 同目录下设计稿

> 执行原则：本地 commit 用 Git Bash 直接做（不联网、安全）；**只有 push 到 GitHub 才需要联网**（梯子/热点/GitHub Desktop），push 动作由用户手动完成。所有新建依赖必须纯 Go（无 CGO），Go 工具链本机已可用（`GOPROXY=goproxy.cn`）。

---

## 阶段 0 · 现状复核（避免重复劳动）

执行前先确认以下"已就绪"项，**不要重做**：
- [x] `frontend/vite.config.js` 已 `base: './'` → T4 免做，仅做静态打开验证。
- [x] 前端 fallback 机制健壮（`stores/insight.js` 的 `withTimeout` + 结构校验 + `fillWithFallback`）→ 后端不可达自动降级，不白屏。
- [x] `douyin_adapter.go` 的 `Kpi`/`ViewsTrend` 已映射真实形状，`resolveTokenProvider(cfg, true)` 已支持 `client_token` + 用户授权 `access_token`；`requireToken` 无凭证返回 `ErrNotImplemented` 触发回退。
- [x] `oauth.go` 已有 `StaticToken` / `ClientToken` 抽象。
- [x] `real_data_test.go` 已有 `TestDouyinKpiMapping`（httptest 假抖音）。

**待办起点**：T1–T8 从"增强抖音真实映射测试 + 新增演示合成数据 + 文档重写 + Pages 上线"开始。

---

## T1 · 抖音洞察域写实增强（client_token + 用户授权两模式）

**目标**：让 `douyin_adapter.go` 的 `Kpi`/`ViewsTrend` 在两种 token 下都正确映射真实响应，并补注释说明端点差异。

**文件**：`backend/internal/service/source/douyin_adapter.go`

**步骤**：
1. 确认 `NewDouyinAdapter` 调用 `resolveTokenProvider(cfg, true)`（已是）。
2. `Kpi`：保持 `/data/external/user/overview/` 映射（fans_total/play_total/digg_total/comment_total/share_total/collab_total）。补充：当走 `client_token` 公开数据时，端点不同——若 `cfg.AccessToken==""` 且 `ClientKey` 有值，改用公开发现类端点（如 `/data/externel/item/...` 或文档允许的用户公开数据）。**若不确定公开端点字段，保持现状并在注释标注"用户授权路径为已验证形状，client_token 公开路径待凭证核对"**——不夸大。
3. `ViewsTrend`：保持 `/data/external/user/play/` 映射（list[date,views] → ViewsTrendPoint 含 PrevViews/Ratio）。
4. 在文件顶部 `DouyinAdapter` 注释补一句：两种 token 命中同一套映射函数，差异仅在 `get()` 的端点与鉴权头。

**验收**：`cd backend && go build ./...` 0 错；`go test ./internal/service/source/ -run Douyin` 通过。

---

## T2 · httptest 映射测试强化（面试硬通货）

**目标**：新增"真实形状"映射测试，证明适配器对接的是真实平台响应结构，无需凭证即可运行。

**文件**：`backend/internal/service/source/real_insight_test.go`（新建）

**步骤**：
1. `TestDouyinInsightRealMapping`：httptest 假抖音服务器返回**真实形状** JSON——
   - overview：`{"data":{"error_code":0,"fans_total":<真实量级>,"play_total":...,"digg_total":...,"comment_total":...,"share_total":...,"collab_total":...}}`
   - play：`{"data":{"error_code":0,"list":[{"date":"2026-01-01","views":123},...]}}`
   断言 KPI 5 卡（followers/views 的 `Raw` 正确）、趋势点含正确 `PrevViews`/`Ratio`。
2. `TestDouyinViewsTrendMapping`：单独断言趋势映射（含首点 ratio=0、环比计算）。
3. 复用 `NewDouyinAdapter(PlatformConfig{AccessToken:"t", BaseURL:srv.URL})` + `a.httpDo = srv.Client().Do`（已有模式）。

**验收**：`go test ./internal/service/source/ -run 'Mapping|Insight'` 全绿；`go test ./...` 全绿。

---

## T3 · 前端演示合成真数据（demo-real-data.js）

**目标**：GitHub Pages 静态站用"形状等价于真实抖音映射结果"的数据演示，诚实标注"有凭证即实时真数据"。

**文件**：`frontend/src/api/demo-real-data.js`（新建）；`frontend/src/stores/insight.js`（微调）

**步骤**：
1. `demo-real-data.js` 导出与 `douyin_adapter` 输出**同构**的洞察域数据：
   - `kpi`：key 用 `creators/followers/views/engagement/collabs`，数值取"单个真实授权账号"量级的合理值（如 followers 1.2M、views 38M），`unit` 用 `humanize` 同款空/M/B。
   - `viewsTrend`：近 30 天播放序列（含 prev/ratio）。
   - `platformShare`：抖音 100%（单账号视角）。
   - 其余（tracks/radar/age/insights/topCreators）沿用真实风格的合理值，**不要**多平台混排（与单账号视角一致）。
2. `stores/insight.js`：`fillWithFallback()` 改为优先 `import demoReal from '../api/demo-real-data'`，用于 insight 页；保留 `fallback-data` 作为其他页（creator/content/market/brand）通用兜底。或在 insight store 顶部 `const fallback = demoReal` 直接切换（最省事，单账号视角更诚实）。
3. 在文件头注释标注：此数据形状等价于 `douyin_adapter.go` 真实映射输出，仅用于无后端时的演示。

**验收**：`cd frontend && npm run build` 0 错；构建产物在无后端时（双击 dist/index.html 或本地静态服务器）insight 页渲染 demo-real 数据、无白屏。

---

## T4 · 静态部署验证（base 已就位）

**目标**：确认 `dist` 在子路径/双击下都能正确加载资源。

**文件**：`frontend/vite.config.js`（已 `base:'./'`，免改）；验证步骤。

**步骤**：
1. `npm run build` 后，用 `python -m http.server` 在 `frontend/dist` 起静态服务，访问确认资源 200、insight 页渲染。
2. 确认 `dist/index.html` 内资源引用均为相对路径（`./assets/...`）。

**验收**：静态服务下页面正常，控制台无 404/白屏。

---

## T5 · 提交真实数据接入层 + 分析页源码（干净 commit）

**目标**：把今天（及近期）未提交的工作整理成清晰 commit。

**待提交文件清单**（用 `git add <具体文件>` 精确暂存，避免误带大文件）：
- `backend/internal/service/source/`：config.go / factory.go / oauth.go / fallback.go / util.go / douyin_adapter.go / bilibili_adapter.go / xiaohongshu_adapter.go / real_data_test.go（及 T2 新增 real_insight_test.go）
- `backend/.env.example`
- `docs/真实数据接入指南.md`
- 分析页/侧栏等既有改动（按 `git status` 实际列出）
- T3 的 `frontend/src/api/demo-real-data.js` + store 微调

**步骤**：
1. `git status -s` 看实际改动，分批 `git add` 具体路径。
2. 分 2 个语义 commit：
   - commit A：`feat(backend): 可插拔真实数据接入层（抖音/B站/小红书 adapter + fallback 兜底 + httptest 映射测试）`
   - commit B：`feat(frontend): 演示合成真数据 + 洞察域无后端降级到 demo-real`
3. **不 push**（联网动作留给用户）。

**验收**：`git status` 工作区干净（除未跟踪的大文件/忽略项）；`git log` 看到两个清晰 commit。

---

## T6 · 文档重写（面试门面）

**文件**：`README.md`（重写）、`docs/ARCHITECTURE.md`（新建）、`docs/INTERVIEW_CHECKLIST.md`（新建）

**README.md 结构**：
1. 一句话定位 + 技术栈徽章（Vue3 / Go(Gin) / Python(FastAPI) / ECharts）。
2. **架构图**：三层（前端演示面 / 后端能力证明 / AI 引擎）+ 可插拔 `DataSource` 适配器（Mock ⇄ 抖音/B站/小红书）的 ASCII 或 mermaid 图。
3. 在线演示链接（GitHub Pages，T7 后填）+ 本地全栈运行步骤（3 终端）。
4. **真实数据接入**小节：明确"有凭证即实时真数据（抖音已实装 `Kpi`/`ViewsTrend` 映射 + httptest 证明）+ 无凭证自动回退 Mock"，列 `backend/.env.example` 的变量。
5. 已知边界：个人号拿不到全网达人汇总（企业资质/消耗门槛），架构已预留。
6. 目录约定 + 测试命令（`go test ./...` / `npm run build`）。

**ARCHITECTURE.md**：总入口，收敛分散的 spec，给面试官一条阅读路径（README → ARCHITECTURE → source 层源码 → 测试）。

**INTERVIEW_CHECKLIST.md**：5 条叙事线（见设计稿 §6）对应"打开 Pages 看 X / 跑 `go test` 看 Y / 讲 Z"，含自查步骤。

**验收**：三份文档覆盖设计稿 §6 全部叙事点；README 可独立让陌生人看懂项目。

---

## T7 · GitHub Pages 上线（零成本自动化）

**目标**：push `main` 后自动构建并发布前端到公网，无需手动提交 dist、无需 VPS。

**方案**：GitHub Actions 工作流（规避本机被墙 + `frontend/dist` 被 gitignore 的问题）。

**文件**：`.github/workflows/deploy.yml`（新建）

**步骤**：
1. 写 `deploy.yml`：
   - `on: push: branches: [main]`
   - `permissions: pages: write; id-token: write`
   - job：checkout → setup-node → `npm ci` → `npm run build`（在 `frontend/`）→ `actions/upload-pages-artifact@v3`（path `frontend/dist`）→ `actions/deploy-pages@v4`。
   - 注意：`npm ci` 需要 `package-lock.json` 已提交；若无则改用 `npm install`。
2. 仓库 Settings → Pages → Source 选 "GitHub Actions"。
3. 用户用 GitHub Desktop 推 `main` 后，Actions 自动构建部署；拿到 `https://a974823304-syg.github.io/insta360-insight/`。
4. 在 README 填入该链接。

**验收**：Actions 跑绿；公网 URL 打开 insight 页渲染 demo-real 数据、无白屏；移动端/桌面端均正常。

---

## T8 · 终验（全栈绿 + 文档齐）

**步骤**：
1. 后端：`cd backend && go build ./... && go test ./...` → 全绿。
2. 前端：`cd frontend && npm run build` → 0 错；静态服务验证渲染。
3. 文档：README / ARCHITECTURE / INTERVIEW_CHECKLIST 三份齐全且在位。
4. 仓库：`git status` 干净；commit 历史清晰（含 T5 两个 commit + 设计稿 commit）。

**验收**：以上四项全部通过，输出"交付清单"给用户（含 Pages 链接、本地运行命令、面试叙事要点）。

---

## 检查点（Checkpoints）

- **CP1**（T1+T2 后）：后端 `go test ./...` 全绿，映射测试证明真实形状对接 → 可暂停让用户看测试。
- **CP2**（T3+T4 后）：前端 `npm run build` 静态可渲染 demo-real → 可暂停让用户看本地效果。
- **CP3**（T5+T6 后）：commit 干净 + 文档齐全 → 让用户复核文档叙事。
- **CP4**（T7 后）：Pages 上线 → 终验 T8。

> 每个检查点后可停下等用户确认，避免一次性摊大饼。方案 B（本人 OAuth 实时真数据）留作 CP4 之后的可选扩展。
