package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Dimonchester/ttk-room-map/backend/internal/database"
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
	_ = userRepo

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
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

	log.Printf("listening on :%s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
