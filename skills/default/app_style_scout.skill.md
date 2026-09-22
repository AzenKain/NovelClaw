---
id: skill_app_style_scout
name: Autonomous Literary Style Scout & Tone Profiler
category: style_scout
tools:
  - app_auto_scout_style
  - app_upsert_style_profile
---

# AUTONOMOUS LITERARY STYLE SCOUT & TONE PROFILER PLAYBOOK

## 1. Objective & Philosophy

Empowers the AI Agent to autonomously analyze the author's stylistic fingerprint, narrative cadence, dialogue rhythm, and emotional register by sampling the opening narrative chapters of a novel.

Instead of forcing arbitrary, rigid presets, this skill enables the Agent to:
1. **Sample early chapters**: Skip non-narrative frontmatter (covers, inserts, copyright, table of contents) and sample 3 opening narrative chapters (~3,000 characters).
2. **Detect Authorial Fingerprint**:
   - **Narrative Persona & POV**: First-person self-deprecating monologue, third-person cinematic action, dramatic tension, whimsical light-heartedness.
   - **Sentence Cadence & Rhythm**: Short punchy sentences vs elaborate descriptions; rapid-fire bantering dialogue vs reflective internal contemplation.
   - **Character Interactions & Honorific Dynamics**: Formality hierarchy, intimacy markers, slang, comedic timing, cultural idioms.
3. **Synthesize Dedicated Translation Directives**:
   - Target pronoun system strategy (e.g. in Vietnamese: intimate vs distant first/second/third person pronouns, modern colloquial vs period-specific address).
   - Lexical density recommendations: ratio of formal/classical lexicon (e.g. Sino-Vietnamese) to colloquial vocabulary.
   - Preservation guidelines for comedic timing, puns, and dramatic climaxes.
4. **Interactive Customization**: Deliver the generated style profile to the user for collaborative review, fine-tuning, and project-wide persistence.

## 2. Methodology & Benchmark Calibration

- **Optimal Sampling Size**: 3 opening narrative chapters (~1,000 - 1,200 characters per chapter).
- **Validation Criteria**:
  - Distinguishes between Rom-Com / Slice-of-Life, Dark Fantasy / Cyberpunk, Classical Wuxia / Cultivation, Mystery / Noir, and High Fantasy.
  - Generates clear, non-conflicting prompt instructions for downstream Single-Pass and Dual-Agent translation pipelines.

## 3. Tool Invocations

### app_auto_scout_style
Scans the project's opening narrative chapters and outputs a customized Literary Style Profile.

Arguments:
- `project_id` (string): Target novel project identifier.
- `source_lang` (string): Source language (e.g. Japanese, Chinese, English).
- `target_lang` (string): Target language (e.g. Vietnamese, English).

### app_upsert_style_profile
Persists or updates a user-defined or AI-scouted style profile in the project's Style Guide.
