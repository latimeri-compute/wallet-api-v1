package main

import (
	"fmt"
	"sync"

	"github.com/latimeri-compute/wallet-api-v1/internal/models"
)

// запрос к кошельку
type WalletRequest struct {
	Wallet    models.Wallet
	Amount    float64             // сумма для изменения
	SeqNumber int64               // номер в очереди
	RespChan  chan walletResponse // канал для обработки ответов
}

type walletQueue struct {
	mu     sync.Mutex
	queues map[string][]WalletRequest
}

// ответ от кошелька
type walletResponse struct {
	NewBalance float64 // баланс
	SeqNumber  int64   // номер в очереди
	Error      error   // ответ об ошибке
}

// обработчик запросов к изменению кошельков
func (app *application) StartWalletProcessor(model models.WalletModelInterface) chan<- WalletRequest {
	requestChan := make(chan WalletRequest, 2000) // увеличенное количество на случай скачков

	go func() {
		// упорядоченные реквесты для каждого кошелька
		wallets := walletQueue{
			queues: make(map[string][]WalletRequest),
		}

		for req := range requestChan {
			func() {
				defer func() {
					if r := recover(); r != nil {
						app.logger.Error("паника во время обработки кошелька:", req.Wallet.ID, r)
						select {
						case req.RespChan <- walletResponse{Error: fmt.Errorf("внутренняя ошибка обработчика")}:
						default:
						}
					}
				}()

				// добавление реквестов в очередь
				wallets.add(req)

				// обработка всех готовых запросов для данного кошелька
				wallets.processReadyRequests(model)
			}()
		}
	}()

	return requestChan
}

func (wq *walletQueue) add(req WalletRequest) {
	wq.mu.Lock()
	defer wq.mu.Unlock()
	wq.queues[req.Wallet.ID] = append(wq.queues[req.Wallet.ID], req)
}

func (wq *walletQueue) processReadyRequests(model models.WalletModelInterface) {
	wq.mu.Lock()
	defer wq.mu.Unlock()

	// обработка запросов для каждого кошелька
	for walletId, queue := range wq.queues {
		if len(queue) == 0 {
			continue
		}

		// первый запрос из очереди
		requestToProcess := queue[0]
		wq.queues[walletId] = queue[1:]

		// обработка запроса
		err := model.ChangeBalance(&requestToProcess.Wallet, requestToProcess.Amount)
		newBalance := requestToProcess.Wallet.Balance

		// отправление ответа
		select {
		case requestToProcess.RespChan <- walletResponse{
			NewBalance: newBalance,
			SeqNumber:  requestToProcess.SeqNumber,
			Error:      err,
		}:
		default:
		}
	}
}
