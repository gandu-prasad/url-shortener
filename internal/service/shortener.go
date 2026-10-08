package service

import (
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"urlshortener/internal/store"
)

const (
	codeLength  = 8
	maxAttempts = 10
)

type Shortener struct {
	store *store.Store
}

func NewShortener(s *store.Store) *Shortener {
	return &Shortener{store: s}
}

func hashKey(longURL string, attempt int) string {
	input := longURL

	if attempt > 0 {
		input = fmt.Sprintf("%s#%d", longURL, attempt)
	}

	sum := sha256.Sum256([]byte(input))
	encoded := base64.RawURLEncoding.EncodeToString(sum[:])
	return encoded[:codeLength]
}

func (s *Shortener) Shorten(longURL string) (string, error) {
	// if url already shortened then return exisiting code
	if code, err := s.store.GetCodeByLongURL(longURL); err == nil {
		return code, nil
	} else if !errors.Is(err, store.ErrNotFound) {
		return "", err
	}

	for attempt := range maxAttempts {
		code := hashKey(longURL, attempt)
		err := s.store.Insert(code, longURL)

		if err == nil {
			return code, nil
		}

		if !store.IsUniqueViolation(err) {
			return "", err
		}
	}
	return "", errors.New("Could not generate a unique code")

}

func (s *Shortener) Resolve(code string) (string, error) {
	longURL, err := s.store.GetLongURL(code)

	if err != nil {
		return "", err
	}

	_ = s.store.IncrementClicks(code)
	return longURL, nil
}
