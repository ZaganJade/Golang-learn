package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"backend-golang/Minggu4/Tugas4/app/repository"
	"backend-golang/Minggu4/Tugas4/app/service"
	"backend-golang/Minggu4/Tugas4/config"
	"backend-golang/Minggu4/Tugas4/database"
)

func main() {
	config.LoadEnv()
	logger := config.NewLogger()

	pool, err := database.NewPool(context.Background())
	if err != nil {
		logger.Error("gagal terhubung ke database", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer pool.Close()

	studentRepository := repository.NewStudentRepository(pool)
	studentService := service.NewStudentService(studentRepository)

	app := config.NewApp(logger, pool, studentService)
	port := config.GetEnv("APP_PORT", "3000")

	go func() {
		if err := app.Listen(":" + port); err != nil {
			logger.Error("server berhenti", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()
	logger.Info("server berjalan", slog.String("port", port))

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("sinyal berhenti diterima, menutup server")
	if err := app.ShutdownWithTimeout(10 * time.Second); err != nil {
		logger.Error("gagal menutup server dengan rapi",
			slog.String("error", err.Error()))
	}
	logger.Info("server berhenti dengan rapi")
}
