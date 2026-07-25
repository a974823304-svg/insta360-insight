// Package model 定义整个后端共享的数据结构。
//
// 这些结构同时是 JSON 序列化的契约,前端通过 src/api/insight.js 引用同名字段。
// 当后续接入 ClickHouse / Doris 时,只需在 service 层把查询结果映射到这些结构。
package model

// ============================================================
// 数据洞察 (Insight Dashboard) 相关模型
// ============================================================

// Filter 通用筛选条件。所有聚合接口都接受这套条件。
// 对应架构文档:不同客户 (Row-Level Security) 需在 service 层做权限裁剪。
type Filter struct {
	DateRange []string `json:"date_range" form:"date_range"` // [start, end] YYYY-MM-DD
	Regions   []string `json:"regions" form:"regions"`       // 北美 / 欧洲 / 亚太 / 全球
	Tracks    []string `json:"tracks" form:"tracks"`         // 滑雪 / 冲浪 / 骑行 / 潜水 / 攀岩
	Platforms []string `json:"platforms" form:"platforms"`   // 抖音 / B站 / 小红书
	AgeBands  []string `json:"age_bands" form:"age_bands"`   // 18-24 / 25-34 / 35-44 / 45+
}

// KpiCard 数据总览卡片
type KpiCard struct {
	Key         string  `json:"key"`          // 唯一 key: creators / followers / views / engagement / collabs
	Label       string  `json:"label"`        // 中文标签
	Value       string  `json:"value"`        // 展示用 (e.g. "2.38B")
	Raw         float64 `json:"raw"`          // 原始数值,供前端做动效 / 排序
	DeltaPct    float64 `json:"delta_pct"`    // 环比 / 同比 %
	DeltaUp     bool    `json:"delta_up"`     // 上升为 true, 下降为 false
	Unit        string  `json:"unit"`         // "" / "%" / "M" / "B"
	Description string  `json:"description"`  // 副标题
}

// ViewsTrendPoint 单日播放量趋势点
type ViewsTrendPoint struct {
	Date       string  `json:"date"`        // MM-DD
	Views      int64   `json:"views"`       // 当期播放量
	PrevViews  int64   `json:"prev_views"`  // 上周期值,用于画虚线对比
	HasAnomaly bool    `json:"has_anomaly"` // 是否触发 AI 洞察标记
	AnomalyTag string  `json:"anomaly_tag"` // 异常原因,UI 在该点高亮
	Ratio      float64 `json:"ratio"`       // 增长率 (用于 tooltip)
}

// PlatformShare 平台分布
type PlatformShare struct {
	Platform string  `json:"platform"`
	Share    float64 `json:"share"`    // 占比 %
	Views    int64   `json:"views"`    // 原始播放量
	Color    string  `json:"color"`    // UI 配色
}

// TrackPerformance 运动赛道表现
type TrackPerformance struct {
	Track string `json:"track"` // 滑雪 / 冲浪 ...
	Views int64  `json:"views"`
	Color string `json:"color"`
}

// RadarMetric 引爆力维度
type RadarMetric struct {
	Dimension string  `json:"dimension"` // 内容质量 / 粉丝互动力 / 商业配合力 / 成活性
	Value     float64 `json:"value"`     // 0~100
	Avg       float64 `json:"avg"`       // 全平台均值,用于画虚线
}

// AgeShare 粉丝年龄分布
type AgeShare struct {
	Bucket string  `json:"bucket"` // 18-24 / 25-34 / 35-44 / 45+
	Share  float64 `json:"share"`  // 占比 %
	Color  string  `json:"color"`
}

// Insight AI 洞察条目
type Insight struct {
	Icon     string `json:"icon"`     // 图标标识: surge / alert / star
	Title    string `json:"title"`    // 标题
	Body     string `json:"body"`     // 业务可读正文
	Severity string `json:"severity"` // info / warning / success
}

// TopCreator 热门达人
type TopCreator struct {
	Rank        int     `json:"rank"`
	Avatar      string  `json:"avatar"`       // emoji / 头像 URL
	Name        string  `json:"name"`
	Platform    string  `json:"platform"`
	Followers   int64   `json:"followers"`    // 粉丝数
	TotalViews  int64   `json:"total_views"`  // 播放量
	Engagement  float64 `json:"engagement"`   // 互动率 %
	Growth30d   float64 `json:"growth_30d"`   // 近 30 天增长率 %
	Blacklist   bool    `json:"blacklist"`    // 黑名单标记
	Explosive   float64 `json:"explosive"`    // 引爆力评分
	Tags        []string `json:"tags"`        // #骑行 #极限 #摄影 ...
}

// GenderShare 粉丝性别分布
type GenderShare struct {
	Gender string  `json:"gender"` // 男 / 女
	Share  float64 `json:"share"`  // 占比 %
	Color  string  `json:"color"`
}

// Audience 达人粉丝画像(年龄 + 性别), 供 /api/creator/audience 返回
type Audience struct {
	Age    []AgeShare     `json:"age"`
	Gender []GenderShare  `json:"gender"`
}

// FilterOptions 筛选面板的可选项(用于渲染 SideFilter)
type FilterOptions struct {
	Regions   []Option `json:"regions"`
	Tracks    []Option `json:"tracks"`
	Platforms []Option `json:"platforms"`
	AgeBands  []Option `json:"age_bands"`
	Presets   []Option `json:"presets"` // 时间范围快捷: 近 7 天 / 近 30 天 ...
}

// Option 通用下拉项
type Option struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

// ============================================================
// 响应包装
// ============================================================

// APIResponse 统一响应结构
//   - Code: 0 表示成功,非 0 为业务错误码
//   - Message: 错误时填,成功时可空
//   - Data: 业务负载,任意类型
type APIResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data"`
}

// OK 构造成功响应
func OK(data interface{}) APIResponse {
	return APIResponse{Code: 0, Data: data}
}

// Fail 构造业务错误响应。注意:鉴权类错误建议同时返回 HTTP 401（见 middleware）。
func Fail(code int, message string) APIResponse {
	return APIResponse{Code: code, Message: message}
}
