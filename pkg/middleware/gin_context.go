package middleware

import (
	"context"
	"github.com/gin-gonic/gin"
)

func GinContextToContextMiddleware() gin.HandlerFunc {
	return func(c *gin.Context){
		ctx := context.WithValue(c.Request.Context(), "GinContextKey", c)
		c.Request = c.Request.WithContext(ctx)
		//resp := &dto.SignInRequest{}
		//decode := json.NewDecoder(c.Request.Body)
		//if err := decode.Decode(resp); err != nil {
		//	c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
		//		"Success": false,
		//		"Message": "Bad Request",
		//	})
		//}
		//c.Set("Username", resp.Email)
		//c.Set("Password", resp.Password)
		c.Next()
	}
}