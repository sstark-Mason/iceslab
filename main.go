package main

import (
	"embed"
	_ "embed"
	"flag"
	"os"
	"time"

	"iceslab/utils"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

//go:embed all:assets
var embedded embed.FS

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
		err := client.UpdateBookmarkYamls()
		if err != nil {
			log.Err(err).Msg("Failed to update bookmarks")
		}
		err = utils.InsertBookmarksInPolicies(stationID)
		if err != nil {
			log.Err(err).Msg("Failed to install bookmarks")
		}
		err = utils.CopyDirectoryTo("assets/etc/", "/etc/")
		if err != nil {
			log.Error().Err(err).Msg("Failed to copy assets/etc/ to /etc/")
			return
		}
		return
	}

	if *install {
		installPath := "/opt/iceslab/"
		log.Info().Str("installPath", installPath).Msg("Installing iceslab to path")

		execPath, err := os.Executable()
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to get executable path")
		}

		if execPath != installPath+"iceslab" {
			log.Warn().Str("execPath", execPath).Msg("Executable is not running from install path; moving to /opt/iceslab/")
			err = utils.InstallTo(installPath)
			if err != nil {
				log.Fatal().Err(err).Msg("Failed to install to path")
			}
			// Re-run the program from the new location with same flags
			err = utils.RerunBinary(installPath+"iceslab", os.Args[1:]...)
			os.Exit(0)
		}

		// if _, err := os.Stat(installPath); err == nil {
		// 	log.Error().Msg("Existing installation found at installPath; exiting")
		// 	os.Exit(1)
		// }

		// err = utils.MoveFile(execPath, installPath+"iceslab")
		// if err != nil {
		// 	log.Fatal().Err(err).Msg("Failed to move binary to installPath")
		// }

		err = utils.DumpAssets(embedded, "assets", installPath+"assets")
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to dump assets in installPath")
		}

		err = utils.InstallPackages(installPath)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to install packages")
		}

		err = utils.SetupGuestUser(installPath)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to set up guest user")
		}

		err = utils.InsertBookmarksInPolicies(stationID)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to install bookmarks")
		}

		err = utils.CopyDirectoryTo(installPath+"assets/etc/", "/etc/")
		if err != nil {
			log.Error().Err(err).Msg("Failed to copy assets/etc/ to /etc/")
			return
		}

		// log.Info().Msg("Installation completed successfully. Rebooting system to apply changes.")
		// err = utils.RunShellCommand("sudo reboot")
		// if err != nil {
		// 	log.Error().Err(err).Msg("Failed to reboot system")
		// }

		return
	}

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
