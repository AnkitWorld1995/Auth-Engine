package utility

import (
	errs "github.com/chsys/userauthenticationengine/pkg/lib/error"
	"golang.org/x/crypto/bcrypt"
	"sync"
	"unicode"
)


type primitives struct {
	cond 	*sync.Cond
}

func GenHashAndSaltPassword(password string) (string, *errs.AppError) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		return "", errs.NewUnexpectedError(err.Error())
	}
	return string(hash), nil
}

func ComparePassword(hashedPassword []byte, newPwd string) bool {
	transMutePwd := []byte(newPwd)
	err := bcrypt.CompareHashAndPassword(hashedPassword, transMutePwd)
	if err != nil {
		return false
	}
	return true
}

func PasswordValidator(password string, isSignIn bool) *errs.AppError {
	var (
		hasMinLen  = false
		hasUpper   = false
		hasLower   = false
		hasNumber  = false
		hasSpecial = false
	)
	if len(password) >= 8 {
		hasMinLen = true
	}
	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	/*
		1. If Sign-In is True:
			1.1. Compare The Password with Hashed.
	*/

	if isSignIn {
		// Compare Password
	}

	if hasMinLen && hasUpper && hasLower && hasNumber && hasSpecial {
		return nil
	} else {
		return errs.NewValidationError("password do not match the criteria of at least one upper case, one lower case, one number, one special character and minimum of 8 characters.")
	}

}