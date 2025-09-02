package main

import (
	"net/http"
	"encoding/json"
	"github.com/tjtreem/Chirpy/internal/database"
	"github.com/tjtreem/Chirpy/internal/auth"


)


func (cfg *apiConfig) handlerUsersUpdate(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
	    respondWithError(w, http.StatusUnauthorized, "missing or invalid authorization header")
	    return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
	    respondWithError(w, http.StatusUnauthorized, "couldn't validate token")
	    return
	}

	dec := json.NewDecoder(r.Body)

	type updateParams struct {
	    Email	string  `json:"email"`
	    Password	string  `json:"password"`
	}
	
	var p updateParams

	if err := dec.Decode(&p); err != nil {
	    respondWithError(w, http.StatusBadRequest, "Couldn't decode parameters")
	    return
	}
	
	if p.Email == "" || p.Password == "" {
	    respondWithError(w, http.StatusBadRequest, "email and password are required")
	    return
	}

	hashed, err := auth.HashPassword(p.Password)
	if err != nil {
	    respondWithError(w, http.StatusInternalServerError, "couldn't hash password")
	    return
	}
	
	updated, err := cfg.db.UpdateUserByID(r.Context(), database.UpdateUserByIDParams{
	    ID:			userID,
	    Email:		p.Email,
	    HashedPassword:	string(hashed),
	})

	if err != nil {
	    respondWithError(w, http.StatusBadRequest, "couldn't update user")
	    return
	}
	
	respondWithJSON(w, http.StatusOK, map[string]any{
	    "id":	updated.ID,
	    "email":	updated.Email,
	})

}
