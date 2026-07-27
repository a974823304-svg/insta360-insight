package source

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"insta360-insight/internal/model"
)

// TestFallbackReturnsMockWhenNoCredential 验证:真实源无凭证(全 ErrNotImplemented)时,
// FallbackDataSource 自动回退 MockAdapter,看板永远有数据、不崩。
// 这是"没填 appkey 也能跑"的核心保障。
func TestFallbackReturnsMockWhenNoCredential(t *testing.T) {
	real := NewDouyinAdapter(PlatformConfig{}) // 无凭证 -> tp=nil
	mock := NewMockAdapter()
	fb := NewFallbackDataSource(real, mock)
	ctx := context.Background()
	f := model.Filter{}

	kpi, err := fb.Kpi(ctx, f)
	if err != nil || len(kpi) != 5 {
		t.Fatalf("fallback Kpi 应回退 mock 返回 5 卡片, got %d err %v", len(kpi), err)
	}
	if kpi[0].Label != "达人数" {
		t.Fatalf("fallback 未回退到 mock, got %q", kpi[0].Label)
	}
	vt, err := fb.ViewsTrend(ctx, f)
	if err != nil || len(vt) != 30 {
		t.Fatalf("fallback ViewsTrend 应回退 mock, got %d err %v", len(vt), err)
	}
	ps, err := fb.PlatformShare(ctx, f)
	if err != nil || len(ps) != 3 {
		t.Fatalf("fallback PlatformShare 应回退 mock, got %v", err)
	}
}

// TestFactoryFallbackWiring 验证 SOURCE=douyin 时返回带兜底的 DataSource。
func TestFactoryFallbackWiring(t *testing.T) {
	ds, err := NewDataSource("douyin")
	if err != nil {
		t.Fatalf("NewDataSource(douyin) err = %v", err)
	}
	if ds.Name() != "douyin" {
		t.Fatalf("Name() = %q, want douyin", ds.Name())
	}
	// 无凭证环境,所有方法应回退 mock(有数据、err=nil)
	kpi, err := ds.Kpi(context.Background(), model.Filter{})
	if err != nil || len(kpi) == 0 {
		t.Fatalf("factory fallback Kpi 应返回 mock 数据, got %d err %v", len(kpi), err)
	}
}

// TestDouyinKpiMapping 用 httptest 验证抖音经营数据接口的字段映射正确。
func TestDouyinKpiMapping(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"data":{"error_code":0,"fans_total":12345678,"play_total":200000000,"digg_total":1000000,"comment_total":50000,"share_total":20000,"collab_total":30}}`)
	}))
	defer srv.Close()

	a := NewDouyinAdapter(PlatformConfig{AccessToken: "test-token", BaseURL: srv.URL})
	a.httpDo = srv.Client().Do

	kpi, err := a.Kpi(context.Background(), model.Filter{})
	if err != nil {
		t.Fatalf("Kpi err = %v", err)
	}
	if len(kpi) != 5 {
		t.Fatalf("want 5 cards, got %d", len(kpi))
	}
	if kpi[1].Key != "followers" || kpi[1].Raw != 12345678 {
		t.Fatalf("followers 映射错误: %+v", kpi[1])
	}
	if kpi[2].Key != "views" || kpi[2].Raw != 200000000 {
		t.Fatalf("views 映射错误: %+v", kpi[2])
	}
}

// TestBilibiliSignAndMapping 验证 B站 v2 签名头被写入且字段映射正确。
func TestBilibiliSignAndMapping(t *testing.T) {
	var gotVer, gotAuth, gotAK string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotVer = r.Header.Get("X-Bili-Signature-Version")
		gotAuth = r.Header.Get("Authorization")
		gotAK = r.Header.Get("X-Bili-Accesskeyid")
		fmt.Fprint(w, `{"code":0,"data":{"current_fans":8888,"current_archive_view":99999,"inc_like":200,"inc_reply":40,"inc_share":70}}`)
	}))
	defer srv.Close()

	a := NewBilibiliAdapter(PlatformConfig{
		AccessToken: "tok", ClientKey: "ak", ClientSecret: "sk", BaseURL: srv.URL,
	})
	a.httpDo = srv.Client().Do

	kpi, err := a.Kpi(context.Background(), model.Filter{})
	if err != nil {
		t.Fatalf("Kpi err = %v", err)
	}
	if gotVer != "2.0" {
		t.Fatalf("签名版本头缺失, got %q", gotVer)
	}
	if gotAuth == "" {
		t.Fatalf("Authorization 签名缺失")
	}
	if gotAK != "ak" {
		t.Fatalf("Accesskeyid 头缺失, got %q", gotAK)
	}
	if len(kpi) != 5 {
		t.Fatalf("want 5 cards, got %d", len(kpi))
	}
	if kpi[1].Raw != 8888 {
		t.Fatalf("followers 映射错误: %+v", kpi[1])
	}
}

// TestXiaohongshuKpiMapping 验证小红书蒲公英核心指标字段映射正确(聚合/官方通用)。
func TestXiaohongshuKpiMapping(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"code":0,"data":{"fans":5555555,"notes":300,"avg_play":8000,"avg_like":400,"avg_comment":50,"avg_collect":120,"avg_share":30}}`)
	}))
	defer srv.Close()

	a := NewXiaohongshuAdapter(PlatformConfig{AccessToken: "apikey", UserID: "u123", BaseURL: srv.URL})
	a.httpDo = srv.Client().Do

	kpi, err := a.Kpi(context.Background(), model.Filter{})
	if err != nil {
		t.Fatalf("Kpi err = %v", err)
	}
	if len(kpi) != 5 {
		t.Fatalf("want 5 cards, got %d", len(kpi))
	}
	if kpi[1].Key != "followers" || kpi[1].Raw != 5555555 {
		t.Fatalf("followers 映射错误: %+v", kpi[1])
	}
}
