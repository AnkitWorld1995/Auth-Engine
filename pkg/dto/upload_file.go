package dto

import "mime/multipart"

type UploadFileResp struct {
	Message string
	Data    map[string]interface{}
}

type MultiUploadFileResp struct {
	Message string
	Data    []struct {
		UniqueID         string `json:"unique_id"`
		FileName         string `json:"file_name"`
		FileSize         int64  `json:"file_size"`
		URL              string `json:"url"`
		Mime             string `json:"mime"`
		Ext              string `json:"ext"`
		DataStreamSHA256 string `json:"data_stream_sha_256"`
	}
}

type UploadFileInput struct {
	File       multipart.File       `json:"file,omitempty" validate:"required"`
	FileHeader multipart.FileHeader `json:"file_header" validate:"required"`
}

type UploadFileListInput struct {
	FileHeader []*multipart.FileHeader `json:"file_header" validate:"required"`
}
