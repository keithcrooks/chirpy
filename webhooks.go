package main

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
)

type Webhook struct {
	Event string `json:"event"`
	Data  struct {
		UserID uuid.UUID `json:"user_id"`
	} `json:"data"`
}

func (cfg *apiConfig) handlerPolkaWebhook(w http.ResponseWriter, req *http.Request) {
	var webhook Webhook

	decoder := json.NewDecoder(req.Body)
	if err := decoder.Decode(&webhook); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid webhook")
		return
	}

	if webhook.Event != "user.upgraded" {
		respondWithJSON(w, http.StatusNoContent, nil)
		return
	}

	if err := cfg.db.UpgradeUserToChirpyRed(req.Context(), webhook.Data.UserID); err != nil {
		if err == sql.ErrNoRows {
			respondWithError(w, http.StatusNotFound, "User not found")
			return
		}

		respondWithError(w, http.StatusInternalServerError, "Could not upgrade user")
		return
	}

	respondWithJSON(w, http.StatusNoContent, nil)
}
