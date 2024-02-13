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
	"os"
	"strconv"
	"time"
)

const (
	postgresDriver = "postgres"
	mongoDriver    = "mongodb"
)

type PostgresConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	Schema   string
	Database string
}

type MongoConfig struct {
	Host            string
	Port            string
	Username        string
	Password        string
	MaxPool         string
	Database        string
	UserCollection  map[string]string
	AdminCollection map[string]string
}

func RdmsInit(config *PostgresConfig) (*sql.DB, error) {

	//"postgres://postgres:password@localhost/DB_1?sslmode=disable"
	//jdbc:postgresql://localhost:5432/postgres
	log.Println("===>", config.Host, config.Port)
	dbUrl := fmt.Sprintf("host=%s port=%s user=%s "+"password=%s dbname=%s sslmode=disable",
		config.Host, config.Port, config.Username, config.Password, config.Database)

	maxOpenConnection, err := strconv.Atoi(os.Getenv("POSTGRES_MAX_CONN"))
	if err != nil {
		log.Println(err)
		maxOpenConnection = 5
	}
	maxIdleTime, err := strconv.Atoi(os.Getenv("POSTGRES_MAX_IDLE_TIME"))
	if err != nil {
		log.Println(err)
		maxIdleTime = 5
	}
	maxConnectionLifetime, err := strconv.Atoi(os.Getenv("POSTGRES_MAX_LIFETIME"))
	if err != nil {
		log.Println(err)
		maxConnectionLifetime = 2
	}

	db, err := sql.Open(postgresDriver, dbUrl)
	if err != nil {
		return nil, err
	}

	db.SetMaxIdleConns(maxOpenConnection)
	db.SetConnMaxLifetime(time.Duration(maxConnectionLifetime) * time.Minute)
	db.SetConnMaxIdleTime(time.Duration(maxIdleTime) * time.Minute)

	//defer func(db *sql.DB) {
	//	err := db.Close()
	//	if err != nil {
	//		return
	//	}
	//}(db)

	err = db.Ping()
	if err != nil {
		log.Print(err)
		return nil, err
	}
	// Log The Connection

	return db, nil
}

func MongoInit(dbConfig *MongoConfig) (*mongo.Client, error) {
	//mongodb://localhost:27017/
	dataSource := fmt.Sprintf("%s://%s:%s/",
		mongoDriver, dbConfig.Host, dbConfig.Port)

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
