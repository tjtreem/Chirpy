package main

import (
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
)


type apiConfig struct {
    fileserverHits	atomic.Int32
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
    w.Header().Add("Content-Type", "text/plain; charset=utf-8")
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("OK"))
    cfg.fileserverHits.Store(0)
}


func handlerReadiness(w http.ResponseWriter, r *http.Request) {
    w.Header().Add("Content-Type", "text/plain; charset=utf-8")
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("OK"))
}
	

func main () {
	const port = "8080"

	apiCfg := apiConfig{
	    fileserverHits:  atomic.Int32{},
	}

	mux := http.NewServeMux()

	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(".")))))
	mux.HandleFunc("GET /api/healthz", handlerReadiness)
	mux.HandleFunc("GET /admin/metrics", apiCfg.handlerAdminMetrics)	
	mux.HandleFunc("POST /admin/reset", apiCfg.handlerAdminReset)

	srv := &http.Server{
	    Addr:	":" + port,
	    Handler: mux,
	}
	log.Fatal(srv.ListenAndServe())
	

	
}
