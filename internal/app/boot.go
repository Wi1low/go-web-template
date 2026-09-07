// Package app is the main package of the application.
// It contains the main function and the main loop.
// Including config、logger、gorm、gin components
package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
)

func Run() {
	// Reading config
	LoadConfig()

	// Logger
	initLogger(AppConfig.Server.Mode)

	// Gorm
	initGorm()

	// Gin
	g := initGin()

	go func() {
		initPrometheus()
	}()

	srv := http.Server{
		Addr:    AppConfig.Server.Addr,
		Handler: g,
	}
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		zap.L().Fatal("server failed to start", zap.Any("err", err))
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	fmt.Println("Block signal received, will shutdown server gracefully")
	<-quit
	fmt.Println("Shutting down server...")
	zap.L().Info("server will be shundown, now clearing resources")
	zap.L().Warn("server will be shundown, now clearing resources")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		zap.L().Fatal("server forced to shutdown", zap.Any("err", err))
	}

	zap.L().Info("server exiting")
}
