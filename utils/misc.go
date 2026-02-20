package utils

import (
	"embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog/log"
)

func RunShellCommand(command string) error {
	cmd := exec.Command("sh", "-c", command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func CheckIfCorrectUser() error {
	expectedUser := "admin"
	currentUser := os.Getenv("USER")
	if currentUser != expectedUser {
		log.Warn().Msgf("Current user '%s' does not match expected user '%s'. Continue? (y/n)", currentUser, expectedUser)
		var response string
		_, err := fmt.Scanln(&response)
		if err != nil {
			return fmt.Errorf("failed to read user response: %w", err)
		}

		if response != "y" && response != "Y" {
			return fmt.Errorf("user did not confirm continuation; exiting")
		} else {
			log.Info().Msg("User confirmed continuation despite mismatch")
		}
	}
	return nil
}

func GetStationID() (string, error) {
	bytes, err := os.ReadFile(".station_id")
	switch err.(type) {
	case nil:
		ID := string(bytes)
		log.Debug().Str("station_id", ID).Msg("Loaded station ID from file")
		return ID, nil
	case *os.PathError:
		log.Info().Msg("No station ID file found; prompting user for station ID")
		ID, err := PromptForStationID()
		if err != nil {
			return "", fmt.Errorf("failed to prompt for station ID: %w", err)
		}
		err = os.WriteFile(".station_id", []byte(ID), 0644)
		if err != nil {
			return "", fmt.Errorf("failed to save station ID: %w", err)
		}
		log.Info().Str("station_id", ID).Msg("Station ID saved successfully")
		return ID, nil
	default:
		return "", fmt.Errorf("failed to read station ID file: %w", err)
	}
}

func PromptForStationID() (string, error) {
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

func MoveFile(src, dest string) error {

	src = filepath.Clean(src)
	dest = filepath.Clean(dest)

	if strings.HasPrefix(dest, src+string(filepath.Separator)) {
		return fmt.Errorf("destination %s is a subpath of source %s", dest, src)
	}

	if _, err := os.Stat(src); os.IsNotExist(err) {
		return fmt.Errorf("source file %s does not exist: %w", src, err)
	}

	destDir := filepath.Dir(dest)
	err := os.MkdirAll(destDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create destination directory %s: %w", destDir, err)
	}

	err = os.Rename(src, dest)
	if err != nil {
		return fmt.Errorf("failed to move file from %s to %s: %w", src, dest, err)
	}

	log.Debug().Str("src", src).Str("dest", dest).Msg("File moved successfully")
	return nil
}

func DumpAssets(efs embed.FS, src, dest string) error {
	entries, err := efs.ReadDir(src)
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
			err = DumpAssets(efs, fullSrc, fullDest)
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
				data, err := efs.ReadFile(fullSrc)
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

func RunScript(path string) error {
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

func RerunBinary(path string, args ...string) error {
	if _, err := os.Stat(path); err == nil {
		log.Info().Str("binary", path).Msg("Running binary")
		cmd := exec.Command(path, args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Start()
		if err != nil {
			return fmt.Errorf("failed to run binary: %w", err)
		}
		log.Info().Str("binary", path).Msg("Binary completed successfully")
	}
	return nil
}

func CopyDirectoryTo(src, dest string) error {
	src = filepath.Clean(src)
	dest = filepath.Clean(dest)

	if strings.HasPrefix(dest, src+string(filepath.Separator)) {
		return fmt.Errorf("destination %s is a subpath of source %s", dest, src)
	}

	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		destPath := filepath.Join(dest, relPath)
		if info.IsDir() {
			return os.MkdirAll(destPath, 0755)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return writeFile(destPath, data, 0644)
	})
}
