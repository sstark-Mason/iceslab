package utils

import (
	"fmt"
	"os"

	"github.com/rs/zerolog/log"
)

func InstallTo(path string) error {
	log.Info().Str("path", path).Msg("Installing to path")
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}
	err = MoveFile(execPath, path)
	if err != nil {
		return fmt.Errorf("failed to move binary to %s: %w", path, err)
	}
	log.Debug().Msgf("Binary moved to %s successfully", path)
	return nil
}

func SetupGuestUser(installPath string) error {
	log.Info().Str("installPath", installPath).Msg("Setting up guest user")
	err := RunShellCommand("sudo chmod +x " + installPath + "assets/scripts/create_guest_template.sh")
	if err != nil {
		return fmt.Errorf("failed to make guest template script executable: %w", err)
	}

	err = RunShellCommand("sudo " + installPath + "assets/scripts/create_guest_template.sh")
	if err != nil {
		return fmt.Errorf("failed to run guest template script: %w", err)
	}

	log.Info().Msg("Guest user setup completed successfully")

	return nil
}

func InstallPackages(installPath string) error {
	log.Info().Str("installPath", installPath).Msg("Installing software dependencies")
	err := RunShellCommand("sudo chmod +x " + installPath + "assets/scripts/install_upgrade_packages.sh")
	if err != nil {
		return fmt.Errorf("failed to make software install script executable: %w", err)
	}

	err = RunShellCommand("sudo " + installPath + "assets/scripts/install_upgrade_packages.sh")
	if err != nil {
		return fmt.Errorf("failed to run software install script: %w", err)
	}
	log.Info().Msg("Software dependencies installed successfully")

	return nil
}
