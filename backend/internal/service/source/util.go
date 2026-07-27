package source

import "fmt"

// humanize 把大数格式化为带单位的中文展示串(用于 KPI Value)。
func humanize(v int64) string {
	switch {
	case v >= 1e8:
		return fmt.Sprintf("%.2f亿", float64(v)/1e8)
	case v >= 1e4:
		return fmt.Sprintf("%.2f万", float64(v)/1e4)
	default:
		return fmt.Sprintf("%d", v)
	}
}

// unitOf 返回 KPI 的单位后缀(M/B),供前端动效 / 排序用。
func unitOf(v int64) string {
	switch {
	case v >= 1e9:
		return "B"
	case v >= 1e6:
		return "M"
	default:
		return ""
	}
}
