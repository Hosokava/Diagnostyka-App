package utils

import (
	"strconv"
)

func IsValidPESEL(pesel string) bool {
	if len(pesel) != 11 {
		return false
	}

	weights := []int{1, 3, 7, 9, 1, 3, 7, 9, 1, 3}
	sum := 0

	for i := 0; i < 10; i++ {
		digit, err := strconv.Atoi(string(pesel[i]))
		if err != nil {
			return false
		}
		sum += digit * weights[i]
	}

	controlDigit, err := strconv.Atoi(string(pesel[10]))
	if err != nil {
		return false
	}

	checksum := (10 - (sum % 10)) % 10

	return checksum == controlDigit
}
