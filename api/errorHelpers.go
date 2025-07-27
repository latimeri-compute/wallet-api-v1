package main

import (
	"net/http"
)

// запись ошибки в лог, метода и uri запроса
func (app *application) logError(r *http.Request, err error) {
	var (
		method = r.Method
		uri    = r.URL.RequestURI()
	)
	app.logger.Error(err.Error(), "method", method, "uri", uri)
}

// отправляет ошибку в формате json клиенту
func (app *application) errorResponse(w http.ResponseWriter, r *http.Request, status int, message any) {
	err := app.writeJSON(w, status, JSONEnveloper{"ошибка": message}, nil)
	if err != nil {
		app.logError(r, err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// ответ если не найден запрашиваемый кошелёк
func (app *application) walletNotFoundResponse(w http.ResponseWriter, r *http.Request) {
	app.errorResponse(w, r, http.StatusNotFound, "запрашиваемый кошелёк не найден")
}

// записывает ошибку в лог и отправляет клиенту ответ о внутренней ошибке 500
func (app *application) internalErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.errorResponse(w, r, http.StatusInternalServerError, "сервер столкнулся с непредвиденной проблемой")
	app.logError(r, err)
}
