package dto

import "mime/multipart"

type UploadFileResp struct {
	Message 	string
	Data 		map[string]interface{}
}

type UploadFileInput struct {
	File 		multipart.File `json:"file,omitempty" validate:"required"`
	FileHeader 	multipart.FileHeader `json:"file_header" validate:"required"`
}

type UploadFileListInput struct {
	FileHeader 	[]*multipart.FileHeader `json:"file_header" validate:"required"`
}


