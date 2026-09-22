package novelclaw

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"novelclaw/internal/dtos"
	"novelclaw/internal/gen/sqlc"
	"novelclaw/pkg/llm"
	"novelclaw/pkg/storage"
)

// AppActionEmitter is an interface to broadcast actions dispatched by tools.
type AppActionEmitter interface {
	EmitAction(payload dtos.AppActionPayload)
}

// DefaultNoopEmitter is a fallback emitter when no event emitter is provided.
type DefaultNoopEmitter struct{}

func (e DefaultNoopEmitter) EmitAction(payload dtos.AppActionPayload) {}

// AppToolsHooks allows injecting custom service behaviors into the tool executor.
type AppToolsHooks struct {
	OnStartTranslation func(ctx context.Context, projectID string, startChap, endChap int64) error
	OnPauseTranslation func(ctx context.Context, projectID string) error
	OnResumeTranslation func(ctx context.Context, projectID string) error
	OnSoftStop         func(ctx context.Context, projectID string) error
	OnAbort            func(ctx context.Context, projectID string) error
	OnRollback         func(ctx context.Context, projectID string, chapterIndex int64, checkpointType string) error
	OnExportBook       func(ctx context.Context, projectID string, format string, includeCover bool, outputPath string) error
	OnTriggerLearn     func(ctx context.Context, projectID string, notes string) error
	OnRunBenchmark     func(ctx context.Context, projectID string, chapterIndices []int64, modes []string) error
	OnScanGraph         func(ctx context.Context, projectID string, startChap, endChap int64) error
	OnScanWorld         func(ctx context.Context, projectID string, startChap, endChap int64) error
	OnSetSoul           func(ctx context.Context, projectID, soulID string) error
}

// AppToolsExecutor manages the 28 Omni-App tools for NovelClaw.
type AppToolsExecutor struct {
	store   *storage.Storage
	emitter AppActionEmitter
	hooks   AppToolsHooks
}

// NewAppToolsExecutor creates a new AppToolsExecutor.
func NewAppToolsExecutor(store *storage.Storage, emitter AppActionEmitter, hooks AppToolsHooks) *AppToolsExecutor {
	if emitter == nil {
		emitter = DefaultNoopEmitter{}
	}
	return &AppToolsExecutor{
		store:   store,
		emitter: emitter,
		hooks:   hooks,
	}
}

// GetOmniAppToolDefinitions returns all 28 tool definitions for LLM function calling.
func GetOmniAppToolDefinitions() []llm.ToolDefinition {
	return []llm.ToolDefinition{
		// -------------------------------------------------------------
		// Group 1: Navigation & UI
		// -------------------------------------------------------------
		{
			Type: "function",
			Function: llm.FunctionSchema{
				Name:        "app_switch_tab",
				Description: "Switch active UI tab to specified view (workspace, graph, glossary, world_bible, benchmark)",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"tab": map[string]any{
							"type":        "string",
							"enum":        []string{"workspace", "graph", "glossary", "world_bible", "benchmark"},
							"description": "Target tab name to navigate to",
						},
					},
					"required": []string{"tab"},
				},
			},
		},
		{
			Type: "function",
			Function: llm.FunctionSchema{
				Name:        "app_select_volume",
				Description: "Select or filter chapter list to a specific volume",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"volume_index": map[string]any{
							"type":        "integer",
							"description": "Volume index number, or 0 to show all",
						},
						"volume_tag": map[string]any{
							"type":        "string",
							"description": "Optional volume tag label such as 'vol1', 'Volume 2'",
						},
					},
					"required": []string{"volume_index"},
				},
			},
		},
		{
			Type: "function",
			Function: llm.FunctionSchema{
				Name:        "app_select_chapter",
				Description: "Navigate reader and translation workspace to a specific chapter",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"chapter_index": map[string]any{
							"type":        "integer",
							"description": "Chapter index number (1-based index)",
						},
					},
					"required": []string{"chapter_index"},
				},
			},
		},
		{
			Type: "function",
			Function: llm.FunctionSchema{
				Name:        "app_open_modal",
				Description: "Open application dialog modal (settings, export, import, project_manager, style_scout)",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"modal_name": map[string]any{
							"type":        "string",
							"enum":        []string{"settings", "export", "import", "project_manager", "style_scout"},
							"description": "Modal identifier to open",
						},
					},
					"required": []string{"modal_name"},
				},
			},
		},
		{
			Type: "function",
			Function: llm.FunctionSchema{
				Name:        "app_scroll_to_text",
				Description: "Auto-scroll reader document viewport to passage containing keyword",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"keyword": map[string]any{
							"type":        "string",
							"description": "Keyword or sentence excerpt to scroll towards",
						},
					},
					"required": []string{"keyword"},
				},
			},
		},

		// -------------------------------------------------------------
		// Group 2: Temporal Character Graph (L2)
		// -------------------------------------------------------------
		{
			Type: "function",
			Function: llm.FunctionSchema{
				Name:        "app_create_character",
				Description: "Create a new character entity in the L2 Character Graph",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"name": map[string]any{
							"type":        "string",
							"description": "Primary character name",
						},
						"aliases": map[string]any{
							"type":        "array",
							"items":       map[string]any{"type": "string"},
							"description": "Aliases, titles, or secondary designations",
						},
						"role": map[string]any{
							"type":        "string",
							"description": "Character narrative role (protagonist, antagonist, supporting, etc.)",
						},
						"gender": map[string]any{
							"type":        "string",
							"description": "Gender designation (male, female, other, unknown)",
						},
						"description": map[string]any{
							"type":        "string",
							"description": "Brief biography and role description",
						},
					},
					"required": []string{"name"},
				},
			},
		},
		{
			Type: "function",
			Function: llm.FunctionSchema{
				Name:        "app_set_relation_temporal",
				Description: "Set or update temporal address relationship between characters for a volume or starting from a chapter",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"volume_index": map[string]any{
							"type":        "integer",
							"description": "Volume index (1, 2, 3...). Set to 0 or 1 for series-wide default",
						},
						"from_char": map[string]any{
							"type":        "string",
							"description": "Calling character name (speaker)",
						},
						"to_char": map[string]any{
							"type":        "string",
							"description": "Addressed character name (listener)",
						},
						"call_as": map[string]any{
							"type":        "string",
							"description": "What from_char calls to_char (e.g. Master, Brother, Milady)",
						},
						"self_call_as": map[string]any{
							"type":        "string",
							"description": "What from_char refers to oneself when speaking to to_char (e.g. Disciple, I, Junior Brother)",
						},
						"tone": map[string]any{
							"type":        "string",
							"description": "Relational tone nuance (respectful, intimate, hostile, haughty)",
						},
						"since_chapter": map[string]any{
							"type":        "integer",
							"description": "Effective starting chapter index (e.g. chapter 25 onwards)",
						},
					},
					"required": []string{"from_char", "to_char", "call_as"},
				},
			},
		},
		{
			Type: "function",
			Function: llm.FunctionSchema{
				Name:        "app_scan_character_graph",
				Description: "Trigger AI scan to extract characters and relationships across chapter range",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"start_chapter": map[string]any{
							"type":        "integer",
							"description": "Starting chapter index for scan",
						},
						"end_chapter": map[string]any{
							"type":        "integer",
							"description": "Ending chapter index for scan",
						},
					},
					"required": []string{"start_chapter", "end_chapter"},
				},
			},
		},

		// -------------------------------------------------------------
		// Group 3: Universal World Bible Control
		// -------------------------------------------------------------
		{
			Type: "function",
			Function: llm.FunctionSchema{
				Name:        "app_scan_and_build_world",
				Description: "Scan and identify novel genre, synthesizing taxonomy categories and world lorebook entries",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"start_chapter": map[string]any{
							"type":        "integer",
							"description": "Starting chapter index for scan",
						},
						"end_chapter": map[string]any{
							"type":        "integer",
							"description": "Ending chapter index for scan",
						},
					},
					"required": []string{"start_chapter", "end_chapter"},
				},
			},
		},
		{
			Type: "function",
			Function: llm.FunctionSchema{
				Name:        "app_upsert_world_category",
				Description: "Upsert world encyclopedia category (e.g. megacorps, magic_schools, court_ranks, cultivation_realms)",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"slug": map[string]any{
							"type":        "string",
							"description": "Slug identifier for category (e.g. megacorps, items, locations)",
						},
						"name": map[string]any{
							"type":        "string",
							"description": "Display name for category (e.g. Megacorporations, Artifacts, Locations)",
						},
						"icon": map[string]any{
							"type":        "string",
							"description": "Lucide icon identifier (e.g. Shield, Building, Globe, Zap)",
						},
						"description": map[string]any{
							"type":        "string",
							"description": "Description of world category purpose and scope",
						},
					},
					"required": []string{"slug", "name"},
				},
			},
		},
		{
			Type: "function",
			Function: llm.FunctionSchema{
				Name:        "app_upsert_world_entry",
				Description: "Upsert world lorebook entry with metadata and custom attributes",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"category_slug": map[string]any{
							"type":        "string",
							"description": "Category slug containing this lorebook entry",
						},
						"name": map[string]any{
							"type":        "string",
							"description": "Lore entry name (e.g. Arasaka, Sword of Light)",
						},
						"aliases": map[string]any{
							"type":        "array",
							"items":       map[string]any{"type": "string"},
							"description": "Alternative names, acronyms, or translation aliases",
						},
						"summary": map[string]any{
							"type":        "string",
							"description": "Concise summary of the lore entry",
						},
						"attributes_json": map[string]any{
							"type":        "string",
							"description": "JSON string of custom attributes (e.g. {\"founder\": \"Saburo\", \"hq\": \"Tokyo\"})",
						},
					},
					"required": []string{"category_slug", "name"},
				},
			},
		},
		{
			Type: "function",
			Function: llm.FunctionSchema{
				Name:        "app_delete_world_entry",
				Description: "Delete an entry from the World Bible",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"entry_id": map[string]any{
							"type":        "string",
							"description": "ID or name of entry to delete",
						},
					},
					"required": []string{"entry_id"},
				},
			},
		},

		// -------------------------------------------------------------
		// Group 4: Glossary & Terminology (L3)
		// -------------------------------------------------------------
		{
			Type: "function",
			Function: llm.FunctionSchema{
				Name:        "app_add_glossary_term",
				Description: "Add term to the L3 Domain Glossary",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"source_term": map[string]any{
							"type":        "string",
							"description": "Source original term (Chinese, Japanese, English, etc.)",
						},
						"target_term": map[string]any{
							"type":        "string",
							"description": "Target term translation",
						},
						"category": map[string]any{
							"type":        "string",
							"description": "Term category (proper_name, skill, realm, item, location, general)",
						},
						"notes": map[string]any{
							"type":        "string",
							"description": "Usage notes and context rules",
						},
					},
					"required": []string{"source_term", "target_term"},
				},
			},
		},
		{
			Type: "function",
			Function: llm.FunctionSchema{
				Name:        "app_bulk_import_glossary",
				Description: "Bulk import glossary items from JSON array",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"terms_json": map[string]any{
							"type":        "string",
							"description": "JSON array containing items [{source_term, target_term, category, notes}]",
						},
					},
					"required": []string{"terms_json"},
				},
			},
		},
		{
			Type: "function",
			Function: llm.FunctionSchema{
				Name:        "app_delete_glossary_term",
				Description: "Delete term from project glossary",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"source_term": map[string]any{
							"type":        "string",
							"description": "Source term to delete",
						},
					},
					"required": []string{"source_term"},
				},
			},
		},

		// -------------------------------------------------------------
		// Group 5: Pipeline & Model Routing
		// -------------------------------------------------------------
		{
			Type: "function",
			Function: llm.FunctionSchema{
				Name:        "app_update_translation_config",
				Description: "Update translation configuration mode, post-filters, hot-patching, agentic RAG, and style guide",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"mode": map[string]any{
							"type":        "string",
							"enum":        []string{"single_pass", "hierarchical_3pass", "concurrent_dual_agent", "swarm_arc_parallel", "SinglePass", "Hierarchical3Pass", "ConcurrentDualAgent"},
							"description": "Translation pipeline mode",
						},
						"enable_hot_patch": map[string]any{
							"type":        "boolean",
							"description": "Enable runtime hot-patching overrides",
						},
						"enable_r19": map[string]any{
							"type":        "boolean",
							"description": "Enable R19 post-filtering compliance",
						},
						"enable_agentic_rag": map[string]any{
							"type":        "boolean",
							"description": "Enable Agentic RAG tool invocation loop",
						},
						"style_guide": map[string]any{
							"type":        "string",
							"description": "Global style directive guide",
						},
					},
				},
			},
		},
		{
			Type: "function",
			Function: llm.FunctionSchema{
				Name:        "app_configure_llm_provider",
				Description: "Configure LLM provider settings or switch default model",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"provider": map[string]any{
							"type":        "string",
							"description": "Provider name (gemini, openai, deepseek, etc.)",
						},
						"model_name": map[string]any{
							"type":        "string",
							"description": "Model identifier (e.g. gemini-2.5-flash, gpt-4o)",
						},
						"api_key": map[string]any{
							"type":        "string",
							"description": "Optional new API key string",
						},
						"is_default": map[string]any{
							"type":        "boolean",
							"description": "Set as default system model",
						},
					},
					"required": []string{"provider", "model_name"},
				},
			},
		},
		{
			Type: "function",
			Function: llm.FunctionSchema{
				Name:        "app_set_agent_soul",
				Description: "Switch active NovelClaw AI companion persona (Soul)",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"soul_id": map[string]any{
							"type":        "string",
							"description": "Soul persona ID (e.g. soul_neko_assistant, soul_tieu_mai_wuxia)",
						},
					},
					"required": []string{"soul_id"},
				},
			},
		},
		{
			Type: "function",
			Function: llm.FunctionSchema{
				Name:        "app_toggle_skill",
				Description: "Toggle specialized skill in Skills library",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"skill_id": map[string]any{
							"type":        "string",
							"description": "Skill identifier (e.g. app_graph_master, app_world_builder)",
						},
						"enabled": map[string]any{
							"type":        "boolean",
							"description": "True to enable, False to disable",
						},
					},
					"required": []string{"skill_id", "enabled"},
				},
			},
		},

		// -------------------------------------------------------------
		// Group 6: Translation Execution & Rollback
		// -------------------------------------------------------------
		{
			Type: "function",
			Function: llm.FunctionSchema{
				Name:        "app_start_translation",
				Description: "Execute translation pipeline from start chapter to end chapter",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"start_chapter": map[string]any{
							"type":        "integer",
							"description": "Starting chapter index to translate",
						},
						"end_chapter": map[string]any{
							"type":        "integer",
							"description": "Ending chapter index (equals start_chapter for single chapter)",
						},
					},
					"required": []string{"start_chapter"},
				},
			},
		},
		{
			Type: "function",
			Function: llm.FunctionSchema{
				Name:        "app_pause_translation",
				Description: "Pause currently running translation process",
				Parameters: map[string]any{
					"type":       "object",
					"properties": map[string]any{},
				},
			},
		},
		{
			Type: "function",
			Function: llm.FunctionSchema{
				Name:        "app_resume_translation",
				Description: "Resume paused translation process",
				Parameters: map[string]any{
					"type":       "object",
					"properties": map[string]any{},
				},
			},
		},
		{
			Type: "function",
			Function: llm.FunctionSchema{
				Name:        "app_soft_stop_translation",
				Description: "Request safe soft-stop after completing current paragraph or chapter",
				Parameters: map[string]any{
					"type":       "object",
					"properties": map[string]any{},
				},
			},
		},
		{
			Type: "function",
			Function: llm.FunctionSchema{
				Name:        "app_abort_translation",
				Description: "Trigger immediate emergency abort on active translation pipeline",
				Parameters: map[string]any{
					"type":       "object",
					"properties": map[string]any{},
				},
			},
		},
		{
			Type: "function",
			Function: llm.FunctionSchema{
				Name:        "app_rollback_checkpoint",
				Description: "Rollback chapter translation content to a previous checkpoint",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"chapter_index": map[string]any{
							"type":        "integer",
							"description": "Chapter index to rollback",
						},
						"checkpoint_type": map[string]any{
							"type":        "string",
							"description": "Checkpoint type (draft, raw, audited, etc.)",
						},
					},
					"required": []string{"chapter_index"},
				},
			},
		},

		// -------------------------------------------------------------
		// Group 7: Export, Evolution & Benchmark
		// -------------------------------------------------------------
		{
			Type: "function",
			Function: llm.FunctionSchema{
				Name:        "app_export_book",
				Description: "Export novel project to epub, pdf, docx, or txt document",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"format": map[string]any{
							"type":        "string",
							"enum":        []string{"epub", "pdf", "docx", "txt"},
							"description": "Target file format export",
						},
						"include_cover": map[string]any{
							"type":        "boolean",
							"description": "Whether to include book cover image",
						},
						"output_path": map[string]any{
							"type":        "string",
							"description": "Optional output file path destination",
						},
					},
					"required": []string{"format"},
				},
			},
		},
		{
			Type: "function",
			Function: llm.FunctionSchema{
				Name:        "app_trigger_learn_evolution",
				Description: "Trigger Reflexion Engine to learn from human edits and evolve skill guidelines",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"notes": map[string]any{
							"type":        "string",
							"description": "Notes and key insights to reflect upon",
						},
					},
				},
			},
		},
		{
			Type: "function",
			Function: llm.FunctionSchema{
				Name:        "app_run_benchmark",
				Description: "Execute translation benchmark to compare speed, quality, and cost across modes",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"chapter_indices": map[string]any{
							"type":        "array",
							"items":       map[string]any{"type": "integer"},
							"description": "List of chapter indices to benchmark",
						},
						"modes": map[string]any{
							"type":        "array",
							"items":       map[string]any{"type": "string"},
							"description": "List of translation modes to compare (SinglePass, Hierarchical3Pass, ConcurrentDualAgent)",
						},
					},
					"required": []string{"chapter_indices"},
				},
			},
		},
	}
}

// Execute dispatches and runs an Omni-App tool call, updating storage and emitting UI actions.
func (e *AppToolsExecutor) Execute(ctx context.Context, projectID string, call llm.ToolCall) (string, *dtos.AppActionPayload, error) {
	name := call.Function.Name
	rawArgs := call.Function.Arguments

	var args map[string]any
	if strings.TrimSpace(rawArgs) != "" {
		if err := json.Unmarshal([]byte(rawArgs), &args); err != nil {
			return "", nil, fmt.Errorf("failed to decode parameters for tool %s: %w", name, err)
		}
	}
	if args == nil {
		args = make(map[string]any)
	}

	var payload *dtos.AppActionPayload
	var resultText string
	var err error

	switch name {
	// =========================================================================
	// Group 1: Navigation & UI
	// =========================================================================
	case "app_switch_tab":
		tab, _ := args["tab"].(string)
		if tab == "" {
			tab = "workspace"
		}
		payload = &dtos.AppActionPayload{
			Action:      "switch_tab",
			Category:    "navigation",
			Description: fmt.Sprintf("Switched to tab %s", tab),
			Data:        map[string]any{"tab": tab},
		}
		resultText = fmt.Sprintf("Switched the interface to the '%s' tab.", tab)

	case "app_select_volume":
		volIdx := getInt64(args["volume_index"])
		volTag, _ := args["volume_tag"].(string)
		payload = &dtos.AppActionPayload{
			Action:      "select_volume",
			Category:    "navigation",
			Description: fmt.Sprintf("Selected volume %d (%s)", volIdx, volTag),
			Data: map[string]any{
				"volume_index": volIdx,
				"volume_tag":   volTag,
			},
		}
		resultText = fmt.Sprintf("Selected volume %d (%s).", volIdx, volTag)

	case "app_select_chapter":
		chapIdx := getInt64(args["chapter_index"])
		payload = &dtos.AppActionPayload{
			Action:      "select_chapter",
			Category:    "navigation",
			Description: fmt.Sprintf("Navigated to chapter %d", chapIdx),
			Data:        map[string]any{"chapter_index": chapIdx},
		}
		resultText = fmt.Sprintf("Switched the view to chapter %d.", chapIdx)

	case "app_open_modal":
		modal, _ := args["modal_name"].(string)
		payload = &dtos.AppActionPayload{
			Action:      "open_modal",
			Category:    "navigation",
			Description: fmt.Sprintf("Open the %s dialog", modal),
			Data:        map[string]any{"modal_name": modal},
		}
		resultText = fmt.Sprintf("Opened the '%s' dialog.", modal)

	case "app_scroll_to_text":
		kw, _ := args["keyword"].(string)
		payload = &dtos.AppActionPayload{
			Action:      "scroll_to_text",
			Category:    "navigation",
			Description: fmt.Sprintf("Scroll to the excerpt containing: %s", kw),
			Data:        map[string]any{"keyword": kw},
		}
		resultText = fmt.Sprintf("Scrolled to the excerpt containing '%s'.", kw)

	// =========================================================================
	// Group 2: Temporal Character Graph
	// =========================================================================
	case "app_create_character":
		name := getString(args["name"])
		if name == "" {
			return "", nil, fmt.Errorf("missing character name")
		}
		aliases := getStringSlice(args["aliases"])
		role := getString(args["role"])
		gender := getString(args["gender"])
		desc := getString(args["description"])

		if e.store != nil && projectID != "" {
			ent := storage.Entity{
				ID:               fmt.Sprintf("ent_%s_%s", projectID, strings.ToLower(name)),
				ProjectID:        projectID,
				Name:             name,
				Aliases:          aliases,
				Category:         "character",
				Gender:           gender,
				Role:             role,
				FirstSeenChapter: 1,
				Metadata:         map[string]any{"description": desc},
			}
			if err = e.store.UpsertEntity(ctx, ent); err != nil {
				return "", nil, fmt.Errorf("failed to save character to the database: %w", err)
			}
		}
		payload = &dtos.AppActionPayload{
			Action:      "character_created",
			Category:    "graph",
			Description: fmt.Sprintf("Add character: %s (%s)", name, role),
			Data: map[string]any{
				"name":        name,
				"aliases":     aliases,
				"role":        role,
				"gender":      gender,
				"description": desc,
			},
		}
		resultText = fmt.Sprintf("Added character '%s' (role: %s) to the L2 Graph.", name, role)

	case "app_set_relation_temporal":
		fromChar := getString(args["from_char"])
		toChar := getString(args["to_char"])
		callAs := getString(args["call_as"])
		selfCallAs := getString(args["self_call_as"])
		tone := getString(args["tone"])
		volIdx := getInt64(args["volume_index"])
		sinceChap := getInt64(args["since_chapter"])

		if fromChar == "" || toChar == "" || callAs == "" {
			return "", nil, fmt.Errorf("from_char, to_char and call_as are all required")
		}
		if sinceChap <= 0 {
			sinceChap = 1
		}

		if e.store != nil && projectID != "" {
			relID := fmt.Sprintf("rel_%s_%s_%s_ch%d", projectID, strings.ToLower(fromChar), strings.ToLower(toChar), sinceChap)
			relParams := sqlc.UpsertRelationParams{
				ID:           relID,
				ProjectID:    projectID,
				FromChar:     fromChar,
				ToChar:       toChar,
				CallAs:       callAs,
				SelfCallAs:   selfCallAs,
				SinceChapter: sinceChap,
				Tone:         tone,
			}
			if err = e.store.UpsertRelation(ctx, relParams); err != nil {
				return "", nil, fmt.Errorf("failed to save relationship to the database: %w", err)
			}
		}

		payload = &dtos.AppActionPayload{
			Action:      "relation_updated",
			Category:    "graph",
			Description: fmt.Sprintf("Set relationship: %s calls %s '%s' (from ch.%d, Volume %d)", fromChar, toChar, callAs, sinceChap, volIdx),
			Data: map[string]any{
				"volume_index":  volIdx,
				"from_char":     fromChar,
				"to_char":       toChar,
				"call_as":       callAs,
				"self_call_as":  selfCallAs,
				"tone":          tone,
				"since_chapter": sinceChap,
			},
		}
		resultText = fmt.Sprintf("Configured address relation: '%s' calls '%s' as '%s', self-reference '%s' (from chapter %d, Volume %d).",
			fromChar, toChar, callAs, selfCallAs, sinceChap, volIdx)

	case "app_scan_character_graph":
		startCh := getInt64(args["start_chapter"])
		endCh := getInt64(args["end_chapter"])
		if startCh <= 0 {
			startCh = 1
		}
		if endCh < startCh {
			endCh = startCh
		}
		if e.hooks.OnScanGraph != nil {
			_ = e.hooks.OnScanGraph(ctx, projectID, startCh, endCh)
		}
		payload = &dtos.AppActionPayload{
			Action:      "graph_scan_triggered",
			Category:    "graph",
			Description: fmt.Sprintf("Scan character graph for chapters %d - %d", startCh, endCh),
			Data: map[string]any{
				"start_chapter": startCh,
				"end_chapter":   endCh,
			},
		}
		resultText = fmt.Sprintf("Started scanning and extracting characters and relationships from chapter %d to %d.", startCh, endCh)

	// =========================================================================
	// Group 3: Universal World Bible Control
	// =========================================================================
	case "app_scan_and_build_world":
		startCh := getInt64(args["start_chapter"])
		endCh := getInt64(args["end_chapter"])
		if startCh <= 0 {
			startCh = 1
		}
		if endCh < startCh {
			endCh = startCh
		}
		if e.hooks.OnScanWorld != nil {
			_ = e.hooks.OnScanWorld(ctx, projectID, startCh, endCh)
		}
		payload = &dtos.AppActionPayload{
			Action:      "world_bible_scanned",
			Category:    "world_bible",
			Description: fmt.Sprintf("Scan and build the World Bible from chapters %d - %d", startCh, endCh),
			Data: map[string]any{
				"start_chapter": startCh,
				"end_chapter":   endCh,
			},
		}
		resultText = fmt.Sprintf("Started scanning and building World Bible entities for chapters %d - %d.", startCh, endCh)

	case "app_upsert_world_category":
		slug := getString(args["slug"])
		catName := getString(args["name"])
		icon := getString(args["icon"])
		desc := getString(args["description"])
		if slug == "" || catName == "" {
			return "", nil, fmt.Errorf("slug and name are required for a world category")
		}

		if e.store != nil && projectID != "" {
			catID := fmt.Sprintf("cat_%s_%s", projectID, slug)
			_, err = e.store.UpsertWorldCategory(ctx, sqlc.UpsertWorldCategoryParams{
				ID:           catID,
				ProjectID:    projectID,
				Slug:         slug,
				Name:         catName,
				Icon:         icon,
				Description:  desc,
				DisplayOrder: 0,
			})
			if err != nil {
				return "", nil, fmt.Errorf("failed to save world category to the database: %w", err)
			}
		}

		payload = &dtos.AppActionPayload{
			Action:      "world_category_upserted",
			Category:    "world_bible",
			Description: fmt.Sprintf("World category: %s (%s)", catName, slug),
			Data: map[string]any{
				"slug":        slug,
				"name":        catName,
				"icon":        icon,
				"description": desc,
			},
		}
		resultText = fmt.Sprintf("Saved world-building category '%s' (slug: %s).", catName, slug)

	case "app_upsert_world_entry":
		catSlug := getString(args["category_slug"])
		entryName := getString(args["name"])
		aliases := getStringSlice(args["aliases"])
		summary := getString(args["summary"])
		attrJSON := getString(args["attributes_json"])

		if catSlug == "" || entryName == "" {
			return "", nil, fmt.Errorf("category_slug and name are required for a world entity")
		}

		var attrs map[string]any
		if attrJSON != "" {
			_ = json.Unmarshal([]byte(attrJSON), &attrs)
		}
		if attrs == nil {
			attrs = make(map[string]any)
		}
		attrs["summary"] = summary

		if e.store != nil && projectID != "" {
			cat, catErr := e.store.GetWorldCategoryBySlug(ctx, projectID, catSlug)
			catID := cat.ID
			if catErr != nil || catID == "" {
				catID = fmt.Sprintf("cat_%s_%s", projectID, catSlug)
				_, _ = e.store.UpsertWorldCategory(ctx, sqlc.UpsertWorldCategoryParams{
					ID:           catID,
					ProjectID:    projectID,
					Slug:         catSlug,
					Name:         catSlug,
					Icon:         "Boxes",
					Description:  "Auto-created from Chat",
					DisplayOrder: 0,
				})
			}

			entryID := fmt.Sprintf("we_%s_%s", projectID, strings.ToLower(strings.ReplaceAll(entryName, " ", "_")))
			aliasBytes, _ := json.Marshal(aliases)
			_, err = e.store.UpsertWorldEntry(ctx, sqlc.UpsertWorldEntryParams{
				ID:                 entryID,
				ProjectID:          projectID,
				CategoryID:         catID,
				Name:               entryName,
				AliasesJson:        string(aliasBytes),
				Summary:            summary,
				FullDescription:    summary,
				AttributesJson:     attrJSON,
				DiscoveredBy:       "chat_command",
				SourceChapterIndex: 1,
				IsVerified:         1,
			})
			if err != nil {
				return "", nil, fmt.Errorf("failed to save world entity to the database: %w", err)
			}
		}

		payload = &dtos.AppActionPayload{
			Action:      "world_entry_upserted",
			Category:    "world_bible",
			Description: fmt.Sprintf("World entity: %s [%s]", entryName, catSlug),
			Data: map[string]any{
				"category_slug":   catSlug,
				"name":            entryName,
				"aliases":         aliases,
				"summary":         summary,
				"attributes_json": attrJSON,
			},
		}
		resultText = fmt.Sprintf("Saved world lore entry '%s' into category '%s'.", entryName, catSlug)

	case "app_delete_world_entry":
		entryID := getString(args["entry_id"])
		if entryID == "" {
			return "", nil, fmt.Errorf("missing entry_id to delete")
		}
		if e.store != nil {
			_ = e.store.DeleteWorldEntry(ctx, entryID)
		}
		payload = &dtos.AppActionPayload{
			Action:      "world_entry_deleted",
			Category:    "world_bible",
			Description: fmt.Sprintf("Delete world entity: %s", entryID),
			Data:        map[string]any{"entry_id": entryID},
		}
		resultText = fmt.Sprintf("Deleted world entry '%s' from the World Bible.", entryID)

	// =========================================================================
	// Group 4: Glossary & Terminology (L3)
	// =========================================================================
	case "app_add_glossary_term":
		src := getString(args["source_term"])
		tgt := getString(args["target_term"])
		cat := getString(args["category"])
		notes := getString(args["notes"])

		if src == "" || tgt == "" {
			return "", nil, fmt.Errorf("source_term and target_term are required")
		}
		if cat == "" {
			cat = "general"
		}

		if e.store != nil && projectID != "" {
			gParams := sqlc.UpsertGlossaryTermParams{
				ID:         "gloss_" + strings.ReplaceAll(uuid.New().String(), "-", "")[:12],
				ProjectID:  projectID,
				SourceTerm: src,
				TargetTerm: tgt,
				Category:   cat,
				Notes:      notes,
			}
			if err = e.store.UpsertGlossaryTerm(ctx, gParams); err != nil {
				return "", nil, fmt.Errorf("failed to save glossary term to the database: %w", err)
			}
		}

		payload = &dtos.AppActionPayload{
			Action:      "glossary_term_added",
			Category:    "glossary",
			Description: fmt.Sprintf("Add glossary term: %s = %s (%s)", src, tgt, cat),
			Data: map[string]any{
				"source_term": src,
				"target_term": tgt,
				"category":    cat,
				"notes":       notes,
			},
		}
		resultText = fmt.Sprintf("Added glossary term '%s' = '%s' (category: %s) to L3 glossary.", src, tgt, cat)

	case "app_bulk_import_glossary":
		termsJSON := getString(args["terms_json"])
		var rawList []struct {
			SourceTerm string `json:"source_term"`
			TargetTerm string `json:"target_term"`
			Category   string `json:"category"`
			Notes      string `json:"notes"`
		}
		if err := json.Unmarshal([]byte(termsJSON), &rawList); err != nil {
			return "", nil, fmt.Errorf("failed to decode terms_json: %w", err)
		}

		imported := 0
		if e.store != nil && projectID != "" {
			for _, item := range rawList {
				if item.SourceTerm == "" || item.TargetTerm == "" {
					continue
				}
				cat := item.Category
				if cat == "" {
					cat = "general"
				}
				_ = e.store.UpsertGlossaryTerm(ctx, sqlc.UpsertGlossaryTermParams{
					ID:         "gloss_" + strings.ReplaceAll(uuid.New().String(), "-", "")[:12],
					ProjectID:  projectID,
					SourceTerm: item.SourceTerm,
					TargetTerm: item.TargetTerm,
					Category:   cat,
					Notes:      item.Notes,
				})
				imported++
			}
		}

		payload = &dtos.AppActionPayload{
			Action:      "glossary_bulk_imported",
			Category:    "glossary",
			Description: fmt.Sprintf("Batch-imported %d glossary terms", imported),
			Data:        map[string]any{"count": imported},
		}
		resultText = fmt.Sprintf("Successfully imported %d glossary terms L3.", imported)

	case "app_delete_glossary_term":
		src := getString(args["source_term"])
		if src == "" {
			return "", nil, fmt.Errorf("missing source_term to delete")
		}
		if e.store != nil && projectID != "" {
			// Find term and delete
			terms, _ := e.store.ListGlossaryByProject(ctx, projectID)
			for _, t := range terms {
				if strings.EqualFold(t.SourceTerm, src) {
					_ = e.store.DeleteGlossaryTerm(ctx, t.ID)
					break
				}
			}
		}

		payload = &dtos.AppActionPayload{
			Action:      "glossary_term_deleted",
			Category:    "glossary",
			Description: fmt.Sprintf("Delete glossary term: %s", src),
			Data:        map[string]any{"source_term": src},
		}
		resultText = fmt.Sprintf("Deleted term '%s' from the project glossary.", src)

	// =========================================================================
	// Group 5: Pipeline & Model Routing
	// =========================================================================
	case "app_update_translation_config":
		mode := getString(args["mode"])
		modeLower := strings.ToLower(mode)
		if modeLower == "singlepass" || modeLower == "single_pass" {
			mode = "single_pass"
		} else if modeLower == "hierarchical3pass" || modeLower == "hierarchical_3pass" {
			mode = "hierarchical_3pass"
		} else if modeLower == "concurrentdualagent" || modeLower == "concurrent_dual_agent" {
			mode = "concurrent_dual_agent"
		} else if modeLower == "swarmarc" || modeLower == "swarm_arc_parallel" {
			mode = "swarm_arc_parallel"
		}

		enableHotPatch := getBool(args["enable_hot_patch"])
		enableR19 := getBool(args["enable_r19"])
		enableAgenticRAG := getBool(args["enable_agentic_rag"])
		styleGuide := getString(args["style_guide"])

		payload = &dtos.AppActionPayload{
			Action:      "update_translation_config",
			Category:    "pipeline",
			Description: fmt.Sprintf("Translation config: Mode=%s, HotPatch=%v, R19=%v, AgenticRAG=%v", mode, enableHotPatch, enableR19, enableAgenticRAG),
			Data: map[string]any{
				"mode":               mode,
				"enable_hot_patch":   enableHotPatch,
				"enable_r19":         enableR19,
				"enable_agentic_rag": enableAgenticRAG,
				"style_guide":        styleGuide,
			},
		}
		resultText = fmt.Sprintf("Updated translation configuration: Mode=%s, HotPatch=%v, R19=%v, AgenticRAG=%v.", mode, enableHotPatch, enableR19, enableAgenticRAG)

	case "app_configure_llm_provider":
		provider := getString(args["provider"])
		modelName := getString(args["model_name"])
		apiKey := getString(args["api_key"])
		isDefault := getBool(args["is_default"])

		if e.store != nil && (provider != "" || modelName != "") {
			apiURL := "https://api.openai.com/v1"
			if strings.EqualFold(provider, "gemini") {
				apiURL = "https://generativelanguage.googleapis.com/v1beta"
			} else if strings.EqualFold(provider, "deepseek") {
				apiURL = "https://api.deepseek.com/v1"
			}
			_ = e.store.SaveLLMConfig(ctx, storage.LLMConfig{
				ProviderName:   provider,
				ModelName:      modelName,
				Token:          apiKey,
				ApiURL:         apiURL,
				IsActive:       true,
				IsDefault:      isDefault,
				TimeoutSeconds: 120,
				MaxRetries:     3,
			})
		}

		payload = &dtos.AppActionPayload{
			Action:      "llm_configured",
			Category:    "pipeline",
			Description: fmt.Sprintf("LLM config: %s (%s)", provider, modelName),
			Data: map[string]any{
				"provider":   provider,
				"model_name": modelName,
				"api_key":    apiKey,
				"is_default": isDefault,
			},
		}
		resultText = fmt.Sprintf("Configured LLM provider '%s' with model '%s' (default=%v).", provider, modelName, isDefault)

	case "app_set_agent_soul":
		soulID := getString(args["soul_id"])
		if e.hooks.OnSetSoul != nil && projectID != "" && soulID != "" {
			_ = e.hooks.OnSetSoul(ctx, projectID, soulID)
		}
		payload = &dtos.AppActionPayload{
			Action:      "soul_changed",
			Category:    "pipeline",
			Description: fmt.Sprintf("Change Persona Soul: %s", soulID),
			Data:        map[string]any{"soul_id": soulID},
		}
		resultText = fmt.Sprintf("Switched the NovelClaw Persona Soul to '%s'.", soulID)

	case "app_toggle_skill":
		skillID := getString(args["skill_id"])
		enabled := getBool(args["enabled"])
		payload = &dtos.AppActionPayload{
			Action:      "skill_toggled",
			Category:    "pipeline",
			Description: fmt.Sprintf("Skill %s: %v", skillID, enabled),
			Data: map[string]any{
				"skill_id": skillID,
				"enabled":  enabled,
			},
		}
		statusText := "disabled"
		if enabled {
			statusText = "enabled"
		}
		resultText = fmt.Sprintf("Skill '%s' has been %s.", skillID, statusText)

	// =========================================================================
	// Group 6: Translation Execution & Rollback
	// =========================================================================
	case "app_start_translation":
		startCh := getInt64(args["start_chapter"])
		endCh := getInt64(args["end_chapter"])
		if startCh <= 0 {
			startCh = 1
		}
		if endCh < startCh {
			endCh = startCh
		}

		if e.hooks.OnStartTranslation != nil {
			_ = e.hooks.OnStartTranslation(ctx, projectID, startCh, endCh)
		}

		payload = &dtos.AppActionPayload{
			Action:      "start_translation",
			Category:    "execution",
			Description: fmt.Sprintf("Start translating chapters %d - %d", startCh, endCh),
			Data: map[string]any{
				"start_chapter": startCh,
				"end_chapter":   endCh,
			},
		}
		resultText = fmt.Sprintf("Started the translation pipeline from chapter %d to %d.", startCh, endCh)

	case "app_pause_translation":
		if e.hooks.OnPauseTranslation != nil {
			_ = e.hooks.OnPauseTranslation(ctx, projectID)
		}
		payload = &dtos.AppActionPayload{
			Action:      "pause_translation",
			Category:    "execution",
			Description: "Pause the translation pipeline",
			Data:        map[string]any{},
		}
		resultText = "Sent the command to pause the translation pipeline."

	case "app_resume_translation":
		if e.hooks.OnResumeTranslation != nil {
			_ = e.hooks.OnResumeTranslation(ctx, projectID)
		}
		payload = &dtos.AppActionPayload{
			Action:      "resume_translation",
			Category:    "execution",
			Description: "Resume the translation pipeline",
			Data:        map[string]any{},
		}
		resultText = "Sent the command to resume the translation pipeline."

	case "app_soft_stop_translation":
		if e.hooks.OnSoftStop != nil {
			_ = e.hooks.OnSoftStop(ctx, projectID)
		}
		payload = &dtos.AppActionPayload{
			Action:      "soft_stop_translation",
			Category:    "execution",
			Description: "Soft stop after the current sentence",
			Data:        map[string]any{},
		}
		resultText = "Triggered a soft stop (Soft Stop). The pipeline will finish the current sentence and halt safely."

	case "app_abort_translation":
		if e.hooks.OnAbort != nil {
			_ = e.hooks.OnAbort(ctx, projectID)
		}
		payload = &dtos.AppActionPayload{
			Action:      "abort_translation",
			Category:    "execution",
			Description: "Emergency abort of the translation pipeline",
			Data:        map[string]any{},
		}
		resultText = "Triggered an emergency abort (Hard Abort). The entire translation pipeline was stopped immediately."

	case "app_rollback_checkpoint":
		chapIdx := getInt64(args["chapter_index"])
		cpType := getString(args["checkpoint_type"])
		if cpType == "" {
			cpType = "raw"
		}
		if e.hooks.OnRollback != nil {
			_ = e.hooks.OnRollback(ctx, projectID, chapIdx, cpType)
		}
		payload = &dtos.AppActionPayload{
			Action:      "rollback_checkpoint",
			Category:    "execution",
			Description: fmt.Sprintf("Roll back chapter %d to checkpoint %s", chapIdx, cpType),
			Data: map[string]any{
				"chapter_index":   chapIdx,
				"checkpoint_type": cpType,
			},
		}
		resultText = fmt.Sprintf("Rolled back chapter %d to checkpoint '%s'.", chapIdx, cpType)

	// =========================================================================
	// Group 7: Export, Evolution & Benchmark
	// =========================================================================
	case "app_export_book":
		fmtStr := getString(args["format"])
		if fmtStr == "" {
			fmtStr = "epub"
		}
		incCover := getBool(args["include_cover"])
		outPath := getString(args["output_path"])

		if e.hooks.OnExportBook != nil {
			_ = e.hooks.OnExportBook(ctx, projectID, fmtStr, incCover, outPath)
		}

		payload = &dtos.AppActionPayload{
			Action:      "export_book",
			Category:    "export",
			Description: fmt.Sprintf("Export book as %s", fmtStr),
			Data: map[string]any{
				"format":        fmtStr,
				"include_cover": incCover,
				"output_path":   outPath,
			},
		}
		resultText = fmt.Sprintf("Triggered book export in %s format (cover image=%v).", strings.ToUpper(fmtStr), incCover)

	case "app_trigger_learn_evolution":
		notes := getString(args["notes"])
		if e.hooks.OnTriggerLearn != nil {
			_ = e.hooks.OnTriggerLearn(ctx, projectID, notes)
		}
		payload = &dtos.AppActionPayload{
			Action:      "trigger_learn",
			Category:    "export",
			Description: fmt.Sprintf("Self-evolution learning: %s", notes),
			Data:        map[string]any{"notes": notes},
		}
		resultText = fmt.Sprintf("Triggered the Reflexion Engine to learn from text edits. Notes: %s", notes)

	case "app_run_benchmark":
		var chapIndices []int64
		if rawChaps, ok := args["chapter_indices"].([]any); ok {
			for _, item := range rawChaps {
				chapIndices = append(chapIndices, getInt64(item))
			}
		}
		modes := getStringSlice(args["modes"])
		if len(modes) == 0 {
			modes = []string{"SinglePass", "Hierarchical3Pass", "ConcurrentDualAgent"}
		}

		if e.hooks.OnRunBenchmark != nil {
			_ = e.hooks.OnRunBenchmark(ctx, projectID, chapIndices, modes)
		}

		payload = &dtos.AppActionPayload{
			Action:      "run_benchmark",
			Category:    "export",
			Description: fmt.Sprintf("Run benchmark on %d chapters", len(chapIndices)),
			Data: map[string]any{
				"chapter_indices": chapIndices,
				"modes":           modes,
			},
		}
		resultText = fmt.Sprintf("Launched translation benchmark comparison across modes %v for chapters %v.", modes, chapIndices)

	default:
		return "", nil, fmt.Errorf("unsupported tool: %s", name)
	}

	// Broadcast action to frontend via emitter
	if payload != nil && e.emitter != nil {
		e.emitter.EmitAction(*payload)
	}

	return resultText, payload, nil
}

// Helpers
func getString(v any) string {
	if s, ok := v.(string); ok {
		return strings.TrimSpace(s)
	}
	return ""
}

func getInt64(v any) int64 {
	switch n := v.(type) {
	case int:
		return int64(n)
	case int64:
		return n
	case float64:
		return int64(n)
	default:
		return 0
	}
}

func getBool(v any) bool {
	if b, ok := v.(bool); ok {
		return b
	}
	return false
}

func getStringSlice(v any) []string {
	var result []string
	if list, ok := v.([]any); ok {
		for _, item := range list {
			if s, ok := item.(string); ok && strings.TrimSpace(s) != "" {
				result = append(result, strings.TrimSpace(s))
			}
		}
	} else if sl, ok := v.([]string); ok {
		return sl
	}
	return result
}
