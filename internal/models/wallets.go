package models

import "database/sql"

type WalletModelInterface interface {
	Create(wallet *Wallet) error
	GetOne(wallet *Wallet) error
}

type WalletModel struct {
	db *sql.DB
}

type Wallet struct {
	ID      int
	Balance float64
}

func NewWalletModel(db *sql.DB) *WalletModel {
	return &WalletModel{
		db: db,
	}
}

func (m *WalletModel) Create(wallet *Wallet) error {
	// TODO
	return nil
}

func (m *WalletModel) GetOne(wallet *Wallet) error {
	// TODO
	return nil
}
