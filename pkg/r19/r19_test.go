package r19

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMaskerBasicAndAge(t *testing.T) {
	cfg := DefaultConfig()
	masker := NewMasker(nil)

	src := "Cô bé 14 tuổi bị trói trong căn phòng xích lại, một cảnh tượng 赤裸 và tràn đầy 呻吟."
	res := masker.Mask(src, cfg)

	if !strings.Contains(res.MaskedText, "<r:age_") {
		t.Fatalf("expected age tag in masked text, got: %s", res.MaskedText)
	}
	if !strings.Contains(res.MaskedText, "<r:") {
		t.Fatalf("expected term tag in masked text, got: %s", res.MaskedText)
	}

	if len(res.Entries) < 3 {
		t.Fatalf("expected at least 3 entries, got: %d", len(res.Entries))
	}
}

func TestMaskerMultiLingualAny(t *testing.T) {
	cfg := DefaultConfig()
	masker := NewMasker(nil)

	// Test JP: 中出し, 12歳
	jpText := "彼女は12歳で、彼に中出しされた。"
	jpRes := masker.Mask(jpText, cfg)
	if !strings.Contains(jpRes.MaskedText, "<r:age_") || !strings.Contains(jpRes.MaskedText, "<r:") {
		t.Errorf("expected JP tags, got: %s", jpRes.MaskedText)
	}

	// Test KR: 질내사정, 14살
	krText := "소녀는 14살이고 질내사정 당했다."
	krRes := masker.Mask(krText, cfg)
	if !strings.Contains(krRes.MaskedText, "<r:age_") || !strings.Contains(krRes.MaskedText, "<r:") {
		t.Errorf("expected KR tags, got: %s", krRes.MaskedText)
	}

	// Test EN: creampie, 15 years old
	enText := "She is 15 years old and experienced a creampie."
	enRes := masker.Mask(enText, cfg)
	if !strings.Contains(enRes.MaskedText, "<r:age_") || !strings.Contains(enRes.MaskedText, "<r:") {
		t.Errorf("expected EN tags, got: %s", enRes.MaskedText)
	}
}

func TestCustomTermsFileLoader(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "custom_r19.txt")
	content := "tentacle = xúc tu ma quái\n마법소녀 = ma pháp thiếu nữ\n"
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("write file err: %v", err)
	}

	engine := NewEngine(DefaultConfig(), nil)
	if err := engine.LoadCustomTerms(filePath); err != nil {
		t.Fatalf("load custom terms err: %v", err)
	}

	res, _ := engine.Prepare("A tentacle caught the 마법소녀.")
	if !strings.Contains(res.MaskedText, "<r:") {
		t.Errorf("expected custom term masked, got: %s", res.MaskedText)
	}

	restored := engine.Restore(res.MaskedText, res, false)
	if !strings.Contains(restored, "xúc tu ma quái") || !strings.Contains(restored, "ma pháp thiếu nữ") {
		t.Errorf("expected custom terms restored, got: %s", restored)
	}
}

func TestSanitizerAndExports(t *testing.T) {
	engine := NewEngine(DefaultConfig(), nil)

	rawContext := "Cảnh tượng 赤裸 và 呻吟 tràn ngập khắp nơi."
	sanitized := engine.SanitizeContext(rawContext)
	if strings.Contains(sanitized, "赤裸") || strings.Contains(sanitized, "呻吟") {
		t.Errorf("expected sensitive terms camouflaged, got: %s", sanitized)
	}
	if !strings.Contains(sanitized, "[...]") {
		t.Errorf("expected [...] replacement, got: %s", sanitized)
	}

	stripped := engine.StripTerms(rawContext)
	if strings.Contains(stripped, "赤裸") || strings.Contains(stripped, "呻吟") {
		t.Errorf("expected sensitive terms stripped, got: %s", stripped)
	}

	var tsvBuf strings.Builder
	if err := engine.ExportToTSV(&tsvBuf); err != nil {
		t.Fatalf("export to tsv failed: %v", err)
	}
	if !strings.Contains(tsvBuf.String(), "中出し\txuất tinh trong") {
		t.Errorf("expected TSV export to contain term, got output length %d", tsvBuf.Len())
	}

	var jsonBuf strings.Builder
	if err := engine.ExportToJSON(&jsonBuf); err != nil {
		t.Fatalf("export to json failed: %v", err)
	}
	if !strings.Contains(jsonBuf.String(), "中出し") {
		t.Errorf("expected JSON export to contain term")
	}
}

func TestRestorerFuzzyAndCapitalization(t *testing.T) {
	restorer := NewRestorer()
	entries := []Entry{
		{ID: 1, Source: "赤裸", Translation: "trần trụi", Tag: "<r:1/>"},
		{ID: 2, Source: "呻吟", Translation: "tiếng rên rỉ", Tag: "<r:2/>"},
	}

	input := "Cô gái đang <r:1/>. <r:2/> vang lên trong đêm. Người đàn ông nhìn thấy [r:1] lần nữa."
	restored := restorer.Restore(input, entries, false)

	if !strings.Contains(restored, "Cô gái đang trần trụi.") {
		t.Errorf("expected inline restoration, got: %s", restored)
	}
	if !strings.Contains(restored, "Tiếng rên rỉ vang lên") {
		t.Errorf("expected capitalized restoration at sentence start, got: %s", restored)
	}
	if !strings.Contains(restored, "nhìn thấy trần trụi lần nữa") {
		t.Errorf("expected fuzzy bracket [r:1] restoration, got: %s", restored)
	}
}

func TestEngineWithSafetyFallback(t *testing.T) {
	engine := NewEngine(DefaultConfig(), nil)

	primaryGen := func(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
		return "", errors.New("400 Bad Request: blocked by safety filter harm_category")
	}

	uncensoredGen := func(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
		return "Cô bé <r:age_1/> đã <r:2/> dữ dội.", nil
	}

	text := "Cô bé 14 tuổi đã 呻吟 dữ dội."
	result, err := engine.ExecuteWithBypass(context.Background(), primaryGen, uncensoredGen, "sys", text, false)
	if err != nil {
		t.Fatalf("expected fallback success, got err: %v", err)
	}

	if !strings.Contains(result, "14 tuổi") {
		t.Errorf("expected age restored, got: %s", result)
	}
	if !strings.Contains(result, "rên rỉ") {
		t.Errorf("expected term restored, got: %s", result)
	}
}

func TestMaskerEnglishTarget(t *testing.T) {
	cfg := DefaultConfig()
	cfg.TargetLang = "English"
	masker := NewMasker(nil)

	enText := "She was 15 years old and his penis touched her breasts."
	res := masker.Mask(enText, cfg)

	// Age should be formatted as "15 years old"
	for _, entry := range res.Entries {
		if entry.Category == CategoryUnderage {
			if entry.Translation != "15 years old" {
				t.Fatalf("expected English age '15 years old', got '%s'", entry.Translation)
			}
		}
		// English terms should NOT be translated to Vietnamese when target is English
		if entry.Source == "penis" && entry.Translation != "penis" {
			t.Fatalf("expected English term 'penis' preserved, got '%s'", entry.Translation)
		}
	}
}

func BenchmarkMasker_453Terms(b *testing.B) {
	cfg := DefaultConfig()
	masker := NewMasker(nil)
	sample := "彼女は14歳で、彼に中出しされた。Cô bé 14 tuổi một cảnh tượng 赤裸 và 呻吟 trong đêm. A tentacle intercourse penetration occurred."

	
	for b.Loop() {
		_ = masker.Mask(sample, cfg)
	}
}

