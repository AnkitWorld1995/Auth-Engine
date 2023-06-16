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
	"io"
	"net/http"
	"os"
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

type MultiPartFileData struct {
	FileName 			*string   `json:"file_name"`
	FileSize 			*int64	  `json:"file_size"`
	FileBufferByte 		*[]byte   `json:"file_buffer_byte"`
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

func S3MultiUpload(sess *session.Session, inputData *dto.UploadFileListInput) ([]*UploadFileMetaData,*errs.AppError) {
	// Convert the File Stored In String into []Byte.
	var fileBufferLst []*MultiPartFileData
	//errChan := make(chan error, 2)
	for i:= 0; i < len(inputData.FileHeader); i++ {

		var fileData MultiPartFileData
		fileOut, err := os.Create(inputData.FileHeader[i].Filename)
		if err != nil {
			//errChan <- err
			return nil, errs.NewUnexpectedError(err.Error())
		}

		fileBuffer, err := io.ReadAll(fileOut)
		if err != nil {
			//errChan <- err
			return nil, errs.NewUnexpectedError(err.Error())
		}

		fileData.FileName = &inputData.FileHeader[i].Filename
		fileData.FileSize = &inputData.FileHeader[i].Size
		fileData.FileBufferByte = &fileBuffer

		fileBufferLst = append(fileBufferLst, &fileData)
	}
	resp, err := multiUpload(sess,fileBufferLst)
	if err != nil {
		logger.Error("Service/Multi-Upload/", zap.String("S3 Multi-Upload: ERROR", err.Message))
		return nil, errs.NewValidationError(err.Message)
	}
	return resp, nil
}

func multiUpload(sess *session.Session, fileDataList []*MultiPartFileData) ([]*UploadFileMetaData,*errs.AppError) {
	// Config settings: this is where you choose the bucket, filename, content-type etc.
	// of the file you're uploading.

	mapFileNameURL := make(map[string]string)
	for j:= 0; j < len(fileDataList); j++ {
		_ , err := s3.New(sess).PutObject(&s3.PutObjectInput{
			Bucket:               aws.String(utility.ReadS3Bucket()),
			Key:                  aws.String(*fileDataList[j].FileName),
			ACL:                  aws.String("private"),
			Body:                 bytes.NewReader(*fileDataList[j].FileBufferByte),
			ContentLength:        aws.Int64(*fileDataList[j].FileSize),
			ContentType:          aws.String(http.DetectContentType(*fileDataList[j].FileBufferByte)),
			ContentDisposition:   aws.String("attachment"),
			ServerSideEncryption: aws.String("AES256"),
		})
		if err != nil {
			logger.Error("Service/Multi-Upload/", zap.String("S3 Multi-Upload: ERROR", err.Error()))
			return nil, errs.NewValidationError(err.Error())
		}

		req, _ := s3.New(sess).GetObjectRequest(&s3.GetObjectInput{
			Bucket: aws.String(utility.ReadS3Bucket()),
			Key:    aws.String(*fileDataList[j].FileName),
		})
		rest.Build(req)
		mapFileNameURL[*fileDataList[j].FileName] = req.HTTPRequest.URL.String()
	}

	s3RespObject := make([]*UploadFileMetaData, len(fileDataList))
	for k:=0; k < len(fileDataList); k++ {

		resp := &UploadFileMetaData{
			UniqueID:         uuid.New().String(),
			FileName:         *fileDataList[k].FileName,
			FileSize:         *fileDataList[k].FileSize,
			URL:              mapFileNameURL[*fileDataList[k].FileName],
			Mime:             http.DetectContentType(*fileDataList[k].FileBufferByte),
			Ext:              strings.Split(*fileDataList[k].FileName, ".")[1],
			DataStreamSHA256: utility.SingleSHA(*fileDataList[k].FileBufferByte),
		}
		s3RespObject = append(s3RespObject, resp)
	}
	return s3RespObject, nil
}