package utils

import (
	"os"
	"time"

	"go.yaml.in/yaml/v4"
)

type Manifest struct {
	BinaryHash  string `yaml:"binary_hash" json:"binary_hash"`
	AssetsHash  string `yaml:"assets_hash" json:"assets_hash"`
	LastUpdated string `yaml:"last_updated" json:"last_updated"`
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
		BinaryHash:  binaryHash,
		AssetsHash:  assetsHash,
		LastUpdated: time.Now().Format(time.RFC3339),
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
