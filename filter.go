package main

import "strings"

func filterChirp(chirp Chirp) CleanedChirp {
	const (
		replacement = "****"
		sep         = " "
	)

	var cleaned CleanedChirp

	words := strings.Split(chirp.Body, sep)

	for i, word := range words {
		if isProfane(word) {
			words[i] = replacement
		}
	}

	cleaned.CleanedBody = strings.Join(words, sep)

	return cleaned
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
