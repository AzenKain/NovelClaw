package stylestat

import (
	"testing"
)

func TestAnalyzerCleanText(t *testing.T) {
	analyzer := NewAnalyzer()
	text := `Nayuta bước vào phòng làm việc của Itsuki. Cô mỉm cười nhẹ nhàng và đặt tách trà nóng lên bàn.
Itsuki ngẩng đầu lên cảm ơn cô em gái nhỏ, ánh mắt tràn đầy sự dịu dàng. Cả hai cùng nhau bắt đầu một ngày làm việc mới.`

	glossary := map[string]string{
		"Nayuta": "Nayuta",
		"Itsuki": "Itsuki",
	}

	report := analyzer.Compute(text, glossary)
	if report.DuplicateSentenceCount != 0 {
		t.Errorf("expected 0 duplicate sentences, got: %d", report.DuplicateSentenceCount)
	}
	if report.GlossaryAdherenceRate < 1.0 {
		t.Errorf("expected 100%% glossary adherence, got: %f", report.GlossaryAdherenceRate)
	}
	if report.QualityScore < 9.5 {
		t.Errorf("expected high quality score, got: %f", report.QualityScore)
	}
}

func TestAnalyzerRepetitiveTics(t *testing.T) {
	analyzer := NewAnalyzer()
	text := `Không kìm được mà quay đầu nhìn lại. Không kìm được mà quay đầu nhìn lại.
Không kìm được mà thở dài trong lòng. Không kìm được mà bước nhanh hơn.
Không kìm được mà quay đầu nhìn lại.`

	report := analyzer.Compute(text, nil)
	if report.DuplicateSentenceCount < 1 {
		t.Errorf("expected at least 1 duplicate sentence, got: %d", report.DuplicateSentenceCount)
	}
	if len(report.RepetitivePhrases) == 0 {
		t.Errorf("expected detected repetitive phrases, got 0")
	}
	if report.QualityScore >= 8.0 {
		t.Errorf("expected penalized quality score, got: %f", report.QualityScore)
	}
}
