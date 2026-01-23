package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/keithcrooks/chirpy/internal/auth"
	"github.com/keithcrooks/chirpy/internal/database"
)

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Chirps struct {
	Entries []Chirp
}

type ValidResponse struct {
	Valid bool `json:"valid"`
}

func (cfg *apiConfig) handlerCreateChirp(w http.ResponseWriter, req *http.Request) {
	chirp, err := validateChirp(req)
	if err != nil {
		log.Printf("Error validating Chirp: %v", err)
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	token, err := auth.GetBearerToken(req.Header)
	if err != nil {
		log.Printf("Error getting Bearer token: %v", err)
		respondWithError(w, http.StatusBadRequest, "Invalid bearer token")
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.tokenSecret)
	if err != nil {
		log.Printf("Error validating JWT: %v", err)
		respondWithError(w, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		return
	}

	if userID != chirp.UserID {
		log.Printf("Unauthorized user %s attempted to Chirp as %s", userID, chirp.UserID)
		respondWithError(w, http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
		return
	}

	filterChirp(&chirp)

	params := database.CreateChirpParams{Body: chirp.Body, UserID: chirp.UserID}
	dbChirp, err := cfg.db.CreateChirp(req.Context(), params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	chirp.ID = dbChirp.ID
	chirp.CreatedAt = dbChirp.CreatedAt
	chirp.UpdatedAt = dbChirp.UpdatedAt

	respondWithJSON(w, http.StatusCreated, chirp)
}

func (cfg *apiConfig) handlerGetAllChirps(w http.ResponseWriter, req *http.Request) {
	dbChirps, err := cfg.db.GetAllChirps(req.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not get all chirps!")
		return
	}

	chirps := Chirps{Entries: []Chirp{}}

	for _, dbChirp := range dbChirps {
		chirp := Chirp{
			ID:        dbChirp.ID,
			CreatedAt: dbChirp.CreatedAt,
			UpdatedAt: dbChirp.UpdatedAt,
			Body:      dbChirp.Body,
			UserID:    dbChirp.UserID,
		}

		chirps.Entries = append(chirps.Entries, chirp)
	}

	respondWithJSON(w, http.StatusOK, chirps.Entries)
}

func (cfg *apiConfig) handlerGetChirp(w http.ResponseWriter, req *http.Request) {
	chirpID := req.PathValue("chirpID")
	chirpUUID, err := uuid.Parse(chirpID)
	if err != nil {
		log.Printf("Error parsing Chirp ID: %s", err)
		respondWithError(w, http.StatusBadRequest, "Invalid Chirp ID")
		return
	}

	log.Printf("chirpUUID: %v", chirpUUID)

	dbChirp, err := cfg.db.GetChirp(req.Context(), chirpUUID)
	if err != nil {
		log.Printf("Error getting Chirp from DB: %s", err)

		switch err {
		case sql.ErrNoRows:
			respondWithError(w, http.StatusNotFound, "Chirp not found")
		default:
			respondWithError(w, http.StatusInternalServerError, "Unknown error getting Chirp")
		}
		return
	}

	chirp := Chirp{
		ID:        dbChirp.ID,
		CreatedAt: dbChirp.CreatedAt,
		UpdatedAt: dbChirp.UpdatedAt,
		Body:      dbChirp.Body,
		UserID:    dbChirp.UserID,
	}

	respondWithJSON(w, http.StatusOK, chirp)
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
