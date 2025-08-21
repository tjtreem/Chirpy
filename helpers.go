package main

import (
	"regexp"
	"net/http"
	"encoding/json"
)


func cleanProfanity(text string, profaneWords []string) string {
	for _, profaneWord := range profaneWords {
	    pattern := `(?i)\b` + profaneWord + `\b`
	    regex := regexp.MustCompile(pattern)
	    text = regex.ReplaceAllString(text, "****")
	}
	return text
}



func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	dat, err := json.Marshal(payload)
	if err != nil {
	    http.Error(w, "Error marshalling JSON", http.StatusInternalServerError)
	    return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(dat)
}



func respondWithError(w http.ResponseWriter, code int, msg string) {
	type errorResponse struct {
	    Error string `json:"error"`
	}

	resp := errorResponse{Error: msg}
	respondWithJSON(w, code, resp)
}





