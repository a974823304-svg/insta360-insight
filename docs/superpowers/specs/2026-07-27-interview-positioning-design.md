# 面试定位设计稿 · Insta360 达人营销数据洞察平台

- 日期: 2026-07-27
- 状态: 结构已与用户逐批确认（两批全确认）
- 目标岗位: 全栈开发
- 时间约束: >3 个月（从容）
- 预算约束: 在校大学生，零成本优先（不买 SaaS / 域名 / VPS）

---

## 0. 背景与目标

项目已具备较完整的全栈骨架（Vue3 看板 + Go(Gin) 后端 + Python(FastAPI) AI 引擎 + 登录/分析页 + 可插拔真实数据接入层）。
但今天是**面试项目**，核心不是"再堆功能"，而是把已有能力**讲清楚、展得开、零成本可看**。

目标：
1. 把"可插拔数据接入层 + Mock 兜底"从架构图变成**有测试证明、能对接真实平台**的硬货。
2. 用 **GitHub Pages 免费**上线一个"点开即看"的演示，真后端作为代码/本地展示。
3. 交付一份面试可直接照着讲的文档（README / ARCHITECTURE / INTERVIEW_CHECKLIST）。
4. 诚实说明个人号拿不到"全网达人汇总"的边界——这反而显出对平台生态与成本的理解。

非目标（本期不做）：
- 不买域名 / VPS / 任何付费 SaaS。
- 不打通方案 B（本人 OAuth 实时真数据）——列为后续可选扩展。
- 不补完达人/内容/市场/品牌四域的真实映射（本期只补洞察域一域）。

---

## 1. 整体架构与交付形态

保持现有三层，明确分工以服务面试：

| 层 | 角色 | 上线方式 |
|---|---|---|
| 前端 Vue3 + ECharts | **演示面**：GitHub Pages 上线，用 fallback / demo-real 数据独立运行，不依赖后端 | 公网免费静态托管 |
| 后端 Go + Gin | **能力证明**：可插拔 `DataSource` 适配器 + 真实/Mock 兜底架构，本地可跑、仓库可见 | 不部署（省钱） |
| AI 引擎 Python | 同后端，规则兜底 + 可接大模型 | 不部署 |

关键叙事：**前端是"演示面"，后端是"能力证明"**。面试时打开 Pages 看效果，打开仓库讲后端架构与测试。

---

## 2. 真实数据补一域（核心，走抖音洞察域）

把首页 **KPI 卡 + 播放趋势图** 对应的 `Kpi` + `ViewsTrend` 从"企业权限占位"做成**个人抖音号真实可达**。

### 2.1 真实可达端点（个人号免费）
- **应用级 `client_token`**：用 `DOUYIN_CLIENT_KEY/SECRET` 拿抖音**公开数据**（部分榜单/视频类接口）。无需用户授权。
- **用户授权 `access_token`**：拿**你自己账号**的真实数据（`/oauth/userinfo/` 昵称/头像/粉丝相关 + 数据接口）。

### 2.2 字段映射
在 `backend/internal/service/source/douyin_adapter.go` 把真实响应 JSON 映射到 `Kpi` / `ViewsTrend` schema，**两种 token 模式都支持**（已有 `TokenProvider` + `StaticToken`/`ClientToken` 抽象，复用即可）。`requireToken` 拿不到 token 仍返回 `ErrNotImplemented` 触发回退，保持兜底不崩。

### 2.3 测试证明（面试硬通货，重点）
用 `httptest` 起一个**假抖音服务器**，喂**真实形状的响应 JSON**，断言适配器正确映射到我们的 schema。
- 新增 `backend/internal/service/source/real_insight_test.go`：`TestDouyinInsightRealMapping`（httptest 假抖音 → 断言 KPI/趋势映射正确）。
- 这套测试**不需要任何凭证即可运行**，`go test` 直接亮给面试官看，证明"不是假接口"。

### 2.4 演示合成真数据
- 新增 `frontend/src/api/demo-real-data.js`：数据形状**完全等价于"真实抖音响应经映射后"的结果**，GitHub Pages 用它演示。
- README 写明："有凭证时同一套代码返回的就是实时真数据，下面是测试证明。"——诚实且硬。

---

## 3. GitHub Pages 静态部署

- `vite.config.js` 设 `base: './'`（相对路径，项目页 `a974823304-syg.github.io/insta360-insight/` 与本地双击都能跑）。
- 构建产物 `frontend/dist` 推到仓库 `gh-pages` 分支（或 `/docs` 目录），开启 GitHub Pages。
- 前端已有"后端不可达自动回退 fallback"机制，静态托管下优雅降级为 `demo-real` 数据，**不白屏、不报错**。
- 真后端（Go+Python）仅在 README 给"可选本地全栈运行"步骤，不部署。

---

## 4. 仓库整理与文档

- **先提交真实数据接入层源码**（今天未提交）：`backend/internal/service/source/` 新文件（config/factory/oauth/fallback/util + 三平台 adapter + 测试）、`backend/.env.example`、`docs/真实数据接入指南.md`，连同分析页/侧栏等改动整理成干净 commit。
- **README 重写（面试门面）**：项目定位、技术栈一句话、架构图（三层 + 可插拔适配器）、本地全栈运行步骤、GitHub Pages 在线演示链接、真实数据接入说明（"有凭证即真数据 + 测试证明"）、目录约定、已知边界（个人号拿不到全网汇总）。
- **收敛架构说明**：整理出 `docs/ARCHITECTURE.md` 作为总入口，避免面试官在十几个 spec 里迷路。
- 保留 `AGENTS.md` / `CODEX.md`（方便后续接 Codex 继续改）。

---

## 5. 测试策略（面试硬通货）

- **Go 单测全绿**：现有 `go test ./...` 已覆盖 fallback 回退、工厂接线、抖音/B站/小红书映射；补 `TestDouyinInsightRealMapping`（见 §2.3）。
- **前端构建验证**：`npm run build` 零报错且产物能纯静态打开（用 demo-real 数据）；写步骤/脚本验证"无后端也能渲染"。
- **演示自检清单** `docs/INTERVIEW_CHECKLIST.md`：列清"打开 Pages 看什么、跑 `go test` 看什么、讲哪几个架构点"。

---

## 6. 面试叙事（把项目讲成亮点）

核心故事线（建议背熟）：
1. **可插拔数据接入层**：定义 `DataSource` 接口，Mock 与真实平台适配器可热插拔，凭证缺失自动回退——体现接口抽象与容错设计。
2. **真实对接抖音，且可验证**：适配器按真实 API 契约实现，用 httptest 模拟平台响应做映射测试，证明"不是假接口"。
3. **前端降级优雅**：后端不可达时前端用 fallback/demo-real 数据，不白屏——体现工程健壮性。
4. **全栈分层清晰**：Vue3 + Go(Gin) + Python(FastAPI)，JWT 鉴权、确定性筛选变换、ECharts 可视化各司其职。
5. **诚实边界**：主动说明个人号拿不到全网汇总数据（企业资质门槛），但架构已为真实数据预留——显出对平台生态与成本的理解，是加分项不是减分项。

---

## 7. 任务拆解与验收

| # | 任务 | 验收 |
|---|---|---|
| T1 | 抖音洞察域 `Kpi`/`ViewsTrend` 真实字段映射补全（client_token + 用户授权两模式） | `go build ./...` 0 错；`douyin_adapter.go` 两模式均可映射 |
| T2 | 新增 `real_insight_test.go` httptest 映射测试 | `go test ./internal/service/source/` 全绿，含真实形状响应断言 |
| T3 | 新增 `frontend/src/api/demo-real-data.js` 合成真数据 | 形状等价于真实映射结果；`npm run build` 静态可渲染 |
| T4 | `vite.config.js` 设 `base:'./'` + 验证静态打开 | dist 双击/index 直开不白屏 |
| T5 | 提交真实数据接入层 + 分析页等源码（干净 commit） | `git status` 干净，commit 信息清晰 |
| T6 | 重写 README + 新增 `docs/ARCHITECTURE.md` + `docs/INTERVIEW_CHECKLIST.md` | 文档覆盖第 6 节叙事点 |
| T7 | GitHub Pages 上线（推 dist 到 gh-pages / /docs） | 公网 URL 可访问，演示数据正常 |
| T8 | 全栈 `go test` + `npm run build` 终验 | 两者均 0 错误 |

---

## 8. 风险与开放项

- **风险**：GitHub 直连本机被墙，push 需梯子/手机热点/GitHub Desktop（仅 push 时联网，本地 commit 不受影响）。
- **风险**：真实抖音接口字段名/单位以官方文档为准，拿凭证后首要核对（adapter 注释已标注）。
- **开放项（后续可选）**：方案 B 本人 OAuth 实时真数据；补完其他域真实映射；自定义域名 + VPS（预算允许时）。
