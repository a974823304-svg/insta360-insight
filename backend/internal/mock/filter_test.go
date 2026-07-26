package mock

import (
	"testing"

	"insta360-insight/internal/model"
)

func TestRenormalizeSums100(t *testing.T) {
	in := []model.PlatformShare{
		{Platform: "抖音", Share: 60, Views: 600},
		{Platform: "B站", Share: 40, Views: 400},
		{Platform: "小红书", Share: 0, Views: 0},
	}
	out := RenormalizePlatformShares(in, []string{"抖音", "B站"})
	var sum float64
	for _, s := range out {
		sum += s.Share
	}
	if len(out) != 2 {
		t.Fatalf("expected 2 items, got %d", len(out))
	}
	if sum < 99.5 || sum > 100.5 {
		t.Fatalf("shares should sum ~100, got %v", sum)
	}
}

func TestScaleFactorBounds(t *testing.T) {
	empty := model.Filter{} // 全空 -> 1.0
	if wf := WeightedFactor(empty); wf != 1.0 {
		t.Fatalf("empty factor want 1.0 got %v", wf)
	}
	one := model.Filter{Platforms: []string{"抖音"}} // 1/3 平台
	wf := WeightedFactor(one)
	if wf < 0.3 || wf > 1.0 {
		t.Fatalf("weighted factor out of [0.3,1.0]: %v", wf)
	}
}

func TestWeightedFactorDeterministic(t *testing.T) {
	f := model.Filter{Platforms: []string{"抖音"}, Tracks: []string{"滑雪"}}
	if WeightedFactor(f) != WeightedFactor(f) {
		t.Fatal("must be deterministic")
	}
}

func TestFilterTopCreatorsShrinks(t *testing.T) {
	rows := []model.TopCreator{
		{Name: "A", Platform: "抖音", Tags: []string{"#滑雪"}},
		{Name: "B", Platform: "B站", Tags: []string{"#冲浪"}},
		{Name: "C", Platform: "抖音", Tags: []string{"#冲浪"}},
	}
	out := FilterTopCreators(rows, model.Filter{Platforms: []string{"抖音"}})
	if len(out) != 2 {
		t.Fatalf("want 2 rows for platform=抖音, got %d", len(out))
	}
	out2 := FilterTopCreators(rows, model.Filter{Tracks: []string{"冲浪"}})
	if len(out2) != 2 {
		t.Fatalf("want 2 rows for track=冲浪, got %d", len(out2))
	}
}

func TestWindowTrendPoints(t *testing.T) {
	pts := make([]model.ViewsTrendPoint, 30)
	for i := range pts {
		pts[i] = model.ViewsTrendPoint{Date: "01-01", Views: 100, PrevViews: 80}
	}
	out := WindowTrend(pts, model.Filter{DateRange: []string{"2024-04-20", "2024-05-20"}})
	if len(out) < 1 {
		t.Fatal("expected points")
	}
	if out[0].Views < 90 || out[0].Views > 110 {
		t.Fatalf("30d window should keep magnitude, got %d", out[0].Views)
	}
}

func TestFormatScaledValuePatterns(t *testing.T) {
	cases := []struct {
		orig string
		raw  float64
		want string
	}{
		{"8.2%", 16.4, "16.4%"},
		{"2.38B", 1_190_000_000, "1.19B"},
		{"46.7M", 23_350_000, "23.4M"},
		{"86.4K", 43_200, "43.2K"},
		{"¥3.2B", 1_600_000_000, "¥1.60B"},
		{"12,856", 6428, "6,428"},
		{"20", 10, "10"},
		{"4.2", 2.1, "2.1"},
	}
	for _, c := range cases {
		if got := FormatScaledValue(c.orig, c.raw); got != c.want {
			t.Fatalf("FormatScaledValue(%q,%v) = %q, want %q", c.orig, c.raw, got, c.want)
		}
	}
}
