package main

import (
	"encoding/json"
	"net/http"
	"github.com/google/uuid"
	"github.com/tjtreem/Chirpy/internal/auth"
)


func (cfg *apiConfig) handlerWebhooks(w http.ResponseWriter, r *http.Request) {
	type polkaWebhookRequest struct {
	    Event  string  `json:"event"`
	    Data   struct  {
		UserID  uuid.UUID `json:"user_id"`
	    } `json:"data"`
	}

	var req polkaWebhookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
	    respondWithError(w, http.StatusBadRequest, "bad request")
	    return
	}
	
	key, err := auth.GetAPIKey(r.Header)
	if err !=nil || key != cfg.polkaKey {
	    respondWithError(w, http.StatusUnauthorized, "unauthorized")
	    return
	}


	if req.Event != "user.upgraded" {
	    w.WriteHeader(http.StatusNoContent)
	    return
	}

	res, err := cfg.db.UpgradeUserToChirpyRed(r.Context(), req.Data.UserID)
	if err != nil {
	    respondWithError(w, http.StatusInternalServerError, "internal error")
	    return
	}

	n, _ := res.RowsAffected()
	if n == 0 {
	    respondWithError(w, http.StatusNotFound, "not found")
	    return
	}
	
	w.WriteHeader(http.StatusNoContent)

}









