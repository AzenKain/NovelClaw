<!--
  Rewrite this file for every release. It is the BODY of the GitHub release
  (the title is set from the pushed tag by .github/workflows/build.yml).
  The version itself lives in pkg/appmeta/version.go — see AGENT.md §8.
  This comment is not rendered by GitHub.
-->

The first public release of NovelClaw — a desktop studio for translating and
publishing novels without losing the things that make a book a book.

It is not a wrapper around a translation API. It models the editorial process a
publishing house actually uses and puts an autonomous agent in the middle of it:
you import a book, the agent profiles its style, builds a knowledge base of
characters and terminology, translates it chapter by chapter, critiques its own
draft, and exports a readable ebook with the original artwork and structure
intact.

## Highlights

- **Full pipeline in one app** — import, profile, translate, review, export,
  without leaving the window.
- **Pronoun and relationship consistency** — the relationship between two
  characters is stored as a graph edge with a chapter timestamp, so a pair who
  address each other one way in chapter 1 and another way in chapter 10 stay
  consistent.
- **Layered memory** — memory is organised by narrative function and retrieval
  is bounded by chapter index, so a search can only ever see the past.
- **Formatting survives the round trip** — EPUB, MOBI, AZW3, PDF, DOCX, ODT,
  FB2, RTF, TXT, Markdown, HTML, CBZ and CBR import and export with
  illustrations, footnotes, cover art and tables of contents preserved.
- **Cost control** — the prompt is assembled in a stable prefix order so
  providers can cache it, and context selection is a trie lookup rather than a
  similarity search.
- **Agent skills and souls** — the capability playbooks and personas ship
  embedded in the binary, so a fresh install is never an empty agent. They are
  materialised on first launch and can be edited afterwards.
- **Bring your own model** — any OpenAI-compatible endpoint, plus native
  OpenAI, Gemini, Anthropic, DeepSeek, ViLao and Ollama presets, with a
  fallback chain.
- **Five interface languages** — English, Vietnamese, Japanese, Chinese, Korean.

## Downloads

| Platform | Architecture | File | Notes |
| --- | --- | --- | --- |
| Windows | x64 | `novelclaw-windows-amd64.exe` | Portable — just run it. `novelclaw-windows-amd64-installer.exe` installs it instead. |
| Windows | ARM64 | `novelclaw-windows-arm64.exe` | Portable. Installer variant also provided. |
| macOS | Apple Silicon | `novelclaw-macos-arm64.dmg` | Drag into Applications. `novelclaw-darwin-arm64.zip` is the archive the updater uses. |
| macOS | Intel | `novelclaw-macos-amd64.dmg` | Drag into Applications. `novelclaw-darwin-amd64.zip` is the archive the updater uses. |
| Linux | x86_64 | `novelclaw-linux-amd64` | Portable bare binary — make it executable and run it. |
| Linux | x86_64 | `novelclaw-linux.AppImage`, `.deb`, `.rpm`, `.pkg.tar.zst` | Package formats for your distribution. |

Nothing else is required — no Go, no Node, no compiler.

> On macOS the app is not notarised yet. If Gatekeeper refuses the first
> launch, right-click the app and choose **Open**, or clear the quarantine
> attribute: `xattr -dr com.apple.quarantine /Applications/NovelClaw.app`.

## First run

A setup wizard asks for your interface language, an LLM provider and its API
key, and a fallback chain. After that, import a book and start translating.

Your data lives in one of two places:

| Platform | Data directory |
| --- | --- |
| Windows | `./` — portable, next to the executable |
| macOS / Linux | `~/.novelclaw/` |

That directory holds the SQLite database, your soul personas, the agent skill
playbooks, and the distilled skills learned from your edits. It is created and
migrated automatically on startup and is never touched by an update.

## Updates

NovelClaw updates itself from this Releases page. The startup check is silent:
if you are already current, or the network is down, you will not see anything.
When an update does exist, a window offers it to you and the binary is swapped
in place on restart. You can also check manually from
**Settings → About → Check for updates**.

## Verifying your download

Every asset is listed with its SHA-256 hash in the `SHA256SUMS` file attached to
this release. To verify a download:

```bash
# Linux / macOS
sha256sum --ignore-missing -c SHA256SUMS     # or: shasum -a 256 -c SHA256SUMS

# Windows (PowerShell)
(Get-FileHash .\novelclaw-windows-amd64.exe -Algorithm SHA256).Hash
```

Compare the output against the matching line in `SHA256SUMS`. The in-app updater
performs this check automatically and refuses to install an asset that does not
match.

## Known limitations

- macOS builds are unsigned and unnotarised (see the note above).
- The ARM64 Windows build is not covered by CI smoke tests yet.
- Automatic update requires the release to keep the asset naming above — the
  updater selects the first asset whose name contains both the platform and the
  architecture token.

## Feedback

Bug reports and feature requests: <https://github.com/AzenKain/NovelClaw/issues>
