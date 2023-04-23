package domain

import (
	"context"
	"database/sql"
	"fmt"
	errs "github.com/chsys/userauthenticationengine/pkg/lib/error"
	"github.com/chsys/userauthenticationengine/pkg/lib/logger"
	"github.com/jmoiron/sqlx"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserRepoClass struct {
	db            *sql.DB
	mongo         *mongo.Client
	dbSchema      string
	nosqlDatabase string
	collection    map[string]string
}

func NewUserRepoClass(rdbClient *sql.DB, mdbClient *mongo.Client, sqlSchema, nosqlDatabase string, collection map[string]string) *UserRepoClass {
	return &UserRepoClass{
		db:            rdbClient,
		mongo:         mdbClient,
		dbSchema:      sqlSchema,
		nosqlDatabase: nosqlDatabase,
		collection:    collection,
	}
}

type UserRepository interface {
	FindByEmail(ctx context.Context, email string) (bool, *errs.AppError)
	FindByUserName(ctx context.Context, userName string) (bool, *errs.AppError)
	SaveUser(ctx context.Context, user *Users) (*UserResponse, *errs.AppError)
	GetPassword(ctx context.Context, cond *string) (string, *errs.AppError)
	GetUser(ctx context.Context, cond *string) (*Users, *errs.AppError)
}

func (r *UserRepoClass) FindByEmail(ctx context.Context, email string) (bool, *errs.AppError) {
	var emailExist sql.NullBool
	query := fmt.Sprintf(`SELECT 1 FROM %s."users" WHERE "email" = $1`, r.dbSchema)
	err := r.db.QueryRowContext(ctx, query, email).Scan(&emailExist)
	if err != nil {
		return false, errs.NewUnexpectedError(err.Error())
	}
	return emailExist.Bool, nil
}

func (r *UserRepoClass) FindByUserName(ctx context.Context, userName string) (bool, *errs.AppError) {
	var userExist sql.NullBool
	query := fmt.Sprintf(`SELECT 1 FROM %s."users" WHERE "user_name" = $1`, r.dbSchema)
	err := r.db.QueryRowContext(ctx, query, userName).Scan(&userExist)
	if err != nil {
		return false, errs.NewUnexpectedError(err.Error())
	}
	return userExist.Bool, nil
}

func (r *UserRepoClass) SaveUser(ctx context.Context, user *Users) (*UserResponse, *errs.AppError) {

	inputArgs := make([]interface{}, 0, 10)

	sqlQuery := fmt.Sprintf(`INSERT INTO %s.users
				(user_name, first_name, last_name, "password", email, phone, address, is_admin, user_type, created_at, updated_at)
				VALUES( ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);`, r.dbSchema)
	inputArgs = append(inputArgs, user.UserName, user.FirstName, user.LastName, user.Password, user.Email,
				user.Phone, user.Address, user.IsAdmin, user.UserType, user.CreatedAt, user.UpdatedAt)

	tx, err := r.db.Begin()
	if err != nil {
		return nil, errs.NewUnexpectedError(err.Error())
	}

	sqlQuery = sqlx.Rebind(sqlx.DOLLAR, sqlQuery)

	txRow, txErr := tx.ExecContext(ctx, sqlQuery, inputArgs...)
	if txErr != nil {
		logger.Error(fmt.Sprintf("txErr Error: User/SaveUser API %s", txErr.Error()))
		_ = tx.Rollback()
		return nil, errs.NewUnexpectedError(txErr.Error())
	}

	if err = tx.Commit(); err != nil {
		logger.Error(fmt.Sprintf("txErr Commit Error: User/SaveUser API %s", txErr.Error()))
		return nil, errs.NewUnexpectedError(err.Error())
	}

	rows, err := txRow.RowsAffected()
	if rows > 0 {
		userResp := UserResponse{
			Success: true,
			Message: "User Inserted Successfully.",
		}
		return &userResp, nil
	} else {
		userResp := UserResponse{
			Success: false,
			Message: "User Insertion Failed.",
		}
		return &userResp, errs.NewUnexpectedError("Rows Unaffected.")
	}

}

func (r *UserRepoClass) GetPassword(ctx context.Context, cond *string) (string, *errs.AppError) {
	var password sql.NullString

	sqlQuery := fmt.Sprintf(`SELECT urs."password" FROM %s."users" urs WHERE (urs.user_name = $1 or urs.email = $2)`, r.dbSchema)

	err := r.db.QueryRowContext(ctx, sqlQuery, cond, cond).Scan(&password)
	if err != nil || !password.Valid {
		return "", errs.NewUnexpectedError(err.Error())
	}

	return password.String, nil
}

func (r *UserRepoClass) GetUser(ctx context.Context, cond *string) (*Users, *errs.AppError) {
	var userResp = Users{}
	sqlQuery := fmt.Sprintf(`select
									id,
									user_id,
									user_name,
									first_name,
									last_name,
									"password",
									email,
									phone,
									address,
									is_admin,
									user_type,
									created_at,
									updated_at
								from
									%s.users
								where
									(user_name = ? or email = ?);`, r.dbSchema)

	sqlQuery = sqlx.Rebind(sqlx.DOLLAR, sqlQuery)
	err := r.db.QueryRowContext(ctx, sqlQuery, cond, cond).Scan(&userResp.ID,
																&userResp.UserID,
																&userResp.UserName,
																&userResp.FirstName,
																&userResp.LastName,
																&userResp.Password,
																&userResp.Email,
																&userResp.Phone,
																&userResp.Address,
																&userResp.IsAdmin,
																&userResp.UserType,
																&userResp.CreatedAt,
																&userResp.UpdatedAt)
	if err != nil {
		return nil, errs.NewUnexpectedError(err.Error())
	}

	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			logger.Error(  fmt.Sprintf("GetUser: Defer Sql Func Error: %s", err.Error()))
			return
		}
	}(r.db)

	return &userResp, nil
}