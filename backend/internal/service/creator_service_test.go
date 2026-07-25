package service

import (
	"context"
	"testing"

	"insta360-insight/internal/model"
	"insta360-insight/internal/service/source"
)

func TestCreatorServiceKpi(t *testing.T) {
	svc := NewCreatorService(source.NewMockAdapter())
	got, err := svc.Kpi(context.Background(), model.Filter{})
	if err != nil {
		t.Fatalf("Kpi 应成功, 实际 %v", err)
	}
	if len(got) != 5 {
		t.Fatalf("期望 5 张 KPI, 实际 %d", len(got))
	}
}

func TestCreatorServiceAudience(t *testing.T) {
	svc := NewCreatorService(source.NewMockAdapter())
	a, err := svc.Audience(context.Background(), model.Filter{})
	if err != nil || a == nil || len(a.Gender) != 2 {
		t.Fatalf("Audience 异常: err=%v a=%+v", err, a)
	}
}
