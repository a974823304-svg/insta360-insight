"""
ai.core.prompt_builder
======================
把 Features + Filter 拼装成 LLM prompt。

设计原则
--------
- System 段:固定角色 + 输出格式约束(JSON Schema),降低模型幻觉概率。
- User 段:facts(已计算的特征) + 上下文(filter / 时间范围),用 markdown 列表,
  模型只需要"翻译"为业务文案。
- max_items 决定 Insight 条数。
"""
from __future__ import annotations

import json
from typing import List

from ai.core.feature_extractor import Features
from ai.models.schemas import Filter, Insight


SYSTEM_TEMPLATE = """你是 Insta360 达人营销数据洞察平台的 AI 业务分析助手。
你的任务是基于"已计算好的事实"生成 {n} 条面向市场/运营同事的洞察, 帮助他们快速做投放决策。

要求:
- 每条洞察 1 句话标题 + 1~2 句话正文, 总长度不超过 80 个汉字。
- 必须基于下面的 facts 给出, 不要编造没有的数字。
- 按业务重要性排序(最值得关注的在前)。
- 严格用 JSON 数组输出, 每个元素格式:
  {{"icon": "surge|alert|star|info", "title": "...", "body": "...", "severity": "info|warning|success"}}
- 不要输出任何 JSON 之外的内容。
"""


class PromptBuilder:
    """纯函数式组件, 不依赖 LLM, 方便单测。"""

    def build_messages(
        self,
        features: Features,
        flt: Filter,
        max_items: int = 3,
    ) -> List[dict]:
        return [
            {"role": "system", "content": SYSTEM_TEMPLATE.format(n=max_items)},
            {"role": "user", "content": self._render_user(features, flt)}
        ]

    @staticmethod
    def _render_user(features: Features, flt: Filter) -> str:
        facts = "\n".join(f"- {line}" for line in features.summary_lines())
        ctx_parts = []
        if flt.date_range:
            ctx_parts.append(f"时间范围: {flt.date_range[0]} ~ {flt.date_range[-1]}")
        if flt.tracks:
            ctx_parts.append(f"运动赛道: {', '.join(flt.tracks)}")
        if flt.platforms:
            ctx_parts.append(f"平台: {', '.join(flt.platforms)}")
        if flt.regions:
            ctx_parts.append(f"地区: {', '.join(flt.regions)}")
        ctx = "\n".join(ctx_parts) or "(无额外筛选)"

        return (
            "## 上下文\n"
            f"{ctx}\n\n"
            "## 关键事实(已计算好)\n"
            f"{facts}\n\n"
            "请输出 JSON 数组。"
        )

    @staticmethod
    def parse_response(raw: str) -> List[Insight]:
        """解析 LLM 输出。失败时抛 ValueError, 由调用方决定兜底。"""
        # 容错:模型有时会把 JSON 包在 ```json ... ``` 里
        text = raw.strip()
        if text.startswith("```"):
            text = text.strip("`")
            if text.startswith("json"):
                text = text[4:]
        data = json.loads(text)
        if not isinstance(data, list):
            raise ValueError("LLM 输出不是 JSON 数组")
        return [Insight(**item) for item in data]
