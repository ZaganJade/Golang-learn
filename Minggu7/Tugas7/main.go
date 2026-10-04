package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"backend-golang/Minggu7/Tugas7/app/repository"
	"backend-golang/Minggu7/Tugas7/app/service"
	"backend-golang/Minggu7/Tugas7/config"
	"backend-golang/Minggu7/Tugas7/database"
	"backend-golang/Minggu7/Tugas7/helper"
	"backend-golang/Minggu7/Tugas7/route"
)

const minSecretLength = 32

func main() {
	config.LoadEnv()
	logger := config.NewLogger()

	jwtSecret := config.GetEnv("JWT_SECRET", "")
	if len(jwtSecret) < minSecretLength {
		logger.Error("JWT_SECRET tidak diisi atau terlalu pendek",
			slog.Int("minimal_karakter", minSecretLength))
		os.Exit(1)
	}

	pool, err := database.NewPool(context.Background())
	if err != nil {
		logger.Error("gagal terhubung ke database", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer pool.Close()

	jwtManager := helper.NewJWTManager(
		jwtSecret,
		config.GetEnv("JWT_ISSUER", "praktikum-backend"),
		time.Duration(config.GetEnvInt("JWT_ACCESS_TTL_MINUTES", 15))*time.Minute,
	)
	userRepository := repository.NewUserRepository(pool)
	tokenRepository := repository.NewTokenRepository(pool)
	studentRepository := repository.NewStudentRepository(pool)
	roleRepository := repository.NewRoleRepository(pool)

	rawPermissions, err := roleRepository.LoadPermissions(context.Background())
	if err != nil {
		logger.Error("gagal memuat permission", slog.String("error", err.Error()))
		os.Exit(1)
	}
	permissions := helper.NewPermissionSet(rawPermissions)
	logger.Info("permission dimuat", slog.Any("roles", permissions.KnownRoles()))

	studentService := service.NewStudentService(studentRepository, permissions)
	authService := service.NewAuthService(
		userRepository, tokenRepository, jwtManager, permissions,
		time.Duration(config.GetEnvInt("JWT_REFRESH_TTL_DAYS", 7))*24*time.Hour,
	)
	userService := service.NewUserService(userRepository, permissions)

	app := config.NewApp(logger, route.Dependencies{
		Pool:		pool,
		JWT:		jwtManager,
		Permissions:	permissions,
		StudentService:	studentService,
		AuthService:	authService,
		UserService:	userService,
	})
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
