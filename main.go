package main

import (
	"github.com/sirupsen/logrus"
	"net/http"
	"os"
	"os/signal"
	"simpleHttpRequest/database"
	"simpleHttpRequest/server"
	"syscall"
	"time"
)

const shutDownTimeOut = 10 * time.Second

func main() {
	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, os.Intrupt(), syscall.SIGTERM)
	srv := server.SetupRoutes()

	if err := database.ConnectAndMigrate(
		"localhost",
		"5433",
		"todo",
		"local",
		"local",
		database.SSLModeDisable); err != nil {
		logrus.Panicf("Failed to initialize and migrate database with error: %+v", err)
	}
	logrus.Print("migrations successful!!")

	go func() {
		if err := srv.Run(":8080"); err != nil && err != http.ErrServerClosed {
			logrus.Panicf("Failed to run server with error: %+v", err)
		}
	}()
	logrus.Print("Server started at :8080")

	<-done

	logrus.Info("shutting down server")
	if err := database.ShutdownDatabase(); err != nil {
		logrus.WithError(err).Error("failed to close database connection")
	}

	if err := srv.Shutdown(shutDownTimeOut); err != nil {
		logrus.WithError(err).Panic("failed to gracefully shutdown server")
	}

}
