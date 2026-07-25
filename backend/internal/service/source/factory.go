package source

import "fmt"

// NewDataSource 按 SOURCE 环境变量选择数据源。
//   mock(默认) -> MockAdapter; douyin/bilibili/xiaohongshu -> 对应空壳(阶段三填充)
func NewDataSource(kind string) (DataSource, error) {
	switch kind {
	case "", "mock":
		return NewMockAdapter(), nil
	case "douyin":
		return &DouyinAdapter{}, nil
	case "bilibili":
		return &BilibiliAdapter{}, nil
	case "xiaohongshu":
		return &XiaohongshuAdapter{}, nil
	default:
		return nil, fmt.Errorf("unknown data source: %q", kind)
	}
}
