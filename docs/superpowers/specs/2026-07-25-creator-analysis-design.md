# 达人分析页 (Creator Analysis) — 设计文档 v19

- 日期: 2026-07-25
- 状态: 已批准 (用户 brainstorming 流程通过)
- 关联: 阶段二 adapter 层 (2026-07-25-phase2), 数据洞察页 (v1–v18), 个人资料页 (v18)

## 1. 背景与目标

「数据洞察」主页面经 v1–v18 打磨已完整;个人资料页 (v18) 刚完成。当前 5 个 Tab 中,
达人分析 / 内容分析 / 市场洞察 / 品牌分析 / 自定义看板 仍为 `PlaceholderView` 占位页。

本设计只做**达人分析 (Creator Analysis)** 一个 Tab,但目标不止于此:
把它**做深做透**,并建立一套**可复用「分析页模板」**,后续 内容/市场/品牌 Tab 直接复制骨架、换数据集。

### 目标
- 交付一个与「数据洞察」视觉/架构完全一致的「达人分析」真实页面 (概览仪表盘)。
- 后端接入阶段二 `DataSource` adapter 层,与现有架构统一。
- 沉淀可复用的分析页骨架 (view 外壳 + store + mock + 路由组),降低后续 Tab 成本。

### 非目标 (YAGNI / 范围边界)
- 不做逐人下钻详情 (用户明确选「概览仪表盘」)。
- 不做真实 API 接入 (卡在抖音/B站/小红书开放平台 appkey,属阶段三)。
- 不引入 ClickHouse / Doris,不做账号角色改动。
- 不重建已有的图表组件 (`KpiCard`/`TrendChart`/`PlatformDonut`/`TrackBarChart`/`AgeDonut`/`TopCreatorsTable`)。

## 2. 架构总览

```
前端 CreatorAnalysis.vue (view 外壳, 网格遵循 minmax 铁律)
   └─ stores/creator.js  (loadAll + fallback + filter, 镜像 insight.js)
        └─ api/request.js  GET /api/creator/*
             ├─ 成功 → 写 ref
             └─ 失败/非JSON/超时 → fallback-data.js (creator* 数据集)
后端 handler/creator.go
   └─ service/creator_service.go
        └─ service/source.DataSource (扩展 creator 域方法)
             ├─ MockAdapter        (真实 mock 实现)
             └─ douyin/bilibili/xiaohongshu_adapter (阶段三空壳, ErrNotImplemented)
```

### 复用约定 (项目铁律, 必须遵守)
- ECharts 容器: 严禁 `min-height` 硬编码;必备 `min-width:0; min-height:0; overflow:hidden` + `ResizeObserver` 主动 resize;容器 100% 自适应父级。
- CSS Grid 列宽: 必须用 `minmax(0, Xfr)`, 不用 `1fr`(会被内容撑爆); grid 内 `.card` 全加 `min-width:0`。
- 响应约定: 业务失败 HTTP 200 + `{code:非0, message}`;中间件未授权 HTTP 401。
- 暗色主题 + 玻璃拟态 + 品牌色 `--brand:#3DD9EB`。

## 3. 后端设计

### 3.1 DataSource 接口扩展
在 `backend/internal/service/source/adapter.go` 的 `DataSource` 接口新增 creator 域方法
(与现有 9 个 insight 方法并列):
```go
CreatorOverview(ctx, f Filter) (*CreatorOverview, error)
CreatorTrend(ctx, f Filter) ([]TrendPoint, error)
CreatorPlatforms(ctx, f Filter) ([]Share, error)
CreatorTracks(ctx, f Filter) ([]TrackStat, error)
CreatorAudience(ctx, f Filter) (*Audience, error)
CreatorList(ctx, f Filter) ([]Creator, error)
```
- `MockAdapter` 实现上述 6 方法 (调 `mock` 包新数据)。
- `douyin/bilibili/xiaohongshu_adapter.go` 对应方法返回 `ErrNotImplemented` (保持阶段三空壳一致)。

### 3.2 数据模型 `backend/internal/model/creator.go`
```go
type Creator struct {
    ID          int64            `db:"id" json:"id"`
    Name        string           `db:"name" json:"name"`
    Platform    string           `db:"platform" json:"platform"`     // 抖音/B站/小红书
    Track       string           `db:"track" json:"track"`           // 运动赛道
    Followers   int64            `db:"followers" json:"followers"`
    Engagement  float64          `db:"engagement" json:"engagement"` // 互动率 %
    Growth      int64            `db:"growth" json:"growth"`         // 近30日粉丝净增
    AvatarColor string           `db:"avatar_color" json:"avatarColor"`
    Tags        []string         `db:"tags" json:"tags"`
    Cooperation bool             `db:"cooperation" json:"cooperation"`
    AudienceAge map[string]float64 `db:"audience_age" json:"audienceAge"`
    AudienceGender map[string]float64 `db:"audience_gender" json:"audienceGender"`
}
```
字段与现有 `TopCreators` 对齐 (platform/track/followers/engagement/tags/avatarColor),
`TopCreatorsTable.vue` 零改动复用。新增 `growth`/`cooperation`/`audience*` 供本页专用。

### 3.3 Mock 数据 `backend/internal/mock/creator_data.go`
- 生成 ~20 个 `Creator` (覆盖 3 平台 × 多赛道, 与数据洞察现有 10 个达人风格一致, 扩充到 20 以撑满总表)。
- 派生函数:
  - `CreatorOverview`: 总数 / 本月新增 / 平均粉丝 / 平均互动率 / 合作占比。
  - `CreatorTrend`: 近 30 日粉丝净增序列。
  - `CreatorPlatforms`: 按平台聚合占比。
  - `CreatorTracks`: 按赛道聚合 (取 Top N)。
  - `CreatorAudience`: 年龄分段 + 性别占比。
  - `CreatorList`: 全量达人 (供总表)。

### 3.4 Service `backend/internal/service/creator_service.go`
- `CreatorService` 注入 `source.DataSource`。
- 6 个方法透传 source, 补 `ctx`; adapter 报错 → 返回 error 由 handler 转 `fail`。

### 3.5 Handler `backend/internal/api/handler/creator.go`
- `CreatorKPI` / `CreatorTrend` / `CreatorPlatforms` / `CreatorTracks` / `CreatorAudience` / `CreatorList`
  各方法 `c *gin.Context`, 解析 `Filter` querystring, 调 service, `ok(c, data)` / `fail(c, err)`。

### 3.6 路由 `backend/internal/api/router/router.go`
受保护组 (已有 JWT 中间件) 下新增:
```
GET /api/creator/kpi
GET /api/creator/trend
GET /api/creator/platforms
GET /api/creator/tracks
GET /api/creator/audience
GET /api/creator/list
```
`main.go` 改动: 构造 `CreatorService{Source: source}`, 并把 `router.New(...)` 签名增加 `creatorSvc` 参数传入
(与 insightSvc/aiSvc 注入方式一致)。

## 4. 前端设计

### 4.1 Store `frontend/src/stores/creator.js`
镜像 `insight.js`:
- state: `kpi / trend / platforms / tracks / audience / list / filter / loading`。
- `loadAll(filter)`: 并发 axios GET 6 个端点;成功写 ref;拦截器拒非 JSON + store 守卫非数组 → 回退 `fallback-data.js`。
- `setFilter(filter)`: 更新 filter 并重拉。

### 4.2 Fallback `frontend/src/api/fallback-data.js`
新增 `creatorKpi / creatorTrend / creatorPlatforms / creatorTracks / creatorAudience / creatorList`
(与后端 mock 同口径, 保证后端不可达时演示完整)。

### 4.3 视图 `frontend/src/views/CreatorAnalysis.vue`
外壳镜像 `InsightDashboard.vue` 的网格系统:
- `.page` padding 同数据洞察; `grid-row` 组合复用 `minmax(0,Xfr)` 规则。
- 板块顺序: KPI 行 → 粉丝增长趋势(全宽) → 平台分布 + 赛道分布(并排) → 粉丝画像 + 达人总表(并排, 总表占宽)。
- 顶部可选复用 `SideFilter` (与数据洞察同筛选侧栏), 经 `creatorStore.setFilter` 驱动。

### 4.4 组件复用清单
| 组件 | 用途 | 改动 |
|---|---|---|
| `KpiCard` | KPI 卡 ×5 | 零改动 |
| `TrendChart` | 粉丝增长趋势 | 零改动 (传 creator 趋势数据) |
| `PlatformDonut` | 平台分布 | 零改动 |
| `TrackBarChart` | 赛道分布 | 零改动 |
| `AgeDonut` | 粉丝画像(年龄+性别) | 零改动 (数据洞察 v10 删过,本页独立复用) |
| `TopCreatorsTable` | 达人总表 | 零改动 (full 10+ 行) |
| `SideFilter` | 筛选侧栏 | 零改动复用 |

→ **新增前端组件: 0** (仅新 view + store + fallback 数据)。

## 5. 接口契约 (响应体)
统一 `{ "code": 0, "data": ... }`;Filter 通过 querystring 多选:
`date_range / regions / tracks / platforms / age_bands` (mock 接受, 阶段三生效)。

示例:
```
GET /api/creator/kpi  → { "code":0, "data":[
  {key:"total",   label:"达人总数", value:20,   deltaPct:8.5},
  {key:"new",     label:"本月新增", value:3,    deltaPct:50},
  {key:"followers",label:"平均粉丝", value:1820000, deltaPct:4.2},
  {key:"engage",  label:"平均互动率", value:6.8, deltaPct:0.6},
  {key:"coop",    label:"合作占比", value:45,   deltaPct:12}
]}
```

## 6. 数据流与错误处理
1. `CreatorAnalysis.vue` `onMounted` → `creatorStore.loadAll(currentFilter)`。
2. store 并发 GET 6 端点;任一成功即写对应 ref。
3. `request.js` 拦截器: `typeof data === 'string'` → reject → 上层回退 fallback。
4. store 对数组类数据做 `Array.isArray` 守卫,异常不写 ref,保持 fallback。
5. 筛选变化 → `setFilter` → 重新 `loadAll`。

## 7. 测试与验收

### 后端 (用户在放行网络执行, 沙箱 Go 被网络拦截)
- 新增 `source/adapter_test.go` 补 creator 方法断言 (MockAdapter 返回非 nil, 空壳返回 ErrNotImplemented)。
- 新增 `creator_service_test.go` / `handler/creator_test.go` / `mock/creator_data_test.go` (TDD 先写后实现)。
- 验收: `cd backend && go build ./... && go test ./...` 全绿。
- curl 抽查: `GET /api/creator/kpi` 返回 `{code:0,...}`;`SOURCE=douyin` 时返回 502 `ErrNotImplemented`。

### 前端 (本环境可验证)
- `vite build` 通过 (CreatorAnalysis 独立 chunk)。
- Edge 无头 `--dump-dom`: KPI 卡渲染 (kpi-row 存在)、图表 canvas 存在、达人总表行存在、零页面级控制台错误。
- 路由: TopBar「达人分析」tab 激活高亮; 访问 `/#/creator` 渲染概览。

### 可复用模板交付物
- 在 `docs/superpowers/` 下补一份「分析页模板说明」(view 外壳结构 + store 范式 + mock 范式 + 路由注册步骤),
  供 内容/市场/品牌 Tab 复制。

## 8. 文件变更清单
后端 (新增/改):
- `internal/model/creator.go` (新)
- `internal/mock/creator_data.go` (新)
- `internal/service/source/adapter.go` (改: 接口 +6 方法)
- `internal/service/source/mock_adapter.go` (改: +6 实现)
- `internal/service/source/{douyin,bilibili,xiaohongshu}_adapter.go` (改: +6 返回 ErrNotImplemented)
- `internal/service/creator_service.go` (新)
- `internal/api/handler/creator.go` (新)
- `internal/api/router/router.go` (改: +6 路由 + `New` 增加 creatorSvc 参数)
- `backend/main.go` (改: 构造 CreatorService 并传入 router.New)
- 对应 `*_test.go` (新)

前端 (新增/改):
- `src/stores/creator.js` (新)
- `src/api/fallback-data.js` (改: +creator* 6 数据集)
- `src/views/CreatorAnalysis.vue` (新)
- (组件 / SideFilter / router 零改动; router 已有 `/creator` → PlaceholderView, 改指向 CreatorAnalysis)

## 9. 风险与注意
- 后端 Go 改动无法在沙箱编译 (网络拦截), 与 v18 同, 需用户本地 `go build`/`go test` 验收。
- 复用 `AgeDonut` 需注意: 数据洞察页 v10 删过粉丝画像, 但 `AgeDonut.vue` 组件仍在 (v7 预防性修复保留),
  本页独立引入不冲突。
- `TopCreatorsTable` 当前在 数据洞察页用 `slice(0,5)`;本页用全量 (去掉 slice), 两页互不影响。

## 10. 后续 (本设计不覆盖)
- 内容分析 / 市场洞察 / 品牌分析 Tab: 复制本模板, 换数据集。
- 自定义看板: 用户自选组件拼装, 形态不同, 单独设计。
- 阶段三真实 API: 填充三个空壳 adapter + 落 ClickHouse。
