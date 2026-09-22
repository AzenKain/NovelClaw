---
id: skill_app_publishing_export
name: Book Publishing, Self-Evolution & Benchmarking
category: publishing
tools:
  - app_export_book
  - app_trigger_learn_evolution
  - app_run_benchmark
---

# BOOK PUBLISHING, SELF-EVOLUTION & BENCHMARKING PLAYBOOK

## 1. Objective

Support final delivery workflows: compiling translated chapters into multi-format ebooks (EPUB, PDF, DOCX, TXT), triggering the Reflexion self-learning cycle (`/learn`), and running translation quality benchmarks across modes.

## 2. Parameter Conventions

- **`app_export_book`**:
  - `format`: `epub`, `pdf`, `docx`, `txt`.
  - `include_cover`: boolean (default: `true`).
  - `output_path`: custom save destination (or empty string for project default directory).
- **`app_trigger_learn_evolution`**:
  - `notes`: Specific editorial guidance notes (e.g. "Learn intimate address conventions from Chapter 2 edits").

## 3. Few-Shot Examples

### Example 1: Exporting EPUB Book

**User**: "Export the novel to EPUB format with book cover"
**Agent Tool Call**:

```json
{
  "name": "app_export_book",
  "arguments": "{\"format\":\"epub\",\"include_cover\":true,\"output_path\":\"\"}"
}
```

**Agent Response**: "Packaging all completed chapters into a publication-ready EPUB ebook with cover artwork..."

### Example 2: Triggering Self-Evolution

**User**: "Learn from the edits I just made to chapter 1"
**Agent Tool Call**:

```json
{
  "name": "app_trigger_learn_evolution",
  "arguments": "{\"notes\":\"Learn from human editor modifications in Chapter 1\"}"
}
```

**Agent Response**: "Triggered Reflexion self-evolution engine. Extracting few-shot translation patterns from chapter 1 revisions."
