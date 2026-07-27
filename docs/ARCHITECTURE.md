# 架构说明（总入口）

> 本文是 Insta360 达人营销数据洞察平台的架构总入口。按阅读目的分流：
> - **想快速看懂整体**：直接看第 1、2 节。
> - **想了解「真实数据怎么接」**：看第 3 节 + [`真实数据接入指南.md`](./真实数据接入指南.md)。
> - **想跑起来 / 部署**：看第 5 节。
> - **想面试讲解**：看 [`INTERVIEW_CHECKLIST.md`](./INTERVIEW_CHECKLIST.md)。
> - **想看某次功能的设计细节**：见 [`superpowers/specs/`](./superpowers/specs/) 与 [`superpowers/plans/`](./superpowers/plans/)。

---

## 1. 系统分层

| 层 | 技术 | 职责 | 生产替换 |
| --- | --- | --- | --- |
| 展示层 | Vue 3 + Vite + ECharts | 看板、筛选联动、图表、后端不可达降级 | 不变 |
| 接入 / 业务层 | Go (Gin) | `DataSource` 适配器调度、筛选变换、鉴权 | OLAP 查询替换 mock |
| 智能层 | Python (FastAPI) | 洞察生成（规则兜底 + 可接 LLM） | 接通义千问 / OpenAI |
| 存储层 | Mock 内存 + SQLite | 演示数据 + 用户表 | ClickHouse/Doris + 多角色 RBAC |

---

## 2. 可插拔数据接入层（核心）

代码位置：`backend/internal/service/source/`

```
DataSource (interface)
 ├─ MockAdapter            全量演示数据（确定性）
 ├─ DouyinAdapter          KPI/趋势 已写实；其余 ErrNotImplemented
 ├─ BilibiliAdapter        KPI/趋势 已写实；其余 ErrNotImplemented
 └─ XiaohongshuAdapter     KPI/趋势 已写实；其余 ErrNotImplemented

FallbackDataSource(real, mock)  装饰器：real 返回 ErrNotImplemented 或 error → 回退 mock
```

- **工厂 `NewDataSource(kind)`**：`SOURCE=mock|""` 返回裸 Mock；`douyin|bilibili|xiaohongshu` 返回 `FallbackDataSource(真实Adapter, Mock)`，凭证缺失自动降级。
- **Token 抽象 `TokenProvider`**：`StaticToken`（用配置的 access_token）+ `ClientToken`（抖音 client_token，内存缓存）。`resolveTokenProvider` 按「access_token → client_token（仅抖音）→ nil」优先级解析。
- **兜底铁律**：真实方法返回 `ErrNotImplemented`（无凭证 / 未实现）或任意 error，外层 `FallbackDataSource` 自动回退 Mock，看板永远有数据不崩。
- **前端同构**：前端 `demo-real-data.js` 的数据形状与后端真实映射输出完全等价；`stores/insight.js` 后端可达时 `loadAll()` 用真实数据覆盖，不可达则用 `demoReal` 填充。

### 测试证明（真实对接，无需凭证）
`real_insight_test.go` 用 `httptest` 起「假抖音服务器」，喂真实形状的 overview / play 响应，断言 `Kpi` / `ViewsTrend` 映射正确（含首点 ratio=0、持平、下跌边界），并验证 `client_token` 模式下端到端命中同一映射。

---

## 3. 前端结构

```
frontend/src/
  api/        axios(request, 数组参数裸 key 序列化) + insight.js(拆信封) + fallback-data / demo-real-data
  components/ TopBar / SideFilter / KpiCard / TrendChart / PlatformDonut / TrackBarChart / RadarChart / AgeDonut / TopCreatorsTable / DataTable
  views/      InsightDashboard / ContentAnalysis / MarketInsights / BrandAnalysis / Login / Profile
  stores/     filter(Pinia 全局筛选) / insight(数据) / user
  router/     路由（看板公开 + 登录可选，无强制守卫重定向）
```

关键约定（来自交接手册 `AGENTS.md`）：
- 所有走 `request` 的数组筛选参数必须序列化为**同名 key 重复**（`platforms=抖音&platforms=B站`），后端 `QueryArray` 读裸 key。
- ECharts 容器严禁 `min-height` 硬编码，必备 `min-width/min-height:0 + overflow:hidden + ResizeObserver`。
- CSS Grid 列宽用 `minmax(0, Xfr)`，禁止裸 `1fr` 撑爆。

---

## 4. 鉴权与用户

- 自建账号密码 + JWT（HS256），SQLite 存储（`modernc.org/sqlite`，纯 Go 无 CGO）。
- 密码 bcrypt 哈希，库内无明文；种子账号 `admin/insta360`（仅 dev/演示）。
- `ENV=dev` 或 `AUTH_DISABLE=1` 时 middleware 跳过 JWT，注入默认 admin；看板公开、登录为可选操作。

---

## 5. 部署

### 5.1 前端（GitHub Pages，零成本）
- `vite.config.js` 已设 `base: './'`，`frontend/dist` 推 `gh-pages` 分支 / 仓库 `/docs` 目录即可。
- 纯静态托管，后端不可达自动降级为 `demo-real` 数据，不白屏。
- 预期地址：`https://a974823304-syg.github.io/insta360-insight/`。

### 5.2 后端 / AI（本地，不部署公网）
- 见 `README.md` 第 5 节三终端启动；需公网实时真数据时才上 VPS / CloudStudio。

---

## 6. 设计文档索引

- 后端鉴权 / adapter：`superpowers/specs/2026-07-25-backend-auth-adapter-design.md`
- 分析页（内容 / 市场 / 品牌）：`superpowers/specs/2026-07-26-analysis-tabs-design.md` + `分析页模板说明.md`
- 筛选联动：`superpowers/specs/2026-07-26-filter-linkage-design.md`
- 侧栏折叠：`superpowers/specs/2026-07-26-sidefilter-collapse-design.md`
- 面试定位设计：`superpowers/specs/2026-07-27-interview-positioning-design.md`
- 面试实现计划：`superpowers/plans/2026-07-27-interview-implementation-plan.md`
