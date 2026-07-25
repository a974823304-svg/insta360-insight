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
