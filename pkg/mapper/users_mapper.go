package mapper

import (
	"context"
	"github.com/chsys/userauthenticationengine/pkg/domain"
	errs "github.com/chsys/userauthenticationengine/pkg/lib/error"
	"github.com/chsys/userauthenticationengine/pkg/lib/utility"
)

type RequestValidation struct {
	Repo domain.UserRepository
}

type RequestValidationInterface interface {
	ValidatePassword(ctx context.Context,password, condition string) *errs.AppError
	ValidateUserName(ctx context.Context, userName string) (bool, *errs.AppError)
	ValidateEmail(ctx context.Context, email string) (bool, *errs.AppError)
}

func (r *RequestValidation) ValidateUserName(ctx context.Context, userName string) (bool, *errs.AppError) {
	return  r.Repo.FindByUserName(ctx, userName)
}

func (r *RequestValidation) ValidateEmail(ctx context.Context, email string) (bool, *errs.AppError) {
	return  r.Repo.FindByEmail(ctx, utility.ParseMail(email))
}

func (r *RequestValidation) ValidatePassword(ctx context.Context,password, condition string) *errs.AppError  {
	orgPassword, err := r.Repo.GetPassword(ctx, &condition)
	if err != nil {
		return err
	}

	isSame := utility.ComparePassword(orgPassword, []byte(password))
	if !isSame {
		return errs.NewValidationError("Password Incorrect")
	}
	return nil
}