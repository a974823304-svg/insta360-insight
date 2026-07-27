package source

import (
	"context"
	"errors"
	"testing"

	"insta360-insight/internal/model"
)

func TestAnalysisMockAllReturn(t *testing.T) {
	a := NewMockAdapter()
	ctx := context.Background()
	f := model.Filter{}
	var err error
	if _, err = a.ContentKpi(ctx, f); err != nil {
		t.Fatalf("ContentKpi: %v", err)
	}
	if _, err = a.ContentTrend(ctx, f); err != nil {
		t.Fatalf("ContentTrend: %v", err)
	}
	if _, err = a.ContentForms(ctx, f); err != nil {
		t.Fatalf("ContentForms: %v", err)
	}
	if _, err = a.ContentTopics(ctx, f); err != nil {
		t.Fatalf("ContentTopics: %v", err)
	}
	if _, err = a.ContentDurations(ctx, f); err != nil {
		t.Fatalf("ContentDurations: %v", err)
	}
	if _, err = a.ContentList(ctx, f); err != nil {
		t.Fatalf("ContentList: %v", err)
	}
	if _, err = a.MarketKpi(ctx, f); err != nil {
		t.Fatalf("MarketKpi: %v", err)
	}
	if _, err = a.MarketTrend(ctx, f); err != nil {
		t.Fatalf("MarketTrend: %v", err)
	}
	if _, err = a.MarketCompetitors(ctx, f); err != nil {
		t.Fatalf("MarketCompetitors: %v", err)
	}
	if _, err = a.MarketRegions(ctx, f); err != nil {
		t.Fatalf("MarketRegions: %v", err)
	}
	if _, err = a.MarketPrices(ctx, f); err != nil {
		t.Fatalf("MarketPrices: %v", err)
	}
	if _, err = a.MarketList(ctx, f); err != nil {
		t.Fatalf("MarketList: %v", err)
	}
	if _, err = a.BrandKpi(ctx, f); err != nil {
		t.Fatalf("BrandKpi: %v", err)
	}
	if _, err = a.BrandTrend(ctx, f); err != nil {
		t.Fatalf("BrandTrend: %v", err)
	}
	if _, err = a.BrandPlatforms(ctx, f); err != nil {
		t.Fatalf("BrandPlatforms: %v", err)
	}
	if _, err = a.BrandSentiment(ctx, f); err != nil {
		t.Fatalf("BrandSentiment: %v", err)
	}
	if _, err = a.BrandKeywords(ctx, f); err != nil {
		t.Fatalf("BrandKeywords: %v", err)
	}
	if _, err = a.BrandList(ctx, f); err != nil {
		t.Fatalf("BrandList: %v", err)
	}
}

func TestAnalysisStubAllNotImplemented(t *testing.T) {
	a := NewDouyinAdapter(PlatformConfig{})
	ctx := context.Background()
	f := model.Filter{}
	calls := []func() error{
		func() error { _, e := a.ContentKpi(ctx, f); return e },
		func() error { _, e := a.ContentTrend(ctx, f); return e },
		func() error { _, e := a.ContentForms(ctx, f); return e },
		func() error { _, e := a.ContentTopics(ctx, f); return e },
		func() error { _, e := a.ContentDurations(ctx, f); return e },
		func() error { _, e := a.ContentList(ctx, f); return e },
		func() error { _, e := a.MarketKpi(ctx, f); return e },
		func() error { _, e := a.MarketTrend(ctx, f); return e },
		func() error { _, e := a.MarketCompetitors(ctx, f); return e },
		func() error { _, e := a.MarketRegions(ctx, f); return e },
		func() error { _, e := a.MarketPrices(ctx, f); return e },
		func() error { _, e := a.MarketList(ctx, f); return e },
		func() error { _, e := a.BrandKpi(ctx, f); return e },
		func() error { _, e := a.BrandTrend(ctx, f); return e },
		func() error { _, e := a.BrandPlatforms(ctx, f); return e },
		func() error { _, e := a.BrandSentiment(ctx, f); return e },
		func() error { _, e := a.BrandKeywords(ctx, f); return e },
		func() error { _, e := a.BrandList(ctx, f); return e },
	}
	for i, c := range calls {
		if !errors.Is(c(), ErrNotImplemented) {
			t.Fatalf("抖音空壳第 %d 个方法应返回 ErrNotImplemented", i)
		}
	}
}
