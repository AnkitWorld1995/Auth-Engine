package handler

import (
	"github.com/chsys/userauthenticationengine/pkg/dto"
	"github.com/chsys/userauthenticationengine/pkg/services"
	"github.com/gin-gonic/gin"
	"net/http"
)

type UploadHandler struct {
	UploadFileService services.UploadFileServices
}

func (u *UploadHandler) UploadFileToS3() gin.HandlerFunc{
	return func(ctx *gin.Context) {

		// FormFile returns the first formFile for the provided form key
		formFile, fileHeader, err := ctx.Request.FormFile("file")
		if err != nil {
			ctx.JSON(http.StatusNotFound,
				dto.UploadFileResp{
					Message:  "Sorry!No File Found.",
					Data: map[string]interface{}{"Data" : "No Data Found. Please insert a File To Upload."},
				},
			)
			return
		}else {
			req := dto.UploadFileInput{
				File:       formFile,
				FileHeader: *fileHeader,
			}
			resp, err := u.UploadFileService.Upload(ctx, &req)
			if err != nil {
				ctx.JSON(err.Code,
					dto.UploadFileResp{
						Message: err.Message,
						Data: map[string]interface{}{"Data" : "Nil"},
					},
				)
				return
			}else {
				ctx.JSON(http.StatusOK,
					dto.UploadFileResp{
						Message: resp.Message,
						Data: resp.Data,
					},
				)
				return
			}
		}
	}
}


func (u *UploadHandler) UploadAllFileToS3() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		multiPartForm, _:= ctx.MultipartForm()
		fileHeader := multiPartForm.File["file"]

		if len(fileHeader) > 0 {

			fileData := new(dto.UploadFileListInput)
			for _, fileValue := range fileHeader {
				fileData.FileHeader = append(fileData.FileHeader, fileValue)
			}

			// Service Method.
			u.UploadFileService.UploadAll(ctx, fileData)

		}else {
			ctx.JSON(http.StatusNotFound,
				dto.UploadFileResp{
					Message:  "Sorry!No File Found.",
					Data: map[string]interface{}{"Data" : "No Data Found. Please insert a File To Upload."},
				},
			)
			return
		}
	}
}