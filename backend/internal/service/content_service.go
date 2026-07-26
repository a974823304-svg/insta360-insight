// Package service 业务逻辑层。
package service

import (
	"context"

	"insta360-insight/internal/model"
	"insta360-insight/internal/service/source"
)

type ContentService struct{ src source.DataSource }

func NewContentService(src source.DataSource) *ContentService { return &ContentService{src: src} }

func (s *ContentService) Kpi(ctx context.Context, f model.Filter) ([]model.KpiCard, error) {
	return s.src.ContentKpi(ctx, f)
}
func (s *ContentService) Trend(ctx context.Context, f model.Filter) ([]model.ViewsTrendPoint, error) {
	return s.src.ContentTrend(ctx, f)
}
func (s *ContentService) Forms(ctx context.Context, f model.Filter) ([]model.PlatformShare, error) {
	return s.src.ContentForms(ctx, f)
}
func (s *ContentService) Topics(ctx context.Context, f model.Filter) ([]model.TrackPerformance, error) {
	return s.src.ContentTopics(ctx, f)
}
func (s *ContentService) Durations(ctx context.Context, f model.Filter) ([]model.AgeShare, error) {
	return s.src.ContentDurations(ctx, f)
}
func (s *ContentService) List(ctx context.Context, f model.Filter) ([]model.ContentItem, error) {
	return s.src.ContentList(ctx, f)
}
