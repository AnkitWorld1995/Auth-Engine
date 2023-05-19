package sso

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Nerzal/gocloak/v7"
	"github.com/agrison/go-commons-lang/stringUtils"
	"github.com/chsys/userauthenticationengine/pkg/domain"
	"github.com/chsys/userauthenticationengine/pkg/dto"
	"github.com/chsys/userauthenticationengine/pkg/lib/constants"
	errs "github.com/chsys/userauthenticationengine/pkg/lib/error"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

type KeyCloakContracts interface {
	extractBearerToken(token string) string
	ExtractAccessTokenData(ctx *gin.Context) *gin.Context
	GetUserWithJWTCred(ctx *gin.Context, req *dto.SignInRequest) *gin.Context
	VerifyJWTToken(ctx *gin.Context) *gin.Context
	ResetPassword(ctx *gin.Context, req *dto.ResetPasswordRequest) *gin.Context
}

func (auth *KeyCloakMiddleware) extractBearerToken(token string) string {
	return strings.Replace(token, "Bearer ", "", 1)
}

func (auth *KeyCloakMiddleware) ExtractAccessTokenData(ctx *gin.Context) (*gin.Context, *errs.AppError)  {

	tokenHeader := ctx.GetHeader("Authorization")
	if stringUtils.IsBlank(tokenHeader) {
		return nil, errs.NewNotFoundError("Authorization JwtHeader Missing.")
	}

	token := auth.extractBearerToken(tokenHeader)
	info, err := auth.Keycloak.GoCloak.GetUserInfo(ctx, token, auth.Keycloak.Realm )
	if err != nil {
		return nil, errs.NewNotFoundError("Invalid Token.")
	}

	infos, _ := json.Marshal(info)
	userDetails := make(map[string]string)
	userDetails[constants.UserContextDetails] = string(infos)
	userDetails[constants.UserContextJwtDetails]  = token
	ctx.Set(constants.UserContextDetails, userDetails)
	return ctx, nil
}


func(auth *KeyCloakMiddleware) GetToken(ctx *gin.Context, req domain.JWT, password string) (*domain.TokenDetails, *errs.AppError) {
	var (
		cred 		 domain.TokenDetails
		tokenOptions domain.TokenOptions
		grantType    = constants.GrantTypePassword
	)

	tokenOptions = domain.TokenOptions{
		ClientID: 		&auth.Keycloak.ClientId,
		ClientSecret: 	&auth.Keycloak.ClientSecret,
		Username: 		&req.PreferredUsername,
		Password: 		&password,
		GrantType:      &grantType,
	}

	jwtCred, err := auth.Keycloak.GoCloak.GetToken(context.Background(), auth.Keycloak.Realm, gocloak.TokenOptions(tokenOptions))
	if err != nil {
		return nil, errs.NewNotFoundError( fmt.Sprintf("Authorization error %s", err.Error()))
	}

	cred = domain.TokenDetails{
		AccessToken:  		jwtCred.AccessToken,
		RefreshToken: 		jwtCred.RefreshToken,
		ExpiresIn:    		jwtCred.ExpiresIn,
		RefreshExpiresIn: jwtCred.RefreshExpiresIn,
	}
	return &cred, nil
}

func (auth *KeyCloakMiddleware) VerifyJWTToken(ctx *gin.Context, authToken string) (bool,*errs.AppError) {

	// call Keycloak API to verify the access token
	result, err := auth.Keycloak.GoCloak.RetrospectToken(context.Background(), authToken, auth.Keycloak.ClientId, auth.Keycloak.ClientSecret, auth.Keycloak.Realm)
	if err != nil {
		return false, errs.NewUnexpectedError(fmt.Sprintf("Invalid or malformed token: %s", err.Error()))
	}

	// check if the token isn't expired and valid
	if !*result.Active {
		return false, errs.NewValidationError("Token Expired.")
	}

	return true, nil
}


func (auth *KeyCloakMiddleware) GetClaims(ctx *gin.Context ,token string) (*dto.MapClaims, *errs.AppError) {
	_,jwtClaims, err := auth.Keycloak.GoCloak.DecodeAccessToken(ctx, token, auth.Keycloak.Realm, "")
	if err != nil {
		return nil, errs.NewUnexpectedError(err.Error())
	}

	return (*dto.MapClaims)(jwtClaims), nil
}

func (auth *KeyCloakMiddleware) ResetPassword(ctx *gin.Context, req *dto.ResetPasswordRequest) *gin.Context {

	resp := req.OnDTO()
	newContext, appErr := auth.ExtractAccessTokenData(ctx)
	if appErr != nil {
		return nil
	}
	contextMap, ok := newContext.Get("User-Info")
	if !ok {
		ctx.AbortWithStatusJSON(http.StatusUnprocessableEntity, gin.H{
			"Success": false,
			"Message": fmt.Sprintf("Context Procession Failed With Data: %s", contextMap),
		})
		ctx.Abort()
		return ctx
	}

	ctxMapValue  := contextMap.(map[string]string)
	ctxUserValue := ctxMapValue[constants.UserContextDetails]
	accessToken  := ctxMapValue[constants.UserContextJwtDetails]

	var userCloakDetails *domain.UserInfo
	err := json.Unmarshal([]byte(ctxUserValue), &userCloakDetails)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusUnprocessableEntity, gin.H{
			"Success": false,
			"Message": fmt.Sprintf("Context Procession Failed With Data: %s", err.Error()),
		})
		ctx.Abort()
		return ctx
	}

	err = auth.Keycloak.GoCloak.SetPassword(ctx, accessToken, *userCloakDetails.Sub, auth.Keycloak.Realm, resp.Password, false)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusUnprocessableEntity, gin.H{
			"Success": false,
			"Message": fmt.Sprintf("Context Procession Failed To Set Password: %s", err.Error()),
		})
		ctx.Abort()
		return ctx
	}
	return ctx
}