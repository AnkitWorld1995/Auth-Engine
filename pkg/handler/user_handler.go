package handler

import (
	"encoding/json"
	"fmt"
	"github.com/chsys/userauthenticationengine/pkg/client/sso"
	"github.com/chsys/userauthenticationengine/pkg/dto"
	"github.com/chsys/userauthenticationengine/pkg/services"
	"github.com/gin-gonic/gin"
	"net/http"
)


type UserHandler struct {
	UserService services.UserService
}

func (u *UserHandler) SignUp() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var request dto.SignUpRequest
		err := json.NewDecoder(ctx.Request.Body).Decode(&request)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"Message": err.Error(),
			})
			return
		}else {
			userRegister, err := u.UserService.SignUp(ctx, request)
			if err != nil {
				if err != nil {
					ctx.JSON(http.StatusBadRequest, gin.H{
						"Message": err.Message,
					})
					return
				}
			}else{
				ctx.JSON(http.StatusCreated, userRegister)
				return
			}
		}
	}
}

func(u *UserHandler) SSOLogIn(auth sso.KeyCloakMiddleware) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		resp := new(dto.SignInRequest)
		err := json.NewDecoder(ctx.Request.Body).Decode(&resp)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"Message": err.Error(),
			})
			return
		}

		ctx = auth.ExtractJWTCred( ctx, resp)
		if ctx.IsAborted(){

			ctx.JSON(http.StatusExpectationFailed, gin.H{
				"Success": false,
				"Message": "Authorization error in handler",
			})
			ctx.Abort()
			return

		}else {

			if value, ok := ctx.Get("userMapKey"); !ok {
				ctx.JSON(http.StatusExpectationFailed, gin.H{
					"Success": false,
					"Message": "Failed To Get Keys",
				})
				ctx.Abort()
				return
			}else{

				mapUserContext := value.(map[string]string)
				userContextValue, ok := mapUserContext["userCred"]
				if !ok {
					ctx.JSON(http.StatusBadGateway, gin.H{
						"Success": false,
						"Message": "Failed To Fetch Data.",
					})
					ctx.Abort()
					return
				}

				response,isDataValid, errs := u.UserService.SSOSignIn(ctx, userContextValue)
				if !isDataValid || (errs != nil && errs.Code == http.StatusUnprocessableEntity) {
					ctx.JSON(http.StatusUnprocessableEntity, gin.H{
						"Success": false,
						"Message": fmt.Sprintf("Invalid Request %s", errs.Message),
					})
					ctx.Abort()
					return
				}

				marshalledResp, _ := json.Marshal(response)
				_, err = ctx.Writer.Write(marshalledResp)
				if err != nil {
					return
				}
			}
		}
	}
}

func (u *UserHandler) GetUser(auth sso.KeyCloakMiddleware) gin.HandlerFunc{
	return func (ctx *gin.Context){
		auth.VerifyJWTToken(ctx)
		if ctx.IsAborted(){
			ctx.JSON(http.StatusExpectationFailed, gin.H{
				"Success": false,
				"Message": "Authorization error in handler",
			})
			ctx.Abort()
			return
		}else {

			if val, ok := ctx.Get("User-Info"); !ok {
				ctx.JSON(http.StatusExpectationFailed, gin.H{
					"Success": false,
					"Message": "Failed To User-Info Keys",
				})
				ctx.Abort()
				return
			}else{
				getUser, err := u.UserService.GetUser(ctx, val)
				if err != nil {
					if err != nil {
						ctx.JSON(http.StatusBadRequest, gin.H{
							"Message": err.Message,
						})
						return
					}
				}else{
					ctx.JSON(http.StatusFound, getUser)
					return
				}
			}
		}
	}
}

func (u *UserHandler) SignIn() gin.HandlerFunc{
	return func (ctx *gin.Context){
		resp := new(dto.SignInRequest)
		err := json.NewDecoder(ctx.Request.Body).Decode(&resp)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"Message": err.Error(),
			})
			return
		}else {
			userDetails, err := u.UserService.SignIn(ctx, resp)
			if err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{
					"Message": err.Message,
				})
				return
			} else {
				ctx.JSON(http.StatusOK, userDetails)
			}
		}
	}
}

func (u *UserHandler) ResetPassword(auth sso.KeyCloakMiddleware) gin.HandlerFunc{
	return func(ctx *gin.Context) {
		var request dto.ResetPasswordRequest
		err := json.NewDecoder(ctx.Request.Body).Decode(&request)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"Message": err.Error(),
			})
			return
		}else{

			response, err := u.UserService.ResetPassword(ctx, &request)
			if err != nil {
				ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
					"Message": response,
				})
			}

			/*
				Note: Need Role To Provide access To Such Method in Key-Cloak.
				1. auth.ResetPassword() Can be Omitted if the Password in Key-Cloak
			       server is already reset via Client.
			*/
			context := auth.ResetPassword(ctx, &request)
			if context.IsAborted() {
				context.JSON(http.StatusExpectationFailed, gin.H{
					"Success": false,
					"Message": "Password Reset Ramification Failed.",
				})
				context.Abort()
				return
			}

			ctx.JSON(http.StatusCreated, response)
			return
		}
	}
}