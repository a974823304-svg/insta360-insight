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
