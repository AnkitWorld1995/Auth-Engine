package domain

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/private/protocol/rest"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/chsys/userauthenticationengine/pkg/dto"
	"github.com/chsys/userauthenticationengine/pkg/lib/constants"
	errs "github.com/chsys/userauthenticationengine/pkg/lib/error"
	"github.com/chsys/userauthenticationengine/pkg/lib/logger"
	"github.com/chsys/userauthenticationengine/pkg/lib/utility"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"strings"
	"sync"
	"time"
)

type IRepoClass struct {
	db            *sql.DB
	mongo         *mongo.Client
	dbSchema      string
	nosqlDatabase string
	collection    map[string]string
	dynamoDB      *dynamodb.DynamoDB
	AwsSession    *session.Session
}

func IRepoNewClass(rdbClient *sql.DB, mdbClient *mongo.Client, sqlSchema, nosqlDatabase string, collection map[string]string, dynoSess *dynamodb.DynamoDB, session *session.Session) *IRepoClass {
	return &IRepoClass{
		db:            rdbClient,
		mongo:         mdbClient,
		dbSchema:      sqlSchema,
		nosqlDatabase: nosqlDatabase,
		collection:    collection,
		dynamoDB:      dynoSess,
		AwsSession:    session,
	}
}

type IRepository interface {
	FindByUserId(ctx context.Context, id *int) (bool, *errs.AppError)
	FindByEmail(ctx context.Context, email string) (bool, *errs.AppError)
	FindByUserName(ctx context.Context, userName string) (bool, *errs.AppError)
	SaveUser(ctx context.Context, user *Users) (*UserResponse, *errs.AppError)
	GetPassword(ctx context.Context, cond *string) (string, *errs.AppError)
	GetUser(ctx context.Context, userID *int, userName, email *string) (*Users, *errs.AppError)
	UpdatePassword(ctx context.Context, email, password string) (*dto.GenericResponse, *errs.AppError)
	WriteUploadFileDb(s3RespData *dto.UploadFileMetaDataResp) *errs.AppError
	ReadDynamoAllUploadFiles() (*dto.UploadFileMetaDataListResp, *errs.AppError)
	S3Upload(ctx context.Context, buffer *bytes.Buffer, inputData *dto.UploadFileInput) (*dto.UploadFileMetaDataResp, *errs.AppError)
}

func (r *IRepoClass) FindByEmail(ctx context.Context, email string) (bool, *errs.AppError) {
	var emailExist sql.NullBool
	query := fmt.Sprintf(`SELECT 1 FROM %s."users" WHERE "email" = $1`, r.dbSchema)
	err := r.db.QueryRowContext(ctx, query, email).Scan(&emailExist)
	if err != nil {
		return false, errs.NewNotFoundError("Email Not Found")
	}
	return emailExist.Bool, nil
}

func (r *IRepoClass) FindByUserName(ctx context.Context, userName string) (bool, *errs.AppError) {
	var userExist sql.NullBool
	query := fmt.Sprintf(`SELECT 1 FROM %s."users" WHERE "user_name" = $1`, r.dbSchema)
	err := r.db.QueryRowContext(ctx, query, userName).Scan(&userExist)
	if err != nil {
		return false, errs.NewNotFoundError("User Name Not Found")
	}
	return userExist.Bool, nil
}

func (r *IRepoClass) UpdatePassword(ctx context.Context, email, password string) (*dto.GenericResponse, *errs.AppError) {

	sqlQuery := fmt.Sprintf(`UPDATE 	%s.users 
									set
										"password" = ?
									where
										email = ?
									`, r.dbSchema)

	sqlQuery = sqlx.Rebind(sqlx.DOLLAR, sqlQuery)
	rows, err := r.db.ExecContext(ctx, sqlQuery, password, email)
	if err != nil {
		logger.Error(fmt.Sprintf("SQL: Update Password ERROR\t %s", err.Error()))
		return &dto.GenericResponse{
			Success: false,
			Message: fmt.Sprintf("SQL: Update Password ERROR\t %s", err.Error()),
		}, errs.NewUnexpectedError(err.Error())
	}

	affectedRows, _ := rows.RowsAffected()
	if affectedRows < 1 {
		return &dto.GenericResponse{
			Success: false,
			Message: "Password Not Updated In DB.",
		}, errs.NewUnexpectedError("Password Not Updated In DB.")
	} else {
		return &dto.GenericResponse{
			Success: true,
			Message: "Password Updated Successfully In DB.",
		}, nil
	}
}

func (r *IRepoClass) SaveUser(ctx context.Context, user *Users) (*UserResponse, *errs.AppError) {

	inputArgs := make([]interface{}, 0, 10)

	sqlQuery := fmt.Sprintf(`INSERT INTO %s.users
				(user_name, first_name, last_name, "password", email, phone, address, is_admin, user_type, created_at, updated_at)
				VALUES( ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);`, r.dbSchema)
	inputArgs = append(inputArgs, user.UserName, user.FirstName, user.LastName, user.Password, user.Email,
		user.Phone, user.Address, user.IsAdmin, user.UserType, user.CreatedAt, user.UpdatedAt)

	tx, err := r.db.Begin()
	if err != nil {
		return nil, errs.NewUnexpectedError(err.Error())
	}

	sqlQuery = sqlx.Rebind(sqlx.DOLLAR, sqlQuery)

	txRow, txErr := tx.ExecContext(ctx, sqlQuery, inputArgs...)
	if txErr != nil {
		logger.Error(fmt.Sprintf("txErr Error: User/SaveUser API %s", txErr.Error()))
		_ = tx.Rollback()
		return nil, errs.NewUnexpectedError(txErr.Error())
	}

	if err = tx.Commit(); err != nil {
		logger.Error(fmt.Sprintf("txErr Commit Error: User/SaveUser API %s", txErr.Error()))
		return nil, errs.NewUnexpectedError(err.Error())
	}

	rows, err := txRow.RowsAffected()
	if rows > 0 {
		userResp := UserResponse{
			Success: true,
			Message: "User Inserted Successfully.",
		}
		return &userResp, nil
	} else {
		userResp := UserResponse{
			Success: false,
			Message: "User Insertion Failed.",
		}
		return &userResp, errs.NewUnexpectedError("Rows Unaffected.")
	}

}

func (r *IRepoClass) GetPassword(ctx context.Context, cond *string) (string, *errs.AppError) {
	var password sql.NullString

	sqlQuery := fmt.Sprintf(`SELECT urs."password" FROM %s."users" urs WHERE (urs.user_name = $1 or urs.email = $2)`, r.dbSchema)

	err := r.db.QueryRowContext(ctx, sqlQuery, cond, cond).Scan(&password)
	if err != nil || !password.Valid {
		return "", errs.NewUnexpectedError(err.Error())
	}

	return password.String, nil
}

func (r *IRepoClass) FindByUserId(ctx context.Context, id *int) (bool, *errs.AppError) {
	var userID sql.NullBool

	sqlQuery := fmt.Sprintf(`SELECT 1 FROM %s."users" urs WHERE urs.user_id = $1;`, r.dbSchema)

	err := r.db.QueryRowContext(ctx, sqlQuery, id).Scan(&userID)
	if err != nil || !userID.Valid {
		return false, errs.NewNotFoundError(err.Error())
	}

	return userID.Bool, nil
}

func (r *IRepoClass) GetUser(ctx context.Context, userID *int, userName, email *string) (*Users, *errs.AppError) {
	var userResp = Users{}
	sqlQuery := fmt.Sprintf(`select
									id,
									user_id,
									user_name,
									first_name,
									last_name,
									"password",
									email,
									phone,
									address,
									is_admin,
									user_type,
									created_at,
									updated_at
								from
									%s.users urs
								where
									urs."user_id" = ? or (urs."user_name" = ? or urs."email" = ?);`, r.dbSchema)

	sqlQuery = sqlx.Rebind(sqlx.DOLLAR, sqlQuery)
	err := r.db.QueryRowContext(ctx, sqlQuery, userID, userName, email).Scan(&userResp.ID,
		&userResp.UserID,
		&userResp.UserName,
		&userResp.FirstName,
		&userResp.LastName,
		&userResp.Password,
		&userResp.Email,
		&userResp.Phone,
		&userResp.Address,
		&userResp.IsAdmin,
		&userResp.UserType,
		&userResp.CreatedAt,
		&userResp.UpdatedAt)
	if err != nil {
		defer func(db *sql.DB) {
			err := db.Close()
			if err != nil {
				logger.Error(fmt.Sprintf("GetUserById: Defer Sql Func Error: %s", err.Error()))
				return
			}
		}(r.db)
		return nil, errs.NewUnexpectedError(err.Error())
	}
	return &userResp, nil
}

func (r *IRepoClass) S3Upload(ctx context.Context, buffer *bytes.Buffer, inputData *dto.UploadFileInput) (*dto.UploadFileMetaDataResp, *errs.AppError) {
	// Config settings: this is where you choose the bucket, filename, content-type etc.
	// of the file you're uploading.
	_, err := s3.New(r.AwsSession).PutObject(&s3.PutObjectInput{
		Bucket:               aws.String(utility.ReadS3Bucket()),
		Key:                  aws.String(inputData.FileHeader.Filename),
		ACL:                  aws.String("private"),
		Body:                 bytes.NewReader(buffer.Bytes()),
		ContentLength:        aws.Int64(inputData.FileHeader.Size),
		ContentType:          aws.String(http.DetectContentType(buffer.Bytes())),
		ContentDisposition:   aws.String("attachment"),
		ServerSideEncryption: aws.String("AES256"),
	})
	if err != nil {
		logger.Error("Service/Upload/", zap.String("S3 Upload: ERROR", err.Error()))
		return nil, errs.NewValidationError(err.Error())
	}

	req, _ := s3.New(r.AwsSession).GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(utility.ReadS3Bucket()),
		Key:    aws.String(inputData.FileHeader.Filename),
	})
	rest.Build(req)
	uploadedResourceLocation := req.HTTPRequest.URL.String()

	s3UploadResponse := UploadFileMetaData{
		UniqueID:         uuid.New().String(),
		FileName:         inputData.FileHeader.Filename,
		FileSize:         inputData.FileHeader.Size,
		URL:              uploadedResourceLocation,
		DataStreamSHA256: utility.SingleSHA(buffer.Bytes()),
		Ext:              strings.Split(inputData.FileHeader.Filename, ".")[1],
		Mime:             http.DetectContentType(buffer.Bytes()),
	}

	return s3UploadResponse.DTO(), nil
}

func (r *IRepoClass) S3MultiUpload(ctx context.Context, inputData *dto.UploadFileListInput) ([]*UploadFileMetaData, *errs.AppError){
	// Convert the File Stored In String into []Byte.

	start := time.Now()
	var fileBufferLst []*MultiPartFileData

	// Concurrent Code.
	var wg sync.WaitGroup
	var mux sync.RWMutex

	for _, fileHeader := range inputData.FileHeader {
		var fileData MultiPartFileData
		wg.Add(1)
		go func(fileData *MultiPartFileData, mux *sync.RWMutex, fileHeader *multipart.FileHeader) {
			defer wg.Done()
			fileOut, err := fileHeader.Open()
			if err != nil {
				//errChan <- err
				return
			}

			fileBuffer, err := io.ReadAll(fileOut)
			if err != nil {
				//errChan <- err
				return
			}

			mux.Lock()
			fileData.FileName = &fileHeader.Filename
			fileData.FileSize = &fileHeader.Size
			fileData.FileBufferByte = &fileBuffer
			fileBufferLst = append(fileBufferLst, fileData)
			mux.Unlock()

		}(&fileData, &mux, fileHeader)
	}
	wg.Wait()

	resp, err := multiUploadRepo(r.AwsSession, fileBufferLst, start)
	if err != nil {
		logger.Error("Service/S3-Multi-Upload/", zap.String("S3 Multi-Upload: ERROR", err.Message))
		return nil, errs.NewValidationError(err.Message)
	}
	return resp, nil
}

func (r *IRepoClass) WriteUploadFileDb(s3RespData *dto.UploadFileMetaDataResp) *errs.AppError {

	mappedData, err := dynamodbattribute.MarshalMap(s3RespData)
	if err != nil {
		return errs.NewUnexpectedError(err.Error())
	} else {
		input := &dynamodb.PutItemInput{
			Item:      mappedData,
			TableName: aws.String(constants.DynamoDBS3UploadTable),
		}

		_, err = r.dynamoDB.PutItem(input)
		if err != nil {
			logger.Error(fmt.Sprintf("Got error calling PutItem: %s", err.Error()))
			return errs.NewUnexpectedError(err.Error())
		} else {
			return nil
		}
	}
}

func (r *IRepoClass) ReadDynamoAllUploadFiles() (*dto.UploadFileMetaDataListResp, *errs.AppError) {

	filterExp := expression.Name("UniqueID").AttributeType(expression.String)
	projEx := expression.NamesList(expression.Name("DataStreamSHA256"), expression.Name("Ext"), expression.Name("Mime"), expression.Name("URL"),
		expression.Name("FileSize"), expression.Name("FileName"), expression.Name("UniqueID"))

	expr, err := expression.NewBuilder().WithFilter(filterExp).WithProjection(projEx).Build()
	if err != nil {
		logger.Error(fmt.Sprintf("REPO/ReadAllUploadFile: Failed to build expression,  %s", err.Error()))
		return nil, errs.NewUnexpectedError(err.Error())
	}

	resp, err := r.dynamoDB.Scan(&dynamodb.ScanInput{
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		FilterExpression:          expr.Filter(),
		ProjectionExpression:      expr.Projection(),
		TableName:                 aws.String(constants.DynamoDBS3UploadTable),
	})

	if err != nil {
		logger.Error(fmt.Sprintf("REPO/ReadAllUploadFile: Failed to Scan Query,  %s", err.Error()))
		return nil, errs.NewUnexpectedError(err.Error())
	}

	var uploadFileResp []UploadFileMetaResp
	err = dynamodbattribute.UnmarshalListOfMaps(resp.Items, &uploadFileResp)
	if err != nil {
		logger.Error(fmt.Sprintf("REPO/ReadAllUploadFile: Failed to Unmarshell Object,  %s", err.Error()))
		return nil, errs.NewUnexpectedError(err.Error())
	}

	uploadFileList := &UploadFileMetaList{
		MetaUnmarshalList: &uploadFileResp,
	}

	return uploadFileList.DTO(), nil
}

func multiUploadRepo(sess *session.Session, fileDataList []*MultiPartFileData, start time.Time) ([]*UploadFileMetaData, *errs.AppError) {
	// Config settings: this is where you choose the bucket, filename, content-type etc.
	// of the file you're uploading.

	//Concurrent Code
	var wg sync.WaitGroup
	var mux sync.RWMutex
	mapFileNameURL := make(map[string]string)
	uploadSuccess := make(chan bool)
	errChan := make(chan error, 2)

	for _, fileRawValue := range fileDataList {

		wg.Add(1)
		go func(fileRawValue *MultiPartFileData) {
			defer wg.Done()
			_, err := s3.New(sess).PutObject(&s3.PutObjectInput{
				Bucket:               aws.String(utility.ReadS3Bucket()),
				Key:                  aws.String(*fileRawValue.FileName),
				ACL:                  aws.String("private"),
				Body:                 bytes.NewReader(*fileRawValue.FileBufferByte),
				ContentLength:        aws.Int64(*fileRawValue.FileSize),
				ContentType:          aws.String(http.DetectContentType(*fileRawValue.FileBufferByte)),
				ContentDisposition:   aws.String("attachment"),
				ServerSideEncryption: aws.String("AES256"),
			})
			if err != nil {
				logger.Error("Service/Multi-Upload/", zap.String("Multi-Upload: ERROR", err.Error()))
				errChan <- err
				return
			}
			log.Println("Exiting The 1st Go")
			uploadSuccess <- true

		}(fileRawValue)

		go func(mux *sync.RWMutex, fileRawValue *MultiPartFileData) {
			if <-uploadSuccess {
				log.Println("Entering The 2st Go")
				req, _ := s3.New(sess).GetObjectRequest(&s3.GetObjectInput{
					Bucket: aws.String(utility.ReadS3Bucket()),
					Key:    aws.String(*fileRawValue.FileName),
				})
				rest.Build(req)
				mux.Lock()
				mapFileNameURL[*fileRawValue.FileName] = req.HTTPRequest.URL.String()
				fmt.Println("\nAfter here == >", req.HTTPRequest.URL.String())
				mux.Unlock()
			}
		}(&mux, fileRawValue)

	}

	go func() {
		wg.Wait()
		close(errChan)
		close(uploadSuccess)
	}()

	err := utility.JoinChannelError(errChan)
	if err != nil {
		newError := errs.AppError{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		}
		return nil, &newError
	}

	s3RespObject := make([]*UploadFileMetaData, len(fileDataList))
	for _, fileDataResp := range fileDataList {
		wg.Add(1)
		go func(fileDataResp *MultiPartFileData, mux *sync.RWMutex) {
			defer wg.Done()
			resp := &UploadFileMetaData{
				UniqueID:         uuid.New().String(),
				FileName:         *fileDataResp.FileName,
				FileSize:         *fileDataResp.FileSize,
				URL:              mapFileNameURL[*fileDataResp.FileName],
				Mime:             http.DetectContentType(*fileDataResp.FileBufferByte),
				Ext:              strings.Split(*fileDataResp.FileName, ".")[1],
				DataStreamSHA256: utility.SingleSHA(*fileDataResp.FileBufferByte),
			}
			mux.Lock()
			s3RespObject = append(s3RespObject, resp)
			mux.Unlock()
		}(fileDataResp, &mux)
	}
	wg.Wait()
	log.Println("\n\n Time Ellipsed S3 Upload at ==> ", time.Since(start))
	return s3RespObject, nil
}