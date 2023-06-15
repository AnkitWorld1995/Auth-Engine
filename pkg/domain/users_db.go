package domain

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/chsys/userauthenticationengine/pkg/dto"
	errs "github.com/chsys/userauthenticationengine/pkg/lib/error"
	"github.com/chsys/userauthenticationengine/pkg/lib/logger"
	"github.com/jmoiron/sqlx"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

type UserRepoClass struct {
	db            *sql.DB
	mongo         *mongo.Client
	dbSchema      string
	nosqlDatabase string
	collection    map[string]string
	awsConfig 	  *aws.Config
}

func NewUserRepoClass(rdbClient *sql.DB, mdbClient *mongo.Client, sqlSchema, nosqlDatabase string, collection map[string]string, awsConfig *aws.Config) *UserRepoClass {
	return &UserRepoClass{
		db:            rdbClient,
		mongo:         mdbClient,
		dbSchema:      sqlSchema,
		nosqlDatabase: nosqlDatabase,
		collection:    collection,
		awsConfig: 	   awsConfig,
	}
}

type UserRepository interface {
	FindByUserId(ctx context.Context, id *int) (bool, *errs.AppError)
	FindByEmail(ctx context.Context, email string) (bool, *errs.AppError)
	FindByUserName(ctx context.Context, userName string) (bool, *errs.AppError)
	SaveUser(ctx context.Context, user *Users) (*UserResponse, *errs.AppError)
	GetPassword(ctx context.Context, cond *string) (string, *errs.AppError)
	GetUser(ctx context.Context, userID *int, userName, email *string) (*Users, *errs.AppError)
	UpdatePassword(ctx context.Context, email, password string) (*dto.GenericResponse, *errs.AppError)
	GetAllUsersCount(ctx context.Context, req *dto.AllUsersRequest) (*int32, *errs.AppError)
	GetAllUsers(ctx context.Context, req *dto.AllUsersRequest) ([]*Users, *errs.AppError)
	UploadFilesWriteDB(ctx context.Context, userId int ,req *UploadFileMetaData) (*UploadFileDBResponse,*errs.AppError)
}

func (r *UserRepoClass) FindByEmail(ctx context.Context, email string) (bool, *errs.AppError) {
	var emailExist sql.NullBool
	query := fmt.Sprintf(`SELECT 1 FROM %s."users" WHERE "email" = $1`, r.dbSchema)
	err := r.db.QueryRowContext(ctx, query, email).Scan(&emailExist)
	if err != nil {
		return false, errs.NewNotFoundError("Email Not Found")
	}
	return emailExist.Bool, nil
}

func (r *UserRepoClass) FindByUserName(ctx context.Context, userName string) (bool, *errs.AppError) {
	var userExist sql.NullBool
	query := fmt.Sprintf(`SELECT 1 FROM %s."users" WHERE "user_name" = $1`, r.dbSchema)
	err := r.db.QueryRowContext(ctx, query, userName).Scan(&userExist)
	if err != nil {
		return false, errs.NewNotFoundError("User Name Not Found")
	}
	return userExist.Bool, nil
}

func (r *UserRepoClass) UpdatePassword(ctx context.Context, email, password string) (*dto.GenericResponse, *errs.AppError) {

	sqlQuery := fmt.Sprintf(`UPDATE 	%s.users 
									set
										"password" = ?
									where
										email = ?
									`, r.dbSchema)

	sqlQuery = sqlx.Rebind(sqlx.DOLLAR, sqlQuery)
	rows, err := r.db.ExecContext(ctx, sqlQuery, password, email)
	if err != nil {
		logger.Error(fmt.Sprintf("SQL: Update Password ERROR\t %s", err.Error()))
		return &dto.GenericResponse{
			Success: false,
			Message: fmt.Sprintf("SQL: Update Password ERROR\t %s", err.Error()),
		}, errs.NewUnexpectedError(err.Error())
	}

	affectedRows, _ := rows.RowsAffected()
	if affectedRows < 1 {
		return &dto.GenericResponse{
			Success: false,
			Message: "Password Not Updated In DB.",
		}, errs.NewUnexpectedError("Password Not Updated In DB.")
	} else {
		return &dto.GenericResponse{
			Success: true,
			Message: "Password Updated Successfully In DB.",
		}, nil
	}
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

func (r *UserRepoClass) FindByUserId(ctx context.Context, id *int) (bool, *errs.AppError) {
	var userID sql.NullBool

	sqlQuery := fmt.Sprintf(`SELECT 1 FROM %s."users" urs WHERE urs.user_id = $1;`, r.dbSchema)

	err := r.db.QueryRowContext(ctx, sqlQuery, id).Scan(&userID)
	if err != nil || !userID.Valid {
		return false, errs.NewNotFoundError(err.Error())
	}

	return userID.Bool, nil
}

func (r *UserRepoClass) GetUser(ctx context.Context, userID *int, userName, email *string) (*Users, *errs.AppError) {
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
									%s.users urs
								where
									urs."user_id" = ? or (urs."user_name" = ? or urs."email" = ?);`, r.dbSchema)

	sqlQuery = sqlx.Rebind(sqlx.DOLLAR, sqlQuery)
	err := r.db.QueryRowContext(ctx, sqlQuery, userID, userName, email).Scan(&userResp.ID,
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
		defer func(db *sql.DB) {
			err := db.Close()
			if err != nil {
				logger.Error(fmt.Sprintf("GetUserById: Defer Sql Func Error: %s", err.Error()))
				return
			}
		}(r.db)
		return nil, errs.NewUnexpectedError(err.Error())
	}
	return &userResp, nil
}

func (r *UserRepoClass) GetAllUsersCount(ctx context.Context, req *dto.AllUsersRequest) (*int32, *errs.AppError) {

	var count sql.NullInt32
	query := fmt.Sprintf(`SELECT distinct(count(*)) FROM %s.users urs `, r.dbSchema)
	query, inputArgs := FilterUserClauses(req, query, true)


	query, inputArgs, err := sqlx.In(query, inputArgs...)
	if err != nil {
		logger.Error("=========== DB Sql In Error: Count ============")
		return nil, errs.NewUnexpectedError(err.Error())
	}

	query = sqlx.Rebind(sqlx.DOLLAR, query)
	err = r.db.QueryRowContext(ctx, query, inputArgs...).Scan(&count)
	if err != nil {
		logger.Error("=========== DB Sql Error:Count ============")
		return nil, errs.NewUnexpectedError(err.Error())
	}

	if !count.Valid {
		return nil, errs.NewNoContentError("Sorry, Count is Empty!! ")
	} else {
		return &count.Int32, nil
	}
}

func (r *UserRepoClass) GetAllUsers(ctx context.Context, req *dto.AllUsersRequest) ([]*Users, *errs.AppError) {

	query := fmt.Sprintf(`SELECT * FROM %s.users urs `, r.dbSchema)
	query, inputArgs := FilterUserClauses(req, query, false)



	query, inputArgs, err := sqlx.In(query, inputArgs...)
	if err != nil {
		logger.Error("=========== DB Sql In Error ============")
		return nil, errs.NewUnexpectedError(err.Error())
	}

	query = sqlx.Rebind(sqlx.DOLLAR, query)
	rows, err := r.db.QueryContext(ctx, query, inputArgs...)
	if err != nil {
		logger.Error("=========== DB Sql Error ============")
		return nil, errs.NewUnexpectedError(err.Error())
	}

	defer rows.Close()

	var users []*Users
	var count int
	for rows.Next() {
		var userResp = Users{}
		err := rows.Scan(&userResp.ID,
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
			logger.Error("=========== DB Sql Scan Error ============")
			return nil, errs.NewUnexpectedError(err.Error())
		}

		users = append(users, &userResp)
		count++
	}

	if count < 1 {
		return nil, errs.NewNoContentError("Sorry, No Data Found!! ")
	} else {
		return users, nil
	}
}

func (r *UserRepoClass) UploadFilesWriteDB(ctx context.Context, userId int ,req *UploadFileMetaData) (*UploadFileDBResponse,*errs.AppError) {

	var uploadId string
	marshalledValue, _ := json.Marshal(req)

	sqlQuery := fmt.Sprintf(`INSERT INTO %s.upload_files
				(user_id, upload_file, created_at, updated_at)
				VALUES( ?, ?, now(), now()) returning id;`, r.dbSchema)

	sqlQuery = sqlx.Rebind(sqlx.DOLLAR, sqlQuery)
	tx, err := r.db.Begin()
	if err != nil {
		logger.Error("Service/Upload/UploadFilesWriteDB", zap.String("Upload: TX ERROR", err.Error()))
		return nil,  errs.NewUnexpectedError(err.Error())
	}

	sqlErr := tx.QueryRowContext(ctx, sqlQuery, userId, string(marshalledValue)).Scan(&uploadId)
	if err != nil {
		logger.Error("Service/Upload/UploadFilesWriteDB", zap.String("Upload: TX Execute ERROR", err.Error()))
		return nil, errs.NewUnexpectedError(err.Error())
	}

	sqlErr = tx.Commit()
	if sqlErr != nil {
		sqlErr := tx.Rollback()
		if sqlErr != nil {
			return nil, nil
		}
		return nil, nil
	}

	return &UploadFileDBResponse{
		UploadedID: uploadId,
		URL:        req.URL,
	}, nil
}


