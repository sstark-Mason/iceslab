package utils

import (
	"os"
	"os/exec"

	"github.com/rs/zerolog/log"
)

// This feels illegal for some reason

func RunShellCommand(command string) error {
	cmd := exec.Command("sh", "-c", command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func Build() error {
	log.Info().Msg("Building iceslab binary")
	err := RunShellCommand(`go build -ldflags="-s -w" .`)
	if err != nil {
		return err
	}
	return nil
}

func BuildDebug() error {
	log.Info().Msg("Building iceslab binary")
	err := RunShellCommand(`go build .`)
	if err != nil {
		return err
	}
	return nil
}
