package main

import (
	"log"
	"net/http"
	"os"

	"mediahub/internal"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := internal.NewServer()
	srv.SetupRoutes()

	addr := ":" + port
	log.Printf("MediaHub démarre sur http://localhost%s", addr)
	if err := http.ListenAndServe(addr, srv.Middleware(srv.Mux)); err != nil {
		log.Fatal(err)
	}
}
