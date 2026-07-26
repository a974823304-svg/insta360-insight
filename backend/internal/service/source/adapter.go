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
	// 达人分析域(阶段三真实接入时由平台 adapter 实现)
	CreatorKpi(ctx context.Context, f model.Filter) ([]model.KpiCard, error)
	CreatorTrend(ctx context.Context, f model.Filter) ([]model.ViewsTrendPoint, error)
	CreatorPlatforms(ctx context.Context, f model.Filter) ([]model.PlatformShare, error)
	CreatorTracks(ctx context.Context, f model.Filter) ([]model.TrackPerformance, error)
	CreatorAudience(ctx context.Context, f model.Filter) (*model.Audience, error)
	CreatorList(ctx context.Context, f model.Filter) ([]model.TopCreator, error)

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
}
