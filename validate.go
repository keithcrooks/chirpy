package main

import (
	"encoding/json"
	"log"
	"net/http"
)

type Chirp struct {
	Body string `json:"body"`
}

type CleanedChirp struct {
	CleanedBody string `json:"cleaned_body"`
}

type ValidResponse struct {
	Valid bool `json:"valid"`
}

func handlerValidateChirp(w http.ResponseWriter, req *http.Request) {
	decoder := json.NewDecoder(req.Body)
	c := Chirp{}
	if err := decoder.Decode(&c); err != nil {
		log.Printf("Error decoding Chirp: %s", err)
		respondWithError(w, http.StatusBadRequest, "Could not read Chirp")
		return
	}

	if len(c.Body) > 140 {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long")
		return
	}

	cleaned := filterChirp(c)

	respondWithJSON(w, http.StatusOK, cleaned)
}
