package models

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/lib/pq"
)

var (
	ErrNotFound            = errors.New("кошелёк не найден")
	ErrInsufficientBalance = errors.New("недостаточный баланс кошелька для совершения операции")
)

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
}

func NewWalletModel(db *sql.DB) *WalletModel {
	return &WalletModel{
		db: db,
	}
}

// изменяет баланс на указанную сумму, если при этом баланс не опускается ниже нуля
func (m *WalletModel) ChangeBalance(wallet *Wallet, amount float64) error {
	err := m.GetOne(wallet)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `UPDATE wallets
	SET balance = balance + $1
	WHERE id = $2 AND (balance + $1) >= 0
	RETURNING balance`
	err = tx.QueryRowContext(ctx, query, amount, wallet.ID).Scan(&wallet.Balance)
	if err != nil {
		tx.Rollback()
		if errors.Is(err, sql.ErrNoRows) {
			return ErrInsufficientBalance
		}
		return err
	}

	err = tx.Commit()

	return err
}

// получает данные о кошельке по id и записывает их на переданный Wallet
func (m *WalletModel) GetOne(wallet *Wallet) error {
	query := `SELECT balance
	FROM wallets
	WHERE id = $1`

	args := []any{wallet.ID}
	err := m.db.QueryRow(query, args...).Scan(&wallet.Balance)

	if err != nil {
		// код 22P02 означает неверное формирование формата uuid
		if errors.Is(err, sql.ErrNoRows) || err.(*pq.Error).Code == "22P02" {
			return ErrNotFound
		}
		return err
	}

	return nil
}
