package store

import (
	"database/sql"
	"errors"
	"strings"

	_ "modernc.org/sqlite"
)

var ErrNotFound = errors.New("URL Not Found")

type Store struct {
	db *sql.DB
}

func Open(path string) (*Store, error) {

	dsn := path + "?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)"
	db, err := sql.Open("sqlite", dsn)

	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(4)

	schema := `
	create table if not exists urls (
		id integer primary key autoincrement,
		short_code text not null unique,
		long_url text not null
		created_at text not null default (datetime('now')),
		click_count interger not null default 0
	);`

	if _, err := db.Exec(schema); err != nil {
		return nil, err
	}

	return &Store{db: db}, nil

}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) Insert(code, longURL string) error {
	_, err := s.db.Exec(
		`insert into urls (short_code, long_url) values (?, ?)`, code, longURL,
	)
	return err
}

func IsUniqueViolation(err error) bool {
	return err != nil && strings.Contains(err.Error(), "UNIQUE constraint failed")
}

func (s *Store) GetLongURL(code string) (string, error) {
	var longURL string

	err := s.db.QueryRow(
		`select long_url from urls where short_code = ?`, code,
	).Scan(&longURL)

	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrNotFound
	}

	return longURL, err
}

func (s *Store) GetCodeByLongURL(longURL string) (string, error) {
	var code string

	err := s.db.QueryRow(
		`select short_code from urls where longURL = ? limit 1`, longURL,
	).Scan(&code)

	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrNotFound
	}

	return code, err
}

func (s *Store) IncrementClicks(code string) error {
	_, err := s.db.Exec(
		`update urls set click_count = click_count + 1 where short_code = ?`, code,
	)

	return err
}
