package db

import (
	"errors"
	"fmt"
	"github.com/agrison/go-commons-lang/stringUtils"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/chsys/userauthenticationengine/pkg/lib/constants"
	"github.com/chsys/userauthenticationengine/pkg/lib/logger"
	"github.com/chsys/userauthenticationengine/pkg/lib/utility"
	"log"
	"sort"
	"strings"
)

func GetDynamoDbTableParams(sess *dynamodb.DynamoDB) {
	_, err := sess.DescribeTable(&dynamodb.DescribeTableInput{TableName: aws.String(constants.DynamoDBS3UploadTable)})
	if err != nil {
		log.Fatalf("DynamoDB describe table Error: %s", err.Error())
	}

}

func ValidateDynamoDbTable(sess *dynamodb.DynamoDB) bool {
	fetchedTableListV2 := make([]string, 0)
	//tableName := constants.DynamoDBS3UploadTable
	tableNameList := &dynamodb.ListTablesInput{}

	for {
		list, err := sess.ListTables(tableNameList)
		if err != nil {
			var awsErr awserr.Error
			if errors.As(err, &awsErr) {
				switch awsErr.Code() {
				case dynamodb.ErrCodeInternalServerError:
					fmt.Println(dynamodb.ErrCodeInternalServerError, awsErr.Error())
				default:
					fmt.Println(awsErr.Error())
				}
			} else {
				fmt.Println(err.Error())
			}
			return false
		}

		for _, name := range list.TableNames {
			if name != nil && stringUtils.IsNotBlank(strings.TrimSpace(*name)) {
				fetchedTableListV2 = append(fetchedTableListV2, *name)
			}
		}

		if list.LastEvaluatedTableName == nil {
			logger.Info("No table list in DynamoDB is found.")
			break
		}
	}

	if fetchedTableListV2 != nil && len(fetchedTableListV2) > 0 {
		sort.Strings(fetchedTableListV2)
		isExist := utility.MapTableName(fetchedTableListV2)
		return isExist

	} else {
		return false
	}
}

func CreateDynamoDbTable(sess *dynamodb.DynamoDB) {
	dynamoDbTableAttribute := make([]*dynamodb.AttributeDefinition, 2)
	dynamoDbSchema := make([]*dynamodb.KeySchemaElement, 2)

	dynamoDbTableAttribute = []*dynamodb.AttributeDefinition{
		{
			AttributeName: aws.String("UniqueID"),
			AttributeType: aws.String("S"),
		},
		{
			AttributeName: aws.String("FileName"),
			AttributeType: aws.String("S"),
		},
	}

	dynamoDbSchema = []*dynamodb.KeySchemaElement{
		{
			AttributeName: aws.String("UniqueID"),
			KeyType:       aws.String("HASH"),
		},
		{
			AttributeName: aws.String("FileName"),
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
		TableName:             aws.String(constants.DynamoDBS3UploadTable),
	}
	createNewTable, err := sess.CreateTable(tableInput)
	if err != nil {
		log.Fatalf("DynamoDB create table Error: %s", err.Error())
	}

	log.Println("\n The New Created table (Dynamodb) response =>", createNewTable)
}

func DropDynamoDbTable(sess *dynamodb.DynamoDB) {

	dropTable, err := sess.DeleteTable(&dynamodb.DeleteTableInput{
		TableName: aws.String(constants.DynamoDBS3UploadTable),
	})
	if err != nil {
		log.Fatalf("DynamoDB drop table Error: %s", err.Error())
	}

	log.Println("\n The Existing table (Dynamodb) Drop response =>", dropTable)
}
