package main

import (
	"embed"
	_ "embed"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"iceslab/utils"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

//go:embed all:assets
var embedded embed.FS

func dumpAssets(src, dest string) error {
	entries, err := embedded.ReadDir(src)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		fullSrc := filepath.Join(src, entry.Name())
		fullDest := filepath.Join(dest, entry.Name())
		if entry.IsDir() {
			err = os.MkdirAll(fullDest, 0755)
			if err != nil {
				return err
			}
			err = dumpAssets(fullSrc, fullDest)
			if err != nil {
				return err
			}
		} else {
			alreadyExists := false
			if _, err := os.Stat(fullDest); err == nil {
				alreadyExists = true
				log.Debug().Str("file", entry.Name()).Str("path", fullDest).Msg("Asset already exists, skipping")
				continue
			}
			if !alreadyExists {
				data, err := embedded.ReadFile(fullSrc)
				if err != nil {
					log.Warn().Err(err).Str("file", entry.Name()).Str("path", fullDest).Msg("Failed to read asset")
					continue
				}
				err = writeFile(fullDest, data, 0644)
				if err != nil {
					log.Warn().Err(err).Str("file", entry.Name()).Str("path", fullDest).Msg("Failed to write asset")
					continue
				} else {
					log.Info().Str("file", entry.Name()).Str("path", fullDest).Msg("Asset written")
					continue
				}
			}
		}
	}

	return nil
}

func writeFile(path string, data []byte, perm os.FileMode) error {
	log.Debug().Str("path", path).Msg("Writing file")
	dir := filepath.Dir(path)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, perm)
}

func promptForStationNumber() (string, error) {
	var stationNum string
	log.Info().Msg("Please enter the station number (e.g., 01, 02, etc.):")
	_, err := fmt.Scanln(&stationNum)
	if err != nil {
		return "", err
	}
	if len(stationNum) == 1 && stationNum >= "1" && stationNum <= "9" {
		stationNum = "0" + stationNum
	}
	return stationNum, nil
}

func getAssetPath() string {
	// Check for local "assets" directory first
	if _, err := os.Stat("assets"); err == nil {
		log.Debug().Msg("Using local assets path")
		return "assets"
	}
	// Fallback to embedded assets
	log.Debug().Msg("Using embedded assets path")
	return "embedded"
}

func runScript(path string) error {
	if _, err := os.Stat(path); err == nil {
		log.Info().Str("script", path).Msg("Running script")
		err = os.Chmod(path, 0755)
		if err != nil {
			return fmt.Errorf("failed to make script executable: %w", err)
		}
		cmd := fmt.Sprintf("./%s", path)
		if err := exec.Command(cmd).Run(); err != nil {
			return fmt.Errorf("failed to run script: %w", err)
		}
		log.Info().Str("script", path).Msg("Script completed successfully")
	}
	return nil
}

func main() {

	zerolog.TimeFieldFormat = time.RFC3339
	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
	})
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	var err error

	hash1, err := utils.HashDirectory("assets")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to hash assets directory")
	}
	log.Info().Str("assets_hash", hash1).Msg("Assets directory hash")

	hash2, err := utils.HashFile("iceslab")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to hash iceslab binary")
	}
	log.Info().Str("embedded_hash", hash2).Msg("Iceslab binary hash")

	if hash1 != hash2 {
		log.Warn().Msg("Assets directory and embedded assets hashes do not match")
	} else {
		log.Info().Msg("Assets directory and embedded assets hashes match")
	}

	flagManifest := flag.Bool("manifest", false, "Generate and save manifest.yaml with current binary and assets hashes")
	flagUpdate := flag.Bool("update", false, "Fetch latest repo state from GitHub before setup")
	flag.Parse()

	if *flagManifest {
		log.Info().Msg("Manifest flag provided; generating manifest")
		manifest, err := utils.GenerateManifest()
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to generate manifest")
		}
		log.Info().Str("manifest", manifest.BinaryHash).Msg("Generated manifest")

		err = utils.SaveManifest(manifest, "manifest.yaml")
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to save manifest")
		}
	}

	if *flagUpdate {
		log.Info().Msg("Update flag provided; checking for updates")
		err = utils.Update()
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to update repo state")
		}
		return
	}

	// update := flag.Bool("update", false, "Fetch latest repo state from GitHub before setup")
	// flag.Parse()

	// if *update {
	// 	err = utils.FetchUpdates()
	// 	if err != nil {
	// 		log.Fatal().Err(err).Msg("Failed to update repo state")
	// 	}
	// 	return
	// }

	// stationNum, err := promptForStationNumber()
	// if err != nil {
	// 	log.Fatal().Err(err).Msg("Failed to read station number")
	// }

	// log.Info().Str("station_number", stationNum).Msg("Starting setup for station")

	// assetPath := getAssetPath()
	// log.Info().Str("asset_path", assetPath).Msg("Using asset path for setup")

	// if assetPath == "embedded" {
	// 	log.Info().Msg("Using embedded assets for setup")
	// 	err = dumpAssets("assets", "assets")
	// 	assetPath = "assets"
	// 	if err != nil {
	// 		log.Fatal().Err(err).Msg("Failed to dump assets")
	// 		return
	// 	}
	// }

	// bookmarks, err := utils.CollectBookmarks(filepath.Join(assetPath, "bookmarks"), stationNum)
	// if err != nil {
	// 	log.Fatal().Err(err).Msg("Failed to collect bookmarks")
	// 	return
	// }
	// log.Info().Int("count", len(bookmarks)).Msg("Collected bookmarks")

	// err = utils.InsertBookmarks("firefox", filepath.Join(assetPath, "etc/firefox/policies/policies.json"), bookmarks)
	// if err != nil {
	// 	log.Fatal().Err(err).Msg("Failed to insert bookmarks into Firefox policies.json")
	// 	return
	// }
	// log.Info().Msg("Bookmarks inserted into Firefox policies.json successfully")

	// err = utils.InsertBookmarks("chromium", filepath.Join(assetPath, "etc/chromium/policies/managed/policies.json"), bookmarks)
	// if err != nil {
	// 	log.Fatal().Err(err).Msg("Failed to insert bookmarks into Chromium policies.json")
	// 	return
	// }
	// log.Info().Msg("Bookmarks inserted into Chromium policies.json successfully")

	// // Run scripts/post-install.sh
	// postInstallPath := filepath.Join(assetPath, "scripts/post-install.sh")
	// if err := runScript(postInstallPath); err != nil {
	// 	log.Fatal().Err(err).Msg("Failed to run post-install script")
	// 	return
	// }

	// // Run scripts/guest-template.sh
	// guestTemplatePath := filepath.Join(assetPath, "scripts/guest-template.sh")
	// if err := runScript(guestTemplatePath); err != nil {
	// 	log.Fatal().Err(err).Msg("Failed to run guest-template script")
	// 	return
	// }

	// log.Info().Msg("Setup completed successfully")
}
