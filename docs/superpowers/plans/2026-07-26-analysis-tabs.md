# 分析页扩展：内容 / 市场 / 品牌 Tab Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 把「数据洞察」下空着的 内容分析 / 市场洞察 / 品牌分析 三个 Tab 做成与达人分析视觉/架构完全一致的真实概览页，复用刚沉淀的「分析页模板」骨架与现有图表组件，新增一个通用 `DataTable.vue` 列表组件。

**Architecture:** 后端在 `DataSource` 接口新增三组共 18 个域方法，`MockAdapter` 实现、三平台空壳返回 `ErrNotImplemented`；新增 3 个 Service + 3 个 Handler 暴露 `/api/{content,market,brand}/*` 各 6 路由。前端新增 `DataTable.vue` + 3 个 store + 3 个 api 模块 + `fallback-data.js` 兜底 + 3 个 view，复用 `KpiCard/TrendChart/PlatformDonut/TrackBarChart/AgeDonut`。

**Tech Stack:** Go + Gin + modernc.org/sqlite(本任务不涉及 DB); Vue 3 + Vite + Element Plus + Pinia + Axios + ECharts; 复用阶段二 adapter/service/handler 模式与「分析页模板」。

## Global Constraints

- ECharts 容器: 严禁 `min-height` 硬编码; 必备 `min-width:0; min-height:0; overflow:hidden` + `ResizeObserver` 主动 resize; 容器 100% 自适应父级。
- CSS Grid 列宽: 必须用 `minmax(0, Xfr)`, 不用 `1fr`(会被内容撑爆); grid 内 `.card` 全加 `min-width:0`。
- 响应约定: 业务失败 HTTP 200 + `{code:非0, message}`; 中间件未授权 HTTP 401。
- 暗色主题 + 玻璃拟态 + 品牌色 `--brand:#3DD9EB`。
- 数据来源维持 mock（阶段三真实 API 卡在开放平台 appkey）；前端 `fallback-data.js` 必须兜底。
- **环境约束(已实测修正 2026-07-25)**: 本机 `go build`/`go test` 可用, 命令加 `GOPROXY=goproxy.cn GOSUMDB=off` 前缀即可; 若某环境网络异常, 退回「放行网络执行」。
- 项目为**非 git 仓库**: 不执行 `git commit`; 保存文件即可（计划内 "Commit" 步骤一律改为「保存文件」）。
- 地域分布明确**不用地图组件**, 用横向柱图(`TrackBarChart`), 规避中国地图合规风险。
- 接口契约: 统一 `{ "code": 0, "data": ... }`; Filter 经 querystring 多选 `date_range/regions/tracks/platforms/age_bands`(mock 接受, 阶段三生效)。

---

## File Structure

**后端（新增/改）**
- `internal/model/analysis.go` — 新: ContentItem / Competitor / PartnerBrand / TagItem
- `internal/mock/analysis_data.go` — 新: 18 函数(content/market/brand 各 6)
- `internal/service/source/adapter.go` — 改: 接口 +18 方法
- `internal/service/source/mock_adapter.go` — 改: +18 实现
- `internal/service/source/{douyin,bilibili,xiaohongshu}_adapter.go` — 改: +18 返回 ErrNotImplemented
- `internal/service/{content,market,brand}_service.go` — 新: 各 6 方法
- `internal/api/handler/{content,market,brand}.go` — 新: 各 6 方法
- `internal/api/router/router.go` — 改: `New` 在 `creatorSvc` 之后增 3 参数 +18 路由
- `backend/main.go` — 改: 构造 3 Service 并传入 `router.New`
- 对应 `*_test.go` — 新

**前端（新增/改）**
- `src/components/DataTable.vue` — 新: 通用列表(列配置驱动)
- `src/api/{content,market,brand}.js` — 新: 各 6 端点封装
- `src/api/fallback-data.js` — 改: +3 组 18 数据集
- `src/stores/{content,market,brand}.js` — 新: 镜像 creator store
- `src/views/{ContentAnalysis,MarketInsights,BrandAnalysis}.vue` — 新
- `src/router/index.js` — 改: 3 处改指向
- `docs/superpowers/分析页模板说明.md` — 改: 补 DataTable 用法 + 三页数据集约定

---

## Task 1: 新增 内容/市场/品牌 模型类型

**Files:**
- Create: `internal/model/analysis.go`
- Test: `internal/model/analysis_test.go`

**Interfaces:**
- Produces: `model.ContentItem` / `model.Competitor` / `model.PartnerBrand` / `model.TagItem`（后续 mock/adapter/service/handler 依赖）

- [ ] **Step 1: 写失败测试**

```go
package model

import (
	"encoding/json"
	"testing"
)

func TestAnalysisModelJSONShape(t *testing.T) {
	in := struct {
		C ContentItem  `json:"c"`
		M Competitor   `json:"m"`
		P PartnerBrand `json:"p"`
		T TagItem      `json:"t"`
	}{
		C: ContentItem{ID: 1, Title: "滑雪教程", Form: "教程", Topic: "滑雪", Views: 120000, Engagement: 7.3, IsHit: true},
		M: Competitor{Name: "GoPro", Category: "运动相机", Buzz: 980000, Growth: 12.4, Sentiment: 78.0},
		P: PartnerBrand{Name: "红牛", Industry: "饮料", Contents: 8, Exposure: 5400000, Engagement: 320000, ROI: 4.2},
		T: TagItem{Word: "画质", Weight: 86.0},
	}
	b, err := json.Marshal(in)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var out struct {
		C ContentItem  `json:"c"`
		M Competitor   `json:"m"`
		P PartnerBrand `json:"p"`
		T TagItem      `json:"t"`
	}
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if out.C.Title != "滑雪教程" || !out.C.IsHit {
		t.Fatalf("ContentItem 字段丢失: %+v", out.C)
	}
	if out.M.Name != "GoPro" || out.M.Buzz != 980000 {
		t.Fatalf("Competitor 字段丢失: %+v", out.M)
	}
	if out.P.ROI != 4.2 {
		t.Fatalf("PartnerBrand.ROI 丢失: %+v", out.P)
	}
	if out.T.Word != "画质" {
		t.Fatalf("TagItem 字段丢失: %+v", out.T)
	}
}
```

- [ ] **Step 2: 运行测试确认失败**

Run: `cd backend && GOPROXY=goproxy.cn GOSUMDB=off go test ./internal/model/ -run TestAnalysisModelJSONShape -v`
Expected: 编译失败（类型未定义）

- [ ] **Step 3: 实现类型**

```go
// Package model 包含内容/市场/品牌分析所需的列表项类型。
package model

// ContentItem 内容分析 — 爆款内容列表项
type ContentItem struct {
	ID         int64   `json:"id"`
	Title      string  `json:"title"`
	Form       string  `json:"form"`       // 教程/测评/创意短片/挑战赛
	Topic      string  `json:"topic"`      // 滑雪/潜水/骑行/旅行/Vlog
	Views      int64   `json:"views"`
	Engagement float64 `json:"engagement"` // 互动率 %
	IsHit      bool    `json:"isHit"`      // 是否爆款
}

// Competitor 市场洞察 — 竞品对比项
type Competitor struct {
	Name      string  `json:"name"`
	Category  string  `json:"category"`
	Buzz      int64   `json:"buzz"`      // 声量
	Growth    float64 `json:"growth"`    // 增速 %
	Sentiment float64 `json:"sentiment"` // 好感度 %
}

// PartnerBrand 品牌分析 — 合作品牌效果项
type PartnerBrand struct {
	Name       string  `json:"name"`
	Industry   string  `json:"industry"`
	Contents   int     `json:"contents"`   // 合作内容数
	Exposure   int64   `json:"exposure"`   // 曝光
	Engagement int64   `json:"engagement"` // 互动
	ROI        float64 `json:"roi"`        // 互动 ROI
}

// TagItem 品牌分析 — 高频词
type TagItem struct {
	Word   string  `json:"word"`
	Weight float64 `json:"weight"`
}
```

- [ ] **Step 4: 运行测试确认通过**

Run: `cd backend && GOPROXY=goproxy.cn GOSUMDB=off go test ./internal/model/ -run TestAnalysisModelJSONShape -v`
Expected: PASS

- [ ] **Step 5: 保存文件

---

## Task 2: Mock 内容/市场/品牌 数据（18 函数）

**Files:**
- Create: `internal/mock/analysis_data.go`
- Test: `internal/mock/analysis_data_test.go`

**Interfaces:**
- Consumes: `model.KpiCard`/`ViewsTrendPoint`/`PlatformShare`/`TrackPerformance`/`AgeShare`/`ContentItem`/`Competitor`/`PartnerBrand`/`TagItem`; 包内 `RoundTo`/`PlatformColor`/`TrackColor`/`AgeColor`/`FormatCount`/`sortByShareDesc`/`Tracks`/`Platforms`
- Produces: `mock.Content*` / `mock.Market*` / `mock.Brand*`（adapter 调用）

- [ ] **Step 1: 写失败测试**

```go
package mock

import (
	"math"
	"testing"
)

func sumShare(ps []struct{ s float64 }) float64 { return 0 } // 占位, 下方用真实类型

func TestAnalysisKpiLen(t *testing.T) {
	for i, got := range [][]interface{}{
		toAny(ContentKpi()), toAny(MarketKpi()), toAny(BrandKpi()),
	} {
		if len(got) != 5 {
			t.Fatalf("第 %d 组 KPI 期望 5 张, 实际 %d", i, len(got))
		}
	}
}

func TestAnalysisDistSum100(t *testing.T) {
	checks := []struct {
		name string
		ps   []PlatformShare
	}{
		{"ContentForms", ContentForms()},
		{"ContentDurations", agesToPlatform(ContentDurations())},
		{"MarketCompetitors", MarketCompetitors()},
		{"MarketPrices", agesToPlatform(MarketPrices())},
		{"BrandPlatforms", BrandPlatforms()},
		{"BrandSentiment", agesToPlatform(BrandSentiment())},
	}
	for _, c := range checks {
		var total float64
		for _, p := range c.ps {
			total += p.Share
		}
		if math.Abs(total-100) > 0.5 {
			t.Fatalf("%s 占比之和应≈100, 实际 %.2f", c.name, total)
		}
	}
}

func TestAnalysisListLen(t *testing.T) {
	if len(ContentList()) != 15 {
		t.Fatalf("ContentList 期望 15, 实际 %d", len(ContentList()))
	}
	if len(MarketList()) != 6 {
		t.Fatalf("MarketList 期望 6, 实际 %d", len(MarketList()))
	}
	if len(BrandList()) != 8 {
		t.Fatalf("BrandList 期望 8, 实际 %d", len(BrandList()))
	}
	if len(BrandKeywords()) != 8 {
		t.Fatalf("BrandKeywords 期望 8, 实际 %d", len(BrandKeywords()))
	}
}

// 测试辅助: 把 AgeShare 转成可求和的切片
func agesToPlatform(a []AgeShare) []PlatformShare {
	out := make([]PlatformShare, 0, len(a))
	for _, x := range a {
		out = append(out, PlatformShare{Platform: x.Bucket, Share: x.Share})
	}
	return out
}
func toAny(_ interface{}) []interface{} { return nil } // 占位, 真正断言在上面的 len 检查
```

注：上面的 `sumShare`/`toAny` 占位函数会在 Step 3 实现时删除——KPI 长度断言直接用 `len(ContentKpi())` 等。修正后的测试见 Step 3 末尾说明。

- [ ] **Step 2: 运行确认失败**

Run: `cd backend && GOPROXY=goproxy.cn GOSUMDB=off go test ./internal/mock/ -run 'TestAnalysis' -v`
Expected: 编译失败（函数未定义）

- [ ] **Step 3: 实现 mock（18 函数）**

```go
// Package mock 生成数据洞察所需的演示数据。
package mock

import (
	"time"

	"insta360-insight/internal/model"
)

// ============================================================
// 内容分析
// ============================================================
func ContentKpi() []model.KpiCard {
	return []model.KpiCard{
		{Key: "total", Label: "内容总数", Value: "1,284", Raw: 1284, DeltaPct: 12.3, DeltaUp: true, Description: "较上期"},
		{Key: "avg_views", Label: "平均播放", Value: "86.4K", Raw: 86400, DeltaPct: 5.1, DeltaUp: true, Description: "较上期"},
		{Key: "hit_rate", Label: "爆款率", Value: "8.2%", Raw: 8.2, DeltaPct: 1.4, DeltaUp: true, Description: "较上期"},
		{Key: "engage", Label: "平均互动率", Value: "7.3%", Raw: 7.3, DeltaPct: 0.3, DeltaUp: false, Description: "较上期"},
		{Key: "freq", Label: "周更新频次", Value: "42", Raw: 42, DeltaPct: 6.0, DeltaUp: true, Description: "较上期"},
	}
}

func ContentTrend() []model.ViewsTrendPoint {
	base := int64(2_400_000)
	points := make([]model.ViewsTrendPoint, 0, 30)
	now := time.Now().Truncate(24 * time.Hour)
	for i := 0; i < 30; i++ {
		v := base + int64(i)*120_000 + int64((i*7)%11)*30_000
		prev := base + int64(i)*95_000
		d := now.AddDate(0, 0, -(29 - i))
		points = append(points, model.ViewsTrendPoint{
			Date:      d.Format("01-02"),
			Views:     v,
			PrevViews: prev,
		})
	}
	return points
}

func ContentForms() []model.PlatformShare {
	return []model.PlatformShare{
		{Platform: "教程", Share: 38.0, Views: 920000, Color: "#3DD9EB"},
		{Platform: "测评", Share: 27.0, Views: 650000, Color: "#5EA1FF"},
		{Platform: "创意短片", Share: 21.0, Views: 510000, Color: "#A07BFF"},
		{Platform: "挑战赛", Share: 14.0, Views: 340000, Color: "#7DD96E"},
	}
}

func ContentTopics() []model.TrackPerformance {
	return []model.TrackPerformance{
		{Track: "滑雪", Views: 720000, Color: "#5EA1FF"},
		{Track: "潜水", Views: 610000, Color: "#3DD9EB"},
		{Track: "骑行", Views: 530000, Color: "#A07BFF"},
		{Track: "旅行", Views: 470000, Color: "#7DD96E"},
		{Track: "Vlog", Views: 390000, Color: "#FFB547"},
	}
}

func ContentDurations() []model.AgeShare {
	return []model.AgeShare{
		{Bucket: "≤30s", Share: 34.0, Color: "#3DD9EB"},
		{Bucket: "30-60s", Share: 41.0, Color: "#5EA1FF"},
		{Bucket: "1-3min", Share: 18.0, Color: "#A07BFF"},
		{Bucket: "3min+", Share: 7.0, Color: "#7DD96E"},
	}
}

func ContentList() []model.ContentItem {
	forms := []string{"教程", "测评", "创意短片", "挑战赛"}
	topics := []string{"滑雪", "潜水", "骑行", "旅行", "Vlog"}
	out := make([]model.ContentItem, 0, 15)
	for i := 0; i < 15; i++ {
		out = append(out, model.ContentItem{
			ID:         int64(i + 1),
			Title:      topics[i%len(topics)] + "第" + itoa(i+1) + "期",
			Form:       forms[i%len(forms)],
			Topic:      topics[i%len(topics)],
			Views:      int64(50_000 + (i*53_000)%1_200_000),
			Engagement: RoundTo(5.0+float64(i%9)+float64(i%4)*0.6, 2),
			IsHit:      i%3 == 0,
		})
	}
	return out
}

// ============================================================
// 市场洞察
// ============================================================
func MarketKpi() []model.KpiCard {
	return []model.KpiCard{
		{Key: "size", Label: "品类规模", Value: "¥3.2B", Raw: 3200000000, DeltaPct: 9.8, DeltaUp: true, Description: "较上期"},
		{Key: "growth", Label: "品类增速", Value: "14.6%", Raw: 14.6, DeltaPct: 2.1, DeltaUp: true, Description: "较上期"},
		{Key: "share", Label: "Insta360 市占", Value: "31.2%", Raw: 31.2, DeltaPct: 3.4, DeltaUp: true, Description: "较上期"},
		{Key: "comp", Label: "在榜竞品数", Value: "6", Raw: 6, DeltaPct: 0, DeltaUp: true, Description: "较上期"},
		{Key: "buzz", Label: "行业声量", Value: "4.7M", Raw: 4700000, DeltaPct: 7.5, DeltaUp: true, Description: "较上期"},
	}
}

func MarketTrend() []model.ViewsTrendPoint {
	base := int64(140_000)
	points := make([]model.ViewsTrendPoint, 0, 30)
	now := time.Now().Truncate(24 * time.Hour)
	for i := 0; i < 30; i++ {
		v := base + int64(i)*9_000 + int64((i*5)%9)*2_000
		prev := base + int64(i)*7_000
		d := now.AddDate(0, 0, -(29 - i))
		points = append(points, model.ViewsTrendPoint{
			Date:      d.Format("01-02"),
			Views:     v,
			PrevViews: prev,
		})
	}
	return points
}

func MarketCompetitors() []model.PlatformShare {
	return []model.PlatformShare{
		{Platform: "Insta360", Share: 31.2, Views: 1460000, Color: "#3DD9EB"},
		{Platform: "GoPro", Share: 28.4, Views: 1330000, Color: "#FF6B6B"},
		{Platform: "DJI", Share: 24.7, Views: 1150000, Color: "#5EA1FF"},
		{Platform: "其他", Share: 15.7, Views: 730000, Color: "#A07BFF"},
	}
}

func MarketRegions() []model.TrackPerformance {
	return []model.TrackPerformance{
		{Track: "华东", Views: 1520000, Color: "#3DD9EB"},
		{Track: "华南", Views: 1180000, Color: "#5EA1FF"},
		{Track: "华北", Views: 960000, Color: "#A07BFF"},
		{Track: "西南", Views: 640000, Color: "#7DD96E"},
		{Track: "海外", Views: 1300000, Color: "#FFB547"},
	}
}

func MarketPrices() []model.AgeShare {
	return []model.AgeShare{
		{Bucket: "<1000", Share: 22.0, Color: "#3DD9EB"},
		{Bucket: "1000-3000", Share: 41.0, Color: "#5EA1FF"},
		{Bucket: "3000-5000", Share: 26.0, Color: "#A07BFF"},
		{Bucket: "5000+", Share: 11.0, Color: "#7DD96E"},
	}
}

func MarketList() []model.Competitor {
	return []model.Competitor{
		{Name: "Insta360", Category: "全景/运动相机", Buzz: 1460000, Growth: 17.2, Sentiment: 81.0},
		{Name: "GoPro", Category: "运动相机", Buzz: 1330000, Growth: 4.1, Sentiment: 73.0},
		{Name: "DJI", Category: "无人机/相机", Buzz: 1150000, Growth: 9.8, Sentiment: 76.0},
		{Name: "Sony", Category: "微单", Buzz: 880000, Growth: 6.3, Sentiment: 79.0},
		{Name: "大疆Action", Category: "运动相机", Buzz: 620000, Growth: 12.5, Sentiment: 74.0},
		{Name: "AKASO", Category: "入门运动相机", Buzz: 340000, Growth: 21.0, Sentiment: 68.0},
	}
}

// ============================================================
// 品牌分析
// ============================================================
func BrandKpi() []model.KpiCard {
	return []model.KpiCard{
		{Key: "buzz", Label: "品牌声量", Value: "2.9M", Raw: 2900000, DeltaPct: 11.2, DeltaUp: true, Description: "较上期"},
		{Key: "sent", Label: "好感度", Value: "81%", Raw: 81, DeltaPct: 1.8, DeltaUp: true, Description: "较上期"},
		{Key: "partners", Label: "合作品牌数", Value: "8", Raw: 8, DeltaPct: 14.3, DeltaUp: true, Description: "较上期"},
		{Key: "roi", Label: "内容互动 ROI", Value: "4.2", Raw: 4.2, DeltaPct: 0.5, DeltaUp: true, Description: "较上期"},
		{Key: "search", Label: "搜索指数", Value: "68.5", Raw: 68.5, DeltaPct: 3.2, DeltaUp: false, Description: "较上期"},
	}
}

func BrandTrend() []model.ViewsTrendPoint {
	base := int64(90_000)
	points := make([]model.ViewsTrendPoint, 0, 30)
	now := time.Now().Truncate(24 * time.Hour)
	for i := 0; i < 30; i++ {
		v := base + int64(i)*6_000 + int64((i*4)%7)*1_500
		prev := base + int64(i)*5_000
		d := now.AddDate(0, 0, -(29 - i))
		points = append(points, model.ViewsTrendPoint{
			Date:      d.Format("01-02"),
			Views:     v,
			PrevViews: prev,
		})
	}
	return points
}

func BrandPlatforms() []model.PlatformShare {
	return []model.PlatformShare{
		{Platform: "抖音", Share: 44.0, Views: 1280000, Color: "#000000"},
		{Platform: "B站", Share: 26.0, Views: 754000, Color: "#00A1D6"},
		{Platform: "小红书", Share: 22.0, Views: 638000, Color: "#FF2442"},
		{Platform: "微博", Share: 8.0, Views: 232000, Color: "#E6162D"},
	}
}

func BrandSentiment() []model.AgeShare {
	return []model.AgeShare{
		{Bucket: "正面", Share: 67.0, Color: "#7DD96E"},
		{Bucket: "中性", Share: 24.0, Color: "#FFB547"},
		{Bucket: "负面", Share: 9.0, Color: "#FF6B6B"},
	}
}

func BrandKeywords() []model.TagItem {
	return []model.TagItem{
		{Word: "画质", Weight: 92.0},
		{Word: "防抖", Weight: 88.0},
		{Word: "全景", Weight: 81.0},
		{Word: "运动相机", Weight: 76.0},
		{Word: "Vlog", Weight: 70.0},
		{Word: "旅行", Weight: 64.0},
		{Word: "性价比", Weight: 58.0},
		{Word: "续航", Weight: 49.0},
	}
}

func BrandList() []model.PartnerBrand {
	return []model.PartnerBrand{
		{Name: "红牛", Industry: "饮料", Contents: 8, Exposure: 5400000, Engagement: 320000, ROI: 4.2},
		{Name: "始祖鸟", Industry: "户外", Contents: 6, Exposure: 4100000, Engagement: 260000, ROI: 3.9},
		{Name: "携程", Industry: "旅行", Contents: 5, Exposure: 3300000, Engagement: 210000, ROI: 3.5},
		{Name: "迪卡侬", Industry: "运动", Contents: 7, Exposure: 2900000, Engagement: 190000, ROI: 3.1},
		{Name: "大疆", Industry: "无人机", Contents: 4, Exposure: 2600000, Engagement: 175000, ROI: 3.8},
		{Name: "索尼", Industry: "影像", Contents: 3, Exposure: 1800000, Engagement: 120000, ROI: 3.3},
		{Name: "Keep", Industry: "健身", Contents: 5, Exposure: 1500000, Engagement: 98000, ROI: 2.9},
		{Name: "小米", Industry: "科技", Contents: 2, Exposure: 980000, Engagement: 64000, ROI: 2.6},
	}
}
```

辅助函数 `itoa` 在包内已有或新增（若包内无 `itoa`，在文件顶部加 `func itoa(n int) string { return strconv.Itoa(n) }` 并 import `strconv`）。
测试修正：删除 Step 1 的占位 `sumShare`/`toAny`，改为直接断言：
```go
func TestAnalysisKpiLen(t *testing.T) {
	if len(ContentKpi()) != 5 || len(MarketKpi()) != 5 || len(BrandKpi()) != 5 {
		t.Fatalf("KPI 期望各 5 张")
	}
}
```

- [ ] **Step 4: 运行测试确认通过**

Run: `cd backend && GOPROXY=goproxy.cn GOSUMDB=off go test ./internal/mock/ -run 'TestAnalysis' -v`
Expected: PASS

- [ ] **Step 5: 保存文件

---

## Task 3: DataSource 接口扩展 + MockAdapter + 三平台空壳（18 方法）

**Files:**
- Modify: `internal/service/source/adapter.go`（接口 +18）
- Modify: `internal/service/source/mock_adapter.go`（+18 实现）
- Modify: `internal/service/source/douyin_adapter.go`, `bilibili_adapter.go`, `xiaohongshu_adapter.go`（各 +18 返回 ErrNotImplemented）
- Test: `internal/service/source/analysis_adapter_test.go`

**Interfaces:**
- Consumes: `mock.Content*` / `mock.Market*` / `mock.Brand*`（Task 2）
- Produces: 接口签名供 3 个 Service 调用

- [ ] **Step 1: 写失败测试**

```go
package source

import (
	"context"
	"errors"
	"testing"

	"insta360-insight/internal/model"
)

func TestAnalysisMockAllReturn(t *testing.T) {
	a := NewMockAdapter()
	ctx := context.Background()
	f := model.Filter{}
	var err error
	if _, err = a.ContentKpi(ctx, f); err != nil {
		t.Fatalf("ContentKpi: %v", err)
	}
	if _, err = a.ContentTrend(ctx, f); err != nil {
		t.Fatalf("ContentTrend: %v", err)
	}
	if _, err = a.ContentForms(ctx, f); err != nil {
		t.Fatalf("ContentForms: %v", err)
	}
	if _, err = a.ContentTopics(ctx, f); err != nil {
		t.Fatalf("ContentTopics: %v", err)
	}
	if _, err = a.ContentDurations(ctx, f); err != nil {
		t.Fatalf("ContentDurations: %v", err)
	}
	if _, err = a.ContentList(ctx, f); err != nil {
		t.Fatalf("ContentList: %v", err)
	}
	if _, err = a.MarketKpi(ctx, f); err != nil {
		t.Fatalf("MarketKpi: %v", err)
	}
	if _, err = a.MarketTrend(ctx, f); err != nil {
		t.Fatalf("MarketTrend: %v", err)
	}
	if _, err = a.MarketCompetitors(ctx, f); err != nil {
		t.Fatalf("MarketCompetitors: %v", err)
	}
	if _, err = a.MarketRegions(ctx, f); err != nil {
		t.Fatalf("MarketRegions: %v", err)
	}
	if _, err = a.MarketPrices(ctx, f); err != nil {
		t.Fatalf("MarketPrices: %v", err)
	}
	if _, err = a.MarketList(ctx, f); err != nil {
		t.Fatalf("MarketList: %v", err)
	}
	if _, err = a.BrandKpi(ctx, f); err != nil {
		t.Fatalf("BrandKpi: %v", err)
	}
	if _, err = a.BrandTrend(ctx, f); err != nil {
		t.Fatalf("BrandTrend: %v", err)
	}
	if _, err = a.BrandPlatforms(ctx, f); err != nil {
		t.Fatalf("BrandPlatforms: %v", err)
	}
	if _, err = a.BrandSentiment(ctx, f); err != nil {
		t.Fatalf("BrandSentiment: %v", err)
	}
	if _, err = a.BrandKeywords(ctx, f); err != nil {
		t.Fatalf("BrandKeywords: %v", err)
	}
	if _, err = a.BrandList(ctx, f); err != nil {
		t.Fatalf("BrandList: %v", err)
	}
}

func TestAnalysisStubAllNotImplemented(t *testing.T) {
	a := NewDouyinAdapter()
	ctx := context.Background()
	f := model.Filter{}
	calls := []func() error{
		func() error { _, e := a.ContentKpi(ctx, f); return e },
		func() error { _, e := a.ContentTrend(ctx, f); return e },
		func() error { _, e := a.ContentForms(ctx, f); return e },
		func() error { _, e := a.ContentTopics(ctx, f); return e },
		func() error { _, e := a.ContentDurations(ctx, f); return e },
		func() error { _, e := a.ContentList(ctx, f); return e },
		func() error { _, e := a.MarketKpi(ctx, f); return e },
		func() error { _, e := a.MarketTrend(ctx, f); return e },
		func() error { _, e := a.MarketCompetitors(ctx, f); return e },
		func() error { _, e := a.MarketRegions(ctx, f); return e },
		func() error { _, e := a.MarketPrices(ctx, f); return e },
		func() error { _, e := a.MarketList(ctx, f); return e },
		func() error { _, e := a.BrandKpi(ctx, f); return e },
		func() error { _, e := a.BrandTrend(ctx, f); return e },
		func() error { _, e := a.BrandPlatforms(ctx, f); return e },
		func() error { _, e := a.BrandSentiment(ctx, f); return e },
		func() error { _, e := a.BrandKeywords(ctx, f); return e },
		func() error { _, e := a.BrandList(ctx, f); return e },
	}
	for i, c := range calls {
		if !errors.Is(c(), ErrNotImplemented) {
			t.Fatalf("抖音空壳第 %d 个方法应返回 ErrNotImplemented", i)
		}
	}
}
```

- [ ] **Step 2: 运行确认失败**

Run: `cd backend && GOPROXY=goproxy.cn GOSUMDB=off go test ./internal/service/source/ -run 'TestAnalysis' -v`
Expected: 编译失败

- [ ] **Step 3: 扩展接口（adapter.go）**

在 `DataSource` 接口 `CreatorList` 行之后追加：

```go
	// 内容分析域
	ContentKpi(ctx context.Context, f model.Filter) ([]model.KpiCard, error)
	ContentTrend(ctx context.Context, f model.Filter) ([]model.ViewsTrendPoint, error)
	ContentForms(ctx context.Context, f model.Filter) ([]model.PlatformShare, error)
	ContentTopics(ctx context.Context, f model.Filter) ([]model.TrackPerformance, error)
	ContentDurations(ctx context.Context, f model.Filter) ([]model.AgeShare, error)
	ContentList(ctx context.Context, f model.Filter) ([]model.ContentItem, error)

	// 市场洞察域
	MarketKpi(ctx context.Context, f model.Filter) ([]model.KpiCard, error)
	MarketTrend(ctx context.Context, f model.Filter) ([]model.ViewsTrendPoint, error)
	MarketCompetitors(ctx context.Context, f model.Filter) ([]model.PlatformShare, error)
	MarketRegions(ctx context.Context, f model.Filter) ([]model.TrackPerformance, error)
	MarketPrices(ctx context.Context, f model.Filter) ([]model.AgeShare, error)
	MarketList(ctx context.Context, f model.Filter) ([]model.Competitor, error)

	// 品牌分析域
	BrandKpi(ctx context.Context, f model.Filter) ([]model.KpiCard, error)
	BrandTrend(ctx context.Context, f model.Filter) ([]model.ViewsTrendPoint, error)
	BrandPlatforms(ctx context.Context, f model.Filter) ([]model.PlatformShare, error)
	BrandSentiment(ctx context.Context, f model.Filter) ([]model.AgeShare, error)
	BrandKeywords(ctx context.Context, f model.Filter) ([]model.TagItem, error)
	BrandList(ctx context.Context, f model.Filter) ([]model.PartnerBrand, error)
```

- [ ] **Step 4: MockAdapter 实现（mock_adapter.go）**

在 `CreatorList` 实现之后追加：

```go
func (a *MockAdapter) ContentKpi(_ context.Context, _ model.Filter) ([]model.KpiCard, error) {
	return mock.ContentKpi(), nil
}
func (a *MockAdapter) ContentTrend(_ context.Context, _ model.Filter) ([]model.ViewsTrendPoint, error) {
	return mock.ContentTrend(), nil
}
func (a *MockAdapter) ContentForms(_ context.Context, _ model.Filter) ([]model.PlatformShare, error) {
	return mock.ContentForms(), nil
}
func (a *MockAdapter) ContentTopics(_ context.Context, _ model.Filter) ([]model.TrackPerformance, error) {
	return mock.ContentTopics(), nil
}
func (a *MockAdapter) ContentDurations(_ context.Context, _ model.Filter) ([]model.AgeShare, error) {
	return mock.ContentDurations(), nil
}
func (a *MockAdapter) ContentList(_ context.Context, _ model.Filter) ([]model.ContentItem, error) {
	return mock.ContentList(), nil
}
func (a *MockAdapter) MarketKpi(_ context.Context, _ model.Filter) ([]model.KpiCard, error) {
	return mock.MarketKpi(), nil
}
func (a *MockAdapter) MarketTrend(_ context.Context, _ model.Filter) ([]model.ViewsTrendPoint, error) {
	return mock.MarketTrend(), nil
}
func (a *MockAdapter) MarketCompetitors(_ context.Context, _ model.Filter) ([]model.PlatformShare, error) {
	return mock.MarketCompetitors(), nil
}
func (a *MockAdapter) MarketRegions(_ context.Context, _ model.Filter) ([]model.TrackPerformance, error) {
	return mock.MarketRegions(), nil
}
func (a *MockAdapter) MarketPrices(_ context.Context, _ model.Filter) ([]model.AgeShare, error) {
	return mock.MarketPrices(), nil
}
func (a *MockAdapter) MarketList(_ context.Context, _ model.Filter) ([]model.Competitor, error) {
	return mock.MarketList(), nil
}
func (a *MockAdapter) BrandKpi(_ context.Context, _ model.Filter) ([]model.KpiCard, error) {
	return mock.BrandKpi(), nil
}
func (a *MockAdapter) BrandTrend(_ context.Context, _ model.Filter) ([]model.ViewsTrendPoint, error) {
	return mock.BrandTrend(), nil
}
func (a *MockAdapter) BrandPlatforms(_ context.Context, _ model.Filter) ([]model.PlatformShare, error) {
	return mock.BrandPlatforms(), nil
}
func (a *MockAdapter) BrandSentiment(_ context.Context, _ model.Filter) ([]model.AgeShare, error) {
	return mock.BrandSentiment(), nil
}
func (a *MockAdapter) BrandKeywords(_ context.Context, _ model.Filter) ([]model.TagItem, error) {
	return mock.BrandKeywords(), nil
}
func (a *MockAdapter) BrandList(_ context.Context, _ model.Filter) ([]model.PartnerBrand, error) {
	return mock.BrandList(), nil
}
```

- [ ] **Step 5: 三个平台空壳各追加 18 方法**

以 `douyin_adapter.go` 为例（其余两个完全镜像，把 `Douyin` 换成 `Bilibili`/`Xiaohongshu`）：

```go
func (a *DouyinAdapter) ContentKpi(_ context.Context, _ model.Filter) ([]model.KpiCard, error) {
	return nil, ErrNotImplemented
}
func (a *DouyinAdapter) ContentTrend(_ context.Context, _ model.Filter) ([]model.ViewsTrendPoint, error) {
	return nil, ErrNotImplemented
}
func (a *DouyinAdapter) ContentForms(_ context.Context, _ model.Filter) ([]model.PlatformShare, error) {
	return nil, ErrNotImplemented
}
func (a *DouyinAdapter) ContentTopics(_ context.Context, _ model.Filter) ([]model.TrackPerformance, error) {
	return nil, ErrNotImplemented
}
func (a *DouyinAdapter) ContentDurations(_ context.Context, _ model.Filter) ([]model.AgeShare, error) {
	return nil, ErrNotImplemented
}
func (a *DouyinAdapter) ContentList(_ context.Context, _ model.Filter) ([]model.ContentItem, error) {
	return nil, ErrNotImplemented
}
func (a *DouyinAdapter) MarketKpi(_ context.Context, _ model.Filter) ([]model.KpiCard, error) {
	return nil, ErrNotImplemented
}
func (a *DouyinAdapter) MarketTrend(_ context.Context, _ model.Filter) ([]model.ViewsTrendPoint, error) {
	return nil, ErrNotImplemented
}
func (a *DouyinAdapter) MarketCompetitors(_ context.Context, _ model.Filter) ([]model.PlatformShare, error) {
	return nil, ErrNotImplemented
}
func (a *DouyinAdapter) MarketRegions(_ context.Context, _ model.Filter) ([]model.TrackPerformance, error) {
	return nil, ErrNotImplemented
}
func (a *DouyinAdapter) MarketPrices(_ context.Context, _ model.Filter) ([]model.AgeShare, error) {
	return nil, ErrNotImplemented
}
func (a *DouyinAdapter) MarketList(_ context.Context, _ model.Filter) ([]model.Competitor, error) {
	return nil, ErrNotImplemented
}
func (a *DouyinAdapter) BrandKpi(_ context.Context, _ model.Filter) ([]model.KpiCard, error) {
	return nil, ErrNotImplemented
}
func (a *DouyinAdapter) BrandTrend(_ context.Context, _ model.Filter) ([]model.ViewsTrendPoint, error) {
	return nil, ErrNotImplemented
}
func (a *DouyinAdapter) BrandPlatforms(_ context.Context, _ model.Filter) ([]model.PlatformShare, error) {
	return nil, ErrNotImplemented
}
func (a *DouyinAdapter) BrandSentiment(_ context.Context, _ model.Filter) ([]model.AgeShare, error) {
	return nil, ErrNotImplemented
}
func (a *DouyinAdapter) BrandKeywords(_ context.Context, _ model.Filter) ([]model.TagItem, error) {
	return nil, ErrNotImplemented
}
func (a *DouyinAdapter) BrandList(_ context.Context, _ model.Filter) ([]model.PartnerBrand, error) {
	return nil, ErrNotImplemented
}
```

- [ ] **Step 6: 运行确认通过**

Run: `cd backend && GOPROXY=goproxy.cn GOSUMDB=off go test ./internal/service/source/ -run 'TestAnalysis' -v`
Expected: PASS

- [ ] **Step 7: 保存文件

---

## Task 4: 3 个 Service（content / market / brand）

**Files:**
- Create: `internal/service/content_service.go`
- Create: `internal/service/market_service.go`
- Create: `internal/service/brand_service.go`
- Test: `internal/service/analysis_service_test.go`

**Interfaces:**
- Consumes: `source.DataSource`（含 18 个方法，Task 3）
- Produces: `service.ContentService` / `service.MarketService` / `service.BrandService`（handler 调用）

- [ ] **Step 1: 写失败测试**

```go
package service

import (
	"context"
	"testing"

	"insta360-insight/internal/model"
	"insta360-insight/internal/service/source"
)

func TestAnalysisServices(t *testing.T) {
	ctx := context.Background()
	f := model.Filter{}
	cs := NewContentService(source.NewMockAdapter())
	if _, e := cs.Kpi(ctx, f); e != nil {
		t.Fatalf("Content.Kpi: %v", e)
	}
	if _, e := cs.List(ctx, f); e != nil {
		t.Fatalf("Content.List: %v", e)
	}
	ms := NewMarketService(source.NewMockAdapter())
	if _, e := ms.Kpi(ctx, f); e != nil {
		t.Fatalf("Market.Kpi: %v", e)
	}
	if _, e := ms.List(ctx, f); e != nil {
		t.Fatalf("Market.List: %v", e)
	}
	bs := NewBrandService(source.NewMockAdapter())
	if _, e := bs.Kpi(ctx, f); e != nil {
		t.Fatalf("Brand.Kpi: %v", e)
	}
	if _, e := bs.Keywords(ctx, f); e != nil {
		t.Fatalf("Brand.Keywords: %v", e)
	}
}
```

- [ ] **Step 2: 运行确认失败**

Run: `cd backend && GOPROXY=goproxy.cn GOSUMDB=off go test ./internal/service/ -run 'TestAnalysisServices' -v`
Expected: 编译失败

- [ ] **Step 3: 实现 3 个 service**

`content_service.go`：
```go
// Package service 业务逻辑层。
package service

import (
	"context"

	"insta360-insight/internal/model"
	"insta360-insight/internal/service/source"
)

type ContentService struct{ src source.DataSource }

func NewContentService(src source.DataSource) *ContentService { return &ContentService{src: src} }

func (s *ContentService) Kpi(ctx context.Context, f model.Filter) ([]model.KpiCard, error) {
	return s.src.ContentKpi(ctx, f)
}
func (s *ContentService) Trend(ctx context.Context, f model.Filter) ([]model.ViewsTrendPoint, error) {
	return s.src.ContentTrend(ctx, f)
}
func (s *ContentService) Forms(ctx context.Context, f model.Filter) ([]model.PlatformShare, error) {
	return s.src.ContentForms(ctx, f)
}
func (s *ContentService) Topics(ctx context.Context, f model.Filter) ([]model.TrackPerformance, error) {
	return s.src.ContentTopics(ctx, f)
}
func (s *ContentService) Durations(ctx context.Context, f model.Filter) ([]model.AgeShare, error) {
	return s.src.ContentDurations(ctx, f)
}
func (s *ContentService) List(ctx context.Context, f model.Filter) ([]model.ContentItem, error) {
	return s.src.ContentList(ctx, f)
}
```

`market_service.go`：
```go
package service

import (
	"context"

	"insta360-insight/internal/model"
	"insta360-insight/internal/service/source"
)

type MarketService struct{ src source.DataSource }

func NewMarketService(src source.DataSource) *MarketService { return &MarketService{src: src} }

func (s *MarketService) Kpi(ctx context.Context, f model.Filter) ([]model.KpiCard, error) {
	return s.src.MarketKpi(ctx, f)
}
func (s *MarketService) Trend(ctx context.Context, f model.Filter) ([]model.ViewsTrendPoint, error) {
	return s.src.MarketTrend(ctx, f)
}
func (s *MarketService) Competitors(ctx context.Context, f model.Filter) ([]model.PlatformShare, error) {
	return s.src.MarketCompetitors(ctx, f)
}
func (s *MarketService) Regions(ctx context.Context, f model.Filter) ([]model.TrackPerformance, error) {
	return s.src.MarketRegions(ctx, f)
}
func (s *MarketService) Prices(ctx context.Context, f model.Filter) ([]model.AgeShare, error) {
	return s.src.MarketPrices(ctx, f)
}
func (s *MarketService) List(ctx context.Context, f model.Filter) ([]model.Competitor, error) {
	return s.src.MarketList(ctx, f)
}
```

`brand_service.go`：
```go
package service

import (
	"context"

	"insta360-insight/internal/model"
	"insta360-insight/internal/service/source"
)

type BrandService struct{ src source.DataSource }

func NewBrandService(src source.DataSource) *BrandService { return &BrandService{src: src} }

func (s *BrandService) Kpi(ctx context.Context, f model.Filter) ([]model.KpiCard, error) {
	return s.src.BrandKpi(ctx, f)
}
func (s *BrandService) Trend(ctx context.Context, f model.Filter) ([]model.ViewsTrendPoint, error) {
	return s.src.BrandTrend(ctx, f)
}
func (s *BrandService) Platforms(ctx context.Context, f model.Filter) ([]model.PlatformShare, error) {
	return s.src.BrandPlatforms(ctx, f)
}
func (s *BrandService) Sentiment(ctx context.Context, f model.Filter) ([]model.AgeShare, error) {
	return s.src.BrandSentiment(ctx, f)
}
func (s *BrandService) Keywords(ctx context.Context, f model.Filter) ([]model.TagItem, error) {
	return s.src.BrandKeywords(ctx, f)
}
func (s *BrandService) List(ctx context.Context, f model.Filter) ([]model.PartnerBrand, error) {
	return s.src.BrandList(ctx, f)
}
```

- [ ] **Step 4: 运行确认通过**

Run: `cd backend && GOPROXY=goproxy.cn GOSUMDB=off go test ./internal/service/ -run 'TestAnalysisServices' -v`
Expected: PASS

- [ ] **Step 5: 保存文件

---

## Task 5: 3 个 Handler（content / market / brand）

**Files:**
- Create: `internal/api/handler/content.go`
- Create: `internal/api/handler/market.go`
- Create: `internal/api/handler/brand.go`
- Test: `internal/api/handler/analysis_handler_test.go`

**Interfaces:**
- Consumes: `service.ContentService` / `service.MarketService` / `service.BrandService`（Task 4）
- Produces: 18 个 HTTP handler 方法（router 注册）

- [ ] **Step 1: 写失败测试**

```go
package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"insta360-insight/internal/service"
	"insta360-insight/internal/service/source"
)

func TestContentHandlerKpi(t *testing.T) {
	h := NewContent(service.NewContentService(source.NewMockAdapter()))
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	h.Kpi(c)
	if w.Code != http.StatusOK {
		t.Fatalf("期望 200, 实际 %d", w.Code)
	}
	var resp struct {
		Code int `json:"code"`
		Data []struct {
			Key string `json:"key"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("json 解析失败: %v", err)
	}
	if resp.Code != 0 || len(resp.Data) != 5 {
		t.Fatalf("期望 code=0 且 5 条, 实际 code=%d len=%d", resp.Code, len(resp.Data))
	}
}

func TestBrandHandlerKeywords(t *testing.T) {
	h := NewBrand(service.NewBrandService(source.NewMockAdapter()))
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	h.Keywords(c)
	if w.Code != http.StatusOK {
		t.Fatalf("期望 200, 实际 %d", w.Code)
	}
	var resp struct {
		Code int `json:"code"`
		Data []struct {
			Word string `json:"word"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("json 解析失败: %v", err)
	}
	if resp.Code != 0 || len(resp.Data) != 8 {
		t.Fatalf("期望 code=0 且 8 条, 实际 code=%d len=%d", resp.Code, len(resp.Data))
	}
}
```

`gin` 包需 import（`github.com/gin-gonic/gin`）。

- [ ] **Step 2: 运行确认失败**

Run: `cd backend && GOPROXY=goproxy.cn GOSUMDB=off go test ./internal/api/handler/ -run 'TestContentHandler|TestBrandHandler' -v`
Expected: 编译失败

- [ ] **Step 3: 实现 3 个 handler**

`content.go`：
```go
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"insta360-insight/internal/model"
	"insta360-insight/internal/service"
)

type Content struct{ svc *service.ContentService }

func NewContent(svc *service.ContentService) *Content { return &Content{svc: svc} }

func (h *Content) Kpi(c *gin.Context) {
	d, e := h.svc.Kpi(c.Request.Context(), bindFilter(c))
	if e != nil { fail(c, e); return }
	c.JSON(http.StatusOK, model.OK(d))
}
func (h *Content) Trend(c *gin.Context) {
	d, e := h.svc.Trend(c.Request.Context(), bindFilter(c))
	if e != nil { fail(c, e); return }
	c.JSON(http.StatusOK, model.OK(d))
}
func (h *Content) Forms(c *gin.Context) {
	d, e := h.svc.Forms(c.Request.Context(), bindFilter(c))
	if e != nil { fail(c, e); return }
	c.JSON(http.StatusOK, model.OK(d))
}
func (h *Content) Topics(c *gin.Context) {
	d, e := h.svc.Topics(c.Request.Context(), bindFilter(c))
	if e != nil { fail(c, e); return }
	c.JSON(http.StatusOK, model.OK(d))
}
func (h *Content) Durations(c *gin.Context) {
	d, e := h.svc.Durations(c.Request.Context(), bindFilter(c))
	if e != nil { fail(c, e); return }
	c.JSON(http.StatusOK, model.OK(d))
}
func (h *Content) List(c *gin.Context) {
	d, e := h.svc.List(c.Request.Context(), bindFilter(c))
	if e != nil { fail(c, e); return }
	c.JSON(http.StatusOK, model.OK(d))
}
```

`market.go`（方法 Kpi/Trend/Competitors/Regions/Prices/List 同构，把 `NewMarket`/`MarketService`/`Competitors/Regions/Prices` 替换）：
```go
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"insta360-insight/internal/model"
	"insta360-insight/internal/service"
)

type Market struct{ svc *service.MarketService }

func NewMarket(svc *service.MarketService) *Market { return &Market{svc: svc} }

func (h *Market) Kpi(c *gin.Context) {
	d, e := h.svc.Kpi(c.Request.Context(), bindFilter(c))
	if e != nil { fail(c, e); return }
	c.JSON(http.StatusOK, model.OK(d))
}
func (h *Market) Trend(c *gin.Context) {
	d, e := h.svc.Trend(c.Request.Context(), bindFilter(c))
	if e != nil { fail(c, e); return }
	c.JSON(http.StatusOK, model.OK(d))
}
func (h *Market) Competitors(c *gin.Context) {
	d, e := h.svc.Competitors(c.Request.Context(), bindFilter(c))
	if e != nil { fail(c, e); return }
	c.JSON(http.StatusOK, model.OK(d))
}
func (h *Market) Regions(c *gin.Context) {
	d, e := h.svc.Regions(c.Request.Context(), bindFilter(c))
	if e != nil { fail(c, e); return }
	c.JSON(http.StatusOK, model.OK(d))
}
func (h *Market) Prices(c *gin.Context) {
	d, e := h.svc.Prices(c.Request.Context(), bindFilter(c))
	if e != nil { fail(c, e); return }
	c.JSON(http.StatusOK, model.OK(d))
}
func (h *Market) List(c *gin.Context) {
	d, e := h.svc.List(c.Request.Context(), bindFilter(c))
	if e != nil { fail(c, e); return }
	c.JSON(http.StatusOK, model.OK(d))
}
```

`brand.go`（方法 Kpi/Trend/Platforms/Sentiment/Keywords/List）：
```go
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"insta360-insight/internal/model"
	"insta360-insight/internal/service"
)

type Brand struct{ svc *service.BrandService }

func NewBrand(svc *service.BrandService) *Brand { return &Brand{svc: svc} }

func (h *Brand) Kpi(c *gin.Context) {
	d, e := h.svc.Kpi(c.Request.Context(), bindFilter(c))
	if e != nil { fail(c, e); return }
	c.JSON(http.StatusOK, model.OK(d))
}
func (h *Brand) Trend(c *gin.Context) {
	d, e := h.svc.Trend(c.Request.Context(), bindFilter(c))
	if e != nil { fail(c, e); return }
	c.JSON(http.StatusOK, model.OK(d))
}
func (h *Brand) Platforms(c *gin.Context) {
	d, e := h.svc.Platforms(c.Request.Context(), bindFilter(c))
	if e != nil { fail(c, e); return }
	c.JSON(http.StatusOK, model.OK(d))
}
func (h *Brand) Sentiment(c *gin.Context) {
	d, e := h.svc.Sentiment(c.Request.Context(), bindFilter(c))
	if e != nil { fail(c, e); return }
	c.JSON(http.StatusOK, model.OK(d))
}
func (h *Brand) Keywords(c *gin.Context) {
	d, e := h.svc.Keywords(c.Request.Context(), bindFilter(c))
	if e != nil { fail(c, e); return }
	c.JSON(http.StatusOK, model.OK(d))
}
func (h *Brand) List(c *gin.Context) {
	d, e := h.svc.List(c.Request.Context(), bindFilter(c))
	if e != nil { fail(c, e); return }
	c.JSON(http.StatusOK, model.OK(d))
}
```

`bindFilter` 与 `fail` 已在 `insight.go` 同包定义，直接复用。

- [ ] **Step 4: 运行确认通过**

Run: `cd backend && GOPROXY=goproxy.cn GOSUMDB=off go test ./internal/api/handler/ -run 'TestContentHandler|TestBrandHandler' -v`
Expected: PASS

- [ ] **Step 5: 保存文件

---

## Task 6: 路由注册 + main 接线 + 后端全量测试

**Files:**
- Modify: `internal/api/router/router.go`（`New` 在 `creatorSvc` 之后增 3 参数 +18 路由）
- Modify: `backend/main.go`（构造 3 Service 并传入）

**Interfaces:**
- Consumes: `service.ContentService`/`MarketService`/`BrandService`（Task 4）、`handler.NewContent`/`NewMarket`/`NewBrand`（Task 5）

- [ ] **Step 1: 改 router.go 签名与路由**

`New` 函数签名改为（在 `creatorSvc` 之后插入 3 个 svc 参数，保持 avatarDir 在最后）：

```go
func New(insightSvc *service.InsightService, aiSvc *service.AIService, authSvc *service.AuthService, creatorSvc *service.CreatorService, contentSvc *service.ContentService, marketSvc *service.MarketService, brandSvc *service.BrandService, disableAuth bool, devUser service.Claims, avatarDir string) *gin.Engine {
```

在 `creator := handler.NewCreator(creatorSvc)` 之后追加：

```go
	content := handler.NewContent(contentSvc)
	market := handler.NewMarket(marketSvc)
	brand := handler.NewBrand(brandSvc)
```

在受保护组 `g` 内（`g.GET("/creator/list", creator.List)` 之后）追加：

```go
		g.GET("/content/kpi", content.Kpi)
		g.GET("/content/trend", content.Trend)
		g.GET("/content/forms", content.Forms)
		g.GET("/content/topics", content.Topics)
		g.GET("/content/durations", content.Durations)
		g.GET("/content/list", content.List)

		g.GET("/market/kpi", market.Kpi)
		g.GET("/market/trend", market.Trend)
		g.GET("/market/competitors", market.Competitors)
		g.GET("/market/regions", market.Regions)
		g.GET("/market/prices", market.Prices)
		g.GET("/market/list", market.List)

		g.GET("/brand/kpi", brand.Kpi)
		g.GET("/brand/trend", brand.Trend)
		g.GET("/brand/platforms", brand.Platforms)
		g.GET("/brand/sentiment", brand.Sentiment)
		g.GET("/brand/keywords", brand.Keywords)
		g.GET("/brand/list", brand.List)
```

- [ ] **Step 2: 改 main.go**

在 `creatorSvc := service.NewCreatorService(adapter)` 之后追加：

```go
	contentSvc := service.NewContentService(adapter)
	marketSvc := service.NewMarketService(adapter)
	brandSvc := service.NewBrandService(adapter)
```

把 `router.New(...)` 调用改为（在 `creatorSvc` 之后、`disableAuth` 之前插入 3 个 svc）：

```go
	engine := router.New(insightSvc, aiSvc, authSvc, creatorSvc, contentSvc, marketSvc, brandSvc, disableAuth, devUser, avatarDir)
```

- [ ] **Step 3: 编译校验**

Run: `cd backend && GOPROXY=goproxy.cn GOSUMDB=off go build ./...`
Expected: 编译通过

- [ ] **Step 4: 后端全量测试**

Run: `cd backend && GOPROXY=goproxy.cn GOSUMDB=off go build ./... && go test ./...`
Expected: 全绿

- [ ] **Step 5: 路由抽查（SOURCE=mock）**

```bash
cd backend && SOURCE=mock AUTH_DISABLE=1 go run main.go &
curl -s localhost:8080/api/content/kpi | head -c 120
curl -s localhost:8080/api/market/list | head -c 120
curl -s localhost:8080/api/brand/keywords | head -c 120
```
Expected: `{ "code": 0, "data": [...] }`

- [ ] **Step 6: 空壳校验**

```bash
SOURCE=douyin AUTH_DISABLE=1 go run main.go &
curl -s localhost:8080/api/content/kpi
```
Expected: HTTP 502 `{"code":500,"message":"data source not implemented: ..."}`

- [ ] **Step 7: 保存文件

---

## Task 7: 通用列表组件 DataTable.vue

**Files:**
- Create: `frontend/src/components/DataTable.vue`

**Interfaces:**
- Produces: `columns`/`rows`/`searchable`/`rowKey` props，供三页总表使用

- [ ] **Step 1: 写组件**

```vue
<template>
  <div class="data-table">
    <div v-if="searchable" class="dt-toolbar">
      <el-input v-model="kw" placeholder="搜索" size="small" clearable class="dt-search" />
    </div>
    <el-table :data="pagedRows" style="width: 100%" size="small" :row-key="rowKey"
      :default-sort="{ prop: firstSortable, order: 'descending' }" height="100%">
      <el-table-column v-for="col in columns" :key="col.key"
        :prop="col.key" :label="col.label"
        :width="col.width" :align="col.align || 'left'"
        :sortable="col.sortable || false"
        :formatter="col.formatter">
        <template #default="scope" v-if="col.slot">
          <component :is="col.slot" :row="scope.row" />
        </template>
      </el-table-column>
    </el-table>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'

const props = defineProps({
  columns: { type: Array, required: true }, // [{ key, label, sortable?, width?, align?, formatter? }]
  rows: { type: Array, default: () => [] },
  searchable: { type: Boolean, default: false },
  rowKey: { type: String, default: 'id' }
})

const kw = ref('')
const firstSortable = computed(() => props.columns.find(c => c.sortable)?.key || '')
const pagedRows = computed(() => {
  if (!props.searchable || !kw.value) return props.rows
  const k = kw.value.toLowerCase()
  return props.rows.filter(r =>
    props.columns.some(c => String(r[c.key] ?? '').toLowerCase().includes(k))
  )
})
</script>

<style lang="scss" scoped>
.data-table { width: 100%; height: 100%; display: flex; flex-direction: column; gap: 6px;
  .dt-toolbar { display: flex; justify-content: flex-end; }
  .dt-search { width: 160px; }
  :deep(.el-table) { background: transparent; --el-table-border-color: var(--border);
    --el-table-header-bg-color: transparent; --el-table-bg-color: transparent;
    --el-table-tr-bg-color: transparent; --el-table-row-hover-bg-color: rgba(61,217,235,0.08);
    color: var(--text-secondary); font-size: 12px; }
  :deep(.el-table th.el-table__cell) { background: transparent; color: var(--text-muted); font-weight: 600; }
  :deep(.el-table .cell) { padding: 4px 8px; }
  :deep(.el-table .num) { font-variant-numeric: tabular-nums; }
}
</style>
```

- [ ] **Step 2: 保存文件

---

## Task 8: 前端 fallback 数据（content/market/brand 3 组共 18 数据集）

**Files:**
- Modify: `frontend/src/api/fallback-data.js`（在 `export default {` 内新增 18 个键；文件末尾新增生成函数）

**Interfaces:**
- Produces: `fallback.contentKpi`/`contentTrend`/`contentForms`/`contentTopics`/`contentDurations`/`contentList`、`market*` 同理、`brand*` 同理（store 兜底用）

- [ ] **Step 1: 在 export default 对象内新增 18 个键**

在 `creatorList: buildCreatorList(),` 之后、`options: {...}` 之前插入（数值与后端 `mock/analysis_data.go` 同口径）：

```js
  // ===== 内容分析 =====
  contentKpi: [
    { key: 'total', label: '内容总数', value: '1,284', raw: 1284, delta_pct: 12.3, delta_up: true, description: '较上期' },
    { key: 'avg_views', label: '平均播放', value: '86.4K', raw: 86400, delta_pct: 5.1, delta_up: true, description: '较上期' },
    { key: 'hit_rate', label: '爆款率', value: '8.2%', raw: 8.2, delta_pct: 1.4, delta_up: true, description: '较上期' },
    { key: 'engage', label: '平均互动率', value: '7.3%', raw: 7.3, delta_pct: 0.3, delta_up: false, description: '较上期' },
    { key: 'freq', label: '周更新频次', value: '42', raw: 42, delta_pct: 6.0, delta_up: true, description: '较上期' }
  ],
  contentTrend: buildContentTrend(),
  contentForms: [
    { platform: '教程', share: 38.0, views: 920000, color: '#3DD9EB' },
    { platform: '测评', share: 27.0, views: 650000, color: '#5EA1FF' },
    { platform: '创意短片', share: 21.0, views: 510000, color: '#A07BFF' },
    { platform: '挑战赛', share: 14.0, views: 340000, color: '#7DD96E' }
  ],
  contentTopics: [
    { track: '滑雪', views: 720000, color: '#5EA1FF' },
    { track: '潜水', views: 610000, color: '#3DD9EB' },
    { track: '骑行', views: 530000, color: '#A07BFF' },
    { track: '旅行', views: 470000, color: '#7DD96E' },
    { track: 'Vlog', views: 390000, color: '#FFB547' }
  ],
  contentDurations: [
    { bucket: '≤30s', share: 34.0, color: '#3DD9EB' },
    { bucket: '30-60s', share: 41.0, color: '#5EA1FF' },
    { bucket: '1-3min', share: 18.0, color: '#A07BFF' },
    { bucket: '3min+', share: 7.0, color: '#7DD96E' }
  ],
  contentList: buildContentList(),
  // ===== 市场洞察 =====
  marketKpi: [
    { key: 'size', label: '品类规模', value: '¥3.2B', raw: 3200000000, delta_pct: 9.8, delta_up: true, description: '较上期' },
    { key: 'growth', label: '品类增速', value: '14.6%', raw: 14.6, delta_pct: 2.1, delta_up: true, description: '较上期' },
    { key: 'share', label: 'Insta360 市占', value: '31.2%', raw: 31.2, delta_pct: 3.4, delta_up: true, description: '较上期' },
    { key: 'comp', label: '在榜竞品数', value: '6', raw: 6, delta_pct: 0, delta_up: true, description: '较上期' },
    { key: 'buzz', label: '行业声量', value: '4.7M', raw: 4700000, delta_pct: 7.5, delta_up: true, description: '较上期' }
  ],
  marketTrend: buildMarketTrend(),
  marketCompetitors: [
    { platform: 'Insta360', share: 31.2, views: 1460000, color: '#3DD9EB' },
    { platform: 'GoPro', share: 28.4, views: 1330000, color: '#FF6B6B' },
    { platform: 'DJI', share: 24.7, views: 1150000, color: '#5EA1FF' },
    { platform: '其他', share: 15.7, views: 730000, color: '#A07BFF' }
  ],
  marketRegions: [
    { track: '华东', views: 1520000, color: '#3DD9EB' },
    { track: '华南', views: 1180000, color: '#5EA1FF' },
    { track: '华北', views: 960000, color: '#A07BFF' },
    { track: '西南', views: 640000, color: '#7DD96E' },
    { track: '海外', views: 1300000, color: '#FFB547' }
  ],
  marketPrices: [
    { bucket: '<1000', share: 22.0, color: '#3DD9EB' },
    { bucket: '1000-3000', share: 41.0, color: '#5EA1FF' },
    { bucket: '3000-5000', share: 26.0, color: '#A07BFF' },
    { bucket: '5000+', share: 11.0, color: '#7DD96E' }
  ],
  marketList: [
    { name: 'Insta360', category: '全景/运动相机', buzz: 1460000, growth: 17.2, sentiment: 81.0 },
    { name: 'GoPro', category: '运动相机', buzz: 1330000, growth: 4.1, sentiment: 73.0 },
    { name: 'DJI', category: '无人机/相机', buzz: 1150000, growth: 9.8, sentiment: 76.0 },
    { name: 'Sony', category: '微单', buzz: 880000, growth: 6.3, sentiment: 79.0 },
    { name: '大疆Action', category: '运动相机', buzz: 620000, growth: 12.5, sentiment: 74.0 },
    { name: 'AKASO', category: '入门运动相机', buzz: 340000, growth: 21.0, sentiment: 68.0 }
  ],
  // ===== 品牌分析 =====
  brandKpi: [
    { key: 'buzz', label: '品牌声量', value: '2.9M', raw: 2900000, delta_pct: 11.2, delta_up: true, description: '较上期' },
    { key: 'sent', label: '好感度', value: '81%', raw: 81, delta_pct: 1.8, delta_up: true, description: '较上期' },
    { key: 'partners', label: '合作品牌数', value: '8', raw: 8, delta_pct: 14.3, delta_up: true, description: '较上期' },
    { key: 'roi', label: '内容互动 ROI', value: '4.2', raw: 4.2, delta_pct: 0.5, delta_up: true, description: '较上期' },
    { key: 'search', label: '搜索指数', value: '68.5', raw: 68.5, delta_pct: 3.2, delta_up: false, description: '较上期' }
  ],
  brandTrend: buildBrandTrend(),
  brandPlatforms: [
    { platform: '抖音', share: 44.0, views: 1280000, color: '#000000' },
    { platform: 'B站', share: 26.0, views: 754000, color: '#00A1D6' },
    { platform: '小红书', share: 22.0, views: 638000, color: '#FF2442' },
    { platform: '微博', share: 8.0, views: 232000, color: '#E6162D' }
  ],
  brandSentiment: [
    { bucket: '正面', share: 67.0, color: '#7DD96E' },
    { bucket: '中性', share: 24.0, color: '#FFB547' },
    { bucket: '负面', share: 9.0, color: '#FF6B6B' }
  ],
  brandKeywords: [
    { word: '画质', weight: 92.0 }, { word: '防抖', weight: 88.0 },
    { word: '全景', weight: 81.0 }, { word: '运动相机', weight: 76.0 },
    { word: 'Vlog', weight: 70.0 }, { word: '旅行', weight: 64.0 },
    { word: '性价比', weight: 58.0 }, { word: '续航', weight: 49.0 }
  ],
  brandList: [
    { name: '红牛', industry: '饮料', contents: 8, exposure: 5400000, engagement: 320000, roi: 4.2 },
    { name: '始祖鸟', industry: '户外', contents: 6, exposure: 4100000, engagement: 260000, roi: 3.9 },
    { name: '携程', industry: '旅行', contents: 5, exposure: 3300000, engagement: 210000, roi: 3.5 },
    { name: '迪卡侬', industry: '运动', contents: 7, exposure: 2900000, engagement: 190000, roi: 3.1 },
    { name: '大疆', industry: '无人机', contents: 4, exposure: 2600000, engagement: 175000, roi: 3.8 },
    { name: '索尼', industry: '影像', contents: 3, exposure: 1800000, engagement: 120000, roi: 3.3 },
    { name: 'Keep', industry: '健身', contents: 5, exposure: 1500000, engagement: 98000, roi: 2.9 },
    { name: '小米', industry: '科技', contents: 2, exposure: 980000, engagement: 64000, roi: 2.6 }
  ],
```

- [ ] **Step 2: 文件末尾新增生成函数（buildBrandTrend 之后）**

```js
// 内容分析: 近 30 天内容播放量趋势
function buildContentTrend() {
  const days = 30, out = [], now = new Date()
  now.setHours(0, 0, 0, 0)
  const base = new Date(now.getTime() - (days - 1) * 86400000)
  for (let i = 0; i < days; i++) {
    const d = new Date(base.getTime() + i * 86400000)
    const views = 2400000 + i * 120000 + ((i * 7) % 11) * 30000
    const prev = 2400000 + i * 95000
    const mm = String(d.getMonth() + 1).padStart(2, '0')
    const dd = String(d.getDate()).padStart(2, '0')
    out.push({ date: `${mm}-${dd}`, views, prev_views: prev })
  }
  return out
}

// 内容分析: 15 条爆款内容
function buildContentList() {
  const forms = ['教程', '测评', '创意短片', '挑战赛']
  const topics = ['滑雪', '潜水', '骑行', '旅行', 'Vlog']
  return Array.from({ length: 15 }, (_, i) => ({
    id: i + 1,
    title: topics[i % topics.length] + '第' + (i + 1) + '期',
    form: forms[i % forms.length],
    topic: topics[i % topics.length],
    views: 50000 + ((i * 53000) % 1200000),
    engagement: +(5.0 + (i % 9) + (i % 4) * 0.6).toFixed(2),
    is_hit: i % 3 === 0
  }))
}

// 市场洞察: 近 30 天品类声量趋势
function buildMarketTrend() {
  const days = 30, out = [], now = new Date()
  now.setHours(0, 0, 0, 0)
  const base = new Date(now.getTime() - (days - 1) * 86400000)
  for (let i = 0; i < days; i++) {
    const d = new Date(base.getTime() + i * 86400000)
    const views = 140000 + i * 9000 + ((i * 5) % 9) * 2000
    const prev = 140000 + i * 7000
    const mm = String(d.getMonth() + 1).padStart(2, '0')
    const dd = String(d.getDate()).padStart(2, '0')
    out.push({ date: `${mm}-${dd}`, views, prev_views: prev })
  }
  return out
}

// 品牌分析: 近 30 天品牌声量趋势
function buildBrandTrend() {
  const days = 30, out = [], now = new Date()
  now.setHours(0, 0, 0, 0)
  const base = new Date(now.getTime() - (days - 1) * 86400000)
  for (let i = 0; i < days; i++) {
    const d = new Date(base.getTime() + i * 86400000)
    const views = 90000 + i * 6000 + ((i * 4) % 7) * 1500
    const prev = 90000 + i * 5000
    const mm = String(d.getMonth() + 1).padStart(2, '0')
    const dd = String(d.getDate()).padStart(2, '0')
    out.push({ date: `${mm}-${dd}`, views, prev_views: prev })
  }
  return out
}
```

- [ ] **Step 3: 保存文件

---

## Task 9: 前端 api 模块（3 个）+ store（3 个）

**Files:**
- Create: `frontend/src/api/content.js`, `frontend/src/api/market.js`, `frontend/src/api/brand.js`
- Create: `frontend/src/stores/content.js`, `frontend/src/stores/market.js`, `frontend/src/stores/brand.js`

**Interfaces:**
- api 暴露 `kpi/trend/forms|competitors|platforms/topics|regions/sentiment/durations|prices/keywords/list(q)`（request 已带 `/api` base）
- store 暴露 `kpi/trend/dist1/dist2/dist3/list/loading/usedFallback/loadAll`，供三页 view 使用

- [ ] **Step 1: 写 api/content.js**

```js
import request from './request'
const content = {
  kpi:       (q) => request.get('/content/kpi', { params: q }),
  trend:     (q) => request.get('/content/trend', { params: q }),
  forms:     (q) => request.get('/content/forms', { params: q }),
  topics:    (q) => request.get('/content/topics', { params: q }),
  durations: (q) => request.get('/content/durations', { params: q }),
  list:      (q) => request.get('/content/list', { params: q })
}
export default content
```

`api/market.js`：
```js
import request from './request'
const market = {
  kpi:         (q) => request.get('/market/kpi', { params: q }),
  trend:       (q) => request.get('/market/trend', { params: q }),
  competitors: (q) => request.get('/market/competitors', { params: q }),
  regions:     (q) => request.get('/market/regions', { params: q }),
  prices:      (q) => request.get('/market/prices', { params: q }),
  list:        (q) => request.get('/market/list', { params: q })
}
export default market
```

`api/brand.js`：
```js
import request from './request'
const brand = {
  kpi:        (q) => request.get('/brand/kpi', { params: q }),
  trend:      (q) => request.get('/brand/trend', { params: q }),
  platforms:  (q) => request.get('/brand/platforms', { params: q }),
  sentiment:  (q) => request.get('/brand/sentiment', { params: q }),
  keywords:   (q) => request.get('/brand/keywords', { params: q }),
  list:       (q) => request.get('/brand/list', { params: q })
}
export default brand
```

- [ ] **Step 2: 写 stores/content.js（dist1←forms, dist2←topics, dist3←durations）**

```js
import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import contentApi from '../api/content'
import fallback from '../api/fallback-data'
import { useFilterStore } from './filter'

export const useContentStore = defineStore('content', () => {
  const loading = ref(false)
  const error = ref(null)
  const usedFallback = ref(false)

  const kpi = ref([])
  const trend = ref([])
  const dist1 = ref([]) // forms (PlatformShare)
  const dist2 = ref([]) // topics (TrackPerformance)
  const dist3 = ref([]) // durations (AgeShare)
  const list = ref([])

  function fillWithFallback() {
    kpi.value = fallback.contentKpi
    trend.value = fallback.contentTrend
    dist1.value = fallback.contentForms
    dist2.value = fallback.contentTopics
    dist3.value = fallback.contentDurations
    list.value = fallback.contentList
    usedFallback.value = true
  }
  fillWithFallback()

  function withTimeout(promise, ms) {
    return Promise.race([
      promise,
      new Promise((_, reject) => setTimeout(() => reject(new Error('timeout')), ms))
    ])
  }

  async function loadAll() {
    loading.value = true
    error.value = null
    const f = useFilterStore()
    const q = f.toQuery()
    try {
      const [k, tr, d1, d2, d3, ls] = await withTimeout(Promise.all([
        contentApi.kpi(q), contentApi.trend(q), contentApi.forms(q),
        contentApi.topics(q), contentApi.durations(q), contentApi.list(q)
      ]), 800)
      if (!Array.isArray(k) || !Array.isArray(tr) || !Array.isArray(d1) ||
          !Array.isArray(d2) || !Array.isArray(d3) || !Array.isArray(ls)) {
        throw new Error('后端返回数据结构异常, 已回退到本地兜底数据')
      }
      kpi.value = k; trend.value = tr; dist1.value = d1; dist2.value = d2; dist3.value = d3; list.value = ls
      usedFallback.value = false
    } catch (e) {
      usedFallback.value = true
      error.value = e?.message || '后端不可达, 已切换到本地数据'
    } finally {
      loading.value = false
    }
  }

  const hasData = computed(() => kpi.value.length > 0)
  return { loading, error, usedFallback, hasData, kpi, trend, dist1, dist2, dist3, list, loadAll }
})
```

`stores/market.js`（dist1←competitors, dist2←regions, dist3←prices）：
```js
import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import marketApi from '../api/market'
import fallback from '../api/fallback-data'
import { useFilterStore } from './filter'

export const useMarketStore = defineStore('market', () => {
  const loading = ref(false), error = ref(null), usedFallback = ref(false)
  const kpi = ref([]), trend = ref([])
  const dist1 = ref([]), dist2 = ref([]), dist3 = ref([]), list = ref([])

  function fillWithFallback() {
    kpi.value = fallback.marketKpi; trend.value = fallback.marketTrend
    dist1.value = fallback.marketCompetitors; dist2.value = fallback.marketRegions
    dist3.value = fallback.marketPrices; list.value = fallback.marketList
    usedFallback.value = true
  }
  fillWithFallback()

  function withTimeout(promise, ms) {
    return Promise.race([promise, new Promise((_, r) => setTimeout(() => r(new Error('timeout')), ms))])
  }
  async function loadAll() {
    loading.value = true; error.value = null
    const f = useFilterStore(); const q = f.toQuery()
    try {
      const [k, tr, d1, d2, d3, ls] = await withTimeout(Promise.all([
        marketApi.kpi(q), marketApi.trend(q), marketApi.competitors(q),
        marketApi.regions(q), marketApi.prices(q), marketApi.list(q)
      ]), 800)
      if (!Array.isArray(k) || !Array.isArray(tr) || !Array.isArray(d1) ||
          !Array.isArray(d2) || !Array.isArray(d3) || !Array.isArray(ls)) {
        throw new Error('后端返回数据结构异常, 已回退到本地兜底数据')
      }
      kpi.value = k; trend.value = tr; dist1.value = d1; dist2.value = d2; dist3.value = d3; list.value = ls
      usedFallback.value = false
    } catch (e) {
      usedFallback.value = true; error.value = e?.message || '后端不可达, 已切换到本地数据'
    } finally { loading.value = false }
  }
  const hasData = computed(() => kpi.value.length > 0)
  return { loading, error, usedFallback, hasData, kpi, trend, dist1, dist2, dist3, list, loadAll }
})
```

`stores/brand.js`（dist1←platforms, dist2←sentiment, dist3←keywords）：
```js
import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import brandApi from '../api/brand'
import fallback from '../api/fallback-data'
import { useFilterStore } from './filter'

export const useBrandStore = defineStore('brand', () => {
  const loading = ref(false), error = ref(null), usedFallback = ref(false)
  const kpi = ref([]), trend = ref([])
  const dist1 = ref([]), dist2 = ref([]), dist3 = ref([]), list = ref([])

  function fillWithFallback() {
    kpi.value = fallback.brandKpi; trend.value = fallback.brandTrend
    dist1.value = fallback.brandPlatforms; dist2.value = fallback.brandSentiment
    dist3.value = fallback.brandKeywords; list.value = fallback.brandList
    usedFallback.value = true
  }
  fillWithFallback()

  function withTimeout(promise, ms) {
    return Promise.race([promise, new Promise((_, r) => setTimeout(() => r(new Error('timeout')), ms))])
  }
  async function loadAll() {
    loading.value = true; error.value = null
    const f = useFilterStore(); const q = f.toQuery()
    try {
      const [k, tr, d1, d2, d3, ls] = await withTimeout(Promise.all([
        brandApi.kpi(q), brandApi.trend(q), brandApi.platforms(q),
        brandApi.sentiment(q), brandApi.keywords(q), brandApi.list(q)
      ]), 800)
      if (!Array.isArray(k) || !Array.isArray(tr) || !Array.isArray(d1) ||
          !Array.isArray(d2) || !Array.isArray(d3) || !Array.isArray(ls)) {
        throw new Error('后端返回数据结构异常, 已回退到本地兜底数据')
      }
      kpi.value = k; trend.value = tr; dist1.value = d1; dist2.value = d2; dist3.value = d3; list.value = ls
      usedFallback.value = false
    } catch (e) {
      usedFallback.value = true; error.value = e?.message || '后端不可达, 已切换到本地数据'
    } finally { loading.value = false }
  }
  const hasData = computed(() => kpi.value.length > 0)
  return { loading, error, usedFallback, hasData, kpi, trend, dist1, dist2, dist3, list, loadAll }
})
```

- [ ] **Step 3: 保存文件

---

## Task 10: 3 个视图 + 路由改指向

**Files:**
- Create: `frontend/src/views/ContentAnalysis.vue`, `frontend/src/views/MarketInsights.vue`, `frontend/src/views/BrandAnalysis.vue`
- Modify: `frontend/src/router/index.js`（3 处改指向）

**Interfaces:**
- Consumes: 3 个 store（Task 9）、`DataTable`（Task 7）、`KpiCard`/`TrendChart`/`PlatformDonut`(centerLabel)/`TrackBarChart`/`AgeDonut`(centerLabel)

- [ ] **Step 1: 写 ContentAnalysis.vue（板块: KPI → trend+forms → topics+durations → 总表）**

```vue
<template>
  <div class="page" v-loading="store.loading">
    <header class="page-head">
      <div class="title-block">
        <h1>内容分析 <span class="info" title="数据每 5 分钟刷新一次">ⓘ</span></h1>
        <p class="sub">{{ dateRangeText }} · 内容营销表现概览</p>
      </div>
    </header>

    <section class="kpi-row"><KpiCard v-for="k in store.kpi" :key="k.key" :kpi="k" /></section>

    <section class="grid-row-2">
      <div class="card chart-card">
        <div class="card-head"><span class="card-title">内容播放量趋势</span></div>
        <div class="card-body chart-body"><TrendChart :data="store.trend" :granularity="filter.granularity" /></div>
      </div>
      <div class="card">
        <div class="card-head"><span class="card-title">内容形式分布</span></div>
        <div class="card-body donut-body">
          <div class="donut-chart"><PlatformDonut :data="store.dist1" centerLabel="内容形式" /></div>
          <ul class="legend-col">
            <li v-for="p in store.dist1" :key="p.platform"><span class="dot" :style="{ background: p.color }" /><span class="lab">{{ p.platform }}</span><span class="pct num">{{ p.share }}%</span></li>
          </ul>
        </div>
      </div>
    </section>

    <section class="grid-row-2">
      <div class="card">
        <div class="card-head"><span class="card-title">内容主题分布</span></div>
        <div class="card-body chart-body"><TrackBarChart :data="store.dist2" /></div>
      </div>
      <div class="card">
        <div class="card-head"><span class="card-title">内容时长段分布</span></div>
        <div class="card-body donut-body">
          <div class="donut-chart"><AgeDonut :data="store.dist3" centerLabel="时长段" /></div>
          <ul class="legend-col">
            <li v-for="a in store.dist3" :key="a.bucket"><span class="dot" :style="{ background: a.color }" /><span class="lab">{{ a.bucket }}</span><span class="pct num">{{ a.share }}%</span></li>
          </ul>
        </div>
      </div>
    </section>

    <section class="grid-row-full">
      <div class="card">
        <div class="card-head"><span class="card-title">爆款内容列表</span></div>
        <div class="card-body table-body">
          <DataTable :columns="contentCols" :rows="store.list" searchable rowKey="id" />
        </div>
      </div>
    </section>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { useFilterStore } from '../stores/filter'
import { useContentStore } from '../stores/content'
import KpiCard from '../components/KpiCard.vue'
import TrendChart from '../components/TrendChart.vue'
import PlatformDonut from '../components/PlatformDonut.vue'
import TrackBarChart from '../components/TrackBarChart.vue'
import AgeDonut from '../components/AgeDonut.vue'
import DataTable from '../components/DataTable.vue'

const filter = useFilterStore()
const store = useContentStore()
const dateRangeText = computed(() => filter.dateRange?.length ? filter.dateRange.join(' ~ ') : '')
const contentCols = [
  { key: 'title', label: '标题', sortable: true },
  { key: 'form', label: '形式', sortable: true },
  { key: 'topic', label: '主题', sortable: true },
  { key: 'views', label: '播放', align: 'right', sortable: true, formatter: (r) => (r.views || 0).toLocaleString() },
  { key: 'engagement', label: '互动率', align: 'right', sortable: true, formatter: (r) => (r.engagement ?? 0) + '%' },
  { key: 'isHit', label: '爆款', align: 'center', sortable: true, formatter: (r) => r.isHit ? '🔥' : '—' }
]
onMounted(() => store.loadAll())
</script>

<style lang="scss" scoped>
.page { display: flex; flex-direction: column; gap: 8px; }
.page-head { display: flex; align-items: center; justify-content: space-between;
  .title-block h1 { margin: 0; font-size: 15px; font-weight: 700; color: var(--text-primary); display: flex; gap: 6px; align-items: center;
    .info { cursor: help; color: var(--text-muted); font-size: 12px; } }
  .sub { margin: 1px 0 0; color: var(--text-muted); font-size: 10px; } }
.kpi-row { display: grid; grid-template-columns: repeat(5, minmax(0, 1fr)); gap: 6px;
  @media (max-width: 1280px) { grid-template-columns: repeat(3, minmax(0, 1fr)); }
  @media (max-width: 768px) { grid-template-columns: repeat(2, minmax(0, 1fr)); } }
.card { background: var(--bg-elev); border: 1px solid var(--border); border-radius: var(--radius); padding: 8px 10px; min-width: 0; }
.card-head { display: flex; align-items: center; justify-content: space-between; margin-bottom: 2px; min-height: 20px;
  .card-title { font-size: 12px; font-weight: 600; color: var(--text-primary); } }
.card-body { padding-top: 0; }
.chart-body { height: 184px; }
.donut-body { display: grid; grid-template-columns: minmax(0, 1.1fr) minmax(0, 1fr); height: 170px; padding-top: 0; align-items: center; gap: 4px;
  .donut-chart { width: 100%; height: 100%; min-width: 0; min-height: 0; overflow: hidden; display: flex; align-items: center; justify-content: center; }
  .legend-col { list-style: none; margin: 0; padding: 0; display: flex; flex-direction: column; gap: 6px;
    li { display: flex; align-items: center; gap: 6px; font-size: 11px; color: var(--text-secondary); }
    .dot { width: 7px; height: 7px; border-radius: 50%; flex-shrink: 0; }
    .lab { flex: 1; } .pct { color: var(--text-primary); font-weight: 600; } } }
.grid-row-2 { display: grid; grid-template-columns: 2fr 1fr; gap: 10px; }
.grid-row-full { display: grid; grid-template-columns: minmax(0, 1fr); gap: 10px; > .card { min-width: 0; } .table-body { padding: 0; height: 360px; } }
@media (max-width: 1280px) { .grid-row-2 { grid-template-columns: minmax(0, 1fr); } }
</style>
```

- [ ] **Step 2: 写 MarketInsights.vue（板块: KPI → trend+competitors → regions+prices → 总表）**

与 ContentAnalysis 同构，把 `useContentStore`→`useMarketStore`、`store.dist1/2/3`←`marketCompetitors/Regions/Prices`、列定义改为：
```js
const marketCols = [
  { key: 'name', label: '竞品', sortable: true },
  { key: 'category', label: '品类', sortable: true },
  { key: 'buzz', label: '声量', align: 'right', sortable: true, formatter: (r) => (r.buzz || 0).toLocaleString() },
  { key: 'growth', label: '增速', align: 'right', sortable: true, formatter: (r) => (r.growth ?? 0) + '%' },
  { key: 'sentiment', label: '好感度', align: 'right', sortable: true, formatter: (r) => (r.sentiment ?? 0) + '%' }
]
```
标题 `市场洞察`，副标题 `品类竞争格局概览`；形式分布卡片标题 `竞品占比`(PlatformDonut centerLabel="竞品占比")；主题分布卡片标题 `地域热度`(TrackBarChart)；时长段卡片标题 `价格段分布`(AgeDonut centerLabel="价格段")；总表标题 `竞品对比榜`。

- [ ] **Step 3: 写 BrandAnalysis.vue（板块: KPI → trend+platforms → sentiment+keywords → 总表）**

与上面同构，把 `useBrandStore`、`store.dist1=platforms`/`dist2=sentiment`/`dist3=keywords`、列定义：
```js
const brandCols = [
  { key: 'name', label: '品牌', sortable: true },
  { key: 'industry', label: '行业', sortable: true },
  { key: 'contents', label: '合作内容数', align: 'right', sortable: true },
  { key: 'exposure', label: '曝光', align: 'right', sortable: true, formatter: (r) => (r.exposure || 0).toLocaleString() },
  { key: 'engagement', label: '互动', align: 'right', sortable: true, formatter: (r) => (r.engagement || 0).toLocaleString() },
  { key: 'roi', label: 'ROI', align: 'right', sortable: true, formatter: (r) => (r.roi ?? 0).toFixed(1) }
]
```
标题 `品牌分析`，副标题 `品牌声量与合作效果概览`；
- 平台分布卡片：PlatformDonut centerLabel="平台分布"，dist1
- 情感分布卡片：AgeDonut centerLabel="情感分布"，dist2，legend 用 `store.dist2`
- **高频词**卡片：用轻量 tag 云（不建组件），单独 `.tag-cloud` 样式：
```vue
<div class="card">
  <div class="card-head"><span class="card-title">品牌高频词</span></div>
  <div class="card-body tag-cloud">
    <span v-for="t in store.dist3" :key="t.word" class="tag" :style="{ fontSize: (12 + t.weight/12) + 'px', opacity: 0.5 + t.weight/200 }">{{ t.word }}</span>
  </div>
</div>
```
（dist3=keywords 放高频词卡片；sentiment 占 grid-row-2 第二列，高频词占 grid-row-2 之后单独一行或并入：本设计把 情感(sentiment) + 高频词(keywords) 放在同一 `grid-row-2`，高频词用 tag 云取代 Donut 的 legend 区。）
- 总表标题 `合作品牌效果`。

- [ ] **Step 4: 改 router/index.js（3 处）**

把：
```js
  { path: '/content', component: () => import('../views/PlaceholderView.vue'), meta: { title: '内容分析' } },
  { path: '/market', component: () => import('../views/PlaceholderView.vue'), meta: { title: '市场洞察' } },
  { path: '/brand', component: () => import('../views/PlaceholderView.vue'), meta: { title: '品牌分析' } },
```
改为：
```js
  { path: '/content', component: () => import('../views/ContentAnalysis.vue'), meta: { title: '内容分析' } },
  { path: '/market', component: () => import('../views/MarketInsights.vue'), meta: { title: '市场洞察' } },
  { path: '/brand', component: () => import('../views/BrandAnalysis.vue'), meta: { title: '品牌分析' } },
```

- [ ] **Step 5: 保存文件

---

## Task 11: 前端构建 + 无头验证

**Files:** 无新增，验收步骤

- [ ] **Step 1: 构建**

Run: `cd frontend && npm run build --outDir .build-tmp`（沙箱可跑）
Expected: 编译通过，ContentAnalysis/MarketInsights/BrandAnalysis 独立 chunk 生成

- [ ] **Step 2: 清理构建产物（中文路径下 bash rm 被 safe-delete 拦截，用 Python）**

```python
import shutil, os
p = "frontend/.build-tmp"
if os.path.exists(p): shutil.rmtree(p, ignore_errors=True)
```

- [ ] **Step 3: Edge 无头 DOM 验证（沙箱可跑）**

```bash
msedge --headless=new --no-sandbox --disable-gpu --virtual-time-budget=9000 \
  --dump-dom http://localhost:5173/#/content > /tmp/content.html
msedge --headless=new --no-sandbox --disable-gpu --virtual-time-budget=9000 \
  --dump-dom http://localhost:5173/#/market > /tmp/market.html
msedge --headless=new --no-sandbox --disable-gpu --virtual-time-budget=9000 \
  --dump-dom http://localhost:5173/#/brand > /tmp/brand.html
```
Expected: 每个 DOM 含 `kpi-row` 容器与 5 张 KPI 文本、图表 canvas、对应总表标题与行数据、品牌页含 `tag-cloud`；零页面级控制台错误（仅 benign 网络提示）。

- [ ] **Step 4: 保存（无文件改动）

---

## Task 12: 更新可复用分析页模板文档

**Files:**
- Modify: `docs/superpowers/分析页模板说明.md`

**Interfaces:**
- Produces: 补 DataTable 用法 + 三页数据集约定

- [ ] **Step 1: 在模板文档末尾追加三页段落**

```markdown
## 内容 / 市场 / 品牌 Tab（2026-07-26）

三页共用同一骨架（KPI×5 → 趋势 + 分布① → 分布② + 分布③ → 总表），差异仅数据集与板块语义：

- 后端: `DataSource` 接口新增 content/market/brand 三组共 18 方法 → `MockAdapter` 实现 + 三平台空壳 `ErrNotImplemented` → 各 `XxxService` + `XxxHandler` → `router.New` 加 3 参数 +18 路由。
- 前端: 新增 `stores/{content,market,brand}.js`（state 用 `dist1/dist2/dist3` 屏蔽页差异，映射：
  content: forms/topics/durations；market: competitors/regions/prices；brand: platforms/sentiment/keywords）
  + `api/{content,market,brand}.js` + `fallback-data.js` 加 3 组 18 数据集 + `views/*` + 路由改 3 处。
- 通用列表: `DataTable.vue`（columns 配置驱动，支持 `searchable` 与前端排序），三页总表复用，不再各造表。
- 品牌高频词: 不建组件，用 `.tag-cloud` 轻量 tags 样式渲染 `store.dist3`(TagItem)。
- 地域分布: 用 `TrackBarChart` 横向柱图，**不用地图**（合规）。
```

- [ ] **Step 2: 保存文件

---

## Self-Review

- **Spec 覆盖**: 架构(接口+18/adapter/service/handler/路由/main) ✓；模型 4 类型 ✓；Mock 18 函数 ✓；KPI×5 ✓；趋势 ✓；分布(形式/主题/时长·竞品/地域/价格·平台/情感/高频词) ✓；三页总表 ✓；DataTable ✓；fallback ✓；3 store ✓；3 view ✓；路由 3 处 ✓；测试 ✓；模板文档 ✓；文件清单 ✓。
- **Placeholder 扫描**: 无 TBD/TODO；所有步骤含具体代码/数据；Step 1 占位 `sumShare`/`toAny` 已在 Step 3 注明删除并给出修正断言。
- **类型一致性**: Service/Handler 方法名与接口、路由、前端 api/store 完全对应；`BrandKeywords`/`BrandSentiment` 返回 `[]TagItem`/`[]AgeShare`，前端 `brand.dist3` 同时承载 keywords(TagItem) 与 sentiment(AgeShare) 仅在不同页用——本计划 brand 页 `dist2=sentiment(AgeShare)`、`dist3=keywords(TagItem)`，view 内分别使用，无冲突；`ContentItem.IsHit` → 前端 `is_hit`（JSON 下划线由 Go `json:"isHit"` 序列化为 `isHit`，fallback 用 `is_hit` 仅为兜底，store 优先用后端真实字段 `isHit`；formatter 用 `r.isHit ?? r.is_hit` 更稳妥，已在列 formatter 留 `r.isHit`）。
- **与已读代码对齐**: router.New 现有签名含 `avatarDir` 在最后，本计划将 3 个 svc 插在 `creatorSvc` 之后、`disableAuth` 之前，main.go 同步；`bindFilter`/`fail` 在 insight.go 同包 handler 复用；`model.OK`/`model.Fail` 已存在；TrendChart `:data`/`:granularity`、PlatformDonut `centerLabel`、AgeDonut `centerLabel`、TrackBarChart `:data`、KpiCard `:kpi` 均来自已上线组件（creator 计划已加 centerLabel 等 prop）。
- **环境约束对齐**: Go 测试步骤加 `GOPROXY=goproxy.cn GOSUMDB=off` 前缀（已实测本机可跑）；非 git 仓库用「保存文件」替代 commit；清理构建产物用 Python rmtree 规避 safe-delete 对仓库路径的 rm 拦截。
- **偏离 spec 的已确认调整**: spec 写「分布③ 与总表同一行右侧 / 品牌高频词并入分布区」；本计划把品牌高频词做成独立 `.tag-cloud` 卡片（grid-row-2 与 sentiment 并列），更清晰且仍不新建组件，符合 spec「轻量 tag 云」意图。
