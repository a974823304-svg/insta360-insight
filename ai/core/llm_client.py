"""
ai.core.llm_client
==================
对大模型 API 的统一抽象。

支持 provider:
- qwen  : 阿里云百炼 / 通义千问 (dashscope SDK)
- openai: 兼容 OpenAI Chat Completions 的服务 (azure / oneapi / 自建都可用)

调用方只需:
    client = LLMClient.from_env()
    raw = client.chat(messages)

如果环境里没有相关 SDK / Key, 构造时会抛 RuntimeError,
由上层 InsightEngine 捕获并回退到 rule-based 方案。
"""
from __future__ import annotations

import os
from dataclasses import dataclass
from typing import List, Optional


@dataclass
class LLMConfig:
    provider: str          # qwen / openai
    model: str             # qwen-plus / gpt-4o-mini ...
    api_key: str
    base_url: Optional[str] = None


class LLMClient:
    """统一的 Chat Completions 客户端。"""

    def __init__(self, cfg: LLMConfig):
        self.cfg = cfg

    # ----------------------------------------------------------------
    # 工厂方法:从环境变量构建
    # ----------------------------------------------------------------
    @classmethod
    def from_env(cls) -> "LLMClient":
        """
        优先级:
        1) INSTINSIGHT_LLM_PROVIDER=openai + OPENAI_API_KEY
        2) INSTINSIGHT_LLM_PROVIDER=qwen  + DASHSCOPE_API_KEY
        """
        provider = os.environ.get("INSTINSIGHT_LLM_PROVIDER", "qwen").lower()
        if provider == "qwen":
            api_key = os.environ.get("DASHSCOPE_API_KEY", "")
            model = os.environ.get("INSTINSIGHT_LLM_MODEL", "qwen-plus")
            if not api_key:
                raise RuntimeError("DASHSCOPE_API_KEY not set; 无法启用通义千问模式")
            return cls(LLMConfig(provider="qwen", model=model, api_key=api_key))
        elif provider == "openai":
            api_key = os.environ.get("OPENAI_API_KEY", "")
            model = os.environ.get("INSTINSIGHT_LLM_MODEL", "gpt-4o-mini")
            base_url = os.environ.get("OPENAI_BASE_URL")
            if not api_key:
                raise RuntimeError("OPENAI_API_KEY not set; 无法启用 OpenAI 模式")
            return cls(LLMConfig(provider="openai", model=model, api_key=api_key, base_url=base_url))
        else:
            raise RuntimeError(f"未知的 LLM provider: {provider}")

    # ----------------------------------------------------------------
    # 统一调用入口
    # ----------------------------------------------------------------
    def chat(self, messages: List[dict], timeout: float = 30.0) -> str:
        if self.cfg.provider == "qwen":
            return self._call_qwen(messages, timeout)
        if self.cfg.provider == "openai":
            return self._call_openai(messages, timeout)
        raise RuntimeError(f"未实现的 provider: {self.cfg.provider}")

    # ----------------------------------------------------------------
    # provider 私有实现
    # ----------------------------------------------------------------
    def _call_qwen(self, messages: List[dict], timeout: float) -> str:
        # dashscope 是可选依赖, 缺失时抛 ImportError, 由上层兜底
        import dashscope
        from dashscope import Generation

        dashscope.api_key = self.cfg.api_key
        resp = Generation.call(
            model=self.cfg.model,
            messages=messages,
            result_format="message",
            timeout=timeout,
        )
        if getattr(resp, "status_code", 200) != 200:
            raise RuntimeError(f"qwen call failed: {getattr(resp, 'code', '?')} {getattr(resp, 'message', '')}")
        return resp.output.choices[0].message.content

    def _call_openai(self, messages: List[dict], timeout: float) -> str:
        # openai>=1.x SDK
        from openai import OpenAI

        client_kwargs = {"api_key": self.cfg.api_key, "timeout": timeout}
        if self.cfg.base_url:
            client_kwargs["base_url"] = self.cfg.base_url
        client = OpenAI(**client_kwargs)

        resp = client.chat.completions.create(
            model=self.cfg.model,
            messages=messages,
            temperature=0.4,
        )
        return resp.choices[0].message.content or ""
