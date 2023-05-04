package sso

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/agrison/go-commons-lang/stringUtils"
	"github.com/chsys/userauthenticationengine/pkg/domain"
	"github.com/chsys/userauthenticationengine/pkg/dto"
	"github.com/chsys/userauthenticationengine/pkg/lib/constants"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"strings"
)

func (auth *KeyCloakMiddleware) extractBearerToken(token string) string {
	return strings.Replace(token, "Bearer ", "", 1)
}

func (auth *KeyCloakMiddleware) ExtractAccessTokenData(ctx *gin.Context) *gin.Context {

	tokenHeader := ctx.GetHeader("Authorization")
	if stringUtils.IsBlank(tokenHeader) {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"Success": false,
			"Message": "Authorization JwtHeader Missing.",
		})
		return &gin.Context{}
	}

	token := auth.extractBearerToken(tokenHeader)
	info, err := auth.Keycloak.GoCloak.GetUserInfo(ctx, token, auth.Keycloak.Realm )
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"Success": false,
			"Message": "Data Extraction Failed",
		})
		return &gin.Context{}
	}

	infos, _ := json.Marshal(info)
	userDetails := make(map[string]string)
	userDetails[constants.UserContextDetails] = string(infos)
	userDetails[constants.UserContextJwtDetails]  = token
	ctx.Set(constants.UserContextDetails, userDetails)
	return ctx
}


func(auth *KeyCloakMiddleware) GetUserWithJWTCred(ctx *gin.Context, req *dto.SignInRequest) *gin.Context {
	var cred *domain.JWTRequest

	jwtCred, err := auth.Keycloak.GoCloak.Login(context.Background(),auth.Keycloak.ClientId, auth.Keycloak.ClientSecret, auth.Keycloak.Realm, req.UserName, req.Password)
	if err != nil {
		log.Println(err.Error())
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"Success": false,
			"Message": "Authorization error",
		})
		ctx.Abort()
		return &gin.Context{}
	}

	cred = &domain.JWTRequest{
		Username:     req.UserName,
		Email: 		  req.Email,
		Password:     req.Password,
		AccessToken:  jwtCred.AccessToken,
		RefreshToken: jwtCred.RefreshToken,
		ExpiresIn:    jwtCred.ExpiresIn,
	}

	credJson, _ := json.Marshal(cred)
	userMapKey := make(map[string]string)
	userMapKey[constants.UserCredentials] = string(credJson)
	ctx.Set(constants.UserMapKey, userMapKey)
	return ctx
}

func (auth *KeyCloakMiddleware) VerifyJWTToken(ctx *gin.Context) *gin.Context{
	newContext := auth.ExtractAccessTokenData(ctx)
	contextMap, ok := newContext.Get(constants.UserContextDetails)
	if !ok {
		ctx.AbortWithStatusJSON(http.StatusUnprocessableEntity, gin.H{
			"Success": false,
			"Message": fmt.Sprintf("Context Procession Failed With Data: %s", contextMap),
		})
		ctx.Abort()
		return ctx
	}

	ctxMapValue := contextMap.(map[string]string)
	token := ctxMapValue[constants.UserContextJwtDetails]

	// call Keycloak API to verify the access token
	result, err := auth.Keycloak.GoCloak.RetrospectToken(context.Background(), token, auth.Keycloak.ClientId, auth.Keycloak.ClientSecret, auth.Keycloak.Realm)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"Success": false,
			"Message": fmt.Sprintf("Invalid or malformed token: %s", err.Error()),
		})
		ctx.Abort()
		return &gin.Context{}
	}

	// check if the token isn't expired and valid
	if !*result.Active {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"Success": false,
			"Message": fmt.Sprintf("Invalid or malformed token: %v", *result.Active),
		})
		return &gin.Context{}
	}
	ctx.Next()
	return ctx
}

func (auth *KeyCloakMiddleware) ResetPassword(ctx *gin.Context, req *dto.ResetPasswordRequest) *gin.Context {

	resp := req.OnDTO()
	newContext := auth.ExtractAccessTokenData(ctx)
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