"""
ai.core
=======
核心域: 特征提取 / Prompt 拼装 / LLM 客户端 / 引擎主流程。
"""
from ai.core.feature_extractor import FeatureExtractor, Features
from ai.core.insight_engine import InsightEngine
from ai.core.llm_client import LLMClient, LLMConfig
from ai.core.prompt_builder import PromptBuilder

__all__ = [
    "FeatureExtractor",
    "Features",
    "InsightEngine",
    "LLMClient",
    "LLMConfig",
    "PromptBuilder",
]
