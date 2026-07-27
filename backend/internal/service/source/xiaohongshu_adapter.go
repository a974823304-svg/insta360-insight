package source

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"

	"insta360-insight/internal/model"
)

// XiaohongshuAdapter 小红书数据源。
//
// 官方蒲公英平台 API 门槛极高(品牌年消耗 500W+ 或白名单),普通开发者基本拿不到。
// 因此本期同时支持两种真实来源:
//  1. 蒲公英官方(ark.xiaohongshu.com):需商家授权 access_token,经授权码流程获取后填 XHS_ACCESS_TOKEN
//  2. 第三方聚合(如 Just One API / rnote):把 API key 填 XHS_ACCESS_TOKEN,并把
//     XHS_BASE_URL 指向聚合网关(如 https://api.justoneapi.com)
// 无凭证或接口未授权时,方法返回 ErrNotImplemented,由 FallbackDataSource 回退 MockAdapter。
//
// 下方映射基于蒲公英"创作者核心指标"接口的合理响应结构,字段名需按平台/聚合最新文档核对。
type XiaohongshuAdapter struct {
	cfg    PlatformConfig
	tp     TokenProvider
	httpDo func(*http.Request) (*http.Response, error)
}

func NewXiaohongshuAdapter(cfg PlatformConfig) *XiaohongshuAdapter {
	return &XiaohongshuAdapter{
		cfg:    cfg,
		tp:     resolveTokenProvider(cfg, false), // 蒲公英需商家授权 token / 聚合 API key
		httpDo: http.DefaultClient.Do,
	}
}

func (a *XiaohongshuAdapter) Name() string { return "xiaohongshu" }

// ---- 平台响应结构(示意为蒲公英创作者核心指标,字段需按最新文档核对) ----

type xhsCoreDataResp struct {
	Code int `json:"code"`
	Data struct {
		Fans       int64 `json:"fans"`
		Notes      int64 `json:"notes"`
		AvgPlay    int64 `json:"avg_play"`
		AvgLike    int64 `json:"avg_like"`
		AvgComment int64 `json:"avg_comment"`
		AvgCollect int64 `json:"avg_collect"`
		AvgShare   int64 `json:"avg_share"`
	} `json:"data"`
}

// xhsTrendResp 笔记/播放按天趋势(示意)。
type xhsTrendResp struct {
	Code int `json:"code"`
	Data struct {
		List []struct {
			Date  string `json:"date"`
			Views int64  `json:"views"`
		} `json:"list"`
	} `json:"data"`
}

func (a *XiaohongshuAdapter) requireToken(ctx context.Context) (string, error) {
	if a.tp == nil {
		return "", ErrNotImplemented
	}
	token, err := a.tp.Token(ctx)
	if err != nil {
		return "", ErrNotImplemented
	}
	return token, nil
}

// get 聚合/官方通用 GET:token 作为 query 参数传递(聚合),官方蒲公英则需放 Authorization/body。
func (a *XiaohongshuAdapter) get(ctx context.Context, path string, q map[string]string) ([]byte, error) {
	token, err := a.requireToken(ctx)
	if err != nil {
		return nil, err
	}
	u := strings.TrimRight(a.cfg.BaseURL, "/") + path
	params := url.Values{}
	params.Set("token", token)
	for k, v := range q {
		params.Set(k, v)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u+"?"+params.Encode(), nil)
	if err != nil {
		return nil, err
	}
	resp, err := a.httpDo(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

// ---- 洞察域(真实映射范例) ----

func (a *XiaohongshuAdapter) Kpi(ctx context.Context, _ model.Filter) ([]model.KpiCard, error) {
	if a.tp == nil {
		return nil, ErrNotImplemented
	}
	q := map[string]string{}
	if a.cfg.UserID != "" {
		q["userId"] = a.cfg.UserID
	}
	body, err := a.get(ctx, "/api/xiaohongshu-pgy/api/pgy/kol/data/core_data/v1", q)
	if err != nil {
		return nil, ErrNotImplemented
	}
	var r xhsCoreDataResp
	if err := json.Unmarshal(body, &r); err != nil || r.Code != 0 {
		return nil, ErrNotImplemented
	}
	eng := r.Data.AvgLike + r.Data.AvgComment + r.Data.AvgCollect + r.Data.AvgShare
	return []model.KpiCard{
		{Key: "creators", Label: "达人账号数", Value: "1", Raw: 1, Description: "当前授权达人"},
		{Key: "followers", Label: "粉丝总量", Value: humanize(r.Data.Fans), Raw: float64(r.Data.Fans), Unit: unitOf(r.Data.Fans), Description: "小红书平台"},
		{Key: "views", Label: "平均播放", Value: humanize(r.Data.AvgPlay), Raw: float64(r.Data.AvgPlay), Unit: unitOf(r.Data.AvgPlay), Description: "单篇均值"},
		{Key: "engagement", Label: "平均互动", Value: humanize(eng), Raw: float64(eng), Description: "赞+评+藏+转"},
		{Key: "collabs", Label: "笔记数", Value: humanize(r.Data.Notes), Raw: float64(r.Data.Notes), Description: "累计笔记"},
	}, nil
}

func (a *XiaohongshuAdapter) ViewsTrend(ctx context.Context, _ model.Filter) ([]model.ViewsTrendPoint, error) {
	if a.tp == nil {
		return nil, ErrNotImplemented
	}
	q := map[string]string{}
	if a.cfg.UserID != "" {
		q["userId"] = a.cfg.UserID
	}
	body, err := a.get(ctx, "/api/xiaohongshu-pgy/api/pgy/kol/data/trend/v1", q)
	if err != nil {
		return nil, ErrNotImplemented
	}
	var r xhsTrendResp
	if err := json.Unmarshal(body, &r); err != nil || r.Code != 0 {
		return nil, ErrNotImplemented
	}
	pts := make([]model.ViewsTrendPoint, 0, len(r.Data.List))
	for i, it := range r.Data.List {
		prev := int64(0)
		if i > 0 {
			prev = r.Data.List[i-1].Views
		}
		ratio := 0.0
		if prev > 0 {
			ratio = float64(it.Views-prev) / float64(prev) * 100
		}
		pts = append(pts, model.ViewsTrendPoint{Date: it.Date, Views: it.Views, PrevViews: prev, Ratio: ratio})
	}
	return pts, nil
}

func (a *XiaohongshuAdapter) PlatformShare(ctx context.Context, _ model.Filter) ([]model.PlatformShare, error) {
	if a.tp == nil {
		return nil, ErrNotImplemented
	}
	return []model.PlatformShare{{Platform: "小红书", Share: 100, Views: 0, Color: "#FF2442"}}, nil
}

func (a *XiaohongshuAdapter) TrackPerformance(ctx context.Context, _ model.Filter) ([]model.TrackPerformance, error) {
	return nil, ErrNotImplemented
}
func (a *XiaohongshuAdapter) Radar(ctx context.Context, _ model.Filter) ([]model.RadarMetric, error) {
	return nil, ErrNotImplemented
}
func (a *XiaohongshuAdapter) AudienceAge(ctx context.Context, _ model.Filter) ([]model.AgeShare, error) {
	return nil, ErrNotImplemented
}
func (a *XiaohongshuAdapter) TopCreators(ctx context.Context, _ model.Filter) ([]model.TopCreator, error) {
	return nil, ErrNotImplemented
}
func (a *XiaohongshuAdapter) Options(ctx context.Context, _ model.Filter) (model.FilterOptions, error) {
	return model.FilterOptions{}, ErrNotImplemented
}
func (a *XiaohongshuAdapter) Insights(ctx context.Context, _ model.Filter) ([]model.Insight, error) {
	return nil, ErrNotImplemented
}

// ---- 达人分析域(回退) ----

func (a *XiaohongshuAdapter) CreatorKpi(ctx context.Context, f model.Filter) ([]model.KpiCard, error) {
	return nil, ErrNotImplemented
}
func (a *XiaohongshuAdapter) CreatorTrend(ctx context.Context, f model.Filter) ([]model.ViewsTrendPoint, error) {
	return nil, ErrNotImplemented
}
func (a *XiaohongshuAdapter) CreatorPlatforms(ctx context.Context, f model.Filter) ([]model.PlatformShare, error) {
	return nil, ErrNotImplemented
}
func (a *XiaohongshuAdapter) CreatorTracks(ctx context.Context, f model.Filter) ([]model.TrackPerformance, error) {
	return nil, ErrNotImplemented
}
func (a *XiaohongshuAdapter) CreatorAudience(ctx context.Context, f model.Filter) (*model.Audience, error) {
	return nil, ErrNotImplemented
}
func (a *XiaohongshuAdapter) CreatorList(ctx context.Context, f model.Filter) ([]model.TopCreator, error) {
	return nil, ErrNotImplemented
}

// ---- 内容分析域(回退) ----

func (a *XiaohongshuAdapter) ContentKpi(ctx context.Context, f model.Filter) ([]model.KpiCard, error) {
	return nil, ErrNotImplemented
}
func (a *XiaohongshuAdapter) ContentTrend(ctx context.Context, f model.Filter) ([]model.ViewsTrendPoint, error) {
	return nil, ErrNotImplemented
}
func (a *XiaohongshuAdapter) ContentForms(ctx context.Context, f model.Filter) ([]model.PlatformShare, error) {
	return nil, ErrNotImplemented
}
func (a *XiaohongshuAdapter) ContentTopics(ctx context.Context, f model.Filter) ([]model.TrackPerformance, error) {
	return nil, ErrNotImplemented
}
func (a *XiaohongshuAdapter) ContentDurations(ctx context.Context, f model.Filter) ([]model.AgeShare, error) {
	return nil, ErrNotImplemented
}
func (a *XiaohongshuAdapter) ContentList(ctx context.Context, f model.Filter) ([]model.ContentItem, error) {
	return nil, ErrNotImplemented
}

// ---- 市场洞察域(回退) ----

func (a *XiaohongshuAdapter) MarketKpi(ctx context.Context, f model.Filter) ([]model.KpiCard, error) {
	return nil, ErrNotImplemented
}
func (a *XiaohongshuAdapter) MarketTrend(ctx context.Context, f model.Filter) ([]model.ViewsTrendPoint, error) {
	return nil, ErrNotImplemented
}
func (a *XiaohongshuAdapter) MarketCompetitors(ctx context.Context, f model.Filter) ([]model.PlatformShare, error) {
	return nil, ErrNotImplemented
}
func (a *XiaohongshuAdapter) MarketRegions(ctx context.Context, f model.Filter) ([]model.TrackPerformance, error) {
	return nil, ErrNotImplemented
}
func (a *XiaohongshuAdapter) MarketPrices(ctx context.Context, f model.Filter) ([]model.AgeShare, error) {
	return nil, ErrNotImplemented
}
func (a *XiaohongshuAdapter) MarketList(ctx context.Context, f model.Filter) ([]model.Competitor, error) {
	return nil, ErrNotImplemented
}

// ---- 品牌分析域(回退) ----

func (a *XiaohongshuAdapter) BrandKpi(ctx context.Context, f model.Filter) ([]model.KpiCard, error) {
	return nil, ErrNotImplemented
}
func (a *XiaohongshuAdapter) BrandTrend(ctx context.Context, f model.Filter) ([]model.ViewsTrendPoint, error) {
	return nil, ErrNotImplemented
}
func (a *XiaohongshuAdapter) BrandPlatforms(ctx context.Context, f model.Filter) ([]model.PlatformShare, error) {
	return nil, ErrNotImplemented
}
func (a *XiaohongshuAdapter) BrandSentiment(ctx context.Context, f model.Filter) ([]model.AgeShare, error) {
	return nil, ErrNotImplemented
}
func (a *XiaohongshuAdapter) BrandKeywords(ctx context.Context, f model.Filter) ([]model.TagItem, error) {
	return nil, ErrNotImplemented
}
func (a *XiaohongshuAdapter) BrandList(ctx context.Context, f model.Filter) ([]model.PartnerBrand, error) {
	return nil, ErrNotImplemented
}
