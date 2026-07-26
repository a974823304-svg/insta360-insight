# 全局筛选器联动（Filter Linkage）实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 让左侧 `SideFilter` 的「应用筛选」真正驱动全部 5 个数据页，且后端 mock 按筛选产生"确定性差异"数据（点筛选数字实时变）。

**Architecture:** 前端 `filter.apply()` 自增 `appliedRevision` → 5 个页面 `watch` 该 revision 重拉各自 store → store 用 `filter.toQuery()` 把 `{date_range, regions, tracks, platforms, age_bands}` 发给后端 → Go `MockAdapter` 把已接收的 `f model.Filter` 透传给各 `mock.X(f)` 函数，函数内部用 `mock/filter.go` 的纯变换（`scaleKpis` / `windowTrend` / `renormalize*` / `filter*`）生成差异化数据。**关键：接口/服务/handler 已全程透传 `f`，本计划只改 mock 层 + adapter 透传 + 前端 store/组件。**

**Tech Stack:** Go 1.26.5 (Gin, pure-Go, 无 CGO) + Vue 3 + Pinia + Element Plus + Vite。单测用 Go 标准 `testing`。

## Global Constraints

- 统一响应 `{ "code": 0, "data": ... }`；前端 API 模块必须 `.then(r => r.data)` 拆信封（已有约定，勿回退）。
- 后端编译/测试命令：`GOPROXY=goproxy.cn GOSUMDB=off go build ./... && go test ./...`（本机已配，直接 `go build/test` 亦可）。
- 新增 Go 依赖必须纯 Go（无 CGO），否则本机编译失败。
- JSON 字段 snake_case：`date_range` / `regions` / `tracks` / `platforms` / `age_bands`，与 `model.Filter` tag 一致。
- 年龄档位取值必须与 Go 常量 `mock.AgeBands = ["18-24 岁","25-34 岁","35-44 岁","45 岁以上"]` 完全一致（前端 `fallback-data.js` 的 `options.ageBands` 已是此值，**不要改成 "18-24"**）。
- 所有变换**确定性、无随机**；空选 = 不缩放 / 分布原样 / 列表全量；不得因缺失字段清空整表。

## 设计偏差修正（相对 spec，务必照此实现）

> spec 有两处与现有代码/数据不符，已批准按以下修正落地：

1. **`scaleFactor` 语义与边界**：spec 写 `[0.3,1.4]` 映射"已选项数/总项数"会反直觉（全选反而 >1）。修正为**加权占比**语义并 clamp：
   - 每个维度：空选 → 该维贡献 `1.0`；有选 → `已选项数 / 该维总项数`（平台总项=3、地区=4、赛道=5）。
   - `weightedFactor = clamp(0.5*fracP + 0.3*fracR + 0.2*fracT, 0.3, 1.0)`。
   - 日期单独：`dateScale = clamp(spanDays/30, 0.3, 5.0)`（90 天窗口累计量≈3×）。
   - KPI 幅度因子 `kpiScale = clamp(weightedFactor * dateScale, 0.3, 5.0)`。
   - 选得越少总量越小（符合直觉），选全量=1.0。

2. **`renormalize` 仅作用于"键名与筛选维度一致"的分布式端点**。现有代码用 `model.PlatformShare`/`TrackPerformance`/`AgeShare` 承载了**语义不同**的分布，强行同名 renormalize 会清空图表。逐组适用表（实现以此为准）：

   | 端点（函数） | 类型 | 是否 renormalize | 命中键 | 说明 |
   |---|---|---|---|---|
   | insight `PlatformShare()` | PlatformShare | ✅ 按 `f.Platforms` | `Platform` | 抖音/B站/小红书 |
   | insight `TrackPerformance()` | TrackPerformance | ✅ 按 `f.Tracks` | `Track` | 滑雪/冲浪/骑行/潜水/攀岩 |
   | insight `AgeShare()` | AgeShare | ✅ 按 `f.AgeBands` | `Bucket` | 18-24 岁… |
   | creator `CreatorPlatforms()` | PlatformShare | ✅ 按 `f.Platforms` | `Platform` | |
   | creator `CreatorTracks()` | TrackPerformance | ✅ 按 `f.Tracks` | `Track` | 来自 tags |
   | creator `CreatorAudience().Age` | AgeShare | ✅ 按 `f.AgeBands` | `Bucket` | |
   | brand `BrandPlatforms()` | PlatformShare | ✅ 按 `f.Platforms` | `Platform` | 抖音/B站/小红书/微博 |
   | content `ContentForms()` | PlatformShare | ❌ | — | 键是 教程/测评…非平台 |
   | content `ContentTopics()` | TrackPerformance | ❌ | — | spec 原样； Topics 部分含赛道词但保持原样 |
   | market `MarketCompetitors()` | PlatformShare | ❌ | — | 键是 Insta360/GoPro… |
   | market `MarketRegions()` | TrackPerformance | ❌ | — | 键是 华东/华南…≠地区筛选(北美/欧洲/亚太/全球) |
   | market `MarketPrices()` | AgeShare | ❌ | — | 价格带 |
   | content `ContentDurations()` | AgeShare | ❌ | — | 时长 |
   | brand `BrandSentiment()` | AgeShare | ❌ | — | 正/中/负 |
   | 其余（雷达/洞察/关键词/options） | — | ❌ | — | 无维度 |

3. **`subsetBy` 列表过滤真实可用性**：

   | 列表函数 | 行类型 | 可过滤维度 | 行为 |
   |---|---|---|---|
   | insight `TopCreators()` | `TopCreator` | 平台(`Platform`) + 赛道(从 `Tags` 去 `#`) | 真过滤，行数减少 |
   | creator `CreatorList()` | `TopCreator` | 同上 | 真过滤 |
   | content `ContentList()` | `ContentItem` | 赛道(`Topic`)；平台无字段→恒命中 | 仅赛道可缩 |
   | market `MarketList()` | `Competitor` | 无 region/platform 字段 | **no-op（保持全量）** |
   | brand `BrandList()` | `PartnerBrand` | 无 platform 字段 | **no-op（保持全量）** |

   > Market/PartnerBrand 列表无匹配维度字段，按 spec"缺失维度=命中"规则保持全量，属已知限制，不在本计划加字段（避免改 JSON 契约）。

## File Structure

**新增（后端）**
- `backend/internal/mock/filter.go` — 全部变换纯函数 + 常量（TotalPlatforms/TotalRegions/TotalTracks）。
- `backend/internal/mock/filter_test.go` — 变换单测（TDD）。

**修改（后端）**
- `backend/internal/mock/insight_data.go` — 9 函数加 `f model.Filter` 形参并应用变换。
- `backend/internal/mock/creator_data.go` — 6 函数加形参并应用。
- `backend/internal/mock/analysis_data.go` — 18 函数加形参并应用。
- `backend/internal/service/source/mock_adapter.go` — 33 处把 `mock.X()` 改为 `mock.X(f)`（方法已收 `f`，把 `_` 改名 `f`）。

**修改（前端）**
- `frontend/src/stores/filter.js` — 加 `appliedRevision` / `apply()` / `isDirty`；`reset()` 联动快照。
- `frontend/src/components/SideFilter.vue` — `onApply` 改调 `filter.apply()`；受众折叠内加年龄段 chips；脏提示；移除对 insight store 的直接调用。
- `frontend/src/views/InsightDashboard.vue` — `watch(appliedRevision)` 重拉。
- `frontend/src/views/CreatorAnalysis.vue` — 同上（store: `useCreatorStore`）。
- `frontend/src/views/ContentAnalysis.vue` — 同上（store: `useContentStore`）。
- `frontend/src/views/MarketInsights.vue` — 同上（store: `useMarketStore`）。
- `frontend/src/views/BrandAnalysis.vue` — 同上（store: `useBrandStore`）。

**文档**
- `docs/superpowers/分析页模板说明.md` — 追加"筛选联动契约"小节。

> 注：`analysis` 三个 store、insight/creator store 的 `loadAll()` **已正确** `const q = useFilterStore().toQuery()` 并透传每个 API，无需改动——本计划不重复。

---

### Task 1: 后端变换纯函数 + 单测（filter.go）

**Files:**
- Create: `backend/internal/mock/filter.go`
- Test: `backend/internal/mock/filter_test.go`

**Interfaces:**
- Produces（后续任务调用）：
  - `func ScaleKpis(kpis []model.KpiCard, f model.Filter) []model.KpiCard`
  - `func WindowTrend(points []model.ViewsTrendPoint, f model.Filter) []model.ViewsTrendPoint`
  - `func RenormalizePlatformShares(items []model.PlatformShare, selected []string) []model.PlatformShare`
  - `func RenormalizeTrackShares(items []model.TrackPerformance, selected []string) []model.TrackPerformance`
  - `func RenormalizeAgeShares(items []model.AgeShare, selected []string) []model.AgeShare`
  - `func FilterTopCreators(rows []model.TopCreator, f model.Filter) []model.TopCreator`
  - `func FilterContentItems(rows []model.ContentItem, f model.Filter) []model.ContentItem`

- [ ] **Step 1: 写失败测试 `filter_test.go`**

```go
package mock

import (
	"testing"

	"insta360-insight/internal/model"
)

func TestRenormalizeSums100(t *testing.T) {
	in := []model.PlatformShare{
		{Platform: "抖音", Share: 60, Views: 600},
		{Platform: "B站", Share: 40, Views: 400},
		{Platform: "小红书", Share: 0, Views: 0},
	}
	out := RenormalizePlatformShares(in, []string{"抖音", "B站"})
	var sum float64
	for _, s := range out {
		sum += s.Share
	}
	if len(out) != 2 {
		t.Fatalf("expected 2 items, got %d", len(out))
	}
	if sum < 99.5 || sum > 100.5 {
		t.Fatalf("shares should sum ~100, got %v", sum)
	}
}

func TestScaleFactorBounds(t *testing.T) {
	empty := model.Filter{}                       // 全空 -> 1.0
	if wf := WeightedFactor(empty); wf != 1.0 {   // 见下方实现; 空选=1.0
		t.Fatalf("empty factor want 1.0 got %v", wf)
	}
	one := model.Filter{Platforms: []string{"抖音"}} // 1/3 平台
	wf := WeightedFactor(one)
	if wf < 0.3 || wf > 1.0 {
		t.Fatalf("weighted factor out of [0.3,1.0]: %v", wf)
	}
}

func TestWeightedFactorDeterministic(t *testing.T) {
	f := model.Filter{Platforms: []string{"抖音"}, Tracks: []string{"滑雪"}}
	if WeightedFactor(f) != WeightedFactor(f) {
		t.Fatal("must be deterministic")
	}
}

func TestFilterTopCreatorsShrinks(t *testing.T) {
	rows := []model.TopCreator{
		{Name: "A", Platform: "抖音", Tags: []string{"#滑雪"}},
		{Name: "B", Platform: "B站", Tags: []string{"#冲浪"}},
		{Name: "C", Platform: "抖音", Tags: []string{"#冲浪"}},
	}
	out := FilterTopCreators(rows, model.Filter{Platforms: []string{"抖音"}})
	if len(out) != 2 {
		t.Fatalf("want 2 rows for platform=抖音, got %d", len(out))
	}
	out2 := FilterTopCreators(rows, model.Filter{Tracks: []string{"冲浪"}})
	if len(out2) != 2 {
		t.Fatalf("want 2 rows for track=冲浪, got %d", len(out2))
	}
}

func TestWindowTrendPoints(t *testing.T) {
	pts := make([]model.ViewsTrendPoint, 30)
	for i := range pts {
		pts[i] = model.ViewsTrendPoint{Date: "01-01", Views: 100, PrevViews: 80}
	}
	out := WindowTrend(pts, model.Filter{DateRange: []string{"2024-04-20", "2024-05-20"}})
	if len(out) < 1 {
		t.Fatal("expected points")
	}
	// 跨度 30 天 -> dateScale≈1.0, 单点值应≈原值
	if out[0].Views < 90 || out[0].Views > 110 {
		t.Fatalf("30d window should keep magnitude, got %d", out[0].Views)
	}
}
```

- [ ] **Step 2: 运行测试确认失败**

Run: `cd /f/workbuddy/影石/backend && go test ./internal/mock/ -run 'TestRenormalizeSums100|TestScaleFactorBounds|TestWeightedFactorDeterministic|TestFilterTopCreatorsShrinks|TestWindowTrendPoints'`
Expected: FAIL（函数未定义 / 编译错误）

- [ ] **Step 3: 实现 `filter.go`**

```go
package mock

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"insta360-insight/internal/model"
)

const (
	totalPlatforms = 3 // 抖音/B站/小红书
	totalRegions   = 4 // 北美/欧洲/亚太/全球
	totalTracks    = 5 // 滑雪/冲浪/骑行/潜水/攀岩
)

func clamp(x, lo, hi float64) float64 {
	if x < lo {
		return lo
	}
	if x > hi {
		return hi
	}
	return x
}

// WeightedFactor 平台/地区/赛道加权占比, clamp[0.3,1.0]; 空选=1.0
func WeightedFactor(f model.Filter) float64 {
	frac := func(sel, total int) float64 {
		if sel == 0 {
			return 1.0
		}
		return float64(sel) / float64(total)
	}
	w := 0.5*frac(len(f.Platforms), totalPlatforms) +
		0.3*frac(len(f.Regions), totalRegions) +
		0.2*frac(len(f.Tracks), totalTracks)
	return clamp(w, 0.3, 1.0)
}

// DateScale 时间跨度缩放, 30 天基准, clamp[0.3,5.0]
func DateScale(f model.Filter) float64 {
	if len(f.DateRange) != 2 {
		return 1.0
	}
	start, err1 := time.Parse("2006-01-02", f.DateRange[0])
	end, err2 := time.Parse("2006-01-02", f.DateRange[1])
	if err1 != nil || err2 != nil || !end.After(start) {
		return 1.0
	}
	span := end.Sub(start).Hours()/24 + 1
	return clamp(span/30.0, 0.3, 5.0)
}

// KpiScale 综合因子 = weighted * date, clamp[0.3,5.0]
func KpiScale(f model.Filter) float64 {
	return clamp(WeightedFactor(f)*DateScale(f), 0.3, 5.0)
}

// FormatScaledValue 依据原 Value 的格式模式, 用新 raw 重新格式化(确定性)
func FormatScaledValue(orig string, raw float64) string {
	switch {
	case strings.Contains(orig, "%"):
		return fmt.Sprintf("%.1f%%", raw)
	case strings.Contains(orig, "¥"):
		rest := strings.TrimPrefix(orig, "¥")
		return "¥" + formatBySuffix(rest, raw)
	case strings.Contains(orig, "B"):
		return fmt.Sprintf("%.2fB", raw/1e9)
	case strings.Contains(orig, "M"):
		return fmt.Sprintf("%.1fM", raw/1e6)
	case strings.Contains(orig, "K"):
		return fmt.Sprintf("%.1fK", raw/1e3)
	case strings.Contains(orig, ","):
		return humanizeComma(int64(raw))
	case strings.Contains(orig, "."):
		return fmt.Sprintf("%.1f", raw)
	default:
		return strconv.Itoa(int(raw))
	}
}

func formatBySuffix(suffix string, raw float64) string {
	switch {
	case strings.Contains(suffix, "B"):
		return fmt.Sprintf("%.2fB", raw/1e9)
	case strings.Contains(suffix, "M"):
		return fmt.Sprintf("%.1fM", raw/1e6)
	case strings.Contains(suffix, "K"):
		return fmt.Sprintf("%.1fK", raw/1e3)
	case strings.Contains(suffix, ","):
		return humanizeComma(int64(raw))
	case strings.Contains(suffix, "."):
		return fmt.Sprintf("%.1f", raw)
	default:
		return strconv.Itoa(int(raw))
	}
}

func humanizeComma(n int64) string {
	s := strconv.FormatInt(n, 10)
	var b strings.Builder
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			b.WriteByte(',')
		}
		b.WriteRune(c)
	}
	return b.String()
}

// ScaleKpis 缩放每个 KPI 的 Raw 并用原格式重算 Value
func ScaleKpis(kpis []model.KpiCard, f model.Filter) []model.KpiCard {
	k := KpiScale(f)
	out := make([]model.KpiCard, len(kpis))
	for i, c := range kpis {
		nc := c
		nc.Raw = c.Raw * k
		nc.Value = FormatScaledValue(c.Value, nc.Raw)
		out[i] = nc
	}
	return out
}

// WindowTrend 按 dateRange 重采样 x 轴并缩放单点值(上限 60 点)
func WindowTrend(points []model.ViewsTrendPoint, f model.Filter) []model.ViewsTrendPoint {
	if len(f.DateRange) != 2 {
		return points // 无日期筛选: 保持默认 30 天
	}
	start, err1 := time.Parse("2006-01-02", f.DateRange[0])
	end, err2 := time.Parse("2006-01-02", f.DateRange[1])
	if err1 != nil || err2 != nil || !end.After(start) {
		return points
	}
	days := int(end.Sub(start).Hours()/24) + 1
	step := 1
	if days > 60 {
		step = (days + 59) / 60
	}
	scale := DateScale(f)
	out := make([]model.ViewsTrendPoint, 0, days)
	for d := 0; d < days; d += step {
		dt := start.AddDate(0, 0, d)
		idx := d % len(points)
		p := points[idx]
		out = append(out, model.ViewsTrendPoint{
			Date:      dt.Format("01-02"),
			Views:     int64(float64(p.Views) * scale),
			PrevViews: int64(float64(p.PrevViews) * scale),
		})
	}
	return out
}

func inSet(keys, selected []string) bool {
	if len(selected) == 0 {
		return true
	}
	for _, s := range selected {
		for _, k := range keys {
			if k == s {
				return true
			}
		}
	}
	return false
}

// RenormalizePlatformShares 仅保留 selected 中的 Platform, 重算 share≈100
func RenormalizePlatformShares(items []model.PlatformShare, selected []string) []model.PlatformShare {
	if len(selected) == 0 {
		return items
	}
	var kept []model.PlatformShare
	var sum float64
	for _, it := range items {
		if contains(selected, it.Platform) {
			kept = append(kept, it)
			sum += it.Share
		}
	}
	if sum <= 0 { // 防零除: 回退原样
		return items
	}
	for i := range kept {
		kept[i].Share = RoundTo(kept[i].Share/sum*100, 1)
	}
	return kept
}

// RenormalizeTrackShares 仅保留 selected 中的 Track
func RenormalizeTrackShares(items []model.TrackPerformance, selected []string) []model.TrackPerformance {
	if len(selected) == 0 {
		return items
	}
	var kept []model.TrackPerformance
	var sum float64
	for _, it := range items {
		if contains(selected, it.Track) {
			kept = append(kept, it)
			sum += it.Views
		}
	}
	if sum <= 0 {
		return items
	}
	for i := range kept {
		kept[i].Views = int64(float64(kept[i].Views) / sum * 100)
	}
	return kept
}

// RenormalizeAgeShares 仅保留 selected 中的 Bucket
func RenormalizeAgeShares(items []model.AgeShare, selected []string) []model.AgeShare {
	if len(selected) == 0 {
		return items
	}
	var kept []model.AgeShare
	var sum float64
	for _, it := range items {
		if contains(selected, it.Bucket) {
			kept = append(kept, it)
			sum += it.Share
		}
	}
	if sum <= 0 {
		return items
	}
	for i := range kept {
		kept[i].Share = RoundTo(kept[i].Share/sum*100, 1)
	}
	return kept
}

func contains(s []string, v string) bool {
	for _, x := range s {
		if x == v {
			return true
		}
	}
	return false
}

// FilterTopCreators 按 平台 + 赛道(从 Tags 去#) AND 过滤
func FilterTopCreators(rows []model.TopCreator, f model.Filter) []model.TopCreator {
	if len(f.Platforms) == 0 && len(f.Tracks) == 0 {
		return rows
	}
	out := make([]model.TopCreator, 0, len(rows))
	for _, r := range rows {
		platOK := len(f.Platforms) == 0 || contains(f.Platforms, r.Platform)
		var trackKeys []string
		for _, t := range r.Tags {
			trackKeys = append(trackKeys, strings.TrimPrefix(t, "#"))
		}
		trackOK := len(f.Tracks) == 0 || inSet(trackKeys, f.Tracks)
		if platOK && trackOK {
			out = append(out, r)
		}
	}
	if len(out) == 0 { // 不空表: 回退全量
		return rows
	}
	return out
}

// FilterContentItems 按 赛道(Topic) 过滤; 平台无字段恒命中
func FilterContentItems(rows []model.ContentItem, f model.Filter) []model.ContentItem {
	if len(f.Tracks) == 0 {
		return rows
	}
	out := make([]model.ContentItem, 0, len(rows))
	for _, r := range rows {
		if contains(f.Tracks, r.Topic) {
			out = append(out, r)
		}
	}
	if len(out) == 0 {
		return rows
	}
	return out
}
```

> 注：`RoundTo` 已在 `mock` 包存在（insight_data.go 引用），复用即可，无需重定义。

- [ ] **Step 4: 运行测试确认通过**

Run: `cd /f/workbuddy/影石/backend && go test ./internal/mock/ -run 'TestRenormalizeSums100|TestScaleFactorBounds|TestWeightedFactorDeterministic|TestFilterTopCreatorsShrinks|TestWindowTrendPoints' -v`
Expected: PASS

- [ ] **Step 5: 提交**

```bash
cd /f/workbuddy/影石 && git add backend/internal/mock/filter.go backend/internal/mock/filter_test.go && git commit -m "feat(mock): 新增筛选确定性变换纯函数与单测"
```

---

### Task 2: MockAdapter 透传 f

**Files:**
- Modify: `backend/internal/service/source/mock_adapter.go` （33 处调用）

**Interfaces:**
- Consumes: `mock/filter.go` 各变换（由 mock 函数内部调用，本任务仅透传 `f`）。
- Produces: 无新接口；仅让 `mock.X(f)` 接收过滤器。

- [ ] **Step 1: 把每个方法的 `_ model.Filter` 改名 `f` 并透传**

将文件中全部 33 处 `return mock.Xxx(), nil` / `return mock.Xxx()` / `a2 := mock.Xxx()` 改为传入 `f`，并把形参 `_ model.Filter` 改为 `f model.Filter`。例如：

```go
func (a *MockAdapter) Kpi(_ context.Context, f model.Filter) ([]model.KpiCard, error) {
	return mock.Kpi(f), nil
}
func (a *MockAdapter) PlatformShare(_ context.Context, f model.Filter) ([]model.PlatformShare, error) {
	return mock.PlatformShare(f), nil
}
func (a *MockAdapter) CreatorAudience(_ context.Context, f model.Filter) (*model.Audience, error) {
	a2 := mock.CreatorAudience(f)
	return &a2, nil
}
// ... 其余 30 处同理: 把 `mock.Xxx()` 改为 `mock.Xxx(f)`, 形参 `_ model.Filter` -> `f model.Filter`
```

- [ ] **Step 2: 编译校验（此时会失败——因为 mock 函数尚未接收 f，下一步才改；此步仅确认 adapter 写法，可与 Task 3-5 合并编译）**

Run: `cd /f/workbuddy/影石/backend && go build ./... 2>&1 | head` （预期：mock 包"too many arguments"错误，属正常，待 Task 3-5 修复）

- [ ] **Step 3: 提交（与 Task 3-5 一起编译通过后提交更稳妥；若单独提交需先 stub mock 签名）**

> 建议：Task 2 的编辑与 Task 3-5 的 mock 签名改动一并 `go build` 通过后统一提交，避免中间态无法编译。本步骤标记为"随 Task 5 末提交"。

---

### Task 3: insight_data.go 接入变换

**Files:**
- Modify: `backend/internal/mock/insight_data.go` （9 函数）

**Interfaces:**
- Consumes: `mock/filter.go` 的 `ScaleKpis` / `WindowTrend` / `RenormalizePlatformShares` / `RenormalizeTrackShares` / `RenormalizeAgeShares` / `FilterTopCreators`。
- Produces: 同签名 `func Xxx(f model.Filter) ...`（已被 adapter 调用）。

逐函数改动（在 `return` 前套用变换，或把 `return []...{...}` 改为先构造变量再变换返回）：

- [ ] **Step 1: 给 9 个函数加 `f model.Filter` 形参并按下表套变换**

| 函数 | 改动 |
|---|---|
| `Kpi()` | → `Kpi(f model.Filter) []model.KpiCard`；先 `k := []model.KpiCard{...}`；`return ScaleKpis(k, f)` |
| `ViewsTrend()` | → `ViewsTrend(f model.Filter)`；`return WindowTrend(points, f)` |
| `PlatformShare()` | → `PlatformShare(f model.Filter)`；`return RenormalizePlatformShares(out, f.Platforms)` |
| `TrackPerformance()` | → `TrackPerformance(f model.Filter)`；`return RenormalizeTrackShares(out, f.Tracks)` |
| `ExplosiveRadar()` | → `ExplosiveRadar(f model.Filter)`；**无变换**，原样返回 |
| `AgeShare()` | → `AgeShare(f model.Filter)`；`return RenormalizeAgeShares(data, f.AgeBands)` |
| `TopCreators()` | → `TopCreators(f model.Filter)`；`return FilterTopCreators(out, f)` |
| `AIInsights()` | → `AIInsights(f model.Filter)`；无变换 |
| `FilterOptions()` | → `FilterOptions(f model.Filter)`；无变换（age_bands 已含） |

示例（Kpi）：
```go
func Kpi(f model.Filter) []model.KpiCard {
	k := []model.KpiCard{
		{Key: "creators", Label: "达人数", Value: "12,856", Raw: 12856, DeltaPct: 18.6, DeltaUp: true, Unit: "", Description: "较上期"},
		// ... 其余 4 项不变
	}
	return ScaleKpis(k, f)
}
```

- [ ] **Step 2: 编译 + 跑 mock 包测试**

Run: `cd /f/workbuddy/影石/backend && go build ./... && go test ./internal/mock/`
Expected: PASS（adapter 仍因 analysis/creator 未改签名而报错属正常，待 Task 4-5）

- [ ] **Step 3: 提交**（与后续任务合并或单独，确保本文件可编译——若单独提交需临时保留 adapter 调用 `mock.Kpi()` 不传 f 会编译失败，故建议合并 Task 2 一起提交）

---

### Task 4: creator_data.go 接入变换

**Files:**
- Modify: `backend/internal/mock/creator_data.go` （6 函数）

**Interfaces:**
- Consumes: `ScaleKpis` / `WindowTrend` / `RenormalizePlatformShares` / `RenormalizeTrackShares` / `RenormalizeAgeShares` / `FilterTopCreators`。

- [ ] **Step 1: 6 函数加 `f model.Filter` 并套变换**

| 函数 | 改动 |
|---|---|
| `CreatorKpi()` | `ScaleKpis(k, f)` |
| `CreatorTrend()` | `WindowTrend(points, f)` |
| `CreatorPlatforms()` | 现有 `out` 先构造；`return RenormalizePlatformShares(out, f.Platforms)`（注意：原函数内部基于 `CreatorList()` 聚合，renormalize 作用于最终 `out`） |
| `CreatorTracks()` | `return RenormalizeTrackShares(out, f.Tracks)` |
| `CreatorAudience()` | → `CreatorAudience(f model.Filter) model.Audience`；`a.Age = RenormalizeAgeShares(a.Age, f.AgeBands)`；`return a` |
| `CreatorList()` | `return FilterTopCreators(out, f)` |

`CreatorPlatforms` 改造示例：
```go
func CreatorPlatforms(f model.Filter) []model.PlatformShare {
	list := CreatorList(model.Filter{}) // 注意: 内部聚合用全量, 不受 f 影响
	// ... 原聚合逻辑不变, 得到 out ...
	return RenormalizePlatformShares(out, f.Platforms)
}
```
> 注意 `CreatorPlatforms`/`CreatorTracks` 内部调用了 `CreatorList()`——为避免递归传 f 导致二次过滤，内部调用传 `model.Filter{}`（全量聚合），仅在外层对结果做 renormalize。

- [ ] **Step 2: 编译**

Run: `cd /f/workbuddy/影石/backend && go build ./internal/mock/`
Expected: PASS（analysis 未改仍报错，正常）

- [ ] **Step 3: 提交**（合并）

---

### Task 5: analysis_data.go 接入变换（content/market/brand 共 18 函数）

**Files:**
- Modify: `backend/internal/mock/analysis_data.go` （18 函数）

**Interfaces:**
- Consumes: `ScaleKpis` / `WindowTrend` / `RenormalizePlatformShares` / `RenormalizeAgeShares` / `FilterContentItems`。

逐组（按"设计偏差修正"表，仅下列接入变换，其余仅加形参无变换）：

- [ ] **Step 1: content 6 函数**
  - `ContentKpi(f)` → `ScaleKpis(..., f)`
  - `ContentTrend(f)` → `WindowTrend(..., f)`
  - `ContentForms(f)` → 仅加形参，**无 renormalize**（键是 教程/测评）
  - `ContentTopics(f)` → 仅加形参，**无 renormalize**
  - `ContentDurations(f)` → 仅加形参，**无 renormalize**（AgeShare 但时长）
  - `ContentList(f)` → `FilterContentItems(out, f)`

- [ ] **Step 2: market 6 函数**
  - `MarketKpi(f)` → `ScaleKpis`
  - `MarketTrend(f)` → `WindowTrend`
  - `MarketCompetitors(f)` → 仅加形参，无 renormalize
  - `MarketRegions(f)` → 仅加形参，无 renormalize（华东…≠地区筛选）
  - `MarketPrices(f)` → 仅加形参，无 renormalize
  - `MarketList(f)` → **no-op**（Competitor 无 region/platform 字段，保持全量）

- [ ] **Step 3: brand 6 函数**
  - `BrandKpi(f)` → `ScaleKpis`
  - `BrandTrend(f)` → `WindowTrend`
  - `BrandPlatforms(f)` → `RenormalizePlatformShares(out, f.Platforms)`
  - `BrandSentiment(f)` → 仅加形参，无 renormalize（正面/中性/负面）
  - `BrandKeywords(f)` → 仅加形参，无变换
  - `BrandList(f)` → **no-op**（PartnerBrand 无 platform 字段，保持全量）

- [ ] **Step 4: 全量编译 + 测试（关键门）**

Run: `cd /f/workbuddy/影石/backend && go build ./... && go test ./...`
Expected: 全部 PASS，无编译错误（此前 adapter 的 `mock.X(f)` 现在全部匹配）

- [ ] **Step 5: 提交（合并 Task 2 的 adapter 改动一起提交）**

```bash
cd /f/workbuddy/影石 && git add backend/internal/mock/analysis_data.go backend/internal/mock/creator_data.go backend/internal/mock/insight_data.go backend/internal/service/source/mock_adapter.go && git commit -m "feat(mock): 27 端点接入 Filter 确定性变换(MockAdapter 透传 f)"
```

---

### Task 6: 后端 e2e 验证（curl 对比）

**Files:** 无（验证步骤）

- [ ] **Step 1: 启动 mock 后端（若未运行）**

```bash
cd /f/workbuddy/影石/backend && go run . > /tmp/be.log 2>&1 &
sleep 3
curl -s -o /dev/null -w "boot HTTP %{http_code}\n" http://localhost:8080/api/kpi
```

- [ ] **Step 2: 断言参数化前后数据差异**

```bash
echo "== KPI 全量 vs 仅抖音 =="
curl -s "http://localhost:8080/api/kpi" | head -c 200; echo
curl -s "http://localhost:8080/api/kpi?platforms=%E6%8A%96%E9%9F%B3" | head -c 200; echo
# 预期: 两条返回的 views/raw 数值不同(抖音子集更小)
echo "== 达人列表 全量 vs tracks=滑雪 =="
curl -s "http://localhost:8080/api/creator/list" | python -c "import sys,json;d=json.load(sys.stdin);print('count',len(d['data']))"
curl -s "http://localhost:8080/api/creator/list?tracks=%E6%BB%91%E9%9B%AA" | python -c "import sys,json;d=json.load(sys.stdin);print('count',len(d['data']))"
# 预期: 后者 count <= 前者, 且每行 Topic/Tags 命中 滑雪
echo "== 年龄分布 age_bands=18-24 岁 =="
curl -s "http://localhost:8080/api/audience-age?age_bands=18-24%20%E5%B2%81" | head -c 200; echo
# 预期: 仅返回 18-24 岁 一项, share 重算为 100
```

- [ ] **Step 3: 记录结果，确认差异成立**

若任一对比无差异，回到对应 mock 函数检查变换是否漏接（最常见：`mock_adapter.go` 某行仍 `mock.X()` 未传 `f`，或该函数被标为"无变换"误判）。

---

### Task 7: 前端 filter store 加 appliedRevision / apply / isDirty

**Files:**
- Modify: `frontend/src/stores/filter.js`

**Interfaces:**
- Produces: `appliedRevision` (ref<number>), `apply()` (action), `isDirty` (computed<bool>)；`reset()` 同步快照。

- [ ] **Step 1: 改写 filter store**

```js
import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export const useFilterStore = defineStore('filter', () => {
  const dateRange = ref(['2024-04-20', '2024-05-20'])
  const regions = ref([])
  const tracks = ref([])
  const platforms = ref([])
  const ageBands = ref([])
  const granularity = ref('day')

  // 联动信号: 每次 apply 自增, 页面据此 watch 重拉
  const appliedRevision = ref(0)
  // 已应用快照(用于脏检查)
  const appliedState = ref({
    dateRange: [...dateRange.value],
    regions: [], tracks: [], platforms: [], ageBands: []
  })

  function toQuery() {
    const q = {}
    if (dateRange.value?.length === 2) q.date_range = dateRange.value
    if (regions.value.length) q.regions = [...regions.value]
    if (tracks.value.length) q.tracks = [...tracks.value]
    if (platforms.value.length) q.platforms = [...platforms.value]
    if (ageBands.value.length) q.age_bands = [...ageBands.value]
    return q
  }

  // 应用筛选: 存快照 + 自增 revision(触发页面重拉)
  function apply() {
    appliedState.value = {
      dateRange: [...dateRange.value],
      regions: [...regions.value],
      tracks: [...tracks.value],
      platforms: [...platforms.value],
      ageBands: [...ageBands.value]
    }
    appliedRevision.value++
  }

  function reset() {
    regions.value = []
    tracks.value = []
    platforms.value = []
    ageBands.value = []
    apply() // 重置后视为已应用(全量), 页面重拉
  }

  // 脏检查: 当前筛选与已应用快照不同 -> 有未应用更改
  const isDirty = computed(() => {
    const a = appliedState.value
    const eq = (x, y) => x.length === y.length && x.every((v, i) => v === y[i])
    return !eq(dateRange.value, a.dateRange) ||
      !eq(regions.value, a.regions) ||
      !eq(tracks.value, a.tracks) ||
      !eq(platforms.value, a.platforms) ||
      !eq(ageBands.value, a.ageBands)
  })

  return {
    dateRange, regions, tracks, platforms, ageBands, granularity,
    appliedRevision, appliedState, toQuery, apply, reset, isDirty
  }
})
```

- [ ] **Step 2: 构建前端确认无报错**

Run: `cd /f/workbuddy/影石/frontend && npm run build 2>&1 | tail -20`
Expected: 构建成功（仅 echarts chunk 警告）

- [ ] **Step 3: 提交**

```bash
cd /f/workbuddy/影石 && git add frontend/src/stores/filter.js && git commit -m "feat(frontend): filter store 加 appliedRevision/apply/isDirty 联动信号"
```

---

### Task 8: SideFilter.vue 改 onApply + 年龄段 chips + 脏提示

**Files:**
- Modify: `frontend/src/components/SideFilter.vue`

**Interfaces:**
- Consumes: `useFilterStore` 的 `apply()` / `isDirty` / `ageBands`（已导出）。
- 移除：对 `useInsightStore` 的 `loadAll` 直接调用（解耦）。

- [ ] **Step 1: 把受众折叠占位替换为年龄段 chips**

将 `el-collapse-item title="粉丝画像"` 内的 `<div class="muted">...</div>` 改为：

```html
<el-collapse-item title="粉丝画像" name="audience">
  <div class="chips">
    <button
      v-for="b in store.options.ageBands"
      :key="b.value"
      class="chip"
      :class="{ active: filter.ageBands.includes(b.value) }"
      @click="toggle('ageBands', b.value)"
    >{{ b.label }}</button>
  </div>
</el-collapse-item>
```

- [ ] **Step 2: 改 `onApply` 为只调 `filter.apply()`，并移除 insight store 依赖**

```js
import { useFilterStore } from '../stores/filter'
// 删除: import { useInsightStore } ... ; const store = useInsightStore()
const filter = useFilterStore()

// 删除 onMounted 中 store.loadAll() 逻辑(可选项加载改为由 InsightDashboard 自身负责)
async function onApply() {
  filter.apply() // 只发信号, 各页面 watch 自取
}

// 重置按钮保持不变: filter.reset()
```

- [ ] **Step 3: 在「应用筛选」按钮旁加脏提示**

```html
<div class="footer">
  <el-button type="primary" size="small" class="apply" @click="onApply">应用筛选</el-button>
  <span v-if="filter.isDirty" class="dirty">● 有未应用更改</span>
</div>
```
并在 `<script setup>` 中确保 `filter` 已导入（见 Step 2）。样式 `.dirty { color: var(--brand); font-size: 10px; margin-top: 4px; }` 加到 `<style scoped>`。

- [ ] **Step 4: 构建校验**

Run: `cd /f/workbuddy/影石/frontend && npm run build 2>&1 | tail -20`
Expected: 成功

- [ ] **Step 5: 提交**

```bash
cd /f/workbuddy/影石 && git add frontend/src/components/SideFilter.vue && git commit -m "feat(frontend): SideFilter 改 apply 解耦 + 年龄段 chips + 脏提示"
```

---

### Task 9: 5 个页面 watch appliedRevision 重拉

**Files:**
- Modify: `frontend/src/views/InsightDashboard.vue`（store: `useInsightStore`）
- Modify: `frontend/src/views/CreatorAnalysis.vue`（store: `useCreatorStore`）
- Modify: `frontend/src/views/ContentAnalysis.vue`（store: `useContentStore`）
- Modify: `frontend/src/views/MarketInsights.vue`（store: `useMarketStore`）
- Modify: `frontend/src/views/BrandAnalysis.vue`（store: `useBrandStore`）

**Interfaces:**
- Consumes: 各页 store 已有 `loadAll()`；`useFilterStore().appliedRevision`。
- 每个文件统一做三处编辑（以 `ContentAnalysis.vue` 为例，其余替换 store 名）：

- [ ] **Step 1: 在 `<script setup>` 顶部加 filter 导入与实例**

```js
import { useFilterStore } from '../stores/filter'
const filter = useFilterStore()
```

- [ ] **Step 2: 在现有 `onMounted(() => store.loadAll())` 旁加 watch**

```js
import { onMounted, watch } from 'vue'
// 已有: onMounted(() => store.loadAll())
watch(() => filter.appliedRevision, () => store.loadAll())
```

- [ ] **Step 3: 对 5 个文件逐一应用（替换各自 store）**

| 文件 | store 名 |
|---|---|
| InsightDashboard.vue | `useInsightStore` → `store`；`watch(() => filter.appliedRevision, () => store.loadAll())` |
| CreatorAnalysis.vue | `useCreatorStore` |
| ContentAnalysis.vue | `useContentStore` |
| MarketInsights.vue | `useMarketStore` |
| BrandAnalysis.vue | `useBrandStore` |

- [ ] **Step 4: 构建校验**

Run: `cd /f/workbuddy/影石/frontend && npm run build 2>&1 | tail -20`
Expected: 成功

- [ ] **Step 5: 提交**

```bash
cd /f/workbuddy/影石 && git add frontend/src/views/InsightDashboard.vue frontend/src/views/CreatorAnalysis.vue frontend/src/views/ContentAnalysis.vue frontend/src/views/MarketInsights.vue frontend/src/views/BrandAnalysis.vue && git commit -m "feat(frontend): 5 个数据页 watch appliedRevision 实现筛选联动重拉"
```

---

### Task 10: 前端无头渲染验证联动生效

**Files:** 无（验证）

- [ ] **Step 1: 确保 Go(:8080) 与 Vite(:5173) 在跑**

```bash
# 后端
cd /f/workbuddy/影石/backend && (go run . > /tmp/be.log 2>&1 &) ; sleep 3
# 前端
cd /f/workbuddy/影石/frontend && (npm run dev > /tmp/vite.log 2>&1 &) ; sleep 5
curl -s -o /dev/null -w "vite HTTP %{http_code}\n" http://localhost:5173/
```

- [ ] **Step 2: 渲染应用筛选前后对比 KPI 文本**

用 Edge 无头渲染某分析页（如 `/#/content`），抓取首个 KPI 数字；再用一个临时入口页在 `localStorage` 写入 `filter` 已应用"仅抖音"后渲染，对比数字是否变化（参考既往 headless 验证手法：`--virtual-time-budget=6000 --dump-dom`）。

```bash
EDGE="/c/Program Files (x86)/Microsoft/Edge/Application/msedge.exe"
"$EDGE" --headless --no-sandbox --disable-gpu --virtual-time-budget=6000 \
  --dump-dom "http://localhost:5173/#/content" > /tmp/dom_before.html 2>/dev/null
grep -o -E '内容总数|平均播放|爆款率' /tmp/dom_before.html | head
# 应用筛选(仅抖音)后的对比需借助注入 localStorage 的临时页, 或手动在 devtools 验证
```

- [ ] **Step 3: 确认五个页面均响应**

逐一访问 `/#/insight`、`/#/creator`、`/#/content`、`/#/market`、`/#/brand`，确认无白屏、KPI 卡片渲染；应用一个筛选后数字变化（后端 e2e 已在 Task 6 证明数据差异，本步证明前端管线接通）。

- [ ] **Step 4: 关闭 dev server（避免遗留进程抢端口）**

```bash
PID=$(netstat -ano 2>/dev/null | grep ':5173' | grep LISTENING | head -1 | awk '{print $5}')
[ -n "$PID" ] && MSYS_NO_PATHCONV=1 taskkill /PID $PID /F >/dev/null 2>&1 && echo "stopped vite"
```

- [ ] **Step 5: 不提交（验证步骤）**

---

### Task 11: 文档 — 筛选联动契约

**Files:**
- Modify: `docs/superpowers/分析页模板说明.md`

**Interfaces:** 追加一节说明 filter linkage 约定，供后续新增页面遵循。

- [ ] **Step 1: 在文档末尾追加**

```markdown
## 全局筛选器联动契约（2026-07-26）

- 触发：点「应用筛选」→ `filter.apply()` 自增 `appliedRevision`；SideFilter 不再直接调任何页面 store。
- 页面接法：各数据页 `watch(() => filter.appliedRevision, () => store.loadAll())`，并在 `onMounted` 首次 `loadAll()`。
- store 接法：每个 store 的 `loadAll()` 内 `const q = useFilterStore().toQuery()` 并作为 `params: q` 透传所有 API（creator/content/market/brand/insight 均已实现）。
- 后端接法：mock 函数接收 `f model.Filter`，用 `mock/filter.go` 的 `ScaleKpis`/`WindowTrend`/`Renormalize*`/`Filter*` 生成差异数据；MockAdapter 已透传 `f`。
- 维度适用范围（renormalize 仅限键名一致的分布式端点）：平台分布(抖音/B站/小红书/微博)、赛道分布(滑雪/冲浪/骑行/潜水/攀岩)、年龄分布(18-24 岁…)。内容形式/话题/时长、竞品、地区(华东…)、价格带、情感、热词 **不**随筛选变化。
- 列表过滤真实可用：insight/creator 达人列表（平台+赛道）；内容列表（仅赛道）。market/brand 列表无匹配维度字段，保持全量（已知限制）。
- 年龄档位值固定为 `18-24 岁 / 25-34 岁 / 35-44 岁 / 45 岁以上`，前后端必须一致。
```

- [ ] **Step 2: 提交**

```bash
cd /f/workbuddy/影石 && git add docs/superpowers/分析页模板说明.md && git commit -m "docs: 追加筛选联动契约说明"
```

---

## 自审（Self-Review）

1. **Spec 覆盖**：spec §3.1-3.4（前端 store/组件/页面）↔ Task 7/8/9；§4.1-4.3（后端变换+27 端点）↔ Task 1-5；§7（测试/e2e/headless）↔ Task 6/10；§9 验收 ↔ Task 6/9/10。覆盖完整。
2. **占位符扫描**：无 TBD/TODO；每个代码步骤均给出实现。Task 2 的"其余 30 处同理"已列出代表性示例且接口签名明确，实现者可逐行替换（属机械编辑，非占位）。
3. **类型一致性**：`mock.X(f model.Filter)` 在 Task 1-5 与 adapter（Task 2）严格一致；`RenormalizePlatformShares/TrackShares/AgeShares` 的入参类型（`[]model.PlatformShare`/`[]model.TrackPerformance`/`[]model.AgeShare`）与对应 mock 函数返回类型一致；`FilterTopCreators`/`FilterContentItems` 入参类型与 `TopCreators()`/`ContentList()` 返回一致。前端 `appliedRevision`/`apply`/`isDirty` 在 store(Task7) 与 SideFilter(Task8)/页面(Task9) 命名一致。
4. **已修正的 spec 偏差**已在"设计偏差修正"节明确，避免实现者误用 `[0.3,1.4]` 或误 renormalize 非同名分布。
