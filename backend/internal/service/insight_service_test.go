package service

import (
	"context"
	"errors"
	"testing"

	"insta360-insight/internal/model"
	"insta360-insight/internal/service/source"
)

func TestInsightServiceViaMock(t *testing.T) {
	svc := NewInsightService(source.NewMockAdapter())
	ctx := context.Background()
	f := model.Filter{}
	kpi, err := svc.Kpi(ctx, f)
	if err != nil || len(kpi) != 5 {
		t.Fatalf("Kpi() = (%v,%v), want 5", kpi, err)
	}
	if _, err := svc.ViewsTrend(ctx, f); err != nil {
		t.Fatalf("ViewsTrend err = %v", err)
	}
	if _, err := svc.PlatformShare(ctx, f); err != nil {
		t.Fatalf("PlatformShare err = %v", err)
	}
	if _, err := svc.TrackPerformance(ctx, f); err != nil {
		t.Fatalf("TrackPerformance err = %v", err)
	}
	if _, err := svc.Radar(ctx, f); err != nil {
		t.Fatalf("Radar err = %v", err)
	}
	if _, err := svc.AudienceAge(ctx, f); err != nil {
		t.Fatalf("AudienceAge err = %v", err)
	}
	if _, err := svc.TopCreators(ctx, f); err != nil {
		t.Fatalf("TopCreators err = %v", err)
	}
	if _, err := svc.Options(ctx, f); err != nil {
		t.Fatalf("Options err = %v", err)
	}
}

func TestInsightServiceStubPropagatesError(t *testing.T) {
	svc := NewInsightService(&source.DouyinAdapter{})
	_, err := svc.Kpi(context.Background(), model.Filter{})
	if !errors.Is(err, source.ErrNotImplemented) {
		t.Fatalf("Kpi() err = %v, want ErrNotImplemented", err)
	}
}

func TestAIServiceViaMock(t *testing.T) {
	ai := NewAIService(source.NewMockAdapter())
	ins, err := ai.Generate(context.Background(), model.Filter{})
	if err != nil || len(ins) != 3 {
		t.Fatalf("Generate() = (%v,%v), want 3", ins, err)
	}
}
