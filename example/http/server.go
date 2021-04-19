package main

import (
	"net/http"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type H struct {
}

func (h H) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("hello"))
}

func main() {
	var h H
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	log.Print("hello world")
	return
	http.ListenAndServe(":7070", &h)
}
