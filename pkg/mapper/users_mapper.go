package mapper

import (
	"context"
	strUtil "github.com/agrison/go-commons-lang/stringUtils"
	"github.com/chsys/userauthenticationengine/pkg/domain"
	"github.com/chsys/userauthenticationengine/pkg/dto"
	errs "github.com/chsys/userauthenticationengine/pkg/lib/error"
	"github.com/chsys/userauthenticationengine/pkg/lib/utility"
	"net/mail"
	"strings"
)

type RequestMapper struct {
	Repo 	domain.UserRepository
}

func NewRequestMapper(repo domain.UserRepository) RequestMapper {
	return RequestMapper{
		Repo: repo,
	}
}

func (r *RequestMapper) ValidatedSignIn(ctx context.Context,req *dto.SignInRequest) *errs.AppError {
	condition := &req.UserName

	if strUtil.IsBlank(strings.TrimSpace(*condition)) {
		return errs.NewValidationError("User Name is empty")
	}else {
		if email, ok := mail.ParseAddress(*condition); ok == nil {
			isExist, err := r.Repo.FindByUserName(ctx,email.Address)
			if err != nil || !isExist {
				return errs.NewValidationError("Email Not Found.")
			}
		}else {
			isExist, err := r.Repo.FindByUserName(ctx,*condition)
			if err != nil || !isExist {
				return errs.NewValidationError("User Name Not Found.")
			}
		}
	}

	if strUtil.IsBlank(strings.TrimSpace(req.Password)) {
		return errs.NewValidationError("Password is empty")
	}else {

		orgPassword, err := r.Repo.GetPassword(ctx, condition)
		if err != nil {
			return err
		}

		isSame := utility.ComparePassword(orgPassword, []byte(req.Password))
		if !isSame {
			return errs.NewValidationError("Password Incorrect")
		}
	}

	return nil
}