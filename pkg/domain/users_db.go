package domain

import (
	"context"
	"database/sql"
	"fmt"
	errs "github.com/chsys/userauthenticationengine/pkg/lib/error"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserRepoClass struct {
	db 				*sql.DB
	mongo         	*mongo.Client
	dbSchema      	string
	nosqlDatabase 	string
	collection 		map[string]string
}

func NewUserRepoClass(rdbClient *sql.DB, mdbClient *mongo.Client, sqlSchema, nosqlDatabase string,  collection 	map[string]string ) *UserRepoClass {
	return &UserRepoClass{
		db:            rdbClient,
		mongo:         mdbClient,
		dbSchema:      sqlSchema,
		nosqlDatabase: nosqlDatabase,
		collection:    collection,
	}
}



type UserRepository interface {
	FindByEmail(ctx context.Context, email string) (bool,*errs.AppError)
	FindByUserName(ctx context.Context, userName string) (bool, *errs.AppError)
	SaveUser(ctx context.Context, user *Users) (*UserResponse, *errs.AppError)
}



func (r *UserRepoClass) FindByEmail(ctx context.Context, email string) (bool,*errs.AppError){
	var emailExist sql.NullBool
	query := fmt.Sprintf(`SELECT 1 FROM %s."users" WHERE %s."email" = $1`, r.dbSchema, r.dbSchema)
	err := r.db.QueryRowContext(ctx, query, email).Scan(&emailExist)
	if err != nil {
		return false, errs.NewUnexpectedError(err.Error())
	}
	return emailExist.Bool, nil
}

func (r *UserRepoClass) FindByUserName(ctx context.Context, userName string) (bool, *errs.AppError)  {
	var userExist sql.NullBool
	query := fmt.Sprintf(`SELECT 1 FROM %s."users" WHERE %s."email" = $1`, r.dbSchema, r.dbSchema)
	err := r.db.QueryRowContext(ctx, query, userName).Scan(&userExist)
	if err != nil {
		return false, errs.NewUnexpectedError(err.Error())
	}
	return userExist.Bool, nil
}

func (r *UserRepoClass) SaveUser(ctx context.Context, user *Users) (*UserResponse, *errs.AppError) {
	return nil, nil
}
