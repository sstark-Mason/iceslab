package utils

import (
	"fmt"
	"net/http"

	"github.com/rs/zerolog/log"
	"go.yaml.in/yaml/v4"
)

const (
	owner  = "sstark-mason"
	repo   = "iceslab"
	branch = "main"
)

type GitHubCommit struct {
	SHA string `json:"sha"`
}

func Update() error {
	manifest, err := GenerateManifest()
	if err != nil {
		return err
	}

	log.Info().Str("binary_hash", manifest.BinaryHash).Str("assets_hash", manifest.AssetsHash).Msg("Current binary and assets hashes")

	remoteManifest, err := FetchRemoteManifest()
	if err != nil {
		return err
	}

	log.Info().Str("remote_binary_hash", remoteManifest.BinaryHash).Str("remote_assets_hash", remoteManifest.AssetsHash).Msg("Remote binary and assets hashes")

	if manifest.BinaryHash != remoteManifest.BinaryHash || manifest.AssetsHash != remoteManifest.AssetsHash {
		log.Warn().Msg("Local binary or assets are outdated compared to GitHub")
	} else {
		log.Info().Msg("Local binary and assets are up to date with GitHub")
	}

	return nil
}

func FetchRemoteManifest() (Manifest, error) {

	client := &http.Client{}
	req, _ := http.NewRequest(
		"GET",
		fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s/manifest.yaml", owner, repo, branch),
		nil,
	)
	req.Header.Set("User-Agent", "iceslab-config-setup-script")

	resp, err := client.Do(req)
	if err != nil {
		return Manifest{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Manifest{}, fmt.Errorf("failed to fetch remote manifest: status %d", resp.StatusCode)
	}

	var manifest Manifest
	err = yaml.NewDecoder(resp.Body).Decode(&manifest)
	if err != nil {
		return Manifest{}, err
	}

	return manifest, nil
}
