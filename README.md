# Insta360 达人营销数据洞察平台

> 面向**全栈开发岗位面试**的数据可视化分析平台 Demo。
> 定位：用一套**可插拔的数据接入层**，把「Mock 演示数据」与「真实平台（抖音 / B站 / 小红书）API」统一抽象，凭证缺失自动降级，凭证到位无需改代码即可切换真实数据。

[在线演示（GitHub Pages，上线后生效）](https://a974823304-syg.github.io/insta360-insight/) · [架构说明](./docs/ARCHITECTURE.md) · [面试讲解清单](./docs/INTERVIEW_CHECKLIST.md) · [真实数据接入指南](./docs/真实数据接入指南.md)

---

## 1. 技术栈

| 层 | 技术 | 说明 |
| --- | --- | --- |
| 前端 | Vue 3 + Vite + Element Plus + Pinia + Axios + ECharts（暗色主题） | GitHub Pages 静态托管，后端不可达时优雅降级 |
| 后端 | Go (Gin) | 可插拔 `DataSource` 适配器 + JWT 鉴权 + SQLite |
| AI 引擎 | Python (FastAPI + dashscope / openai) | 规则兜底 + 可接大模型生成洞察 |
| 数据存储 | Mock 内存数据；生产候选 OLAP：ClickHouse / Doris | 用户存储：SQLite（纯 Go，无 CGO） |

---

## 2. 架构总览

```
                         ┌─────────────────────────────────────┐
                         │           前端 Vue3 + ECharts         │
                         │  看板 / 筛选联动 / 后端不可达自动降级    │
                         └───────────────┬─────────────────────┘
                                         │  /api/*  (统一 {code,data} 信封)
                         ┌───────────────▼─────────────────────┐
                         │          Go 后端 (Gin)               │
                         │  ┌───────────────────────────────┐  │
                         │  │  service.DataSource 接口        │  │
                         │  │   ├─ MockAdapter（演示数据）      │  │
                         │  │   ├─ DouyinAdapter  ──┐          │  │
                         │  │   ├─ BilibiliAdapter ─┤ fallback │  │
                         │  │   └─ XiaohongshuAdapter┘ 装饰器   │  │
                         │  │      ↓ 真实方法 ErrNotImplemented │  │
                         │  │      ↓ 或任意 error → 自动回退 Mock│ │
                         │  └───────────────────────────────┘  │
                         │  JWT 鉴权 · SQLite 用户 · CORS        │
                         └───────────────┬─────────────────────┘
                                         │  /v1/insights
                         ┌───────────────▼─────────────────────┐
                         │     Python AI 引擎 (FastAPI)          │
                         │  特征提取 → Prompt → LLM → 规则兜底    │
                         └─────────────────────────────────────┘
```

**核心设计点**：定义 `DataSource` 接口，Mock 与三平台真实适配器实现同一契约；外层 `FallbackDataSource` 装饰器在真实方法返回 `ErrNotImplemented`（无凭证 / 未实现）或任意错误时，自动回退 Mock —— 看板永远有数据、不崩。

---

## 3. 真实数据接入（项目最大亮点）

真实平台「全网达人营销汇总数据」被企业资质 + 消耗门槛锁死（小红书蒲公英需近一年消耗 > 500 万、抖音星图需蓝 V + 保证金），个人身份基本拿不到。本项目因此采用**「代码就绪、凭证可插拔」**策略：

- **抖音洞察域已写实**：`Kpi` + `ViewsTrend` 按真实 API 响应形状做字段映射，支持两种免费 token 模式：
  - `client_token`（应用级公开 token，个人可申）：拉公开发现类数据；
  - 用户授权 `access_token`：拉**你自己账号**的真实粉丝 / 播放 / 互动数据。
- **测试即证据**：`backend/internal/service/source/real_insight_test.go` 用 `httptest` 起「假抖音服务器」，喂**真实形状的响应 JSON**，断言适配器映射正确。**无需任何凭证即可 `go test` 验证「适配器真对接了真实平台结构，不是假接口」**。
- **前端演示诚实**：GitHub Pages 静态站用 `demo-real-data.js`（形状等价于真实映射输出）演示；README 与代码注释都写明：「有凭证时同一套代码经后端适配器返回的就是实时真数据」。
- **切换零代码**：拿到任意凭证（哪怕是自己的创作者 token），填进 `backend/.env`（参考 `backend/.env.example`），改一行 `SOURCE` 即可切真实数据。

详见 [真实数据接入指南](./docs/真实数据接入指南.md)。

---

## 4. 在线演示（GitHub Pages）

前端构建产物已配置 `base: './'`，纯静态托管即可运行，后端不可达时自动降级为 `demo-real` 数据，**不会出现白屏**。

- 预期地址：`https://a974823304-syg.github.io/insta360-insight/`（部署后生效，见 [ARCHITECTURE.md](./docs/ARCHITECTURE.md) 部署章节）
- 真实后端（Go + Python）仅在本地运行演示，不部署公网（省钱）。

---

## 5. 本地全栈运行（可选，用于看真实后端 / 跑测试）

完整启动需要 3 个终端。Go 工具链本机已配置 `GOPROXY=goproxy.cn`，`go build` / `go test` 可直接跑通。

### 终端 1：后端（Go，:8080）
```bash
cd backend
go mod tidy
go run main.go
```

### 终端 2：AI 引擎（Python，:9000，可选）
```bash
cd .
python -m venv .venv && .venv/Scripts/activate
pip install -r ai/requirements.txt
python -m ai.main
```

### 终端 3：前端（Vue，:5173）
```bash
cd frontend
npm install
npm run dev
# 打开 http://localhost:5173
```

### 跑后端测试（面试硬通货）
```bash
cd backend
go test ./...
# 含 real_insight_test.go：httptest 模拟抖音真实响应，证明映射层正确
```

---

## 6. 核心 API 一览

| Method | Path | 用途 |
| --- | --- | --- |
| GET | `/api/health` | 健康检查 |
| GET | `/api/filters/options` | 筛选面板可选项 |
| GET | `/api/kpi` | 总览 KPI 卡 |
| GET | `/api/views-trend` | 播放趋势 + 环比 + 异常点 |
| GET | `/api/platform-distribution` | 平台占比 |
| GET | `/api/track-performance` | 赛道表现 |
| GET | `/api/explosive-radar` | 引爆力雷达 |
| GET | `/api/audience-age` | 粉丝年龄画像 |
| GET | `/api/insights` | AI 关键洞察 |
| GET | `/api/top-creators` | 热门达人 TOP |
| GET | `/api/{content,market,brand}/*` | 内容 / 市场 / 品牌分析域 |

所有 `/api/*` 支持多选筛选：`date_range`、`regions`、`tracks`、`platforms`、`age_bands`。

---

## 7. 已知边界（诚实说明）

- 个人账号 **live 拉不到**「全网达人营销汇总」（企业资质 / 消耗门槛）；本项目真实域聚焦「你自己账号的真实数据」+ 公开数据，架构已为汇总数据预留接入位。
- 三平台响应字段名基于公开 / 文档结构的合理实现，上线前以官方最新文档为准。
- 前端演示数据为「真实映射输出同构」的合成数据，非实时真数据（实时真数据需上线后端 + 凭证）。

---

## 8. 目录约定

```
backend/internal/
  api/{handler,router}   # HTTP 层（瘦）
  service/source/        # DataSource 接口 + Mock/三平台 adapter + fallback + 测试
  mock/                  # 演示数据集（确定性）
  model/                 # 数据结构 + JSON 契约
  middleware/            # CORS / JWT / AccessLog
frontend/src/
  api/                   # axios 封装 + insight API + fallback-data / demo-real-data
  components/            # 图表 / 表格 / 筛选组件
  views/                 # 页面（洞察 / 内容 / 市场 / 品牌 / 登录 / 资料）
  stores/                # Pinia（filter / insight / user）
ai/                      # Python FastAPI 智能洞察引擎
docs/                    # 架构 / 设计 spec / 计划 / 面试清单
```

---

## 9. AI 代理交接

已为 AI 编码工具（OpenAI Codex / Claude Code / Cursor / GitHub Copilot）准备交接手册：根目录 `AGENTS.md` / `CODEX.md`（内容一致）。接手前请先阅读，严格遵守其中的「铁律」与「已知 bug 修复」。
