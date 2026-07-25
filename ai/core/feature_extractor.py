"""
ai.core.feature_extractor
=========================
从看板快照中提取"对业务判断有用的"统计特征。

为什么要单独做特征提取?
- LLM 不擅长看一长串 JSON 算百分比, 先在代码层把核心信号算出来,
  prompt 里只喂"已经提取过的事实", 模型可以专注于"用业务语言描述"。
- 同时特征也作为规则方案的输入, 保证 LLM 兜底 / LLM 在线时输出风格一致。
"""
from __future__ import annotations

from dataclasses import dataclass
from typing import List, Optional

from ai.models.schemas import Snapshot


@dataclass
class Features:
    """结构化特征,供 PromptBuilder 和规则引擎共用。"""

    # 总体规模
    total_views_raw: float           # 原始总播放量
    total_views_growth_pct: float    # 较上期增长 %

    # 平台侧
    top_platform: Optional[str]      # 占比最大的平台
    top_platform_share: float

    # 赛道侧
    top_track: Optional[str]
    top_track_growth_pct: float      # 较上期增长 %

    # 趋势异常
    anomaly_dates: List[str]         # 触发 AI 标记的日期

    # 达人侧
    rising_creators: List[str]       # 增长率 > 阈值 的达人名
    blacklisted_creators: List[str]

    # 引爆力
    above_avg_dimensions: List[str]  # 高于均值的维度

    def summary_lines(self) -> List[str]:
        """返回一组人类可读的摘要行, 用于塞进 prompt 的 facts 段。"""
        lines = []
        lines.append(
            f"总播放量较上期 {self.total_views_growth_pct:+.1f}%,"
            f" 当前规模 {self.total_views_raw:,.0f} (原始数值)"
        )
        if self.top_platform:
            lines.append(
                f"最大平台: {self.top_platform} (占比 {self.top_platform_share:.1f}%)"
            )
        if self.top_track:
            lines.append(
                f"增长最快赛道: {self.top_track} (播放量增长 {self.top_track_growth_pct:+.1f}%)"
            )
        if self.anomaly_dates:
            lines.append(
                f"近期存在播放量异常波动的日期: {', '.join(self.anomaly_dates)}"
            )
        if self.rising_creators:
            lines.append(
                f"高增长达人(>20%): {', '.join(self.rising_creators[:5])}"
            )
        if self.blacklisted_creators:
            lines.append(
                f"处于黑名单的达人: {', '.join(self.blacklisted_creators)}"
            )
        if self.above_avg_dimensions:
            lines.append(
                f"高于均值的引爆力维度: {', '.join(self.above_avg_dimensions)}"
            )
        return lines


class FeatureExtractor:
    """把 Snapshot 抽取成 Features 的可测试组件。"""

    GROWTH_THRESHOLD = 20.0  # 增长率 > 20% 视为"高增长"

    def extract(self, snap: Snapshot) -> Features:
        return Features(
            total_views_raw=self._sum_views(snap),
            total_views_growth_pct=self._overall_growth(snap),
            top_platform=self._top_platform(snap),
            top_platform_share=self._top_platform_share(snap),
            top_track=self._top_track(snap),
            top_track_growth_pct=self._top_track_growth(snap),
            anomaly_dates=self._anomaly_dates(snap),
            rising_creators=self._rising_creators(snap),
            blacklisted_creators=self._blacklisted(snap),
            above_avg_dimensions=self._above_avg_dimensions(snap),
        )

    # ---------- helpers ----------

    @staticmethod
    def _sum_views(snap: Snapshot) -> float:
        if snap.kpi:
            for k in snap.kpi:
                if k.get("key") == "views":
                    return float(k.get("raw", 0))
        # 退路:用 views_trend 累加
        return sum(float(p.get("views", 0)) for p in snap.views_trend)

    @staticmethod
    def _overall_growth(snap: Snapshot) -> float:
        for k in snap.kpi:
            if k.get("key") == "views":
                return float(k.get("delta_pct", 0))
        return 0.0

    @staticmethod
    def _top_platform(snap: Snapshot) -> Optional[str]:
        if not snap.platform_share:
            return None
        top = max(snap.platform_share, key=lambda x: x.get("share", 0))
        return top.get("platform")

    @staticmethod
    def _top_platform_share(snap: Snapshot) -> float:
        if not snap.platform_share:
            return 0.0
        top = max(snap.platform_share, key=lambda x: x.get("share", 0))
        return float(top.get("share", 0))

    @staticmethod
    def _top_track(snap: Snapshot) -> Optional[str]:
        if not snap.track_performance:
            return None
        top = max(snap.track_performance, key=lambda x: x.get("views", 0))
        return top.get("track")

    @staticmethod
    def _top_track_growth(snap: Snapshot) -> float:
        # 真实实现会从数仓取同比;这里给一个与 views trend 联动的近似值
        if not snap.views_trend:
            return 0.0
        ratios = [float(p.get("ratio", 0)) for p in snap.views_trend]
        return sum(ratios) / len(ratios) if ratios else 0.0

    @staticmethod
    def _anomaly_dates(snap: Snapshot) -> List[str]:
        return [p["date"] for p in snap.views_trend if p.get("has_anomaly")]

    def _rising_creators(self, snap: Snapshot) -> List[str]:
        return [
            c["name"]
            for c in snap.top_creators
            if float(c.get("growth_30d", 0)) > self.GROWTH_THRESHOLD
        ]

    @staticmethod
    def _blacklisted(snap: Snapshot) -> List[str]:
        return [c["name"] for c in snap.top_creators if c.get("blacklist")]

    @staticmethod
    def _above_avg_dimensions(snap: Snapshot) -> List[str]:
        return [
            d["dimension"]
            for d in snap.explosive_radar
            if float(d.get("value", 0)) > float(d.get("avg", 0))
        ]
