package main

import (
	"net/http"
)

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET api/v1/wallets/{WALLET_UUID}", app.showWallet)
	mux.HandleFunc("POST api/v1/wallet", app.changeWalletBalance)

	return mux
}
