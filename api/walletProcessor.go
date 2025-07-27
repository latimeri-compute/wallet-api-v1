package main

import (
	"github.com/latimeri-compute/wallet-api-v1/internal/models"
)

// запрос к кошельку
type WalletRequest struct {
	Wallet    *models.Wallet
	Amount    float64             // сумма для изменения
	SeqNumber int64               // номер в очереди
	RespChan  chan walletResponse // канал для обработки ответов
}

// ответ от кошелька
type walletResponse struct {
	NewBalance float64 // баланс
	SeqNumber  int64   // номер в очереди
	Error      error   // ответ об ошибке
}

// обработчик запросов к изменению кошельков
func StartWalletProcessor(model models.WalletModelInterface) chan<- WalletRequest {
	requestChan := make(chan WalletRequest, 10000) // увеличенное количество на случай скачков

	go func() {
		// упорядоченные реквесты для каждого кошелька
		wallets := make(map[string][]WalletRequest)
		var currentSeq int64 = 1

		for req := range requestChan {
			// добавление реквестов
			wallets[req.Wallet.ID] = append(wallets[req.Wallet.ID], req)

			// обработка следующего в очереди запроса (в случае его наличия)
			for len(wallets[req.Wallet.ID]) > 0 && wallets[req.Wallet.ID][0].SeqNumber == currentSeq {

				// достаём запрос и убираем его из очереди
				requestToProcess := wallets[req.Wallet.ID][0]
				wallets[req.Wallet.ID] = wallets[req.Wallet.ID][1:]

				err := model.ChangeBalance(req.Wallet, requestToProcess.Amount)

				// отправка ответа
				requestToProcess.RespChan <- walletResponse{
					NewBalance: req.Wallet.Balance,
					SeqNumber:  currentSeq,
					Error:      err,
				}
				close(requestToProcess.RespChan)
				currentSeq++
			}
		}
	}()

	return requestChan
}
