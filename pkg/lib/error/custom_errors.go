package errs

import "net/http"

type AppError struct {
	Code    int    `json:",omitempty"`
	Message string `json:"message"`
}

/*
	Note: AppErrorOption & ErrMessage Follows Option Design Pattern.
	Need To Implement In Later Stages Of Development.
*/

type AppErrorOptionList struct {
	ErrList []AppError
}

type AppErrorOption func(AppError) *AppErrorOptionList

func NewErrMessageList(err AppError) AppErrorOption {
	return func(appError AppError) *AppErrorOptionList {
		errLst := new(AppErrorOptionList)
		errLst.AddErr(err)
		return &AppErrorOptionList{
			ErrList: errLst.ErrList,
		}
	}
}


func(e *AppErrorOptionList) AddErr(err AppError) {
	e.ErrList = append(e.ErrList, err)
}

func (e AppError) AsMessage() *AppError {
	return &AppError{Message: e.Message}
}

func NewNotFoundError(message string) *AppError {
	return &AppError{
		Message: message,
		Code:    http.StatusNotFound,
	}
}

func NewForbiddenRequest(message string) *AppError{
	return &AppError{
		Code:    http.StatusForbidden,
		Message: message,
	}
}

func NewUnexpectedError(message string) *AppError {
	return &AppError{
		Message: message,
		Code:    http.StatusInternalServerError,
	}
}

func NewValidationError(message string) *AppError {
	return &AppError{
		Message: message,
		Code:    http.StatusUnprocessableEntity,
	}
}

func NewUnauthorizedError(message string) *AppError {
	return &AppError{
		Message: message,
		Code:    http.StatusUnauthorized,
	}
}

func NewNoContentError(message string) *AppError {
	return &AppError{
		Message: message,
		Code:    http.StatusNoContent,
	}
}
