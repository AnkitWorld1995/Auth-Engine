package handler

import (
	"encoding/json"
	"fmt"
	"github.com/chsys/userauthenticationengine/pkg/client/sso"
	"github.com/chsys/userauthenticationengine/pkg/dto"
	_ "github.com/chsys/userauthenticationengine/pkg/lib/constants"
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
				ctx.JSON(err.Code, gin.H{
					"Message": err.Message,
				})
				return
			} else {
				ctx.JSON(http.StatusOK, userDetails)
			}
		}
	}
}

func(u *UserHandler) SSOLogIn() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		resp := new(dto.SSOSignInRequest)
		err := json.NewDecoder(ctx.Request.Body).Decode(&resp)
		if err != nil {
			if marshallingErr, ok := err.(*json.UnmarshalTypeError); ok {
				ctx.JSON(http.StatusBadRequest, gin.H{
					"Message": fmt.Sprintf("The field %s must be a %s", marshallingErr.Field, marshallingErr.Type.String()),
				})
				return
			}
		} else {
			response, errs := u.UserService.SSOSignIn(ctx, *resp)
			if errs != nil  {
				ctx.JSON(errs.Code, gin.H{
					"Message": fmt.Sprintf("Invalid Request: %s", errs.Message),
				})
				return
			} else {
			ctx.JSON(http.StatusFound, response)
			return
			}
		}
	}
}

func (u *UserHandler) GetUserById() gin.HandlerFunc{
	return func (ctx *gin.Context){
		var request dto.GetUserByIdRequest
		err := json.NewDecoder(ctx.Request.Body).Decode(&request)
		if err != nil {
			if marshallingErr, ok := err.(*json.UnmarshalTypeError); ok{
				ctx.JSON(http.StatusBadRequest, gin.H{
					"Message": fmt.Sprintf("The field %s must be a %s", marshallingErr.Field, marshallingErr.Type.String()),
				})
				return
			}
		}else {
			getUser, err := u.UserService.GetUserById(ctx, request)
			if err != nil {
				if err != nil {
					ctx.JSON(err.Code, gin.H{
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
				return
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