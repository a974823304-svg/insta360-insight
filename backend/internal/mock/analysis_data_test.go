package mock

import (
	"math"
	"testing"

	"insta360-insight/internal/model"
)

func TestAnalysisKpiLen(t *testing.T) {
	if len(ContentKpi(model.Filter{})) != 5 {
		t.Fatalf("ContentKpi 期望 5 张, 实际 %d", len(ContentKpi(model.Filter{})))
	}
	if len(MarketKpi(model.Filter{})) != 5 {
		t.Fatalf("MarketKpi 期望 5 张, 实际 %d", len(MarketKpi(model.Filter{})))
	}
	if len(BrandKpi(model.Filter{})) != 5 {
		t.Fatalf("BrandKpi 期望 5 张, 实际 %d", len(BrandKpi(model.Filter{})))
	}
}

func TestAnalysisDistSum100(t *testing.T) {
	checks := []struct {
		name string
		ps   []model.PlatformShare
	}{
		{"ContentForms", ContentForms(model.Filter{})},
		{"ContentDurations", agesToPlatform(ContentDurations(model.Filter{}))},
		{"MarketCompetitors", MarketCompetitors(model.Filter{})},
		{"MarketPrices", agesToPlatform(MarketPrices(model.Filter{}))},
		{"BrandPlatforms", BrandPlatforms(model.Filter{})},
		{"BrandSentiment", agesToPlatform(BrandSentiment(model.Filter{}))},
	}
	for _, c := range checks {
		var total float64
		for _, p := range c.ps {
			total += p.Share
		}
		if math.Abs(total-100) > 0.5 {
			t.Fatalf("%s 占比之和应≈100, 实际 %.2f", c.name, total)
		}
	}
}

func TestAnalysisListLen(t *testing.T) {
	if len(ContentList(model.Filter{})) != 15 {
		t.Fatalf("ContentList 期望 15, 实际 %d", len(ContentList(model.Filter{})))
	}
	if len(MarketList(model.Filter{})) != 6 {
		t.Fatalf("MarketList 期望 6, 实际 %d", len(MarketList(model.Filter{})))
	}
	if len(BrandList(model.Filter{})) != 8 {
		t.Fatalf("BrandList 期望 8, 实际 %d", len(BrandList(model.Filter{})))
	}
	if len(BrandKeywords(model.Filter{})) != 8 {
		t.Fatalf("BrandKeywords 期望 8, 实际 %d", len(BrandKeywords(model.Filter{})))
	}
}

// agesToPlatform 把 AgeShare 转成可求和的切片（便于复用占比和校验）
func agesToPlatform(a []model.AgeShare) []model.PlatformShare {
	out := make([]model.PlatformShare, 0, len(a))
	for _, x := range a {
		out = append(out, model.PlatformShare{Platform: x.Bucket, Share: x.Share})
	}
	return out
}
