"""
ai.core.insight_engine
======================
洞察生成主流程: FeatureExtractor -> PromptBuilder -> LLMClient -> parse -> 兜底。

任何步骤失败(LLM 不可用 / JSON 解析失败 / 限流), 都会无缝切到 rule-based 输出,
确保调用方总能拿到 N 条 Insight。
"""
from __future__ import annotations

import time
from typing import List, Optional

from ai.core.feature_extractor import FeatureExtractor, Features
from ai.core.llm_client import LLMClient
from ai.core.prompt_builder import PromptBuilder
from ai.models.schemas import (
    Filter,
    Insight,
    InsightRequest,
    InsightResponse,
    Snapshot,
)


class InsightEngine:
    """
    主入口。 接受一个可选的 LLMClient, 没传则只用规则模式。
    """

    def __init__(self, llm: Optional[LLMClient] = None):
        self.llm = llm
        self.extractor = FeatureExtractor()
        self.builder = PromptBuilder()

    # ----------------------------------------------------------------
    # 公共 API
    # ----------------------------------------------------------------
    def generate(self, req: InsightRequest) -> InsightResponse:
        t0 = time.time()
        features = self.extractor.extract(req.snapshot)

        # 1) 尝试 LLM
        if self.llm is not None:
            try:
                items = self._generate_with_llm(features, req.filter, req.max_items)
                return InsightResponse(
                    items=items,
                    source=f"llm-{self._provider_name()}",
                    model=self.llm.cfg.model,
                    elapsed_ms=int((time.time() - t0) * 1000),
                )
            except Exception as e:  # noqa: BLE001
                # LLM 失败不致命, 回退到规则方案, 记到 stderr
                import sys
                print(f"[insight-engine] LLM 失败, 回退到规则方案: {e}", file=sys.stderr)

        # 2) 规则兜底
        items = self._generate_with_rules(features, req.max_items)
        return InsightResponse(
            items=items,
            source="rule",
            elapsed_ms=int((time.time() - t0) * 1000),
        )

    # ----------------------------------------------------------------
    # LLM 路径
    # ----------------------------------------------------------------
    def _generate_with_llm(self, features: Features, flt: Filter, n: int) -> List[Insight]:
        messages = self.builder.build_messages(features, flt, max_items=n)
        raw = self.llm.chat(messages)
        return self.builder.parse_response(raw)

    def _provider_name(self) -> str:
        return self.llm.cfg.provider if self.llm else "none"

    # ----------------------------------------------------------------
    # 规则路径
    # ----------------------------------------------------------------
    def _generate_with_rules(self, f: Features, n: int) -> List[Insight]:
        out: List[Insight] = []

        # 1) 总播放量增长
        if f.total_views_growth_pct >= 20:
            out.append(Insight(
                icon="surge",
                title=f"整体播放量增长 {f.total_views_growth_pct:+.1f}%",
                body=(
                    f"近一期总播放量较上期增长 {f.total_views_growth_pct:+.1f}%,"
                    f"高于历史均值, 建议持续加大投放力度。"
                ),
                severity="success",
            ))
        elif f.total_views_growth_pct < 0:
            out.append(Insight(
                icon="alert",
                title=f"整体播放量下滑 {f.total_views_growth_pct:+.1f}%",
                body=(
                    "整体播放量较上期出现负增长, 建议排查爆款流失原因, "
                    "并对高 ROI 达人增加排期。"
                ),
                severity="warning",
            ))

        # 2) 头部平台
        if f.top_platform and f.top_platform_share >= 35:
            out.append(Insight(
                icon="star",
                title=f"{f.top_platform} 占比突出",
                body=(
                    f"{f.top_platform} 平台贡献 {f.top_platform_share:.1f}% 播放量,"
                    f"建议在该平台加投垂类达人, 撬动更多自然流量。"
                ),
                severity="info",
            ))

        # 3) 头部赛道
        if f.top_track and abs(f.top_track_growth_pct) >= 10:
            sev = "success" if f.top_track_growth_pct > 0 else "warning"
            arrow = "增长" if f.top_track_growth_pct > 0 else "下降"
            out.append(Insight(
                icon="surge" if f.top_track_growth_pct > 0 else "alert",
                title=f"{f.top_track} 赛道{arrow} {f.top_track_growth_pct:+.1f}%",
                body=(
                    f"{f.top_track} 赛道近一期播放量较上期{arrow} {f.top_track_growth_pct:+.1f}%,"
                    f"建议复盘该赛道爆款选题。"
                ),
                severity=sev,
            ))

        # 4) 异常波动
        if f.anomaly_dates:
            out.append(Insight(
                icon="info",
                title=f"检测到 {len(f.anomaly_dates)} 个播放量异常日",
                body=(
                    f"在 {', '.join(f.anomaly_dates[:3])} 出现明显波动,"
                    f"建议结合选题/平台活动做归因分析。"
                ),
                severity="info",
            ))

        # 5) 黑名单提醒(运营安全)
        if f.blacklisted_creators:
            out.append(Insight(
                icon="alert",
                title="Top 列表中出现黑名单达人",
                body=(
                    f"{', '.join(f.blacklisted_creators)} 命中合作黑名单,"
                    f"请运营同学在排期前复核。"
                ),
                severity="warning",
            ))

        # 6) 上升达人
        if f.rising_creators:
            out.append(Insight(
                icon="star",
                title="潜力新锐达人浮现",
                body=(
                    f"近 30 天增速超过 20% 的达人: {', '.join(f.rising_creators[:3])},"
                    f"建议尽早建立合作以锁定优质档期。"
                ),
                severity="info",
            ))

        # 截断 / 兜底
        if not out:
            out.append(Insight(
                icon="info",
                title="数据平稳, 暂无显著异常",
                body="当前各核心指标运行平稳, 建议保持现有投放节奏并持续关注 AI 洞察。",
                severity="info",
            ))

        return out[:n]
