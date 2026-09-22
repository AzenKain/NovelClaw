# NovelClaw — Project Rules & Code Prohibitions

Binding rules for every AI agent, subagent, and developer working on the
`novelclaw` repository. Violating them is a serious architectural defect.

Scope: **architecture, storage, packaging, release, data safety.**
Translation quality rules (style, prompts, terminology) live with the skills
and personas under `skills/` and `souls/`, not here.

---

## 1. Architecture & module boundaries

```
main.go              → wiring only: bootstrap, DI, window, run
internal/services/   → Wails bindings bridging to pkg/*
internal/dtos/       → types crossing the IPC boundary (JSON tags required)
pkg/*                → pure logic; must NOT import wails (appmeta excepted)
frontend/src/        → React UI; frontend/bindings/ is generated code
db/schema|query      → SQL source of truth; internal/gen/sqlc is generated
```

**Prohibited**

- Putting business logic in `main.go`. It only wires things together.
- Importing `github.com/wailsapp/wails/v3/...` from `pkg/*`. The one
  exception is `pkg/appmeta`, which is the app integration layer by design.
  To emit events, accept an interface (see `NovelClawEventEmitter` in
  `pkg/novelclaw/engine.go`).
- Hand-editing generated files: `internal/gen/sqlc/**`, `frontend/bindings/**`.
  Change the SQL source and run `sqlc generate`, or run
  `wails3 generate bindings`.

---

## 2. Storage & paths — no hardcoded paths

Every data path **must** go through `pkg/paths`. Never write `"./data"`,
`"./souls"`, `"./skills"`, `"~/.novelclaw"`, or `"novelclaw.db"` anywhere else.

| OS | DataDir |
| --- | --- |
| Windows | `./` (portable, next to the executable) |
| macOS / Linux | `~/.novelclaw/` |

Use `paths.DB()`, `paths.Data()`, `paths.Souls()`, `paths.Skills()`,
`paths.SkillsDefault()`, `paths.SkillsEvolved()`, `paths.Resolve(p)`,
`paths.Ensure()`.

**Prohibited**

- Calling `os.UserHomeDir()` yourself to build a data path.
- Creating data files or directories outside the `paths.DataDir()` tree.
- Changing the data tree without updating `paths.Ensure()` and
  `appmeta.EnsureDataLayout()`.

---

## 3. Data migration — never destroy user data

Any change to where or how data is stored **requires** an idempotent
migration path that preserves existing data.

**Prohibited**

- Overwriting user data. Migration must be additive: `souls/` and `skills/`
  merge in missing files only and never replace existing ones; `data/` moves
  as a unit and only when the destination has no real files.
- Ignoring old directory names on a rename. This already happened once: the
  `~/.neko-novel` → `~/.novelclaw` rename split the vault key in two, making
  every stored API key undecryptable and hiding all saved LLM configs. Every
  data-directory rename must ship a migration (see `legacyKeyDirs` in
  `pkg/security/vault.go` and `migrateLegacyData` in `pkg/appmeta/layout.go`).
- Deleting a user's old files as part of a migration.

---

## 4. Default resources must be embedded

`skills/` and `souls/` **must be embedded** into the binary. The installer does
not copy these directories next to the executable, so a fresh install has to
materialise them on its own.

- Embed: `embed_defaults.go` (`//go:embed all:skills all:souls`) → `defaultsFS`.
- Seed: `appmeta.SeedSkillBaselineFromDefaults(defaultsFS)` in `main.go`.

**Prohibited**

- Adding a default resource (skill, soul, prompt, schema, font, template)
  without embedding it. An unembedded resource is missing on a fresh install.
- Writing a seeder that overwrites existing files — the user may have edited a
  soul or added a skill.
- Assuming `skills/` or `souls/` exists next to the executable.
- Dropping `SOUL.md` from the seed: without it the engine falls back to a
  generic three-line prompt.

---

## 5. Data safety & secrets

- API tokens are encrypted with AES-256-GCM via `pkg/security`. Never write
  plaintext tokens to the database or the logs.
- DTOs sent to the frontend must carry masked tokens (`dtos.MaskToken`). Never
  send a real token over IPC.
- One bad record must never break a whole query. `List*` functions must handle
  each row independently, log the failure, and skip the row instead of
  returning an error for everything (this already happened: a single
  undecryptable token hid the entire LLM config list).
- `pkg/security` keeps a **key ring**: it tries the primary key, then any
  legacy key, and transparently re-encrypts old rows under the new key. Do not
  remove this mechanism.

**Prohibited**

- Logging tokens, keys, or ciphertext.
- Returning a global error because a single row is corrupt.
- Minting a new vault key without first checking whether an older key exists.

---

## 6. No hardcoded business data

Code in `pkg/`, `internal/`, and `cmd/` must stay **generic** and independent
of any specific book.

**Prohibited**

- Hardcoding character names, series names, or story-specific terminology into
  logic. Entities must be loaded dynamically from the database (`entities`,
  `character_relations`).
- Using regex to "patch" translated content (rewording, restyling). Regex is
  for pure technical cleanup only: HTML/XML tags, code fences, encoding
  artifacts, stray SVG.
- Inventing data that does not exist (phantom entities, phantom fallback
  tiers). Lists and sequences must be derived from real database rows.
- Seeding sample data that looks like real user data.

**Test:** the code must work correctly on a brand-new novel with no edits.

---

## 7. Enums must stay in sync across FE and BE

Enum values that cross IPC (translation mode, trigger condition, provider
preset, …) need one canonical source, and both sides must agree.

**Prohibited**

- Naming a value differently on each side. This already happened: the frontend
  sent `swarm_arc` while the backend only accepted `swarm_arc_parallel`, so the
  mode was **silently ignored** and the run fell back to the default.
- Silent enum parse failures. Either normalise aliases or fail loudly; never
  swallow an unknown value and quietly run the default.
- Duplicating an enum list across frontend components — use a shared constant
  (`frontend/src/constants/providers.ts`).

---

## 8. Versioning & self-update

There is exactly **one** source of truth for the version:

```go
// pkg/appmeta/version.go
const Version = "1.0.0"
```

To cut a release: edit that constant, commit, tag `v<same-version>` and push
the tag. The GitHub tag must match the constant — app `1.2.3` ↔ tag `v1.2.3`.
`make sync-version` and CI both mirror the constant into `build/config.yml`
(`info.version`) **and** run `wails3 update build-assets` to regenerate
`build/darwin/Info.plist` / `build/windows/info.json` / `nfpm.yaml` — the
darwin Taskfile copies the committed plist verbatim into the `.app`, so editing
`config.yml` alone would ship the wrong bundle version. CI also fails the build
when the pushed tag disagrees with the constant.

`release/RELEASE_NOTES.md` is the **body** of the GitHub release (the title is
taken from the pushed tag). Rewrite it before pushing a tag — there is no
version field in it, and no `release.json`.

Asset naming is **load-bearing**, because the updater picks the first asset
whose name contains both the platform and the arch token:

| Artifact | Purpose |
| --- | --- |
| `novelclaw-linux-amd64`, `novelclaw-windows-amd64.exe`, `novelclaw-darwin-{amd64,arm64}.zip` | Updatable — must carry the arch token |
| `novelclaw-linux.AppImage` / `.deb` / `.rpm` / `.pkg.tar.zst`, `novelclaw-macos-*.dmg`, `*-installer.exe` | Human downloads — must NOT carry an arch token the matcher accepts |

**Prohibited**

- Adding `-ldflags` version injection, a `VERSION` env var, or a
  `release.json`-style version file. One constant, edited by hand.
- Letting `build/config.yml` `info.version` diverge from `appmeta.Version`.
- Renaming release assets without re-checking the updater's matcher (`.deb`,
  `.rpm`, `.AppImage` and `.dmg` are **not** skipped by it).
- Publishing a release without a `SHA256SUMS` covering every updatable asset.
- Hardcoding the version in a frontend component. Read it from
  `appmeta.GetAppInfo()`.
- Auto-update that shows errors or noise when the network is down or when the
  app is already current. The startup check must be **silent** on error and on
  "up to date", and only surface when an update actually exists.
- Making the manual "check for updates" button silent on failure — that path is
  user-initiated and **must** surface errors.

---

## 9. Frontend

- Do not edit `frontend/bindings/**`; it is generated.
- Do not hardcode display text. Use i18n (`frontend/src/locales/*`) and cover
  all five languages: `vi`, `en`, `ja`, `ko`, `zh`.
- Configuration state must come from the database through a service, never be
  reconstructed from `localStorage`.

**Prohibited**

- Treating `localStorage` as the source of truth for configuration.
- Adding an i18n key to only one language.
- Swallowing a service error and rendering an empty UI as if there were no data.

---

## 10. Testing & handoff

Before handing off, these **must** pass:

```bash
go build ./...
go vet ./...
go test ./...                     # ignore failures caused by sandboxed sockets
cd frontend && npm run build      # tsc + vite must pass
```

**Prohibited**

- Handing off with `tsc` or `go build` errors.
- Adding or changing a feature without a matching test (migration, seeding,
  security, and enum parsing all require tests).
- Making a test pass by loosening assertions or hardcoding data.
- Changing storage, migration, or security behaviour without documenting it.

---

## Pre-commit checklist

- [ ] Any hardcoded path, character name, or resource left behind?
- [ ] Are new default resources embedded?
- [ ] Is the migration idempotent and non-destructive to user data?
- [ ] Can one bad record break a list query?
- [ ] Do the FE and BE enums match?
- [ ] Were any generated files (`gen/sqlc`, `bindings`) edited by hand?
- [ ] Do `go build ./...`, `go vet ./...`, `go test ./...`, and `npm run build` pass?
