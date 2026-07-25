"""
ai.main
=======
FastAPI 应用入口。

启动方式 (在项目根目录 F:\\workbuddy\\影石 下):

    # 1) 装依赖
    python -m venv .venv && . .venv/Scripts/activate
    pip install -r ai/requirements.txt

    # 2) 启动(无 LLM 也可,自动走规则兜底)
    python -m ai.main

启用通义千问:
    set DASHSCOPE_API_KEY=sk-xxx     # Windows
    export DASHSCOPE_API_KEY=sk-xxx  # macOS / Linux
    python -m ai.main

启用 OpenAI:
    set INSTINSIGHT_LLM_PROVIDER=openai
    set OPENAI_API_KEY=sk-xxx
    python -m ai.main

默认监听 0.0.0.0:9000, 可通过环境变量 AI_PORT 调整。
"""
from __future__ import annotations

import os
import sys

from fastapi import FastAPI

from ai.api.routes import init_engine, router as insight_router
from ai.core.llm_client import LLMClient


def _build_llm_or_none() -> LLMClient | None:
    """
    尝试从环境变量构造 LLM 客户端; 失败也不致命,
    InsightEngine 内部会自动回退到 rule-based 方案。
    """
    try:
        return LLMClient.from_env()
    except RuntimeError as e:
        print(f"[ai-main] LLM 未启用, 使用规则模式: {e}", file=sys.stderr)
        return None


def create_app() -> FastAPI:
    app = FastAPI(
        title="Insta360 AI Insight Engine",
        description="Insta360 达人营销数据洞察平台 — AI 智能洞察引擎",
        version="0.1.0",
    )

    # 启动时尝试构造 LLM, 失败回退到规则模式
    init_engine(_build_llm_or_none())

    app.include_router(insight_router)

    @app.get("/")
    def root() -> dict:
        return {
            "service": "insta360-ai-insight",
            "endpoints": ["/v1/insights", "/v1/health"],
        }

    return app


app = create_app()


if __name__ == "__main__":
    import uvicorn

    port = int(os.environ.get("AI_PORT", "9000"))
    uvicorn.run("ai.main:app", host="0.0.0.0", port=port, reload=False)
