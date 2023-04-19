package dto

import (
	strUtil "github.com/agrison/go-commons-lang/stringUtils"
	errs "github.com/chsys/userauthenticationengine/pkg/lib/error"
	"github.com/chsys/userauthenticationengine/pkg/lib/utility"
	"net/mail"
	"strings"
)

type SignUpRequest struct {
	UserName 	string 	`json:"user_name"`
	FirstName	string 	`json:"first_name"`
	LastName	string 	`json:"last_name"`
	Email		string 	`json:"email"`
	Password  	string 	`json:"password"`
	PhoneNumber int32 	`json:"phone_number"`
	Address		*string `json:"address"`
	IsAdmin		*bool	`json:"is_admin"`
	UserType	string	`json:"user_type"`
}

type UserResponse struct {
	UserID		string 	`json:"id"`
	UserName 	string	`json:"user_name"`
	FirstName	string	`json:"first_name"`
	LastName	string	`json:"last_name"`
	Password	string  `json:"password"`
	Email		string	`json:"email"`
	Phone		int32	`json:"phone"`
	Address 	*string	`json:"address"`
	IsAdmin		bool	`json:"is_admin"`
	CreatedAt	string	`json:"created_at"`
	UpdatedAt 	string	`json:"updated_at"`
}


type SignUpResponse struct {
	Success		bool
	Message		string
}


func (r *SignUpRequest) SignUpValidate() *errs.AppError{
	_, err := mail.ParseAddress(r.Email)
	if err != nil {
		return errs.NewValidationError("invalid email address")
	}
	if strUtil.IsBlank(strings.TrimSpace(r.FirstName)) {
		return errs.NewValidationError("Firstname cannot be empty")
	}
	if strUtil.IsBlank(strings.TrimSpace(r.LastName)) {
		return errs.NewValidationError("Lastname cannot be empty")
	}

	if strUtil.IsBlank(strings.TrimSpace(r.UserType)) {
		return errs.NewValidationError("User Type cannot be empty")
	}
	passwordValidatorError := utility.PasswordValidator(r.Password, false)
	if passwordValidatorError != nil {
		return passwordValidatorError
	}

	return nil
}