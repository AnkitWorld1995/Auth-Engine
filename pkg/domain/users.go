package domain

import (
	"github.com/chsys/userauthenticationengine/pkg/dto"
	"time"
)

type Users struct {
	UserID		string 	`json:"id"`
	UserName 	string	`json:"user_name"`
	FirstName	string	`json:"first_name"`
	LastName	string	`json:"last_name"`
	Password	string  `json:"password"`
	Email		string	`json:"email"`
	Phone		int32	`json:"phone"`
	Address 	*string	`json:"address"`
	IsAdmin		bool	`json:"is_admin"`
	CreatedAt	string	`json:"created_at"`
	UpdatedAt 	string	`json:"updated_at"`
}

type UserResponse struct {
	Success			bool
	Message			string
	UserDetails 	*Users
}



func (u *UserResponse) ToDto() *dto.SignUpResponse {
	return &dto.SignUpResponse{
		Success:     u.Success,
		Message:     u.Message,
		UserDetails: &dto.UserResponse {
			UserID:    u.UserDetails.UserID,
			UserName:  u.UserDetails.UserName,
			FirstName: u.UserDetails.FirstName,
			LastName:  u.UserDetails.LastName,
			Password:  u.UserDetails.Password,
			Email:     u.UserDetails.Email,
			Phone:     u.UserDetails.Phone,
			Address:   u.UserDetails.Address,
			IsAdmin:   u.UserDetails.IsAdmin,
			CreatedAt: u.UserDetails.CreatedAt,
			UpdatedAt: u.UserDetails.UpdatedAt,
		},
	}
}


func CreateNewUser(userName , firstName, lastName, password, email string, address *string, phone int32, isAdmin bool) *Users{
	return &Users{
		UserName:  userName,
		FirstName: firstName,
		LastName:  lastName,
		Password:  password,
		Email:     email,
		Phone:     phone,
		Address:   address,
		IsAdmin:   isAdmin,
		CreatedAt: time.Now().Format("2014-11-12 11:45:26.371 +05:30 UTC"),
		UpdatedAt: time.Now().Format("2014-11-12 11:45:26.371 +05:30 UTC"),
	}
}