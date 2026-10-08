package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/url"
	"os"

	"urlshortener/internal/service"
	"urlshortener/internal/store"
)

type shortenRequest struct {
	URL string `json:"url"`
}

type shortenResponse struct {
	ShortURL string `json:"short_url"`
}

func main() {

	dbPath := getEnv("DB_PATH", "urls.db")
	baseURL := getEnv("BASE_URL", "http://localhost:6969")
	addr := getEnv("ADDR", ":8080")

	st, err := store.Open(dbPath)
	if err != nil {
		log.Fatal(err)
	}

	defer st.Close()

	svc := service.NewShortener(st)
	mux := http.NewServeMux()

	mux.HandleFunc("POST /storten", func(w http.ResponseWriter, r *http.Request) {
		var req shortenRequest

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "Invalid json body")
			return
		}
	})
}
