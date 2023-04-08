package utility

import "os"

func ReadPort() string {
	port := os.Getenv("PORT")
	return port
}

func ReadRDB() string {
	db := os.Getenv("DATABASE")
	return db
}

func ReadRDBHost() string {
	dbHost := os.Getenv("DB_HOST")
	return dbHost
}

func ReadRDBPort() string {
	dbPort := os.Getenv("DB_PORT")
	return dbPort
}

func ReadRDBUser() string {
	dbUser := os.Getenv("DB_USERNAME")
	return dbUser
}

func ReadRDBPassword() string {
	dbPassword := os.Getenv("DB_PASSWORD")
	return dbPassword
}

func ReadNSQLHost() string {
	NoSqlDbHost := os.Getenv("NOSQL_HOST")
	return NoSqlDbHost
}

func ReadNSQLPort() string {
	NoSqlDbPort := os.Getenv("NOSQL_PORT")
	return NoSqlDbPort
}

func ReadNSQLDatabase() string {
	NoSqlDb := os.Getenv("NOSQL_DATABASE")
	return NoSqlDb
}

func ReadNSQLCollection() string {
	NoSqlDbCollection := os.Getenv("NOSQL_COLLECTION")
	return NoSqlDbCollection
}