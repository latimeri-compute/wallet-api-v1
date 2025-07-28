package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/joho/godotenv"
	"github.com/latimeri-compute/wallet-api-v1/internal/models"

	_ "github.com/lib/pq"
)

// конфигурация приложения
type config struct {
	port int
	dsn  string
}

// структура приложения
type application struct {
	logger               *slog.Logger                // логгер
	walletModel          models.WalletModelInterface // модель кошелька
	walletProcessorInput chan<- WalletRequest        // канал обработки запросов на изменение баланса
	walletMu             sync.Mutex
}

func main() {
	// инициализация логгера
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// загрузка переменных среды из *.env файлов
	err := godotenv.Load("../config.env")
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	var cfg config

	// считывание флажков
	flag.IntVar(&cfg.port, "port", 8080, "порт сервера API")
	flag.StringVar(&cfg.dsn, "dsn", fmt.Sprintf("host=localhost user=%s password=%s dbname=%s port=5432 sslmode=disable TimeZone=Europe/Moscow", os.Getenv("POSTGRES_USER"), os.Getenv("POSTGRES_PASSWORD"), os.Getenv("POSTGRES_DB")), "PostgeSQL connection string")
	flag.Parse()

	// подключение к дб
	db, err := OpenDB(cfg.dsn)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	logger.Info("соединение с базой данных установлено")

	walletModel := models.NewWalletModel(db)
	app := &application{
		logger:      logger,
		walletModel: walletModel,
	}

	// канал обработки запросов на изменение баланса
	app.walletProcessorInput = app.StartWalletProcessor(walletModel)

	srv := &http.Server{
		Addr:        fmt.Sprintf(":%d", cfg.port),
		Handler:     app.routes(),
		IdleTimeout: 5 * time.Second,
		// SetConnMaxIdleTime: 5 * time.Minute,
	}

	logger.Info("запуск сервера", "addr", srv.Addr)
	err = srv.ListenAndServe()

	logger.Error(err.Error())
	os.Exit(1)

}

func OpenDB(dsn string) (*sql.DB, error) {
	// открытие пула подключения
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	// настройка максимального количества подключений
	db.SetMaxOpenConns(10)
	db.SetConnMaxIdleTime(time.Minute * 15)

	// контекст с 5 секундами таймаута
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	// пинг с контекстом
	// если соединение не будет установлено в указанный промежуток времени - возвращает ошибку
	err = db.PingContext(ctx)
	if err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}
