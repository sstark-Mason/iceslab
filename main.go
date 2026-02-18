package main

import (
	"embed"
	_ "embed"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"time"

	"iceslab/utils"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

//go:embed all:assets
var embedded embed.FS

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
	var stationID string

	if _, err := os.Stat(".git"); err == nil {
		log.Fatal().Msg(".git directory found; exiting. Don't run this in your git repo, dumbass.")
		os.Exit(1)
	}

	// flagManifest := flag.Bool("manifest", false, "Generate and save manifest.yaml with current binary and assets hashes")
	// flagUpdate := flag.Bool("update", false, "Fetch latest repo state from GitHub before setup")
	// flagBuild := flag.Bool("build", false, "Build the iceslab binary from source before setup")

	update := flag.String("u", "", "Update 's'ource or 'b'ookmarks")
	install := flag.Bool("i", false, "Install to /opt/iceslab/ and run setup scripts")
	verbose := flag.Bool("v", false, "Enable verbose logging")
	dump := flag.Bool("dump", false, "Dump embedded assets")

	flag.Parse()

	if *verbose {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	if *dump {
		log.Info().Msg("Dumping embedded assets")
		err = utils.DumpAssets(embedded, "assets", "assets")
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to dump embedded assets")
		}
		log.Info().Msg("Embedded assets dumped successfully")
	}

	err = utils.CheckIfCorrectUser()
	if err != nil {
		log.Fatal().Err(err).Msg("Exiting on user check")
		return
	}

	stationID, err = utils.GetStationID()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to get station ID")
	}

	switch *update {
	case "s", "source":
		log.Info().Msg("Updating source code")
		client := utils.NewClient("")
		err := client.UpdateSource()
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to update source code")
		}
		return
	case "b", "bookmarks":
		log.Info().Msg("Updating bookmarks")
		client := utils.NewClient("")
		err := client.UpdateBookmarks()
		if err != nil {
			log.Err(err).Msg("Failed to update bookmarks")
		}
		err = utils.InstallBookmarks(stationID)
		if err != nil {
			log.Err(err).Msg("Failed to install bookmarks")
		}
		return
	}

	if *install {
		log.Info().Msg("Installing to /opt/iceslab/")
		execPath, err := os.Executable()
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to get executable path")
		}
		err = utils.MoveFile(execPath, "/opt/iceslab/iceslab")
		err = utils.DumpAssets(embedded, "assets", "/opt/iceslab/assets")
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to move binary and dump assets to /opt/iceslab/")
		}
		err = utils.InstallBookmarks(stationID)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to install bookmarks")
		}
		log.Info().Msg("Installation completed successfully")
		return
	}

	// log.Info().Msgf("Running iceslab with flags: manifest=%t, update=%t, build=%t", *flagManifest, *flagUpdate, *flagBuild)

	// if *flagBuild {
	// 	log.Info().Msg("Build flag provided; building iceslab binary")
	// 	err = utils.Build()
	// 	if err != nil {
	// 		log.Fatal().Err(err).Msg("Failed to build iceslab binary")
	// 	}

	// 	manifest, err := utils.GenerateManifest()
	// 	if err != nil {
	// 		log.Fatal().Err(err).Msg("Failed to generate manifest")
	// 	}
	// 	log.Info().Str("manifest", manifest.BinaryHash).Msg("Generated manifest")

	// 	err = utils.SaveManifest(manifest, "manifest.yaml")
	// 	if err != nil {
	// 		log.Fatal().Err(err).Msg("Failed to save manifest")
	// 	}

	// 	return
	// }

	// if *flagManifest {
	// 	log.Info().Msg("Manifest flag provided; generating manifest")
	// 	manifest, err := utils.GenerateManifest()
	// 	if err != nil {
	// 		log.Fatal().Err(err).Msg("Failed to generate manifest")
	// 	}
	// 	log.Info().Str("manifest", manifest.BinaryHash).Msg("Generated manifest")

	// 	err = utils.SaveManifest(manifest, "manifest.yaml")
	// 	if err != nil {
	// 		log.Fatal().Err(err).Msg("Failed to save manifest")
	// 	}
	// }

	// if *flagUpdate {
	// 	log.Info().Msg("Update flag provided; checking for updates")
	// 	client := utils.NewClient("") // Limited to 60 reqs/hour without auth
	// 	err := client.Update()
	// 	if err != nil {
	// 		log.Fatal().Err(err).Msg("Failed to update local repo state")
	// 	}
	// 	log.Info().Msg("Update check completed successfully")
	// 	return
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
