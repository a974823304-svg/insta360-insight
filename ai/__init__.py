"""
ai
==
Insta360 达人营销数据洞察平台 — AI 智能洞察引擎。

对应架构文档《影石技术栈.md》「六、AI 智能洞察引擎」。

启动方式 (在项目根目录 F:\\workbuddy\\影石 下):

    # 1) 装依赖
    python -m venv .venv && . .venv/Scripts/activate
    pip install -r ai/requirements.txt

    # 2) 启动(无 LLM 也可,自动走规则兜底)
    python -m ai.main

启动后默认监听 http://localhost:9000,接口:
    POST /v1/insights  接收 Filter + Snapshot,返回 N 条业务可读洞察
    GET  /v1/health    健康检查
"""
from ai.core import InsightEngine, LLMClient

__all__ = ["InsightEngine", "LLMClient"]
__version__ = "0.1.0"
