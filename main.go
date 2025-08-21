package main

import (
	"encoding/json"
	"os"
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
	"database/sql"
	"github.com/google/uuid"
	"time"
	_ "github.com/lib/pq"
	"github.com/joho/godotenv"
	"github.com/tjtreem/Chirpy/internal/database"

)


type apiConfig struct {
    fileserverHits	atomic.Int32
    db			*database.Queries
    platform		string
}


func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	cfg.fileserverHits.Add(1)
	next.ServeHTTP(w, r)
	})
}


func (cfg *apiConfig) handlerAdminMetrics(w http.ResponseWriter, r *http.Request) {
    w.Header().Add("Content-Type", "text/html; charset=utf-8")
    w.WriteHeader(http.StatusOK)
    w.Write([]byte(fmt.Sprintf(
	`<html>
  	  <body>
    	    <h1>Welcome, Chirpy Admin</h1>
    	    <p>Chirpy has been visited %d times!</p>
  	  </body>
	</html>`, 
	cfg.fileserverHits.Load())))
}


func (cfg *apiConfig) handlerAdminReset(w http.ResponseWriter, r *http.Request) {
	if cfg.platform != "dev" {
	    respondWithError(w, http.StatusForbidden, "403 forbidden")
	    return
        }

	err := cfg.db.DeleteAllUsers(r.Context())
	if err != nil {
	    respondWithError(w, http.StatusInternalServerError, "Failed to delete users")
	    return
	}

	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
    	w.WriteHeader(http.StatusOK)
    	w.Write([]byte("OK"))
}



func handlerReadiness(w http.ResponseWriter, r *http.Request) {
    w.Header().Add("Content-Type", "text/plain; charset=utf-8")
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("OK"))
}
	

func (cfg *apiConfig) handlerValidateChirp(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
	    Body string `json:"body"`
	}

	type validateChirpResponse struct {
	    CleanedBody string `json:"cleaned_body"`
	}


	// Parse incoming json
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	
	// Decoding errors
	if err != nil {
	    respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters")
	    return
	}
	
	// If Chirp is too long
	if len(params.Body) > 140 {
	    respondWithError(w, http.StatusBadRequest, "Chirp is too long")
	    return
	}

		
	// Clean any profanity
	cleaned := cleanProfanity(params.Body, []string{"kerfuffle", "sharbert", "fornax"})
	respondWithJSON(w, http.StatusOK, validateChirpResponse{
	    CleanedBody: cleaned,
	})


}


func(cfg *apiConfig) handlerCreateUser(w http.ResponseWriter, r *http.Request) {
	type createUserParams struct {
	    Email string `json:"email"`
	}
	
	decoder := json.NewDecoder(r.Body)
	params := createUserParams{}
	err := decoder.Decode(&params)

	if err != nil {
	    respondWithError(w, http.StatusBadRequest, "Couldn't decode parameters")
	    return
	}
	
	id := uuid.New()
	now := time.Now().UTC()


	dbUser, err := cfg.db.CreateUser(r.Context(), database.CreateUserParams{
	    ID:		id,
	    CreatedAt:	now,
    	    UpdatedAt:	now,
    	    Email:	params.Email,
    	})

	    if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create user")
		return
	}

	type User struct {
	    ID		uuid.UUID `json:"id"`
	    CreatedAt	time.Time `json:"created_at"`
	    UpdatedAt	time.Time `json:"updated_at"`
	    Email	string	  `json:"email"`
	}

	responseUser := User{
	    ID:		dbUser.ID,
	    CreatedAt:	dbUser.CreatedAt,
	    UpdatedAt:	dbUser.UpdatedAt,
	    Email:	dbUser.Email,
	}

	respondWithJSON(w, http.StatusCreated, responseUser)
}





func main () {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
	    fmt.Println("Error opening database")
	    return
	}

	dbQueries := database.New(db)


	const port = "8080"

	apiCfg := apiConfig{
	    fileserverHits:  atomic.Int32{},
	    db: dbQueries,
	    platform: os.Getenv("PLATFORM"),
	}

	mux := http.NewServeMux()

	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(".")))))
	mux.HandleFunc("GET /api/healthz", handlerReadiness)
	mux.HandleFunc("POST /api/validate_chirp", apiCfg.handlerValidateChirp)
	mux.HandleFunc("GET /admin/metrics", apiCfg.handlerAdminMetrics)	
	mux.HandleFunc("POST /admin/reset", apiCfg.handlerAdminReset)
	mux.HandleFunc("POST /api/users", apiCfg.handlerCreateUser)
	


	srv := &http.Server{
	    Addr:	":" + port,
	    Handler: mux,
	}
	log.Fatal(srv.ListenAndServe())

}
	

	

