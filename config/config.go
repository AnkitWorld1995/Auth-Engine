package config

import (
	"fmt"
	"github.com/Nerzal/gocloak/v7"
	"github.com/chsys/userauthenticationengine/pkg/client/db"
	"github.com/chsys/userauthenticationengine/pkg/client/sso"
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
	KeyCloak *sso.KeyCloak
}

var applicationConfig *AppConfig

func Init()  {
	userCollection := make(map[string]string)
	userCollection[constants.MongoCollectionName] = utility.ReadNSQLCollection()
	appConfig := &AppConfig{
		RdmsDB:  &db.PostgresConfig{
			Host:     	utility.ReadRDBHost(),
			Port:     	utility.ReadRDBPort(),
			Username: 	utility.ReadRDBUser(),
			Password: 	utility.ReadRDBPassword(),
			Schema: 	utility.ReadRDBSchema(),
			Database: 	utility.ReadRDB(),
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
		KeyCloak: &sso.KeyCloak{
			GoCloak:      gocloak.NewClient("http://localhost:8086"),
			ClientId:     "auth-svc",
			ClientSecret: "SACqjRGHCnNs3po9V4BcwKqLj4hDVmZg",
			Realm:        "Authentication-SVC",
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