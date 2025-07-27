package main

import (
	"errors"
	"net/http"
	"strings"
	"sync/atomic"

	"github.com/latimeri-compute/wallet-api-v1/internal/models"
	"github.com/latimeri-compute/wallet-api-v1/internal/models/validator"
)

var sequenceCounter int64

// GET api/v1/wallet/{WALLET_UUID}
func (app *application) showWallet(w http.ResponseWriter, r *http.Request) {
	// получение id из ссылки
	id := r.PathValue("WALLET_UUID")

	// ищем заданный кошелёк
	wallet := &models.Wallet{
		ID: id,
	}
	err := app.walletModel.GetOne(wallet)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			app.walletNotFoundResponse(w, r)
		} else {
			app.internalErrorResponse(w, r, err)
		}
		return
	}

	// ответ сервера с кошельком
	err = app.writeJSON(w, http.StatusOK, JSONEnveloper{"кошелёк": wallet}, nil)
	if err != nil {
		app.internalErrorResponse(w, r, err)
	}
}

// POST /api/v1/wallet
func (app *application) changeWalletBalance(w http.ResponseWriter, r *http.Request) {
	var unpackedJSON struct {
		ID            string  `json:"walletId"`
		OperationType string  `json:"operationType"`
		Amount        float64 `json:"amount"`
	}

	err := app.unpackJSON(w, r, &unpackedJSON)
	if err != nil {
		app.errorResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}

	// проверка полей полученного json
	operationType := strings.ToUpper(unpackedJSON.OperationType)
	val := validator.NewValidator()
	val.Check(len(unpackedJSON.ID) != 0, "walletId", "значение поля не может быть пустым")
	val.Check(unpackedJSON.Amount != 0, "amount", "значение поля не может быть пустым")
	val.Check(unpackedJSON.Amount > 0, "amount", "значение поля должно быть больше нуля")
	val.Check(len(unpackedJSON.OperationType) != 0, "operationType", "значение поля не может быть пустым")
	val.Check(validator.IsPermittedValue(operationType, []string{"DEPOSIT", "WITHDRAW"}...), "operationType", "несуществующий тип операции. доступные операции: DEPOSIT, WITHDRAW")

	if val.Valid() != true {
		app.errorResponse(w, r, http.StatusUnprocessableEntity, val.Errors)
		return
	}

	// проверяем существование кошелька, заодно получаем его баланс
	wallet := &models.Wallet{ID: unpackedJSON.ID}
	err = app.walletModel.GetOne(wallet)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			app.walletNotFoundResponse(w, r)
		} else {
			app.internalErrorResponse(w, r, err)
		}
		return
	}

	// если пользователь снимает деньги, то сумма для изменения становится отрицательной
	if operationType == "WITHDRAW" {
		unpackedJSON.Amount = -unpackedJSON.Amount
	}

	// создание канала для получения ответа об обработке кошелька
	respChan := make(chan walletResponse, 1)
	seqNumber := atomic.AddInt64(&sequenceCounter, 1)

	// отправка в очередь на обработку запроса
	app.walletProcessorInput <- WalletRequest{
		Wallet:    wallet,
		Amount:    unpackedJSON.Amount,
		SeqNumber: seqNumber,
		RespChan:  respChan,
	}

	// получение ответа
	resp := <-respChan

	switch {
	case errors.Is(resp.Error, models.ErrInsufficientBalance):
		app.errorResponse(w, r, http.StatusBadRequest, resp.Error.Error())
		return
	case errors.Is(resp.Error, models.ErrNotFound):
		app.walletNotFoundResponse(w, r)
		return
	case errors.Is(resp.Error, nil):
		break
	default:
		app.internalErrorResponse(w, r, err)
		return
	}

	// ответ клиенту
	err = app.writeJSON(w, http.StatusOK, JSONEnveloper{"кошелёк": wallet}, nil)
	if err != nil {
		app.internalErrorResponse(w, r, err)
	}

}
