package source

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"insta360-insight/internal/model"
)

// BilibiliAdapter B站数据源。
//
// 真实接入依赖哔哩哔哩开放平台创作者数据接口(需用户授权 access_token + HMAC-SHA256 v2 签名)。
// 无凭证或接口未授权时,方法返回 ErrNotImplemented,由 FallbackDataSource 回退 MockAdapter。
//
// 说明:B站没有"无需用户授权的应用级 token",必须先用授权码流程拿到用户 access_token 再填到
// BILIBILI_ACCESS_TOKEN。只配 client_id/secret 不足以拉数(会回退)。
type BilibiliAdapter struct {
	cfg    PlatformConfig
	tp     TokenProvider
	httpDo func(*http.Request) (*http.Response, error)
}

func NewBilibiliAdapter(cfg PlatformConfig) *BilibiliAdapter {
	return &BilibiliAdapter{
		cfg:    cfg,
		tp:     resolveTokenProvider(cfg, false), // B站无应用级 client_token
		httpDo: http.DefaultClient.Do,
	}
}

func (a *BilibiliAdapter) Name() string { return "bilibili" }

// ---- 平台响应结构 ----

type biliIncStatsResp struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		IncArchiveView     int64 `json:"inc_archive_view"`
		IncFans            int64 `json:"inc_fans"`
		IncLike            int64 `json:"inc_like"`
		IncDanmaku         int64 `json:"inc_danmaku"`
		IncReply           int64 `json:"inc_reply"`
		IncFavorite        int64 `json:"inc_favorite"`
		IncCoin            int64 `json:"inc_coin"`
		IncShare           int64 `json:"inc_share"`
		CurrentArchiveView int64 `json:"current_archive_view"`
		CurrentFans        int64 `json:"current_fans"`
	} `json:"data"`
}

// biliArcDataResp 稿件按天数据(示意,需按最新文档核对字段)。
type biliArcDataResp struct {
	Code int `json:"code"`
	Data struct {
		Result []struct {
			Date string `json:"date"`
			View int64  `json:"view"`
		} `json:"result"`
	} `json:"data"`
}

func (a *BilibiliAdapter) requireToken(ctx context.Context) (string, error) {
	if a.tp == nil {
		return "", ErrNotImplemented
	}
	token, err := a.tp.Token(ctx)
	if err != nil {
		return "", ErrNotImplemented
	}
	return token, nil
}

// sign 按 B站开放平台 v2 签名算法对请求头部签名(HMAC-SHA256)。
// 文档:https://openhome.bilibili.com 接口签名说明。
func (a *BilibiliAdapter) sign(req *http.Request) {
	nonce := fmt.Sprintf("%d", time.Now().UnixNano())
	ts := fmt.Sprintf("%d", time.Now().Unix())
	contentMD5 := md5Base64("")
	req.Header.Set("X-Bili-Accesskeyid", a.cfg.ClientKey)
	req.Header.Set("X-Bili-Content-Md5", contentMD5)
	req.Header.Set("X-Bili-Signature-Method", "HMAC-SHA256")
	req.Header.Set("X-Bili-Signature-Nonce", nonce)
	req.Header.Set("X-Bili-Signature-Version", "2.0")
	req.Header.Set("X-Bili-Timestamp", ts)
	keys := []string{
		"X-Bili-Accesskeyid", "X-Bili-Content-Md5", "X-Bili-Signature-Method",
		"X-Bili-Signature-Nonce", "X-Bili-Signature-Version", "X-Bili-Timestamp",
	}
	var sb strings.Builder
	for i, k := range keys {
		if i > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString(k + ":" + req.Header.Get(k))
	}
	mac := hmac.New(sha256.New, []byte(a.cfg.ClientSecret))
	mac.Write([]byte(sb.String()))
	req.Header.Set("Authorization", hex.EncodeToString(mac.Sum(nil)))
}

func (a *BilibiliAdapter) get(ctx context.Context, token, path string) ([]byte, error) {
	u := strings.TrimRight(a.cfg.BaseURL, "/") + path
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Access-Token", token)
	a.sign(req)
	resp, err := a.httpDo(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

// ---- 洞察域(真实映射范例) ----

func (a *BilibiliAdapter) Kpi(ctx context.Context, _ model.Filter) ([]model.KpiCard, error) {
	token, err := a.requireToken(ctx)
	if err != nil {
		return nil, err
	}
	body, err := a.get(ctx, token, "/arcopen/fn/data/arc/inc-stats")
	if err != nil {
		return nil, ErrNotImplemented
	}
	var r biliIncStatsResp
	if err := json.Unmarshal(body, &r); err != nil || r.Code != 0 {
		return nil, ErrNotImplemented
	}
	eng := r.Data.IncLike + r.Data.IncReply + r.Data.IncShare
	return []model.KpiCard{
		{Key: "creators", Label: "UP主数", Value: "1", Raw: 1, Description: "当前授权账号"},
		{Key: "followers", Label: "粉丝总量", Value: humanize(r.Data.CurrentFans), Raw: float64(r.Data.CurrentFans), Unit: unitOf(r.Data.CurrentFans), Description: "B站平台"},
		{Key: "views", Label: "播放总量", Value: humanize(r.Data.CurrentArchiveView), Raw: float64(r.Data.CurrentArchiveView), Unit: unitOf(r.Data.CurrentArchiveView), Description: "累计播放"},
		{Key: "engagement", Label: "互动总量", Value: humanize(eng), Raw: float64(eng), Description: "赞+评+转"},
		{Key: "collabs", Label: "商业合作数", Value: "-", Raw: 0, Description: "花火商单(待接入)"},
	}, nil
}

func (a *BilibiliAdapter) ViewsTrend(ctx context.Context, _ model.Filter) ([]model.ViewsTrendPoint, error) {
	token, err := a.requireToken(ctx)
	if err != nil {
		return nil, err
	}
	body, err := a.get(ctx, token, "/arcopen/fn/data/arc/data")
	if err != nil {
		return nil, ErrNotImplemented
	}
	var r biliArcDataResp
	if err := json.Unmarshal(body, &r); err != nil || r.Code != 0 {
		return nil, ErrNotImplemented
	}
	pts := make([]model.ViewsTrendPoint, 0, len(r.Data.Result))
	for i, it := range r.Data.Result {
		prev := int64(0)
		if i > 0 {
			prev = r.Data.Result[i-1].View
		}
		ratio := 0.0
		if prev > 0 {
			ratio = float64(it.View-prev) / float64(prev) * 100
		}
		pts = append(pts, model.ViewsTrendPoint{Date: it.Date, Views: it.View, PrevViews: prev, Ratio: ratio})
	}
	return pts, nil
}

func (a *BilibiliAdapter) PlatformShare(ctx context.Context, _ model.Filter) ([]model.PlatformShare, error) {
	if a.tp == nil {
		return nil, ErrNotImplemented
	}
	return []model.PlatformShare{{Platform: "B站", Share: 100, Views: 0, Color: "#FB7299"}}, nil
}

func (a *BilibiliAdapter) TrackPerformance(ctx context.Context, _ model.Filter) ([]model.TrackPerformance, error) {
	return nil, ErrNotImplemented
}
func (a *BilibiliAdapter) Radar(ctx context.Context, _ model.Filter) ([]model.RadarMetric, error) {
	return nil, ErrNotImplemented
}
func (a *BilibiliAdapter) AudienceAge(ctx context.Context, _ model.Filter) ([]model.AgeShare, error) {
	return nil, ErrNotImplemented
}
func (a *BilibiliAdapter) TopCreators(ctx context.Context, _ model.Filter) ([]model.TopCreator, error) {
	return nil, ErrNotImplemented
}
func (a *BilibiliAdapter) Options(ctx context.Context, _ model.Filter) (model.FilterOptions, error) {
	return model.FilterOptions{}, ErrNotImplemented
}
func (a *BilibiliAdapter) Insights(ctx context.Context, _ model.Filter) ([]model.Insight, error) {
	return nil, ErrNotImplemented
}

// ---- 达人分析域(回退) ----

func (a *BilibiliAdapter) CreatorKpi(ctx context.Context, f model.Filter) ([]model.KpiCard, error) {
	return nil, ErrNotImplemented
}
func (a *BilibiliAdapter) CreatorTrend(ctx context.Context, f model.Filter) ([]model.ViewsTrendPoint, error) {
	return nil, ErrNotImplemented
}
func (a *BilibiliAdapter) CreatorPlatforms(ctx context.Context, f model.Filter) ([]model.PlatformShare, error) {
	return nil, ErrNotImplemented
}
func (a *BilibiliAdapter) CreatorTracks(ctx context.Context, f model.Filter) ([]model.TrackPerformance, error) {
	return nil, ErrNotImplemented
}
func (a *BilibiliAdapter) CreatorAudience(ctx context.Context, f model.Filter) (*model.Audience, error) {
	return nil, ErrNotImplemented
}
func (a *BilibiliAdapter) CreatorList(ctx context.Context, f model.Filter) ([]model.TopCreator, error) {
	return nil, ErrNotImplemented
}

// ---- 内容分析域(回退) ----

func (a *BilibiliAdapter) ContentKpi(ctx context.Context, f model.Filter) ([]model.KpiCard, error) {
	return nil, ErrNotImplemented
}
func (a *BilibiliAdapter) ContentTrend(ctx context.Context, f model.Filter) ([]model.ViewsTrendPoint, error) {
	return nil, ErrNotImplemented
}
func (a *BilibiliAdapter) ContentForms(ctx context.Context, f model.Filter) ([]model.PlatformShare, error) {
	return nil, ErrNotImplemented
}
func (a *BilibiliAdapter) ContentTopics(ctx context.Context, f model.Filter) ([]model.TrackPerformance, error) {
	return nil, ErrNotImplemented
}
func (a *BilibiliAdapter) ContentDurations(ctx context.Context, f model.Filter) ([]model.AgeShare, error) {
	return nil, ErrNotImplemented
}
func (a *BilibiliAdapter) ContentList(ctx context.Context, f model.Filter) ([]model.ContentItem, error) {
	return nil, ErrNotImplemented
}

// ---- 市场洞察域(回退) ----

func (a *BilibiliAdapter) MarketKpi(ctx context.Context, f model.Filter) ([]model.KpiCard, error) {
	return nil, ErrNotImplemented
}
func (a *BilibiliAdapter) MarketTrend(ctx context.Context, f model.Filter) ([]model.ViewsTrendPoint, error) {
	return nil, ErrNotImplemented
}
func (a *BilibiliAdapter) MarketCompetitors(ctx context.Context, f model.Filter) ([]model.PlatformShare, error) {
	return nil, ErrNotImplemented
}
func (a *BilibiliAdapter) MarketRegions(ctx context.Context, f model.Filter) ([]model.TrackPerformance, error) {
	return nil, ErrNotImplemented
}
func (a *BilibiliAdapter) MarketPrices(ctx context.Context, f model.Filter) ([]model.AgeShare, error) {
	return nil, ErrNotImplemented
}
func (a *BilibiliAdapter) MarketList(ctx context.Context, f model.Filter) ([]model.Competitor, error) {
	return nil, ErrNotImplemented
}

// ---- 品牌分析域(回退) ----

func (a *BilibiliAdapter) BrandKpi(ctx context.Context, f model.Filter) ([]model.KpiCard, error) {
	return nil, ErrNotImplemented
}
func (a *BilibiliAdapter) BrandTrend(ctx context.Context, f model.Filter) ([]model.ViewsTrendPoint, error) {
	return nil, ErrNotImplemented
}
func (a *BilibiliAdapter) BrandPlatforms(ctx context.Context, f model.Filter) ([]model.PlatformShare, error) {
	return nil, ErrNotImplemented
}
func (a *BilibiliAdapter) BrandSentiment(ctx context.Context, f model.Filter) ([]model.AgeShare, error) {
	return nil, ErrNotImplemented
}
func (a *BilibiliAdapter) BrandKeywords(ctx context.Context, f model.Filter) ([]model.TagItem, error) {
	return nil, ErrNotImplemented
}
func (a *BilibiliAdapter) BrandList(ctx context.Context, f model.Filter) ([]model.PartnerBrand, error) {
	return nil, ErrNotImplemented
}
