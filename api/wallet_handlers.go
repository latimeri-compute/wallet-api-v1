package main

import (
	"errors"
	"net/http"

	"github.com/latimeri-compute/wallet-api-v1/internal/models"
)

func (app *application) showWallet(w http.ResponseWriter, r *http.Request) {
	// получение id из ссылки
	id := r.PathValue("WALLET_UUID")

	wallet := &models.Wallet{
		ID: id,
	}

	err := app.walletModel.GetOne(wallet)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			app.notFoundResponse(w, r)
			return
		}
		app.internalErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, JSONEnveloper{"кошелёк": wallet}, nil)
	if err != nil {
		app.internalErrorResponse(w, r, err)
	}
}

func (app *application) changeWalletBalance(w http.ResponseWriter, r *http.Request) {
	var unpackedJSON struct {
		ID            string  `json:"valletId"`
		OperationType string  `json:"operationType"`
		Amount        float64 `json:"amount"`
	}

	err := app.unpackJSON(w, r, &unpackedJSON)
	if err != nil {
		app.errorResponse(w, r, http.StatusBadRequest, err)
	}

	var wallet *models.Wallet
	err = app.walletModel.GetOne(wallet)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			app.notFoundResponse(w, r)
			return
		}
		app.internalErrorResponse(w, r, err)
		return
	}

	if unpackedJSON.OperationType != "DEPOSIT" && unpackedJSON.OperationType != "WITHDRAW" {
		app.errorResponse(w, r, http.StatusBadRequest, "неверный тип операции operationType")
		return
	}
	if unpackedJSON.Amount <= 0 {
		app.errorResponse(w, r, http.StatusBadRequest, "сумма не может быть меньше или равна нулю")
		return
	}

	// TODO закончить логику изменения баланса

}

// снятие
func (app *application) withdrawal(wallet *models.Wallet, amount float64) error {
	if wallet.Balance < amount {
		return errors.New("недостаточный баланс кошелька для совершения операции")
	}

	err := app.walletModel.ChangeBalance(wallet, -amount)
	if err != nil {
		return err
	}

	return nil

}
