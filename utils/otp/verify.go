package otp

import (
	

	"golang.org/x/crypto/bcrypt"
)

func HashOTP(plain string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(plain), 12)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

