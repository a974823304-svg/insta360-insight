# 面试讲解清单（INTERVIEW CHECKLIST）

> 给候选人的「开箱即讲」清单。把项目讲成亮点，比堆功能更重要。
> 配套设计：[`superpowers/specs/2026-07-27-interview-positioning-design.md`](../superpowers/specs/2026-07-27-interview-positioning-design.md)
> 实现计划：[`superpowers/plans/2026-07-27-interview-implementation-plan.md`](../superpowers/plans/2026-07-27-interview-implementation-plan.md)

---

## 一、必讲的故事线（5 条，建议背熟）

1. **可插拔的数据接入层**
   - 定义 `DataSource` 接口，Mock 与真实平台（抖音 / B站 / 小红书）适配器实现同一契约。
   - 凭证缺失自动回退 Mock —— 体现接口抽象 + 容错设计，而不是写死的假数据。

2. **真实对接抖音，且可验证**
   - 适配器按真实 API 契约做字段映射（`client_token` 公开数据 / 用户授权拿自身数据两种模式）。
   - 用 `httptest` 模拟平台响应做映射测试，**不需要任何凭证就能 `go test` 证明「不是假接口」**。

3. **前端降级优雅**
   - 后端不可达时前端用 `fallback` / `demo-real` 数据，绝不会出现白屏 —— 体现工程健壮性。

4. **全栈分层清晰**
   - Vue3 + Go(Gin) + Python(FastAPI) 三层各司其职；JWT 鉴权、确定性筛选变换、ECharts 可视化、AI 洞察兜底。

5. **诚实边界（加分项，主动说）**
   - 个人号拿不到「全网达人营销汇总」（企业资质 / 消耗门槛），但架构已为真实数据预留接入位。
   - 这反而显出你对平台生态、成本、合规的理解 —— 是加分不是减分。

---

## 二、演示自检（打开页面看什么）

- [ ] 打开 GitHub Pages 演示链接，看板正常渲染（KPI / 趋势 / 平台分布 / 赛道 / 雷达 / 粉丝画像 / 达人表）。
- [ ] 左侧筛选面板切换平台 / 赛道 / 时间，看图表联动。
- [ ] 趋势图有「当期 vs 上周期」对比与 AI 异常点标记。
- [ ] 切换深色主题一致、无溢出 / 白屏。
- [ ] 右上角「登录」按钮 → 登录后变头像 + 登出下拉（看板未登录也能看）。

---

## 三、代码 / 测试自检（打开仓库看什么）

- [ ] 后端 `go build ./...` 零错误。
- [ ] 后端 `go test ./...` 全绿，重点亮 `real_insight_test.go`（httptest 证明抖音映射正确）。
- [ ] `backend/internal/service/source/` 下 `adapter.go`(接口) / `factory.go`(接线) / `fallback.go`(兜底) / `douyin_adapter.go`(真实映射) / `oauth.go`(token 抽象)。
- [ ] `backend/.env.example` 展示凭证可插拔；`docs/真实数据接入指南.md` 解释 OAuth2 流程与切换方式。
- [ ] 前端 `demo-real-data.js` 形状 = 后端真实映射输出（诚实说明是合成演示数据）。

---

## 四、可能被追问 & 应答要点

| 追问 | 建议应答 |
| --- | --- |
| 这数据是真的吗？ | 演示数据是真实映射输出同构的合成数据；有凭证时同一套代码经后端适配器返回实时真数据，且 `httptest` 已证明映射层对接了真实平台结构。 |
| 为什么不用实时真数据上线？ | 个人身份拿不到全网汇总（企业门槛），且 GitHub Pages 只托管前端省钱；真实后端本地可跑，架构已就绪。 |
| 多个平台数据怎么统一？ | 统一 `DataSource` 接口 + 字段映射层，前端只认业务 schema，不关心来源。 |
| 接入失败怎么办？ | `FallbackDataSource` 装饰器捕获 `ErrNotImplemented` / error 自动回退 Mock，看板不崩。 |
| 鉴权怎么做的？ | JWT(HS256) + SQLite(纯 Go) + bcrypt；dev 可免登录，看板公开、登录可选。 |
| 为什么选 Mock 兜底而不是报错？ | 演示 / 容错优先；真实环境可改为告警或灰度。 |

---

## 五、下一步可扩展（体现规划力）

- [ ] 方案 B：打通本人抖音 OAuth 授权码流程，本地看自己真实数据。
- [ ] 补全达人 / 内容 / 市场 / 品牌四域的真实映射（当前仅洞察域写实）。
- [ ] 接 ClickHouse / Doris 替换 Mock，做物化视图预聚合。
