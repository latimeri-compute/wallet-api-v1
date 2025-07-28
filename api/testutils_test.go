package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/latimeri-compute/wallet-api-v1/internal/models"
	"github.com/latimeri-compute/wallet-api-v1/internal/models/mocks"
)

// создание тестового приложения
func newTestApplication(t *testing.T) *application {

	return &application{
		logger:      slog.New(slog.NewTextHandler(io.Discard, nil)),
		walletModel: &mocks.MockWalletModel{},
	}
}

// полноценное тестовое приложение с базой данных
func newTestApplicationWithDB(t *testing.T) *application {
	t.Helper()

	db := newSQLTestDB(t)
	m := models.NewWalletModel(db)
	app := &application{
		logger:      slog.New(slog.NewTextHandler(io.Discard, nil)),
		walletModel: m,
	}
	app.walletProcessorInput = app.StartWalletProcessor(m)
	return app
}

// создание тестового сервера
func newTestServer(t *testing.T, h http.Handler) *httptest.Server {
	srv := httptest.NewServer(h)
	srv.Config.IdleTimeout = 5 * time.Second
	return srv
}

// отправляет POST запрос json на /api/v1/wallet,
// возвращает только тело ответа
func sendRequest(t *testing.T, ts *httptest.Server, jsonBody []byte) []byte {
	t.Helper()
	resp, err := http.Post(ts.URL+"/api/v1/wallet", "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	b = bytes.TrimSpace(b)
	return b
}

func newSQLTestDB(t *testing.T) *sql.DB {
	dsn := fmt.Sprintf("host=db user=%s password=%s dbname=%s port=5432 sslmode=disable TimeZone=Europe/Moscow", os.Getenv("POSTGRES_USER"), os.Getenv("POSTGRES_PASSWORD"), os.Getenv("POSTGRES_DB_TEST"))
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}

	script, err := os.ReadFile("../internal/models/testdata/setup.sql")
	if err != nil {
		db.Close()
		t.Fatal(err)
	}
	_, err = db.Exec(string(script))
	if err != nil {
		db.Close()
		t.Fatal(err)
	}

	t.Cleanup(func() {
		defer db.Close()
		script, err := os.ReadFile("../internal/models/testdata/teardown.sql")
		if err != nil {
			t.Fatal(err)
		}
		_, err = db.Exec(string(script))
		if err != nil {
			t.Fatal(err)
		}
	})

	return db
}
