package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/keithcrooks/chirpy/internal/database"
)

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ValidResponse struct {
	Valid bool `json:"valid"`
}

func (cfg *apiConfig) handlerCreateChirp(w http.ResponseWriter, req *http.Request) {
	chirp, err := validateChirp(req)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	filterChirp(&chirp)

	params := database.CreateChirpParams{Body: chirp.Body, UserID: chirp.UserID}
	dbChirp, err := cfg.db.CreateChirp(req.Context(), params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
	}

	chirp.ID = dbChirp.ID
	chirp.CreatedAt = dbChirp.CreatedAt
	chirp.UpdatedAt = dbChirp.UpdatedAt

	respondWithJSON(w, http.StatusCreated, chirp)
}

func filterChirp(chirp *Chirp) {
	const (
		replacement = "****"
		sep         = " "
	)

	words := strings.Split(chirp.Body, sep)

	for i, word := range words {
		if isProfane(word) {
			words[i] = replacement
		}
	}

	chirp.Body = strings.Join(words, sep)
}

func isProfane(word string) bool {
	profaneWords := []string{"kerfuffle", "sharbert", "fornax"}

	for _, profaneWord := range profaneWords {
		if strings.ToLower(word) == profaneWord {
			return true
		}
	}

	return false
}

func validateChirp(req *http.Request) (Chirp, error) {
	decoder := json.NewDecoder(req.Body)
	chirp := Chirp{}
	if err := decoder.Decode(&chirp); err != nil {
		log.Printf("Error decoding Chirp: %s", err)
		return chirp, errors.New("Could not read Chirp")
	}

	if len(chirp.Body) > 140 {
		return chirp, errors.New("Chirp is too long")
	}

	return chirp, nil
}
