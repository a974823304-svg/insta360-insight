package source

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"insta360-insight/internal/model"
)

// DouyinAdapter 抖音数据源。
//
// 真实接入依赖抖音开放平台 / 巨量星图经营数据接口(需企业资质 + OAuth2 token)。
// 无凭证或接口未授权时,方法返回 ErrNotImplemented,由 FallbackDataSource 回退 MockAdapter。
//
// 重要:抖音开放平台公开接口仅能取授权账号的公开资料 / 视频;达人营销汇总数据
// (粉丝 / 播放 / 互动 / 榜单)在巨量星图(企业资质)。下方映射基于星图经营数据接口的
// 合理响应结构,字段名需按平台最新文档核对。有企业 appkey + scope 后即可真实返回。
type DouyinAdapter struct {
	cfg    PlatformConfig
	tp     TokenProvider
	httpDo func(*http.Request) (*http.Response, error)
}

func NewDouyinAdapter(cfg PlatformConfig) *DouyinAdapter {
	return &DouyinAdapter{
		cfg:    cfg,
		tp:     resolveTokenProvider(cfg, true), // 抖音支持 client_token(无需用户授权)
		httpDo: http.DefaultClient.Do,
	}
}

func (a *DouyinAdapter) Name() string { return "douyin" }

// ---- 平台响应结构(示意为星图经营数据,字段需按最新文档核对) ----

type douyinOverviewResp struct {
	Data struct {
		ErrorCode    int    `json:"error_code"`
		Description  string `json:"description"`
		FansTotal    int64  `json:"fans_total"`
		PlayTotal    int64  `json:"play_total"`
		DiggTotal    int64  `json:"digg_total"`
		CommentTotal int64  `json:"comment_total"`
		ShareTotal   int64  `json:"share_total"`
		CollabTotal  int64  `json:"collab_total"`
	} `json:"data"`
}

type douyinTrendResp struct {
	Data struct {
		ErrorCode int `json:"error_code"`
		List      []struct {
			Date  string `json:"date"`
			Views int64  `json:"views"`
		} `json:"list"`
	} `json:"data"`
}

type douyinUserinfoResp struct {
	Data struct {
		ErrorCode int    `json:"error_code"`
		Nickname  string `json:"nickname"`
		Avatar    string `json:"avatar"`
		Gender    string `json:"gender"`
		City      string `json:"city"`
		Province  string `json:"province"`
	} `json:"data"`
}

// ---- HTTP 辅助 ----

func (a *DouyinAdapter) requireToken(ctx context.Context) (string, error) {
	if a.tp == nil {
		return "", ErrNotImplemented
	}
	token, err := a.tp.Token(ctx)
	if err != nil {
		return "", ErrNotImplemented
	}
	return token, nil
}

func (a *DouyinAdapter) get(ctx context.Context, token, path string, q map[string]string) ([]byte, error) {
	u := strings.TrimRight(a.cfg.BaseURL, "/") + path
	if len(q) > 0 {
		params := make([]string, 0, len(q))
		for k, v := range q {
			params = append(params, k+"="+v)
		}
		u += "?" + strings.Join(params, "&")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := a.httpDo(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

// ---- 洞察域(真实映射范例) ----

// mapOverviewToKpi 把抖音经营总览响应映射成首页 KPI 卡。
// 用户授权 access_token 与 client_token(应用级)两种模式命中同一映射,差异仅在 token 来源。
func mapOverviewToKpi(r douyinOverviewResp) []model.KpiCard {
	eng := r.Data.DiggTotal + r.Data.CommentTotal + r.Data.ShareTotal
	return []model.KpiCard{
		{Key: "creators", Label: "授权账号数", Value: "1", Raw: 1, Description: "当前授权抖音账号"},
		{Key: "followers", Label: "粉丝总量", Value: humanize(r.Data.FansTotal), Raw: float64(r.Data.FansTotal), Unit: unitOf(r.Data.FansTotal), Description: "抖音平台"},
		{Key: "views", Label: "播放总量", Value: humanize(r.Data.PlayTotal), Raw: float64(r.Data.PlayTotal), Unit: unitOf(r.Data.PlayTotal), Description: "累计播放"},
		{Key: "engagement", Label: "互动总量", Value: humanize(eng), Raw: float64(eng), Description: "赞+评+转"},
		{Key: "collabs", Label: "商业合作数", Value: fmt.Sprintf("%d", r.Data.CollabTotal), Raw: float64(r.Data.CollabTotal), Description: "星图合作"},
	}
}

// mapPlayToTrend 把抖音播放趋势响应映射成趋势点(含环比 ratio)。首点无前值,ratio=0。
func mapPlayToTrend(r douyinTrendResp) []model.ViewsTrendPoint {
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
	return pts
}

// Kpi 真实映射(洞察域)。端点 /data/external/user/overview/ 返回授权账号的经营总览。
// 注意: 个人号 live 拉到的是「你自己账号」的真实数据; client_token 覆盖的是公开发现类接口,
// 端点与字段以抖音开放平台最新文档为准(本适配器注释已标注),拿凭证后首要核对。
func (a *DouyinAdapter) Kpi(ctx context.Context, _ model.Filter) ([]model.KpiCard, error) {
	token, err := a.requireToken(ctx)
	if err != nil {
		return nil, err
	}
	body, err := a.get(ctx, token, "/data/external/user/overview/", nil)
	if err != nil {
		return nil, ErrNotImplemented
	}
	var r douyinOverviewResp
	if err := json.Unmarshal(body, &r); err != nil || r.Data.ErrorCode != 0 {
		return nil, ErrNotImplemented
	}
	return mapOverviewToKpi(r), nil
}

// ViewsTrend 真实映射(洞察域)。端点 /data/external/user/play/ 返回授权账号的播放趋势。
func (a *DouyinAdapter) ViewsTrend(ctx context.Context, _ model.Filter) ([]model.ViewsTrendPoint, error) {
	token, err := a.requireToken(ctx)
	if err != nil {
		return nil, err
	}
	body, err := a.get(ctx, token, "/data/external/user/play/", nil)
	if err != nil {
		return nil, ErrNotImplemented
	}
	var r douyinTrendResp
	if err := json.Unmarshal(body, &r); err != nil || r.Data.ErrorCode != 0 {
		return nil, ErrNotImplemented
	}
	return mapPlayToTrend(r), nil
}

func (a *DouyinAdapter) PlatformShare(ctx context.Context, _ model.Filter) ([]model.PlatformShare, error) {
	if a.tp == nil {
		return nil, ErrNotImplemented
	}
	return []model.PlatformShare{{Platform: "抖音", Share: 100, Views: 0, Color: "#FE2C55"}}, nil
}

func (a *DouyinAdapter) TrackPerformance(ctx context.Context, _ model.Filter) ([]model.TrackPerformance, error) {
	return nil, ErrNotImplemented
}

func (a *DouyinAdapter) Radar(ctx context.Context, _ model.Filter) ([]model.RadarMetric, error) {
	return nil, ErrNotImplemented
}

func (a *DouyinAdapter) AudienceAge(ctx context.Context, _ model.Filter) ([]model.AgeShare, error) {
	return nil, ErrNotImplemented
}

func (a *DouyinAdapter) TopCreators(ctx context.Context, _ model.Filter) ([]model.TopCreator, error) {
	token, err := a.requireToken(ctx)
	if err != nil {
		return nil, err
	}
	// 单授权账号视角:用经营总览填充(星图接口)
	body, err := a.get(ctx, token, "/data/external/user/overview/", nil)
	if err != nil {
		return nil, ErrNotImplemented
	}
	var r douyinOverviewResp
	if err := json.Unmarshal(body, &r); err != nil || r.Data.ErrorCode != 0 {
		return nil, ErrNotImplemented
	}
	ui := douyinUserinfoResp{}
	if b, e := a.get(ctx, token, "/oauth/userinfo/", nil); e == nil {
		_ = json.Unmarshal(b, &ui)
	}
	return []model.TopCreator{{
		Rank:       1,
		Name:       ui.Data.Nickname,
		Platform:   "抖音",
		Avatar:     ui.Data.Avatar,
		Followers:  r.Data.FansTotal,
		TotalViews: r.Data.PlayTotal,
	}}, nil
}

func (a *DouyinAdapter) Options(ctx context.Context, _ model.Filter) (model.FilterOptions, error) {
	return model.FilterOptions{}, ErrNotImplemented
}

func (a *DouyinAdapter) Insights(ctx context.Context, _ model.Filter) ([]model.Insight, error) {
	return nil, ErrNotImplemented
}

// ---- 达人分析域(抖音公开接口无法提供多账号/跨账号数据,回退) ----

func (a *DouyinAdapter) CreatorKpi(ctx context.Context, f model.Filter) ([]model.KpiCard, error) {
	return nil, ErrNotImplemented
}
func (a *DouyinAdapter) CreatorTrend(ctx context.Context, f model.Filter) ([]model.ViewsTrendPoint, error) {
	return nil, ErrNotImplemented
}
func (a *DouyinAdapter) CreatorPlatforms(ctx context.Context, f model.Filter) ([]model.PlatformShare, error) {
	return nil, ErrNotImplemented
}
func (a *DouyinAdapter) CreatorTracks(ctx context.Context, f model.Filter) ([]model.TrackPerformance, error) {
	return nil, ErrNotImplemented
}
func (a *DouyinAdapter) CreatorAudience(ctx context.Context, f model.Filter) (*model.Audience, error) {
	return nil, ErrNotImplemented
}
func (a *DouyinAdapter) CreatorList(ctx context.Context, f model.Filter) ([]model.TopCreator, error) {
	return nil, ErrNotImplemented
}

// ---- 内容分析域(回退) ----

func (a *DouyinAdapter) ContentKpi(ctx context.Context, f model.Filter) ([]model.KpiCard, error) {
	return nil, ErrNotImplemented
}
func (a *DouyinAdapter) ContentTrend(ctx context.Context, f model.Filter) ([]model.ViewsTrendPoint, error) {
	return nil, ErrNotImplemented
}
func (a *DouyinAdapter) ContentForms(ctx context.Context, f model.Filter) ([]model.PlatformShare, error) {
	return nil, ErrNotImplemented
}
func (a *DouyinAdapter) ContentTopics(ctx context.Context, f model.Filter) ([]model.TrackPerformance, error) {
	return nil, ErrNotImplemented
}
func (a *DouyinAdapter) ContentDurations(ctx context.Context, f model.Filter) ([]model.AgeShare, error) {
	return nil, ErrNotImplemented
}
func (a *DouyinAdapter) ContentList(ctx context.Context, f model.Filter) ([]model.ContentItem, error) {
	return nil, ErrNotImplemented
}

// ---- 市场洞察域(回退) ----

func (a *DouyinAdapter) MarketKpi(ctx context.Context, f model.Filter) ([]model.KpiCard, error) {
	return nil, ErrNotImplemented
}
func (a *DouyinAdapter) MarketTrend(ctx context.Context, f model.Filter) ([]model.ViewsTrendPoint, error) {
	return nil, ErrNotImplemented
}
func (a *DouyinAdapter) MarketCompetitors(ctx context.Context, f model.Filter) ([]model.PlatformShare, error) {
	return nil, ErrNotImplemented
}
func (a *DouyinAdapter) MarketRegions(ctx context.Context, f model.Filter) ([]model.TrackPerformance, error) {
	return nil, ErrNotImplemented
}
func (a *DouyinAdapter) MarketPrices(ctx context.Context, f model.Filter) ([]model.AgeShare, error) {
	return nil, ErrNotImplemented
}
func (a *DouyinAdapter) MarketList(ctx context.Context, f model.Filter) ([]model.Competitor, error) {
	return nil, ErrNotImplemented
}

// ---- 品牌分析域(回退) ----

func (a *DouyinAdapter) BrandKpi(ctx context.Context, f model.Filter) ([]model.KpiCard, error) {
	return nil, ErrNotImplemented
}
func (a *DouyinAdapter) BrandTrend(ctx context.Context, f model.Filter) ([]model.ViewsTrendPoint, error) {
	return nil, ErrNotImplemented
}
func (a *DouyinAdapter) BrandPlatforms(ctx context.Context, f model.Filter) ([]model.PlatformShare, error) {
	return nil, ErrNotImplemented
}
func (a *DouyinAdapter) BrandSentiment(ctx context.Context, f model.Filter) ([]model.AgeShare, error) {
	return nil, ErrNotImplemented
}
func (a *DouyinAdapter) BrandKeywords(ctx context.Context, f model.Filter) ([]model.TagItem, error) {
	return nil, ErrNotImplemented
}
func (a *DouyinAdapter) BrandList(ctx context.Context, f model.Filter) ([]model.PartnerBrand, error) {
	return nil, ErrNotImplemented
}
