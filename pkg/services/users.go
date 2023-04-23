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
	"net/mail"
	"strings"
)

type UserServiceClass struct {
	repo 	domain.UserRepository
	mapper  mapper.RequestMapper
}


func NewUserServiceClass(repo domain.UserRepository) *UserServiceClass {
	return &UserServiceClass{repo: repo, mapper: mapper.NewRequestMapper(repo)}
}

type UserService interface {
	SignUp(ctx context.Context, request dto.SignUpRequest) (*dto.SignUpResponse, *errs.AppError)
	SignIn(ctx context.Context, request *dto.SignInRequest) (*dto.SignInResponse, *errs.AppError)
}


func (u *UserServiceClass) SignUp(ctx context.Context, request dto.SignUpRequest) (*dto.SignUpResponse, *errs.AppError) {
	err := request.SignUpValidate()
	if err != nil {
		return nil, err
	}
	/*
		1. Validate Email From Database.
		2. Validate Username Already Exist in DB.
		3. Create hashed + salt Password Encryption.
		4. Create Date and Time.
		5. Store Data.
	*/

	email, err2 := mail.ParseAddress(request.Email)
	if err2 != nil {
		logger.Debug(err2.Error())
		return nil, errs.NewUnexpectedError(err2.Error())
	}

	emailExist, err := u.repo.FindByEmail(ctx, strings.TrimSpace(email.Address))
	if emailExist {
		logger.Debug("Users already Exists.")
		return nil, errs.NewValidationError("user already exist")
	}else if err != nil  && err.Code != http.StatusInternalServerError {
		logger.Debug(err.Message)
		return nil, errs.NewUnexpectedError(err.Message)
	}

	usernameExist, err := u.repo.FindByUserName(ctx, request.UserName)
	if usernameExist {
		logger.Debug("User Name already Exists.")
		return nil, errs.NewValidationError("User Name already exist")
	}else if err != nil && err.Code != http.StatusInternalServerError {
		logger.Debug(err.Message)
		return nil, errs.NewUnexpectedError(err.Message)
	}

	/*
		1. Create Hashed + Salt Password Encryption.
	*/
	hashedPassword, err := utility.GenHashAndSaltPassword(request.Password)
	if stringUtils.IsBlank(hashedPassword) && err != nil {
		logger.Debug(err.Message)
		return nil, errs.NewValidationError(err.Message)
	}

	newUser := domain.CreateNewUser(request.UserName, request.FirstName, request.LastName, hashedPassword, email.Address, request.UserType ,request.Address ,request.PhoneNumber, false )
	userDetails, err := u.repo.SaveUser(ctx, newUser)
	if err != nil {
		logger.Debug(err.Message)
		return nil, errs.NewUnexpectedError(err.Message)
	}

	// Get Data From DTO
	resp := userDetails.ToSignUpDTO()

	return resp, nil
}

func (u *UserServiceClass) SignIn(ctx context.Context, request *dto.SignInRequest) (*dto.SignInResponse, *errs.AppError){
	err := u.mapper.ValidatedSignIn(ctx,request)
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