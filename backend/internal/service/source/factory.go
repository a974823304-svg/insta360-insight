package source

import "fmt"

// NewDataSource 按 SOURCE 环境变量选择数据源。
//   mock(默认) -> MockAdapter(纯演示,无兜底装饰)
//   douyin/bilibili/xiaohongshu -> 对应真实 adapter,外层包 FallbackDataSource:
//     真实源返回 ErrNotImplemented(未接入 / 无凭证)或任意 error 时,自动回退 MockAdapter。
//   因此即使没填 appkey,看板也永远有数据;appkey 到位即可切真实源(看板零改动)。
func NewDataSource(kind string) (DataSource, error) {
	cfg := LoadConfig()
	mock := NewMockAdapter()
	switch kind {
	case "", "mock":
		return mock, nil
	case "douyin":
		return NewFallbackDataSource(NewDouyinAdapter(cfg.Douyin), mock), nil
	case "bilibili":
		return NewFallbackDataSource(NewBilibiliAdapter(cfg.Bilibili), mock), nil
	case "xiaohongshu":
		return NewFallbackDataSource(NewXiaohongshuAdapter(cfg.Xiaohongshu), mock), nil
	default:
		return nil, fmt.Errorf("unknown data source: %q", kind)
	}
}
