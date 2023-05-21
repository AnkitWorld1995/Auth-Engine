package main

import (
	"context"
	ginadapter "github.com/awslabs/aws-lambda-go-api-proxy/gin"
	"github.com/chsys/userauthenticationengine/config"
	"github.com/chsys/userauthenticationengine/pkg/app"
	"github.com/joho/godotenv"
	"log"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

var ginLambda *ginadapter.GinLambda

/*
	Main Function To Initialize the project.
*/
func init(){
	err := godotenv.Load("app.env")
	if err != nil {
		log.Fatalln("Error loading .env file: " + err.Error())
	}
	config.Init()
}

// Handler AWS Lambda Function handler.
func Handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// If no name is provided in the HTTP request body, throw an error
	return ginLambda.ProxyWithContext(ctx, req)
}

func main() {
	configuration := config.AppConfigs()
	ginLambda = app.StartApp(configuration, ginLambda)
	lambda.Start(Handler)
}
