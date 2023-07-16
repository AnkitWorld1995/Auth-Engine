package main

import (
	"context"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	ginadapter "github.com/awslabs/aws-lambda-go-api-proxy/gin"
	"github.com/chsys/userauthenticationengine/config"
	"github.com/chsys/userauthenticationengine/pkg/app"
	"github.com/gin-gonic/gin"
	"log"
)

var ginLambda *ginadapter.GinLambdaV2

/*
	Main Function To Initialize the project.
*/
func init(){
	config.Init()
	configuration := config.AppConfigs()
	r := gin.Default()
	r = app.StartApp(configuration, r)
	ginLambda = ginadapter.NewV2(r)
}

// Handler AWS Lambda Function handler (Gin Adapter specific For HTTP AWS Gateway).
func Handler(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	// If no name is provided in the HTTP request body, throw an error
	log.Println(" ======== AWS Lambda Request handler ========: ", req.Body, req.RawPath, req.PathParameters)
	return ginLambda.ProxyWithContext(ctx, req)
}

func main() {
	lambda.Start(Handler)
}
