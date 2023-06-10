package dto

import (
	strUtil "github.com/agrison/go-commons-lang/stringUtils"
	"github.com/chsys/userauthenticationengine/pkg/domain"
	errs "github.com/chsys/userauthenticationengine/pkg/lib/error"
	"github.com/chsys/userauthenticationengine/pkg/lib/utility"
	"net/http"
	"net/mail"
	"strings"
)

type MapClaims map[string]interface{}

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

type SSOSignInRequest struct {
	AuthToken	string	`json:"auth_token"`
	Username 	string  `json:"username"`
	Password 	string  `json:"password"`
}

type SignInRequest struct {
	UserID 		int		`json:"user_id"`
	UserName	string	`json:"user_name"`
	Email 		string 	`json:"email"`
	Password	string  `json:"password"`
}

type SignUpResponse struct {
	Success		bool
	Message		string
}

type SignInResponse struct {
	ID 			string	`json:"id"`
	UserID		string 	`json:"user_id"`
	UserName 	string	`json:"user_name"`
	FirstName	string	`json:"first_name"`
	LastName	string	`json:"last_name"`
	Email		string	`json:"email"`
	Phone		int32	`json:"phone"`
	Address 	*string	`json:"address"`
	IsAdmin		bool	`json:"is_admin"`
	UserType	string	`json:"user_type"`
	CreatedAt	string	`json:"created_at"`
	UpdatedAt 	string	`json:"updated_at"`
}

type SSOSignUpRequest struct {
	AuthToken 	string `json:"auth_token,omitempty"`
}

type SSOSignInResponse struct {
	AccessToken      string `json:"accessToken"`
	RefreshToken     string `json:"refreshToken"`
	RefreshExpiresIn int    `json:"refresh_expires_in"`
	SessionState string 	`json:"SessionState"`
	IssuedAt	 float64	`json:"issued_at"`
	ExpiresIn    float64    `json:"expiresIn"`
}

type ResetPasswordRequest struct {
	Email       string `json:"email"`
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

type GenericResponse struct {
	Success 	bool	`json:"success"`
	Message 	string	`json:"message"`
}

type GetUserByIdRequest struct {
	UserID 		int		`json:"user_id"`
	UserName 	*string	`json:"user_name"`
	Email 		*string	`json:"email"`
}

type AllUsersRequest struct {
	UserID		*int `json:"id"`
	Email  		*string `json:"email"`
	IsVerified  *bool   `json:"is_verified"`
	IsBlocked   *bool   `json:"is_blocked"`
	Limit       *int    `json:"limit"`
	Offset      *int    `json:"offset"`
}

type AllUsersResponse struct {
	Count 		int		`json:"count"`
	Users 		*domain.Users `json:"users"`
}

func (r *SSOSignInRequest) SSOSignInValidation() *errs.AppError {
	if strUtil.IsBlank(strings.TrimSpace(r.AuthToken)) {
		return errs.NewValidationError("Auth-Token cannot be empty")
	}
	return nil
}

func (r *GetUserByIdRequest) GetUserReqValidate() *errs.AppError {
	if r.UserID == 0 {
		return errs.NewValidationError("User ID Is Invalid. Please Provide a Correct ID.")
	}

	if r.UserName != nil && strUtil.IsBlank(strings.TrimSpace(*r.UserName)) {
		return errs.NewValidationError("Firstname cannot be empty")
	}

	if r.Email != nil && strUtil.IsBlank(strings.TrimSpace(*r.Email)) {
		return errs.NewValidationError("Email cannot be empty")
	}

	return nil
}

func (r *SignUpRequest) SignUpValidate() *errs.AppError{
	_, err := mail.ParseAddress(r.Email)
	if err != nil {
		return errs.NewValidationError("invalid email address")
	}

	if strUtil.IsBlank(strings.TrimSpace(r.UserName)) {
		return errs.NewValidationError("Firstname cannot be empty")
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

	userAccountType, mapErr := utility.MapUserAccountType(r.UserType)
	if mapErr !=nil && mapErr.Code == http.StatusUnprocessableEntity {
		return errs.NewValidationError(mapErr.Message)
	}
	r.UserType = userAccountType

	passwordValidatorError := utility.PasswordValidator(r.Password, false)
	if passwordValidatorError != nil {
		return passwordValidatorError
	}

	return nil
}

func (r *SignInRequest) SignInValidate() *errs.AppError {
	if strUtil.IsBlank(strings.TrimSpace(r.UserName)) {
		return errs.NewValidationError("User Name is empty")
	}

	if strUtil.IsBlank(strings.TrimSpace(r.Password)) {
		return errs.NewValidationError("Password is empty")
	}

	return nil
}

func(r *ResetPasswordRequest) OnDTO() *SignInRequest {
	return &SignInRequest{
		UserName: r.Email,
		Password: r.NewPassword,
	}
}

func(r *ResetPasswordRequest) RestPasswordValidation() *errs.AppError {
	if strUtil.IsBlank(strings.TrimSpace(r.Email)) {
		return errs.NewValidationError("Email is empty")
	}

	if strUtil.IsBlank(strings.TrimSpace(r.OldPassword)) {
		return errs.NewValidationError("Old Password is empty")
	}

	if strUtil.IsBlank(strings.TrimSpace(r.NewPassword)) {
		return errs.NewValidationError("New Password is empty")
	}

	return nil
}

func (r *AllUsersRequest) Validate() *errs.AppError {
	if r.Email != nil {
		_, err := mail.ParseAddress(*r.Email)
		if err != nil {
			return errs.NewValidationError("invalid email address")
		}
	}
	if r.UserID != nil {
		if *r.UserID == 0 {
			return errs.NewValidationError("User id Is invalid.")
		}
	}else {
		return errs.NewValidationError("User id is Blank.")
	}
	return nil
}