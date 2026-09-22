package main

import (
	"embed"
	"os"
	"path/filepath"

	"github.com/rs/zerolog/log"
	"github.com/wailsapp/wails/v3/pkg/application"

	"novelclaw/internal/services"
	"novelclaw/pkg/appmeta"
	"novelclaw/pkg/auditor"
	"novelclaw/pkg/paths"
	"novelclaw/pkg/skills"
	"novelclaw/pkg/soul"
	"novelclaw/pkg/storage"
	"novelclaw/pkg/voice"
	"novelclaw/pkg/worldbible"
)

//go:embed all:frontend/dist build/appicon.png
var assets embed.FS

// main initializes database storage, services, creates the application window and runs Wails v3.
func main() {
	checkAndElevate()

	// Canonical storage layout: Windows → ".", mac/Linux → ~/.novelclaw.
	// Creates the tree, migrates legacy cwd-relative data once, logs the
	// version/platform/data report.
	appmeta.EnsureDataLayout()

	// Materialise the shipped skill playbooks and soul personas. The
	// packaged installer does not place skills/ or souls/ next to the
	// executable, so without this a fresh install would run an agent with no
	// playbooks and only a generic persona. Existing user files are kept.
	appmeta.SeedSkillBaselineFromDefaults(defaultsFS)

	store, err := storage.OpenStorage(paths.DB())
	if err != nil {
		log.Fatal().Err(err).Str("db", paths.DB()).Msg("failed to initialize SQLite storage")
	}
	defer store.Close()

	soulCtrl := soul.NewController()
	skillReg := skills.NewRegistry(store)
	projectSvc := services.NewProjectService(store)
	graphSvc := services.NewGraphService(store)
	glossarySvc := services.NewGlossaryService(store)
	llmSvc := services.NewLLMService(store)
	wbMatcher := worldbible.NewMatcher(store)
	voiceRegistry := voice.NewRegistry()
	plotAuditor := auditor.NewPlotAuditor(store)

	transSvc := services.NewTranslationServiceWithSoul(store, soulCtrl)
	transSvc.SetSkillRegistry(skillReg)
	transSvc.SetWorldBibleMatcher(wbMatcher)
	transSvc.SetVoiceRegistry(voiceRegistry)
	transSvc.SetPlotAuditor(plotAuditor)

	wbSvc := services.NewWorldBibleService(store, skillReg, wbMatcher, voiceRegistry, plotAuditor)
	exportSvc := services.NewExportService(store)
	soulSvc := services.NewSoulService(soulCtrl)
	fsSvc := services.NewFSService()
	novelClawSvc := services.NewNovelClawService(store, soulCtrl, wbSvc)
	skillSvc := services.NewSkillService(skillReg)
	updateSvc := appmeta.NewUpdateService()

	app := application.New(application.Options{
		Name:        "novelclaw",
		Description: "NovelClaw - AI Novel Translation Studio",
		Services: []application.Service{
			application.NewService(fsSvc),
			application.NewService(projectSvc),
			application.NewService(graphSvc),
			application.NewService(glossarySvc),
			application.NewService(llmSvc),
			application.NewService(transSvc),
			application.NewService(exportSvc),
			application.NewService(soulSvc),
			application.NewService(novelClawSvc),
			application.NewService(skillSvc),
			application.NewService(wbSvc),
			application.NewService(updateSvc),
		},
		Assets: application.AssetOptions{
			Handler: services.NewAssetHandler(store, application.AssetFileServerFS(assets)),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
		Windows: application.WindowsOptions{
			WebviewUserDataPath: filepath.Join(os.Getenv("APPDATA"), "novelclaw"),
		},
	})

	if err := appmeta.InitUpdater(app); err != nil {
		log.Warn().Err(err).Msg("self-updater disabled")
	}
	appmeta.InstallUpdateMenu(app)
	// NOTE: the automatic startup update check is triggered by the frontend
	// once the UI is up (useAppStore.checkForUpdatesSilent). Do not also call
	// CheckSilent here — two concurrent checks race on the updater's staging
	// directory and can open the update window before the app is ready.

	// Window geometry is derived from the primary display's WORK AREA (not
	// its raw size): the work area already excludes taskbars, docks and
	// panels, so the window never opens partly off-screen.
	width, height := windowSizeForScreen(app)

	window := app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:  "NovelClaw",
		Width:  width,
		Height: height,
		// Floor values that match the layout's narrowest supported mode: the
		// header collapses to icon-only below 1024px, so anything smaller
		// would clip controls instead of adapting.
		MinWidth:  1024,
		MinHeight: 600,
		Mac: application.MacWindow{
			InvisibleTitleBarHeight: 50,
			Backdrop:                application.MacBackdropTranslucent,
			TitleBar:                application.MacTitleBarHiddenInset,
		},
		BackgroundColour: application.NewRGB(6, 7, 15),
		URL:              "/",
	})

	iconBytes, _ := assets.ReadFile("build/appicon.png")
	systemTray := app.SystemTray.New()
	systemTray.SetIcon(iconBytes)
	systemTray.SetTooltip("NovelClaw")

	// Attach the window to the system tray
	trayMenu := application.NewMenu()
	trayMenu.Add("Open").OnClick(func(ctx *application.Context) {
		window.Show()
	})
	trayMenu.AddSeparator()
	trayMenu.Add("Check for Updates…").OnClick(func(ctx *application.Context) {
		// Manual check — surfaces every state (up-to-date / available / error)
		// in the built-in updater window.
		appmeta.CheckForUpdates(app)
	})
	trayMenu.AddSeparator()
	trayMenu.Add("Quit").OnClick(func(ctx *application.Context) {
		if window != nil {
			window.Hide()
		}
		app.Quit()
	})

	systemTray.SetMenu(trayMenu)

	err = app.Run()
	if err != nil {
		log.Fatal().Err(err).Msg("novelclaw application exited with error")
	}
}

// windowSizeForScreen picks an initial window size from the primary display's
// work area, keeping a 16:9-ish shape and leaving a margin so the window is
// never flush against the screen edges. Falls back to 1280x720 when the
// screen manager is unavailable (e.g. headless or server builds).
func windowSizeForScreen(app *application.App) (int, int) {
	const (
		fallbackW = 1280
		fallbackH = 720
		// Never request a window taller/wider than this fraction of the work
		// area, so the window always leaves room to see the desktop behind it.
		maxFraction = 0.92
	)

	primary := app.Screen.GetPrimary()
	if primary == nil {
		return fallbackW, fallbackH
	}

	// Prefer the work area; fall back to the full size when the platform
	// reports a zero work area.
	availW := primary.WorkArea.Width
	availH := primary.WorkArea.Height
	if availW <= 0 || availH <= 0 {
		availW = primary.Size.Width
		availH = primary.Size.Height
	}
	if availW <= 0 || availH <= 0 {
		return fallbackW, fallbackH
	}

	maxW := int(float64(availW) * maxFraction)
	maxH := int(float64(availH) * maxFraction)

	width, height := fallbackW, fallbackH
	switch {
	case availH <= 800: // small laptop / low-res panel
		width, height = 1024, 576
	case availH <= 1080: // 1080p
		width, height = 1280, 720
	case availH <= 1440: // 1440p
		width, height = 1600, 900
	case availH <= 2160: // 4K
		width, height = 1920, 1080
	default: // ultrawide / very large
		width, height = 2560, 1440
	}

	// Clamp to the available area, but never below the layout's minimum.
	if width > maxW {
		width = maxW
	}
	if height > maxH {
		height = maxH
	}
	if width < 1024 {
		width = 1024
	}
	if height < 600 {
		height = 600
	}
	return width, height
}
