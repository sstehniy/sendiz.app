package server

import (
	"errors"
	"regexp"
)

func validatePhoneNumber(phoneNumber string) (bool, error) {
	re := regexp.MustCompile(`^\+?[1-9]\d{1,14}$`)
	if !re.MatchString(phoneNumber) {
		return false, errors.New("invalid phone number format")
	}

	return true, nil
}
