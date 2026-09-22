package websearch

import (
	"strings"
	"testing"
	"time"
)

// TestDeepReader_ExtractContent tests boilerplate removal and markdown structure preservation.
func TestDeepReader_ExtractContent(t *testing.T) {
	htmlContent := `
<!DOCTYPE html>
<html>
<head>
    <title>Test Light Novel Article</title>
    <style>body { font-family: sans-serif; }</style>
    <script>alert("ad popup");</script>
</head>
<body>
    <header>
        <nav>
            <ul>
                <li><a href="/">Home</a></li>
                <li><a href="/about">About</a></li>
            </ul>
        </nav>
    </header>

    <div class="sidebar advertisement">
        <p>Buy our cheap products now!</p>
    </div>

    <main>
        <article class="post-body">
            <h1>Tsundere (ツンデレ)</h1>
            <p>Tsundere describes a character personality development from harsh to sweet.</p>
            <h2>Origins</h2>
            <p>This term originated in the early 2000s Japanese otaku culture.</p>
            <ul>
                <li>Tsun: harsh, aloof</li>
                <li>Dere: sweet, softhearted</li>
            </ul>
            <blockquote>A quintessential character archetype in light novels.</blockquote>
        </article>
    </main>

    <div class="user-comments">
        <p>User123: Great article!</p>
    </div>

    <footer>
        <p>Copyright © 2026 NovelClaw. All rights reserved.</p>
    </footer>
</body>
</html>
`

	dr := NewDeepReader(5*time.Second, 2000)
	content, err := dr.ExtractContent(strings.NewReader(htmlContent))
	if err != nil {
		t.Fatalf("ExtractContent failed: %v", err)
	}

	if strings.Contains(content, "ad popup") {
		t.Errorf("content still contains script content")
	}
	if strings.Contains(content, "Buy our cheap products now") {
		t.Errorf("content still contains advertisement block")
	}
	if strings.Contains(content, "Copyright © 2026") {
		t.Errorf("content still contains footer block")
	}
	if strings.Contains(content, "User123: Great article") {
		t.Errorf("content still contains comment section")
	}

	if !strings.Contains(content, "Tsundere (ツンデレ)") {
		t.Errorf("content missing title heading")
	}
	if !strings.Contains(content, "Tsun: harsh, aloof") {
		t.Errorf("content missing list item")
	}
	if !strings.Contains(content, "A quintessential character archetype") {
		t.Errorf("content missing blockquote")
	}
}
