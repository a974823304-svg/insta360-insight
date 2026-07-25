package service

import (
	"context"

	"insta360-insight/internal/model"
	"insta360-insight/internal/service/source"
)

// ============================================================
// AIService AI 洞察服务
// ============================================================
//
// 当前实现:基于 mock 的本地兜底洞察。
// 数据源经 source.DataSource 注入,切换 SOURCE 时洞察也跟随源。
// 生产环境:把 Generate() 改成 HTTP 调用 ai/ (Python FastAPI) 的 /v1/insights。
type AIService struct {
	src source.DataSource
}

func NewAIService(src source.DataSource) *AIService { return &AIService{src: src} }

// Generate 根据当前数据源返回洞察。
//   真实实现:POST http://ai-service/v1/insights,body 是 Filter + 当前 KPI/趋势快照
func (s *AIService) Generate(ctx context.Context, f model.Filter) ([]model.Insight, error) {
	return s.src.Insights(ctx, f)
}
