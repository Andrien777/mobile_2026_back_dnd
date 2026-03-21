package main

import (
	"dnd_back/server"
	"flag"
	"log"
	"net/http"

	"dnd_back/api"
	"dnd_back/auth"

	"github.com/go-chi/chi/v5"
)

func main() {
	port := flag.String("port", "8080", "port where to serve traffic")

	r := chi.NewRouter()

	// Create a fake authenticator. This allows us to issue tokens, and also
	// implements a validator to check their validity.
	fa, err := auth.NewFakeAuthenticator()
	if err != nil {
		log.Fatalln("error creating authenticator:", err)
	}

	// Create middleware for validating tokens.
	mw, err := auth.CreateMiddleware(fa)
	if err != nil {
		log.Fatalln("error creating middleware:", err)
	}

	h := api.HandlerFromMux(api.NewStrictHandler(new(server.NewServer()), []api.StrictMiddlewareFunc{}), r)
	// wrap the existing handler with our global middleware
	h = mw(h)

	s := &http.Server{
		Handler: h,
		Addr:    "0.0.0.0:" + *port,
	}

	log.Fatal(s.ListenAndServe())
}
