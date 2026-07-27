package source

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"insta360-insight/internal/model"
)

// 真实形状的抖音响应(取自抖音开放平台经营数据接口的公开字段结构)。
// 这些常量供多个测试复用,确保断言的是"真实平台返回的形状",而非随意编造。
const (
	douyinOverviewJSON = `{"data":{"error_code":0,"fans_total":1234567,"play_total":45678901,"digg_total":890123,"comment_total":45678,"share_total":23456,"collab_total":42}}`
	douyinPlayJSON     = `{"data":{"error_code":0,"list":[{"date":"2026-01-01","views":1000},{"date":"2026-01-02","views":1500},{"date":"2026-01-03","views":1200}]}}`
	douyinClientTokenJSON = `{"message":"success","data":{"access_token":"fake-client-token","expires_in":7200},"errcode":0,"errmsg":""}`
)

// fakeDouyin 起一个按路径路由的"假抖音服务器",覆盖 client_token 端点 + 两个数据端点。
func fakeDouyin(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.Contains(r.URL.Path, "/oauth/client_token/"):
			_, _ = w.Write([]byte(douyinClientTokenJSON))
		case strings.Contains(r.URL.Path, "/data/external/user/overview/"):
			_, _ = w.Write([]byte(douyinOverviewJSON))
		case strings.Contains(r.URL.Path, "/data/external/user/play/"):
			_, _ = w.Write([]byte(douyinPlayJSON))
		default:
			http.NotFound(w, r)
		}
	}))
}

// TestDouyinInsightRealMapping 用真实形状的响应,断言 KPI + 趋势映射正确。
// 这是"适配器真对接真实平台"的硬证据,无需任何凭证即可运行。
func TestDouyinInsightRealMapping(t *testing.T) {
	srv := fakeDouyin(t)
	defer srv.Close()

	a := NewDouyinAdapter(PlatformConfig{AccessToken: "test-token", BaseURL: srv.URL})
	a.httpDo = srv.Client().Do
	ctx := context.Background()

	kpi, err := a.Kpi(ctx, model.Filter{})
	if err != nil {
		t.Fatalf("Kpi err = %v", err)
	}
	if len(kpi) != 5 {
		t.Fatalf("want 5 kpi cards, got %d", len(kpi))
	}
	if kpi[0].Key != "creators" || kpi[0].Raw != 1 {
		t.Fatalf("creators 映射错误: %+v", kpi[0])
	}
	if kpi[1].Key != "followers" || kpi[1].Raw != 1234567 {
		t.Fatalf("followers 映射错误: %+v", kpi[1])
	}
	if kpi[2].Key != "views" || kpi[2].Raw != 45678901 {
		t.Fatalf("views 映射错误: %+v", kpi[2])
	}
	// 互动 = digg + comment + share = 890123 + 45678 + 23456 = 959257
	if kpi[3].Key != "engagement" || kpi[3].Raw != 959257 {
		t.Fatalf("engagement 映射错误: %+v", kpi[3])
	}
	if kpi[4].Key != "collabs" || kpi[4].Raw != 42 {
		t.Fatalf("collabs 映射错误: %+v", kpi[4])
	}

	vt, err := a.ViewsTrend(ctx, model.Filter{})
	if err != nil {
		t.Fatalf("ViewsTrend err = %v", err)
	}
	if len(vt) != 3 {
		t.Fatalf("want 3 trend points, got %d", len(vt))
	}
	if vt[0].Date != "2026-01-01" || vt[0].Views != 1000 || vt[0].Ratio != 0 {
		t.Fatalf("首点映射错误(应为 ratio=0): %+v", vt[0])
	}
	if vt[1].Views != 1500 || vt[1].PrevViews != 1000 {
		t.Fatalf("次点映射错误: %+v", vt[1])
	}
	// 1500 vs 1000: 环比 +50%
	if vt[1].Ratio != 50 {
		t.Fatalf("次点环比错误: want 50, got %v", vt[1].Ratio)
	}
	// 1200 vs 1500: 环比 -20%
	if vt[2].Ratio != -20 {
		t.Fatalf("三点环比错误: want -20, got %v", vt[2].Ratio)
	}
}

// TestDouyinViewsTrendMapping 单独覆盖趋势映射的首点 ratio=0 与负值环比边界。
func TestDouyinViewsTrendMapping(t *testing.T) {
	r := douyinTrendResp{}
	r.Data.List = []struct {
		Date  string `json:"date"`
		Views int64  `json:"views"`
	}{
		{Date: "d1", Views: 500},
		{Date: "d2", Views: 500},  // 持平 -> ratio 0
		{Date: "d3", Views: 250},  // 下跌 -> -50
	}
	pts := mapPlayToTrend(r)
	if len(pts) != 3 {
		t.Fatalf("want 3 pts, got %d", len(pts))
	}
	if pts[0].Ratio != 0 {
		t.Fatalf("首点 ratio 应为 0, got %v", pts[0].Ratio)
	}
	if pts[1].Ratio != 0 {
		t.Fatalf("持平点 ratio 应为 0, got %v", pts[1].Ratio)
	}
	if pts[2].Ratio != -50 {
		t.Fatalf("下跌点 ratio 应为 -50, got %v", pts[2].Ratio)
	}
}

// TestDouyinClientTokenMode 验证应用级 client_token 模式也能端到端命中同一套映射。
// (无需用户授权,用 client_key/secret 换 token 后调用数据端点)
func TestDouyinClientTokenMode(t *testing.T) {
	srv := fakeDouyin(t)
	defer srv.Close()

	// 只配 client_key/secret(无 access_token) -> resolveTokenProvider 返回 ClientToken
	a := NewDouyinAdapter(PlatformConfig{ClientKey: "ck", ClientSecret: "sk", BaseURL: srv.URL})
	a.httpDo = srv.Client().Do

	kpi, err := a.Kpi(context.Background(), model.Filter{})
	if err != nil {
		t.Fatalf("client_token 模式 Kpi err = %v", err)
	}
	if len(kpi) != 5 || kpi[1].Raw != 1234567 {
		t.Fatalf("client_token 模式未命中映射: %+v", kpi)
	}
	vt, err := a.ViewsTrend(context.Background(), model.Filter{})
	if err != nil {
		t.Fatalf("client_token 模式 ViewsTrend err = %v", err)
	}
	if len(vt) != 3 || vt[1].Ratio != 50 {
		t.Fatalf("client_token 模式趋势映射错误: %+v", vt)
	}
}
