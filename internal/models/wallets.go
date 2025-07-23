package models

import (
	"database/sql"
	"errors"
)

var ErrNotFound = errors.New("кошелёк не найден")

type WalletModelInterface interface {
	ChangeBalance(wallet *Wallet, amount float64) error
	GetOne(wallet *Wallet) error
}

type WalletModel struct {
	db *sql.DB
}

type Wallet struct {
	ID      string  `json:"walletId"`
	Balance float64 `json:"balance"`
	// TODO добавить версии строчек?
}

func NewWalletModel(db *sql.DB) *WalletModel {
	return &WalletModel{
		db: db,
	}
}

func (m *WalletModel) ChangeBalance(wallet *Wallet, amount float64) error {
	// TODO добавить логику изменения баланса
	return nil
}

func (m *WalletModel) GetOne(wallet *Wallet) error {
	query := `SELECT id, balance
	FROM wallets
	WHERE id = $id`

	args := []any{wallet.ID}
	err := m.db.QueryRow(query, args...).Scan(
		&wallet.Balance,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}

	return nil
}
