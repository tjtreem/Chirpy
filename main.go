package main

import (
	"encoding/json"
	"os"
	"fmt"
	"log"
	"sort"
	"net/http"
	"sync/atomic"
	"database/sql"
	"github.com/google/uuid"
	"time"
	_ "github.com/lib/pq"
	"github.com/joho/godotenv"
	"github.com/golang-jwt/jwt/v5"
	"github.com/tjtreem/Chirpy/internal/database"
	"github.com/tjtreem/Chirpy/internal/auth"

)


type apiConfig struct {
    fileserverHits	atomic.Int32
    db			*database.Queries
    platform		string
    jwtSecret		string
    polkaKey		string
}

type ChirpResponse struct {
    ID		uuid.UUID `json:"id"`
    CreatedAt	time.Time `json:"created_at"`
    UpdatedAt	time.Time `json:"updated_at"`
    Body	string	  `json:"body"`
    UserID	uuid.UUID `json:"user_id"`
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
	    Password string `json:"password"`
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
	
	password, err := auth.HashPassword(params.Password)
	if err != nil {
	    respondWithError(w, http.StatusInternalServerError, "Unable to authenticate password")
	    return
	}

	dbUser, err := cfg.db.CreateUser(r.Context(), database.CreateUserParams{
	    ID:			id,
	    CreatedAt:		now,
    	    UpdatedAt:		now,
    	    Email:		params.Email,
	    HashedPassword:	password,
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
	    IsChirpyRed bool      `json:"is_chirpy_red"`
	}

	responseUser := User{
	    ID:		 dbUser.ID,
	    CreatedAt:	 dbUser.CreatedAt,
	    UpdatedAt:	 dbUser.UpdatedAt,
	    Email:	 dbUser.Email,
	    IsChirpyRed: dbUser.IsChirpyRed,
	}

	respondWithJSON(w, http.StatusCreated, responseUser)
}


func (cfg *apiConfig) handlerUserLogin(w http.ResponseWriter, r *http.Request) {
	type loginParams struct {
	    Email 	string	`json:"email"`
	    Password 	string	`json:"password"`
	}
	
	decoder := json.NewDecoder(r.Body)
	params := loginParams{}
	err := decoder.Decode(&params)

	if err != nil {
	    respondWithError(w, http.StatusBadRequest, "Couldn't decode parameters")
	    return
	}
	
	dbUser, err := cfg.db.GetUserByEmail(r.Context(), params.Email)
	if err != nil {
	    respondWithError(w, http.StatusUnauthorized, "Incorrect email or password")
	    return
	}
	
	

	err = auth.CheckPasswordHash(params.Password, dbUser.HashedPassword)
	if err != nil {
	    respondWithError(w, http.StatusUnauthorized, "Incorrect email or password")
	    return
	}


	dur := time.Hour
	now := time.Now().UTC()

	
	claims := jwt.RegisteredClaims{
	    Subject:	dbUser.ID.String(),
	    IssuedAt:	jwt.NewNumericDate(now),
	    ExpiresAt:	jwt.NewNumericDate(now.Add(dur)),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(cfg.jwtSecret))
	if err != nil {
	    respondWithError(w, http.StatusInternalServerError, "couldn't sign token")
	    return
	}
	
	ref_token, err := auth.MakeRefreshToken()
	if err != nil {
	    respondWithError(w, http.StatusInternalServerError, "unable to create refresh token")
	}

	_, err = cfg.db.RefreshToken(r.Context(), database.RefreshTokenParams{
	    Token:	ref_token,
	    UserID:	dbUser.ID,
	})
	
	if err != nil {
	    respondWithError(w, http.StatusInternalServerError, "unable to create refresh token")
	    return
	}


	type User struct {
	    ID		 uuid.UUID `json:"id"`
	    CreatedAt	 time.Time `json:"created_at"`
	    UpdatedAt	 time.Time `json:"updated_at"`
	    Email	 string    `json:"email"`
	    Token	 string    `json:"token"`
	    RefreshToken string    `json:"refresh_token"`
	    IsChirpyRed	 bool      `json:"is_chirpy_red"`
    	}

	responseUser := User{
	    ID:		  dbUser.ID,
	    CreatedAt:	  dbUser.CreatedAt,
	    UpdatedAt:	  dbUser.UpdatedAt,
	    Email:	  dbUser.Email,
	    Token:	  signed,
	    RefreshToken: ref_token,
	    IsChirpyRed:  dbUser.IsChirpyRed,
	}

	respondWithJSON(w, http.StatusOK, responseUser)
}




func (cfg *apiConfig) handlerCreateChirps(w http.ResponseWriter, r *http.Request) {
	type CreateChirpsParams struct {
	    Body 	string		`json:"body"`
	}
	
	decoder := json.NewDecoder(r.Body)
	params := CreateChirpsParams{}
	err := decoder.Decode(&params)

	if err != nil {
	    respondWithError(w, http.StatusBadRequest, "Couldn't decode parameters")
	    return
	}

	if len(params.Body) > 140 {
	    respondWithError(w, http.StatusBadRequest, "Chirp is too long")
	    return
	}

	cleaned := cleanProfanity(params.Body, []string{"kerfuffle", "sharbert", "fornax"})

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
	    respondWithError(w, http.StatusUnauthorized, "missing or invalid authorization header")
	    return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
	    respondWithError(w, http.StatusUnauthorized, "invalid or expired token")
	    return
	}

	
	dbChirp, err := cfg.db.CreateChirp(r.Context(), database.CreateChirpParams{
	    Body:	cleaned,
    	    UserID:	userID,
    	})

	if err != nil {
	    respondWithError(w, http.StatusInternalServerError, "Couldn't create chirp")
	    return
	}

	type Chirp struct {
	    ID		uuid.UUID `json:"id"`
	    CreatedAt	time.Time `json:"created_at"`
	    UpdatedAt	time.Time `json:"updated_at"`
	    Body	string	  `json:"body"`
	    UserID	uuid.UUID `json:"user_id"`
	}

	responseChirp := Chirp{
	    ID:		dbChirp.ID,
	    CreatedAt:	dbChirp.CreatedAt,
	    UpdatedAt:	dbChirp.UpdatedAt,
	    Body:	dbChirp.Body,
	    UserID:	dbChirp.UserID,
	}
	
	respondWithJSON(w, http.StatusCreated, responseChirp)

}


func (cfg *apiConfig) handlerRetrieveChirps(w http.ResponseWriter, r *http.Request) {
	s := r.URL.Query().Get("author_id")
	sortParam := r.URL.Query().Get("sort")
	sortDirection := "asc"
	if sortParam == "desc" {
	    sortDirection = "desc"
	}


	var (
	    chirps []database.Chirp
	    err    error
	)

	if s == "" {
	    chirps, err = cfg.db.GetChirps(r.Context())
	} else {
	    authorID, parseErr := uuid.Parse(s)
	    if parseErr != nil {
		respondWithError(w, http.StatusBadRequest, "invalid author_id")
		return
	    }

	    chirps, err = cfg.db.GetChirpsByUserID(r.Context(), authorID)
	}

	if err != nil {
	    respondWithError(w, http.StatusInternalServerError, "could not retrieve chirps")
	    return
	}


	apiChirps := make([]ChirpResponse, 0, len(chirps))

	for _, c := range chirps {
	    apiChirps = append(apiChirps, ChirpResponse{
	    	ID:		c.ID,
		CreatedAt:	c.CreatedAt,
	    	UpdatedAt:	c.UpdatedAt,
	    	Body:		c.Body,
	    	UserID:		c.UserID,
	    })
	}

	sort.Slice(apiChirps, func(i, j int) bool {
	    if sortDirection == "desc" {
		return apiChirps[i].CreatedAt.After(apiChirps[j].CreatedAt)
	    } else {
		return apiChirps[i].CreatedAt.Before(apiChirps[j].CreatedAt)
	      }
	})
 
	respondWithJSON(w, http.StatusOK, apiChirps)

}


func (cfg *apiConfig) handlerGetSingleChirp(w http.ResponseWriter, r *http.Request) {
	chirpID := r.PathValue("chirpID")

	id, err := uuid.Parse(chirpID)
	if err != nil {
	    respondWithError(w, http.StatusBadRequest, "invalid chirp ID")
	    return
	}

	chirp, err := cfg.db.GetChirp(r.Context(), id)
	if err == sql.ErrNoRows {
	    respondWithError(w, http.StatusNotFound, "no chirp found")
	    return
	} else if err != nil {
	      respondWithError(w, http.StatusInternalServerError, "unable to load chirp")
	      return
	}

	chirpResponse := ChirpResponse{
	    ID: chirp.ID,
	    CreatedAt: chirp.CreatedAt,
	    UpdatedAt: chirp.UpdatedAt,
	    Body: chirp.Body,
	    UserID: chirp.UserID,
	}

	respondWithJSON(w, http.StatusOK, chirpResponse)

}



func (cfg *apiConfig) handlerDeleteChirp(w http.ResponseWriter, r *http.Request) {
	chirpID := r.PathValue("chirpID")

	id, err := uuid.Parse(chirpID)
	if err != nil {
	    respondWithError(w, http.StatusBadRequest, "invalid chirp ID")
	    return
	}
	
	chirp, err := cfg.db.GetChirp(r.Context(), id)
	if err == sql.ErrNoRows {
	    respondWithError(w, http.StatusNotFound, "no chirp found")
	    return
	} else if err != nil {
	      respondWithError(w, http.StatusInternalServerError, "unable to load chirp")
	      return
	}

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
	
	if chirp.UserID != userID {
	    respondWithError(w, http.StatusForbidden, "forbidden")
	    return
	}

	err = cfg.db.DeleteChirp(r.Context(), id)
	if err != nil {
	    respondWithError(w, http.StatusInternalServerError, "failed to delete chirp")
	    return
	}

	w.WriteHeader(http.StatusNoContent)

}



func main () {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	secret := os.Getenv("Secret")
	polkaKey := os.Getenv("POLKA_KEY")
	if polkaKey == "" {
	    log.Fatal("missing POLKA_KEY env var")
	    return
	}
	if secret == "" {
	    log.Fatal("missing Secret env var")
	    return
	}
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
	    jwtSecret: secret,
	    polkaKey:  polkaKey,
	}

	mux := http.NewServeMux()

	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(".")))))
	mux.HandleFunc("GET /api/healthz", handlerReadiness)
	mux.HandleFunc("POST /api/chirps", apiCfg.handlerCreateChirps)
	mux.HandleFunc("GET /admin/metrics", apiCfg.handlerAdminMetrics)	
	mux.HandleFunc("POST /admin/reset", apiCfg.handlerAdminReset)
	mux.HandleFunc("POST /api/users", apiCfg.handlerCreateUser)
	mux.HandleFunc("PUT /api/users", apiCfg.handlerUsersUpdate)
	mux.HandleFunc("POST /api/login", apiCfg.handlerUserLogin)
	mux.HandleFunc("POST /api/refresh", apiCfg.handlerRefresh)
	mux.HandleFunc("POST /api/revoke", apiCfg.handlerRevoke)
	mux.HandleFunc("GET /api/chirps", apiCfg.handlerRetrieveChirps)	
	mux.HandleFunc("GET /api/chirps/{chirpID}", apiCfg.handlerGetSingleChirp)
	mux.HandleFunc("DELETE /api/chirps/{chirpID}", apiCfg.handlerDeleteChirp)
	mux.HandleFunc("POST /api/polka/webhooks", apiCfg.handlerWebhooks)

	srv := &http.Server{
	    Addr:	":" + port,
	    Handler: mux,
	}
	log.Fatal(srv.ListenAndServe())

}

