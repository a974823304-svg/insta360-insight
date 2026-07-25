// Package mock 生成数据洞察所需的演示数据。
//
// 真实环境替换: service 层从 ClickHouse 物化视图读取,以下函数保留接口签名。
package mock

import (
	"time"

	"insta360-insight/internal/model"
)

// ============================================================
// 5 个 KPI 卡片(数据总览)
// ============================================================

// Kpi 生成 5 张总览卡的数据
//   真实来源: dws_insight_overview_kpi (T+1 物化视图, 5 分钟刷新)
func Kpi() []model.KpiCard {
	return []model.KpiCard{
		{Key: "creators", Label: "达人数", Value: "12,856", Raw: 12856, DeltaPct: 18.6, DeltaUp: true, Unit: "", Description: "较上期"},
		{Key: "followers", Label: "总粉丝", Value: "287.6M", Raw: 287_600_000, DeltaPct: 21.3, DeltaUp: true, Unit: "", Description: "较上期"},
		{Key: "views", Label: "总播放量", Value: "2.38B", Raw: 2_380_000_000, DeltaPct: 24.7, DeltaUp: true, Unit: "", Description: "较上期"},
		{Key: "engagement", Label: "互动量", Value: "46.7M", Raw: 46_700_000, DeltaPct: 19.8, DeltaUp: true, Unit: "", Description: "较上期"},
		{Key: "collabs", Label: "合作内容", Value: "3,682", Raw: 3682, DeltaPct: 17.2, DeltaUp: true, Unit: "", Description: "较上期"},
	}
}

// ============================================================
// 播放量趋势
// ============================================================

// ViewsTrend 生成近 30 天的播放量 + 上周期对比 + 异常点
func ViewsTrend() []model.ViewsTrendPoint {
	cur := TrendGenerator(30)
	// 上周期 = 往前 30 天,整体略低
	prev := make([]int64, len(cur))
	for i, v := range cur {
		prev[i] = int64(float64(v) * 0.78)
	}

	// 模拟 3 个异常点(UI 上加 tag)
	anomalyTag := map[int]string{
		8:  "抖音热门话题 #citysurf 推升当日播放量 +28%",
		17: "雪季开板,滑雪赛道自然爆发 +18%",
		23: "B站推荐算法调整, 长视频流量回升 +12%",
	}

	points := make([]model.ViewsTrendPoint, 0, len(cur))
	for i, v := range cur {
		p := model.ViewsTrendPoint{
			Date:      timeFormat(cur, i),
			Views:     v,
			PrevViews: prev[i],
		}
		if tag, ok := anomalyTag[i]; ok {
			p.HasAnomaly = true
			p.AnomalyTag = tag
		}
		if prev[i] > 0 {
			p.Ratio = float64(v-prev[i]) / float64(prev[i]) * 100
		}
		points = append(points, p)
	}
	return points
}

// timeFormat 把第 i 个点的日期格式化为 MM-DD
func timeFormat(_ []int64, i int) string {
	base := timeNow().Truncate(24 * time.Hour).AddDate(0, 0, -(29 - i))
	return base.Format("01-02")
}

// ============================================================
// 平台分布(环形图)
// ============================================================

// PlatformShare 生成平台分布
//   真实来源: dws_instight_platform_share
func PlatformShare() []model.PlatformShare {
	views := map[string]int64{
		"抖音":   1_247_000_000,
		"B站":    638_000_000,
		"小红书": 495_000_000,
	}
	var total int64
	for _, v := range views {
		total += v
	}
	out := make([]model.PlatformShare, 0, len(views))
	for p, v := range views {
		out = append(out, model.PlatformShare{
			Platform: p,
			Share:    RoundTo(float64(v)/float64(total)*100, 1),
			Views:    v,
			Color:    PlatformColor(p),
		})
	}
	// 按 share 倒序,UI 显示更顺眼
	sortByShareDesc(out)
	return out
}

// ============================================================
// 运动赛道表现(横向条形图)
// ============================================================

// TrackPerformance 生成赛道横向条形
func TrackPerformance() []model.TrackPerformance {
	data := []model.TrackPerformance{
		{Track: "滑雪", Views: 878_000_000, Color: TrackColor("滑雪")},
		{Track: "冲浪", Views: 642_000_000, Color: TrackColor("冲浪")},
		{Track: "骑行", Views: 456_000_000, Color: TrackColor("骑行")},
		{Track: "潜水", Views: 228_000_000, Color: TrackColor("潜水")},
		{Track: "攀岩", Views: 176_000_000, Color: TrackColor("攀岩")},
	}
	return data
}

// ============================================================
// 引爆力维度(雷达图)
// ============================================================

// ExplosiveRadar 生成 4 维引爆力
func ExplosiveRadar() []model.RadarMetric {
	return []model.RadarMetric{
		{Dimension: "内容质量", Value: 92, Avg: 78},
		{Dimension: "粉丝互动力", Value: 78, Avg: 70},
		{Dimension: "商业配合力", Value: 85, Avg: 72},
		{Dimension: "成活性", Value: 81, Avg: 68},
		{Dimension: "成长力", Value: 88, Avg: 73},
	}
}

// ============================================================
// 粉丝年龄分布
// ============================================================

// AgeShare 粉丝画像
func AgeShare() []model.AgeShare {
	data := []model.AgeShare{
		{Bucket: "18-24 岁", Share: 31.2, Color: AgeColor("18-24 岁")},
		{Bucket: "25-34 岁", Share: 42.7, Color: AgeColor("25-34 岁")},
		{Bucket: "35-44 岁", Share: 17.3, Color: AgeColor("35-44 岁")},
		{Bucket: "45 岁以上", Share: 8.8, Color: AgeColor("45 岁以上")},
	}
	return data
}

// ============================================================
// AI 关键洞察
// ============================================================

// AIInsights 默认 3 条洞察(无大模型时返回)
//   真实来源: AI 引擎 (/api/ai/insights) 返回或离线预生成
func AIInsights() []model.Insight {
	return []model.Insight{
		{
			Icon:     "surge",
			Title:    "冲浪赛道增长迅猛",
			Body:     "冲浪相关内容在近 30 天内播放量增长 38.7%,显著高于整体水平(24.7%),建议加大优选达人合作力度。",
			Severity: "success",
		},
		{
			Icon:     "star",
			Title:    "抖音表现突出",
			Body:     "抖音平台播放量占比提升至 52.4%, 同比增长 24.6%, 建议加大抖音渠道投放及头部达人合作。",
			Severity: "info",
		},
		{
			Icon:     "alert",
			Title:     "黑马达人浮现",
			Body:     "粉丝数 < 5 万但播放量 > 30 万的\"内容创作者\" +62%,具备较高投资价值,建议关注 #极限运动 标签下的新锐创作者。",
			Severity: "warning",
		},
	}
}

// ============================================================
// 热门达人 TOP 10
// ============================================================

// TopCreators 生成热门达人列表
func TopCreators() []model.TopCreator {
	return []model.TopCreator{
		{Rank: 1, Avatar: "🤿", Name: "Chris Burkard", Platform: "抖音", Followers: 1_020_000, TotalViews: 186_300_000, Engagement: 6.72, Growth30d: 12.4, Explosive: 89.7, Tags: []string{"#骑行", "#极限", "#摄影"}},
		{Rank: 2, Avatar: "🚴", Name: "Sophie Laurent", Platform: "小红书", Followers: 326_800, TotalViews: 64_200_000, Engagement: 8.91, Growth30d: 18.7, Explosive: 88.6, Tags: []string{"#冲浪", "#滑翔伞", "#生活方式"}},
		{Rank: 3, Avatar: "⛷️", Name: "Jake Wetter", Platform: "抖音", Followers: 48_700, TotalViews: 18_600_000, Engagement: 12.38, Growth30d: 42.3, Blacklist: true, Explosive: 91.2, Tags: []string{"#滑雪", "#极限运动", "#户外"}},
		{Rank: 4, Avatar: "🏄", Name: "Marina Costa", Platform: "B站", Followers: 890_400, TotalViews: 124_500_000, Engagement: 9.65, Growth30d: 15.1, Explosive: 86.4, Tags: []string{"#冲浪", "#海洋", "#环保"}},
		{Rank: 5, Avatar: "🧗", Name: "Liam Hoffmann", Platform: "抖音", Followers: 215_600, TotalViews: 38_900_000, Engagement: 7.43, Growth30d: 9.8, Explosive: 82.1, Tags: []string{"#攀岩", "#登山", "#探险"}},
		{Rank: 6, Avatar: "🏂", Name: "Yuki Tanaka", Platform: "B站", Followers: 1_540_000, TotalViews: 248_700_000, Engagement: 8.12, Growth30d: 21.6, Explosive: 87.9, Tags: []string{"#滑雪", "#单板", "#冬季"}},
		{Rank: 7, Avatar: "🚵", Name: "Felix Becker", Platform: "小红书", Followers: 432_900, TotalViews: 71_200_000, Engagement: 6.85, Growth30d: 6.4, Explosive: 79.5, Tags: []string{"#骑行", "#公路车", "#训练"}},
		{Rank: 8, Avatar: "🤽", Name: "Aria Chen", Platform: "抖音", Followers: 678_200, TotalViews: 96_500_000, Engagement: 10.27, Growth30d: 28.4, Explosive: 88.2, Tags: []string{"#潜水", "#自由潜", "#旅行"}},
		{Rank: 9, Avatar: "🏃", Name: "Marco Rivera", Platform: "抖音", Followers: 312_400, TotalViews: 52_300_000, Engagement: 5.96, Growth30d: 4.7, Explosive: 76.8, Tags: []string{"#越野跑", "#训练", "#营养"}},
		{Rank: 10, Avatar: "🎿", Name: "Elena Petrova", Platform: "小红书", Followers: 198_500, TotalViews: 28_700_000, Engagement: 9.74, Growth30d: 11.9, Explosive: 84.3, Tags: []string{"#滑雪", "#阿尔卑斯", "#度假"}},
	}
}

// ============================================================
// 筛选面板的可选项
// ============================================================

// FilterOptions 返回筛选面板下拉数据
func FilterOptions() model.FilterOptions {
	return model.FilterOptions{
		Regions: toOptions(Regions),
		Tracks:  toOptions(Tracks),
		Platforms: []model.Option{
			{Label: "抖音", Value: "抖音"},
			{Label: "B站", Value: "B站"},
			{Label: "小红书", Value: "小红书"},
		},
		AgeBands: toOptions(AgeBands),
		Presets: []model.Option{
			{Label: "近 7 天", Value: "7d"},
			{Label: "近 30 天", Value: "30d"},
			{Label: "近 90 天", Value: "90d"},
			{Label: "本季度", Value: "this_quarter"},
		},
	}
}

func toOptions(ss []string) []model.Option {
	out := make([]model.Option, 0, len(ss))
	for _, s := range ss {
		out = append(out, model.Option{Label: s, Value: s})
	}
	return out
}

// sortByShareDesc 按 share 倒序
func sortByShareDesc(s []model.PlatformShare) {
	// 简单插入排序,长度 < 10
	for i := 1; i < len(s); i++ {
		j := i
		for j > 0 && s[j].Share > s[j-1].Share {
			s[j], s[j-1] = s[j-1], s[j]
			j--
		}
	}
}
