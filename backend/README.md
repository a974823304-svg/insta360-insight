# Insta360 达人营销数据洞察平台 — 后端 (Go / Gin)

> 对应架构文档《影石技术栈.md》「三、后端服务层」「四、数据存储与基础设施」。

## 目录结构

```
backend/
├── main.go                       # 启动入口
├── go.mod / go.sum
├── internal/
│   ├── api/
│   │   ├── handler/              # HTTP 处理器 (瘦层,只做参数解析 + 响应包装)
│   │   │   ├── health.go
│   │   │   └── insight.go
│   │   └── router/               # 路由注册
│   │       └── router.go
│   ├── service/                  # 业务逻辑层
│   │   ├── insight_service.go    # 看板数据装配
│   │   └── ai_service.go         # AI 洞察(当前走规则,生产可走 ai/ Python)
│   ├── mock/                     # 演示用数据集(后续替换为 ClickHouse / Doris)
│   │   ├── data.go               # 通用工具 + 配色 + 数字格式化
│   │   ├── insight_data.go       # 各模块 Mock 数据
│   │   └── time.go
│   ├── model/                    # 共享数据结构 + JSON 契约
│   │   └── types.go
│   └── middleware/               # CORS / AccessLog
│       └── middleware.go
└── README.md
```

## 快速开始

```powershell
cd F:\workbuddy\影石\backend
go mod tidy
go run main.go
# 监听 http://localhost:8080
```

## API 列表

所有响应统一格式:`{ "code": 0, "data": ... }`。

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| GET | `/api/health` | 健康检查 |
| GET | `/api/filters/options` | 筛选面板可选项 |
| GET | `/api/kpi` | 5 张总览卡 |
| GET | `/api/views-trend` | 近 30 天播放量 + 上周期 + 异常点 |
| GET | `/api/platform-distribution` | 平台占比 |
| GET | `/api/track-performance` | 运动赛道表现 |
| GET | `/api/explosive-radar` | 引爆力雷达 |
| GET | `/api/audience-age` | 粉丝年龄画像 |
| GET | `/api/insights` | AI 关键洞察 |
| GET | `/api/top-creators` | 热门达人 TOP 10 |

所有数据接口都接受以下 querystring(可同时存在)：

- `date_range=2024-04-20&date_range=2024-05-20`
- `regions=北美&regions=欧洲`
- `tracks=冲浪&tracks=骑行`
- `platforms=抖音`
- `age_bands=18-24 岁`

## 与架构文档的对应

| 架构层 | 当前实现 | 生产替换 |
| --- | --- | --- |
| OLAP 数仓 (ClickHouse / Doris) | `internal/mock` 内存数据 | 物化视图 + 异步预聚合 |
| 聚合 API | `internal/service/insight_service.go` | 不变,只是把 mock 替换为 ch 查询 |
| AI 引擎 | `internal/service/ai_service.go` 走规则 | 调 `ai/` (Python FastAPI) `/v1/insights` |
| 鉴权 | 无 (Demo 阶段) | JWT 中间件 + Row-Level Security |

## 后续扩展点

1. **真实 OLAP**:`internal/mock/data.go` 的 `TrendGenerator` 等改为 ch-go 客户端
2. **缓存层**:核心 KPI 接口加 Redis,1-3 分钟过期(架构文档建议)
3. **行级权限**:在 `bindFilter()` 处追加 `tenant_id`
4. **gRPC 拆分**:统计引擎独立成 micro-service,与本服务通过 gRPC 通信
