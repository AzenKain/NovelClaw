package voice

import (
	"context"
	"strings"
	"testing"
)

func TestVoiceRegistry_MatchingAndFormatting(t *testing.T) {
	ctx := context.Background()
	reg := NewRegistry()
	projectID := "proj-voice-test"

	v1 := CharacterVoice{
		CharacterID:   "char-tieu-viem",
		Name:          "Tiêu Viêm",
		Aliases:       []string{"Nham Kiêu", "Dược Vương Đệ Tử"},
		VoiceTone:     "Kiên định, khẩu khí sắc sảo, tự tin",
		DialogueRules: []string{"Xưng 'ta' với kẻ thù", "Gọi Dược Lão là 'Sư phụ'"},
		Catchphrases:  []string{"Ba mươi năm hà đông"},
	}

	v2 := CharacterVoice{
		CharacterID:   "char-my-do-toa",
		Name:          "Mỹ Đỗ Toa",
		Aliases:       []string{"Thải Lân", "Nữ Vương"},
		VoiceTone:     "Lãnh diễm, cao ngạo, uy nghiêm",
		DialogueRules: []string{"Luôn tự xưng 'Bản vương'"},
	}

	reg.UpsertVoice(projectID, v1)
	reg.UpsertVoice(projectID, v2)

	// List
	all := reg.ListVoices(projectID)
	if len(all) != 2 {
		t.Fatalf("expected 2 voices, got %d", len(all))
	}

	// Match in text containing alias "Nham Kiêu"
	text := "Đột nhiên, Nham Kiêu rút ra Huyền Trọng Xích chắn ngang trước mặt."
	matched := reg.MatchVoicesInText(ctx, projectID, text)
	if len(matched) != 1 || matched[0].Name != "Tiêu Viêm" {
		t.Fatalf("expected 1 match for Tiêu Viêm via alias, got %v", matched)
	}

	// Match both
	textBoth := "Tiêu Viêm nhìn Nữ Vương và khẽ mỉm cười."
	matchedBoth := reg.MatchVoicesInText(ctx, projectID, textBoth)
	if len(matchedBoth) != 2 {
		t.Fatalf("expected 2 matches, got %d", len(matchedBoth))
	}

	// Format prompt
	prompt := FormatVoicePrompt(matchedBoth)
	if !strings.Contains(prompt, "[CHARACTER VOICE & DIALOGUE STYLE GUIDE]") {
		t.Errorf("prompt missing header: %s", prompt)
	}
	if !strings.Contains(prompt, "Tiêu Viêm") || !strings.Contains(prompt, "Mỹ Đỗ Toa") {
		t.Errorf("prompt missing characters: %s", prompt)
	}
	if !strings.Contains(prompt, "Bản vương") {
		t.Errorf("prompt missing dialogue rule: %s", prompt)
	}
}
