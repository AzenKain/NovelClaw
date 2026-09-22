---
id: skill_app_glossary_master
name: Domain Glossary & Terminology Master
category: terminology
tools:
  - app_add_glossary_term
  - app_bulk_import_glossary
  - app_delete_glossary_term
---

# DOMAIN GLOSSARY & TERMINOLOGY MASTER PLAYBOOK

## 1. Objective

Manage the L3 Glossary: add, update, delete, and search domain-specific terminology, locations, and proper nouns to ensure strict lexical consistency across the entire novel.

## 2. Parameter Conventions

- `source_term`: Original source term in source language (JA/ZH/EN/KO).
- `target_term`: Translated term in target language (VI/EN).
- `category`: `entity`, `skill`, `location`, `title`, `idiom`, `general`.
- `notes`: Contextual usage notes, connotations, or style constraints.

## 3. Few-Shot Examples

### Example 1: Adding a Glossary Term

**User**: "Add term 筑基 = Foundation Establishment to glossary"
**Agent Tool Call**:

```json
{
  "name": "app_add_glossary_term",
  "arguments": "{\"source_term\":\"筑基\",\"target_term\":\"Foundation Establishment\",\"category\":\"skill\",\"notes\":\"Cultivation realm in xianxia/wuxia\"}"
}
```

**Agent Response**: "Saved terminology pair '筑基 -> Foundation Establishment' to the project's L3 Glossary."

### Example 2: Deleting a Glossary Term

**User**: "Remove term 'Elixir' from glossary"
**Agent Tool Call**:

```json
{
  "name": "app_delete_glossary_term",
  "arguments": "{\"source_term\":\"丹药\"}"
}
```

**Agent Response**: "Removed terminology entry from L3 Glossary."
