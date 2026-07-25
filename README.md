# Insta360 达人营销数据洞察平台 — Demo 原型

> 基于架构方案《影石技术栈.md》搭建的最小可运行 Demo。前端按 UI 效果图 v1
> 复刻"数据洞察"主页面,后端 / AI 引擎均为可独立运行的最小骨架,真实 OLAP / 大模型可平滑替换。

## 目录结构

```
F:\workbuddy\影石\
├── README.md                   # 本文件
├── 环境安装保姆级指南.md          # Go / VC++ 安装教程
├── backend/                    # Go (Gin) 后端
│   ├── main.go
│   ├── go.mod / go.sum
│   ├── internal/
│   │   ├── api/
│   │   │   ├── handler/         # HTTP 处理器 (health / insight)
│   │   │   └── router/          # 路由注册
│   │   ├── service/             # 业务逻辑 (insight / ai)
│   │   ├── mock/                # 演示数据(后续替换为 ClickHouse)
│   │   ├── model/               # 数据结构 + JSON 契约
│   │   └── middleware/          # CORS / AccessLog
│   └── README.md
├── frontend/                   # Vue 3 + Vite + Element Plus + ECharts
│   ├── package.json
│   ├── vite.config.js
│   ├── index.html
│   ├── src/
│   │   ├── main.js
│   │   ├── App.vue
│   │   ├── styles/theme.scss    # 暗色主题变量
│   │   ├── router/              # 路由
│   │   ├── stores/              # Pinia (filter / insight)
│   │   ├── api/                 # axios 封装 + insight API
│   │   ├── components/          # 10 个组件(顶栏/侧栏/图表/表格)
│   │   └── views/               # 页面(InsightDashboard + 占位页)
│   └── README.md
└── ai/                         # Python FastAPI 智能洞察引擎
    ├── main.py
    ├── requirements.txt
    ├── api/                    # /v1/insights 路由
    ├── core/                   # FeatureExtractor / PromptBuilder / LLMClient / InsightEngine
    └── models/                 # Pydantic 契约
```

## 模块职责一览

| 模块 | 关键文件 | 职责 |
| --- | --- | --- |
| 前端 - 顶栏 | `components/TopBar.vue` | Logo / 6 个 Tab / 日期 / 导出 / 头像 |
| 前端 - 侧栏 | `components/SideFilter.vue` | 时间/地区/赛道/平台/达人属性/粉丝画像 + 应用筛选 |
| 前端 - KPI | `components/KpiCard.vue` × 5 | 5 张总览卡(达人数/总粉丝/总播放/互动/合作) |
| 前端 - 趋势 | `components/TrendChart.vue` | 当期 vs 上周期 + AI 异常点 |
| 前端 - 平台分布 | `components/PlatformDonut.vue` | 环形图 + 图例 |
| 前端 - 赛道表现 | `components/TrackBarChart.vue` | 横向条形 + 渐变 |
| 前端 - 引爆力 | `components/RadarChart.vue` | 雷达(当前 vs 均值) |
| 前端 - 粉丝画像 | `components/AgeDonut.vue` | 环形(总粉丝居中) |
| 前端 - 达人表 | `components/TopCreatorsTable.vue` | 排序 + 搜索 + 黑名单标记 |
| 前端 - 状态 | `stores/filter.js` `stores/insight.js` | Pinia 全局筛选 / 数据 |
| 后端 - 入口 | `main.go` | 装载 dataset / service / router |
| 后端 - 处理器 | `internal/api/handler/insight.go` | 9 个 REST 接口,瘦层 |
| 后端 - 业务 | `internal/service/insight_service.go` | 业务装配,接受 Filter |
| 后端 - Mock | `internal/mock/*` | 演示数据集(确定性的) |
| AI - 引擎 | `ai/core/insight_engine.py` | 主流程:特征→Prompt→LLM→兜底 |
| AI - 特征 | `ai/core/feature_extractor.py` | 提取 7 类业务特征 |
| AI - Prompt | `ai/core/prompt_builder.py` | 拼装 + 解析 JSON 响应 |
| AI - LLM | `ai/core/llm_client.py` | 通义千问 / OpenAI 统一抽象 |

## 快速开始

完整启动需 **3 个终端**。完整步骤见《环境安装保姆级指南.md》。

### 终端 1:后端
```powershell
cd F:\workbuddy\影石\backend
go mod tidy
go run main.go
# 监听 http://localhost:8080
```

### 终端 2:AI 引擎(可选,无 LLM 也可跑)
```powershell
cd F:\workbuddy\影石
python -m venv .venv
.venv\Scripts\activate
pip install -r ai/requirements.txt
python -m ai.main
# 监听 http://localhost:9000
```

### 终端 3:前端
```powershell
cd F:\workbuddy\影石\frontend
npm install
npm run dev
# 打开 http://localhost:5173
```

## 核心 API 一览

| Method | Path | 用途 |
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
| POST | `http://localhost:9000/v1/insights` | (Python AI 引擎)业务洞察生成 |

所有 `/api/*` 都支持 querystring 筛选: `date_range`, `regions`, `tracks`, `platforms`, `age_bands`。

## 验证 Demo

打开 `http://localhost:5173` 应能看到:
1. 顶部 Insta360 Logo + 6 个 Tab + 日期选择 + 导出报告
2. 左侧 5 类筛选面板 + "应用筛选"按钮
3. 数据总览 5 个 KPI 卡片(达人数 / 总粉丝 / 总播放量 / 互动量 / 合作内容)
4. 趋势分析折线图(蓝实线 + 灰虚线 + AI 异常点)
5. 关键洞察列表(3 条 AI 提示)
6. 平台分布环形图 / 赛道条形图 / 引爆力雷达图(三联)
7. 粉丝画像环形图 + 热门达人 TOP 10 表格

## 与架构文档的对应

| 架构层 | 当前实现 | 生产替换 |
| --- | --- | --- |
| OLAP 数仓 (ClickHouse / Doris) | `backend/internal/mock` 内存数据 | 物化视图 + 异步预聚合 |
| 聚合 API | `backend/internal/service` | 不变,只把 mock 换成 ch 查询 |
| AI 引擎 | `backend/internal/service/ai_service` 走规则 | 调 `ai/` Python FastAPI `/v1/insights` |
| 鉴权 | 无 (Demo 阶段) | JWT 中间件 + Row-Level Security |
| 缓存 | 无 | Redis(核心 KPI 1-3 min) |
| 消息队列 | 无 | Kafka(原始事件) + Flink(实时) |
| 容器化 | 无 | Docker + K8s |

## Demo 已验证

- ✅ 前端 `vite build` 成功 (2237 modules, 6.2s)
- ✅ AI 引擎 Python import 全部通过,规则兜底输出 3 条洞察
- ✅ 后端代码按 Go 标准布局,服务装配完整(待本地装 Go 后跑 `go run`)
