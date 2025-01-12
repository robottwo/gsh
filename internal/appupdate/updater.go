package appupdate

import (
	"context"

	"github.com/creativeprojects/go-selfupdate"
)

type Release interface {
	Version() string
	AssetURL() string
	AssetName() string
}

type DefaultRelease struct {
	value *selfupdate.Release
}

func (r DefaultRelease) Version() string {
	return r.value.Version()
}

func (r DefaultRelease) AssetURL() string {
	return r.value.AssetURL
}

func (r DefaultRelease) AssetName() string {
	return r.value.AssetName
}

type Updater interface {
	DetectLatest(ctx context.Context, slug string) (Release, bool, error)
	UpdateTo(ctx context.Context, assetURL, assetName, exePath string) error
}

type DefaultUpdater struct{}

func (u DefaultUpdater) DetectLatest(ctx context.Context, slug string) (Release, bool, error) {
	defaultRelease, ok, err := selfupdate.DetectLatest(ctx, selfupdate.ParseSlug(slug))
	release := DefaultRelease{value: defaultRelease}
	return release, ok, err
}

func (u DefaultUpdater) UpdateTo(ctx context.Context, assetURL, assetName, exePath string) error {
	return selfupdate.UpdateTo(ctx, assetURL, assetName, exePath)
}
