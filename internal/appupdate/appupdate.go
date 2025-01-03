package appupdate

import (
	"bytes"
	"context"
	"io"
	"os"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/atinylittleshell/gsh/internal/core"
	"github.com/atinylittleshell/gsh/internal/styles"
	"github.com/atinylittleshell/gsh/pkg/gline"
	"github.com/creativeprojects/go-selfupdate"
	"go.uber.org/zap"
)

func HandleSelfUpdate(currentVersion string, logger *zap.Logger) {
	currentSemVer, err := semver.NewVersion(currentVersion)
	if err != nil {
		logger.Debug("running a dev build, skipping self-update check")
		return
	}

	// Check if we have previously detected a newer version
	updateToLatestVersion(currentSemVer, logger)

	// Check for newer versions from remote repository
	go fetchAndSaveLatestVersion(logger)
}

func readLatestVersion() string {
	file, err := os.Open(core.LatestVersionFile())
	if err != nil {
		return ""
	}
	defer file.Close()

	var buf bytes.Buffer
	_, err = io.Copy(&buf, file)
	if err != nil {
		return ""
	}

	return strings.TrimSpace(buf.String())
}

func updateToLatestVersion(currentSemVer *semver.Version, logger *zap.Logger) {
	latestVersion := readLatestVersion()
	if latestVersion == "" {
		return
	}

	latestSemVer, err := semver.NewVersion(latestVersion)
	if err != nil {
		logger.Error("failed to parse latest version", zap.Error(err))
		return
	}
	if latestSemVer.LessThanEqual(currentSemVer) {
		return
	}

	confirm, err := gline.Gline(
		styles.AGENT_QUESTION("New version of gsh available. Update now? (Y/n)"),
		latestVersion,
		nil,
		nil,
		logger,
		gline.NewOptions(),
	)

	if strings.ToLower(confirm) == "n" {
		return
	}

	latest, found, err := selfupdate.DetectLatest(
		context.Background(),
		selfupdate.ParseSlug("atinylittleshell/gsh"),
	)
	if err != nil {
		logger.Warn("error occurred while detecting latest version", zap.Error(err))
		return
	}
	if !found {
		logger.Warn("latest version could not be detected")
		return
	}

	exe, err := selfupdate.ExecutablePath()
	if err != nil {
		logger.Error("failed to get executable path to update", zap.Error(err))
		return
	}
	if err := selfupdate.UpdateTo(context.Background(), latest.AssetURL, latest.AssetName, exe); err != nil {
		logger.Error("failed to update to latest version", zap.Error(err))
		return
	}

	logger.Info("successfully updated to latest version", zap.String("version", latest.Version()))
}

func fetchAndSaveLatestVersion(logger *zap.Logger) {
	latest, found, err := selfupdate.DetectLatest(
		context.Background(),
		selfupdate.ParseSlug("atinylittleshell/gsh"),
	)
	if err != nil {
		logger.Warn("error occurred while getting latest version from remote", zap.Error(err))
		return
	}
	if !found {
		logger.Warn("latest version could not be found")
		return
	}

	recordFilePath := core.LatestVersionFile()
	file, err := os.Create(recordFilePath)
	defer file.Close()

	if err != nil {
		logger.Error("failed to save latest version", zap.Error(err))
		return
	}

	_, err = file.WriteString(latest.Version())
	if err != nil {
		logger.Error("failed to save latest version", zap.Error(err))
		return
	}
}
