package services

import (
	"context"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/chsys/userauthenticationengine/pkg/domain"
	"github.com/chsys/userauthenticationengine/pkg/dto"
	errs "github.com/chsys/userauthenticationengine/pkg/lib/error"
	"github.com/chsys/userauthenticationengine/pkg/lib/logger"
	"github.com/chsys/userauthenticationengine/pkg/lib/utility"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

type uploadFileServiceClass struct {
	repo       domain.UserRepository
	AwsSession *session.Session
}

func NewUploadFileService(session *session.Session,  repo domain.UserRepository) *uploadFileServiceClass {
	return &uploadFileServiceClass{
		AwsSession: session,
		repo:       repo,
	}
}

type UploadFileServices interface {
	Upload(ctx context.Context, inputData *dto.UploadFileInput) (*dto.UploadFileResp,*errs.AppError)
	UploadAll(ctx context.Context, inputData *dto.UploadFileListInput) (*dto.UploadFileResp,*errs.AppError)
}


func (u *uploadFileServiceClass) Upload(ctx context.Context, inputData *dto.UploadFileInput) (*dto.UploadFileResp,*errs.AppError){

	validate := validator.New()
	err := validate.Struct(inputData)
	if err != nil {
		logger.Error("Service/Upload/", zap.String("Validate: ERROR", err.Error()))
		return nil, errs.NewValidationError(err.Error())
	}

	fileBuffer, appErr := utility.CreateFileBuffer(inputData.File)
	if appErr != nil {
		logger.Error("Service/Upload/", zap.String("Buffer: ERROR", appErr.Message))
		return nil, appErr
	}

	s3Response, appErr := domain.S3Upload(u.AwsSession, fileBuffer, inputData)
	if appErr != nil {
		logger.Error("Service/Upload/", zap.String("S3 Upload: ERROR", appErr.Message))
		return nil, appErr
	}

	//response, appErr := u.repo.UploadFilesWriteDB(ctx, 4, s3Response)
	//if appErr != nil {
	//	logger.Error("Service/Upload/", zap.String("Insert Upload File: ERROR", appErr.Message))
	//	return nil, appErr
	//}

	return &dto.UploadFileResp{
		Message: "Uploaded File SuccessFully.",
		Data: map[string]interface{}{
			"DATA": s3Response,
		},
	}, nil
}

func (u *uploadFileServiceClass) UploadAll(ctx context.Context, inputData *dto.UploadFileListInput) (*dto.UploadFileResp,*errs.AppError) {

	validate := validator.New()
	err := validate.Struct(inputData)
	if err != nil {
		logger.Error("Service/UploadAll/", zap.String("Validate: ERROR", err.Error()))
		return nil,  errs.NewUnexpectedError(err.Error())
	}

	response, appErr := domain.S3MultiUpload(u.AwsSession, inputData)
	if appErr != nil {
		logger.Error("Service/UploadAll/", zap.String("S3MultiUpload: ERROR", appErr.Message))
		return nil, appErr
	}

	return &dto.UploadFileResp{
		Message: "Uploaded File SuccessFully.",
		Data: map[string]interface{}{
			"DATA": response,
		},
	}, nil

}