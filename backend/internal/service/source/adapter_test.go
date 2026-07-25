package source

import (
	"context"
	"testing"

	"insta360-insight/internal/model"
)

func TestMockAdapterReturnsData(t *testing.T) {
	a := NewMockAdapter()
	if a.Name() != "mock" {
		t.Fatalf("Name() = %q, want mock", a.Name())
	}
	ctx := context.Background()
	f := model.Filter{}
	kpi, err := a.Kpi(ctx, f)
	if err != nil || len(kpi) != 5 {
		t.Fatalf("Kpi() = (%v, %v), want 5 cards", kpi, err)
	}
	if kpi[0].Label != "达人数" {
		t.Fatalf("Kpi[0].Label = %q, want 达人数", kpi[0].Label)
	}
	vt, err := a.ViewsTrend(ctx, f)
	if err != nil || len(vt) != 30 {
		t.Fatalf("ViewsTrend() len = %d, err=%v", len(vt), err)
	}
	ps, err := a.PlatformShare(ctx, f)
	if err != nil || len(ps) != 3 {
		t.Fatalf("PlatformShare() len = %d, err=%v", len(ps), err)
	}
	tp, err := a.TrackPerformance(ctx, f)
	if err != nil || len(tp) != 5 {
		t.Fatalf("TrackPerformance() len = %d, err=%v", len(tp), err)
	}
	rd, err := a.Radar(ctx, f)
	if err != nil || len(rd) != 5 {
		t.Fatalf("Radar() len = %d, err=%v", len(rd), err)
	}
	ag, err := a.AudienceAge(ctx, f)
	if err != nil || len(ag) != 4 {
		t.Fatalf("AudienceAge() len = %d, err=%v", len(ag), err)
	}
	tc, err := a.TopCreators(ctx, f)
	if err != nil || len(tc) != 10 {
		t.Fatalf("TopCreators() len = %d, err=%v", len(tc), err)
	}
	opt, err := a.Options(ctx, f)
	if err != nil || len(opt.Platforms) != 3 {
		t.Fatalf("Options() platforms = %d, err=%v", len(opt.Platforms), err)
	}
	ins, err := a.Insights(ctx, f)
	if err != nil || len(ins) != 3 {
		t.Fatalf("Insights() len = %d, err=%v", len(ins), err)
	}
}

func TestStubsReturnErrNotImplemented(t *testing.T) {
	stubs := []DataSource{&DouyinAdapter{}, &BilibiliAdapter{}, &XiaohongshuAdapter{}}
	ctx := context.Background()
	f := model.Filter{}
	for _, s := range stubs {
		if _, err := s.Kpi(ctx, f); err != ErrNotImplemented {
			t.Fatalf("%s.Kpi err = %v, want ErrNotImplemented", s.Name(), err)
		}
		if _, err := s.ViewsTrend(ctx, f); err != ErrNotImplemented {
			t.Fatalf("%s.ViewsTrend err = %v", s.Name(), err)
		}
		if _, err := s.PlatformShare(ctx, f); err != ErrNotImplemented {
			t.Fatalf("%s.PlatformShare err = %v", s.Name(), err)
		}
		if _, err := s.TrackPerformance(ctx, f); err != ErrNotImplemented {
			t.Fatalf("%s.TrackPerformance err = %v", s.Name(), err)
		}
		if _, err := s.Radar(ctx, f); err != ErrNotImplemented {
			t.Fatalf("%s.Radar err = %v", s.Name(), err)
		}
		if _, err := s.AudienceAge(ctx, f); err != ErrNotImplemented {
			t.Fatalf("%s.AudienceAge err = %v", s.Name(), err)
		}
		if _, err := s.TopCreators(ctx, f); err != ErrNotImplemented {
			t.Fatalf("%s.TopCreators err = %v", s.Name(), err)
		}
		if _, err := s.Options(ctx, f); err != ErrNotImplemented {
			t.Fatalf("%s.Options err = %v", s.Name(), err)
		}
		if _, err := s.Insights(ctx, f); err != ErrNotImplemented {
			t.Fatalf("%s.Insights err = %v", s.Name(), err)
		}
	}
}

func TestFactory(t *testing.T) {
	cases := []struct {
		kind  string
		want  string
		isErr bool
	}{
		{"", "mock", false},
		{"mock", "mock", false},
		{"douyin", "douyin", false},
		{"bilibili", "bilibili", false},
		{"xiaohongshu", "xiaohongshu", false},
		{"unknown", "", true},
	}
	for _, c := range cases {
		ds, err := NewDataSource(c.kind)
		if c.isErr {
			if err == nil {
				t.Fatalf("NewDataSource(%q) want error", c.kind)
			}
			continue
		}
		if err != nil {
			t.Fatalf("NewDataSource(%q) err = %v", c.kind, err)
		}
		if ds.Name() != c.want {
			t.Fatalf("NewDataSource(%q).Name() = %q, want %q", c.kind, ds.Name(), c.want)
		}
	}
}
