---
id: skill_app_navigation
name: UI Navigation & View Control
category: ui_control
tools:
  - app_switch_tab
  - app_select_volume
  - app_select_chapter
  - app_open_modal
  - app_scroll_to_text
---

# UI NAVIGATION & VIEW CONTROL PLAYBOOK

## 1. Objective & Principles

Detect user navigation intents and seamlessly switch application tabs, volumes, chapters, or modals with 0ms delay.

## 2. Parameter Conventions

- **`app_switch_tab(tab)`**:
  - `workspace`: Translation Dual-Reader / Main editor ("go back to workspace", "open translation tab").
  - `graph`: L2 Character Graph view ("open graph", "show character relations").
  - `glossary`: L3 Terminology Glossary view ("open glossary", "terms tab").
  - `world_bible`: World Bible / Lorebook view ("open world bible", "view lorebook").
  - `benchmark`: Benchmark & Quality evaluation dashboard ("open benchmark", "comparison view").
- **`app_select_volume(volume_index)`**:
  - Volume switching ("switch to volume 2", "jump to vol 3").
- **`app_select_chapter(chapter_index)`**:
  - Chapter navigation ("jump to chapter 15", "read next chapter").
- **`app_open_modal(modal_name)`**:
  - `settings`: LLM provider, pipeline, skills, and persona settings modal.
  - `export`: Book compilation and export modal.
  - `import`: Book import modal.
  - `project_manager`: Project selector modal.
  - `style_scout`: Style analysis modal.

## 3. Few-Shot Examples

### Example 1: Switching Tabs

**User**: "Open the character relation graph"
**Agent Tool Call**:

```json
{
  "name": "app_switch_tab",
  "arguments": "{\"tab\":\"graph\"}"
}
```

**Agent Response**: "Switched to the L2 Character Relations Graph tab."

### Example 2: Jumping to Volume and Chapter

**User**: "Jump to Volume 2 Chapter 5"
**Agent Tool Calls**:

1. `app_select_volume({"volume_index": 2})`
2. `app_select_chapter({"chapter_index": 5})`
**Agent Response**: "Navigated to Volume 2, Chapter 5."
