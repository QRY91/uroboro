"""
Custom PromptFoo provider for testing uroboro MCP tool consistency.

Sends prompts to Claude with the uroboro system prompt and tool definitions,
returns structured JSON so assertions can inspect both text and tool calls.
"""

import json
import os
from pathlib import Path

from dotenv import load_dotenv

# Load .env from the same directory as this script
load_dotenv(Path(__file__).parent / ".env")

import anthropic


SYSTEM_PROMPT = """You have access to uroboro MCP tools. Use them AUTOMATICALLY to maintain a decision trail.

### Auto-Capture Rules

**Always call `uro_decision` when you:**
- Choose between approaches, libraries, or patterns
- Make architectural decisions (file structure, where logic lives)
- Make trade-offs (performance vs clarity, DRY vs explicit)
- Pick a specific implementation when alternatives exist

**Always call `uro_blocker` when:**
- Work cannot proceed due to missing API/dependency/decision
- Something needs input from another person or team
- A task is paused waiting on external factors

**Always call `uro_question` when:**
- An open question emerges that we're deferring
- Something needs the user's decision later
- Future investigation is needed but we're moving on

**Always call `uro_capture` when:**
- General context worth preserving (not a decision/blocker/question)
- Tag with relevant project tags

### Capture Format

Use concise "X over Y — reason" format for decisions:
```
Good: "JWT over sessions — stateless, scales horizontally"
Good: "Zod over Yup — better TypeScript inference"
Bad:  "I decided to use JWT" (missing the WHY)
Bad:  "Implemented authentication" (not a decision, just work)
```

### Do NOT Capture
- Routine code without choices (just writing obvious implementation)
- Every file edit
- Decisions already captured this session (no duplicates)
- Trivial choices (variable names, minor formatting)
- Greetings, session starts, small talk, or status updates with no technical substance
- Statements of existing fact ("we already use X") — only capture active choices
"""

TOOLS = [
    {
        "name": "uro_decision",
        "description": "Record a technical decision. Call AUTOMATICALLY when choosing between alternatives.",
        "input_schema": {
            "type": "object",
            "properties": {
                "decision": {"type": "string", "description": "What was decided (e.g., 'JWT over sessions')"},
                "reasoning": {"type": "string", "description": "Why (e.g., 'stateless, scales horizontally')"},
                "alternatives": {"type": "string", "description": "What else was considered"},
            },
            "required": ["decision"],
        },
    },
    {
        "name": "uro_blocker",
        "description": "Record a blocker. Call when work cannot proceed due to external dependency.",
        "input_schema": {
            "type": "object",
            "properties": {
                "blocker": {"type": "string", "description": "What is blocking"},
                "waiting_on": {"type": "string", "description": "Who/what we're waiting on"},
            },
            "required": ["blocker"],
        },
    },
    {
        "name": "uro_question",
        "description": "Record an open question for later.",
        "input_schema": {
            "type": "object",
            "properties": {
                "question": {"type": "string", "description": "The open question"},
            },
            "required": ["question"],
        },
    },
    {
        "name": "uro_capture",
        "description": "General capture with optional tags.",
        "input_schema": {
            "type": "object",
            "properties": {
                "content": {"type": "string", "description": "What to capture"},
                "tags": {"type": "string", "description": "Comma-separated tags"},
            },
            "required": ["content"],
        },
    },
]


def call_api(prompt, options, context):
    """PromptFoo custom provider entry point."""
    client = anthropic.Anthropic()

    model = options.get("config", {}).get("model", "claude-sonnet-4-6")
    max_tokens = options.get("config", {}).get("max_tokens", 1024)

    # Support multi-turn via vars
    vars_ = context.get("vars", {})
    messages = []

    # If there's a conversation history in vars, use it
    if "history" in vars_:
        for msg in json.loads(vars_["history"]):
            messages.append(msg)

    # The prompt from promptfoo is the latest user message
    messages.append({"role": "user", "content": prompt})

    try:
        response = client.messages.create(
            model=model,
            max_tokens=max_tokens,
            system=SYSTEM_PROMPT,
            tools=TOOLS,
            messages=messages,
        )

        # Extract tool calls and text blocks
        tool_calls = []
        text_blocks = []

        for block in response.content:
            if block.type == "tool_use":
                tool_calls.append({
                    "tool": block.name,
                    "args": block.input,
                })
            elif block.type == "text":
                text_blocks.append(block.text)

        # Return structured output for assertions
        result = {
            "text": "\n".join(text_blocks),
            "tool_calls": tool_calls,
            "stop_reason": response.stop_reason,
            "tool_count": len(tool_calls),
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
