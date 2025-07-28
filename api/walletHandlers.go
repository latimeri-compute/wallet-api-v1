package main

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"github.com/latimeri-compute/wallet-api-v1/internal/models"
	"github.com/latimeri-compute/wallet-api-v1/internal/validator"
)

var sequenceCounter int64

type walletRequestJSON struct {
	ID            string  `json:"walletId"`
	OperationType string  `json:"operationType"`
	Amount        float64 `json:"amount"`
}

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

	var walletRequestJSON walletRequestJSON
	err := app.unpackJSON(w, r, &walletRequestJSON)
	if err != nil {
		app.errorResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}

	// проверка полей полученного json
	operationType := strings.ToUpper(walletRequestJSON.OperationType)
	val := validator.NewValidator()
	val.Check(len(walletRequestJSON.ID) != 0, "walletId", "значение поля не может быть пустым")
	val.Check(walletRequestJSON.Amount != 0, "amount", "значение поля не может быть пустым")
	val.Check(walletRequestJSON.Amount > 0, "amount", "значение поля должно быть больше нуля")
	val.Check(len(walletRequestJSON.OperationType) != 0, "operationType", "значение поля не может быть пустым")
	val.Check(validator.IsPermittedValue(operationType, []string{"DEPOSIT", "WITHDRAW"}...), "operationType", "несуществующий тип операции. доступные операции: DEPOSIT, WITHDRAW")

	if val.Valid() != true {
		app.errorResponse(w, r, http.StatusUnprocessableEntity, val.Errors)
		return
	}

	wallet := models.Wallet{
		ID: walletRequestJSON.ID,
	}
	// если пользователь снимает деньги, то сумма для изменения становится отрицательной
	if operationType == "WITHDRAW" {
		walletRequestJSON.Amount = -walletRequestJSON.Amount
	}

	// создание канала для получения ответа об обработке кошелька
	respChan := make(chan walletResponse, 1)

	seqNumber := atomic.AddInt64(&sequenceCounter, 1)

	// отправка в очередь на обработку запроса
	select {
	case app.walletProcessorInput <- WalletRequest{
		Wallet:    wallet,
		Amount:    walletRequestJSON.Amount,
		SeqNumber: seqNumber,
		RespChan:  respChan,
	}:
		// запрос отправлен успешно
	case <-time.After(500 * time.Millisecond):
		app.errorResponse(w, r, http.StatusServiceUnavailable, "сервер перегружен")
		return
	case <-r.Context().Done():
		app.errorResponse(w, r, http.StatusRequestTimeout, "запрос отменён")
		return
	}

	// получение ответа
	select {
	case <-time.After(50 * time.Second):
		app.errorResponse(w, r, http.StatusGatewayTimeout, "таймаут операции")
		return
	case <-r.Context().Done():
		app.errorResponse(w, r, http.StatusRequestTimeout, "запрос отменён")
		return
	case resp, ok := <-respChan:
		if !ok {
			app.internalErrorResponse(w, r, fmt.Errorf("канал ответа закрыт"))
			return
		}

		switch {
		case errors.Is(resp.Error, models.ErrInsufficientBalance):
			app.errorResponse(w, r, http.StatusBadRequest, resp.Error.Error())
			return
		case errors.Is(resp.Error, models.ErrNotFound):
			app.walletNotFoundResponse(w, r)
			return
		case resp.Error != nil:
			app.internalErrorResponse(w, r, resp.Error)
		default:
			wallet = models.Wallet{
				ID:      wallet.ID,
				Balance: resp.NewBalance,
			}
			err = app.writeJSON(w, http.StatusOK, JSONEnveloper{"кошелёк": wallet}, nil)
			if err != nil {
				app.internalErrorResponse(w, r, err)
			}
		}
	}
}
