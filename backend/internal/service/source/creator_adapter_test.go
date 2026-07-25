package source

import (
	"context"
	"errors"
	"testing"

	"insta360-insight/internal/model"
)

func TestCreatorMockReturnsData(t *testing.T) {
	a := NewMockAdapter()
	if _, err := a.CreatorKpi(context.Background(), model.Filter{}); err != nil {
		t.Fatalf("mock CreatorKpi 应成功, 实际 %v", err)
	}
}

func TestCreatorStubReturnsNotImplemented(t *testing.T) {
	a := NewDouyinAdapter()
	_, err := a.CreatorList(context.Background(), model.Filter{})
	if !errors.Is(err, ErrNotImplemented) {
		t.Fatalf("抖音空壳应返回 ErrNotImplemented, 实际 %v", err)
	}
}
