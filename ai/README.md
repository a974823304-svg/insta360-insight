# Insta360 达人营销数据洞察平台 — AI 智能洞察引擎 (Python)

> 对应架构文档《影石技术栈.md》「六、AI 智能洞察引擎」。

## 目录结构

```
ai/
├── __init__.py
├── main.py                  # FastAPI 入口
├── requirements.txt
├── README.md
├── api/
│   ├── __init__.py
│   └── routes.py            # /v1/insights, /v1/health
├── core/
│   ├── __init__.py
│   ├── insight_engine.py    # 主流程 (Feature -> Prompt -> LLM -> 兜底)
│   ├── feature_extractor.py # 特征提取
│   ├── prompt_builder.py    # Prompt 拼装 + JSON 解析
│   └── llm_client.py        # 通义千问 / OpenAI 统一抽象
└── models/
    ├── __init__.py
    └── schemas.py           # Pydantic 契约
```

## 快速开始

在项目根目录 `F:\workbuddy\影石` 下执行:

```powershell
# 1) 装依赖
python -m venv .venv
.venv\Scripts\activate
pip install -r ai/requirements.txt

# 2) 启动服务(无 LLM 也能跑, 走规则兜底)
python -m ai.main
# 监听 http://localhost:9000
```

## 启用通义千问

```powershell
$env:DASHSCOPE_API_KEY = "sk-xxxx"
$env:INSTINSIGHT_LLM_MODEL = "qwen-plus"   # 可选, 默认 qwen-plus
python -m ai.main
```

## 启用 OpenAI / 兼容服务

```powershell
$env:INSTINSIGHT_LLM_PROVIDER = "openai"
$env:OPENAI_API_KEY = "sk-xxxx"
$env:OPENAI_BASE_URL = "https://api.openai.com/v1"   # 可选
$env:INSTINSIGHT_LLM_MODEL = "gpt-4o-mini"            # 可选
python -m ai.main
```

## 接口示例

```bash
curl -X POST http://localhost:9000/v1/insights \
  -H "Content-Type: application/json" \
  -d '{
    "filter": { "tracks": ["冲浪"] },
    "snapshot": {
      "kpi": [{"key":"views","label":"总播放量","value":"2.38B","raw":2380000000,"delta_pct":24.7,"delta_up":true,"unit":"","description":"较上期"}],
      "views_trend": [],
      "platform_share": [{"platform":"YouTube","share":45.6,"views":1085000000,"color":"#FF3D5A"}],
      "track_performance": [{"track":"冲浪","views":642000000,"color":"#3DD9EB"}],
      "explosive_radar": [],
      "audience_age": [],
      "top_creators": []
    },
    "max_items": 3
  }'
```

返回:

```json
{
  "items": [
    {"icon": "surge", "title": "...", "body": "...", "severity": "success"}
  ],
  "source": "rule",
  "model": null,
  "elapsed_ms": 3
}
```

## 关键设计

1. **LLM 失败可降级**:`InsightEngine.generate()` 任何异常(网络/解析/限流)都会
   切到 rule-based 兜底, 调用方总能拿到 N 条 Insight。
2. **职责分层**:FeatureExtractor / PromptBuilder / LLMClient 全部独立,
   方便单测和替换模型(未来要接 Qwen-VL / Function Call 都不用改 engine)。
3. **契约对齐**:Pydantic Schema 的字段与 Go 后端 `internal/model/types.go`
   一一对应, JSON 互转无歧义。
4. **服务化**:`/v1/insights` 是 HTTP 接口, Go 后端通过 HTTP 调用即可,
   也可让前端直接调用(生产建议走后端聚合, 控制 QPS)。

## 与 Go 后端的对接方式

`backend/internal/service/ai_service.go` 当前走本地规则, 生产替换为:

```go
// service/ai_service.go (生产版)
func (s *AIService) Generate(f model.Filter) []model.Insight {
    payload := map[string]interface{}{"filter": f, "snapshot": s.snapshot(f)}
    body, _ := json.Marshal(payload)
    resp, err := http.Post("http://ai-service:9000/v1/insights", "application/json", bytes.NewReader(body))
    if err != nil { return ruleFallback(f) }
    var out struct{ Items []model.Insight `json:"items"` }
    json.NewDecoder(resp.Body).Decode(&out)
    return out.Items
}
```
