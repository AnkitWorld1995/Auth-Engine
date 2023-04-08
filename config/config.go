package config

import (
	"fmt"
	"github.com/chsys/userauthenticationengine/pkg/client/db"
	"github.com/chsys/userauthenticationengine/pkg/lib/constants"
	"github.com/chsys/userauthenticationengine/pkg/lib/utility"
	"log"
)

/*
	Configuration Package.
*/


type AppConfig struct {
	RdmsDB	 *db.PostgresConfig
	MongoDB  *db.MongoConfig
}

var applicationConfig *AppConfig

func Init()  {
	userCollection := make(map[string]string)
	userCollection[constants.MongoCollectionName] = utility.ReadNSQLCollection()
	appConfig := &AppConfig{
		RdmsDB:  &db.PostgresConfig{
			Host:     utility.ReadRDBHost(),
			Port:     utility.ReadRDBPort(),
			Username: utility.ReadRDBUser(),
			Password: utility.ReadRDBPassword(),
			Database: utility.ReadRDB(),
		},
		MongoDB: &db.MongoConfig{
			Host:            utility.ReadNSQLHost(),
			Port:            utility.ReadNSQLPort(),
			Username:        "",
			Password:        "",
			MaxPool:         "",
			Database:        utility.ReadNSQLDatabase(),
			UserCollection:  userCollection,
			AdminCollection: nil,
		},
	}
	applicationConfig = appConfig
	fmt.Printf("The value Of App Config :== %v+\n", applicationConfig.RdmsDB)
}

func AppConfigs() *AppConfig{
	if applicationConfig == nil {
		fmt.Printf("The Application Configured is Empty. Value is %v", applicationConfig)
		log.Fatalln("Failed to Initialize Config Variables.")
	}
	log.Println("Config",applicationConfig.RdmsDB)
	return applicationConfig
}