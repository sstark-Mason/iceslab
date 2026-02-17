package utils

import (
	"fmt"
	"os"

	"github.com/rs/zerolog/log"
)

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

func PromptForStationNumber() (string, error) {
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

func SaveStationNumber(stationNum string) error {
	err := os.WriteFile(".station_number", []byte(stationNum), 0644)
	if err != nil {
		return err
	}
	return nil
}

func LoadStationNumber() (string, error) {
	data, err := os.ReadFile(".station_number")
	if err != nil {
		return "", err
	}
	return string(data), nil
}
