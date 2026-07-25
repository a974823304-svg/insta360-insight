"""
ai.api.routes
=============
FastAPI 路由定义。
"""
from __future__ import annotations

from fastapi import APIRouter, HTTPException

from ai.core.insight_engine import InsightEngine
from ai.core.llm_client import LLMClient
from ai.models.schemas import InsightRequest, InsightResponse

router = APIRouter(prefix="/v1", tags=["insights"])

# 单例 engine; 启动时由 main.py 注入(避免在 import 时就尝试读环境变量)
_engine: InsightEngine | None = None


def init_engine(llm: LLMClient | None) -> None:
    """由 main.py 在应用启动时调用。"""
    global _engine
    _engine = InsightEngine(llm=llm)


def _get_engine() -> InsightEngine:
    if _engine is None:
        raise HTTPException(status_code=503, detail="Insight engine not initialized")
    return _engine


@router.post("/insights", response_model=InsightResponse)
def generate_insights(req: InsightRequest) -> InsightResponse:
    """
    接收 Filter + 当前看板快照, 返回 N 条业务可读洞察。

    调用方: Go 后端 internal/service/ai_service.go
            真实部署时: HTTP POST /v1/insights
    """
    return _get_engine().generate(req)


@router.get("/health")
def health() -> dict:
    """健康检查: 用于 k8s liveness probe。"""
    return {"status": "ok", "service": "insta360-ai-insight"}
