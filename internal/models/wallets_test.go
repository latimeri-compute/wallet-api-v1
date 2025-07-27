package models

import (
	"testing"

	"github.com/joho/godotenv"
	"github.com/latimeri-compute/wallet-api-v1/internal/models/assert"
)

func TestGetOne(t *testing.T) {
	if testing.Short() {
		t.Skip("models: пропуск проверки интеграции с базой данных")
	}

	tests := []struct {
		name       string
		wallet     *Wallet
		wantWallet Wallet
		wantErr    error
	}{
		{
			name: "существующий кошелёк",
			wallet: &Wallet{
				ID: "81a4c5c8-0085-45c1-9c44-d05912276715",
			},
			wantWallet: Wallet{
				ID:      "81a4c5c8-0085-45c1-9c44-d05912276715",
				Balance: 1000,
			},
			wantErr: nil,
		},
		{
			name: "несуществующий кошелёк",
			wallet: &Wallet{
				ID: "01757132-f89b-48c6-aa57-1e8c9b2999d3",
			},
			wantWallet: Wallet{
				ID: "01757132-f89b-48c6-aa57-1e8c9b2999d3",
			},
			wantErr: ErrNotFound,
		},
		{
			name: "неверный формат UUID",
			wallet: &Wallet{
				ID: ":)",
			},
			wantWallet: Wallet{
				ID: ":)",
			},
			wantErr: ErrNotFound,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			db := newTestDB(t)
			m := WalletModel{db}

			err := m.GetOne(test.wallet)

			assert.Equal(t, test.wallet.Balance, test.wantWallet.Balance)
			assert.Equal(t, err, test.wantErr)
		})
	}
}

func TestChangeWalletBalance(t *testing.T) {
	if testing.Short() {
		t.Skip("models: пропуск проверки интеграции с базой данных")
	}
	err := godotenv.Load("../../config.env")
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name        string
		wallet      *Wallet
		amount      float64
		wantBalance float64
		wantErr     error
	}{
		{
			name: "существующий кошелёк",
			wallet: &Wallet{
				ID: "81a4c5c8-0085-45c1-9c44-d05912276715",
			},
			amount:      100,
			wantBalance: 1100,
			wantErr:     nil,
		},
		{
			name: "несуществующий кошелёк",
			wallet: &Wallet{
				ID: "01757132-f89b-48c6-aa57-1e8c9b2999d3",
			},
			amount:      100,
			wantBalance: 0,
			wantErr:     ErrNotFound,
		},
		{
			name: "неверный формат UUID",
			wallet: &Wallet{
				ID: ":)",
			},
			wantBalance: 0,
			wantErr:     ErrNotFound,
		},
		{
			name: "снятие суммы больше баланса на кошельке",
			wallet: &Wallet{
				ID:      "81a4c5c8-0085-45c1-9c44-d05912276715",
				Balance: 1000,
			},
			amount:      -9000,
			wantBalance: 1000,
			wantErr:     ErrInsufficientBalance,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			db := newTestDB(t)
			m := WalletModel{db}

			err := m.ChangeBalance(test.wallet, test.amount)

			assert.Equal(t, err, test.wantErr)
			assert.Equal(t, test.wallet.Balance, test.wantBalance)
		})
	}
}
