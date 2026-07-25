# 阶段二：数据接入层抽象 (Adapter) 实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 把看板数据从写死的 `mock` 包解耦为可插拔的 `DataSource` 接入层，默认 `MockAdapter` 顶上，并预留抖音/B站/小红书三个平台空壳，来源由 `SOURCE` 环境变量切换；看板渲染与现有一致、前端零改动。

**Architecture:** 新增 `internal/service/source` 包定义 `DataSource` 接口与各平台 adapter；`InsightService` / `AIService` 改为注入 `DataSource`（移除旧的 `*Dataset`）；handler 方法补 `ctx` 并把 adapter 返回的错误翻译成 HTTP 502（供前端 fallback）；`main.go` 按 `SOURCE` 环境变量用工厂构造 adapter。

**Tech Stack:** Go 1.22 + Gin 1.10；纯 Go 依赖（`modernc.org/sqlite` / `golang-jwt` / `golang.org/x/crypto`）；统一响应 `{code,data,message}`。

## Global Constraints

- **CGO_ENABLED=0**；所有新增依赖必须是**纯 Go**（无 CGO），否则本机编译失败。
- 依赖拉取走 `GOPROXY=https://goproxy.cn,direct`（`proxy.golang.org` 本机超时）。
- JWT HS256，secret 取 `JWT_SECRET`；密码 bcrypt，`PasswordHash` 用 `json:"-"` 不泄露。
- SQLite 文件 `backend/data/app.db`，首启动建表；种子 `admin/insta360`。
- `AUTH_DISABLE=1` 或 `ENV=dev` 跳过鉴权注入默认 admin。
- 统一响应 `{code,data,message}`；业务失败 HTTP 200+`{code非0}`；中间件未授权 HTTP 401。
- **阶段二前端零改动**（接口契约不变：`/api/*` 前缀 + 多筛选 querystring）。
- 项目**非 git 仓库**：不擅自 `git init`；subagent 的 commit 步骤改为"是 git 仓库则提交，否则跳过仅保留改动"。
- 数据接口契约字段以 `backend/internal/model/types.go` 为准，新增/改动必须双向与 `frontend/src/api/fallback-data.js` 保持一致（本次前端不改，仅后端内部重构）。

## 设计说明（对 spec 4.1 的细化）

spec 4.1 给的接口是 `FetchCreators/FetchPlayStats/FetchInteraction` 三方法示意。为 1:1 对接现有 `InsightService` 的 8 个数据方法 + AI 洞察，且把 handler 改动降到最低、风险最小，本计划将接口展开为与 service 方法一一对应的 9 个方法（含 `Insights`）。这样既满足 spec 4.6 全部验收标准（数据经 adapter 返回、改 `SOURCE` 切换源、空壳错误前端 fallback），又避免大改 handler 签名。验收标准不变。

---

## File Structure

**新建**
- `backend/internal/service/source/adapter.go` — `DataSource` 接口 + `ErrNotImplemented`
- `backend/internal/service/source/factory.go` — `NewDataSource(kind string)` 工厂，按 `SOURCE` 选源
- `backend/internal/service/source/mock_adapter.go` — `MockAdapter`，用现有 `mock` 包实现全部方法
- `backend/internal/service/source/douyin_adapter.go` — `DouyinAdapter` 空壳，全部方法返回 `ErrNotImplemented`
- `backend/internal/service/source/bilibili_adapter.go` — `BilibiliAdapter` 空壳
- `backend/internal/service/source/xiaohongshu_adapter.go` — `XiaohongshuAdapter` 空壳
- `backend/internal/service/source/adapter_test.go` — TDD 单测（MockAdapter 数据 / 空壳报错 / 工厂映射）
- `backend/internal/service/insight_service_test.go` — 验证 `InsightService` 经 adapter 委托与错误透传

**修改**
- `backend/internal/service/insight_service.go` — 移除 `Dataset`/`NewDataset` 与 `mock` 依赖，注入 `source.DataSource`，方法加 `ctx`
- `backend/internal/service/ai_service.go` — 注入 `source.DataSource`，`Generate` 加 `ctx` 调 `src.Insights`
- `backend/internal/api/handler/insight.go` — 方法加 `ctx`，adapter 错误→HTTP 502 + `model.Fail`
- `backend/main.go` — 移除 `service.NewDataset()`，按 `SOURCE` 用工厂构造 adapter 注入两 service

**不变**：`backend/internal/api/router/router.go`（方法名未变）、前端全部、`mock` 包（仅被 `MockAdapter` 引用）。

---

## Task 1: source 包（接口 / 工厂 / MockAdapter / 3 空壳）+ TDD 单测

**Files:**
- Create: `backend/internal/service/source/adapter.go`
- Create: `backend/internal/service/source/factory.go`
- Create: `backend/internal/service/source/mock_adapter.go`
- Create: `backend/internal/service/source/douyin_adapter.go`
- Create: `backend/internal/service/source/bilibili_adapter.go`
- Create: `backend/internal/service/source/xiaohongshu_adapter.go`
- Test: `backend/internal/service/source/adapter_test.go`

**Interfaces:**
- 本 task 是源头，定义 `DataSource` 接口供 Task 2 的 service 引用。
- Produces: `source.DataSource` 接口、`source.NewDataSource(kind string) (DataSource, error)`、`source.ErrNotImplemented`、`source.NewMockAdapter() *MockAdapter`、各 `*XxxAdapter` 类型。

- [ ] **Step 1: 写失败测试 `adapter_test.go`**

```go
package source

import (
	"context"
	"testing"

	"insta360-insight/internal/model"
)

func TestMockAdapterReturnsData(t *testing.T) {
	a := NewMockAdapter()
	if a.Name() != "mock" {
		t.Fatalf("Name() = %q, want mock", a.Name())
	}
	ctx := context.Background()
	f := model.Filter{}
	kpi, err := a.Kpi(ctx, f)
	if err != nil || len(kpi) != 5 {
		t.Fatalf("Kpi() = (%v, %v), want 5 cards", kpi, err)
	}
	if kpi[0].Label != "达人数" {
		t.Fatalf("Kpi[0].Label = %q, want 达人数", kpi[0].Label)
	}
	vt, err := a.ViewsTrend(ctx, f)
	if err != nil || len(vt) != 30 {
		t.Fatalf("ViewsTrend() len = %d, err=%v", len(vt), err)
	}
	ps, err := a.PlatformShare(ctx, f)
	if err != nil || len(ps) != 3 {
		t.Fatalf("PlatformShare() len = %d, err=%v", len(ps), err)
	}
	tp, err := a.TrackPerformance(ctx, f)
	if err != nil || len(tp) != 5 {
		t.Fatalf("TrackPerformance() len = %d, err=%v", len(tp), err)
	}
	rd, err := a.Radar(ctx, f)
	if err != nil || len(rd) != 5 {
		t.Fatalf("Radar() len = %d, err=%v", len(rd), err)
	}
	ag, err := a.AudienceAge(ctx, f)
	if err != nil || len(ag) != 4 {
		t.Fatalf("AudienceAge() len = %d, err=%v", len(ag), err)
	}
	tc, err := a.TopCreators(ctx, f)
	if err != nil || len(tc) != 10 {
		t.Fatalf("TopCreators() len = %d, err=%v", len(tc), err)
	}
	opt, err := a.Options(ctx, f)
	if err != nil || len(opt.Platforms) != 3 {
		t.Fatalf("Options() platforms = %d, err=%v", len(opt.Platforms), err)
	}
	ins, err := a.Insights(ctx, f)
	if err != nil || len(ins) != 3 {
		t.Fatalf("Insights() len = %d, err=%v", len(ins), err)
	}
}

func TestStubsReturnErrNotImplemented(t *testing.T) {
	stubs := []DataSource{&DouyinAdapter{}, &BilibiliAdapter{}, &XiaohongshuAdapter{}}
	ctx := context.Background()
	f := model.Filter{}
	for _, s := range stubs {
		if _, err := s.Kpi(ctx, f); err != ErrNotImplemented {
			t.Fatalf("%s.Kpi err = %v, want ErrNotImplemented", s.Name(), err)
		}
		if _, err := s.ViewsTrend(ctx, f); err != ErrNotImplemented {
			t.Fatalf("%s.ViewsTrend err = %v", s.Name(), err)
		}
		if _, err := s.PlatformShare(ctx, f); err != ErrNotImplemented {
			t.Fatalf("%s.PlatformShare err = %v", s.Name(), err)
		}
		if _, err := s.TrackPerformance(ctx, f); err != ErrNotImplemented {
			t.Fatalf("%s.TrackPerformance err = %v", s.Name(), err)
		}
		if _, err := s.Radar(ctx, f); err != ErrNotImplemented {
			t.Fatalf("%s.Radar err = %v", s.Name(), err)
		}
		if _, err := s.AudienceAge(ctx, f); err != ErrNotImplemented {
			t.Fatalf("%s.AudienceAge err = %v", s.Name(), err)
		}
		if _, err := s.TopCreators(ctx, f); err != ErrNotImplemented {
			t.Fatalf("%s.TopCreators err = %v", s.Name(), err)
		}
		if _, err := s.Options(ctx, f); err != ErrNotImplemented {
			t.Fatalf("%s.Options err = %v", s.Name(), err)
		}
		if _, err := s.Insights(ctx, f); err != ErrNotImplemented {
			t.Fatalf("%s.Insights err = %v", s.Name(), err)
		}
	}
}

func TestFactory(t *testing.T) {
	cases := []struct {
		kind   string
		want   string
		isErr  bool
	}{
		{"", "mock", false},
		{"mock", "mock", false},
		{"douyin", "douyin", false},
		{"bilibili", "bilibili", false},
		{"xiaohongshu", "xiaohongshu", false},
		{"unknown", "", true},
	}
	for _, c := range cases {
		ds, err := NewDataSource(c.kind)
		if c.isErr {
			if err == nil {
				t.Fatalf("NewDataSource(%q) want error", c.kind)
			}
			continue
		}
		if err != nil {
			t.Fatalf("NewDataSource(%q) err = %v", c.kind, err)
		}
		if ds.Name() != c.want {
			t.Fatalf("NewDataSource(%q).Name() = %q, want %q", c.kind, ds.Name(), c.want)
		}
	}
}
```

- [ ] **Step 2: 运行测试确认失败**

Run: `cd backend && CGO_ENABLED=0 go test ./internal/service/source/`
Expected: FAIL（`package source` 不存在 / 符号未定义）

- [ ] **Step 3: 写实现 `adapter.go`**

```go
package source

import (
	"context"
	"errors"

	"insta360-insight/internal/model"
)

// ErrNotImplemented 表示对应数据源尚未接入(阶段三需平台 appkey / OAuth2)。
var ErrNotImplemented = errors.New("data source not implemented: requires platform appkey / OAuth2")

// DataSource 看板数据的可插拔接入层。
// MockAdapter 用现有 mock 顶上;真实平台 adapter 在阶段三实现(抖音/B站/小红书)。
// 所有方法都接受 ctx + Filter,Filter 当前被 mock 忽略,保留给真实源做行级裁剪。
type DataSource interface {
	Name() string
	Kpi(ctx context.Context, f model.Filter) ([]model.KpiCard, error)
	ViewsTrend(ctx context.Context, f model.Filter) ([]model.ViewsTrendPoint, error)
	PlatformShare(ctx context.Context, f model.Filter) ([]model.PlatformShare, error)
	TrackPerformance(ctx context.Context, f model.Filter) ([]model.TrackPerformance, error)
	Radar(ctx context.Context, f model.Filter) ([]model.RadarMetric, error)
	AudienceAge(ctx context.Context, f model.Filter) ([]model.AgeShare, error)
	TopCreators(ctx context.Context, f model.Filter) ([]model.TopCreator, error)
	Options(ctx context.Context, f model.Filter) (model.FilterOptions, error)
	Insights(ctx context.Context, f model.Filter) ([]model.Insight, error)
}
```

- [ ] **Step 4: 写实现 `factory.go`**

```go
package source

import "fmt"

// NewDataSource 按 SOURCE 环境变量选择数据源。
//   mock(默认) -> MockAdapter; douyin/bilibili/xiaohongshu -> 对应空壳(阶段三填充)
func NewDataSource(kind string) (DataSource, error) {
	switch kind {
	case "", "mock":
		return NewMockAdapter(), nil
	case "douyin":
		return &DouyinAdapter{}, nil
	case "bilibili":
		return &BilibiliAdapter{}, nil
	case "xiaohongshu":
		return &XiaohongshuAdapter{}, nil
	default:
		return nil, fmt.Errorf("unknown data source: %q", kind)
	}
}
```

- [ ] **Step 5: 写实现 `mock_adapter.go`**

```go
package source

import (
	"context"

	"insta360-insight/internal/mock"
	"insta360-insight/internal/model"
)

// MockAdapter 用现有 mock 包顶上看板,保证看板立刻有"真实结构"的数据。
type MockAdapter struct{}

func NewMockAdapter() *MockAdapter { return &MockAdapter{} }

func (a *MockAdapter) Name() string { return "mock" }

func (a *MockAdapter) Kpi(_ context.Context, _ model.Filter) ([]model.KpiCard, error) {
	return mock.Kpi(), nil
}
func (a *MockAdapter) ViewsTrend(_ context.Context, _ model.Filter) ([]model.ViewsTrendPoint, error) {
	return mock.ViewsTrend(), nil
}
func (a *MockAdapter) PlatformShare(_ context.Context, _ model.Filter) ([]model.PlatformShare, error) {
	return mock.PlatformShare(), nil
}
func (a *MockAdapter) TrackPerformance(_ context.Context, _ model.Filter) ([]model.TrackPerformance, error) {
	return mock.TrackPerformance(), nil
}
func (a *MockAdapter) Radar(_ context.Context, _ model.Filter) ([]model.RadarMetric, error) {
	return mock.ExplosiveRadar(), nil
}
func (a *MockAdapter) AudienceAge(_ context.Context, _ model.Filter) ([]model.AgeShare, error) {
	return mock.AgeShare(), nil
}
func (a *MockAdapter) TopCreators(_ context.Context, _ model.Filter) ([]model.TopCreator, error) {
	return mock.TopCreators(), nil
}
func (a *MockAdapter) Options(_ context.Context, _ model.Filter) (model.FilterOptions, error) {
	return mock.FilterOptions(), nil
}
func (a *MockAdapter) Insights(_ context.Context, _ model.Filter) ([]model.Insight, error) {
	return mock.AIInsights(), nil
}
```

- [ ] **Step 6: 写实现 `douyin_adapter.go`**

```go
package source

import (
	"context"

	"insta360-insight/internal/model"
)

// DouyinAdapter 抖音数据源空壳。阶段三填充 OAuth2 + 字段映射 + 落库。
type DouyinAdapter struct{}

func (a *DouyinAdapter) Name() string { return "douyin" }

func (a *DouyinAdapter) Kpi(_ context.Context, _ model.Filter) ([]model.KpiCard, error) {
	return nil, ErrNotImplemented
}
func (a *DouyinAdapter) ViewsTrend(_ context.Context, _ model.Filter) ([]model.ViewsTrendPoint, error) {
	return nil, ErrNotImplemented
}
func (a *DouyinAdapter) PlatformShare(_ context.Context, _ model.Filter) ([]model.PlatformShare, error) {
	return nil, ErrNotImplemented
}
func (a *DouyinAdapter) TrackPerformance(_ context.Context, _ model.Filter) ([]model.TrackPerformance, error) {
	return nil, ErrNotImplemented
}
func (a *DouyinAdapter) Radar(_ context.Context, _ model.Filter) ([]model.RadarMetric, error) {
	return nil, ErrNotImplemented
}
func (a *DouyinAdapter) AudienceAge(_ context.Context, _ model.Filter) ([]model.AgeShare, error) {
	return nil, ErrNotImplemented
}
func (a *DouyinAdapter) TopCreators(_ context.Context, _ model.Filter) ([]model.TopCreator, error) {
	return nil, ErrNotImplemented
}
func (a *DouyinAdapter) Options(_ context.Context, _ model.Filter) (model.FilterOptions, error) {
	return model.FilterOptions{}, ErrNotImplemented
}
func (a *DouyinAdapter) Insights(_ context.Context, _ model.Filter) ([]model.Insight, error) {
	return nil, ErrNotImplemented
}
```

- [ ] **Step 7: 写实现 `bilibili_adapter.go`（与 douyin 完全相同，仅类型名/Name 不同）**

```go
package source

import (
	"context"

	"insta360-insight/internal/model"
)

// BilibiliAdapter B站数据源空壳。阶段三填充。
type BilibiliAdapter struct{}

func (a *BilibiliAdapter) Name() string { return "bilibili" }

func (a *BilibiliAdapter) Kpi(_ context.Context, _ model.Filter) ([]model.KpiCard, error) {
	return nil, ErrNotImplemented
}
func (a *BilibiliAdapter) ViewsTrend(_ context.Context, _ model.Filter) ([]model.ViewsTrendPoint, error) {
	return nil, ErrNotImplemented
}
func (a *BilibiliAdapter) PlatformShare(_ context.Context, _ model.Filter) ([]model.PlatformShare, error) {
	return nil, ErrNotImplemented
}
func (a *BilibiliAdapter) TrackPerformance(_ context.Context, _ model.Filter) ([]model.TrackPerformance, error) {
	return nil, ErrNotImplemented
}
func (a *BilibiliAdapter) Radar(_ context.Context, _ model.Filter) ([]model.RadarMetric, error) {
	return nil, ErrNotImplemented
}
func (a *BilibiliAdapter) AudienceAge(_ context.Context, _ model.Filter) ([]model.AgeShare, error) {
	return nil, ErrNotImplemented
}
func (a *BilibiliAdapter) TopCreators(_ context.Context, _ model.Filter) ([]model.TopCreator, error) {
	return nil, ErrNotImplemented
}
func (a *BilibiliAdapter) Options(_ context.Context, _ model.Filter) (model.FilterOptions, error) {
	return model.FilterOptions{}, ErrNotImplemented
}
func (a *BilibiliAdapter) Insights(_ context.Context, _ model.Filter) ([]model.Insight, error) {
	return nil, ErrNotImplemented
}
```

- [ ] **Step 8: 写实现 `xiaohongshu_adapter.go`（与上同，Name 返回 "xiaohongshu"）**

```go
package source

import (
	"context"

	"insta360-insight/internal/model"
)

// XiaohongshuAdapter 小红书数据源空壳。阶段三填充。
type XiaohongshuAdapter struct{}

func (a *XiaohongshuAdapter) Name() string { return "xiaohongshu" }

func (a *XiaohongshuAdapter) Kpi(_ context.Context, _ model.Filter) ([]model.KpiCard, error) {
	return nil, ErrNotImplemented
}
func (a *XiaohongshuAdapter) ViewsTrend(_ context.Context, _ model.Filter) ([]model.ViewsTrendPoint, error) {
	return nil, ErrNotImplemented
}
func (a *XiaohongshuAdapter) PlatformShare(_ context.Context, _ model.Filter) ([]model.PlatformShare, error) {
	return nil, ErrNotImplemented
}
func (a *XiaohongshuAdapter) TrackPerformance(_ context.Context, _ model.Filter) ([]model.TrackPerformance, error) {
	return nil, ErrNotImplemented
}
func (a *XiaohongshuAdapter) Radar(_ context.Context, _ model.Filter) ([]model.RadarMetric, error) {
	return nil, ErrNotImplemented
}
func (a *XiaohongshuAdapter) AudienceAge(_ context.Context, _ model.Filter) ([]model.AgeShare, error) {
	return nil, ErrNotImplemented
}
func (a *XiaohongshuAdapter) TopCreators(_ context.Context, _ model.Filter) ([]model.TopCreator, error) {
	return nil, ErrNotImplemented
}
func (a *XiaohongshuAdapter) Options(_ context.Context, _ model.Filter) (model.FilterOptions, error) {
	return model.FilterOptions{}, ErrNotImplemented
}
func (a *XiaohongshuAdapter) Insights(_ context.Context, _ model.Filter) ([]model.Insight, error) {
	return nil, ErrNotImplemented
}
```

- [ ] **Step 9: 运行测试确认通过**

Run: `cd backend && CGO_ENABLED=0 go test ./internal/service/source/ -v`
Expected: PASS（3 个测试函数全绿）

- [ ] **Step 10: 提交（若已是 git 仓库）**

```bash
git add backend/internal/service/source/
git commit -m "feat(phase2): add DataSource adapter layer with MockAdapter + 3 platform stubs"
```
（非 git 仓库则跳过，仅保留改动）

---

## Task 2: service + handler 重构（注入 DataSource, 补 ctx, 错误透传）

**Files:**
- Modify: `backend/internal/service/insight_service.go`
- Modify: `backend/internal/service/ai_service.go`
- Modify: `backend/internal/api/handler/insight.go`
- Test: `backend/internal/service/insight_service_test.go`

**Interfaces:**
- Consumes: `source.DataSource` 接口、`source.NewMockAdapter()`、`source.ErrNotImplemented`（来自 Task 1）
- Produces: `service.NewInsightService(src source.DataSource)`、`service.NewAIService(src source.DataSource)`、`(*InsightService).Kpi(ctx, f) ...`、`(*AIService).Generate(ctx, f)`、handler 各方法错误处理（adapter 错误→HTTP 502）

- [ ] **Step 1: 写失败测试 `insight_service_test.go`**

```go
package service

import (
	"context"
	"errors"
	"testing"

	"insta360-insight/internal/model"
	"insta360-insight/internal/service/source"
)

func TestInsightServiceViaMock(t *testing.T) {
	svc := NewInsightService(source.NewMockAdapter())
	ctx := context.Background()
	f := model.Filter{}
	kpi, err := svc.Kpi(ctx, f)
	if err != nil || len(kpi) != 5 {
		t.Fatalf("Kpi() = (%v,%v), want 5", kpi, err)
	}
	if _, err := svc.ViewsTrend(ctx, f); err != nil {
		t.Fatalf("ViewsTrend err = %v", err)
	}
	if _, err := svc.PlatformShare(ctx, f); err != nil {
		t.Fatalf("PlatformShare err = %v", err)
	}
	if _, err := svc.TrackPerformance(ctx, f); err != nil {
		t.Fatalf("TrackPerformance err = %v", err)
	}
	if _, err := svc.Radar(ctx, f); err != nil {
		t.Fatalf("Radar err = %v", err)
	}
	if _, err := svc.AudienceAge(ctx, f); err != nil {
		t.Fatalf("AudienceAge err = %v", err)
	}
	if _, err := svc.TopCreators(ctx, f); err != nil {
		t.Fatalf("TopCreators err = %v", err)
	}
	if _, err := svc.Options(ctx, f); err != nil {
		t.Fatalf("Options err = %v", err)
	}
	if ins, err := svc.fooInsights(ctx, f); err != nil || len(ins) != 3 {
		t.Fatalf("Insights() = (%v,%v), want 3", ins, err)
	}
}

func TestInsightServiceStubPropagatesError(t *testing.T) {
	svc := NewInsightService(&source.DouyinAdapter{})
	_, err := svc.Kpi(context.Background(), model.Filter{})
	if !errors.Is(err, source.ErrNotImplemented) {
		t.Fatalf("Kpi() err = %v, want ErrNotImplemented", err)
	}
}
```

> 注：`fooInsights` 为测试辅助，需在 InsightService 上临时暴露一个调用 `src.Insights` 的方法（见 Step 3 的 `Insights` 方法——把它作为 `InsightService` 的公开方法即可，handler 不依赖它，AIService 才用 `src.Insights`；这里测试直接通过 `aiSvc` 验证更干净。改测试为下面 Step 1b。

- [ ] **Step 1b: 用更干净的测试（去掉 fooInsights）**

```go
package service

import (
	"context"
	"errors"
	"testing"

	"insta360-insight/internal/model"
	"insta360-insight/internal/service/source"
)

func TestInsightServiceViaMock(t *testing.T) {
	svc := NewInsightService(source.NewMockAdapter())
	ctx := context.Background()
	f := model.Filter{}
	kpi, err := svc.Kpi(ctx, f)
	if err != nil || len(kpi) != 5 {
		t.Fatalf("Kpi() = (%v,%v), want 5", kpi, err)
	}
	if _, err := svc.ViewsTrend(ctx, f); err != nil {
		t.Fatalf("ViewsTrend err = %v", err)
	}
	if _, err := svc.PlatformShare(ctx, f); err != nil {
		t.Fatalf("PlatformShare err = %v", err)
	}
	if _, err := svc.TrackPerformance(ctx, f); err != nil {
		t.Fatalf("TrackPerformance err = %v", err)
	}
	if _, err := svc.Radar(ctx, f); err != nil {
		t.Fatalf("Radar err = %v", err)
	}
	if _, err := svc.AudienceAge(ctx, f); err != nil {
		t.Fatalf("AudienceAge err = %v", err)
	}
	if _, err := svc.TopCreators(ctx, f); err != nil {
		t.Fatalf("TopCreators err = %v", err)
	}
	if _, err := svc.Options(ctx, f); err != nil {
		t.Fatalf("Options err = %v", err)
	}
}

func TestInsightServiceStubPropagatesError(t *testing.T) {
	svc := NewInsightService(&source.DouyinAdapter{})
	_, err := svc.Kpi(context.Background(), model.Filter{})
	if !errors.Is(err, source.ErrNotImplemented) {
		t.Fatalf("Kpi() err = %v, want ErrNotImplemented", err)
	}
}

func TestAIServiceViaMock(t *testing.T) {
	ai := NewAIService(source.NewMockAdapter())
	ins, err := ai.Generate(context.Background(), model.Filter{})
	if err != nil || len(ins) != 3 {
		t.Fatalf("Generate() = (%v,%v), want 3", ins, err)
	}
}
```

- [ ] **Step 2: 运行测试确认失败**

Run: `cd backend && CGO_ENABLED=0 go test ./internal/service/`
Expected: FAIL（`NewInsightService` 签名不匹配 / `InsightService` 无 `ctx` 方法）

- [ ] **Step 3: 重写 `insight_service.go`**

```go
// Package service 业务逻辑层。
//
// 阶段二起：所有数据访问走注入的 source.DataSource(可插拔接入层),
// 真实环境由 source 包下的各平台 adapter 实现(目前 MockAdapter 顶上)。
// handler 层方法签名不变,仅补 ctx。
package service

import (
	"context"

	"insta360-insight/internal/model"
	"insta360-insight/internal/service/source"
)

// InsightService 负责组装 Dashboard 所需的所有数据。
// 数据来源通过注入的 source.DataSource 抽象,当前为 MockAdapter。
type InsightService struct {
	src source.DataSource
}

// NewInsightService 构造函数,注入数据源。
func NewInsightService(src source.DataSource) *InsightService {
	return &InsightService{src: src}
}

// Kpi 看板 KPI
func (s *InsightService) Kpi(ctx context.Context, f model.Filter) ([]model.KpiCard, error) {
	return s.src.Kpi(ctx, f)
}

// ViewsTrend 播放量趋势
func (s *InsightService) ViewsTrend(ctx context.Context, f model.Filter) ([]model.ViewsTrendPoint, error) {
	return s.src.ViewsTrend(ctx, f)
}

// PlatformShare 平台分布
func (s *InsightService) PlatformShare(ctx context.Context, f model.Filter) ([]model.PlatformShare, error) {
	return s.src.PlatformShare(ctx, f)
}

// TrackPerformance 运动赛道表现
func (s *InsightService) TrackPerformance(ctx context.Context, f model.Filter) ([]model.TrackPerformance, error) {
	return s.src.TrackPerformance(ctx, f)
}

// Radar 引爆力雷达
func (s *InsightService) Radar(ctx context.Context, f model.Filter) ([]model.RadarMetric, error) {
	return s.src.Radar(ctx, f)
}

// AudienceAge 粉丝画像
func (s *InsightService) AudienceAge(ctx context.Context, f model.Filter) ([]model.AgeShare, error) {
	return s.src.AudienceAge(ctx, f)
}

// TopCreators 热门达人
func (s *InsightService) TopCreators(ctx context.Context, f model.Filter) ([]model.TopCreator, error) {
	return s.src.TopCreators(ctx, f)
}

// Options 筛选面板可选项
func (s *InsightService) Options(ctx context.Context, f model.Filter) (model.FilterOptions, error) {
	return s.src.Options(ctx, f)
}
```

- [ ] **Step 4: 重写 `ai_service.go`**

```go
package service

import (
	"context"

	"insta360-insight/internal/model"
	"insta360-insight/internal/service/source"
)

// ============================================================
// AIService AI 洞察服务
// ============================================================
//
// 当前实现:基于 mock 的本地兜底洞察。
// 数据源经 source.DataSource 注入,切换 SOURCE 时洞察也跟随源。
// 生产环境:把 Generate() 改成 HTTP 调用 ai/ (Python FastAPI) 的 /v1/insights。
type AIService struct {
	src source.DataSource
}

func NewAIService(src source.DataSource) *AIService { return &AIService{src: src} }

// Generate 根据当前数据源返回洞察。
//   真实实现:POST http://ai-service/v1/insights,body 是 Filter + 当前 KPI/趋势快照
func (s *AIService) Generate(ctx context.Context, f model.Filter) ([]model.Insight, error) {
	return s.src.Insights(ctx, f)
}
```

- [ ] **Step 5: 重写 `handler/insight.go`（补 ctx + 错误透传 HTTP 502）**

```go
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"insta360-insight/internal/model"
	"insta360-insight/internal/service"
)

// Insight 数据洞察 HTTP 处理器
type Insight struct {
	svc *service.InsightService
	ai  *service.AIService
}

func NewInsight(svc *service.InsightService, ai *service.AIService) *Insight {
	return &Insight{svc: svc, ai: ai}
}

// bindFilter 解析 querystring 中的筛选条件。
func bindFilter(c *gin.Context) model.Filter {
	return model.Filter{
		DateRange: c.QueryArray("date_range"),
		Regions:   c.QueryArray("regions"),
		Tracks:    c.QueryArray("tracks"),
		Platforms: c.QueryArray("platforms"),
		AgeBands:  c.QueryArray("age_bands"),
	}
}

func fail(c *gin.Context, err error) {
	c.JSON(http.StatusBadGateway, model.Fail(500, err.Error()))
}

// Kpi GET /api/kpi
func (h *Insight) Kpi(c *gin.Context) {
	data, err := h.svc.Kpi(c.Request.Context(), bindFilter(c))
	if err != nil {
		fail(c, err)
		return
	}
	c.JSON(http.StatusOK, model.OK(data))
}

// ViewsTrend GET /api/views-trend
func (h *Insight) ViewsTrend(c *gin.Context) {
	data, err := h.svc.ViewsTrend(c.Request.Context(), bindFilter(c))
	if err != nil {
		fail(c, err)
		return
	}
	c.JSON(http.StatusOK, model.OK(data))
}

// PlatformShare GET /api/platform-distribution
func (h *Insight) PlatformShare(c *gin.Context) {
	data, err := h.svc.PlatformShare(c.Request.Context(), bindFilter(c))
	if err != nil {
		fail(c, err)
		return
	}
	c.JSON(http.StatusOK, model.OK(data))
}

// TrackPerformance GET /api/track-performance
func (h *Insight) TrackPerformance(c *gin.Context) {
	data, err := h.svc.TrackPerformance(c.Request.Context(), bindFilter(c))
	if err != nil {
		fail(c, err)
		return
	}
	c.JSON(http.StatusOK, model.OK(data))
}

// Radar GET /api/explosive-radar
func (h *Insight) Radar(c *gin.Context) {
	data, err := h.svc.Radar(c.Request.Context(), bindFilter(c))
	if err != nil {
		fail(c, err)
		return
	}
	c.JSON(http.StatusOK, model.OK(data))
}

// AudienceAge GET /api/audience-age
func (h *Insight) AudienceAge(c *gin.Context) {
	data, err := h.svc.AudienceAge(c.Request.Context(), bindFilter(c))
	if err != nil {
		fail(c, err)
		return
	}
	c.JSON(http.StatusOK, model.OK(data))
}

// TopCreators GET /api/top-creators
func (h *Insight) TopCreators(c *gin.Context) {
	data, err := h.svc.TopCreators(c.Request.Context(), bindFilter(c))
	if err != nil {
		fail(c, err)
		return
	}
	c.JSON(http.StatusOK, model.OK(data))
}

// Insights GET /api/insights —— AI 关键洞察
func (h *Insight) Insights(c *gin.Context) {
	data, err := h.ai.Generate(c.Request.Context(), bindFilter(c))
	if err != nil {
		fail(c, err)
		return
	}
	c.JSON(http.StatusOK, model.OK(data))
}

// FilterOptions GET /api/filters/options
func (h *Insight) FilterOptions(c *gin.Context) {
	data, err := h.svc.Options(c.Request.Context(), bindFilter(c))
	if err != nil {
		fail(c, err)
		return
	}
	c.JSON(http.StatusOK, model.OK(data))
}
```

- [ ] **Step 6: 运行测试确认通过**

Run: `cd backend && CGO_ENABLED=0 go test ./internal/service/ -v`
Expected: PASS（含新增的 3 个测试）

- [ ] **Step 7: 提交（若已是 git 仓库）**

```bash
git add backend/internal/service/ backend/internal/api/handler/insight.go backend/internal/service/insight_service_test.go
git commit -m "refactor(phase2): inject DataSource into Insight/AIService, ctx + 502 on error"
```
（非 git 仓库则跳过）

---

## Task 3: main.go 接线 + 构建 + 回归 + E2E 验收

**Files:**
- Modify: `backend/main.go`
- 验收: `backend` 构建 + `go test ./...` + 后端 E2E 脚本

**Interfaces:**
- Consumes: `source.NewDataSource(kind string) (DataSource, error)`（Task 1）、`service.NewInsightService(src)`、`service.NewAIService(src)`（Task 2）
- Produces: 可启动的后端，按 `SOURCE` 切换数据源；`go build ./...` 通过；全量测试通过；E2E 验证渲染一致 + 空壳 502

- [ ] **Step 1: 重写 `main.go` 数据来源部分**

把原 `dataset := service.NewDataset()` 及两处 `NewXxxService(dataset)` 改为：

```go
	// 1. 选择数据源(阶段二:可插拔接入层 adapter)
	srcKind := os.Getenv("SOURCE")
	adapter, err := source.NewDataSource(srcKind)
	if err != nil {
		log.Fatalf("初始化数据源失败: %v", err)
	}
	log.Printf("[insta360-insight] data source = %s", adapter.Name())

	// 2. 装载业务服务(依赖注入,后续方便做单测)
	insightSvc := service.NewInsightService(adapter)
	aiSvc := service.NewAIService(adapter)
```

并在 import 块加入 `"insta360-insight/internal/service/source"`；删除 `service` 包里不再需要的 `NewDataset` 引用（本文件已不再调用）。

- [ ] **Step 2: 确认 `insight_service.go` 已无 `Dataset`/`NewDataset`/`mock` 残留**

检查 `backend/internal/service/insight_service.go` 顶部 import 仅含 `context`、`model`、`service/source`，且无 `Dataset`/`NewDataset` 定义（已在 Task 2 删除）。如有遗漏一并清理。

- [ ] **Step 3: 构建**

Run: `cd backend && CGO_ENABLED=0 go build ./...`
Expected: 构建成功（无输出 / 退出码 0）

- [ ] **Step 4: 全量回归测试**

Run: `cd backend && CGO_ENABLED=0 go test ./...`
Expected: 全部 PASS（model / store / service / source / handler / middleware）

- [ ] **Step 5: E2E —— 默认 SOURCE=mock 渲染一致**

```bash
cd backend
AUTH_DISABLE=1 CGO_ENABLED=0 go build -o /tmp/be2.exe . && /tmp/be2.exe &
sleep 1
curl -s localhost:8080/api/kpi | head -c 200
# 期望: {"code":0,"data":[{"key":"creators","label":"达人数","value":"12,856",...}]}
curl -s localhost:8080/api/top-creators | head -c 120
# 期望含 Chris Burkard
kill %1 2>/dev/null
```

- [ ] **Step 6: E2E —— 空壳 SOURCE=douyin 返回 502（前端将 fallback）**

```bash
cd backend
AUTH_DISABLE=1 SOURCE=douyin CGO_ENABLED=0 go build -o /tmp/be3.exe . && /tmp/be3.exe &
sleep 1
curl -s -o /dev/null -w "%{http_code}\n" localhost:8080/api/kpi
# 期望: 502
curl -s localhost:8080/api/kpi
# 期望: {"code":500,"message":"data source not implemented: ..."}
kill %1 2>/dev/null
```

- [ ] **Step 7: 提交（若已是 git 仓库）**

```bash
git add backend/main.go
git commit -m "feat(phase2): wire SOURCE env into main, select DataSource adapter"
```
（非 git 仓库则跳过）

---

## Self-Review

**1. Spec 覆盖（spec 4.1–4.6）**
- 4.1 接口定义 → Task 1 `adapter.go` `DataSource` 接口（9 方法，覆盖 spec 示意并 1:1 对齐 service）
- 4.2 MockAdapter → Task 1 `mock_adapter.go`
- 4.3 三平台空壳 → Task 1 `douyin/bilibili/xiaohongshu_adapter.go`，返回 `ErrNotImplemented`
- 4.4 insight_service 改造 + SOURCE 配置 → Task 2（注入）+ Task 3（main 接线）
- 4.5 前端零改动 → 全计划未触及前端；handler 方法名/路径不变
- 4.6 验收：数据经 adapter（Task 2/3 E2E）、可切换源（Task 3 空壳 502）、真实接入预留（Task 1 空壳就位）
- 全覆盖，无遗漏。

**2. 占位符扫描**：无 `TBD`/`TODO`/`implement later`；每个代码步骤均给出完整实现；测试含真实断言。空壳方法体显式 `return nil, ErrNotImplemented` 并附 `TODO` 注释说明阶段三填充（这是有意的设计占位，非计划缺失）。

**3. 类型一致性**：
- `DataSource` 方法签名（ctx, f）→ `MockAdapter`/`*Adapter`/service 委托/工厂 完全一致。
- `NewInsightService(source.DataSource)` / `NewAIService(source.DataSource)` 在 Task 1 定义、Task 2 使用、Task 3 main 调用，签名一致。
- `model.Filter`、`model.KpiCard` 等类型来自 `model` 包，source/service/handler 同一导入，无重名冲突。
- handler 错误统一 `model.Fail(500, err.Error())` + HTTP 502，`model.Fail` 签名 `(int, string)` 与 `types.go` 一致。
- `router.go` 注册的方法名（Kpi/ViewsTrend/PlatformShare/TrackPerformance/Radar/AudienceAge/TopCreators/Insights/FilterOptions）与重写后的 handler 方法名完全一致。
