package config

import (
	"fmt"
	"github.com/Nerzal/gocloak/v7"
	"github.com/aws/aws-sdk-go/aws"
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
	RdmsDB    *db.PostgresConfig
	MongoDB   *db.MongoConfig
	KeyCloak  *sso.KeyCloak
	AwsConfig *aws.Config
}

var applicationConfig *AppConfig

func Init() {
	var maxGlobalRetry = 5
	var s3Region = utility.ReadAwsRegion()
	//var dynamoDBURL = utility.ReadDynamoDBURL()
	userCollection := make(map[string]string)
	userCollection[constants.MongoCollectionName] = utility.ReadNSQLCollection()
	appConfig := &AppConfig{
		RdmsDB: &db.PostgresConfig{
			Host:     utility.ReadRDBHost(),
			Port:     utility.ReadRDBPort(),
			Username: utility.ReadRDBUser(),
			Password: utility.ReadRDBPassword(),
			Schema:   utility.ReadRDBSchema(),
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
		KeyCloak: &sso.KeyCloak{
			GoCloak:      gocloak.NewClient("http://localhost:8086"),
			ClientId:     "auth-svc",
			ClientSecret: "SACqjRGHCnNs3po9V4BcwKqLj4hDVmZg",
			Realm:        "Authentication-SVC",
		},
		AwsConfig: &aws.Config{
			Region:     &s3Region,
			MaxRetries: &maxGlobalRetry,
			//Endpoint:   &dynamoDBURL,
		},
	}
	applicationConfig = appConfig
	fmt.Printf("The value Of App Config :== %v+\n", applicationConfig.AwsConfig)
}

func AppConfigs() *AppConfig {
	if applicationConfig == nil {
		fmt.Printf("The Application Configured is Empty. Value is %v", applicationConfig)
		Init()
	}
	log.Println("Config", applicationConfig.RdmsDB)
	return applicationConfig
}
