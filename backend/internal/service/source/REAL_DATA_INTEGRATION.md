# 真实数据接入层 · 设计规格（2026-07-26）

> 本文档记录「接入真实数据」阶段的后端接入层改造方案。配套用户向说明见 `docs/真实数据接入指南.md`。

## 1. 目标与边界

- **目标**：把 `DataSource` 接口从「全 Mock」升级为「可插拔真实平台接入」，让抖音 / B站 / 小红书
  的真实 API 能在凭证到位时切换启用，**凭证缺失或调用失败时不崩溃、自动回退演示数据**。
- **本期范围（聚焦可验证价值）**：
  - 落地通用接入基础设施：env 配置、OAuth2/Token 抽象、Fallback 装饰器。
  - 三平台 **洞察域（主页 9 个卡片/图表）** 完成真实 HTTP 调用 + 字段映射范例。
  - 其余域（达人 / 内容 / 市场 / 品牌）每个平台方法保留 `ErrNotImplemented`，由兜底接管。
- **现实约束（必须告知）**：
  - 抖音 / B站 / 小红书 的**达人营销汇总数据**都在企业资质接口（星图 / 蒲公英），公开接口只能取
    授权账号自身的公开资料。本期代码基于公开文档 + 星图/蒲公英文档的**合理响应结构**实现映射，
    字段名需按平台最新文档核对。
  - 小红书蒲公英官方门槛极高（品牌年消耗 500W+ 或白名单），普通开发者基本拿不到；本期同时支持
    第三方聚合网关（如 Just One API），把 API key 填 `XHS_ACCESS_TOKEN` 即可。
  - 因无凭证、平台门槛高，本期**无法端到端真实验证**；验证靠 httptest 单测覆盖字段映射与回退逻辑。

## 2. 架构

```
main.go
  └─ source.NewDataSource(SOURCE)            // factory
        └─ FallbackDataSource{ real, mock }   // 装饰器:real 失败→mock
              ├─ real = DouyinAdapter | BilibiliAdapter | XiaohongshuAdapter
              └─ mock = MockAdapter           // 永远有数据
```

- `factory.go`：`NewDataSource` 读取 `LoadConfig()`，构造真实 adapter 并包一层 `FallbackDataSource`。
  `SOURCE=mock` 或空 → 直接返回 `MockAdapter`（无装饰）。
- `config.go`：`PlatformConfig` 从环境变量读取三平台凭证（`*.CLIENT_KEY/SECRET/ACCESS_TOKEN/USER_ID/BASE_URL`）。
- `oauth.go`：
  - `TokenProvider` 接口：`Token(ctx) (string, error)`。
  - `StaticToken`：直接用配置里的 `ACCESS_TOKEN`（最简单，推荐先用）。
  - `ClientToken`：抖音 `client_token` 流程（GET `/oauth/client_token/`，应用级、无需用户授权），带内存缓存。
  - `resolveTokenProvider`：按优先级解析（AccessToken → ClientToken(仅抖音) → nil）。
- `fallback.go`：`FallbackDataSource` 实现全部 `DataSource` 方法，统一用 `tryCall` 包裹：
  `real` 成功返回；`real` 返回 `ErrNotImplemented` 或任意 error → 回退 `mock`。
- 三平台 adapter：洞察域方法做真实调用 + 映射；其余域方法 `return nil, ErrNotImplemented`。

## 3. 各平台接入前提

| 平台 | 应用级 token | 用户/商家授权 | 公开可拉数据 | 本期真实映射方法 |
|------|------------|--------------|------------|----------------|
| 抖音 | client_token（无需用户） | 可选 | 授权账号公开资料/视频；汇总在星图(企业) | Kpi / ViewsTrend / PlatformShare / TopCreators |
| B站 | 无 | 必须（access_token） | 稿件增量/按天数据（需 v2 签名） | Kpi / ViewsTrend / PlatformShare |
| 小红书 | 无（蒲公英需商家授权） | 必须 | 蒲公英核心指标（高门槛）/ 聚合可用 | Kpi / ViewsTrend / PlatformShare |

## 4. 字段映射约定

- adapter 内部定义「平台响应结构（示意）」→ 映射到 `model` 包结构。
- 大数展示用 `humanize/unitOf`（万 / 亿 / M / B）。
- 每个真实方法：`requireToken` → 无凭证返回 `ErrNotImplemented` → 构造请求（带 token / 签名）→
  解析 JSON → `error_code != 0` 或解析失败返回 `ErrNotImplemented` → 映射返回。
- **所有平台响应字段名均标注「需按最新文档核对」**，拿到真实 appkey 后首要核对项是字段名与单位。

## 5. 验证

- `go build ./...`：0 错误。
- `go test ./...`：全绿，含：
  - `TestFallbackReturnsMockWhenNoCredential`：无凭证时看板回退 mock。
  - `TestFactoryFallbackWiring`：`SOURCE=douyin` 返回带兜底的 DataSource，无凭证返回 mock 数据。
  - `TestDouyinKpiMapping` / `TestBilibiliSignAndMapping` / `TestXiaohongshuKpiMapping`：
    用 httptest 验证字段映射与 B站 v2 签名头。

## 6. 后续工作（由凭证/资质到位触发）

1. 拿企业 appkey + scope 后，核对并补全各平台「星图/蒲公英」真实字段名。
2. 补全达人 / 内容 / 市场 / 品牌四域的真实映射方法（目前回退 mock）。
3. 实现 OAuth2 授权码回调页（用户/商家登录授权 → 落库 token → 注入 adapter），支持无人值守拉数。
4. 真实数据落 OLAP（ClickHouse/Doris）做行级裁剪与缓存，降低平台 API 调用频率与限流风险。
