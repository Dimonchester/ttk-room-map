package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Dimonchester/ttk-room-map/backend/internal/auth"
	"github.com/Dimonchester/ttk-room-map/backend/internal/database"
	"github.com/Dimonchester/ttk-room-map/backend/internal/logger"
	"github.com/Dimonchester/ttk-room-map/backend/internal/middleware"
	"github.com/Dimonchester/ttk-room-map/backend/internal/space"
	"github.com/Dimonchester/ttk-room-map/backend/internal/user"
	"github.com/joho/godotenv"
)

func main() {
	if err := logger.Init("logs"); err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	}
	defer logger.Close()

	if err := godotenv.Load(); err != nil {
		logger.Warn(logger.CatSystem, "файл .env не найден, используются системные переменные окружения")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	logger.Info(logger.CatSystem, "запуск backend-сервера на порту :%s", port)

	ctx := context.Background()

	pool, err := database.NewPool(ctx)
	if err != nil {
		logger.Fatal(logger.CatDB, "критическая ошибка подключения к БД: %v", err)
	}
	defer pool.Close()

	migrateCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := database.Migrate(migrateCtx, pool); err != nil {
		logger.Fatal(logger.CatDB, "критическая ошибка миграции БД: %v", err)
	}

	mediaProcessor, err := space.NewMediaProcessor("uploads")
	if err != nil {
		logger.Fatal(logger.CatSystem, "ошибка инициализации каталога загрузок: %v", err)
	}

	userRepo := user.NewRepository(pool)
	tokenManager := auth.NewTokenManager(
		os.Getenv("JWT_SECRET"),
		15*time.Minute,
		7*24*time.Hour,
	)
	authHandler := auth.NewHandler(userRepo, tokenManager)

	spaceRepo := space.NewRepository(pool)
	spaceHandler := space.NewHandler(spaceRepo, mediaProcessor, authHandler)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		pingCtx, pingCancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer pingCancel()

		if err := pool.Ping(pingCtx); err != nil {
			logger.Error(logger.CatDB, "проверка /health не пройдена: БД недоступна: %v", err)
			w.WriteHeader(http.StatusServiceUnavailable)
			fmt.Fprintf(w, `{"status":"error","db":"%s"}`, err.Error())
			return
		}

		fmt.Fprint(w, `{"status":"ok","db":"ok"}`)
	})

	mux.HandleFunc("GET /api/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"status":  "ok",
			"message": "pong from backend",
		})
	})

	mux.Handle("GET /uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir("uploads"))))

	authHandler.RegisterRoutes(mux)
	spaceHandler.RegisterRoutes(mux)

	handler := middleware.CORS(middleware.RequestLogger(mux))

	logger.Info(logger.CatSystem, "сервер готов и слушает порт :%s", port)
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		logger.Fatal(logger.CatSystem, "остановка HTTP-сервера с ошибкой: %v", err)
	}
}
