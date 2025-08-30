package main

import (
	"fmt"
	"net/http"
	"errors"
)



func (cfg *apiConfig) handlerRefresh(w http.ResponseWriter, r *http.Request) {
	token, err := GetBearerToken(r.Header)
	if err != nil {
	    respondWithError(w, StatusUnauthorized, "missing or invalid authorization header")
	    return
	}

	dbGetUser, err := cfg.db.GetUserFromRefreshToken(r.Context(), token)
	if err == sql.ErrNoRows {
	    respondWithError(w, StatusUnauthorized, "token does not exist or is invalid")
	    return
	}
	if err != nil {
	    respondWithError(w, StatusInternalServerError, "unable to retrieve token")
	    return
	}

	dur := time.Hour
	now := time.Now().UTC()

	
	claims := jwt.RegisteredClaims{
	    Subject:	dbGetUser.ID.String(),
	    IssuedAt:	jwt.NewNumericDate(now),
	    ExpiresAt:	jwt.NewNumericDate(now.Add(dur)),
	}

	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := tok.SignedString([]byte(cfg.jwtSecret))
	if err != nil {
	    respondWithError(w, http.StatusInternalServerError, "couldn't sign token")
	    return
	}
	
	type Token struct {
	    Token	string	`json:"token"`
	}

	responseToken := Token{
	    Token:	signed,
	}

	respondWithJSON(w, http.StatusOK, responseToken)

}

