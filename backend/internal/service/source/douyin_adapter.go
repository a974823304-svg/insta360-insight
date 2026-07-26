package source

import (
	"context"

	"insta360-insight/internal/model"
)

// DouyinAdapter 抖音数据源空壳。阶段三填充 OAuth2 + 字段映射 + 落库。
type DouyinAdapter struct{}

func NewDouyinAdapter() *DouyinAdapter { return &DouyinAdapter{} }

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
