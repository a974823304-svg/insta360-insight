package mock

import (
	"math"
	"testing"

	"insta360-insight/internal/model"
)

func TestCreatorPlatformsSum100(t *testing.T) {
	ps := CreatorPlatforms(model.Filter{})
	var total float64
	for _, p := range ps {
		total += p.Share
	}
	if math.Abs(total-100) > 0.5 {
		t.Fatalf("平台占比之和应≈100, 实际 %.2f", total)
	}
}

func TestCreatorListLen(t *testing.T) {
	if len(CreatorList(model.Filter{})) != 20 {
		t.Fatalf("期望 20 个达人, 实际 %d", len(CreatorList(model.Filter{})))
	}
}

func TestCreatorKpiLen(t *testing.T) {
	if len(CreatorKpi(model.Filter{})) != 5 {
		t.Fatalf("期望 5 张 KPI, 实际 %d", len(CreatorKpi(model.Filter{})))
	}
}

func TestCreatorAudienceGender(t *testing.T) {
	a := CreatorAudience(model.Filter{})
	if len(a.Age) != 4 || len(a.Gender) != 2 {
		t.Fatalf("期望 age=4 gender=2, 实际 %d/%d", len(a.Age), len(a.Gender))
	}
}
