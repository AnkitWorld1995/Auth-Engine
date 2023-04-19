package handler

import (
	"encoding/json"
	"github.com/chsys/userauthenticationengine/pkg/dto"
	"github.com/chsys/userauthenticationengine/pkg/services"
	"github.com/gin-gonic/gin"
	"net/http"
)

//type UserHandler struct {
//	UserName 		string
//	FirstName 		string
//	LastName 		string
//	EmailAddress 	string
//	PhoneNumber 	string
//	Address 		*string
//}


type UserHandler struct {
	UserService services.UserService
}

func (u *UserHandler) SignUp() gin.HandlerFunc {
	return func(c *gin.Context) {
		var request dto.SignUpRequest
		err := json.NewDecoder(c.Request.Body).Decode(&request)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"Message": err.Error(),
			})
			return
		}
	}
}