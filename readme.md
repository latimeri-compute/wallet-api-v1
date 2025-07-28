## Функционал
###### POST /api/wallet/</br>
формат отправляемого json:</br>
`{
walletId: UUID,
operationType: DEPOSIT or WITHDRAW,
amount: 33.67
}`
###### GET /api/v1/wallets/{WALLET_UUID}
Получение баланса кошелька по указанному UUID

## переменные среды
указываются в файле config.env
### Докер:
POSTGRES_USER=пользователь postgres</br>
POSTGRES_PASSWORD=пароль к пользователю</br>
POSTGRES_DB=название базы данных</br>
### Бэкенд:
берёт те же переменные, что и докер
### Тесты:
POSTGRES_DB_TEST=название бд для тестов моделей
