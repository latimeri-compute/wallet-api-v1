package main

import (
	"encoding/json"
	"net/http"
)

type JSONEnveloper map[string]any

// записывает данные в формат JSON и отправляет клиенту
func (app *application) writeJSON(w http.ResponseWriter, status int, data any, headers http.Header) error {
	js, err := json.Marshal(data)
	if err != nil {
		return err
	}

	// записывает хедеры
	for key, value := range headers {
		w.Header()[key] = value
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(js)

	return nil
}
