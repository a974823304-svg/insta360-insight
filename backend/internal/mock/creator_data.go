// Package mock 生成数据洞察所需的演示数据。
package mock

import (
	"strings"
	"time"

	"insta360-insight/internal/model"
)

// ============================================================
// 达人分析 KPI(5 张卡)
// ============================================================
func CreatorKpi() []model.KpiCard {
	return []model.KpiCard{
		{Key: "creators", Label: "达人总数", Value: "20", Raw: 20, DeltaPct: 8.5, DeltaUp: true, Description: "较上期"},
		{Key: "new", Label: "本月新增", Value: "3", Raw: 3, DeltaPct: 50, DeltaUp: true, Description: "较上期"},
		{Key: "followers", Label: "平均粉丝", Value: FormatCount(1_820_000), Raw: 1_820_000, DeltaPct: 4.2, DeltaUp: true, Description: "较上期"},
		{Key: "engagement", Label: "平均互动率", Value: "6.8%", Raw: 6.8, DeltaPct: 0.6, DeltaUp: true, Description: "较上期"},
		{Key: "collabs", Label: "合作占比", Value: "45%", Raw: 45, DeltaPct: 12, DeltaUp: true, Description: "较上期"},
	}
}

// ============================================================
// 粉丝规模趋势(近 30 天累计粉丝, M 量级, 复用 ViewsTrendPoint 形状)
// ============================================================
func CreatorTrend() []model.ViewsTrendPoint {
	base := int64(180_000_000)
	points := make([]model.ViewsTrendPoint, 0, 30)
	now := time.Now().Truncate(24 * time.Hour)
	for i := 0; i < 30; i++ {
		v := base + int64(i)*180_000 + int64((i*5)%9)*40_000
		prev := base + int64(i)*150_000
		d := now.AddDate(0, 0, -(29 - i))
		points = append(points, model.ViewsTrendPoint{
			Date:      d.Format("01-02"),
			Views:     v,
			PrevViews: prev,
		})
	}
	return points
}

// ============================================================
// 达人列表(20 个, 覆盖 3 平台 × 5 赛道)
// ============================================================
func CreatorList() []model.TopCreator {
	seeds := []struct {
		name, avatar string
	}{
		{"Chris Burkard", "🤿"}, {"Sophie Laurent", "🚴"}, {"Jake Wetter", "⛷️"},
		{"Marina Costa", "🏄"}, {"Liam Hoffmann", "🧗"}, {"Yuki Tanaka", "🏂"},
		{"Felix Becker", "🚵"}, {"Aria Chen", "🤽"}, {"Marco Rivera", "🏃"},
		{"Elena Petrova", "🎿"}, {"Noah Kim", "🏄"}, {"Mia Wong", "🚴"},
		{"Leo Schmidt", "⛷️"}, {"Zoe Martin", "🧗"}, {"Kenji Sato", "🏂"},
		{"Lara Lopez", "🤽"}, {"Owen Brooks", "🏃"}, {"Nina Roth", "🎿"},
		{"Pablo Cruz", "🚵"}, {"Sara Lind", "🤿"},
	}
	platforms := []string{"抖音", "B站", "小红书"}
	out := make([]model.TopCreator, 0, len(seeds))
	for i, s := range seeds {
		p := platforms[i%3]
		t := Tracks[i%len(Tracks)]
		followers := int64(80_000 + (i*37_000)%1_500_000)
		out = append(out, model.TopCreator{
			Rank:       i + 1,
			Avatar:     s.avatar,
			Name:       s.name,
			Platform:   p,
			Followers:  followers,
			TotalViews: followers * (3 + int64(i)%5),
			Engagement: RoundTo(5.0+float64(i%6)+float64(i%3)*0.4, 2),
			Growth30d:  RoundTo(float64((i*7)%45)-5.0, 1),
			Explosive:  RoundTo(70.0+float64(i%25)+float64(i%4)*1.5, 1),
			Blacklist:  i == 2,
			Tags:       []string{"#" + t, "#极限"},
		})
	}
	return out
}

// ============================================================
// 平台分布(按粉丝量聚合)
// ============================================================
func CreatorPlatforms() []model.PlatformShare {
	list := CreatorList()
	m := map[string]int64{}
	for _, c := range list {
		m[c.Platform] += c.Followers
	}
	var total int64
	for _, v := range m {
		total += v
	}
	out := make([]model.PlatformShare, 0, len(m))
	for p, v := range m {
		out = append(out, model.PlatformShare{
			Platform: p,
			Share:    RoundTo(float64(v)/float64(total)*100, 1),
			Views:    v,
			Color:    PlatformColor(p),
		})
	}
	sortByShareDesc(out)
	return out
}

// ============================================================
// 赛道粉丝分布(按粉丝量聚合, M 量级)
// ============================================================
func CreatorTracks() []model.TrackPerformance {
	list := CreatorList()
	m := map[string]int64{}
	for _, c := range list {
		t := strings.TrimPrefix(c.Tags[0], "#")
		m[t] += c.Followers
	}
	out := make([]model.TrackPerformance, 0, len(m))
	for t, v := range m {
		out = append(out, model.TrackPerformance{
			Track: t,
			Views: v,
			Color: TrackColor(t),
		})
	}
	return out
}

// ============================================================
// 粉丝画像(年龄 + 性别)
// ============================================================
func CreatorAudience() model.Audience {
	return model.Audience{
		Age: []model.AgeShare{
			{Bucket: "18-24 岁", Share: 33.5, Color: AgeColor("18-24 岁")},
			{Bucket: "25-34 岁", Share: 41.2, Color: AgeColor("25-34 岁")},
			{Bucket: "35-44 岁", Share: 17.8, Color: AgeColor("35-44 岁")},
			{Bucket: "45 岁以上", Share: 7.5, Color: AgeColor("45 岁以上")},
		},
		Gender: []model.GenderShare{
			{Gender: "男", Share: 58.4, Color: "#5EA1FF"},
			{Gender: "女", Share: 41.6, Color: "#FF6B6B"},
		},
	}
}
