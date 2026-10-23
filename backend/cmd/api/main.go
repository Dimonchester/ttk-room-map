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
	"github.com/Dimonchester/ttk-room-map/backend/internal/middleware"
	"github.com/Dimonchester/ttk-room-map/backend/internal/user"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("warning: .env not found, using system env")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	ctx := context.Background()

	pool, err := database.NewPool(ctx)
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}
	defer pool.Close()
	log.Println("connected to PostgreSQL")

	migrateCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := database.Migrate(migrateCtx, pool); err != nil {
		log.Fatalf("database migration failed: %v", err)
	}
	log.Println("migrations applied")

	userRepo := user.NewRepository(pool)
	tokenManager := auth.NewTokenManager(
		os.Getenv("JWT_SECRET"),
		15*time.Minute,
		7*24*time.Hour,
	)
	authHandler := auth.NewHandler(userRepo, tokenManager)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		pingCtx, pingCancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer pingCancel()

		if err := pool.Ping(pingCtx); err != nil {
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

	authHandler.RegisterRoutes(mux)

	handler := middleware.CORS(mux)

	log.Printf("listening on :%s", port)
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatal(err)
	}
}
