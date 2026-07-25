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
