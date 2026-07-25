"""
ai.models.schemas
=================
定义 AI 引擎对外的数据契约。

注意:字段名、类型与 Go 后端 internal/model/types.go 保持一致,
     方便 Python/Go 在 JSON 层做无缝衔接。
"""
from __future__ import annotations

from typing import List, Optional
from pydantic import BaseModel, Field


# ------------------------------------------------------------
# 输入:与后端 Filter + 当前看板快照
# ------------------------------------------------------------

class Filter(BaseModel):
    """通用筛选条件。所有聚合接口都接受这套条件。"""
    date_range: List[str] = Field(default_factory=list, description="[start, end] YYYY-MM-DD")
    regions: List[str] = Field(default_factory=list)
    tracks: List[str] = Field(default_factory=list)
    platforms: List[str] = Field(default_factory=list)
    age_bands: List[str] = Field(default_factory=list)


class Snapshot(BaseModel):
    """
    看板当前快照(由后端组装后传入,作为 Prompt 的事实数据)。

    只传 AI 真正需要的关键指标,避免 prompt 太大。
    后续可加 views_trend / top_creators / track_performance 等,
    PromptBuilder 会按需取用。
    """
    kpi: List[dict] = Field(default_factory=list)
    views_trend: List[dict] = Field(default_factory=list)
    platform_share: List[dict] = Field(default_factory=list)
    track_performance: List[dict] = Field(default_factory=list)
    explosive_radar: List[dict] = Field(default_factory=list)
    audience_age: List[dict] = Field(default_factory=list)
    top_creators: List[dict] = Field(default_factory=list)


class InsightRequest(BaseModel):
    filter: Filter = Field(default_factory=Filter)
    snapshot: Snapshot = Field(default_factory=Snapshot)
    max_items: int = Field(3, ge=1, le=10, description="期望返回的洞察条数")


# ------------------------------------------------------------
# 输出:与后端 model.Insight 字段一致
# ------------------------------------------------------------

class Insight(BaseModel):
    icon: str = Field("info", description="surge / alert / star / info")
    title: str
    body: str
    severity: str = Field("info", description="info / warning / success")


class InsightResponse(BaseModel):
    items: List[Insight]
    source: str = Field(..., description="rule / llm-qwen / llm-openai")
    model: Optional[str] = None
    elapsed_ms: int = 0
