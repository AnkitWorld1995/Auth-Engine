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

			if val, ok := ctx.Get("userMapKey"); !ok {
				ctx.JSON(http.StatusExpectationFailed, gin.H{
					"Success": false,
					"Message": "Failed To Get Keys",
				})
				ctx.Abort()
				return
			}else{
				marshalledValue, err := json.Marshal(val)
				if err != nil {
					ctx.JSON(http.StatusBadGateway, gin.H{
						"Success": false,
						"Message": "Failed To un Marshall",
					})
					ctx.Abort()
					return
				}

				isDataValid, errs := u.UserService.SSOSignIn(ctx, marshalledValue)
				if !isDataValid || (errs != nil && errs.Code == http.StatusUnprocessableEntity) {
					ctx.JSON(http.StatusUnprocessableEntity, gin.H{
						"Success": false,
						"Message": fmt.Sprintf("Invalid Request %s", errs.Message),
					})
					ctx.Abort()
					return
				}

				_, err = ctx.Writer.Write(marshalledValue)
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