package mocks

import (
	"github.com/latimeri-compute/wallet-api-v1/internal/models"
)

var (
	wallet_one = &models.Wallet{
		ID:      "81a4c5c8-0085-45c1-9c44-d05912276715",
		Balance: 1000,
	}
	wallet_two = &models.Wallet{
		ID:      "a10c1759-ba1a-47a9-86f7-de80387fc3d4",
		Balance: 33.67,
	}
)

type MockWalletModel struct{}

func (m *MockWalletModel) ChangeBalance(wallet *models.Wallet, amount float64) error {
	err := m.GetOne(wallet)
	if err != nil {
		return err
	}

	if wallet.Balance < -amount {
		return models.ErrInsufficientBalance
	}
	wallet.Balance = amount + wallet.Balance
	return nil
}

func (m *MockWalletModel) GetOne(wallet *models.Wallet) error {
	switch wallet.ID {
	case wallet_one.ID:
		wallet.Balance = wallet_one.Balance
		return nil
	case wallet_two.ID:
		wallet.Balance = wallet_two.Balance
		return nil
	default:
		return models.ErrNotFound
	}
}
