// Package mock 提供确定性的演示数据集,模拟数仓预聚合输出。
//
// 当前直接生成内存数据用于前后端联调。后续接入真实 OLAP 时:
//   - GenerateDataset() 由 ETL / 物化视图初始化器替换
//   - 数据集结构保持不变,service 层无需感知
//
// 注释风格:每个生成函数顶部说明它的"业务含义"+"在真实数仓中的来源表",
//          方便后续按字段映射到 ClickHouse / Doris。
package mock

import (
	"fmt"
	"math"
	"time"
)

// ============================================================
// 演示用基础数据(模拟平台维度的原始事实表)
// ============================================================

// Platforms 模拟外部平台枚举
var Platforms = []string{"抖音", "B站", "小红书"}

// Tracks 运动赛道
var Tracks = []string{"滑雪", "冲浪", "骑行", "潜水", "攀岩"}

// Regions 销售地区
var Regions = []string{"北美", "欧洲", "亚太", "全球"}

// AgeBands 粉丝年龄段
var AgeBands = []string{"18-24 岁", "25-34 岁", "35-44 岁", "45 岁以上"}

// ============================================================
// 数字格式化工具(与前端展示口径一致,单位 B/M/K)
// ============================================================

// FormatCount 把大数字格式化为 1.23B / 45.6M / 7.8K
func FormatCount(v int64) string {
	f := float64(v)
	switch {
	case f >= 1_000_000_000:
		return fmt.Sprintf("%.2fB", f/1_000_000_000)
	case f >= 1_000_000:
		return fmt.Sprintf("%.1fM", f/1_000_000)
	case f >= 1_000:
		return fmt.Sprintf("%.1fK", f/1_000)
	default:
		return fmt.Sprintf("%d", v)
	}
}

// FormatPct 百分比保留 1 位小数
func FormatPct(v float64) string {
	return fmt.Sprintf("%.1f%%", v)
}

// RoundTo 保留 n 位小数
func RoundTo(v float64, n int) float64 {
	p := math.Pow10(n)
	return math.Round(v*p) / p
}

// ============================================================
// 平台/赛道/年龄配色(与前端 ECharts 一致)
// ============================================================

// PlatformColor 平台配色
func PlatformColor(p string) string {
	switch p {
	case "抖音":
		return "#000000"
	case "B站":
		return "#00A1D6"
	case "小红书":
		return "#FF2442"
	default:
		return "#7B61FF"
	}
}

// TrackColor 运动赛道配色
func TrackColor(t string) string {
	switch t {
	case "滑雪":
		return "#5EA1FF"
	case "冲浪":
		return "#3DD9EB"
	case "骑行":
		return "#A07BFF"
	case "潜水":
		return "#4DD0E1"
	case "攀岩":
		return "#7DD96E"
	}
	return "#888"
}

// AgeColor 粉丝年龄段配色
func AgeColor(bucket string) string {
	switch bucket {
	case "18-24 岁":
		return "#3DD9EB"
	case "25-34 岁":
		return "#FFB547"
	case "35-44 岁":
		return "#A07BFF"
	case "45 岁以上":
		return "#FF6B6B"
	}
	return "#888"
}

// ============================================================
// 趋势生成(确定性的伪随机)
// ============================================================

// TrendGenerator 生成带周末效应 + 缓慢上升的播放量序列
//   真实数仓来源: dws_instagram_daily_creator_view
func TrendGenerator(days int) []int64 {
	points := make([]int64, 0, days)
	base := time.Now().Truncate(24 * time.Hour).AddDate(0, 0, -(days - 1))
	for i := 0; i < days; i++ {
		d := base.AddDate(0, 0, i)
		trend := int64(72_000_000) + int64(i)*1_200_000
		var weekend int64
		if d.Weekday() == time.Saturday || d.Weekday() == time.Sunday {
			weekend = 18_000_000
		}
		noise := int64((i*7+13)%9) * 1_500_000
		points = append(points, trend+weekend+noise)
	}
	return points
}
