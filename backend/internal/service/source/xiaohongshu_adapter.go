package source

import (
	"context"

	"insta360-insight/internal/model"
)

// XiaohongshuAdapter 小红书数据源空壳。阶段三填充。
type XiaohongshuAdapter struct{}

func NewXiaohongshuAdapter() *XiaohongshuAdapter { return &XiaohongshuAdapter{} }

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
func (a *XiaohongshuAdapter) CreatorKpi(_ context.Context, _ model.Filter) ([]model.KpiCard, error) {
	return nil, ErrNotImplemented
}
func (a *XiaohongshuAdapter) CreatorTrend(_ context.Context, _ model.Filter) ([]model.ViewsTrendPoint, error) {
	return nil, ErrNotImplemented
}
func (a *XiaohongshuAdapter) CreatorPlatforms(_ context.Context, _ model.Filter) ([]model.PlatformShare, error) {
	return nil, ErrNotImplemented
}
func (a *XiaohongshuAdapter) CreatorTracks(_ context.Context, _ model.Filter) ([]model.TrackPerformance, error) {
	return nil, ErrNotImplemented
}
func (a *XiaohongshuAdapter) CreatorAudience(_ context.Context, _ model.Filter) (*model.Audience, error) {
	return nil, ErrNotImplemented
}
func (a *XiaohongshuAdapter) CreatorList(_ context.Context, _ model.Filter) ([]model.TopCreator, error) {
	return nil, ErrNotImplemented
}

func (a *XiaohongshuAdapter) ContentKpi(_ context.Context, _ model.Filter) ([]model.KpiCard, error) {
	return nil, ErrNotImplemented
}
func (a *XiaohongshuAdapter) ContentTrend(_ context.Context, _ model.Filter) ([]model.ViewsTrendPoint, error) {
	return nil, ErrNotImplemented
}
func (a *XiaohongshuAdapter) ContentForms(_ context.Context, _ model.Filter) ([]model.PlatformShare, error) {
	return nil, ErrNotImplemented
}
func (a *XiaohongshuAdapter) ContentTopics(_ context.Context, _ model.Filter) ([]model.TrackPerformance, error) {
	return nil, ErrNotImplemented
}
func (a *XiaohongshuAdapter) ContentDurations(_ context.Context, _ model.Filter) ([]model.AgeShare, error) {
	return nil, ErrNotImplemented
}
func (a *XiaohongshuAdapter) ContentList(_ context.Context, _ model.Filter) ([]model.ContentItem, error) {
	return nil, ErrNotImplemented
}
func (a *XiaohongshuAdapter) MarketKpi(_ context.Context, _ model.Filter) ([]model.KpiCard, error) {
	return nil, ErrNotImplemented
}
func (a *XiaohongshuAdapter) MarketTrend(_ context.Context, _ model.Filter) ([]model.ViewsTrendPoint, error) {
	return nil, ErrNotImplemented
}
func (a *XiaohongshuAdapter) MarketCompetitors(_ context.Context, _ model.Filter) ([]model.PlatformShare, error) {
	return nil, ErrNotImplemented
}
func (a *XiaohongshuAdapter) MarketRegions(_ context.Context, _ model.Filter) ([]model.TrackPerformance, error) {
	return nil, ErrNotImplemented
}
func (a *XiaohongshuAdapter) MarketPrices(_ context.Context, _ model.Filter) ([]model.AgeShare, error) {
	return nil, ErrNotImplemented
}
func (a *XiaohongshuAdapter) MarketList(_ context.Context, _ model.Filter) ([]model.Competitor, error) {
	return nil, ErrNotImplemented
}
func (a *XiaohongshuAdapter) BrandKpi(_ context.Context, _ model.Filter) ([]model.KpiCard, error) {
	return nil, ErrNotImplemented
}
func (a *XiaohongshuAdapter) BrandTrend(_ context.Context, _ model.Filter) ([]model.ViewsTrendPoint, error) {
	return nil, ErrNotImplemented
}
func (a *XiaohongshuAdapter) BrandPlatforms(_ context.Context, _ model.Filter) ([]model.PlatformShare, error) {
	return nil, ErrNotImplemented
}
func (a *XiaohongshuAdapter) BrandSentiment(_ context.Context, _ model.Filter) ([]model.AgeShare, error) {
	return nil, ErrNotImplemented
}
func (a *XiaohongshuAdapter) BrandKeywords(_ context.Context, _ model.Filter) ([]model.TagItem, error) {
	return nil, ErrNotImplemented
}
func (a *XiaohongshuAdapter) BrandList(_ context.Context, _ model.Filter) ([]model.PartnerBrand, error) {
	return nil, ErrNotImplemented
}
