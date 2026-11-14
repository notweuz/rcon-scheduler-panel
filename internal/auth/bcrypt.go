package auth

import (
	"errors"
	"rcon-scheduler-panel/internal/errs"

	"golang.org/x/crypto/bcrypt"
)

var ErrPasswordTooLong = errs.ErrBadRequest.WithCode("PASSWORD_TOO_LONG").WithMessage("password is too long")

func GeneratePasswordHash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		if errors.Is(err, bcrypt.ErrPasswordTooLong) {
			return "", ErrPasswordTooLong.WithErr(err)
		}
		return "", err
	}
	return string(hash), nil
}

func VerifyPasswordHash(hash, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
