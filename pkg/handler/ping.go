package handler

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type PingHandler struct {}

func (p *PingHandler) Ping() gin.HandlerFunc  {
	return func(context *gin.Context) {
		context.JSON(http.StatusOK, gin.H{
			"Success": true,
			"Message": "Welcome to HVTC",
		})
	}
}
