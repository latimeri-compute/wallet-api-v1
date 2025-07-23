package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/latimeri-compute/wallet-api-v1/internal/models"

	// go не предоставляет официальных драйверов бд, но рекомендует этот
	_ "github.com/lib/pq"
)

// конфигурация приложения
type config struct {
	port int
	dsn  string
}

// структура приложения
type application struct {
	logger      *slog.Logger
	walletModel *models.WalletModel
}

func main() {
	var cfg config

	// считывание флажков
	flag.IntVar(&cfg.port, "port", 8080, "порт сервера API")
	flag.StringVar(&cfg.dsn, "dsn", os.Getenv("DSN"), "PostgeSQL connection string")
	flag.Parse()

	// инициализация логгера
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
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

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.port),
		Handler: app.routes(),
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
	db.SetMaxOpenConns(20)
	db.SetConnMaxIdleTime(15)
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
