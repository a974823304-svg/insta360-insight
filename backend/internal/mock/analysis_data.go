// Package mock 生成数据洞察所需的演示数据。
package mock

import (
	"strconv"
	"time"

	"insta360-insight/internal/model"
)

// itoa 包内整型转字符串（mock 包未引入 strconv 全局别名）
func itoa(n int) string { return strconv.Itoa(n) }

// ============================================================
// 内容分析
// ============================================================
func ContentKpi(f model.Filter) []model.KpiCard {
	k := []model.KpiCard{
		{Key: "total", Label: "内容总数", Value: "1,284", Raw: 1284, DeltaPct: 12.3, DeltaUp: true, Description: "较上期"},
		{Key: "avg_views", Label: "平均播放", Value: "86.4K", Raw: 86400, DeltaPct: 5.1, DeltaUp: true, Description: "较上期"},
		{Key: "hit_rate", Label: "爆款率", Value: "8.2%", Raw: 8.2, DeltaPct: 1.4, DeltaUp: true, Description: "较上期"},
		{Key: "engage", Label: "平均互动率", Value: "7.3%", Raw: 7.3, DeltaPct: 0.3, DeltaUp: false, Description: "较上期"},
		{Key: "freq", Label: "周更新频次", Value: "42", Raw: 42, DeltaPct: 6.0, DeltaUp: true, Description: "较上期"},
	}
	return ScaleKpis(k, f)
}

func ContentTrend(f model.Filter) []model.ViewsTrendPoint {
	base := int64(2_400_000)
	points := make([]model.ViewsTrendPoint, 0, 30)
	now := time.Now().Truncate(24 * time.Hour)
	for i := 0; i < 30; i++ {
		v := base + int64(i)*120_000 + int64((i*7)%11)*30_000
		prev := base + int64(i)*95_000
		d := now.AddDate(0, 0, -(29 - i))
		points = append(points, model.ViewsTrendPoint{
			Date:      d.Format("01-02"),
			Views:     v,
			PrevViews: prev,
		})
	}
	return WindowTrend(points, f)
}

func ContentForms(f model.Filter) []model.PlatformShare {
	return []model.PlatformShare{
		{Platform: "教程", Share: 38.0, Views: 920000, Color: "#3DD9EB"},
		{Platform: "测评", Share: 27.0, Views: 650000, Color: "#5EA1FF"},
		{Platform: "创意短片", Share: 21.0, Views: 510000, Color: "#A07BFF"},
		{Platform: "挑战赛", Share: 14.0, Views: 340000, Color: "#7DD96E"},
	}
}

func ContentTopics(f model.Filter) []model.TrackPerformance {
	return []model.TrackPerformance{
		{Track: "滑雪", Views: 720000, Color: "#5EA1FF"},
		{Track: "潜水", Views: 610000, Color: "#3DD9EB"},
		{Track: "骑行", Views: 530000, Color: "#A07BFF"},
		{Track: "旅行", Views: 470000, Color: "#7DD96E"},
		{Track: "Vlog", Views: 390000, Color: "#FFB547"},
	}
}

func ContentDurations(f model.Filter) []model.AgeShare {
	return []model.AgeShare{
		{Bucket: "≤30s", Share: 34.0, Color: "#3DD9EB"},
		{Bucket: "30-60s", Share: 41.0, Color: "#5EA1FF"},
		{Bucket: "1-3min", Share: 18.0, Color: "#A07BFF"},
		{Bucket: "3min+", Share: 7.0, Color: "#7DD96E"},
	}
}

func ContentList(f model.Filter) []model.ContentItem {
	forms := []string{"教程", "测评", "创意短片", "挑战赛"}
	topics := []string{"滑雪", "潜水", "骑行", "旅行", "Vlog"}
	out := make([]model.ContentItem, 0, 15)
	for i := 0; i < 15; i++ {
		out = append(out, model.ContentItem{
			ID:         int64(i + 1),
			Title:      topics[i%len(topics)] + "第" + itoa(i+1) + "期",
			Form:       forms[i%len(forms)],
			Topic:      topics[i%len(topics)],
			Views:      int64(50_000 + (i*53_000)%1_200_000),
			Engagement: RoundTo(5.0+float64(i%9)+float64(i%4)*0.6, 2),
			IsHit:      i%3 == 0,
		})
	}
	return FilterContentItems(out, f)
}

// ============================================================
// 市场洞察
// ============================================================
func MarketKpi(f model.Filter) []model.KpiCard {
	k := []model.KpiCard{
		{Key: "size", Label: "品类规模", Value: "¥3.2B", Raw: 3200000000, DeltaPct: 9.8, DeltaUp: true, Description: "较上期"},
		{Key: "growth", Label: "品类增速", Value: "14.6%", Raw: 14.6, DeltaPct: 2.1, DeltaUp: true, Description: "较上期"},
		{Key: "share", Label: "Insta360 市占", Value: "31.2%", Raw: 31.2, DeltaPct: 3.4, DeltaUp: true, Description: "较上期"},
		{Key: "comp", Label: "在榜竞品数", Value: "6", Raw: 6, DeltaPct: 0, DeltaUp: true, Description: "较上期"},
		{Key: "buzz", Label: "行业声量", Value: "4.7M", Raw: 4700000, DeltaPct: 7.5, DeltaUp: true, Description: "较上期"},
	}
	return ScaleKpis(k, f)
}

func MarketTrend(f model.Filter) []model.ViewsTrendPoint {
	base := int64(140_000)
	points := make([]model.ViewsTrendPoint, 0, 30)
	now := time.Now().Truncate(24 * time.Hour)
	for i := 0; i < 30; i++ {
		v := base + int64(i)*9_000 + int64((i*5)%9)*2_000
		prev := base + int64(i)*7_000
		d := now.AddDate(0, 0, -(29 - i))
		points = append(points, model.ViewsTrendPoint{
			Date:      d.Format("01-02"),
			Views:     v,
			PrevViews: prev,
		})
	}
	return WindowTrend(points, f)
}

func MarketCompetitors(f model.Filter) []model.PlatformShare {
	return []model.PlatformShare{
		{Platform: "Insta360", Share: 31.2, Views: 1460000, Color: "#3DD9EB"},
		{Platform: "GoPro", Share: 28.4, Views: 1330000, Color: "#FF6B6B"},
		{Platform: "DJI", Share: 24.7, Views: 1150000, Color: "#5EA1FF"},
		{Platform: "其他", Share: 15.7, Views: 730000, Color: "#A07BFF"},
	}
}

func MarketRegions(f model.Filter) []model.TrackPerformance {
	return []model.TrackPerformance{
		{Track: "华东", Views: 1520000, Color: "#3DD9EB"},
		{Track: "华南", Views: 1180000, Color: "#5EA1FF"},
		{Track: "华北", Views: 960000, Color: "#A07BFF"},
		{Track: "西南", Views: 640000, Color: "#7DD96E"},
		{Track: "海外", Views: 1300000, Color: "#FFB547"},
	}
}

func MarketPrices(f model.Filter) []model.AgeShare {
	return []model.AgeShare{
		{Bucket: "<1000", Share: 22.0, Color: "#3DD9EB"},
		{Bucket: "1000-3000", Share: 41.0, Color: "#5EA1FF"},
		{Bucket: "3000-5000", Share: 26.0, Color: "#A07BFF"},
		{Bucket: "5000+", Share: 11.0, Color: "#7DD96E"},
	}
}

func MarketList(f model.Filter) []model.Competitor {
	return []model.Competitor{
		{Name: "Insta360", Category: "全景/运动相机", Buzz: 1460000, Growth: 17.2, Sentiment: 81.0},
		{Name: "GoPro", Category: "运动相机", Buzz: 1330000, Growth: 4.1, Sentiment: 73.0},
		{Name: "DJI", Category: "无人机/相机", Buzz: 1150000, Growth: 9.8, Sentiment: 76.0},
		{Name: "Sony", Category: "微单", Buzz: 880000, Growth: 6.3, Sentiment: 79.0},
		{Name: "大疆Action", Category: "运动相机", Buzz: 620000, Growth: 12.5, Sentiment: 74.0},
		{Name: "AKASO", Category: "入门运动相机", Buzz: 340000, Growth: 21.0, Sentiment: 68.0},
	}
}

// ============================================================
// 品牌分析
// ============================================================
func BrandKpi(f model.Filter) []model.KpiCard {
	k := []model.KpiCard{
		{Key: "buzz", Label: "品牌声量", Value: "2.9M", Raw: 2900000, DeltaPct: 11.2, DeltaUp: true, Description: "较上期"},
		{Key: "sent", Label: "好感度", Value: "81%", Raw: 81, DeltaPct: 1.8, DeltaUp: true, Description: "较上期"},
		{Key: "partners", Label: "合作品牌数", Value: "8", Raw: 8, DeltaPct: 14.3, DeltaUp: true, Description: "较上期"},
		{Key: "roi", Label: "内容互动 ROI", Value: "4.2", Raw: 4.2, DeltaPct: 0.5, DeltaUp: true, Description: "较上期"},
		{Key: "search", Label: "搜索指数", Value: "68.5", Raw: 68.5, DeltaPct: 3.2, DeltaUp: false, Description: "较上期"},
	}
	return ScaleKpis(k, f)
}

func BrandTrend(f model.Filter) []model.ViewsTrendPoint {
	base := int64(90_000)
	points := make([]model.ViewsTrendPoint, 0, 30)
	now := time.Now().Truncate(24 * time.Hour)
	for i := 0; i < 30; i++ {
		v := base + int64(i)*6_000 + int64((i*4)%7)*1_500
		prev := base + int64(i)*5_000
		d := now.AddDate(0, 0, -(29 - i))
		points = append(points, model.ViewsTrendPoint{
			Date:      d.Format("01-02"),
			Views:     v,
			PrevViews: prev,
		})
	}
	return WindowTrend(points, f)
}

func BrandPlatforms(f model.Filter) []model.PlatformShare {
	out := []model.PlatformShare{
		{Platform: "抖音", Share: 44.0, Views: 1280000, Color: "#000000"},
		{Platform: "B站", Share: 26.0, Views: 754000, Color: "#00A1D6"},
		{Platform: "小红书", Share: 22.0, Views: 638000, Color: "#FF2442"},
		{Platform: "微博", Share: 8.0, Views: 232000, Color: "#E6162D"},
	}
	return RenormalizePlatformShares(out, f.Platforms)
}

func BrandSentiment(f model.Filter) []model.AgeShare {
	return []model.AgeShare{
		{Bucket: "正面", Share: 67.0, Color: "#7DD96E"},
		{Bucket: "中性", Share: 24.0, Color: "#FFB547"},
		{Bucket: "负面", Share: 9.0, Color: "#FF6B6B"},
	}
}

func BrandKeywords(f model.Filter) []model.TagItem {
	return []model.TagItem{
		{Word: "画质", Weight: 92.0},
		{Word: "防抖", Weight: 88.0},
		{Word: "全景", Weight: 81.0},
		{Word: "运动相机", Weight: 76.0},
		{Word: "Vlog", Weight: 70.0},
		{Word: "旅行", Weight: 64.0},
		{Word: "性价比", Weight: 58.0},
		{Word: "续航", Weight: 49.0},
	}
}

func BrandList(f model.Filter) []model.PartnerBrand {
	return []model.PartnerBrand{
		{Name: "红牛", Industry: "饮料", Contents: 8, Exposure: 5400000, Engagement: 320000, ROI: 4.2},
		{Name: "始祖鸟", Industry: "户外", Contents: 6, Exposure: 4100000, Engagement: 260000, ROI: 3.9},
		{Name: "携程", Industry: "旅行", Contents: 5, Exposure: 3300000, Engagement: 210000, ROI: 3.5},
		{Name: "迪卡侬", Industry: "运动", Contents: 7, Exposure: 2900000, Engagement: 190000, ROI: 3.1},
		{Name: "大疆", Industry: "无人机", Contents: 4, Exposure: 2600000, Engagement: 175000, ROI: 3.8},
		{Name: "索尼", Industry: "影像", Contents: 3, Exposure: 1800000, Engagement: 120000, ROI: 3.3},
		{Name: "Keep", Industry: "健身", Contents: 5, Exposure: 1500000, Engagement: 98000, ROI: 2.9},
		{Name: "小米", Industry: "科技", Contents: 2, Exposure: 980000, Engagement: 64000, ROI: 2.6},
	}
}
