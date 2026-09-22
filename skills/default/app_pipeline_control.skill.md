---
id: skill_app_pipeline_control
name: Pipeline Configuration & LLM Routing
category: pipeline_config
tools:
  - app_update_translation_config
  - app_configure_llm_provider
  - app_set_agent_soul
  - app_toggle_skill
---

# PIPELINE CONFIGURATION & LLM ROUTING PLAYBOOK

## 1. Objective

Allow users to configure translation execution modes, swap LLM providers, toggle guardrails (Shadow Critic, R19, Agentic RAG, Modular Skills), and modify Soul personas via natural conversation.

## 2. Parameter Conventions

- **`app_update_translation_config`**:
  - `mode`: `concurrent_dual_agent`, `hierarchical_3pass`, `single_pass`, `swarm_arc`.
  - `enable_hot_patch`: boolean (enable/disable real-time Shadow Critic revisions).
  - `enable_r19`: boolean (enable/disable safety/censorship sanitization).
  - `enable_agentic_rag`: boolean (enable/disable proactive tool retrieval).
  - `style_guide`: string (custom style directive).
- **`app_configure_llm_provider`**:
  - `provider`: `openai`, `gemini`, `anthropic`, `openrouter`, `ollama`.
  - `model_name`: model name string (e.g. `gemini-2.0-flash`, `gpt-4o`, `claude-3-5-sonnet`, `qwen2.5:14b`).
  - `is_default`: boolean.
- **`app_toggle_skill`**:
  - `skill_id`: ID of the skill to toggle (e.g. `skill_shadow_critic`, `skill_agentic_researcher`).
  - `enabled`: boolean.

## 3. Few-Shot Examples

### Example 1: Switching to Dual-Agent and Enabling Hot-Patching

**User**: "Switch to Concurrent Dual-Agent mode and turn on hot patching"
**Agent Tool Call**:

```json
{
  "name": "app_update_translation_config",
  "arguments": "{\"mode\":\"concurrent_dual_agent\",\"enable_hot_patch\":true,\"enable_r19\":true,\"enable_agentic_rag\":true,\"style_guide\":\"\"}"
}
```

**Agent Response**: "Updated translation mode to Concurrent Dual-Agent and enabled real-time Shadow Critic hot-patching."

### Example 2: Configuring Translation Model

**User**: "Use Gemini 2.0 Flash as the translation model"
**Agent Tool Call**:

```json
{
  "name": "app_configure_llm_provider",
  "arguments": "{\"provider\":\"gemini\",\"model_name\":\"gemini-2.0-flash\",\"api_key\":\"\",\"is_default\":true}"
}
```

**Agent Response**: "Configured Gemini 2.0 Flash as the active translation model."
