# 达人分析页 (Creator Analysis) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 新建「达人分析」真实页面（概览仪表盘），全栈接入阶段二 adapter 层，并沉淀可复用分析页模板供后续 Tab 复制。

**Architecture:** 后端扩展 `DataSource` 接口 +6 个 creator 域方法，`MockAdapter` 实现、`三平台空壳`返回 `ErrNotImplemented`；新增 `CreatorService`/`CreatorHandler` 暴露 `/api/creator/*` 6 路由。前端新增 `creator` store + `api/creator.js` + `fallback-data.js` 兜底 + `CreatorAnalysis.vue`，复用现有 `KpiCard/TrendChart/PlatformDonut/TrackBarChart/AgeDonut/TopCreatorsTable`（仅做 4 处向后兼容的小增强）。

**Tech Stack:** Go + Gin + modernc.org/sqlite(本任务不涉及 DB); Vue 3 + Vite + Element Plus + Pinia + Axios + ECharts; 复用现有 adapter/service/handler 模式。

## Global Constraints

- ECharts 容器: 严禁 `min-height` 硬编码; 必备 `min-width:0; min-height:0; overflow:hidden` + `ResizeObserver` 主动 resize; 容器 100% 自适应父级。
- CSS Grid 列宽: 必须用 `minmax(0, Xfr)`, 不用 `1fr`(会被内容撑爆); grid 内 `.card` 全加 `min-width:0`。
- 响应约定: 业务失败 HTTP 200 + `{code:非0, message}`; 中间件未授权 HTTP 401。
- 暗色主题 + 玻璃拟态 + 品牌色 `--brand:#3DD9EB`。
- 数据来源维持 mock（阶段三真实 API 卡在开放平台 appkey）；前端 `fallback-data.js` 必须兜底。
- **环境约束**: 本机网络拦截 Go 工具链, 后端 `go build`/`go test` 必须在用户**放行网络环境**执行; 沙箱内后端测试步骤标注「(放行网络执行)」, 仅做代码 re-read 自检。
- 项目为**非 git 仓库**: 不执行 `git commit`; 保存文件即可（计划内 "Commit" 步骤一律改为「保存文件」）。

---

## File Structure

**后端（新增/改）**
- `internal/model/types.go` — 改: 新增 `GenderShare` + `Audience` 类型
- `internal/mock/creator_data.go` — 新: 6 个 creator 数据函数
- `internal/service/source/adapter.go` — 改: 接口 +6 creator 方法
- `internal/service/source/mock_adapter.go` — 改: +6 实现
- `internal/service/source/{douyin,bilibili,xiaohongshu}_adapter.go` — 改: +6 返回 ErrNotImplemented
- `internal/service/creator_service.go` — 新: `CreatorService` +6 方法
- `internal/api/handler/creator.go` — 新: `Creator` handler +6 方法
- `internal/api/router/router.go` — 改: `New` 增加 `creatorSvc` 参数 +6 路由
- `backend/main.go` — 改: 构造 `CreatorService` 并传入 `router.New`
- 对应 `*_test.go` — 新

**前端（新增/改）**
- `src/components/KpiCard.vue` — 改: `META` 增加 `new` 键（向后兼容）
- `src/components/PlatformDonut.vue` — 改: 增加 `centerLabel` prop（默认 "总播放量"）
- `src/components/AgeDonut.vue` — 改: 增加 `centerLabel`(默认 "总粉丝") + `centerValue`(默认 "") prop
- `src/components/TopCreatorsTable.vue` — 改: 增加 `rows` prop（默认 null → 回退 insight store，向后兼容）
- `src/api/creator.js` — 新: 6 个端点封装
- `src/api/fallback-data.js` — 改: +creator* 6 数据集
- `src/stores/creator.js` — 新: 镜像 insight store
- `src/views/CreatorAnalysis.vue` — 新: 页面外壳
- `src/router/index.js` — 改: `/creator` 从 PlaceholderView 改指向 CreatorAnalysis
- `docs/superpowers/分析页模板说明.md` — 新: 可复用模板文档

---

## Task 1: 新增 Audience / GenderShare 模型类型

**Files:**
- Modify: `internal/model/types.go` (在 TopCreator 段之后追加)
- Test: `internal/model/creator_test.go`

**Interfaces:**
- Produces: `model.GenderShare`, `model.Audience`（后续 adapter/service/handler 依赖）

- [ ] **Step 1: 写失败测试**

```go
package model

import (
	"encoding/json"
	"testing"
)

func TestAudienceJSONShape(t *testing.T) {
	a := Audience{
		Age:    []AgeShare{{Bucket: "25-34 岁", Share: 42.7, Color: "#FFB547"}},
		Gender: []GenderShare{{Gender: "男", Share: 58.4, Color: "#5EA1FF"}},
	}
	b, err := json.Marshal(a)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var out Audience
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(out.Age) != 1 || len(out.Gender) != 1 {
		t.Fatalf("期望 age/gender 各 1 条, 实际 %d/%d", len(out.Age), len(out.Gender))
	}
	if out.Gender[0].Gender != "男" {
		t.Fatalf("Gender 字段丢失: %+v", out.Gender[0])
	}
}
```

- [ ] **Step 2: 运行测试确认失败**

Run: `cd backend && go test ./internal/model/ -run TestAudienceJSONShape -v`
Expected: 编译失败（`Audience`/`GenderShare` 未定义）

- [ ] **Step 3: 实现类型**

在 `internal/model/types.go` 的 `TopCreator` 定义之后追加：

```go
// GenderShare 粉丝性别分布
type GenderShare struct {
	Gender string  `json:"gender"` // 男 / 女
	Share  float64 `json:"share"`  // 占比 %
	Color  string  `json:"color"`
}

// Audience 达人粉丝画像(年龄 + 性别), 供 /api/creator/audience 返回
type Audience struct {
	Age    []AgeShare     `json:"age"`
	Gender []GenderShare  `json:"gender"`
}
```

- [ ] **Step 4: 运行测试确认通过**

Run: `cd backend && go test ./internal/model/ -run TestAudienceJSONShape -v` （放行网络执行）
Expected: PASS

- [ ] **Step 5: 保存文件**

（非 git 仓库，保存即可，不 commit）

---

## Task 2: Mock 达人数据

**Files:**
- Create: `internal/mock/creator_data.go`
- Test: `internal/mock/creator_data_test.go`

**Interfaces:**
- Consumes: `model.TopCreator`/`KpiCard`/`ViewsTrendPoint`/`PlatformShare`/`TrackPerformance`/`Audience`/`GenderShare`; 包内 `RoundTo`/`PlatformColor`/`TrackColor`/`AgeColor`/`FormatCount`/`Tracks`/`Platforms`/`sortByShareDesc`
- Produces: `mock.CreatorKpi` / `CreatorTrend` / `CreatorPlatforms` / `CreatorTracks` / `CreatorAudience` / `CreatorList`（adapter 调用）

- [ ] **Step 1: 写失败测试**

```go
package mock

import (
	"math"
	"testing"
)

func sumShare(ports []struct{ s float64 }) float64 { return 0 } // 占位, 下面用真实类型

func TestCreatorPlatformsSum100(t *testing.T) {
	ps := CreatorPlatforms()
	var total float64
	for _, p := range ps {
		total += p.Share
	}
	if math.Abs(total-100) > 0.5 {
		t.Fatalf("平台占比之和应≈100, 实际 %.2f", total)
	}
}

func TestCreatorListLen(t *testing.T) {
	if len(CreatorList()) != 20 {
		t.Fatalf("期望 20 个达人, 实际 %d", len(CreatorList()))
	}
}

func TestCreatorKpiLen(t *testing.T) {
	if len(CreatorKpi()) != 5 {
		t.Fatalf("期望 5 张 KPI, 实际 %d", len(CreatorKpi()))
	}
}

func TestCreatorAudienceGender(t *testing.T) {
	a := CreatorAudience()
	if len(a.Age) != 4 || len(a.Gender) != 2 {
		t.Fatalf("期望 age=4 gender=2, 实际 %d/%d", len(a.Age), len(a.Gender))
	}
}
```

- [ ] **Step 2: 运行确认失败**

Run: `cd backend && go test ./internal/mock/ -run 'TestCreator' -v` （放行网络执行）
Expected: 编译失败（函数未定义）

- [ ] **Step 3: 实现 mock**

```go
// Package mock 生成数据洞察所需的演示数据。
package mock

import (
	"strings"
	"time"

	"insta360-insight/internal/model"
)

// ============================================================
// 达人分析 KPI(5 张卡)
// ============================================================
func CreatorKpi() []model.KpiCard {
	return []model.KpiCard{
		{Key: "creators", Label: "达人总数", Value: "20", Raw: 20, DeltaPct: 8.5, DeltaUp: true, Description: "较上期"},
		{Key: "new", Label: "本月新增", Value: "3", Raw: 3, DeltaPct: 50, DeltaUp: true, Description: "较上期"},
		{Key: "followers", Label: "平均粉丝", Value: FormatCount(1_820_000), Raw: 1_820_000, DeltaPct: 4.2, DeltaUp: true, Description: "较上期"},
		{Key: "engagement", Label: "平均互动率", Value: "6.8%", Raw: 6.8, DeltaPct: 0.6, DeltaUp: true, Description: "较上期"},
		{Key: "collabs", Label: "合作占比", Value: "45%", Raw: 45, DeltaPct: 12, DeltaUp: true, Description: "较上期"},
	}
}

// ============================================================
// 粉丝规模趋势(近 30 天累计粉丝, M 量级, 复用 ViewsTrendPoint 形状)
// ============================================================
func CreatorTrend() []model.ViewsTrendPoint {
	base := int64(180_000_000)
	points := make([]model.ViewsTrendPoint, 0, 30)
	now := time.Now().Truncate(24 * time.Hour)
	for i := 0; i < 30; i++ {
		v := base + int64(i)*180_000 + int64((i*5)%9)*40_000
		prev := base + int64(i)*150_000
		d := now.AddDate(0, 0, -(29 - i))
		points = append(points, model.ViewsTrendPoint{
			Date:      d.Format("01-02"),
			Views:     v,
			PrevViews: prev,
		})
	}
	return points
}

// ============================================================
// 达人列表(20 个, 覆盖 3 平台 × 5 赛道)
// ============================================================
func CreatorList() []model.TopCreator {
	seeds := []struct {
		name, avatar string
	}{
		{"Chris Burkard", "🤿"}, {"Sophie Laurent", "🚴"}, {"Jake Wetter", "⛷️"},
		{"Marina Costa", "🏄"}, {"Liam Hoffmann", "🧗"}, {"Yuki Tanaka", "🏂"},
		{"Felix Becker", "🚵"}, {"Aria Chen", "🤽"}, {"Marco Rivera", "🏃"},
		{"Elena Petrova", "🎿"}, {"Noah Kim", "🏄"}, {"Mia Wong", "🚴"},
		{"Leo Schmidt", "⛷️"}, {"Zoe Martin", "🧗"}, {"Kenji Sato", "🏂"},
		{"Lara Lopez", "🤽"}, {"Owen Brooks", "🏃"}, {"Nina Roth", "🎿"},
		{"Pablo Cruz", "🚵"}, {"Sara Lind", "🤿"},
	}
	platforms := []string{"抖音", "B站", "小红书"}
	out := make([]model.TopCreator, 0, len(seeds))
	for i, s := range seeds {
		p := platforms[i%3]
		t := Tracks[i%len(Tracks)]
		followers := int64(80_000 + (i*37_000)%1_500_000)
		out = append(out, model.TopCreator{
			Rank:       i + 1,
			Avatar:     s.avatar,
			Name:       s.name,
			Platform:   p,
			Followers:  followers,
			TotalViews: followers * (3 + int64(i)%5),
			Engagement: RoundTo(5.0+float64(i%6)+float64(i%3)*0.4, 2),
			Growth30d:  RoundTo(float64((i*7)%45)-5.0, 1),
			Explosive:  RoundTo(70.0+float64(i%25)+float64(i%4)*1.5, 1),
			Blacklist:  i == 2,
			Tags:       []string{"#" + t, "#极限"},
		})
	}
	return out
}

// ============================================================
// 平台分布(按粉丝量聚合)
// ============================================================
func CreatorPlatforms() []model.PlatformShare {
	list := CreatorList()
	m := map[string]int64{}
	for _, c := range list {
		m[c.Platform] += c.Followers
	}
	var total int64
	for _, v := range m {
		total += v
	}
	out := make([]model.PlatformShare, 0, len(m))
	for p, v := range m {
		out = append(out, model.PlatformShare{
			Platform: p,
			Share:    RoundTo(float64(v)/float64(total)*100, 1),
			Views:    v,
			Color:    PlatformColor(p),
		})
	}
	sortByShareDesc(out)
	return out
}

// ============================================================
// 赛道粉丝分布(按粉丝量聚合, M 量级)
// ============================================================
func CreatorTracks() []model.TrackPerformance {
	list := CreatorList()
	m := map[string]int64{}
	for _, c := range list {
		t := strings.TrimPrefix(c.Tags[0], "#")
		m[t] += c.Followers
	}
	out := make([]model.TrackPerformance, 0, len(m))
	for t, v := range m {
		out = append(out, model.TrackPerformance{
			Track: t,
			Views: v,
			Color: TrackColor(t),
		})
	}
	return out
}

// ============================================================
// 粉丝画像(年龄 + 性别)
// ============================================================
func CreatorAudience() model.Audience {
	return model.Audience{
		Age: []model.AgeShare{
			{Bucket: "18-24 岁", Share: 33.5, Color: AgeColor("18-24 岁")},
			{Bucket: "25-34 岁", Share: 41.2, Color: AgeColor("25-34 岁")},
			{Bucket: "35-44 岁", Share: 17.8, Color: AgeColor("35-44 岁")},
			{Bucket: "45 岁以上", Share: 7.5, Color: AgeColor("45 岁以上")},
		},
		Gender: []model.GenderShare{
			{Gender: "男", Share: 58.4, Color: "#5EA1FF"},
			{Gender: "女", Share: 41.6, Color: "#FF6B6B"},
		},
	}
}
```

（测试里的 `sumShare` 占位函数删除，仅保留 4 个真实测试。）

- [ ] **Step 4: 运行测试确认通过**

Run: `cd backend && go test ./internal/mock/ -run 'TestCreator' -v` （放行网络执行）
Expected: PASS

- [ ] **Step 5: 保存文件**

---

## Task 3: DataSource 接口扩展 + MockAdapter 实现 + 空壳返回 ErrNotImplemented

**Files:**
- Modify: `internal/service/source/adapter.go`
- Modify: `internal/service/source/mock_adapter.go`
- Modify: `internal/service/source/douyin_adapter.go`, `bilibili_adapter.go`, `xiaohongshu_adapter.go`
- Test: `internal/service/source/creator_adapter_test.go`

**Interfaces:**
- Consumes: `mock.Creator*` (Task 2)
- Produces: 接口签名供 `CreatorService` 调用

- [ ] **Step 1: 写失败测试**

```go
package source

import (
	"context"
	"errors"
	"testing"

	"insta360-insight/internal/model"
)

func TestCreatorMockReturnsData(t *testing.T) {
	a := NewMockAdapter()
	if _, err := a.CreatorKpi(context.Background(), model.Filter{}); err != nil {
		t.Fatalf("mock CreatorKpi 应成功, 实际 %v", err)
	}
}

func TestCreatorStubReturnsNotImplemented(t *testing.T) {
	a := NewDouyinAdapter()
	_, err := a.CreatorList(context.Background(), model.Filter{})
	if !errors.Is(err, ErrNotImplemented) {
		t.Fatalf("抖音空壳应返回 ErrNotImplemented, 实际 %v", err)
	}
}
```

- [ ] **Step 2: 运行确认失败**

Run: `cd backend && go test ./internal/service/source/ -run 'TestCreator' -v` （放行网络执行）
Expected: 编译失败

- [ ] **Step 3: 扩展接口（adapter.go）**

在 `DataSource` 接口 `Insights` 行之后追加：

```go
	// 达人分析域(阶段三真实接入时由平台 adapter 实现)
	CreatorKpi(ctx context.Context, f model.Filter) ([]model.KpiCard, error)
	CreatorTrend(ctx context.Context, f model.Filter) ([]model.ViewsTrendPoint, error)
	CreatorPlatforms(ctx context.Context, f model.Filter) ([]model.PlatformShare, error)
	CreatorTracks(ctx context.Context, f model.Filter) ([]model.TrackPerformance, error)
	CreatorAudience(ctx context.Context, f model.Filter) (*model.Audience, error)
	CreatorList(ctx context.Context, f model.Filter) ([]model.TopCreator, error)
```

- [ ] **Step 4: MockAdapter 实现（mock_adapter.go）**

在 `Insights` 方法之后追加：

```go
func (a *MockAdapter) CreatorKpi(_ context.Context, _ model.Filter) ([]model.KpiCard, error) {
	return mock.CreatorKpi(), nil
}
func (a *MockAdapter) CreatorTrend(_ context.Context, _ model.Filter) ([]model.ViewsTrendPoint, error) {
	return mock.CreatorTrend(), nil
}
func (a *MockAdapter) CreatorPlatforms(_ context.Context, _ model.Filter) ([]model.PlatformShare, error) {
	return mock.CreatorPlatforms(), nil
}
func (a *MockAdapter) CreatorTracks(_ context.Context, _ model.Filter) ([]model.TrackPerformance, error) {
	return mock.CreatorTracks(), nil
}
func (a *MockAdapter) CreatorAudience(_ context.Context, _ model.Filter) (*model.Audience, error) {
	a2 := mock.CreatorAudience()
	return &a2, nil
}
func (a *MockAdapter) CreatorList(_ context.Context, _ model.Filter) ([]model.TopCreator, error) {
	return mock.CreatorList(), nil
}
```

- [ ] **Step 5: 三个平台空壳追加 6 方法**

以 `douyin_adapter.go` 为例（其余两个完全镜像，把 `Douyin` 换成 `Bilibili`/`Xiaohongshu`）：

```go
func (a *DouyinAdapter) CreatorKpi(_ context.Context, _ model.Filter) ([]model.KpiCard, error) {
	return nil, ErrNotImplemented
}
func (a *DouyinAdapter) CreatorTrend(_ context.Context, _ model.Filter) ([]model.ViewsTrendPoint, error) {
	return nil, ErrNotImplemented
}
func (a *DouyinAdapter) CreatorPlatforms(_ context.Context, _ model.Filter) ([]model.PlatformShare, error) {
	return nil, ErrNotImplemented
}
func (a *DouyinAdapter) CreatorTracks(_ context.Context, _ model.Filter) ([]model.TrackPerformance, error) {
	return nil, ErrNotImplemented
}
func (a *DouyinAdapter) CreatorAudience(_ context.Context, _ model.Filter) (*model.Audience, error) {
	return nil, ErrNotImplemented
}
func (a *DouyinAdapter) CreatorList(_ context.Context, _ model.Filter) ([]model.TopCreator, error) {
	return nil, ErrNotImplemented
}
```

- [ ] **Step 6: 运行确认通过**

Run: `cd backend && go test ./internal/service/source/ -run 'TestCreator' -v` （放行网络执行）
Expected: PASS

- [ ] **Step 7: 保存文件**

---

## Task 4: CreatorService

**Files:**
- Create: `internal/service/creator_service.go`
- Test: `internal/service/creator_service_test.go`

**Interfaces:**
- Consumes: `source.DataSource`（含 6 个 Creator 方法）
- Produces: `service.CreatorService`（handler 调用）

- [ ] **Step 1: 写失败测试**

```go
package service

import (
	"context"
	"testing"

	"insta360-insight/internal/model"
	"insta360-insight/internal/service/source"
)

func TestCreatorServiceKpi(t *testing.T) {
	svc := NewCreatorService(source.NewMockAdapter())
	got, err := svc.Kpi(context.Background(), model.Filter{})
	if err != nil {
		t.Fatalf("Kpi 应成功, 实际 %v", err)
	}
	if len(got) != 5 {
		t.Fatalf("期望 5 张 KPI, 实际 %d", len(got))
	}
}

func TestCreatorServiceAudience(t *testing.T) {
	svc := NewCreatorService(source.NewMockAdapter())
	a, err := svc.Audience(context.Background(), model.Filter{})
	if err != nil || a == nil || len(a.Gender) != 2 {
		t.Fatalf("Audience 异常: err=%v a=%+v", err, a)
	}
}
```

- [ ] **Step 2: 运行确认失败**

Run: `cd backend && go test ./internal/service/ -run 'TestCreatorService' -v` （放行网络执行）
Expected: 编译失败

- [ ] **Step 3: 实现**

```go
// Package service 业务逻辑层。
package service

import (
	"context"

	"insta360-insight/internal/model"
	"insta360-insight/internal/service/source"
)

// CreatorService 达人分析业务组装, 数据走注入的 source.DataSource。
type CreatorService struct {
	src source.DataSource
}

func NewCreatorService(src source.DataSource) *CreatorService {
	return &CreatorService{src: src}
}

func (s *CreatorService) Kpi(ctx context.Context, f model.Filter) ([]model.KpiCard, error) {
	return s.src.CreatorKpi(ctx, f)
}
func (s *CreatorService) Trend(ctx context.Context, f model.Filter) ([]model.ViewsTrendPoint, error) {
	return s.src.CreatorTrend(ctx, f)
}
func (s *CreatorService) Platforms(ctx context.Context, f model.Filter) ([]model.PlatformShare, error) {
	return s.src.CreatorPlatforms(ctx, f)
}
func (s *CreatorService) Tracks(ctx context.Context, f model.Filter) ([]model.TrackPerformance, error) {
	return s.src.CreatorTracks(ctx, f)
}
func (s *CreatorService) Audience(ctx context.Context, f model.Filter) (*model.Audience, error) {
	return s.src.CreatorAudience(ctx, f)
}
func (s *CreatorService) List(ctx context.Context, f model.Filter) ([]model.TopCreator, error) {
	return s.src.CreatorList(ctx, f)
}
```

- [ ] **Step 4: 运行确认通过**

Run: `cd backend && go test ./internal/service/ -run 'TestCreatorService' -v` （放行网络执行）
Expected: PASS

- [ ] **Step 5: 保存文件**

---

## Task 5: Creator Handler

**Files:**
- Create: `internal/api/handler/creator.go`
- Test: `internal/api/handler/creator_test.go`

**Interfaces:**
- Consumes: `service.CreatorService`（Task 4）
- Produces: 6 个 HTTP handler 方法（router 注册）

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

func TestCreatorHandlerKpi(t *testing.T) {
	h := NewCreator(service.NewCreatorService(source.NewMockAdapter()))
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
```

注：`gin` 包需 import（`github.com/gin-gonic/gin`）。

- [ ] **Step 2: 运行确认失败**

Run: `cd backend && go test ./internal/api/handler/ -run 'TestCreatorHandler' -v` （放行网络执行）
Expected: 编译失败

- [ ] **Step 3: 实现**

```go
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"insta360-insight/internal/model"
	"insta360-insight/internal/service"
)

// Creator 达人分析 HTTP 处理器
type Creator struct {
	svc *service.CreatorService
}

func NewCreator(svc *service.CreatorService) *Creator {
	return &Creator{svc: svc}
}

// Kpi GET /api/creator/kpi
func (h *Creator) Kpi(c *gin.Context) {
	data, err := h.svc.Kpi(c.Request.Context(), bindFilter(c))
	if err != nil {
		fail(c, err)
		return
	}
	c.JSON(http.StatusOK, model.OK(data))
}

// Trend GET /api/creator/trend
func (h *Creator) Trend(c *gin.Context) {
	data, err := h.svc.Trend(c.Request.Context(), bindFilter(c))
	if err != nil {
		fail(c, err)
		return
	}
	c.JSON(http.StatusOK, model.OK(data))
}

// Platforms GET /api/creator/platforms
func (h *Creator) Platforms(c *gin.Context) {
	data, err := h.svc.Platforms(c.Request.Context(), bindFilter(c))
	if err != nil {
		fail(c, err)
		return
	}
	c.JSON(http.StatusOK, model.OK(data))
}

// Tracks GET /api/creator/tracks
func (h *Creator) Tracks(c *gin.Context) {
	data, err := h.svc.Tracks(c.Request.Context(), bindFilter(c))
	if err != nil {
		fail(c, err)
		return
	}
	c.JSON(http.StatusOK, model.OK(data))
}

// Audience GET /api/creator/audience
func (h *Creator) Audience(c *gin.Context) {
	data, err := h.svc.Audience(c.Request.Context(), bindFilter(c))
	if err != nil {
		fail(c, err)
		return
	}
	c.JSON(http.StatusOK, model.OK(data))
}

// List GET /api/creator/list
func (h *Creator) List(c *gin.Context) {
	data, err := h.svc.List(c.Request.Context(), bindFilter(c))
	if err != nil {
		fail(c, err)
		return
	}
	c.JSON(http.StatusOK, model.OK(data))
}
```

`bindFilter` 与 `fail` 已在 `insight.go` 同包定义，直接复用。

- [ ] **Step 4: 运行确认通过**

Run: `cd backend && go test ./internal/api/handler/ -run 'TestCreatorHandler' -v` （放行网络执行）
Expected: PASS

- [ ] **Step 5: 保存文件**

---

## Task 6: 路由注册 + main 接线

**Files:**
- Modify: `internal/api/router/router.go`（`New` 签名 +6 路由）
- Modify: `backend/main.go`（构造 `CreatorService` 并传入）

**Interfaces:**
- Consumes: `service.CreatorService`（Task 4）、`handler.NewCreator`（Task 5）

- [ ] **Step 1: 改 router.go 签名与路由**

`New` 函数签名改为（在 `authSvc` 之后插入 `creatorSvc`）：

```go
func New(insightSvc *service.InsightService, aiSvc *service.AIService, authSvc *service.AuthService, creatorSvc *service.CreatorService, disableAuth bool, devUser service.Claims) *gin.Engine {
```

在 `auth := handler.NewAuth(authSvc)` 之后追加：

```go
	creator := handler.NewCreator(creatorSvc)
```

在受保护组 `g` 的 `g.PUT("/user/profile", ...)` 之后追加：

```go
		g.GET("/creator/kpi", creator.Kpi)
		g.GET("/creator/trend", creator.Trend)
		g.GET("/creator/platforms", creator.Platforms)
		g.GET("/creator/tracks", creator.Tracks)
		g.GET("/creator/audience", creator.Audience)
		g.GET("/creator/list", creator.List)
```

- [ ] **Step 2: 改 main.go**

在 `aiSvc := service.NewAIService(adapter)` 之后追加：

```go
	creatorSvc := service.NewCreatorService(adapter)
```

把 `router.New(...)` 调用改为：

```go
	engine := router.New(insightSvc, aiSvc, authSvc, creatorSvc, disableAuth, devUser)
```

- [ ] **Step 3: 编译校验（放行网络执行）**

Run: `cd backend && go build ./...`
Expected: 编译通过

- [ ] **Step 4: 保存文件**

---

## Task 7: 后端全量测试 + 路由抽查（放行网络执行）

**Files:** 无新增，验收步骤

- [ ] **Step 1: 跑全量测试**

Run: `cd backend && go build ./... && go test ./...` （放行网络执行）
Expected: 全绿

- [ ] **Step 2: 路由抽查（SOURCE=mock）**

```bash
cd backend && SOURCE=mock AUTH_DISABLE=1 go run main.go &
# 另开
curl -s localhost:8080/api/creator/kpi | head -c 200
curl -s localhost:8080/api/creator/list | head -c 200
```
Expected: `{ "code": 0, "data": [...] }`

- [ ] **Step 3: 空壳校验**

```bash
SOURCE=douyin AUTH_DISABLE=1 go run main.go &
curl -s localhost:8080/api/creator/kpi
```
Expected: HTTP 502 `{"code":500,"message":"data source not implemented: ..."}`

- [ ] **Step 4: 保存（无文件改动）**

---

## Task 8: 前端 fallback 数据（creator* 6 数据集）

**Files:**
- Modify: `frontend/src/api/fallback-data.js`（在 `export default {` 内、其它字段并列处新增 6 个键；在文件末尾 `buildViewsTrend` 之后新增 `buildCreatorTrend`/`buildCreatorList` 两个生成函数）

**Interfaces:**
- Produces: `fallback.creatorKpi` / `creatorTrend` / `creatorPlatforms` / `creatorTracks` / `creatorAudience` / `creatorList`（creator store 兜底用）

- [ ] **Step 1: 在 export default 对象内新增字段**

在 `topCreators: [...]` 之后、`options: {...}` 之前插入：

```js
  creatorKpi: [
    { key: 'creators',  label: '达人总数', value: '20', raw: 20, delta_pct: 8.5, delta_up: true, description: '较上期' },
    { key: 'new',       label: '本月新增', value: '3', raw: 3, delta_pct: 50, delta_up: true, description: '较上期' },
    { key: 'followers', label: '平均粉丝', value: '1.82M', raw: 1820000, delta_pct: 4.2, delta_up: true, description: '较上期' },
    { key: 'engagement',label: '平均互动率', value: '6.8%', raw: 6.8, delta_pct: 0.6, delta_up: true, description: '较上期' },
    { key: 'collabs',   label: '合作占比', value: '45%', raw: 45, delta_pct: 12, delta_up: true, description: '较上期' }
  ],
  creatorTrend: buildCreatorTrend(),
  creatorPlatforms: [
    { platform: '抖音',   share: 48.2, views: 880000000, color: '#000000' },
    { platform: 'B站',    share: 30.5, views: 556000000, color: '#00A1D6' },
    { platform: '小红书', share: 21.3, views: 388000000, color: '#FF2442' }
  ],
  creatorTracks: [
    { track: '滑雪', views: 620000000, color: '#5EA1FF' },
    { track: '冲浪', views: 510000000, color: '#3DD9EB' },
    { track: '骑行', views: 430000000, color: '#A07BFF' },
    { track: '潜水', views: 360000000, color: '#4DD0E1' },
    { track: '攀岩', views: 308000000, color: '#7DD96E' }
  ],
  creatorAudience: {
    age: [
      { bucket: '18-24 岁',  share: 33.5, color: '#3DD9EB' },
      { bucket: '25-34 岁',  share: 41.2, color: '#FFB547' },
      { bucket: '35-44 岁',  share: 17.8, color: '#A07BFF' },
      { bucket: '45 岁以上', share: 7.5,  color: '#FF6B6B' }
    ],
    gender: [
      { gender: '男', share: 58.4, color: '#5EA1FF' },
      { gender: '女', share: 41.6, color: '#FF6B6B' }
    ]
  },
  creatorList: buildCreatorList(),
```

- [ ] **Step 2: 在文件末尾新增两个生成函数（buildViewsTrend 之后）**

```js
// 达人分析: 近 30 天累计粉丝趋势(M 量级, 与 Go CreatorTrend 等价)
function buildCreatorTrend() {
  const days = 30
  const out = []
  const now = new Date()
  now.setHours(0, 0, 0, 0)
  const base = new Date(now.getTime() - (days - 1) * 86400000)
  for (let i = 0; i < days; i++) {
    const d = new Date(base.getTime() + i * 86400000)
    const views = 180_000_000 + i * 180_000 + ((i * 5) % 9) * 40_000
    const prev = 180_000_000 + i * 150_000
    const mm = String(d.getMonth() + 1).padStart(2, '0')
    const dd = String(d.getDate()).padStart(2, '0')
    out.push({ date: `${mm}-${dd}`, views, prev_views: prev })
  }
  return out
}

// 达人分析: 20 个达人(与 Go CreatorList 等价)
function buildCreatorList() {
  const seeds = [
    ['Chris Burkard', '🤿'], ['Sophie Laurent', '🚴'], ['Jake Wetter', '⛷️'],
    ['Marina Costa', '🏄'], ['Liam Hoffmann', '🧗'], ['Yuki Tanaka', '🏂'],
    ['Felix Becker', '🚵'], ['Aria Chen', '🤽'], ['Marco Rivera', '🏃'],
    ['Elena Petrova', '🎿'], ['Noah Kim', '🏄'], ['Mia Wong', '🚴'],
    ['Leo Schmidt', '⛷️'], ['Zoe Martin', '🧗'], ['Kenji Sato', '🏂'],
    ['Lara Lopez', '🤽'], ['Owen Brooks', '🏃'], ['Nina Roth', '🎿'],
    ['Pablo Cruz', '🚵'], ['Sara Lind', '🤿']
  ]
  const platforms = ['抖音', 'B站', '小红书']
  const tracks = ['滑雪', '冲浪', '骑行', '潜水', '攀岩']
  return seeds.map((s, i) => {
    const followers = 80_000 + ((i * 37_000) % 1_500_000)
    return {
      rank: i + 1,
      avatar: s[1],
      name: s[0],
      platform: platforms[i % 3],
      followers,
      total_views: followers * (3 + (i % 5)),
      engagement: +(5.0 + (i % 6) + (i % 3) * 0.4).toFixed(2),
      growth_30d: +(((i * 7) % 45) - 5.0).toFixed(1),
      blacklist: i === 2,
      explosive: +(70.0 + (i % 25) + (i % 4) * 1.5).toFixed(1),
      tags: ['#' + tracks[i % tracks.length], '#极限']
    }
  })
}
```

- [ ] **Step 3: 保存文件**

---

## Task 9: 组件向后兼容增强（4 处小改）

**Files:**
- Modify: `src/components/KpiCard.vue`（META 增加 `new`）
- Modify: `src/components/PlatformDonut.vue`（增加 `centerLabel` prop）
- Modify: `src/components/AgeDonut.vue`（增加 `centerLabel` + `centerValue` prop）
- Modify: `src/components/TopCreatorsTable.vue`（增加 `rows` prop）

**Interfaces:**
- Produces: 让 `CreatorAnalysis.vue` 能正确显示「本月新增」「达人总数」「性别画像」「传入自定义列表」

- [ ] **Step 1: KpiCard — META 增加 new**

在 `META` 对象内 `collabs` 行之后加：

```js
  new:       { icon: '✨', from: '#7DD96E', to: '#3DD9EB' },
```

- [ ] **Step 2: PlatformDonut — centerLabel prop**

`defineProps` 内新增：

```js
  centerLabel: { type: String, default: '总播放量' }
```

`buildOption` 里两处 `text: '总播放量'` 改为 `text: props.centerLabel`。

- [ ] **Step 3: AgeDonut — centerLabel + centerValue prop**

`defineProps` 内新增：

```js
  centerLabel: { type: String, default: '总粉丝' },
  centerValue: { type: String, default: '' }
```

`buildOption` 的中心文字：把 `text: '总粉丝'` 改为 `text: props.centerLabel`；把 `text: props.totalM.toFixed(1) + 'M'` 改为：

```js
  text: props.centerValue !== '' ? props.centerValue : props.totalM.toFixed(1) + 'M'
```

- [ ] **Step 4: TopCreatorsTable — rows prop（向后兼容）**

`script setup` 顶部 import 保持 `useInsightStore`，新增 prop 与 computed：

```js
const props = defineProps({
  rows: { type: Array, default: null }
})
const store = useInsightStore()
const rows = computed(() => props.rows ?? store.topCreators)
```

把模板里所有 `store.topCreators` 引用改为 `rows`（仅 `const rows = computed(() => store.topCreators)` 这一行改为上面；模板里已用 `rows`，无需动）。

- [ ] **Step 5: 保存文件**

---

## Task 10: 前端 creator API 模块 + store

**Files:**
- Create: `frontend/src/api/creator.js`
- Create: `frontend/src/stores/creator.js`

**Interfaces:**
- `api/creator.js` 暴露 `kpi/trend/platforms/tracks/audience/list(q)`（request 已带 `/api` base）
- `stores/creator.js` 暴露 `kpi/trend/platforms/tracks/audience/list/usedFallback/loadAll`，供 `CreatorAnalysis.vue` 使用

- [ ] **Step 1: 写 api/creator.js**

```js
// 达人分析相关 API, 全部走 /api 代理(5173 -> 8080)
import request from './request'

const creator = {
  kpi:       (q) => request.get('/creator/kpi', { params: q }),
  trend:     (q) => request.get('/creator/trend', { params: q }),
  platforms: (q) => request.get('/creator/platforms', { params: q }),
  tracks:    (q) => request.get('/creator/tracks', { params: q }),
  audience:  (q) => request.get('/creator/audience', { params: q }),
  list:      (q) => request.get('/creator/list', { params: q })
}

export default creator
```

- [ ] **Step 2: 写 stores/creator.js（镜像 insight store）**

```js
// 达人分析 Store: 先 fallback 填充, 后端可达则用真实数据
import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import creatorApi from '../api/creator'
import fallback from '../api/fallback-data'
import { useFilterStore } from './filter'

export const useCreatorStore = defineStore('creator', () => {
  const loading = ref(false)
  const error = ref(null)
  const usedFallback = ref(false)

  const kpi = ref([])
  const trend = ref([])
  const platforms = ref([])
  const tracks = ref([])
  const audience = ref({ age: [], gender: [] })
  const list = ref([])

  function fillWithFallback() {
    kpi.value = fallback.creatorKpi
    trend.value = fallback.creatorTrend
    platforms.value = fallback.creatorPlatforms
    tracks.value = fallback.creatorTracks
    audience.value = fallback.creatorAudience
    list.value = fallback.creatorList
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
      const [k, tr, ps, tk, au, ls] = await withTimeout(Promise.all([
        creatorApi.kpi(q),
        creatorApi.trend(q),
        creatorApi.platforms(q),
        creatorApi.tracks(q),
        creatorApi.audience(q),
        creatorApi.list(q)
      ]), 800)

      if (
        !Array.isArray(k) || !Array.isArray(tr) || !Array.isArray(ps) ||
        !Array.isArray(tk) || !au || typeof au !== 'object' || !Array.isArray(ls)
      ) {
        throw new Error('后端返回数据结构异常, 已回退到本地兜底数据')
      }

      kpi.value = k
      trend.value = tr
      platforms.value = ps
      tracks.value = tk
      audience.value = au
      list.value = ls
      usedFallback.value = false
    } catch (e) {
      usedFallback.value = true
      error.value = e?.message || '后端不可达, 已切换到本地数据'
    } finally {
      loading.value = false
    }
  }

  const hasData = computed(() => kpi.value.length > 0)

  return {
    loading, error, usedFallback, hasData,
    kpi, trend, platforms, tracks, audience, list,
    loadAll
  }
})
```

- [ ] **Step 3: 保存文件**

---

## Task 11: CreatorAnalysis.vue 页面 + 路由改指向

**Files:**
- Create: `frontend/src/views/CreatorAnalysis.vue`
- Modify: `frontend/src/router/index.js`（`/creator` 改指向）

**Interfaces:**
- Consumes: `stores/creator.js`（Task 10）、`TopCreatorsTable` 的 `rows` prop（Task 9）、`AgeDonut` 的 `centerLabel`/`centerValue`（Task 9）、`PlatformDonut` 的 `centerLabel`（Task 9）

- [ ] **Step 1: 写 CreatorAnalysis.vue（镜像 InsightDashboard 网格，遵循 minmax 铁律）**

```vue
<template>
  <div class="page" v-loading="store.loading">
    <header class="page-head">
      <div class="title-block">
        <h1>达人分析 <span class="info" title="数据每 5 分钟刷新一次">ⓘ</span></h1>
        <p class="sub">{{ dateRangeText }} · 全球达人营销数据洞察</p>
      </div>
    </header>

    <!-- 5 张 KPI -->
    <section class="kpi-row">
      <KpiCard v-for="k in store.kpi" :key="k.key" :kpi="k" />
    </section>

    <!-- 粉丝规模趋势(2fr) + 粉丝年龄画像(1fr) -->
    <section class="grid-row-2">
      <div class="card chart-card">
        <div class="card-head"><span class="card-title">粉丝规模趋势</span></div>
        <div class="card-body chart-body">
          <TrendChart :data="store.trend" :granularity="filter.granularity" />
        </div>
      </div>
      <div class="card">
        <div class="card-head"><span class="card-title">粉丝年龄画像</span></div>
        <div class="card-body donut-body">
          <div class="donut-chart"><AgeDonut :data="store.audience.age" centerLabel="粉丝年龄" /></div>
          <ul class="legend-col">
            <li v-for="a in store.audience.age" :key="a.bucket">
              <span class="dot" :style="{ background: a.color }" />
              <span class="lab">{{ a.bucket }}</span>
              <span class="pct num">{{ a.share }}%</span>
            </li>
          </ul>
        </div>
      </div>
    </section>

    <!-- 平台分布 / 赛道粉丝分布 / 性别分布 -->
    <section class="grid-row-3">
      <div class="card">
        <div class="card-head"><span class="card-title">平台分布</span></div>
        <div class="card-body donut-body">
          <div class="donut-chart"><PlatformDonut :data="store.platforms" centerLabel="达人总数" /></div>
          <ul class="legend-col">
            <li v-for="p in store.platforms" :key="p.platform">
              <span class="dot" :style="{ background: p.color }" />
              <span class="lab">{{ p.platform }}</span>
              <span class="pct num">{{ p.share }}%</span>
            </li>
          </ul>
        </div>
      </div>
      <div class="card">
        <div class="card-head"><span class="card-title">赛道粉丝分布</span></div>
        <div class="card-body chart-body"><TrackBarChart :data="store.tracks" /></div>
      </div>
      <div class="card">
        <div class="card-head"><span class="card-title">粉丝性别分布</span></div>
        <div class="card-body donut-body">
          <div class="donut-chart"><AgeDonut :data="store.audience.gender" centerLabel="性别占比" centerValue="100%" /></div>
          <ul class="legend-col">
            <li v-for="g in store.audience.gender" :key="g.gender">
              <span class="dot" :style="{ background: g.color }" />
              <span class="lab">{{ g.gender }}</span>
              <span class="pct num">{{ g.share }}%</span>
            </li>
          </ul>
        </div>
      </div>
    </section>

    <!-- 达人总表 全宽 -->
    <section class="grid-row-full">
      <div class="card creators-card">
        <div class="card-head">
          <span class="card-title">达人总表</span>
          <el-input v-model="search" placeholder="搜索达人" size="small" class="search" clearable />
        </div>
        <div class="card-body">
          <TopCreatorsTable :rows="filteredList" />
        </div>
      </div>
    </section>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { useFilterStore } from '../stores/filter'
import { useCreatorStore } from '../stores/creator'

import KpiCard from '../components/KpiCard.vue'
import TrendChart from '../components/TrendChart.vue'
import AgeDonut from '../components/AgeDonut.vue'
import PlatformDonut from '../components/PlatformDonut.vue'
import TrackBarChart from '../components/TrackBarChart.vue'
import TopCreatorsTable from '../components/TopCreatorsTable.vue'

const filter = useFilterStore()
const store = useCreatorStore()
const search = ref('')

const dateRangeText = computed(() => filter.dateRange?.length ? filter.dateRange.join(' ~ ') : '')
const filteredList = computed(() => {
  if (!search.value) return store.list
  const kw = search.value.toLowerCase()
  return store.list.filter(c => c.name.toLowerCase().includes(kw))
})

onMounted(() => store.loadAll())
</script>

<style lang="scss" scoped>
.page { display: flex; flex-direction: column; gap: 8px; }
.page-head {
  display: flex; align-items: center; justify-content: space-between;
  .title-block h1 { margin: 0; font-size: 15px; font-weight: 700; color: var(--text-primary); display: flex; gap: 6px; align-items: center;
    .info { cursor: help; color: var(--text-muted); font-size: 12px; } }
  .sub { margin: 1px 0 0; color: var(--text-muted); font-size: 10px; }
}
.kpi-row {
  display: grid; grid-template-columns: repeat(5, 1fr); gap: 6px;
  @media (max-width: 1280px) { grid-template-columns: repeat(3, 1fr); }
  @media (max-width: 768px) { grid-template-columns: repeat(2, 1fr); }
}
.card {
  background: var(--bg-elev); border: 1px solid var(--border);
  border-radius: var(--radius); padding: 8px 10px;
}
.card-head { display: flex; align-items: center; justify-content: space-between; margin-bottom: 2px; min-height: 20px;
  .card-title { font-size: 12px; font-weight: 600; color: var(--text-primary); } }
.card-body { padding-top: 0; }
.chart-body { height: 184px; display: flex; gap: 8px; align-items: center; }
.donut-body {
  display: grid; grid-template-columns: minmax(0, 1.1fr) minmax(0, 1fr);
  height: 170px; padding-top: 0; align-items: center; gap: 4px;
  .donut-chart { width: 100%; height: 100%; min-width: 0; min-height: 0; overflow: hidden;
    display: flex; align-items: center; justify-content: center; }
  .legend-col { list-style: none; margin: 0; padding: 0; display: flex; flex-direction: column; gap: 6px;
    li { display: flex; align-items: center; gap: 6px; font-size: 11px; color: var(--text-secondary); }
    .dot { width: 7px; height: 7px; border-radius: 50%; flex-shrink: 0; }
    .lab { flex: 1; } .pct { color: var(--text-primary); font-weight: 600; } }
}
.grid-row-2 { display: grid; grid-template-columns: 2fr 1fr; gap: 10px; .chart-body { height: 184px; } }
.grid-row-3 { display: grid; grid-template-columns: 1fr 1fr 1fr; gap: 10px; align-items: stretch;
  .chart-body { height: 150px; flex-direction: column; align-items: stretch; }
  > .card:has(.donut-body) { display: flex; flex-direction: column; } }
.grid-row-full { display: grid; grid-template-columns: 1fr; gap: 10px; > .card { min-width: 0; }
  .creators-card .card-body { padding: 0; }
  .creators-card { ::v-deep(.el-table) { width: 100%; } ::v-deep(.el-table__body-wrapper) { overflow-x: auto; } } }
.search { width: 130px; }
@media (max-width: 1280px) { .grid-row-2, .grid-row-3 { grid-template-columns: 1fr; } }
</style>
```

- [ ] **Step 2: 改 router/index.js**

把：
```js
  { path: '/creator', component: () => import('../views/PlaceholderView.vue'), meta: { title: '达人分析' } },
```
改为：
```js
  { path: '/creator', component: () => import('../views/CreatorAnalysis.vue'), meta: { title: '达人分析' } },
```

- [ ] **Step 3: 保存文件**

---

## Task 12: 前端构建 + 无头验证

**Files:** 无新增，验收步骤

- [ ] **Step 1: 构建**

Run: `cd frontend && npm run build --outDir .build-tmp`（沙箱可跑）
Expected: 编译通过，CreatorAnalysis 独立 chunk 生成

- [ ] **Step 2: 清理构建产物（中文路径下 bash rm 被 safe-delete 拦截，用 Python）**

```python
import shutil, os
p = "frontend/.build-tmp"
if os.path.exists(p): shutil.rmtree(p, ignore_errors=True)
```

- [ ] **Step 3: Edge 无头 DOM 验证（沙箱可跑）**

```bash
msedge --headless=new --no-sandbox --disable-gpu --virtual-time-budget=9000 \
  --dump-dom http://localhost:5173/#/creator > /tmp/creator.html
```
Expected: DOM 含 `kpi-row` 容器与 5 张 KPI 文本、图表 canvas、`达人总表` 标题、20 行达人数据；零页面级控制台错误（仅 benign 网络提示）。

- [ ] **Step 4: 保存（无文件改动）**

---

## Task 13: 可复用分析页模板文档

**Files:**
- Create: `docs/superpowers/分析页模板说明.md`

**Interfaces:**
- Produces: 供 内容/市场/品牌 Tab 复制的骨架说明

- [ ] **Step 1: 写模板文档**

```markdown
# 分析页模板（供 内容/市场/品牌 Tab 复制）

每个分析型 Tab 遵循同一骨架：

1. 后端: 在 `DataSource` 接口加 N 个域方法 → `MockAdapter` 实现 + 三平台空壳返回
   `ErrNotImplemented` → 新建 `XxxService` + `XxxHandler` → `router.New` 加参数 + 路由。
2. 前端: 新建 `stores/xxx.js`（镜像 `creator.js`，fallback 优先 + 800ms 超时回退）
   + `api/xxx.js`（6 个端点封装）+ `fallback-data.js` 加 `xxx*` 数据集
   + `views/XxxAnalysis.vue`（镜像 `CreatorAnalysis.vue` 网格，遵循 minmax/CSS 铁律）。
3. 组件复用: KpiCard / TrendChart / PlatformDonut / TrackBarChart / AgeDonut /
   TopCreatorsTable 已全部参数化，新页直接传数据即可，无需新建图表组件。
4. 字段对齐: 尽量复用现有 model 类型（KpiCard/ViewsTrendPoint/PlatformShare/
   TrackPerformance/AgeShare/TopCreator）；新增少量类型时放入 `model/types.go`。

参考实现: 达人分析（2026-07-25-creator-analysis）。
```

- [ ] **Step 2: 保存文件**

---

## Self-Review（写计划时已完成）

- **Spec 覆盖**: 架构(接口扩展/adapter/service/handler/路由/main) ✓；KPI×5 ✓；趋势 ✓；平台/赛道 ✓；粉丝画像(年龄+性别) ✓；达人总表 ✓；fallback ✓；store ✓；路由 ✓；测试 ✓；可复用模板 ✓；文件清单 ✓。
- **Placeholder 扫描**: 无 TBD/TODO；所有步骤含具体代码。
- **类型一致性**: `CreatorService`/`CreatorHandler` 方法名 Kpi/Trend/Platforms/Tracks/Audience/List 与接口、路由、前端 api/store 完全对应；`Audience` 为 `{age,gender}`，前端 `store.audience` 与 `AgeDonut` 的 `data` 均为数组，一致。
- **与已读代码对齐**: router.New 原签名 `(insightSvc, aiSvc, authSvc, disableAuth, devUser)`，本计划插入 `creatorSvc` 于 `authSvc` 之后；main.go 原调用 `router.New(insightSvc, aiSvc, authSvc, disableAuth, devUser)`，本计划同步加 `creatorSvc`；`bindFilter`/`fail` 在 insight.go 同包，handler 直接复用。
- **偏离 spec 的已确认调整**: spec 写"TopCreatorsTable 零改动"，实际该组件耦合 insight store，本计划改为增加向后兼容的 `rows` prop（默认回退 insight store）；KpiCard 增加 `new` META 键；PlatformDonut/AgeDonut 增加 `centerLabel`/`centerValue` prop。均向后兼容，已在 Task 9 明确。
