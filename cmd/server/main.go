package main

import (
	"context"
	"go.uber.org/zap"
	"net/http"
	"orderHub_L0/internal/cache"
	"orderHub_L0/internal/config"
	"orderHub_L0/internal/consumer"
	"orderHub_L0/internal/db"
	"orderHub_L0/internal/handler"
	"orderHub_L0/internal/service"
)

func main() {
	//инициализация логера
	logger, _ := zap.NewProduction()
	defer logger.Sync()
	//загрузка конфига с .env
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Fatal("failed to load config", zap.Error(err))
	}
	//инициализация базы данных
	db, err := db.NewDB(cfg.PostgresURL, logger)
	if err != nil {
		logger.Fatal("faild to initialaze db", zap.Error(err))
	}
	//инициализации кэша
	cache := cache.NewCache()

	//инициализации сервиса
	svc := service.NewService(db, cache, logger)
	if err := svc.LoadCache(context.Background()); err != nil {
		logger.Fatal("Failed to load cache", zap.Error(err))
	}

	//инициализации консумера запуск в горутине
	cons := consumer.NewConsumer(db, cache, cfg.KafkaBrokers, cfg.KafkaTopic, logger)
	go cons.Start(context.Background())

	//инициализация хендлеров и регистрация routes
	handler := handler.NewHandler(svc, logger)
	router := handler.RegisterRoutes()

	//запуска http сервера
	logger.Info("starting http server", zap.String("port", cfg.HTTPPort))
	server := &http.Server{Addr: ":" + cfg.HTTPPort, Handler: router}
	if err := server.ListenAndServe(); err != nil {
		logger.Fatal("failed to start Http server", zap.Error(err))
	}
}
