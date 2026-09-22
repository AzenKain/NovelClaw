package appmeta

import (
	"context"
	"fmt"
	"runtime"

	"github.com/rs/zerolog/log"
	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/updater"
	"github.com/wailsapp/wails/v3/pkg/updater/providers/github"
)

func InitUpdater(app *application.App) error {
	if app == nil || app.Updater == nil {
		return fmt.Errorf("appmeta: updater unavailable (call after application.New)")
	}

	gh, err := github.New(github.Config{
		Repository:    UpdateRepo,
		ChecksumAsset: ChecksumAsset,
	})
	if err != nil {
		return fmt.Errorf("appmeta: github provider: %w", err)
	}

	cfg := updater.Config{
		CurrentVersion: CurrentVersion(),
		Providers:      []updater.Provider{gh},
	}

	if err := app.Updater.Init(cfg); err != nil {
		return fmt.Errorf("appmeta: updater init: %w", err)
	}

	for _, name := range []string{
		updater.EventUpdateAvailable,
		updater.EventDownloadProgress,
		updater.EventUpdateReady,
		updater.EventError,
	} {
		evt := name
		app.Event.On(evt, func(e *application.CustomEvent) {
			log.Info().Str("event", evt).Any("data", e.Data).Msg("updater")
		})
	}

	log.Info().
		Str("version", CurrentVersion()).
		Str("repo", UpdateRepo).
		Str("checksum", ChecksumAsset).
		Msg("self-updater initialised (GitHub Releases)")
	return nil
}

func CheckForUpdates(app *application.App) {
	go func() {
		if app == nil || app.Updater == nil {
			log.Warn().Msg("appmeta: check for updates skipped (updater not initialised)")
			return
		}
		if err := app.Updater.CheckAndInstall(context.Background()); err != nil {
			log.Error().Err(err).Msg("appmeta: check for updates failed")
		}
	}()
}

func InstallUpdateMenu(app *application.App) {
	if app == nil || runtime.GOOS != "darwin" {
		return
	}
	menu := app.Menu.New()
	app.Menu.SetApplicationMenu(menu)

	appMenu := menu.AddSubmenu("NovelClaw")
	appMenu.Add("Check for Updates…").OnClick(func(ctx *application.Context) {
		CheckForUpdates(app)
	})
	appMenu.AddSeparator()
	appMenu.Add(fmt.Sprintf("About NovelClaw %s", CurrentVersion())).OnClick(func(ctx *application.Context) {
		app.Menu.ShowAbout()
	})
}
