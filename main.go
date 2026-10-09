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
		if !validURL(req.URL) {
			writeError(w, http.StatusBadRequest, "url must be a valid http or https URL")
			return
		}

		code, err := svc.Shorten(req.URL)

		if err != nil {
			log.Println("shorten:", err)
			writeError(w, http.StatusInternalServerError, "could not create url")
			return
		}
		writeJSON(w, http.StatusCreated, shortenResponse{ShortURL: baseURL + "/" + code})
	})

	mux.HandleFunc("GET/{code}", func(w http.ResponseWriter, r *http.Request) {
		code := r.PathValue("code")
		longURL, err := svc.Resolve(code)

		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "short url not found")
			return
		}

		if err != nil {
			log.Println("resolve:", err)
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}
		http.Redirect(w, r, longURL, http.StatusFound) // 302
	})

	log.Printf("listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}

func validURL(raw string) bool {
	u, err := url.ParseRequestURI(raw)
	return err == nil && (u.Scheme == "http" || u.Scheme == "https") && u.Host != ""
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
