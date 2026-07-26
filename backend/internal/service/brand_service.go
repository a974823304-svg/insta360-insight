// Package service 业务逻辑层。
package service

import (
	"context"

	"insta360-insight/internal/model"
	"insta360-insight/internal/service/source"
)

type BrandService struct{ src source.DataSource }

func NewBrandService(src source.DataSource) *BrandService { return &BrandService{src: src} }

func (s *BrandService) Kpi(ctx context.Context, f model.Filter) ([]model.KpiCard, error) {
	return s.src.BrandKpi(ctx, f)
}
func (s *BrandService) Trend(ctx context.Context, f model.Filter) ([]model.ViewsTrendPoint, error) {
	return s.src.BrandTrend(ctx, f)
}
func (s *BrandService) Platforms(ctx context.Context, f model.Filter) ([]model.PlatformShare, error) {
	return s.src.BrandPlatforms(ctx, f)
}
func (s *BrandService) Sentiment(ctx context.Context, f model.Filter) ([]model.AgeShare, error) {
	return s.src.BrandSentiment(ctx, f)
}
func (s *BrandService) Keywords(ctx context.Context, f model.Filter) ([]model.TagItem, error) {
	return s.src.BrandKeywords(ctx, f)
}
func (s *BrandService) List(ctx context.Context, f model.Filter) ([]model.PartnerBrand, error) {
	return s.src.BrandList(ctx, f)
}
