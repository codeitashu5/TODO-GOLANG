package middlewares

import (
	"context"
	"github.com/dgrijalva/jwt-go"
	"net/http"
	"simpleHttpRequest/database/dbHelper"
	"simpleHttpRequest/models"
	"simpleHttpRequest/utils"
	"time"
)

const UserContext = "UserKey"

//no longer using it

func AuthMiddleWare(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("token")
		user, err := dbHelper.GetUserUsingToken(token)

		if err != nil {
			utils.RespondError(w, http.StatusUnauthorized, err, "session not valid")
			return
		}

		authContext := context.WithValue(r.Context(), UserContext, user)
		next.ServeHTTP(w, r.WithContext(authContext))
	})
}

func AuthMiddleWareJwt(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenStr := r.Header.Get("token")

		var claims models.Claims
		tkn, err := jwt.ParseWithClaims(tokenStr, &claims, func(token *jwt.Token) (interface{}, error) {
			return models.JwtKey, nil
		})

		//check if session is valid
		valid, err := dbHelper.CheckIfUserSessionIsValid(claims.SessionToke)
		if err != nil {
			utils.RespondError(w, http.StatusInternalServerError, err, "Un-able to execute")
			return
		}

		if !valid {
			utils.RespondError(w, http.StatusForbidden, err, "Invalid session")
			return
		}

		//check if token time is valid
		if claims.ExpiresAt < time.Now().Unix() {
			//log-out user
			err := dbHelper.LogOutUser(claims.SessionToke)
			if err != nil {
				utils.RespondError(w, http.StatusInternalServerError, err, "Un-able to logout")
				return
			}
			utils.RespondError(w, http.StatusBadRequest, err, "All ready loged out")
			return
		}

		if err != nil {
			if err == jwt.ErrSignatureInvalid {
				utils.RespondError(w, http.StatusUnauthorized, err, "session not valid")
				return
			}
			utils.RespondError(w, http.StatusBadRequest, err, "session not valid")
			return
		}

		if !tkn.Valid {
			utils.RespondError(w, http.StatusForbidden, err, "Log-ing out session got expired")
			return
		}

		authContext := context.WithValue(r.Context(), UserContext, claims)
		next.ServeHTTP(w, r.WithContext(authContext))

	})
}
