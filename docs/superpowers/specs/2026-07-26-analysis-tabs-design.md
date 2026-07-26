# 分析页扩展：内容 / 市场 / 品牌 Tab — 设计文档 v1

- 日期: 2026-07-26
- 状态: 已批准（brainstorming 流程通过，待写实施计划）
- 关联: 达人分析设计 `2026-07-25-creator-analysis-design.md`（已批准）、分析页模板 `docs/superpowers/分析页模板说明.md`、阶段二 adapter 层 `2026-07-25-phase2`

## 1. 背景与目标

「数据洞察」主页 5 个 Tab 中，达人分析已完成真实页面（概览仪表盘 + 可复用模板）。
其余 **内容分析 / 市场洞察 / 品牌分析** 三个 Tab 仍为 `PlaceholderView` 占位页。

本设计**一次覆盖三页**，全部复用达人分析沉淀的「分析页模板」骨架，差异只在数据集与板块语义。
目标：把三个 Tab 都做成与「数据洞察」「达人分析」视觉/架构完全一致的真实页面（概览仪表盘）。

### 目标
- 交付内容 / 市场 / 品牌 三个真实分析页（概览仪表盘），与既有页面视觉统一。
- 后端沿用阶段二 `DataSource` adapter 层，三页各扩展一组域方法。
- 沉淀一个**通用列表组件 `DataTable.vue`**，消除三页"列表"重复造轮子。

### 非目标 (YAGNI / 范围边界)
- 不做逐对象下钻详情（内容详情、竞品详情、品牌详情），只做概览。
- 不做真实 API 接入（卡在开放平台 appkey，属阶段三）；数据维持 mock。
- 不引入 ClickHouse / Doris，不做账号角色改动。
- 不重建已有图表组件（`KpiCard`/`TrendChart`/`PlatformDonut`/`TrackBarChart`/`AgeDonut`）。
- 不使用地图组件（地域分布用横向柱图），规避中国地图合规风险。

## 2. 架构总览

```
前端 ContentAnalysis.vue / MarketInsights.vue / BrandAnalysis.vue (view 外壳, 网格遵循 minmax 铁律)
   └─ stores/{content,market,brand}.js  (镜像 creator.js: loadAll + fallback + 800ms 超时)
        └─ api/{content,market,brand}.js  GET /api/{content,market,brand}/*
             ├─ 成功 → 写 ref
             └─ 失败/非JSON/超时 → fallback-data.js (对应数据集)
   └─ components/DataTable.vue (新增, 通用列表, 列配置)
后端 handler/{content,market,brand}.go
   └─ service/{content,market,brand}_service.go
        └─ service/source.DataSource (扩展 content/market/brand 三组共 18 方法)
             ├─ MockAdapter        (真实 mock 实现)
             └─ douyin/bilibili/xiaohongshu_adapter (阶段三空壳, ErrNotImplemented)
```

### 复用约定（项目铁律，必须遵守）
- ECharts 容器：严禁 `min-height` 硬编码；必备 `min-width:0; min-height:0; overflow:hidden` + `ResizeObserver` 主动 resize；容器 100% 自适应父级。
- CSS Grid 列宽：必须用 `minmax(0, Xfr)`，不用 `1fr`（会被内容撑爆）；grid 内 `.card` 全加 `min-width:0`。
- 响应约定：业务失败 HTTP 200 + `{code:非0, message}`；中间件未授权 HTTP 401。
- 暗色主题 + 玻璃拟态 + 品牌色 `--brand:#3DD9EB`。

## 3. 后端设计

### 3.1 DataSource 接口扩展
在 `backend/internal/service/source/adapter.go` 的 `DataSource` 接口，于现有 creator 方法之后，新增三组共 18 个方法：

```go
// 内容分析域
ContentKpi(ctx, f Filter) ([]KpiCard, error)
ContentTrend(ctx, f Filter) ([]ViewsTrendPoint, error)
ContentForms(ctx, f Filter) ([]PlatformShare, error)   // 形式分布
ContentTopics(ctx, f Filter) ([]TrackPerformance, error) // 主题分布
ContentDurations(ctx, f Filter) ([]AgeShare, error)      // 时长段分布
ContentList(ctx, f Filter) ([]ContentItem, error)

// 市场洞察域
MarketKpi(ctx, f Filter) ([]KpiCard, error)
MarketTrend(ctx, f Filter) ([]ViewsTrendPoint, error)
MarketCompetitors(ctx, f Filter) ([]PlatformShare, error) // 竞品占比
MarketRegions(ctx, f Filter) ([]TrackPerformance, error)  // 地域热度
MarketPrices(ctx, f Filter) ([]AgeShare, error)           // 价格段
MarketList(ctx, f Filter) ([]Competitor, error)

// 品牌分析域
BrandKpi(ctx, f Filter) ([]KpiCard, error)
BrandTrend(ctx, f Filter) ([]ViewsTrendPoint, error)
BrandPlatforms(ctx, f Filter) ([]PlatformShare, error)    // 平台分布
BrandSentiment(ctx, f Filter) ([]AgeShare, error)         // 情感分布
BrandKeywords(ctx, f Filter) ([]TagItem, error)           // 高频词
BrandList(ctx, f Filter) ([]PartnerBrand, error)
```

- `MockAdapter` 实现上述 18 方法（调 `mock` 包新数据）。
- `douyin/bilibili/xiaohongshu_adapter.go` 对应方法返回 `ErrNotImplemented`（保持阶段三空壳一致）。

### 3.2 数据模型 `backend/internal/model/analysis.go`（新）
```go
// 内容分析 — 爆款内容列表项
type ContentItem struct {
    ID         int64   `json:"id"`
    Title      string  `json:"title"`
    Form       string  `json:"form"`       // 教程/测评/创意短片/挑战赛
    Topic      string  `json:"topic"`      // 滑雪/潜水/骑行/旅行/Vlog
    Views      int64   `json:"views"`
    Engagement float64 `json:"engagement"` // 互动率 %
    IsHit      bool    `json:"isHit"`      // 是否爆款
}

// 市场洞察 — 竞品对比项
type Competitor struct {
    Name       string  `json:"name"`
    Category   string  `json:"category"`
    Buzz       int64   `json:"buzz"`       // 声量
    Growth     float64 `json:"growth"`     // 增速 %
    Sentiment  float64 `json:"sentiment"`  // 好感度 %
}

// 品牌分析 — 合作品牌效果项
type PartnerBrand struct {
    Name       string  `json:"name"`
    Industry   string  `json:"industry"`
    Contents   int     `json:"contents"`   // 合作内容数
    Exposure   int64   `json:"exposure"`   // 曝光
    Engagement int64   `json:"engagement"` // 互动
    ROI        float64 `json:"roi"`        // 互动 ROI
}

// 品牌分析 — 高频词
type TagItem struct {
    Word   string  `json:"word"`
    Weight float64 `json:"weight"`
}
```
字段与现有 `KpiCard`/`ViewsTrendPoint`/`PlatformShare`/`TrackPerformance`/`AgeShare` 完全复用（分布类端点沿用既有类型，前端零适配）。

### 3.3 Mock 数据 `backend/internal/mock/analysis_data.go`（新）
- 三个子包函数（或同一文件分组）：`Content*` / `Market*` / `Brand*`，各 6 个，与接口一一对应。
- 生成规模：内容列表 ~15 条；竞品榜单 ~6 条；合作品牌 ~8 条；分布类聚合自列表或常量。
- KPI 各 5 张，形状 `[]model.KpiCard`（与达人页一致）。

### 3.4 Service `backend/internal/service/{content,market,brand}_service.go`（新）
- 三个 `*Service` 各注入 `source.DataSource`，6 方法透传 source 并补 `ctx`；adapter 报错 → 返回 error 由 handler 转 `fail`。

### 3.5 Handler `backend/internal/api/handler/{content,market,brand}.go`（新）
- 各 6 方法 `c *gin.Context`，解析 `Filter` querystring，调 service，`ok(c, data)` / `fail(c, err)`。
- 复用同包已定义的 `bindFilter` / `fail`。

### 3.6 路由 `backend/internal/api/router/router.go`
受保护组下新增三组共 18 路由：
```
GET /api/content/{kpi,trend,forms,topics,durations,list}
GET /api/market/{kpi,trend,competitors,regions,prices,list}
GET /api/brand/{kpi,trend,platforms,sentiment,keywords,list}
```
`main.go` 改动：构造三个 Service 并传入 `router.New(...)`（签名相应增加 `contentSvc`/`marketSvc`/`brandSvc` 参数，置于 `creatorSvc` 之后）。

## 4. 前端设计

### 4.1 通用列表组件 `frontend/src/components/DataTable.vue`（新）
列配置驱动，避免三页各造一个表：
```js
const props = defineProps({
  columns: { type: Array, required: true }, // [{ key, label, sortable?, width?, align?, formatter? }]
  rows:    { type: Array, default: () => [] },
  searchable: { type: Boolean, default: false },
  rowKey:  { type: String, default: 'id' }
})
```
- 内部 `el-table` + `el-table-column` `v-for` 动态渲染；`sortable` 列交给 Element Plus 前端排序。
- `searchable` 时顶部一个 `el-input` 按所有 `label`/字符串列过滤。
- 复用现有暗色主题样式（参考 `TopCreatorsTable` 的 `::v-deep(.el-table)` 处理）。
- **向后兼容**：`TopCreatorsTable` 保持不变（达人页继续用），本组件只服务于三页新列表。

### 4.2 Store（新 3 个，镜像 `creator.js`）
- `stores/content.js` / `stores/market.js` / `stores/brand.js`
- state：`kpi / trend / dist1 / dist2 / dist3 / list / loading / usedFallback`
- `loadAll(filter)`：并发 GET 6 端点；成功写 ref；拦截器拒非 JSON + 数组守卫 → 回退 `fallback-data.js`。
- `setFilter`：更新 filter 并重拉。

### 4.3 Fallback `frontend/src/api/fallback-data.js`
新增三组共 18 个数据集：`contentKpi/contentTrend/contentForms/contentTopics/contentDurations/contentList`、`market*` 同理、`brand*` 同理（与后端 mock 同口径，保证后端不可达时演示完整）。

### 4.4 API 模块（新 3 个）
- `api/content.js` / `api/market.js` / `api/brand.js`，方法名**对齐各页路由**（content：`kpi/trend/forms/topics/durations/list`；market：`kpi/trend/competitors/regions/prices/list`；brand：`kpi/trend/platforms/sentiment/keywords/list`）。
- store 的 `loadAll` 按「第 5 节板块明细」的顺序，把三个分布端点结果分别写入 `dist1/dist2/dist3`（页无关字段名），例如 content：`dist1←forms`、`dist2←topics`、`dist3←durations`。

### 4.5 视图（新 3 个）
- `views/ContentAnalysis.vue` / `views/MarketInsights.vue` / `views/BrandAnalysis.vue`
- 外壳镜像 `CreatorAnalysis.vue` 的网格系统（`.page` padding、`.kpi-row`、`grid-row-2/3/full`、`minmax(0,Xfr)` 铁律）。
- 板块顺序统一（每页 KPI×5 + 趋势 + 3 分布 + 1 总表）：
  - KPI 行（5 卡）
  - `grid-row-2`：趋势(2fr) + 分布①(1fr)
  - `grid-row-2`：分布②(1fr) + 分布③(1fr)
  - `grid-row-full`：总表（DataTable 占宽）
  - 品牌分析「高频词」为轻量 tag 云，置于分布③所在行右侧或直接并入分布区，不单独占一行。
- 复用组件：`KpiCard`（零改）、`TrendChart`（零改）、`PlatformDonut`（centerLabel prop）、`TrackBarChart`（零改）、`AgeDonut`（centerLabel/centerValue prop）、`DataTable`（新）。

### 4.6 路由 `frontend/src/router/index.js`
三处 `PlaceholderView` 改指向对应 Analysis 视图：
```
/content  → ContentAnalysis
/market   → MarketInsights
/brand    → BrandAnalysis
```

## 5. 三页板块明细（数据与语义）

### ① 内容分析
| 板块 | 端点 | 组件 | 数据 |
|---|---|---|---|
| KPI×5 | `/content/kpi` | KpiCard | 内容总数 / 平均播放 / 爆款率 / 平均互动率 / 周更新频次 |
| 趋势 | `/content/trend` | TrendChart | 近30天内容播放量（ViewsTrendPoint） |
| 形式分布 | `/content/forms` | PlatformDonut(centerLabel="内容形式") | 教程/测评/创意短片/挑战赛 占比 |
| 主题分布 | `/content/topics` | TrackBarChart | 滑雪/潜水/骑行/旅行/Vlog 播放聚合 |
| 时长段 | `/content/durations` | AgeDonut(centerLabel="时长段") | ≤30s / 30-60s / 1-3min / 3min+ 占比 |
| 总表 | `/content/list` | DataTable | 爆款内容：标题/形式/主题/播放/互动/爆款标记 |

### ② 市场洞察
| 板块 | 端点 | 组件 | 数据 |
|---|---|---|---|
| KPI×5 | `/market/kpi` | KpiCard | 品类规模 / 品类增速 / Insta360 市占 / 在榜竞品数 / 行业声量 |
| 趋势 | `/market/trend` | TrendChart | 品类声量趋势 |
| 竞品占比 | `/market/competitors` | PlatformDonut(centerLabel="竞品占比") | Insta360/GoPro/DJI/其他 |
| 地域热度 | `/market/regions` | TrackBarChart | 华东/华南/华北/海外…（**不用地图**） |
| 价格段 | `/market/prices` | AgeDonut(centerLabel="价格段") | <1000 / 1000-3000 / 3000-5000 / 5000+ |
| 总表 | `/market/list` | DataTable | 竞品对比：竞品/品类/声量/增速/好感度 |

### ③ 品牌分析
| 板块 | 端点 | 组件 | 数据 |
|---|---|---|---|
| KPI×5 | `/brand/kpi` | KpiCard | 品牌声量 / 好感度 / 合作品牌数 / 内容互动 ROI / 搜索指数 |
| 趋势 | `/brand/trend` | TrendChart | 品牌声量趋势 |
| 平台分布 | `/brand/platforms` | PlatformDonut(centerLabel="平台分布") | 抖音/B站/小红书/微博 |
| 情感分布 | `/brand/sentiment` | AgeDonut(centerLabel="情感分布") | 正面/中性/负面 |
| 高频词 | `/brand/keywords` | tags 样式（轻量，不建组件） | 高频词云（TagItem） |
| 总表 | `/brand/list` | DataTable | 合作品牌效果：品牌/行业/合作内容数/曝光/互动/ROI |

## 6. 接口契约（响应体）
统一 `{ "code": 0, "data": ... }`；Filter 通过 querystring 多选：`date_range / regions / tracks / platforms / age_bands`（mock 接受，阶段三生效）。

KPI 示例（三页同形状）：
```
GET /api/content/kpi → { "code":0, "data":[
  {key:"total", label:"内容总数", value:"1,284", raw:1284, deltaPct:12.3, deltaUp:true, description:"较上期"},
  {key:"avg_views", label:"平均播放", value:"86.4K", raw:86400, deltaPct:5.1, deltaUp:true},
  {key:"hit_rate", label:"爆款率", value:"8.2%", raw:8.2, deltaPct:1.4, deltaUp:true},
  {key:"engage", label:"平均互动率", value:"7.3%", raw:7.3, deltaPct:0.3, deltaUp:false},
  {key:"freq", label:"周更新频次", value:"42", raw:42, deltaPct:6.0, deltaUp:true}
]}
```

## 7. 数据流与错误处理
1. 各 Analysis 视图 `onMounted` → 对应 store `loadAll(currentFilter)`。
2. store 并发 GET 6 端点；任一成功即写对应 ref。
3. `request.js` 拦截器：`typeof data === 'string'` → reject → 上层回退 fallback。
4. store 对数组/对象类数据做 `Array.isArray` / 类型守卫，异常不写 ref，保持 fallback。
5. 筛选变化 → `setFilter` → 重新 `loadAll`。

## 8. 测试与验收

### 后端（用户在放行网络执行，沙箱 Go 被网络拦截）
- 新增 `source/analysis_adapter_test.go`：MockAdapter 18 方法返回非 nil；三平台空壳返回 `ErrNotImplemented`。
- 新增 `mock/analysis_data_test.go`：KPI 各 5 条、列表长度、分布类占比和≈100。
- 新增 `content/market/brand_service_test.go` 与 `handler/*_test.go`（TDD 先写后实现）。
- 验收：`cd backend && go build ./... && go test ./...` 全绿。
- curl 抽查：`GET /api/content/kpi` 返回 `{code:0,...}`；`SOURCE=douyin` 时返回 502 `ErrNotImplemented`。

### 前端（本环境可验证）
- `vite build` 通过（三页独立 chunk）。
- Edge 无头 `--dump-dom`：三页 KPI 卡渲染、图表 canvas 存在、DataTable 行存在、零页面级控制台错误。
- 路由：TopBar 三个 tab 激活高亮；访问 `/#/content`、`/#/market`、`/#/brand` 渲染对应概览。

### 可复用模板交付物
- 更新 `docs/superpowers/分析页模板说明.md`：补充 `DataTable` 通用列表用法与三页数据集约定。

## 9. 文件变更清单
后端（新增/改）：
- `internal/model/analysis.go`（新：ContentItem / Competitor / PartnerBrand / TagItem）
- `internal/mock/analysis_data.go`（新：18 函数）
- `internal/service/source/adapter.go`（改：接口 +18 方法）
- `internal/service/source/mock_adapter.go`（改：+18 实现）
- `internal/service/source/{douyin,bilibili,xiaohongshu}_adapter.go`（改：+18 返回 ErrNotImplemented）
- `internal/service/{content,market,brand}_service.go`（新）
- `internal/api/handler/{content,market,brand}.go`（新）
- `internal/api/router/router.go`（改：`New` +3 参数 +18 路由）
- `backend/main.go`（改：构造三 Service 并传入 `router.New`）
- 对应 `*_test.go`（新）

前端（新增/改）：
- `src/components/DataTable.vue`（新）
- `src/api/{content,market,brand}.js`（新）
- `src/api/fallback-data.js`（改：+3 组 18 数据集）
- `src/stores/{content,market,brand}.js`（新）
- `src/views/{ContentAnalysis,MarketInsights,BrandAnalysis}.vue`（新）
- `src/router/index.js`（改：3 处改指向）
- `docs/superpowers/分析页模板说明.md`（改：补 DataTable 用法）

## 10. 风险与注意
- 后端 Go 改动无法在沙箱编译（网络拦截），与达人分析同，需用户本地 `go build`/`go test` 验收。
- `DataSource` 接口已较大（原 9 insight + 6 creator + 18 本设计 = 33 方法）；保持模式统一优先于接口瘦身，后续若膨胀可拆子接口（不在本设计范围）。
- `DataTable` 与 `TopCreatorsTable` 并存：前者通用、后者达人专用；达人页不迁移，避免扰动已上线页面。
- 地域分布明确**不用地图组件**，规避中国地图合规风险（项目铁律约束）。

## 11. 后续（本设计不覆盖）
- 自定义看板：用户自选组件拼装，形态不同，单独设计。
- 阶段三真实 API：填充三个空壳 adapter + 落 ClickHouse；列表类端点届时接真实分页。
- 逐对象下钻：内容/竞品/品牌详情页，需另起设计。
