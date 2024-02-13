package domain

import (
	"bytes"
	"fmt"
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
	"log"
	"mime/multipart"
	"net/http"
	"strings"
	"sync"
	"time"
)

type UploadFileDBResponse struct {
	UploadedID string `json:"uploaded_id"`
	URL        string `json:"url"`
}

type MultiPartFileData struct {
	FileName       *string `json:"file_name"`
	FileSize       *int64  `json:"file_size"`
	FileBufferByte *[]byte `json:"file_buffer_byte"`
}

type UploadFileMetaData struct {
	UniqueID         string `json:"UniqueID"`
	FileName         string `json:"FileName"`
	FileSize         int64  `json:"FileSize"`
	URL              string `json:"URL"`
	Mime             string `json:"Mime"`
	Ext              string `json:"Ext"`
	DataStreamSHA256 string `json:"DataStreamSHA256"`
}

type UploadFileMetaResp struct {
	DataStreamSHA256 string
	Ext              string
	FileName         string
	FileSize         int64
	Mime             string
	URL              string
	UniqueID         string
}

func (r *UploadFileMetaData) DTO() *dto.UploadFileMetaDataResp {
	return &dto.UploadFileMetaDataResp{
		UniqueID:         r.UniqueID,
		FileName:         r.FileName,
		FileSize:         r.FileSize,
		URL:              r.URL,
		Mime:             r.Mime,
		Ext:              r.Ext,
		DataStreamSHA256: r.DataStreamSHA256,
	}
}

type UploadFileMetaList struct {
	MetaUnmarshalList *[]UploadFileMetaResp
}

func (r *UploadFileMetaList) DTO() *dto.UploadFileMetaDataListResp {

	if r.MetaUnmarshalList == nil {
		return &dto.UploadFileMetaDataListResp{}
	} else {

		metaFileList := make([]*dto.UploadFileMetaDataResp, 0, len(*r.MetaUnmarshalList))
		for _, metaFile := range *r.MetaUnmarshalList {
			uploadMetaFile := dto.UploadFileMetaDataResp{
				UniqueID:         metaFile.UniqueID,
				FileName:         metaFile.FileName,
				FileSize:         metaFile.FileSize,
				URL:              metaFile.URL,
				Mime:             metaFile.Mime,
				Ext:              metaFile.Ext,
				DataStreamSHA256: metaFile.DataStreamSHA256,
			}
			metaFileList = append(metaFileList, &uploadMetaFile)
		}
		return &dto.UploadFileMetaDataListResp{FileMetaList: metaFileList}
	}
}

func S3MultiUpload(sess *session.Session, inputData *dto.UploadFileListInput) ([]*UploadFileMetaData, *errs.AppError) {
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

	resp, err := multiUpload(sess, fileBufferLst, start)
	if err != nil {
		logger.Error("Service/S3-Multi-Upload/", zap.String("S3 Multi-Upload: ERROR", err.Message))
		return nil, errs.NewValidationError(err.Message)
	}
	return resp, nil
}

func multiUpload(sess *session.Session, fileDataList []*MultiPartFileData, start time.Time) ([]*UploadFileMetaData, *errs.AppError) {
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
