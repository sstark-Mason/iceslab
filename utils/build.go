package utils

import (
	"github.com/rs/zerolog/log"
)

// This feels illegal for some reason

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
