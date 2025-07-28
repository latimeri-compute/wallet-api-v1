package main

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/latimeri-compute/wallet-api-v1/internal/assert"
	"github.com/latimeri-compute/wallet-api-v1/internal/models"
	vegeta "github.com/tsenart/vegeta/lib"
)

func TestManyChangeWalletBalance(t *testing.T) {
	if testing.Short() {
		t.Skip("пропуск атаки запросами")
	}

	rate := vegeta.Rate{
		Freq: 1000,
		Per:  time.Second,
	}

	tests := []struct {
		name          string
		walletRequest walletRequestJSON
		wantBalance   float64
	}{
		{
			name: "пополнение атака 1000 rps",
			walletRequest: walletRequestJSON{
				ID:            "81a4c5c8-0085-45c1-9c44-d05912276715",
				OperationType: "deposit",
				Amount:        1,
			},
			wantBalance: 3000,
		},
		{
			name: "снятие атака 1000 rps",
			walletRequest: walletRequestJSON{
				ID:            "81a4c5c8-0085-45c1-9c44-d05912276715",
				OperationType: "withdraw",
				Amount:        100,
			},
			wantBalance: 0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			app := newTestApplicationWithDB(t)
			ts := newTestServer(t, app.routes())
			defer ts.Close()

			JSON, err := json.Marshal(test.walletRequest)
			if err != nil {
				t.Fatal(err)
			}

			duration := 2 * time.Second
			targeter := vegeta.NewStaticTargeter(vegeta.Target{
				Method: "POST",
				URL:    ts.URL + "/api/v1/wallet",
				Body:   JSON,
			})
			attacker := vegeta.NewAttacker()
			for range attacker.Attack(
				targeter, rate, duration, "json") {

			}

			wallet := models.Wallet{ID: test.walletRequest.ID}
			err = app.walletModel.GetOne(&wallet)
			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, wallet.Balance, test.wantBalance)
			ts.CloseClientConnections()
			ts.Close()
		})
	}
}
