package main

import (
	"log"
	"net/http"
)


func main () {
	mux := http.NewServeMux()
	server := http.Server {
	    Handler: mux,
	    Addr: ":8080",
	}
	
	mux.Handle("/", http.FileServer(http.Dir(".")))

	err := server.ListenAndServe ()
	if err != nil {
	    log.Fatal(err)
	}











}
