package utils

import (
	"fmt"
	"net/http"
	"os"

	"go.yaml.in/yaml/v4"
)

type Manifest struct {
	BinaryHash string `yaml:"binary_hash" json:"binary_hash"`
	AssetsHash string `yaml:"assets_hash" json:"assets_hash"`
}

func GenerateManifest() (Manifest, error) {
	binaryHash, err := HashFile("iceslab")
	if err != nil {
		binaryHash = "error"
	}

	assetsHash, err := HashDirectory("assets")
	if err != nil {
		assetsHash = "error"
	}

	return Manifest{
		BinaryHash: binaryHash,
		AssetsHash: assetsHash,
	}, nil
}

func SaveManifest(manifest Manifest, path string) error {
	data, err := yaml.Marshal(manifest)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func LoadManifest(path string) (Manifest, error) {
	var manifest Manifest
	data, err := os.ReadFile(path)
	if err != nil {
		return manifest, err
	}
	err = yaml.Unmarshal(data, &manifest)
	return manifest, err
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
