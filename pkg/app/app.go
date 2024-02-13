package app

import (
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/chsys/userauthenticationengine/config"
	"github.com/chsys/userauthenticationengine/pkg/client/db"
	_ "github.com/chsys/userauthenticationengine/pkg/client/db"
	"github.com/chsys/userauthenticationengine/pkg/domain"
	"github.com/chsys/userauthenticationengine/pkg/handler"
	"github.com/chsys/userauthenticationengine/pkg/middleware"
	"github.com/chsys/userauthenticationengine/pkg/services"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

/*
	1. Initializing Dependency Configurations.
	2. Wiring the Adapters and Ports With Handler (Gin/HTTP).
	3. Return Type For AWS Lambda.
*/

var newAWSSession *session.Session
var newDynamoDBSession *dynamodb.DynamoDB

func StartApp(appConfig *config.AppConfig, router *gin.Engine) *gin.Engine {

	// Start an AWS Session
	newAWSSession, err := session.NewSession(appConfig.AwsConfig)
	if err != nil {
		awsSessionRetry()
	}
	log.Println("------- appConfig AWS-----------", *appConfig.AwsConfig.Region, newAWSSession.Config.Region)

	// Start an AWS DynamoDB Session
	newDynamoDBSession := dynamodb.New(newAWSSession, aws.NewConfig().WithEndpoint("http://localhost:8000"))
	if newDynamoDBSession == nil {
		awsDynamoDbRetry()
	}

	// Check If The Dynamo-DB Table Exists.
	// If not Exists, Create One.
	func() {
		isValidExist := db.ValidateDynamoDbTable(newDynamoDBSession)
		if !isValidExist {
			go func() {
				//db.CreateDynamoDbTable(newDynamoDBSession)
			}()
		} else {
			//db.DropDynamoDbTable(newDynamoDBSession)
			//db.CreateDynamoDbTable(newDynamoDBSession)
		}
	}()

	/*
		Registering A Middleware.
	*/

	router.Use(middleware.RequestContextGinLogger())
	router.Use(middleware.GinContextToContextMiddleware())
	router.Use(middleware.GinCORSMiddleware())
	//router.Use(keyCloakMiddleware.VerifyJWTToken())

	pingHandler := handler.PingHandler{}
	router.Handle(http.MethodGet, "/ping2", pingHandler.Ping2())
	router.Handle(http.MethodGet, "/test2", pingHandler.Test())

	uploadHandler := handler.UploadHandler{UploadFileService: services.NewUploadFileService(newAWSSession, domain.IRepoNewClass(nil, nil, appConfig.RdmsDB.Schema, appConfig.MongoDB.Database, appConfig.MongoDB.UserCollection, newDynamoDBSession, newAWSSession))}
	router.Handle(http.MethodPost, "/upload", uploadHandler.UploadFileToS3())
	router.Handle(http.MethodPost, "/upload-All", uploadHandler.UploadAllFileToS3())
	router.Handle(http.MethodGet, "/read-upload-file", uploadHandler.ReadAllUploadFileS3())

	//userHandler  := handler.UserHandler{UserService: services.NewUserServiceClass(domain.IRepoNewClass(dbClient, mongoClient, appConfig.RdmsDB.Schema, appConfig.MongoDB.Database, appConfig.MongoDB.UserCollection), keyCloakMiddleware)}
	//router.Handle(http.MethodPost, "/sign-up", userHandler.SignUp())
	//router.Handle(http.MethodGet,  "/sign-in", userHandler.SignIn())
	//router.Handle(http.MethodGet, "/sso-sign-in", userHandler.SSOLogIn())
	//router.Handle(http.MethodGet, "/get-user", userHandler.GetUserById())
	//router.Handle(http.MethodPost, "/reset-password", userHandler.ResetPassword(keyCloakMiddleware))
	/*
		1. Register The Router a Method router.GET With Our Request Handler Function.
		2. In the handler function, we return the message back to client.
		3. Run The Router Using router.Run()
	*/
	router.GET("/test", func(context *gin.Context) {
		context.JSON(http.StatusOK, gin.H{
			"message": "Starting Server",
		})
	})

	return router
}

func awsSessionRetry() {
	for i := 0; i < 3; i++ {
		v := config.AppConfigs()
		if v != nil {
			initAwsSession()
			if newAWSSession != nil {
				break
			}
		}
	}
	log.Fatalln(fmt.Sprintf("Failed to Create a New S3 Session  with Error %s: ", errors.New("failed To Initialize AWS Session")))
}

func awsDynamoDbRetry() {
	for i := 0; i < 3; i++ {
		initDyDbSession()
		if newDynamoDBSession != nil {
			break
		}
	}
	log.Fatalln(fmt.Sprintf("Failed to Create a New S3 Session  with Error %s: ", errors.New("failed To Initialize DynamoDB Session")))
}

func initAwsSession() {
	sess, _ := session.NewSession(config.AppConfigs().AwsConfig)
	newAWSSession = sess
	log.Println("\n The New Aws Session", newAWSSession)
}

func initDyDbSession() {
	sess := dynamodb.New(newAWSSession, newAWSSession.Config)
	newDynamoDBSession = sess
	log.Println("\n The New Dynamo DB Session", newDynamoDBSession)
}
