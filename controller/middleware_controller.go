package controller

import (
	"net/http"
	"strings"

	errors "github.com/go-openapi/errors"

	"github.com/freecloudio/freecloud/auth"
	"github.com/freecloudio/freecloud/models"
)

func FileServerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api") || strings.HasPrefix(r.URL.Path, "/swagger.json") {
			next.ServeHTTP(w, r)
		} else {
			http.FileServer(http.Dir("./client/build")).ServeHTTP(w, r)
		}
	})
}

func ValidateToken(token string, scopes []string) (principal *models.Principal, err error) {
	principal = &models.Principal{Token: &models.Token{Token: token}}

	if len(scopes) > 0 {
		var session *models.Session
		session, err = models.ParseSessionTokenString(token)
		if err != nil {
			return nil, errors.New(http.StatusUnauthorized, "Token could not be parsed")
		}

		valid := auth.ValidateSession(session)
		if !valid {
			return nil, errors.New(http.StatusUnauthorized, "No valid session")
		}

		principal.User, err = auth.GetUserByID(session.UserID)
		if err != nil {
			return nil, errors.New(http.StatusInternalServerError, err.Error())
		}

		if isUserScope(scopes) || (isAdminScope(scopes) && principal.User.IsAdmin) {
			return
		} else {
			return nil, errors.New(http.StatusForbidden, "Insufficient privileges")
		}
	}

	return
}

func isUserScope(scopes []string) bool {
	userScope := false
	for _, scope := range scopes {
		if scope == "user" {
			userScope = true
			break
		}
	}

	return userScope
}

func isAdminScope(scopes []string) bool {
	adminScope := false
	for _, scope := range scopes {
		if scope == "admin" {
			adminScope = true
			break
		}
	}

	return adminScope
}
