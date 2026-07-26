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

func (a *MockAdapter) Kpi(_ context.Context, f model.Filter) ([]model.KpiCard, error) {
	return mock.Kpi(f), nil
}
func (a *MockAdapter) ViewsTrend(_ context.Context, f model.Filter) ([]model.ViewsTrendPoint, error) {
	return mock.ViewsTrend(f), nil
}
func (a *MockAdapter) PlatformShare(_ context.Context, f model.Filter) ([]model.PlatformShare, error) {
	return mock.PlatformShare(f), nil
}
func (a *MockAdapter) TrackPerformance(_ context.Context, f model.Filter) ([]model.TrackPerformance, error) {
	return mock.TrackPerformance(f), nil
}
func (a *MockAdapter) Radar(_ context.Context, f model.Filter) ([]model.RadarMetric, error) {
	return mock.ExplosiveRadar(f), nil
}
func (a *MockAdapter) AudienceAge(_ context.Context, f model.Filter) ([]model.AgeShare, error) {
	return mock.AgeShare(f), nil
}
func (a *MockAdapter) TopCreators(_ context.Context, f model.Filter) ([]model.TopCreator, error) {
	return mock.TopCreators(f), nil
}
func (a *MockAdapter) Options(_ context.Context, f model.Filter) (model.FilterOptions, error) {
	return mock.FilterOptions(f), nil
}
func (a *MockAdapter) Insights(_ context.Context, f model.Filter) ([]model.Insight, error) {
	return mock.AIInsights(f), nil
}

func (a *MockAdapter) CreatorKpi(_ context.Context, f model.Filter) ([]model.KpiCard, error) {
	return mock.CreatorKpi(f), nil
}
func (a *MockAdapter) CreatorTrend(_ context.Context, f model.Filter) ([]model.ViewsTrendPoint, error) {
	return mock.CreatorTrend(f), nil
}
func (a *MockAdapter) CreatorPlatforms(_ context.Context, f model.Filter) ([]model.PlatformShare, error) {
	return mock.CreatorPlatforms(f), nil
}
func (a *MockAdapter) CreatorTracks(_ context.Context, f model.Filter) ([]model.TrackPerformance, error) {
	return mock.CreatorTracks(f), nil
}
func (a *MockAdapter) CreatorAudience(_ context.Context, f model.Filter) (*model.Audience, error) {
	a2 := mock.CreatorAudience(f)
	return &a2, nil
}
func (a *MockAdapter) CreatorList(_ context.Context, f model.Filter) ([]model.TopCreator, error) {
	return mock.CreatorList(f), nil
}

func (a *MockAdapter) ContentKpi(_ context.Context, f model.Filter) ([]model.KpiCard, error) {
	return mock.ContentKpi(f), nil
}
func (a *MockAdapter) ContentTrend(_ context.Context, f model.Filter) ([]model.ViewsTrendPoint, error) {
	return mock.ContentTrend(f), nil
}
func (a *MockAdapter) ContentForms(_ context.Context, f model.Filter) ([]model.PlatformShare, error) {
	return mock.ContentForms(f), nil
}
func (a *MockAdapter) ContentTopics(_ context.Context, f model.Filter) ([]model.TrackPerformance, error) {
	return mock.ContentTopics(f), nil
}
func (a *MockAdapter) ContentDurations(_ context.Context, f model.Filter) ([]model.AgeShare, error) {
	return mock.ContentDurations(f), nil
}
func (a *MockAdapter) ContentList(_ context.Context, f model.Filter) ([]model.ContentItem, error) {
	return mock.ContentList(f), nil
}
func (a *MockAdapter) MarketKpi(_ context.Context, f model.Filter) ([]model.KpiCard, error) {
	return mock.MarketKpi(f), nil
}
func (a *MockAdapter) MarketTrend(_ context.Context, f model.Filter) ([]model.ViewsTrendPoint, error) {
	return mock.MarketTrend(f), nil
}
func (a *MockAdapter) MarketCompetitors(_ context.Context, f model.Filter) ([]model.PlatformShare, error) {
	return mock.MarketCompetitors(f), nil
}
func (a *MockAdapter) MarketRegions(_ context.Context, f model.Filter) ([]model.TrackPerformance, error) {
	return mock.MarketRegions(f), nil
}
func (a *MockAdapter) MarketPrices(_ context.Context, f model.Filter) ([]model.AgeShare, error) {
	return mock.MarketPrices(f), nil
}
func (a *MockAdapter) MarketList(_ context.Context, f model.Filter) ([]model.Competitor, error) {
	return mock.MarketList(f), nil
}
func (a *MockAdapter) BrandKpi(_ context.Context, f model.Filter) ([]model.KpiCard, error) {
	return mock.BrandKpi(f), nil
}
func (a *MockAdapter) BrandTrend(_ context.Context, f model.Filter) ([]model.ViewsTrendPoint, error) {
	return mock.BrandTrend(f), nil
}
func (a *MockAdapter) BrandPlatforms(_ context.Context, f model.Filter) ([]model.PlatformShare, error) {
	return mock.BrandPlatforms(f), nil
}
func (a *MockAdapter) BrandSentiment(_ context.Context, f model.Filter) ([]model.AgeShare, error) {
	return mock.BrandSentiment(f), nil
}
func (a *MockAdapter) BrandKeywords(_ context.Context, f model.Filter) ([]model.TagItem, error) {
	return mock.BrandKeywords(f), nil
}
func (a *MockAdapter) BrandList(_ context.Context, f model.Filter) ([]model.PartnerBrand, error) {
	return mock.BrandList(f), nil
}
