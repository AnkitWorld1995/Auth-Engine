package services

import (
	"context"
	"github.com/agrison/go-commons-lang/stringUtils"
	"github.com/chsys/userauthenticationengine/pkg/domain"
	"github.com/chsys/userauthenticationengine/pkg/dto"
	errs "github.com/chsys/userauthenticationengine/pkg/lib/error"
	"github.com/chsys/userauthenticationengine/pkg/lib/logger"
	"github.com/chsys/userauthenticationengine/pkg/lib/utility"
	"github.com/chsys/userauthenticationengine/pkg/mapper"
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
}



func (u *userServiceClass) SignUp(ctx context.Context, request dto.SignUpRequest) (*dto.SignUpResponse, *errs.AppError) {
	err := request.SignUpValidate()
	if err != nil {
		return nil, err
	}

	emailExist, err := u.valid.ValidateEmail(ctx, strings.TrimSpace(request.Email))
	if emailExist || (err != nil  && err.Code != http.StatusInternalServerError) {
		return nil, errs.NewValidationError("user already exist")
	}

	usernameExist, err := u.valid.ValidateUserName(ctx, request.UserName)
	if usernameExist || (err != nil  && err.Code != http.StatusInternalServerError) {
		return nil, errs.NewValidationError("user already exist")
	}

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

func (u *userServiceClass) SignIn(ctx context.Context, request *dto.SignInRequest) (*dto.SignInResponse, *errs.AppError){
	err := request.SignInValidate()
	if err != nil {
		return nil, err
	}
	_, err = u.valid.ValidateUserName(ctx,  request.UserName)
	if err != nil {
		return nil, err
	}

	err = u.valid.ValidatePassword(ctx, request.Password, request.UserName)
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