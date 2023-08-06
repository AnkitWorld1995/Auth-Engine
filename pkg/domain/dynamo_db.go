package domain

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"log"
)

func (r *UserRepoClass) createDynamoDBTable() {

	dynamoDbTableAttribute := make([]*dynamodb.AttributeDefinition, 2)
	dynamoDbSchema := make([]*dynamodb.KeySchemaElement, 2)

	dynamoDbTableAttribute = []*dynamodb.AttributeDefinition{
		{
			AttributeName: aws.String("file_name"),
			AttributeType: aws.String("S"),
		},
		{
			AttributeName: aws.String("file_data"),
			AttributeType: aws.String("S"),
		},
	}

	dynamoDbSchema = []*dynamodb.KeySchemaElement{
		{
			AttributeName: aws.String("file_name"),
			KeyType:       aws.String("HASH"),
		},
		{
			AttributeName: aws.String("file_data"),
			KeyType:       aws.String("RANGE"),
		},
	}

	dynamoDbProvisionThroughPut := &dynamodb.ProvisionedThroughput{
		ReadCapacityUnits:  aws.Int64(10),
		WriteCapacityUnits: aws.Int64(10),
	}

	tableInput := &dynamodb.CreateTableInput{
		AttributeDefinitions:  dynamoDbTableAttribute,
		KeySchema:             dynamoDbSchema,
		ProvisionedThroughput: dynamoDbProvisionThroughPut,
		TableName:             aws.String("uploaded_file"),
	}
	createNewTable, err := r.dynamoDB.CreateTable(tableInput)
	if err != nil {
		log.Fatalf("DynamoDB create table Error: %s", err.Error())
	}

	log.Println("\n The New Created table (Dynamodb) response =>", createNewTable)
}
