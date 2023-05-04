package services

import (
	"context"
	"encoding/json"
	"github.com/agrison/go-commons-lang/stringUtils"
	"github.com/chsys/userauthenticationengine/pkg/domain"
	"github.com/chsys/userauthenticationengine/pkg/dto"
	errs "github.com/chsys/userauthenticationengine/pkg/lib/error"
	"github.com/chsys/userauthenticationengine/pkg/lib/logger"
	"github.com/chsys/userauthenticationengine/pkg/lib/utility"
	"github.com/chsys/userauthenticationengine/pkg/mapper"
	"log"
	"net/http"
	"strings"
)

type userServiceClass struct {
	repo  domain.UserRepository
	valid mapper.RequestValidationInterface
}


func NewUserServiceClass(repo domain.UserRepository) *userServiceClass {
	return &userServiceClass{repo: repo, valid: &mapper.RequestValidation{
		Repo: repo,
	}}
}

type UserService interface {
	SignUp(ctx context.Context, request dto.SignUpRequest) (*dto.SignUpResponse, *errs.AppError)
	SignIn(ctx context.Context, request *dto.SignInRequest) (*dto.SignInResponse, *errs.AppError)
	SSOSignIn(ctx context.Context, data string) (*dto.JWTResponse,bool, *errs.AppError)
	GetUser(ctx context.Context, data any) (*dto.SignInResponse, *errs.AppError)
	CreateUser(ctx context.Context, request *dto.SignUpRequest) (*dto.SignUpResponse, *errs.AppError)
	ResetPassword(ctx context.Context, request *dto.ResetPasswordRequest) (*dto.GenericResponse, *errs.AppError)
}



func (u *userServiceClass) SignUp(ctx context.Context, request dto.SignUpRequest) (*dto.SignUpResponse, *errs.AppError) {
	err := request.SignUpValidate()
	if err != nil {
		return nil, err
	}

	emailExist, err := u.valid.ValidateEmail(ctx, strings.TrimSpace(request.Email))
	if emailExist || (err != nil  && err.Code != http.StatusNotFound) {
		return nil, errs.NewValidationError("Email already exist")
	}

	usernameExist, err := u.valid.ValidateUserName(ctx, request.UserName)
	if usernameExist || (err != nil  && err.Code != http.StatusNotFound) {
		return nil, errs.NewValidationError("user already exist")
	}

	resp, err := u.CreateUser(ctx, &request)
	if err != nil {
		return nil, err
	}

	return resp, nil
}


/*
	Note: Decommission SignIn Method if SSO-LogIn and Get-User Method IS Fully Up and Functional.
*/
func (u *userServiceClass) SignIn(ctx context.Context, request *dto.SignInRequest) (*dto.SignInResponse, *errs.AppError){
	err := request.SignInValidate()
	if err != nil {
		return nil, err
	}
	_, err = u.valid.ValidateUserName(ctx,  request.UserName)
	if err != nil {
		return nil, err
	}

	_, err = u.valid.ValidatePassword(ctx, request.Password, request.UserName)
	if err != nil {
		return nil, err
	}

	userDetails, err :=  u.repo.GetUser(ctx, &request.UserName)
	if err != nil {
		logger.Debug(err.Message)
		return nil, errs.NewUnexpectedError(err.Message)
	}
	resp := userDetails.ToSignInDTO()
	return resp, nil
}

func (u *userServiceClass) SSOSignIn(ctx context.Context, data string) (*dto.JWTResponse,bool, *errs.AppError) {
	var reqData domain.JWTRequest

	err := json.Unmarshal([]byte(data), &reqData)
	if err != nil {
		return nil,false, errs.NewValidationError(err.Error())
	}

	_, appErr := u.valid.ValidateUserName(ctx,  reqData.Username)
	if appErr != nil {
		return nil, false, errs.NewValidationError(appErr.Message)
	}

	_, appErr = u.valid.ValidatePassword(ctx, reqData.Password, reqData.Username)
	if appErr != nil {
		return nil, false, errs.NewValidationError(appErr.Message)
	}

	resp := reqData.ToDTOJwtResponse()
	return resp, true, nil
}

func (u *userServiceClass) GetUser(ctx context.Context, data any) (*dto.SignInResponse, *errs.AppError) {
	var respData domain.UserInfo
	val := data.(map[string]string)
	userData := val["User-Info"]

	err := json.Unmarshal([]byte(userData), &respData)
	if err != nil {
		log.Println(err.Error())
		return nil, errs.NewValidationError(err.Error())
	}

	userDetails, appErr :=  u.repo.GetUser(ctx, respData.PreferredUsername)
	if err != nil {
		logger.Debug(appErr.Message)
		return nil, errs.NewUnexpectedError(appErr.Message)
	}
	resp := userDetails.ToSignInDTO()
	return resp, nil
}

func(u *userServiceClass) CreateUser(ctx context.Context, request *dto.SignUpRequest) (*dto.SignUpResponse, *errs.AppError) {

	hashedPassword, err := utility.GenHashAndSaltPassword(request.Password)
	if stringUtils.IsBlank(hashedPassword) && err != nil {
		logger.Debug(err.Message)
		return nil, errs.NewValidationError(err.Message)
	}

	newUser := domain.CreateNewUser(request.UserName, request.FirstName, request.LastName, hashedPassword, utility.ParseMail(request.Email), request.UserType ,request.Address ,request.PhoneNumber, false )
	userDetails, err := u.repo.SaveUser(ctx, newUser)
	if err != nil {
		logger.Debug(err.Message)
		return nil, errs.NewUnexpectedError(err.Message)
	}

	resp := userDetails.ToSignUpDTO()
	return resp, nil
}

func (u *userServiceClass) ResetPassword(ctx context.Context, request *dto.ResetPasswordRequest) (*dto.GenericResponse, *errs.AppError) {
	err := request.RestPasswordValidation()
	if err != nil {
		return &dto.GenericResponse{
				Success: false,
				Message: err.Message,
		}, err
	}

	samePassword, err := u.valid.ValidatePassword(ctx, request.NewPassword, request.Email)
	if err != nil && err.Code != http.StatusUnprocessableEntity {
		return &dto.GenericResponse{
				Success: false,
				Message: err.Message,
		},  errs.NewValidationError(err.Message)
	}else if samePassword && err == nil{
		return &dto.GenericResponse{
			Success: false,
			Message: "Sorry! Cannot Use Same old Password.",
		},  errs.NewValidationError("Sorry! Cannot Use Same old Password.")
	}

	hashedPassword, err := utility.GenHashAndSaltPassword(request.NewPassword)
	if stringUtils.IsBlank(hashedPassword) && err != nil {
		logger.Debug(err.Message)
		return &dto.GenericResponse{
				Success: false,
				Message: err.Message,
		}, errs.NewValidationError(err.Message)
	}
	// Update the Same In Database.
	// Send Notification Via kafka or aws SNS

	resp, err := u.repo.UpdatePassword(ctx, request.Email, hashedPassword)
	if err != nil {
		return resp, err
	}

	return &dto.GenericResponse{
		Success: true,
		Message: "Password Updated Successfully.",
	}, nil
}