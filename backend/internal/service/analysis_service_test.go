package service

import (
	"context"
	"testing"

	"insta360-insight/internal/model"
	"insta360-insight/internal/service/source"
)

func TestAnalysisServices(t *testing.T) {
	ctx := context.Background()
	f := model.Filter{}
	cs := NewContentService(source.NewMockAdapter())
	if _, e := cs.Kpi(ctx, f); e != nil {
		t.Fatalf("Content.Kpi: %v", e)
	}
	if _, e := cs.List(ctx, f); e != nil {
		t.Fatalf("Content.List: %v", e)
	}
	ms := NewMarketService(source.NewMockAdapter())
	if _, e := ms.Kpi(ctx, f); e != nil {
		t.Fatalf("Market.Kpi: %v", e)
	}
	if _, e := ms.List(ctx, f); e != nil {
		t.Fatalf("Market.List: %v", e)
	}
	bs := NewBrandService(source.NewMockAdapter())
	if _, e := bs.Kpi(ctx, f); e != nil {
		t.Fatalf("Brand.Kpi: %v", e)
	}
	if _, e := bs.Keywords(ctx, f); e != nil {
		t.Fatalf("Brand.Keywords: %v", e)
	}
}
