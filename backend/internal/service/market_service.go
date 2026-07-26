// Package service 业务逻辑层。
package service

import (
	"context"

	"insta360-insight/internal/model"
	"insta360-insight/internal/service/source"
)

type MarketService struct{ src source.DataSource }

func NewMarketService(src source.DataSource) *MarketService { return &MarketService{src: src} }

func (s *MarketService) Kpi(ctx context.Context, f model.Filter) ([]model.KpiCard, error) {
	return s.src.MarketKpi(ctx, f)
}
func (s *MarketService) Trend(ctx context.Context, f model.Filter) ([]model.ViewsTrendPoint, error) {
	return s.src.MarketTrend(ctx, f)
}
func (s *MarketService) Competitors(ctx context.Context, f model.Filter) ([]model.PlatformShare, error) {
	return s.src.MarketCompetitors(ctx, f)
}
func (s *MarketService) Regions(ctx context.Context, f model.Filter) ([]model.TrackPerformance, error) {
	return s.src.MarketRegions(ctx, f)
}
func (s *MarketService) Prices(ctx context.Context, f model.Filter) ([]model.AgeShare, error) {
	return s.src.MarketPrices(ctx, f)
}
func (s *MarketService) List(ctx context.Context, f model.Filter) ([]model.Competitor, error) {
	return s.src.MarketList(ctx, f)
}
