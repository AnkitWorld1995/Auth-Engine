package domain

import (
	"fmt"
	"github.com/chsys/userauthenticationengine/pkg/dto"
)

func FilterUserClauses(req *dto.AllUsersRequest, query string, isCount bool) (string, []interface{}) {
	var (
		inputArgs []interface{}
		isWhere   bool
	)

	// FilterUserClauses
	if req.UserID != nil {
		query += fmt.Sprintf(" Where urs.user_id = ? ")
		inputArgs = append(inputArgs, req.UserID)
		isWhere = true
	}

	if  req.IdIn != nil && len(req.IdIn) > 0 {
		if isWhere {
			query += fmt.Sprintf(" or urs.user_id in (?) ")
			inputArgs = append(inputArgs, req.IdIn)
		}else {
			query += fmt.Sprintf(" Where urs.user_id in (?) ")
			inputArgs = append(inputArgs, req.IdIn)
			isWhere = true
		}
	}

	if req.Email != nil && req.IdIn == nil {
		if isWhere {
			query += fmt.Sprintf(" AND urs.email = ? ")
			inputArgs = append(inputArgs, req.Email)
		} else {
			query += fmt.Sprintf(" Where urs.email = ? ")
			inputArgs = append(inputArgs, req.Email)
			isWhere = true
		}

	}

	if req.IsVerified != nil && req.IdIn == nil {
		if isWhere {
			query += fmt.Sprintf(" AND urs.is_verified = ? ")
			inputArgs = append(inputArgs, req.IsVerified)
		} else {
			query += fmt.Sprintf(" Where urs.is_verified = ? ")
			inputArgs = append(inputArgs, req.IsVerified)
			isWhere = true
		}

	}

	if req.IsBlocked != nil && req.IdIn == nil {
		if isWhere {
			query += fmt.Sprintf(" AND urs.is_blocked = ? ")
			inputArgs = append(inputArgs, req.IsBlocked)
		} else {
			query += fmt.Sprintf(" Where urs.is_blocked = ? ")
			inputArgs = append(inputArgs, req.IsBlocked)
			isWhere = true
		}

	}

	if !isCount {
		if req.Limit != nil && req.Offset != nil {
			query += fmt.Sprintf(" order by urs.user_id asc limit ? offset ? ")
			inputArgs = append(inputArgs, req.Limit, req.Offset)
		} else {
			query += fmt.Sprintf(" order by urs.user_id asc limit 10 offset 0 asc")
		}
	}


	return query, inputArgs
}
