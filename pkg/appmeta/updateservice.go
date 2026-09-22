package appmeta

import (
	"context"
	"errors"

	"github.com/rs/zerolog/log"
	"github.com/wailsapp/wails/v3/pkg/application"
)

type UpdateService struct{}

func NewUpdateService() *UpdateService { return &UpdateService{} }

type AppInfo struct {
	Version  string `json:"version"`
	Repo     string `json:"repo"`
	Platform string `json:"platform"`
	DataDir  string `json:"data_dir"`
}

func (s *UpdateService) GetAppInfo() AppInfo {
	r := VersionReport()
	return AppInfo{
		Version:  r.Version,
		Repo:     r.UpdateRepo,
		Platform: r.Platform,
		DataDir:  r.DataDir,
	}
}

func (s *UpdateService) CheckForUpdates() error {
	app := application.Get()
	if app == nil || app.Updater == nil {
		return errors.New("updater is not initialised")
	}
	go func() {
		if err := app.Updater.CheckAndInstall(context.Background()); err != nil {
			log.Error().Err(err).Msg("manual update check failed (see updater window)")
		}
	}()
	return nil
}

func (s *UpdateService) CheckSilent() error {
	app := application.Get()
	if app == nil || app.Updater == nil {
		return nil
	}
	go func() {
		rel, err := app.Updater.Check(context.Background())
		if err != nil {
			log.Debug().Err(err).Msg("silent update check skipped (network/repo error)")
			return
		}
		if rel == nil {
			return
		}
		if err := app.Updater.CheckAndInstall(context.Background()); err != nil {
			log.Error().Err(err).Msg("update found but install flow failed")
		}
	}()
	return nil
}
