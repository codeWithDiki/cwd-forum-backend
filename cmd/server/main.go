package main

import (
	"context"
	"gin-quickstart/config"
	"gin-quickstart/internal/app"
	"gin-quickstart/pkg/logger"
	"gin-quickstart/pkg/worker"
	"gin-quickstart/routes"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	if err := config.LoadEnv(); err != nil {
		panic("failed to load env: " + err.Error())
	}

	db, err := config.InitDB()
	if err != nil {
		panic("failed to connect to database: " + err.Error())
	}

	redis, err := config.InitRedis()
	if err != nil {
		panic("failed to connect to Redis: " + err.Error())
	}

	isProduction := os.Getenv("LOG_IS_PRODUCTION") == "true"
	logLevel := os.Getenv("LOG_LEVEL")
	isDisableStackTrace := os.Getenv("DISABLE_STACK_TRACE") == "true"
	log := logger.NewLogger(isProduction, logLevel, isDisableStackTrace)
	defer log.Sync()
	appCtx := log.SetTraceID(context.Background())

	workerPool := worker.NewWorker(20)

	deps := app.Dependencies{
		DB:     db,
		Redis:  redis,
		Worker: workerPool,
		Logger: log,
	}

	r := routes.SetupRouter(deps)
	server := &http.Server{
		Addr:              ":8080",
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Info(
		appCtx,
		"starting http server",
		log.Field("addr", server.Addr),
		log.Field("shutdown_timeout", "10s"),
		log.Field("worker_count", 20),
	)

	serverErr := make(chan error, 1)
	go func() {
		serverErr <- server.ListenAndServe()
	}()

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(signalCh)

	select {
	case err := <-serverErr:
		if err != nil && err != http.ErrServerClosed {
			log.Fatal(
				appCtx,
				"server failed to start",
				log.Field("error", err),
			)
		}
	case sig := <-signalCh:
		shutdownStartedAt := time.Now()
		log.Info(
			appCtx,
			"shutdown signal received",
			log.Field("signal", sig.String()),
		)

		log.Info(
			appCtx,
			"starting graceful shutdown",
			log.Field("http_addr", server.Addr),
		)

		shutdownCtx, cancel := context.WithTimeout(appCtx, 10*time.Second)
		defer cancel()

		log.Info(appCtx, "shutting down http server")
		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Error(appCtx, "http server shutdown failed", err)
		} else {
			log.Info(appCtx, "http server shutdown completed")
		}

		log.Info(appCtx, "waiting for worker pool to stop")
		workerPool.Stop()
		log.Info(appCtx, "worker pool stopped")

		log.Info(appCtx, "closing redis client")
		if err := redis.Close(); err != nil {
			log.Error(appCtx, "redis shutdown failed", err)
		} else {
			log.Info(appCtx, "redis client closed")
		}

		log.Info(appCtx, "closing database connection")
		sqlDB, err := db.DB()
		if err != nil {
			log.Error(appCtx, "failed to get sql db", err)
		} else if err := sqlDB.Close(); err != nil {
			log.Error(appCtx, "database shutdown failed", err)
		} else {
			log.Info(appCtx, "database connection closed")
		}

		log.Info(
			appCtx,
			"graceful shutdown completed",
			log.Field("duration_ms", time.Since(shutdownStartedAt).Milliseconds()),
		)
	}
}
