package app

import (
	"fmt"
	"github.com/chsys/userauthenticationengine/config"
	"github.com/chsys/userauthenticationengine/pkg/client/db"
	"github.com/chsys/userauthenticationengine/pkg/client/sso"
	"github.com/chsys/userauthenticationengine/pkg/domain"
	"github.com/chsys/userauthenticationengine/pkg/handler"
	"github.com/chsys/userauthenticationengine/pkg/lib/utility"
	"github.com/chsys/userauthenticationengine/pkg/middleware"
	"github.com/chsys/userauthenticationengine/pkg/services"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

/*
	1. Initializing Dependency Configurations.
	2. Wiring the Adapters and Ports With Handler (Gin/HTTP).
*/

func StartApp(config *config.AppConfig) {

	/*
		Start RDMS Database.
	*/
	dbClient, err := db.RdmsInit(config.RdmsDB)
	if err != nil {
		log.Fatalln("R-Database Error",err.Error())
	}
	log.Println("The DB ClientSecret", dbClient)

	/*
		Start No-SQL Database.
	*/
	mongoClient, err := db.MongoInit(config.MongoDB)
	if err != nil{
		log.Fatalln("Mongo ClientSecret Error",err.Error())
	}
	log.Println("The Mongo ClientSecret", mongoClient)

	keyCloakClient := sso.KeyCloakInit(config.KeyCloak)
	log.Println("The Key-Cloak client", keyCloakClient)

	port := utility.ReadPort()
	log.Printf("Starting Server on http://localhost:%v", port)

	//Create A Router Using Gin.
	router := gin.Default()
	keyCloakMiddleware:= sso.KeyCloakMiddleware{Keycloak: config.KeyCloak}

	/*
		Registering A Middleware.
	*/

	router.Use(middleware.RequestContextGinLogger())
	router.Use(middleware.GinContextToContextMiddleware())
	router.Use(middleware.GinCORSMiddleware())
	//router.Use(keyCloakMiddleware.VerifyJWTToken())

	pingHandler := handler.PingHandler{}
	router.Handle(http.MethodGet, "/ping", pingHandler.Ping())

	userHandler  := handler.UserHandler{UserService: services.NewUserServiceClass(domain.NewUserRepoClass(dbClient, mongoClient, config.RdmsDB.Schema, config.MongoDB.Database, config.MongoDB.UserCollection), keyCloakMiddleware)}
	router.Handle(http.MethodPost, "/sign-up", userHandler.SignUp())
	router.Handle(http.MethodGet,  "/sign-in", userHandler.SignIn())
	router.Handle(http.MethodGet, "/sso-sign-in", userHandler.SSOLogIn())
	router.Handle(http.MethodGet, "/get-user", userHandler.GetUserById())
	router.Handle(http.MethodPost, "/reset-password", userHandler.ResetPassword(keyCloakMiddleware))
	/*
		1. Register The Router a Method router.GET With Our Request Handler Function.
		2. In the handler function, we return the message back to client.
		3. Run The Router Using router.Run()
	*/
	router.GET("/", func(context *gin.Context) {
		context.JSON(http.StatusOK, gin.H{
			"message": "Starting Server",
		})
	})
	err = router.Run(fmt.Sprintf(":%s", port))
	if err != nil {
		//RequestContextGinLogger
		return
	}

}