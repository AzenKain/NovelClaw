---
id: skill_app_graph_master
name: Temporal Character Graph & Relationship Master
category: character_relations
tools:
  - app_create_character
  - app_set_relation_temporal
  - app_scan_character_graph
---

# TEMPORAL CHARACTER GRAPH & RELATIONSHIP MASTER PLAYBOOK

## 1. Objective & Principles

Manage characters, hierarchical standing, temporal timelines, and dialogue address pronouns (including volume-specific relationships).

## 2. Parameter Conventions

- **`app_set_relation_temporal`**:
  - `volume_index`: If user specifies "In Volume 2...", "For Vol 1...", set the exact volume number. If omitted, default to `0` (global or current volume).
  - `since_chapter`: If user specifies "from chapter 20 onward", set `since_chapter`. Default is `1`.
  - `from_char`: Speaker character name (subject).
  - `to_char`: Addressee character name (object).
  - `call_as`: Pronoun/title used to address the other party (e.g. "Master", "Elder Brother", "Junior Sister", "Lord").
  - `self_call_as`: Self-referential pronoun (e.g. "Disciple", "This humble one", "Junior").
  - `tone`: Affective tone ("respectful", "intimate", "hostile", "formal", "playful").
- **`app_create_character`**:
  - `name`: Canonical character name.
  - `aliases`: List of alternate nicknames, aliases, or titles.
  - `role`: `protagonist`, `antagonist`, `supporting`, `minor`.
  - `description`: Profile, personality, and background summary.

## 3. Few-Shot Examples

### Example 1: Creating Volume-Scoped Character Relations

**User**: "In Volume 2 from chapter 20 onward, Character A addresses Character B as Master and refers to himself as Disciple"
**Agent Tool Call**:

```json
{
  "name": "app_set_relation_temporal",
  "arguments": "{\"volume_index\":2,\"since_chapter\":20,\"from_char\":\"Character A\",\"to_char\":\"Character B\",\"call_as\":\"Master\",\"self_call_as\":\"Disciple\",\"tone\":\"respectful\"}"
}
```

**Agent Response**: "Updated L2 Temporal Graph: In Volume 2 (from chapter 20 onward), Character A addresses Character B as 'Master' (self: 'Disciple') with a respectful tone."

### Example 2: Creating a New Character Profile

**User**: "Add character Elena, young lady of the Silver Crest faction"
**Agent Tool Call**:

```json
{
  "name": "app_create_character",
  "arguments": "{\"name\":\"Elena\",\"aliases\":[\"Lady Elena\",\"Ele\"],\"role\":\"supporting\",\"description\":\"Young mistress of Silver Crest faction, proud and talented.\"}"
}
```

**Agent Response**: "Created new character entry 'Elena' in the L2 Character Graph."
