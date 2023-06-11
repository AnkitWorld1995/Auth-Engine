package services

import (
	"context"
	"github.com/agrison/go-commons-lang/stringUtils"
	"github.com/chsys/userauthenticationengine/pkg/client/sso"
	"github.com/chsys/userauthenticationengine/pkg/domain"
	"github.com/chsys/userauthenticationengine/pkg/dto"
	errs "github.com/chsys/userauthenticationengine/pkg/lib/error"
	"github.com/chsys/userauthenticationengine/pkg/lib/logger"
	"github.com/chsys/userauthenticationengine/pkg/lib/utility"
	"github.com/chsys/userauthenticationengine/pkg/mapper"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"strings"
	"sync"
)

type userServiceClass struct {
	repo  domain.UserRepository
	valid mapper.RequestValidationInterface
	keyCloak sso.KeyCloakMiddleware
}


func NewUserServiceClass(repo domain.UserRepository, keyCloakClient sso.KeyCloakMiddleware) *userServiceClass {
	return &userServiceClass{repo: repo, valid: &mapper.RequestValidation{
		Repo: repo,
	},
	keyCloak: keyCloakClient,
	}
}

type UserService interface {
	SignUp(ctx context.Context, request dto.SignUpRequest) (*dto.SignUpResponse, *errs.AppError)
	SignIn(ctx context.Context, request *dto.SignInRequest) (*dto.SignInResponse, *errs.AppError)
	SSOSignIn(ctx *gin.Context, request dto.SSOSignInRequest) (*dto.SSOSignInResponse, *errs.AppError)
	GetUserById(ctx *gin.Context, request dto.GetUserByIdRequest) (*dto.SignInResponse, *errs.AppError)
	CreateUser(ctx context.Context, request *dto.SignUpRequest) (*dto.SignUpResponse, *errs.AppError)
	ResetPassword(ctx context.Context, request *dto.ResetPasswordRequest) (*dto.GenericResponse, *errs.AppError)
	GetAllUser(ctx context.Context, request *dto.AllUsersRequest) (*dto.AllUsersResponse,*errs.AppError)
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

	userDetails, err :=  u.repo.GetUser(ctx, nil ,&request.UserName, &request.Email)
	if err != nil {
		logger.Debug(err.Message)
		return nil, errs.NewUnexpectedError(err.Message)
	}
	resp := userDetails.ToSignInDTO()
	return resp, nil
}

func (u *userServiceClass) SSOSignIn(ctx *gin.Context, request dto.SSOSignInRequest) (*dto.SSOSignInResponse, *errs.AppError) {
	appErr := request.SSOSignInValidation()
	if appErr != nil {
		return nil,appErr
	}

	var (
		respData domain.JWT
	)


	Valid , appErr := u.keyCloak.VerifyJWTToken(ctx, request.AuthToken)
	if appErr != nil || !Valid {
		log.Println("Service: GetUserById API ERROR", appErr)
		return nil, appErr
	}

	claims, appErr := u.keyCloak.GetClaims(ctx, request.AuthToken)
	if appErr != nil  {
		log.Println("Service: GetUserById API ERROR", appErr)
		return nil, appErr
	}

	respData.GetUserClaims(*claims)

	roleSize := len(respData.RealmAccess.Roles)
	for roleIdx, role := range respData.RealmAccess.Roles {
		if roleIdx <= roleSize && role == "admin-role" {
			break
		}else if roleIdx < roleSize && role != "admin-role" {
			continue
		}else {
			return nil, errs.NewForbiddenRequest("Un-Authorized access. User is not Admin.")
		}
	}

	_, appErr = u.valid.ValidateUserName(ctx, respData.PreferredUsername)
	if appErr != nil {
		return nil,errs.NewValidationError(appErr.Message)
	}

	_, appErr = u.valid.ValidateEmail(ctx, respData.Email)
	if appErr != nil {
		return nil, errs.NewValidationError(appErr.Message)
	}

	samePassword, appErr := u.valid.ValidatePassword(ctx, request.Password, respData.Email)
	if !samePassword || appErr != nil {
		return nil, errs.NewValidationError("Sorry! Invalid Password.")
	}

	tokenDetails, appErr := u.keyCloak.GetToken(ctx, respData, request.Password)
	if appErr != nil {
		return nil, errs.NewValidationError(appErr.Message)
	}


	respData.SetTokenDetails(*tokenDetails)
	resp := respData.SSOJWTDetails()
	return resp, nil
}

func (u *userServiceClass) GetUserById(ctx *gin.Context, request dto.GetUserByIdRequest) (*dto.SignInResponse, *errs.AppError) {
	appErr := request.GetUserReqValidate()
	if appErr != nil {
		return nil, appErr
	}

	validID, appErr := u.valid.ValidateUserID(ctx, request.UserID)
	if !validID || (appErr !=nil && appErr.Code == http.StatusNotFound) {
		return nil, errs.NewValidationError("User ID Not Found. Invalid User ID")
	}

	userDetails, appErr :=  u.repo.GetUser(ctx, &request.UserID, request.UserName, request.Email)
	if appErr != nil {
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

func (u *userServiceClass) GetAllUser(ctx context.Context, request *dto.AllUsersRequest) (*dto.AllUsersResponse,*errs.AppError){
	err := request.Validate()
	if err != nil {
		return nil, err
	}


	// Repository Function
	
	var wg sync.WaitGroup
	var users []*domain.Users
	var counts int32
	
	errChan := make(chan *errs.AppError, 2)
	wg.Add(1)
	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		count, err :=  u.repo.GetAllUsersCount(ctx, request)
		if err != nil {
			errChan <- err
		}
		counts = *count

	}(&wg)

	wg.Add(1)
	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		users, err =  u.repo.GetAllUsers(ctx, request)
		if err != nil {
			errChan <- err
		}

	}(&wg)

	go func() {
		wg.Wait()
		close(errChan)
	}()

	for errValue := range errChan {
		if errValue != nil {
			err = errValue
		}
	}

	if err != nil {
		return nil, err
	}

	 userList := domain.GetAllUser(users)

	resp := dto.AllUsersResponse{
		Count: counts,
		User: userList,
	}

	return &resp, nil
}