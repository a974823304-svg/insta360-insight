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
