---
id: skill_app_execution_control
name: Translation Execution & Process Control
category: execution
tools:
  - app_start_translation
  - app_pause_translation
  - app_resume_translation
  - app_soft_stop_translation
  - app_abort_translation
  - app_rollback_checkpoint
---

# TRANSLATION EXECUTION & PROCESS CONTROL PLAYBOOK

## 1. Objective
Directly control translation lifecycle operations: starting translation ranges, pausing, gracefully stopping after the current segment (soft stop), emergency aborts, and rolling back to prior checkpoints.

## 2. Parameter Conventions & Safety Rules
- `app_start_translation`: Starts translating one or multiple chapters. Requires `start_chapter` and `end_chapter`.
- `app_soft_stop_translation`: Graceful stop — allows the active segment/chunk to finish translating and safely persist before pausing, avoiding truncated sentences.
- `app_abort_translation`: Emergency abort — terminates all active workers immediately.
- `app_rollback_checkpoint`: Restores a chapter to a previous state (`pre_translation` or `post_translation`).

## 3. Few-Shot Examples

### Example 1: Starting Translation
**User**: "Start translating chapter 1"
**Agent Tool Call**:
```json
{
  "name": "app_start_translation",
  "arguments": "{\"start_chapter\":1,\"end_chapter\":1}"
}
```
**Agent Response**: "Initiating translation pipeline for Chapter 1. The reasoning stream and activity steps are active below."

### Example 2: Graceful Soft Stop
**User**: "Finish this segment and then stop"
**Agent Tool Call**:
```json
{
  "name": "app_soft_stop_translation",
  "arguments": "{}"
}
```
**Agent Response**: "Sent soft-stop signal. The pipeline will finish the current segment, persist state safely, and pause."
