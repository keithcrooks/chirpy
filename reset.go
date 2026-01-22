package main

import "net/http"

func (cfg *apiConfig) handlerReset(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	cfg.fileserverHits.Store(0)

	if err := cfg.db.DeleteAllUsers(req.Context()); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error deleting all users"))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Counter reset to 0"))
}
