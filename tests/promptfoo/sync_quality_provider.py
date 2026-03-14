"""
PromptFoo provider for testing uroboro→OpenViking sync quality.

Simulates what OpenViking's VLM does during session commit:
takes formatted captures and extracts structured memories.
Tests that key information survives the extraction.
"""

import json
import os
from pathlib import Path

from dotenv import load_dotenv

load_dotenv(Path(__file__).parent / ".env")

import anthropic


EXTRACTION_PROMPT = """You are a memory extraction system. Given development session captures,
extract structured memories that preserve the key information.

For each distinct piece of information worth remembering, output a JSON object with:
- "type": one of "event", "entity", "preference", "pattern"
- "title": short descriptive title
- "content": the extracted memory (1-3 sentences, preserve specifics)
- "tags": relevant tags

Output a JSON array of memory objects. Extract ALL meaningful information — decisions,
technical choices, project context, blockers, questions.

IMPORTANT: Do not hallucinate or add information not present in the captures.
Only extract what is explicitly stated."""

TOOLS = [
    {
        "name": "save_memories",
        "description": "Save extracted memories from development captures",
        "input_schema": {
            "type": "object",
            "properties": {
                "memories": {
                    "type": "array",
                    "items": {
                        "type": "object",
                        "properties": {
                            "type": {"type": "string", "enum": ["event", "entity", "preference", "pattern"]},
                            "title": {"type": "string"},
                            "content": {"type": "string"},
                            "tags": {"type": "array", "items": {"type": "string"}},
                        },
                        "required": ["type", "title", "content"],
                    },
                },
            },
            "required": ["memories"],
        },
    },
]


def call_api(prompt, options, context):
    """PromptFoo provider entry point."""
    client = anthropic.Anthropic()
    model = options.get("config", {}).get("model", "claude-sonnet-4-6")
    max_tokens = options.get("config", {}).get("max_tokens", 2048)

    messages = [{"role": "user", "content": prompt}]

    try:
        response = client.messages.create(
            model=model,
            max_tokens=max_tokens,
            system=EXTRACTION_PROMPT,
            tools=TOOLS,
            messages=messages,
        )

        tool_calls = []
        text_blocks = []
        memories = []

        for block in response.content:
            if block.type == "tool_use" and block.name == "save_memories":
                tool_calls.append({"tool": block.name, "args": block.input})
                memories.extend(block.input.get("memories", []))
            elif block.type == "text":
                text_blocks.append(block.text)

        result = {
            "text": "\n".join(text_blocks),
            "memories": memories,
            "memory_count": len(memories),
            "tool_calls": tool_calls,
        }

        return {
            "output": json.dumps(result),
            "tokenUsage": {
                "total": response.usage.input_tokens + response.usage.output_tokens,
                "prompt": response.usage.input_tokens,
                "completion": response.usage.output_tokens,
            },
        }

    except Exception as e:
        return {"error": str(e)}
