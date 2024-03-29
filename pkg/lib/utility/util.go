package utility

import (
	"crypto/rand"
	"fmt"
	"github.com/chsys/userauthenticationengine/pkg/lib/constants"
	errs "github.com/chsys/userauthenticationengine/pkg/lib/error"
	"golang.org/x/crypto/bcrypt"
	"io"
	"log"
	"net/mail"
	"strings"
	"sync"
	"unicode"
)



var accountType map[string]string

func init() {
	userAccountType := make(map[string]string, 3)
	userAccountType[constants.Root] 	  = constants.RootAdminAccountType
	userAccountType[constants.User]  	  = constants.UserAccountType
	userAccountType[constants.SalesRoot]  = constants.SalesAdminAccountType

	accountType = userAccountType
}


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

func ParseMail(email string) string {
	emailAddress, err := mail.ParseAddress(email)
	if err != nil {
		log.Fatalln(fmt.Sprintf("Utility, ParseMail Error: %s ", err.Error()))
	}
	return emailAddress.Address
}

func ComparePassword(hashedPwd string, plainPwd []byte) bool {
	byteHash := []byte(hashedPwd)
	err := bcrypt.CompareHashAndPassword(byteHash, plainPwd)
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

	if hasMinLen && hasUpper && hasLower && hasNumber && hasSpecial {
		return nil
	} else {
		return errs.NewValidationError("password do not match the criteria of at least one upper case, one lower case, one number, one special character and minimum of 8 characters.")
	}

}


func MapUserAccountType(userType string) (string, *errs.AppError) {
	if accountType, ok := accountType[strings.ToLower(userType)]; !ok {
		return "",errs.NewValidationError("Incorrect Account Type. User account Type Must be Either Sales Root A/C | Root Admin A/C | User A/C")
	}else{
		return accountType, nil
	}
}



var table = [...]byte{'1', '2', '3', '4', '5', '6', '7', '8', '9', '0'}

func GenerateOTP(max int) (string, error) {
	b := make([]byte, max)
	n, err := io.ReadAtLeast(rand.Reader, b, max)
	if n != max {
		return "", err
	}
	for i := 0; i < len(b); i++ {
		b[i] = table[int(b[i])%len(table)]
	}
	return string(b), nil
}