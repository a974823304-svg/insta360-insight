# 全局筛选器联动设计（Filter Linkage）

- 日期：2026-07-26
- 状态：设计已确认（待评审后进入实现计划）

## 1. 目标与背景

平台已有 5 个数据页（数据洞察 / 达人分析 / 内容分析 / 市场洞察 / 品牌分析），左侧 `SideFilter` 在全局布局中常驻。但当前存在两个硬伤：

1. **前端不联动**：`SideFilter` 的「应用筛选」只调用了 insight store 的 `loadAll()`，其余四个分析页从不因筛选变化重拉数据——点了筛选，分析页纹丝不动。
2. **后端不变数据**：五个 handler 组都调用了 `bindFilter(c)` 并把 `model.Filter` 传给了 service，但 `mock` 层拿到的 Filter **完全没被用来改变输出**（KPI / 趋势 / 分布 / 列表都是写死的静态值）。所以即便前端重拉，数字也不变，看起来像坏了。

本设计目标：**让筛选真正驱动全部 5 个页面的数据，且数据按筛选产生"确定性差异"**，使 demo 实时可见变化。

确认的关键决策：
- **联动深度 = 真联动**：后端 mock 按筛选生成确定性差异数据（非仅通管道）。
- **年龄段 UI = 四档**：18-24 / 25-34 / 35-44 / 45+。
- **触发方式 = 保留「应用筛选」按钮**：点 chip 不实时刷新，点按钮才统一重拉（避免每次切换狂打后端、竞态）。

## 2. 架构与数据流

```
SideFilter(用户操作)
   │  onApply()
   ▼
filter.apply()  ──►  filter.appliedRevision++  (filter store 内)
   │
   ├─► InsightDashboard   watch(appliedRevision) ─► insightStore.loadAll()
   ├─► CreatorAnalysis    watch(appliedRevision) ─► creatorStore.loadAll()
   ├─► ContentAnalysis    watch(appliedRevision) ─► contentStore.loadAll()
   ├─► MarketInsights     watch(appliedRevision) ─► marketStore.loadAll()
   └─► BrandAnalysis      watch(appliedRevision) ─► brandStore.loadAll()

每个 store.loadAll():
   q = filter.toQuery()           // {date_range, regions, tracks, platforms, age_bands}
   → api.*(q)  ──►  GET /api/...?platforms=抖音&tracks=骑行...
   → Go handler bindFilter(c) → service → mock(应用 Filter 变换) → 差异化 JSON
   → 前端 store 用真实数据替换 fallback
```

**解耦要点**：`SideFilter` 不再感知任何具体页面 store；它只负责改 filter store 状态并触发 `apply()`。各页面组件自己 watch `appliedRevision` 决定何时重拉。新增页面也只需"watch + loadAll"即可自动获得联动能力。

## 3. 前端改动

### 3.1 `stores/filter.js`
- 新增 `appliedRevision = ref(0)`：每次 `apply()` 自增，作为"筛选已应用"的信号。
- 新增 `appliedState`：快照当前 `{dateRange, regions, tracks, platforms, ageBands}`，用于脏检查。
- 新增 `apply()`：`appliedState = 深拷贝当前筛选；appliedRevision++`。
- 新增 `isDirty` 计算属性：对比当前筛选与 `appliedState`，任一维度不同则为 `true`（供按钮显示"未应用"提示）。
- `reset()` 保持：清空 regions/tracks/platforms/ageBands（已含 ageBands）。重置后也应视为"未应用"——`reset()` 内部一并更新 `appliedState` 快照或直接触发 `apply()` 语义；**决定**：`reset()` 清空后调用一次 `apply()` 使 revision 自增、页面重拉全部数据，且 `isDirty=false`。
- 导出新增 `appliedRevision, apply, isDirty`。

### 3.2 `SideFilter.vue`
- 把折叠面板里「粉丝画像」的占位文字替换为**真实年龄段 chips**：数据源 `filter.ageBands` 选项取自 `store.options.ageBands`（见 4.2）；四档 18-24 / 25-34 / 35-44 / 45+，点击 toggle 写入 `filter.ageBands`。
- 「达人属性」折叠面板保留占位（后续扩展），不加逻辑。
- `onApply` 改为**只调 `filter.apply()`**（删除原先对 insight store 的直接调用）。
- 在「应用筛选」按钮旁：当 `filter.isDirty` 为真时显示提示（如 `● 有未应用更改` 小字或按钮高亮）。
- 时间段 / 市场 / 赛道 / 平台四个区块保持现状（已正确双向绑定 `filter.*`）。

### 3.3 五个页面组件
- `InsightDashboard.vue` / `CreatorAnalysis.vue` / `ContentAnalysis.vue` / `MarketInsights.vue` / `BrandAnalysis.vue`：
  - `onMounted(() => store.loadAll())` 确保进入即加载（若已有则保留）。
  - `watch(() => filter.appliedRevision, () => store.loadAll())` 实现联动重拉。`filter` 来自 `useFilterStore()`。
- 页面组件需 `import { useFilterStore } from '../stores/filter'` 并在 setup 内取实例。

### 3.4 五个 store
- `insight.js`：已正确在 `loadAll()` 内 `const f = useFilterStore(); const q = f.toQuery()` 并透传——保持不变。
- `creator.js` / `content.js` / `market.js` / `brand.js`：**逐项核对** `loadAll()` 是否读取 `useFilterStore().toQuery()` 并作为 `params: q` 传给每个 API 调用；若有遗漏（例如某端点没传 q 或没引 filter store），补齐。这是当前分析页"看似不联动"的潜在第二层原因。

## 4. 后端改动（路线 A：集中式 Filter 变换层）

### 4.1 新增 `backend/internal/mock/filter.go`
纯函数集合（无副作用、易测）：

- `scaleFactor(f model.Filter) float64`
  - 综合 平台 / 地区 / 赛道 选择，合成一个确定性"体量系数"。
  - 规则：每个维度的"已选项数 / 总项数"取加权（平台权重 0.5、地区 0.3、赛道 0.2），映射到区间 **[0.3, 1.4]**；空选（全选或没选）→ 1.0。
  - 用固定公式（非随机）保证同输入同输出。
  - 边界：clamp 到 [0.3, 1.4]，避免极端值或零除。

- `renormalize(items []NamedShare, selected []string) []NamedShare`
  - `NamedShare` 为 `{Name string; Share float64}`（或复用现有分布结构）。
  - 若 `selected` 非空：仅保留 `Name ∈ selected` 的项，Share 重算为 `原值/Σ选中原值*100`，使合计≈100。
  - 若 `selected` 为空：原样返回（全选）。
  - 用于平台占比、地区分布、年龄段分布等。

- `subsetBy(list []Row, f model.Filter) []Row`
  - 对列表行按 平台 / 赛道 / 地区 过滤：
    - 行需带可识别字段（平台、赛道/话题、地区）。各 list 结构不同，由调用方传入"取维度键"的闭包或先在主结构上加字段。
    - 任一维度非空时，行必须命中该维度所选集合；三个维度是 AND 关系。
  - 行数确定性减少（符合直觉：筛得越细，列表越短）。

- `windowTrend(trend []Point, dateRange []string) []Point`
  - 按 `dateRange` 跨度（天数）缩放趋势总额：`point.Value *= scaleFactor 的日期分量`（天数/30 近似，clamp）。
  - x 轴标签：用确定性日序列从 `dateRange[0]` 到 `dateRange[1]` 生成（mock 不依赖真实日历库，按天步进即可）。
  - 点数为日期跨度（上限裁剪到合理值，如 ≤ 60 点，避免超长）。

### 4.2 筛选选项
- `frontend/src/api/fallback-data.js` 的 `options`：补 `ageBands: [{value:'18-24',label:'18-24'}, {value:'25-34',label:'25-34'}, {value:'35-44',label:'35-44'}, {value:'45+',label:'45+'}]`。
- Go `Insight.FilterOptions` handler（及 `service.InsightService`/`mock` 的 options 构造）：在返回里补 `age_bands` 四个选项，结构与现有 regions/tracks/platforms 一致（Go 侧 snake_case：`age_bands`）。

### 4.3 27 个端点接入变换
每个 mock 函数返回前，按自身数据类型调用对应变换（**不改变现有字段结构，只改数值/子集**）。

**变换分配原则**：
- `scale`（整体体量缩放）与 `window`（趋势按时段）适用于**所有**端点组的 KPI 与趋势。
- `renormalize`（子集+重算占比≈100）**仅作用于"维度与某筛选维度同名"的分布式端点**。没有对应筛选维度的分布（内容形式/话题/时长、价格带、情感、品牌热词等）**保持原样**——它们本来就不该被筛选。
- `subsetBy`（列表过滤）作用于列表类端点，按该列表**实际带有的**维度字段过滤。
- `age_bands` 仅作用于受众 / 年龄类分布。

**逐组分配**：

| 端点组 | KPI | 趋势 | 分布类（renormalize 维度） | 列表类（subsetBy 维度） | 受众/年龄 |
|---|---|---|---|---|---|
| insight (9) | scale | window | platform-distribution→平台；track-performance→赛道 | top-creators→平台/赛道 | audience-age→年龄段 |
| creator (6) | scale | window | platforms→平台；tracks→赛道 | creator list→平台/赛道 | audience→年龄段 |
| content (6) | scale | window | 形式/话题/时长→原样（无匹配维度） | content list→平台/赛道 | — |
| market (6) | scale | window | regions→地区；competitors/prices→原样 | market list→地区 | — |
| brand (6) | scale | window | platforms→平台；sentiment/keywords→原样 | brand list→平台 | 受众→年龄段 |

> 说明：地区(`regions`)筛选对 insight/content/brand 无同名分布端点，因此只通过 `scaleFactor` 影响这些页的整体体量；只有 market 的 `regions` 端点会做 renormalize。平台/赛道筛选同理——只 renormalize 同名分布，其余页走 scale。

- KPI 的 `raw` 整数与派生 `value` 展示字符串需一并缩放（注意 `value` 是已格式化字符串如 "86.4K"，缩放后需重新格式化；**约定**：mock 内部以 `raw` 为唯一真值，缩放 `raw` 后再统一格式化，避免字符串二次处理漂移）。
- 分布类的 `share` 字段走 `renormalize`；`bucket`/`name` 不变。
- 列表类的行过滤走 `subsetBy`；若某行无对应维度字段，则该维度视为"命中"（不参与该维度过滤），保证不会误清空整张表。

### 4.4 不改动
- `bindFilter`、handler、`model.Filter` 已就绪，不动。
- 不在 service 层做后处理（保持 mock 纯净、语义清晰）。

## 5. 确定性规则小结（各维度怎么变）

- **平台 / 地区 / 赛道**：选中后整体体量按 `scaleFactor` 缩放；分布类只留选中项并 `renormalize`≈100；列表按选中项过滤。
- **时间段 `date_range`**：按天数跨度缩放累计量；趋势 x 轴改为该区间确定性日序列。
- **年龄段 `age_bands`**：仅作用于受众/年龄分布（AgeDonut、达人/品牌受众），过滤并 `renormalize`；轻微影响互动率（可选，低优先）。
- 全部确定性、无随机；空选=系数 1、分布原样、列表全量。

## 6. 错误处理与边界

- 筛选解析已安全：空数组 = 全部，不会因缺参报错。
- `scaleFactor` 输出 clamp 至 [0.3, 1.4]，杜绝极端值 / 零除。
- `renormalize` 在选中集为空或 Σ=0 时回退原样（不崩）。
- `subsetBy` 不应清空整表：缺失维度字段的行视为命中。
- 前端：后端不可达时仍走 fallback（静态），体验不退化；`isDirty` 仅在本地状态层面提示，不影响数据获取。

## 7. 测试与验证

### 7.1 后端单测（新增 `mock/filter_test.go`）
- `TestRenormalizeSums100`：给若干 share + 选中集，断言输出合计 ≈100（容差 0.5）。
- `TestScaleFactorBounds`：全空→1.0；极端单选→落在 [0.3,1.4]。
- `TestScaleFactorDeterministic`：同输入两次调用结果相等。
- `TestSubsetBy`：构造带维度字段的行，断言过滤后仅含命中行、不空表。
- `TestWindowTrendPoints`：给定 dateRange，断言点数 = min(跨度, 上限) 且值随跨度缩放。
- 现有 `go build ./... && go test ./...` 应全绿。

### 7.2 端到端（curl 对比）
- `curl '/api/kpi'` vs `curl '/api/kpi?platforms=抖音'` → 断言 `raw` 不同。
- `curl '/api/creator/list?tracks=骑行'` vs 无参 → 断言列表行数减少且行赛道命中"骑行"。
- `curl '/api/audience-age?age_bands=18-24'` → 断言仅返回 18-24 项且 share 重算。

### 7.3 前端无头验证
- 启动 Go(:8080) + Vite(:5173)；Edge 无头 `--dump-dom` 渲染某分析页。
- 应用一个筛选（如只选平台"抖音"）前后，对比 DOM 中 KPI 数字文本是否变化，确认联动生效。

## 8. 范围与非目标

**范围内**：上述前端 5 页联动 + 年龄段 UI + 后端 27 端点 Filter 变换 + 单测/e2e。
**非目标（本次不做）**：
- 真实 API（抖音/B站/小红书）接入——仍走 MockAdapter。
- 「达人属性」维度展开（保留折叠占位）。
- 实时（每次点 chip 即刷新）模式——保留按钮触发。
- 筛选条件进 URL（刷新保持）——可选未来项。
- AI 引擎接通（独立模块）。

## 9. 验收标准

1. 进入任意分析页，调整筛选并点「应用筛选」，该页 KPI / 趋势 / 分布 / 列表**数字实时变化**，且与所选维度一致（如只选"抖音"→ 平台分布仅抖音、总量下缩放）。
2. 五个页面**全部**响应筛选（不再只有首页）。
3. 年龄段四档可在侧栏选择并影响受众类图表。
4. 未点「应用筛选」前，按钮显示"有未应用更改"提示；点「重置」后回到全量且提示消失。
5. 后端 `go test ./...` 全绿；e2e curl 证明参数化前后数据差异；无头渲染证明前端联动。
6. 后端不可达时前端仍可用（fallback），不报错。
