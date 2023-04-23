package handler

import (
	"encoding/json"
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