package main

import (
	"github.com/chsys/userauthenticationengine/config"
	"github.com/chsys/userauthenticationengine/pkg/app"
	"github.com/chsys/userauthenticationengine/pkg/lib/logger"
	"github.com/joho/godotenv"
	"log"
)


/*
	Main Function To Initialize the project.
*/
func init(){
	err := godotenv.Load("app.env")
	if err != nil {
		log.Fatalln("Error loading .env file: " + err.Error())
	}
	config.Init()
	logger.LogInit()
}



func main() {
	configuration := config.AppConfigs()
	app.StartApp(configuration)
}
