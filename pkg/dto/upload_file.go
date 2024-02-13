package dto

import (
	"mime/multipart"
)

type UploadFileResp struct {
	Message string
	Data    map[string]interface{}
}

type MultiUploadFileReq struct {
	UniqueID         string `json:"UniqueID"`
	FileName         string `json:"FileName"`
	FileSize         int64  `json:"FileSize"`
	URL              string `json:"URL"`
	Mime             string `json:"Mime"`
	Ext              string `json:"Ext"`
	DataStreamSHA256 string `json:"DataStreamSHA256"`
}

type UploadFileInput struct {
	File       multipart.File       `json:"file,omitempty" validate:"required"`
	FileHeader multipart.FileHeader `json:"file_header" validate:"required"`
}

type UploadFileListInput struct {
	FileHeader []*multipart.FileHeader `json:"file_header" validate:"required"`
}

type UploadFileMetaDataResp struct {
	UniqueID         string
	FileName         string
	FileSize         int64
	URL              string
	Mime             string
	Ext              string
	DataStreamSHA256 string
}

type UploadFileMetaDataListResp struct {
	FileMetaList []*UploadFileMetaDataResp
}
