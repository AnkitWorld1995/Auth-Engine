package domain

import (
	"github.com/chsys/userauthenticationengine/pkg/dto"
	"time"
)

type Users struct {
	ID 			string	`json:"id"`
	UserID		string 	`json:"user_id"`
	UserName 	string	`json:"user_name"`
	FirstName	string	`json:"first_name"`
	LastName	string	`json:"last_name"`
	Password	string  `json:"password"`
	Email		string	`json:"email"`
	Phone		int32	`json:"phone"`
	Address 	*string	`json:"address"`
	IsAdmin		bool	`json:"is_admin"`
	UserType	string	`json:"user_type"`
	CreatedAt	string	`json:"created_at"`
	UpdatedAt 	string	`json:"updated_at"`
}

type UserResponse struct {
	Success			bool
	Message			string
	UserDetails 	*Users
}



func (u *UserResponse) ToSignUpDTO() *dto.SignUpResponse {
	return &dto.SignUpResponse{
		Success:     u.Success,
		Message:     u.Message,
	}
}

func (u *Users) ToSignInDTO() *dto.SignInResponse {
	return &dto.SignInResponse{
		ID:        u.ID,
		UserID:    u.UserID,
		UserName:  u.UserName,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Email:     u.Email,
		Phone:     u.Phone,
		Address:   u.Address,
		IsAdmin:   u.IsAdmin,
		UserType:  u.UserType,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}


func CreateNewUser(userName , firstName, lastName, password, email, userType string, address *string, phone int32, isAdmin bool) *Users{
	return &Users{
		UserName:  userName,
		FirstName: firstName,
		LastName:  lastName,
		Password:  password,
		Email:     email,
		Phone:     phone,
		Address:   address,
		IsAdmin:   isAdmin,
		UserType:  userType,
		CreatedAt: time.Now().UTC().Format(time.RFC822),
		UpdatedAt: time.Now().UTC().Format(time.RFC822),
	}
}

func  GetAllUser(userList []*Users) []*dto.Users {

	var userLst []*dto.Users
	for _, v := range userList {
		var users dto.Users

		users.ID			= v.ID
		users.UserID 		= v.UserID
		users.UserName 		= v.UserName
		users.FirstName 	= v.FirstName
		users.LastName 		= v.LastName
		users.Email 		= v.Email
		users.Password 		= v.Password
		users.Phone 		= v.Phone
		users.Address 		= v.Address
		users.IsAdmin 		= v.IsAdmin
		users.UserType 		= v.UserType
		users.CreatedAt 	= v.CreatedAt
		users.UpdatedAt 	= v.UpdatedAt

		userLst = append(userLst, &users)
	}
	return userLst
}