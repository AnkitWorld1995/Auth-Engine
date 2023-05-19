package domain

import (
	"context"
	"github.com/chsys/userauthenticationengine/pkg/dto"
	"github.com/chsys/userauthenticationengine/pkg/lib/constants"
	errs "github.com/chsys/userauthenticationengine/pkg/lib/error"
	"log"
)

type JWT struct {
	Exp          float64  `json:"exp"`
	Iat          float64  `json:"iat"`
	Jti          string   `json:"jti"`
	Iss          string   `json:"iss"`
	Aud          []string `json:"aud"`
	Sub          string   `json:"sub"`
	Typ          string   `json:"typ"`
	Azp          string   `json:"azp"`
	SessionState string   `json:"session_state"`
	Acr          string   `json:"acr"`
	RealmAccess  struct {
		Roles []string `json:"roles"`
	} `json:"realm_access"`
	ResourceAccess struct {
		RealmManagement struct {
			Roles []string `json:"roles"`
		} `json:"realm-management"`
		AuthSvc struct {
			Roles []string `json:"roles"`
		} `json:"auth-svc"`
		Account struct {
			Roles []string `json:"roles"`
		} `json:"account"`
	} `json:"resource_access"`
	Token 			  TokenDetails `json:"jwt_details"`
	Scope 			  string `json:"scope"`
	Sid               string `json:"sid"`
	EmailVerified     bool   `json:"email_verified"`
	Name              string `json:"name"`
	PreferredUsername string `json:"preferred_username"`
	GivenName         string `json:"given_name"`
	FamilyName        string `json:"family_name"`
	Email             string `json:"email"`
}

type TokenDetails struct {
	AccessToken  		string `json:"accessToken"`
	RefreshToken 		string `json:"refreshToken"`
	ExpiresIn    		int    `json:"expiresIn"`
	RefreshExpiresIn 	int    `json:"refresh_expires_in"`
}

type TokenOptions struct {
	ClientID            *string   `json:"client_id,omitempty"`
	ClientSecret        *string   `json:"-"`
	GrantType           *string   `json:"grant_type,omitempty"`
	RefreshToken        *string   `json:"refresh_token,omitempty"`
	Scopes              *[]string `json:"-"`
	Scope               *string   `json:"scope,omitempty"`
	ResponseTypes       *[]string `json:"-"`
	ResponseType        *string   `json:"response_type,omitempty"`
	Permission          *string   `json:"permission,omitempty"`
	Username            *string   `json:"username,omitempty"`
	Password            *string   `json:"password,omitempty"`
	Totp                *string   `json:"totp,omitempty"`
	Code                *string   `json:"code,omitempty"`
	ClientAssertionType *string   `json:"client_assertion_type,omitempty"`
	ClientAssertion     *string   `json:"client_assertion,omitempty"`
}

type UserInfo struct {
	Sub                 *string          `json:"sub,omitempty"`
	Name                *string          `json:"name,omitempty"`
	GivenName           *string          `json:"given_name,omitempty"`
	FamilyName          *string          `json:"family_name,omitempty"`
	MiddleName          *string          `json:"middle_name,omitempty"`
	Nickname            *string          `json:"nickname,omitempty"`
	PreferredUsername   *string          `json:"preferred_username,omitempty"`
	Profile             *string          `json:"profile,omitempty"`
	Picture             *string          `json:"picture,omitempty"`
	Website             *string          `json:"website,omitempty"`
	Email               *string          `json:"email,omitempty"`
	EmailVerified       *bool            `json:"email_verified,omitempty"`
	Gender              *string          `json:"gender,omitempty"`
	ZoneInfo            *string          `json:"zoneinfo,omitempty"`
	Locale              *string          `json:"locale,omitempty"`
	PhoneNumber         *string          `json:"phone_number,omitempty"`
	PhoneNumberVerified *bool            `json:"phone_number_verified,omitempty"`
	Address  			*struct{
						Formatted     *string `json:"formatted,omitempty"`
						StreetAddress *string `json:"street_address,omitempty"`
						Locality      *string `json:"locality,omitempty"`
						Region        *string `json:"region,omitempty"`
						PostalCode    *string `json:"postal_code,omitempty"`
						Country       *string `json:"country,omitempty"`
	} `json:"address,omitempty"`
	UpdatedAt           *int             `json:"updated_at,omitempty"`
}

func (r *JWT) GetUserClaims(claimsMap dto.MapClaims)  {
	for key, value := range claimsMap {
		if key == "exp" {
			r.Exp = value.(float64)
		}

		if key == "iat" {
			r.Iat = value.(float64)
		}

		if key == "aud" {
			switch value.(type) {
				case []interface{} :
					for _ , audValue := range value.([]interface{}) {
						r.Aud = append(r.Aud, audValue.(string))
					}
				default:
					r.Aud = append(r.Aud, value.(string))
			}
		}

		if key == "session_state" {
			r.SessionState = value.(string)
		}

		if key == "realm_access" {
			for _ , realm := range value.(map[string]interface{}) {
				for _, realmStrRoles := range realm.([]interface{}) {
					r.RealmAccess.Roles = append(r.RealmAccess.Roles, realmStrRoles.(string))
				}
			}
		}

		if key == "resource_access" {
			for index, resources := range value.(map[string]interface{}) {
				if index == "account" {
					for _, account := range resources.(map[string]interface{}) {
						for _, accountRoles := range account.([]interface{}) {
							r.ResourceAccess.Account.Roles = append(r.ResourceAccess.Account.Roles, accountRoles.(string))
						}
					}
				}
				if index == "auth-svc"{
					for _, authSvc := range resources.(map[string]interface{}){
						for _, authSvcRoles := range authSvc.([]interface{}) {
							r.ResourceAccess.AuthSvc.Roles = append(r.ResourceAccess.AuthSvc.Roles , authSvcRoles.(string))
						}
					}
				}
				if index == "realm-management" {
					for _, realmMgmt := range resources.(map[string]interface{}) {
						for _, realmMgmtRoles := range realmMgmt.([]interface{}) {
							r.ResourceAccess.RealmManagement.Roles = append(r.ResourceAccess.RealmManagement.Roles, realmMgmtRoles.(string))
						}
					}
				}
			}
		}

		if key == "scope"{
			r.Scope = value.(string)
		}

		if key == "preferred_username"{
			r.PreferredUsername = value.(string)
		}

		if key == "email"{
			r.Email = value.(string)
		}
	}

	log.Println("\n\nThe Value OF JWT Struct", r)
}

func (r *JWT) SetTokenDetails(req TokenDetails) {
	r.Token.AccessToken 		= req.AccessToken
	r.Token.RefreshToken 		= req.RefreshToken
	r.Token.RefreshExpiresIn 	= req.RefreshExpiresIn
	r.Token.ExpiresIn 			= req.ExpiresIn
}

func (r *JWT) SSOJWTDetails() *dto.SSOSignInResponse {
	return &dto.SSOSignInResponse{
		AccessToken: 	r.Token.AccessToken,
		RefreshToken:   r.Token.RefreshToken,
		RefreshExpiresIn: r.Token.RefreshExpiresIn,
		SessionState: 	r.SessionState,
		IssuedAt: 		r.Iat,
		ExpiresIn:    	r.Exp,
	}
}



// GetUserDetail Must be Refactored To integrate Properly with functions.
func GetUserDetail(ctx context.Context) (*TokenDetails,*errs.AppError){
	value := ctx.Value(constants.UserMapKey)
	mapUserContext, ok := value.(TokenDetails)
	if !ok {
		return nil, errs.NewNotFoundError("GetUserDetail: User Details Not Found.")
	}
	return &mapUserContext, nil
}

