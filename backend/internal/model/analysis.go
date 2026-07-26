// Package model 包含内容/市场/品牌分析所需的列表项类型。
package model

// ContentItem 内容分析 — 爆款内容列表项
type ContentItem struct {
	ID         int64   `json:"id"`
	Title      string  `json:"title"`
	Form       string  `json:"form"`       // 教程/测评/创意短片/挑战赛
	Topic      string  `json:"topic"`      // 滑雪/潜水/骑行/旅行/Vlog
	Views      int64   `json:"views"`
	Engagement float64 `json:"engagement"` // 互动率 %
	IsHit      bool    `json:"isHit"`      // 是否爆款
}

// Competitor 市场洞察 — 竞品对比项
type Competitor struct {
	Name      string  `json:"name"`
	Category  string  `json:"category"`
	Buzz      int64   `json:"buzz"`      // 声量
	Growth    float64 `json:"growth"`    // 增速 %
	Sentiment float64 `json:"sentiment"` // 好感度 %
}

// PartnerBrand 品牌分析 — 合作品牌效果项
type PartnerBrand struct {
	Name       string  `json:"name"`
	Industry   string  `json:"industry"`
	Contents   int     `json:"contents"`   // 合作内容数
	Exposure   int64   `json:"exposure"`   // 曝光
	Engagement int64   `json:"engagement"` // 互动
	ROI        float64 `json:"roi"`        // 互动 ROI
}

// TagItem 品牌分析 — 高频词
type TagItem struct {
	Word   string  `json:"word"`
	Weight float64 `json:"weight"`
}
