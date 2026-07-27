package source

import (
	"context"

	"insta360-insight/internal/model"
)

// FallbackDataSource 装饰器:优先调用真实 adapter,
// 当真实源返回 ErrNotImplemented(未接入 / 无凭证)或任何 error 时,
// 自动回退到 MockAdapter,保证看板永远有数据、不崩溃。
//
// service / handler 层对它完全无感 —— 它实现了 DataSource 接口。
type FallbackDataSource struct {
	real DataSource
	mock DataSource
}

// NewFallbackDataSource 用真实源 + MockAdapter 兜底构造。
func NewFallbackDataSource(real, mock DataSource) *FallbackDataSource {
	return &FallbackDataSource{real: real, mock: mock}
}

func (d *FallbackDataSource) Name() string { return d.real.Name() }

// try 统一处理回退:real 成功返回;real 返回 ErrNotImplemented 或任意 error,
// 都回退到 mock(mock 实现不会出错,error 被丢弃但类型保持)。
func tryCall[T any](realFn func() (T, error), mockFn func() (T, error)) (T, error) {
	v, err := realFn()
	if err == nil {
		return v, nil
	}
	return mockFn()
}

// ---- 洞察域 ----

func (d *FallbackDataSource) Kpi(ctx context.Context, f model.Filter) ([]model.KpiCard, error) {
	return tryCall(
		func() ([]model.KpiCard, error) { return d.real.Kpi(ctx, f) },
		func() ([]model.KpiCard, error) { return d.mock.Kpi(ctx, f) },
	)
}
func (d *FallbackDataSource) ViewsTrend(ctx context.Context, f model.Filter) ([]model.ViewsTrendPoint, error) {
	return tryCall(
		func() ([]model.ViewsTrendPoint, error) { return d.real.ViewsTrend(ctx, f) },
		func() ([]model.ViewsTrendPoint, error) { return d.mock.ViewsTrend(ctx, f) },
	)
}
func (d *FallbackDataSource) PlatformShare(ctx context.Context, f model.Filter) ([]model.PlatformShare, error) {
	return tryCall(
		func() ([]model.PlatformShare, error) { return d.real.PlatformShare(ctx, f) },
		func() ([]model.PlatformShare, error) { return d.mock.PlatformShare(ctx, f) },
	)
}
func (d *FallbackDataSource) TrackPerformance(ctx context.Context, f model.Filter) ([]model.TrackPerformance, error) {
	return tryCall(
		func() ([]model.TrackPerformance, error) { return d.real.TrackPerformance(ctx, f) },
		func() ([]model.TrackPerformance, error) { return d.mock.TrackPerformance(ctx, f) },
	)
}
func (d *FallbackDataSource) Radar(ctx context.Context, f model.Filter) ([]model.RadarMetric, error) {
	return tryCall(
		func() ([]model.RadarMetric, error) { return d.real.Radar(ctx, f) },
		func() ([]model.RadarMetric, error) { return d.mock.Radar(ctx, f) },
	)
}
func (d *FallbackDataSource) AudienceAge(ctx context.Context, f model.Filter) ([]model.AgeShare, error) {
	return tryCall(
		func() ([]model.AgeShare, error) { return d.real.AudienceAge(ctx, f) },
		func() ([]model.AgeShare, error) { return d.mock.AudienceAge(ctx, f) },
	)
}
func (d *FallbackDataSource) TopCreators(ctx context.Context, f model.Filter) ([]model.TopCreator, error) {
	return tryCall(
		func() ([]model.TopCreator, error) { return d.real.TopCreators(ctx, f) },
		func() ([]model.TopCreator, error) { return d.mock.TopCreators(ctx, f) },
	)
}
func (d *FallbackDataSource) Options(ctx context.Context, f model.Filter) (model.FilterOptions, error) {
	return tryCall(
		func() (model.FilterOptions, error) { return d.real.Options(ctx, f) },
		func() (model.FilterOptions, error) { return d.mock.Options(ctx, f) },
	)
}
func (d *FallbackDataSource) Insights(ctx context.Context, f model.Filter) ([]model.Insight, error) {
	return tryCall(
		func() ([]model.Insight, error) { return d.real.Insights(ctx, f) },
		func() ([]model.Insight, error) { return d.mock.Insights(ctx, f) },
	)
}

// ---- 达人分析域 ----

func (d *FallbackDataSource) CreatorKpi(ctx context.Context, f model.Filter) ([]model.KpiCard, error) {
	return tryCall(
		func() ([]model.KpiCard, error) { return d.real.CreatorKpi(ctx, f) },
		func() ([]model.KpiCard, error) { return d.mock.CreatorKpi(ctx, f) },
	)
}
func (d *FallbackDataSource) CreatorTrend(ctx context.Context, f model.Filter) ([]model.ViewsTrendPoint, error) {
	return tryCall(
		func() ([]model.ViewsTrendPoint, error) { return d.real.CreatorTrend(ctx, f) },
		func() ([]model.ViewsTrendPoint, error) { return d.mock.CreatorTrend(ctx, f) },
	)
}
func (d *FallbackDataSource) CreatorPlatforms(ctx context.Context, f model.Filter) ([]model.PlatformShare, error) {
	return tryCall(
		func() ([]model.PlatformShare, error) { return d.real.CreatorPlatforms(ctx, f) },
		func() ([]model.PlatformShare, error) { return d.mock.CreatorPlatforms(ctx, f) },
	)
}
func (d *FallbackDataSource) CreatorTracks(ctx context.Context, f model.Filter) ([]model.TrackPerformance, error) {
	return tryCall(
		func() ([]model.TrackPerformance, error) { return d.real.CreatorTracks(ctx, f) },
		func() ([]model.TrackPerformance, error) { return d.mock.CreatorTracks(ctx, f) },
	)
}
func (d *FallbackDataSource) CreatorAudience(ctx context.Context, f model.Filter) (*model.Audience, error) {
	return tryCall(
		func() (*model.Audience, error) { return d.real.CreatorAudience(ctx, f) },
		func() (*model.Audience, error) { return d.mock.CreatorAudience(ctx, f) },
	)
}
func (d *FallbackDataSource) CreatorList(ctx context.Context, f model.Filter) ([]model.TopCreator, error) {
	return tryCall(
		func() ([]model.TopCreator, error) { return d.real.CreatorList(ctx, f) },
		func() ([]model.TopCreator, error) { return d.mock.CreatorList(ctx, f) },
	)
}

// ---- 内容分析域 ----

func (d *FallbackDataSource) ContentKpi(ctx context.Context, f model.Filter) ([]model.KpiCard, error) {
	return tryCall(
		func() ([]model.KpiCard, error) { return d.real.ContentKpi(ctx, f) },
		func() ([]model.KpiCard, error) { return d.mock.ContentKpi(ctx, f) },
	)
}
func (d *FallbackDataSource) ContentTrend(ctx context.Context, f model.Filter) ([]model.ViewsTrendPoint, error) {
	return tryCall(
		func() ([]model.ViewsTrendPoint, error) { return d.real.ContentTrend(ctx, f) },
		func() ([]model.ViewsTrendPoint, error) { return d.mock.ContentTrend(ctx, f) },
	)
}
func (d *FallbackDataSource) ContentForms(ctx context.Context, f model.Filter) ([]model.PlatformShare, error) {
	return tryCall(
		func() ([]model.PlatformShare, error) { return d.real.ContentForms(ctx, f) },
		func() ([]model.PlatformShare, error) { return d.mock.ContentForms(ctx, f) },
	)
}
func (d *FallbackDataSource) ContentTopics(ctx context.Context, f model.Filter) ([]model.TrackPerformance, error) {
	return tryCall(
		func() ([]model.TrackPerformance, error) { return d.real.ContentTopics(ctx, f) },
		func() ([]model.TrackPerformance, error) { return d.mock.ContentTopics(ctx, f) },
	)
}
func (d *FallbackDataSource) ContentDurations(ctx context.Context, f model.Filter) ([]model.AgeShare, error) {
	return tryCall(
		func() ([]model.AgeShare, error) { return d.real.ContentDurations(ctx, f) },
		func() ([]model.AgeShare, error) { return d.mock.ContentDurations(ctx, f) },
	)
}
func (d *FallbackDataSource) ContentList(ctx context.Context, f model.Filter) ([]model.ContentItem, error) {
	return tryCall(
		func() ([]model.ContentItem, error) { return d.real.ContentList(ctx, f) },
		func() ([]model.ContentItem, error) { return d.mock.ContentList(ctx, f) },
	)
}

// ---- 市场洞察域 ----

func (d *FallbackDataSource) MarketKpi(ctx context.Context, f model.Filter) ([]model.KpiCard, error) {
	return tryCall(
		func() ([]model.KpiCard, error) { return d.real.MarketKpi(ctx, f) },
		func() ([]model.KpiCard, error) { return d.mock.MarketKpi(ctx, f) },
	)
}
func (d *FallbackDataSource) MarketTrend(ctx context.Context, f model.Filter) ([]model.ViewsTrendPoint, error) {
	return tryCall(
		func() ([]model.ViewsTrendPoint, error) { return d.real.MarketTrend(ctx, f) },
		func() ([]model.ViewsTrendPoint, error) { return d.mock.MarketTrend(ctx, f) },
	)
}
func (d *FallbackDataSource) MarketCompetitors(ctx context.Context, f model.Filter) ([]model.PlatformShare, error) {
	return tryCall(
		func() ([]model.PlatformShare, error) { return d.real.MarketCompetitors(ctx, f) },
		func() ([]model.PlatformShare, error) { return d.mock.MarketCompetitors(ctx, f) },
	)
}
func (d *FallbackDataSource) MarketRegions(ctx context.Context, f model.Filter) ([]model.TrackPerformance, error) {
	return tryCall(
		func() ([]model.TrackPerformance, error) { return d.real.MarketRegions(ctx, f) },
		func() ([]model.TrackPerformance, error) { return d.mock.MarketRegions(ctx, f) },
	)
}
func (d *FallbackDataSource) MarketPrices(ctx context.Context, f model.Filter) ([]model.AgeShare, error) {
	return tryCall(
		func() ([]model.AgeShare, error) { return d.real.MarketPrices(ctx, f) },
		func() ([]model.AgeShare, error) { return d.mock.MarketPrices(ctx, f) },
	)
}
func (d *FallbackDataSource) MarketList(ctx context.Context, f model.Filter) ([]model.Competitor, error) {
	return tryCall(
		func() ([]model.Competitor, error) { return d.real.MarketList(ctx, f) },
		func() ([]model.Competitor, error) { return d.mock.MarketList(ctx, f) },
	)
}

// ---- 品牌分析域 ----

func (d *FallbackDataSource) BrandKpi(ctx context.Context, f model.Filter) ([]model.KpiCard, error) {
	return tryCall(
		func() ([]model.KpiCard, error) { return d.real.BrandKpi(ctx, f) },
		func() ([]model.KpiCard, error) { return d.mock.BrandKpi(ctx, f) },
	)
}
func (d *FallbackDataSource) BrandTrend(ctx context.Context, f model.Filter) ([]model.ViewsTrendPoint, error) {
	return tryCall(
		func() ([]model.ViewsTrendPoint, error) { return d.real.BrandTrend(ctx, f) },
		func() ([]model.ViewsTrendPoint, error) { return d.mock.BrandTrend(ctx, f) },
	)
}
func (d *FallbackDataSource) BrandPlatforms(ctx context.Context, f model.Filter) ([]model.PlatformShare, error) {
	return tryCall(
		func() ([]model.PlatformShare, error) { return d.real.BrandPlatforms(ctx, f) },
		func() ([]model.PlatformShare, error) { return d.mock.BrandPlatforms(ctx, f) },
	)
}
func (d *FallbackDataSource) BrandSentiment(ctx context.Context, f model.Filter) ([]model.AgeShare, error) {
	return tryCall(
		func() ([]model.AgeShare, error) { return d.real.BrandSentiment(ctx, f) },
		func() ([]model.AgeShare, error) { return d.mock.BrandSentiment(ctx, f) },
	)
}
func (d *FallbackDataSource) BrandKeywords(ctx context.Context, f model.Filter) ([]model.TagItem, error) {
	return tryCall(
		func() ([]model.TagItem, error) { return d.real.BrandKeywords(ctx, f) },
		func() ([]model.TagItem, error) { return d.mock.BrandKeywords(ctx, f) },
	)
}
func (d *FallbackDataSource) BrandList(ctx context.Context, f model.Filter) ([]model.PartnerBrand, error) {
	return tryCall(
		func() ([]model.PartnerBrand, error) { return d.real.BrandList(ctx, f) },
		func() ([]model.PartnerBrand, error) { return d.mock.BrandList(ctx, f) },
	)
}
