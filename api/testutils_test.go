package main

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/latimeri-compute/wallet-api-v1/internal/models/mocks"
)

// создание тестового приложения
func newTestApplication(t *testing.T) *application {
	t.Helper()

	return &application{
		logger:      slog.New(slog.NewTextHandler(io.Discard, nil)),
		walletModel: &mocks.MockWalletModel{},
	}
}

// создание тестового сервера
func newTestServer(t *testing.T, h http.Handler) *httptest.Server {
	t.Helper()

	srv := httptest.NewServer(h)
	return srv
}

// проверяет, одинаковы ли значения, и в случае их отличия выбрасывает ошибку
func assertEqual[T comparable](t *testing.T, actual, expected T) {
	t.Helper()

	if actual != expected {
		t.Errorf("got: %v; want: %v", actual, expected)
	}
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
