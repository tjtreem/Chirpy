package main

import (
	"time"
	"net/http"
	"database/sql"
	"github.com/golang-jwt/jwt/v5"
	"github.com/tjtreem/Chirpy/internal/auth"
)



func (cfg *apiConfig) handlerRefresh(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
	    respondWithError(w, http.StatusUnauthorized, "missing or invalid authorization header")
	    return
	}

	dbGetUser, err := cfg.db.GetUserFromRefreshToken(r.Context(), token)
	if err == sql.ErrNoRows {
	    respondWithError(w, http.StatusUnauthorized, "token does not exist or is invalid")
	    return
	}
	if err != nil {
	    respondWithError(w, http.StatusInternalServerError, "unable to retrieve token")
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



func (cfg *apiConfig) handlerRevoke(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
	    respondWithError(w, http.StatusUnauthorized, "missing or invalid authorization header")
	    return
	}

	err = cfg.db.RevokeToken(r.Context(), token)
	if err != nil {
	    respondWithError(w, http.StatusInternalServerError, "unable to revoke token")
	    return
	}

	w.WriteHeader(http.StatusNoContent)

}
