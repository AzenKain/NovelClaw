package novelclaw

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"novelclaw/internal/dtos"
	"novelclaw/pkg/llm"
)

func marshalArgs(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}

// SlashCommandResult represents the output of a parsed slash command execution.
type SlashCommandResult struct {
	Handled       bool
	Content       string
	ActionPayload *dtos.AppActionPayload
	Notice        string
}

// ParseAndExecuteSlashCommand parses the input text. If it starts with '/', it executes the command.
func (e *Engine) ParseAndExecuteSlashCommand(ctx context.Context, projectID, threadID, text string) (*SlashCommandResult, error) {
	trimmed := strings.TrimSpace(text)
	if !strings.HasPrefix(trimmed, "/") {
		return &SlashCommandResult{Handled: false}, nil
	}

	parts := strings.SplitN(trimmed, " ", 2)
	cmd := strings.ToLower(parts[0])
	args := ""
	if len(parts) > 1 {
		args = strings.TrimSpace(parts[1])
	}

	switch cmd {
	case "/compact":
		resp, err := e.compactor.CompactThread(ctx, threadID, true)
		if err != nil {
			return nil, fmt.Errorf("context compaction failed: %w", err)
		}
		return &SlashCommandResult{
			Handled: true,
			Content: fmt.Sprintf("⚡ Successfully compacted conversation context: Archived %d messages, current context footprint ~%d tokens.",
				resp.ArchivedMessages, resp.CompactedTokens),
			Notice: resp.SummaryNotice,
		}, nil

	case "/clear":
		if err := e.store.ClearNovelClawThreadMessages(ctx, threadID); err != nil {
			return nil, fmt.Errorf("clear message history failed: %w", err)
		}
		return &SlashCommandResult{
			Handled: true,
			Content: "🧹 All messages in this thread have been cleared.",
		}, nil

	case "/term":
		// Syntax: /term <source> = <target> [category]
		if args == "" || !strings.Contains(args, "=") {
			return &SlashCommandResult{
				Handled: true,
				Content: "⚠️ Syntax: `/term <source_term> = <target_term> [category]`\nExample: `/term Tsukuyomi = Moon Reader proper_name`",
			}, nil
		}
		sub := strings.SplitN(args, "=", 2)
		src := strings.TrimSpace(sub[0])
		rest := strings.TrimSpace(sub[1])
		tgt := rest
		cat := "general"

		fields := strings.Fields(rest)
		if len(fields) > 1 && (fields[len(fields)-1] == "proper_name" || fields[len(fields)-1] == "skill" || fields[len(fields)-1] == "realm" || fields[len(fields)-1] == "item" || fields[len(fields)-1] == "location") {
			cat = fields[len(fields)-1]
			tgt = strings.TrimSpace(strings.TrimSuffix(rest, cat))
		}

		argsBytes := marshalArgs(map[string]any{
			"source_term": src,
			"target_term": tgt,
			"category":    cat,
		})
		resText, payload, err := e.toolsExec.Execute(ctx, projectID, llm.ToolCall{
			Function: struct {
				Name      string `json:"name"`
				Arguments string `json:"arguments"`
			}{
				Name:      "app_add_glossary_term",
				Arguments: argsBytes,
			},
		})
		if err != nil {
			return nil, err
		}
		return &SlashCommandResult{
			Handled:       true,
			Content:       resText,
			ActionPayload: payload,
		}, nil

	case "/char":
		// Syntax: /char <A> -> <B>: <call_as> [self/tự xưng: <self_call_as>]  (Vietnamese alias kept for input compatibility)
		re := regexp.MustCompile(`(?i)^(.*?)\s*->\s*(.*?):\s*([^,;]+)(?:,\s*(?:tự xưng|self|as):\s*(.*))?$`)
		m := re.FindStringSubmatch(args)
		if len(m) < 4 {
			return &SlashCommandResult{
				Handled: true,
				Content: "⚠️ Syntax: `/char <Character A> -> <Character B>: <Call As> [, self: <Self Call As>]`\nExample: `/char Protagonist -> Mentor: Master, self: Disciple` or `/char Senior -> Junior: Junior, self: Senior`",
			}, nil
		}
		fromChar := strings.TrimSpace(m[1])
		toChar := strings.TrimSpace(m[2])
		callAs := strings.TrimSpace(m[3])
		selfCallAs := ""
		if len(m) >= 5 {
			selfCallAs = strings.TrimSpace(m[4])
		}

		argsBytes := marshalArgs(map[string]any{
			"from_char":     fromChar,
			"to_char":       toChar,
			"call_as":       callAs,
			"self_call_as":  selfCallAs,
			"volume_index":  1,
			"since_chapter": 1,
		})
		resText, payload, err := e.toolsExec.Execute(ctx, projectID, llm.ToolCall{
			Function: struct {
				Name      string `json:"name"`
				Arguments string `json:"arguments"`
			}{
				Name:      "app_set_relation_temporal",
				Arguments: argsBytes,
			},
		})
		if err != nil {
			return nil, err
		}
		return &SlashCommandResult{
			Handled:       true,
			Content:       resText,
			ActionPayload: payload,
		}, nil

	case "/steer":
		if args == "" {
			return &SlashCommandResult{
				Handled: true,
				Content: "⚠️ Syntax: `/steer <direct style directive>`\nExample: `/steer Translate upcoming dialogue with a sharper, more sarcastic tone.`",
			}, nil
		}
		argsBytes := marshalArgs(map[string]any{
			"style_guide":      args,
			"enable_hot_patch": true,
		})
		resText, payload, err := e.toolsExec.Execute(ctx, projectID, llm.ToolCall{
			Function: struct {
				Name      string `json:"name"`
				Arguments string `json:"arguments"`
			}{
				Name:      "app_update_translation_config",
				Arguments: argsBytes,
			},
		})
		if err != nil {
			return nil, err
		}
		return &SlashCommandResult{
			Handled:       true,
			Content:       "🎯 " + resText,
			ActionPayload: payload,
		}, nil

	case "/soul":
		if args == "" {
			return &SlashCommandResult{
				Handled: true,
				Content: "⚠️ Syntax: `/soul <persona_id>`\nExample: `/soul soul_neko_assistant` or `/soul soul_tieu_mai_wuxia`",
			}, nil
		}
		argsBytes := marshalArgs(map[string]any{
			"soul_id": args,
		})
		resText, payload, err := e.toolsExec.Execute(ctx, projectID, llm.ToolCall{
			Function: struct {
				Name      string `json:"name"`
				Arguments string `json:"arguments"`
			}{
				Name:      "app_set_agent_soul",
				Arguments: argsBytes,
			},
		})
		if err != nil {
			return nil, err
		}
		return &SlashCommandResult{
			Handled:       true,
			Content:       "🐱 " + resText,
			ActionPayload: payload,
		}, nil

	case "/rollback":
		chapIdx := int64(1)
		if args != "" {
			if n, err := strconv.ParseInt(args, 10, 64); err == nil && n > 0 {
				chapIdx = n
			}
		}
		argsBytes := marshalArgs(map[string]any{
			"chapter_index":   chapIdx,
			"checkpoint_type": "raw",
		})
		resText, payload, err := e.toolsExec.Execute(ctx, projectID, llm.ToolCall{
			Function: struct {
				Name      string `json:"name"`
				Arguments string `json:"arguments"`
			}{
				Name:      "app_rollback_checkpoint",
				Arguments: argsBytes,
			},
		})
		if err != nil {
			return nil, err
		}
		return &SlashCommandResult{
			Handled:       true,
			Content:       "⏪ " + resText,
			ActionPayload: payload,
		}, nil

	case "/pause":
		resText, payload, err := e.toolsExec.Execute(ctx, projectID, llm.ToolCall{
			Function: struct {
				Name      string `json:"name"`
				Arguments string `json:"arguments"`
			}{
				Name: "app_pause_translation",
			},
		})
		if err != nil {
			return nil, err
		}
		return &SlashCommandResult{
			Handled:       true,
			Content:       "⏸️ " + resText,
			ActionPayload: payload,
		}, nil

	case "/resume":
		resText, payload, err := e.toolsExec.Execute(ctx, projectID, llm.ToolCall{
			Function: struct {
				Name      string `json:"name"`
				Arguments string `json:"arguments"`
			}{
				Name: "app_resume_translation",
			},
		})
		if err != nil {
			return nil, err
		}
		return &SlashCommandResult{
			Handled:       true,
			Content:       "▶️ " + resText,
			ActionPayload: payload,
		}, nil

	case "/abort":
		resText, payload, err := e.toolsExec.Execute(ctx, projectID, llm.ToolCall{
			Function: struct {
				Name      string `json:"name"`
				Arguments string `json:"arguments"`
			}{
				Name: "app_abort_translation",
			},
		})
		if err != nil {
			return nil, err
		}
		return &SlashCommandResult{
			Handled:       true,
			Content:       "🛑 " + resText,
			ActionPayload: payload,
		}, nil

	case "/learn":
		argsBytes := marshalArgs(map[string]any{
			"notes": args,
		})
		resText, payload, err := e.toolsExec.Execute(ctx, projectID, llm.ToolCall{
			Function: struct {
				Name      string `json:"name"`
				Arguments string `json:"arguments"`
			}{
				Name:      "app_trigger_learn_evolution",
				Arguments: argsBytes,
			},
		})
		if err != nil {
			return nil, err
		}
		return &SlashCommandResult{
			Handled:       true,
			Content:       "🧠 " + resText,
			ActionPayload: payload,
		}, nil

	case "/help":
		helpText := `### 🐾 NovelClaw Slash Commands (/commands):
- **/compact**: Instantly compacts conversation context (frees up context window).
- **/clear**: Clears all message history in this thread.
- **/term <source> = <target> [category]**: Quickly adds a term into the L3 Domain Glossary.
- **/char <A> -> <B>: <call_as> [, self: <call>]**: Sets up character address relationship.
- **/steer <directive>**: Sends direct runtime style guidance (Hot-patching).
- **/soul <soul_id>**: Switches the active AI companion persona.
- **/rollback [chapter]**: Reverts chapter to raw checkpoint before translation.
- **/pause**: Pauses active translation pipeline.
- **/resume**: Resumes paused translation pipeline.
- **/abort**: Emergency aborts translation process.
- **/learn [notes]**: Triggers autonomous evolutionary learning from editor text edits.
- **/help**: Shows this help table.

💡 *You can also chat in any natural language! NovelClaw understands your intent and automatically invokes the corresponding studio tools.*`
		return &SlashCommandResult{
			Handled: true,
			Content: helpText,
		}, nil
	}

	return &SlashCommandResult{Handled: false}, nil
}
