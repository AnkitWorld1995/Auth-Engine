package handler

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type PingHandler struct {}

func (p *PingHandler) Ping2() gin.HandlerFunc  {
	return func(context *gin.Context) {
		context.JSON(http.StatusOK, gin.H{
			"Success": true,
			"Message": "Pong @ Welcome to HVTC",
		})
	}
}


func (p *PingHandler) Test() gin.HandlerFunc  {
	return func(context *gin.Context) {
		context.JSON(http.StatusOK, gin.H{
			"Success": true,
			"Message": "Server Running!!!!!!!!!!!!!",
		})
	}
}
