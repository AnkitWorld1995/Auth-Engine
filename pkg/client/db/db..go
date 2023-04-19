package db

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"log"
)

const (
	postgresDriver = "postgres"
	mongoDriver	   = "mongodb"
)

type PostgresConfig struct {
	Host 		string
	Port		string
	Username 	string
	Password	string
	Schema		string
	Database	string
}

type MongoConfig struct {
	Host                   string
	Port                   string
	Username               string
	Password               string
	MaxPool                string
	Database               string
	UserCollection         map[string]string
	AdminCollection        map[string]string
}

func RdmsInit(config *PostgresConfig) (*sql.DB, error) {

	//"postgres://postgres:password@localhost/DB_1?sslmode=disable"
	//jdbc:postgresql://localhost:5432/postgres
	log.Println("===>", config.Host, config.Port)
	dbUrl := fmt.Sprintf("host=%s port=%s user=%s " + "password=%s dbname=%s sslmode=disable",
		     config.Host, config.Port, config.Username, config.Password, config.Database)

	db, err := sql.Open(postgresDriver, dbUrl)
	if err != nil {
		return nil, err
	}

	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			return
		}
	}(db)

	err = db.Ping()
	if err != nil {
		return nil, err
	}
	// Log The Connection

	return db, nil
}

func MongoInit(dbConfig *MongoConfig) (*mongo.Client, error) {
	//mongodb://localhost:27017/
	dataSource := fmt.Sprintf("%s://%s:%s/",
		mongoDriver,  dbConfig.Host, dbConfig.Port)

	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(dataSource))
	if err != nil {
		return nil, err
	}

	// verifies connection is db is working
	if err := client.Ping(context.TODO(), readpref.Primary()); err != nil {
		return nil, err
	}
	return client, nil
}