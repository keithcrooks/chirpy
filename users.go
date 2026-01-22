package main

import (
	"database/sql"
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
	user, err := getUserFromRequest(req)
	if err != nil {
		log.Printf("Error getting user from request: %s", err)
		respondWithError(w, http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
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

func (cfg *apiConfig) handlerLoginUser(w http.ResponseWriter, req *http.Request) {
	user, err := getUserFromRequest(req)
	if err != nil {
		log.Printf("Error getting user from request: %s", err)
		respondWithError(w, http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
		return
	}

	dbUser, err := cfg.db.GetUserByEmail(req.Context(), user.Email)
	if err != nil {
		log.Printf("Error getting user from the database: %s", err)
		switch err {
		case sql.ErrNoRows:
			respondWithError(w, http.StatusUnauthorized, "Incorrect email or password")
		default:
			respondWithError(
				w,
				http.StatusInternalServerError,
				http.StatusText(http.StatusInternalServerError),
			)
		}
		return
	}

	passwordOk, err := auth.CheckPassword(user.Password, dbUser.HashedPassword)
	if err != nil {
		log.Printf("Error checking password: %s", err)
		respondWithError(
			w,
			http.StatusInternalServerError,
			http.StatusText(http.StatusInternalServerError),
		)
		return
	}

	if !passwordOk {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password")
		return
	}

	user.Password = ""

	respondWithJSON(w, http.StatusOK, user)
}

func getUserFromRequest(req *http.Request) (User, error) {
	var user User

	decoder := json.NewDecoder(req.Body)

	err := decoder.Decode(&user)

	return user, err
}
