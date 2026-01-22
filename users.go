package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/keithcrooks/chirpy/internal/auth"
	"github.com/keithcrooks/chirpy/internal/database"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
	Password  string    `json:"password,omitempty"`
}

func (cfg *apiConfig) handlerAddUser(w http.ResponseWriter, req *http.Request) {
	var user User

	decoder := json.NewDecoder(req.Body)

	if err := decoder.Decode(&user); err != nil {
		respondWithError(w, http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
		return
	}

	hashedPassword, err := auth.HashPassword(user.Password)
	if err != nil {
		log.Printf("Error hashing password: %s", err)
		respondWithError(w, http.StatusInternalServerError, "Could not hash password")
		return
	}

	params := database.CreateUserParams{Email: user.Email, HashedPassword: hashedPassword}
	dbUser, err := cfg.db.CreateUser(req.Context(), params)
	if err != nil {
		log.Printf("Error creating user: %s", err)
		respondWithError(w, http.StatusInternalServerError, "Could not create user")
	}

	user.ID = dbUser.ID
	user.CreatedAt = dbUser.CreatedAt
	user.UpdatedAt = dbUser.UpdatedAt
	user.Password = ""

	respondWithJSON(w, http.StatusCreated, user)
}
