package sso

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Nerzal/gocloak/v7"
	"github.com/agrison/go-commons-lang/stringUtils"
	"github.com/chsys/userauthenticationengine/pkg/domain"
	"github.com/chsys/userauthenticationengine/pkg/dto"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"strings"
)

type KeyCloakMiddleware struct {
	Keycloak *KeyCloak
}


type KeyCloak struct {
	GoCloak		 gocloak.GoCloak
	ClientId     string
	ClientSecret string
	Realm        string
}


func KeyCloakInit(keycloak *KeyCloak) *KeyCloakMiddleware {
	return &KeyCloakMiddleware{Keycloak: keycloak}
}



func (auth *KeyCloakMiddleware) extractBearerToken(token string) string {
	return strings.Replace(token, "Bearer ", "", 1)
}

func (auth *KeyCloakMiddleware) extractAccessTokenData(ctx *gin.Context) *gin.Context {

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
	userDetails["User-Info"] = string(infos)
	userDetails["User-JWT"]  = token
	ctx.Set("User-Info", userDetails)
	return ctx
}


func(auth *KeyCloakMiddleware) ExtractJWTCred(ctx *gin.Context, req *dto.SignInRequest) *gin.Context {
	var cred *domain.JWTCredentials

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
	log.Println(ctx.Keys)

	cred = &domain.JWTCredentials{
		Username:     req.UserName,
		Password:     req.Password,
		AccessToken:  jwtCred.AccessToken,
		RefreshToken: jwtCred.RefreshToken,
		ExpiresIn:    jwtCred.ExpiresIn,
	}

	userMapKey := make(map[string]any)
	userMapKey["userCred"] = cred
	ctx.Set("userMapKey", userMapKey)
	return ctx
}

func (auth *KeyCloakMiddleware) VerifyJWTToken(ctx *gin.Context) *gin.Context{
	newContext := auth.extractAccessTokenData(ctx)
	contextMap, ok := newContext.Get("User-Info")
	if !ok {
		ctx.AbortWithStatusJSON(http.StatusUnprocessableEntity, gin.H{
			"Success": false,
			"Message": fmt.Sprintf("Context Procession Failed With Data: %s", contextMap),
		})
		ctx.Abort()
		return ctx
	}

	ctxMapValue := contextMap.(map[string]string)
	token := ctxMapValue["User-JWT"]

	log.Println("-------------2---------------------")

	//// call Keycloak API to verify the access token
	result, err := auth.Keycloak.GoCloak.RetrospectToken(context.Background(), token, auth.Keycloak.ClientId, auth.Keycloak.ClientSecret, auth.Keycloak.Realm)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"Success": false,
			"Message": fmt.Sprintf("Invalid or malformed token: %s", err.Error()),
		})
		ctx.Abort()
		return &gin.Context{}
	}
	log.Println("-------------3---------------------")


	// check if the token isn't expired and valid
	if !*result.Active {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"Success": false,
			"Message": fmt.Sprintf("Invalid or malformed token: %v", *result.Active),
		})
		return &gin.Context{}
	}

	log.Println("-------------5---------------------")
	ctx.Next()

	return ctx
}
