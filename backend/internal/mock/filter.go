package mock

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"insta360-insight/internal/model"
)

const (
	totalPlatforms = 3 // 抖音/B站/小红书
	totalRegions   = 4 // 北美/欧洲/亚太/全球
	totalTracks    = 5 // 滑雪/冲浪/骑行/潜水/攀岩
)

func clamp(x, lo, hi float64) float64 {
	if x < lo {
		return lo
	}
	if x > hi {
		return hi
	}
	return x
}

// WeightedFactor 平台/地区/赛道加权占比, clamp[0.3,1.0]; 空选=1.0
func WeightedFactor(f model.Filter) float64 {
	frac := func(sel, total int) float64 {
		if sel == 0 {
			return 1.0
		}
		return float64(sel) / float64(total)
	}
	w := 0.5*frac(len(f.Platforms), totalPlatforms) +
		0.3*frac(len(f.Regions), totalRegions) +
		0.2*frac(len(f.Tracks), totalTracks)
	return clamp(w, 0.3, 1.0)
}

// DateScale 时间跨度缩放, 30 天基准, clamp[0.3,5.0]
func DateScale(f model.Filter) float64 {
	if len(f.DateRange) != 2 {
		return 1.0
	}
	start, err1 := time.Parse("2006-01-02", f.DateRange[0])
	end, err2 := time.Parse("2006-01-02", f.DateRange[1])
	if err1 != nil || err2 != nil || !end.After(start) {
		return 1.0
	}
	span := end.Sub(start).Hours()/24 + 1
	return clamp(span/30.0, 0.3, 5.0)
}

// KpiScale 综合因子 = weighted * date, clamp[0.3,5.0]
func KpiScale(f model.Filter) float64 {
	return clamp(WeightedFactor(f)*DateScale(f), 0.3, 5.0)
}

// FormatScaledValue 依据原 Value 的格式模式, 用新 raw 重新格式化(确定性)
func FormatScaledValue(orig string, raw float64) string {
	switch {
	case strings.Contains(orig, "%"):
		return fmt.Sprintf("%.1f%%", raw)
	case strings.Contains(orig, "¥"):
		rest := strings.TrimPrefix(orig, "¥")
		return "¥" + formatBySuffix(rest, raw)
	case strings.Contains(orig, "B"):
		return fmt.Sprintf("%.2fB", raw/1e9)
	case strings.Contains(orig, "M"):
		return fmt.Sprintf("%.1fM", raw/1e6)
	case strings.Contains(orig, "K"):
		return fmt.Sprintf("%.1fK", raw/1e3)
	case strings.Contains(orig, ","):
		return humanizeComma(int64(raw))
	case strings.Contains(orig, "."):
		return fmt.Sprintf("%.1f", raw)
	default:
		return strconv.Itoa(int(raw))
	}
}

func formatBySuffix(suffix string, raw float64) string {
	switch {
	case strings.Contains(suffix, "B"):
		return fmt.Sprintf("%.2fB", raw/1e9)
	case strings.Contains(suffix, "M"):
		return fmt.Sprintf("%.1fM", raw/1e6)
	case strings.Contains(suffix, "K"):
		return fmt.Sprintf("%.1fK", raw/1e3)
	case strings.Contains(suffix, ","):
		return humanizeComma(int64(raw))
	case strings.Contains(suffix, "."):
		return fmt.Sprintf("%.1f", raw)
	default:
		return strconv.Itoa(int(raw))
	}
}

func humanizeComma(n int64) string {
	s := strconv.FormatInt(n, 10)
	var b strings.Builder
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			b.WriteByte(',')
		}
		b.WriteRune(c)
	}
	return b.String()
}

// ScaleKpis 缩放每个 KPI 的 Raw 并用原格式重算 Value
func ScaleKpis(kpis []model.KpiCard, f model.Filter) []model.KpiCard {
	k := KpiScale(f)
	out := make([]model.KpiCard, len(kpis))
	for i, c := range kpis {
		nc := c
		nc.Raw = c.Raw * k
		nc.Value = FormatScaledValue(c.Value, nc.Raw)
		out[i] = nc
	}
	return out
}

// WindowTrend 按 dateRange 重采样 x 轴并缩放单点值(上限 60 点)
func WindowTrend(points []model.ViewsTrendPoint, f model.Filter) []model.ViewsTrendPoint {
	if len(f.DateRange) != 2 {
		return points // 无日期筛选: 保持默认 30 天
	}
	start, err1 := time.Parse("2006-01-02", f.DateRange[0])
	end, err2 := time.Parse("2006-01-02", f.DateRange[1])
	if err1 != nil || err2 != nil || !end.After(start) {
		return points
	}
	days := int(end.Sub(start).Hours()/24) + 1
	step := 1
	if days > 60 {
		step = (days + 59) / 60
	}
	scale := DateScale(f)
	out := make([]model.ViewsTrendPoint, 0, days)
	for d := 0; d < days; d += step {
		dt := start.AddDate(0, 0, d)
		idx := d % len(points)
		p := points[idx]
		out = append(out, model.ViewsTrendPoint{
			Date:      dt.Format("01-02"),
			Views:     int64(float64(p.Views) * scale),
			PrevViews: int64(float64(p.PrevViews) * scale),
		})
	}
	return out
}

func inSet(keys, selected []string) bool {
	if len(selected) == 0 {
		return true
	}
	for _, s := range selected {
		for _, k := range keys {
			if k == s {
				return true
			}
		}
	}
	return false
}

// RenormalizePlatformShares 仅保留 selected 中的 Platform, 重算 share≈100
func RenormalizePlatformShares(items []model.PlatformShare, selected []string) []model.PlatformShare {
	if len(selected) == 0 {
		return items
	}
	var kept []model.PlatformShare
	var sum float64
	for _, it := range items {
		if contains(selected, it.Platform) {
			kept = append(kept, it)
			sum += it.Share
		}
	}
	if sum <= 0 { // 防零除: 回退原样
		return items
	}
	for i := range kept {
		kept[i].Share = RoundTo(kept[i].Share/sum*100, 1)
	}
	return kept
}

// RenormalizeTrackShares 仅保留 selected 中的 Track, 按 Views 重算占比≈100
func RenormalizeTrackShares(items []model.TrackPerformance, selected []string) []model.TrackPerformance {
	if len(selected) == 0 {
		return items
	}
	var kept []model.TrackPerformance
	var sum float64
	for _, it := range items {
		if contains(selected, it.Track) {
			kept = append(kept, it)
			sum += float64(it.Views)
		}
	}
	if sum <= 0 {
		return items
	}
	for i := range kept {
		kept[i].Views = int64(float64(kept[i].Views) / sum * 100)
	}
	return kept
}

// RenormalizeAgeShares 仅保留 selected 中的 Bucket, 重算 share≈100
func RenormalizeAgeShares(items []model.AgeShare, selected []string) []model.AgeShare {
	if len(selected) == 0 {
		return items
	}
	var kept []model.AgeShare
	var sum float64
	for _, it := range items {
		if contains(selected, it.Bucket) {
			kept = append(kept, it)
			sum += it.Share
		}
	}
	if sum <= 0 {
		return items
	}
	for i := range kept {
		kept[i].Share = RoundTo(kept[i].Share/sum*100, 1)
	}
	return kept
}

func contains(s []string, v string) bool {
	for _, x := range s {
		if x == v {
			return true
		}
	}
	return false
}

// FilterTopCreators 按 平台 + 赛道(从 Tags 去#) AND 过滤
func FilterTopCreators(rows []model.TopCreator, f model.Filter) []model.TopCreator {
	if len(f.Platforms) == 0 && len(f.Tracks) == 0 {
		return rows
	}
	out := make([]model.TopCreator, 0, len(rows))
	for _, r := range rows {
		platOK := len(f.Platforms) == 0 || contains(f.Platforms, r.Platform)
		var trackKeys []string
		for _, t := range r.Tags {
			trackKeys = append(trackKeys, strings.TrimPrefix(t, "#"))
		}
		trackOK := len(f.Tracks) == 0 || inSet(trackKeys, f.Tracks)
		if platOK && trackOK {
			out = append(out, r)
		}
	}
	if len(out) == 0 { // 不空表: 回退全量
		return rows
	}
	return out
}

// FilterContentItems 按 赛道(Topic) 过滤; 平台无字段恒命中
func FilterContentItems(rows []model.ContentItem, f model.Filter) []model.ContentItem {
	if len(f.Tracks) == 0 {
		return rows
	}
	out := make([]model.ContentItem, 0, len(rows))
	for _, r := range rows {
		if contains(f.Tracks, r.Topic) {
			out = append(out, r)
		}
	}
	if len(out) == 0 {
		return rows
	}
	return out
}
