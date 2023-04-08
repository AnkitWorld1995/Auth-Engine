package middleware

import (
	"github.com/google/uuid"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"time"
)

/*
	1. Setting A RequestContextGinLogger Middleware For All Incoming Request.
*/

func RequestContextGinLogger() gin.HandlerFunc {
	return func(context *gin.Context){

		requestID := uuid.New().String()
		location, err := time.LoadLocation("Asia/Kolkata")
		if err != nil {
			context.JSON(http.StatusServiceUnavailable, gin.H{
				"Success": false,
				"Message": "Service Unavailable",
			})
			context.Next()
		}
		start := time.Now().In(location).Format("2006-01-02T15:04:05 -07:00:00")
		clientIP := context.ClientIP()
		log.Printf(" Time: %s \n  Requested Method: %s, URL: %s, ClientIP: %s", start ,context.Request.Method, context.Request.URL, clientIP)


		/*
			1. Setting the Incoming Request With An Unique Universal Unique Identifier and the Time request was made.
			2. Moving to Next Handler with .Next().
		*/
		context.Set("ID", 	 requestID)
		context.Set("Time",  start)
		context.Next()
	}
}