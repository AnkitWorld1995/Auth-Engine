package domain

import (
	"bytes"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/private/protocol/rest"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/chsys/userauthenticationengine/pkg/dto"
	errs "github.com/chsys/userauthenticationengine/pkg/lib/error"
	"github.com/chsys/userauthenticationengine/pkg/lib/logger"
	"github.com/chsys/userauthenticationengine/pkg/lib/utility"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"net/http"
	"strings"
)

type UploadFileMetaData struct {
	UniqueID 			string 	`json:"unique_id"`
	FileName 			string 	`json:"file_name"`
	FileSize 			int64   `json:"file_size"`
	URL             	string  `json:"url"`
	Mime             	string  `json:"mime"`
	Ext              	string  `json:"ext"`
	DataStreamSHA256 	string  `json:"data_stream_sha_256"`
}

type UploadFileDBResponse struct {
	UploadedID 			string  `json:"uploaded_id"`
	URL 				string  `json:"url"`
}


func S3Upload(sess *session.Session, buffer *bytes.Buffer, inputData *dto.UploadFileInput) (*UploadFileMetaData, *errs.AppError) {
	// Config settings: this is where you choose the bucket, filename, content-type etc.
	// of the file you're uploading.
	_ , err := s3.New(sess).PutObject(&s3.PutObjectInput{
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

	req, _ := s3.New(sess).GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(utility.ReadS3Bucket()),
		Key:    aws.String(inputData.FileHeader.Filename),
	})
	rest.Build(req)
	uploadedResourceLocation := req.HTTPRequest.URL.String()

	resp := UploadFileMetaData{
		UniqueID:         uuid.New().String(),
		FileName:         inputData.FileHeader.Filename,
		FileSize:         inputData.FileHeader.Size,
		URL:              uploadedResourceLocation,
		DataStreamSHA256: utility.SingleSHA(buffer.Bytes()),
		Ext:              strings.Split(inputData.FileHeader.Filename, ".")[1],
		Mime:             http.DetectContentType(buffer.Bytes()),
	}

	return &resp, nil
}