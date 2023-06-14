package services

import (
	"bytes"
	"context"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/chsys/userauthenticationengine/pkg/dto"
	errs "github.com/chsys/userauthenticationengine/pkg/lib/error"
	"github.com/chsys/userauthenticationengine/pkg/lib/logger"
	"github.com/chsys/userauthenticationengine/pkg/lib/utility"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
	"io"
	"log"
	"net/http"
)

type uploadFileServiceClass struct {
	s3Session  *session.Session
}


func NewUploadFileService(s3Session *session.Session) *uploadFileServiceClass {
	return &uploadFileServiceClass{
		s3Session: s3Session,
	}
}


type UploadFileServices interface {
	Upload(ctx context.Context, inputData *dto.UploadFileInput) (*dto.UploadFileResp,*errs.AppError)
}

func (u *uploadFileServiceClass) Upload(ctx context.Context, inputData *dto.UploadFileInput) (*dto.UploadFileResp,*errs.AppError){

	validate := validator.New()
	err := validate.Struct(inputData)
	if err != nil {
		logger.Error("Service/Upload/", zap.String("Validate: ERROR", err.Error()))
		return nil, errs.NewValidationError(err.Error())
	}

	// S3 Upload.
	log.Println( inputData.FileHeader.Filename, inputData.FileHeader.Size)
	buffer := bytes.NewBuffer(nil)
	_, err = io.Copy(buffer, inputData.File)
	if err != nil {
		logger.Error("Service/Upload/", zap.String("Buffer: ERROR", err.Error()))
		return nil, errs.NewUnexpectedError(err.Error())
	}

	// Config settings: this is where you choose the bucket, filename, content-type etc.
	// of the file you're uploading.
	s3Resp , err := s3.New(u.s3Session).PutObject(&s3.PutObjectInput{
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

	return &dto.UploadFileResp{
		HttpCode: http.StatusOK,
		Message: "Uploaded File SuccessFully.",
		Data: map[string]interface{}{
			"Id": 	s3Resp.SSEKMSKeyId,
			"Data": s3Resp.GoString(),
		},
	}, nil
}