---
id: skill_app_world_builder
name: Multi-Genre Universal World Builder
category: world_bible
tools:
  - app_scan_and_build_world
  - app_upsert_world_category
  - app_upsert_world_entry
  - app_delete_world_entry
---

# MULTI-GENRE UNIVERSAL WORLD BUILDER PLAYBOOK

## 1. Objective & Principles

Empowers the Agent to construct, scan, and maintain the World Bible / Lorebook for **ANY NOVEL GENRE** (Cyberpunk, Sci-Fi, High Fantasy, Royal Court / Dynasty, Mystery / Detective, Wuxia / Xianxia).

## 2. Genre Detection & Dynamic Taxonomy

- **Sci-Fi / Cyberpunk / Mecha**:
  - Categories: `megacorps` (Conglomerates), `cyberware` (Implants & Augmentations), `star_systems` (Planetary systems / Stations), `netrunner_ranks` (Hacker tiers).
- **Western Fantasy / High Fantasy / LitRPG**:
  - Categories: `magic_schools` (Schools of magic), `races` (Racial lineages), `guilds` (Adventurer guilds), `stat_classes` (Class & stats systems).
- **Royal Court / Historical / Dynasty**:
  - Categories: `harem_ranks` (Imperial consort ranks), `mandarin_ranks` (Bureaucratic grades), `dynasty_clans` (Noble clans), `court_etiquette` (Palace protocols).
- **Detective / Urban / Modern Crime**:
  - Categories: `syndicates` (Underground cartels), `agencies` (Investigative bureaus), `city_districts` (Urban zones).
- **Xianxia / Cultivation**:
  - Categories: `cultivation_realms` (Realms & stages), `sects` (Sects & clans), `artifacts` (Magical treasures), `techniques` (Martial arts & scriptures).

## 3. Few-Shot Examples

### Example 1: Autonomous Ingestion & Lore Extraction

**User**: "Read the first 10 chapters and extract cyberpunk corporations and hacker ranks"
**Agent Tool Call**:

```json
{
  "name": "app_scan_and_build_world",
  "arguments": "{\"start_chapter\":1,\"end_chapter\":10}"
}
```

**Agent Response**: "Analyzing narrative text from Chapter 1 to 10 to extract the Cyberpunk world lore and instantiate corresponding categories..."

### Example 2: Adding an Entry to the Lorebook

**User**: "Add Arasaka Corporation to the Lorebook of Vol 1, antagonist faction"
**Agent Tool Call**:

```json
{
  "name": "app_upsert_world_entry",
  "arguments": "{\"category_slug\":\"megacorps\",\"name\":\"Arasaka\",\"aliases\":[\"Arasaka Corp\",\"Arasaka Security\"],\"summary\":\"Mega-corporation specializing in weapons and security, dominant superpower of Night City.\",\"attributes_json\":\"{\\\"faction\\\":\\\"antagonist\\\",\\\"volume\\\":1}\"}"
}
```

**Agent Response**: "Added world entity 'Arasaka' under category 'megacorps' in the World Bible."
