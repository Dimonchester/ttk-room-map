package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/Dimonchester/ttk-room-map/backend/internal/database"
	"github.com/Dimonchester/ttk-room-map/backend/internal/user"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	pool, err := database.NewPool(ctx)
	if err != nil {
		log.Fatalf("connect error: %v", err)
	}
	defer pool.Close()

	if err := database.Migrate(ctx, pool); err != nil {
		log.Fatalf("migrate error: %v", err)
	}

	repo := user.NewRepository(pool)
	testEmail := fmt.Sprintf("verify_%d@ttk.local", time.Now().UnixNano())

	created, err := repo.Create(ctx, user.CreateUserInput{
		Email:        testEmail,
		PasswordHash: "hashed_secret_123",
		Role:         user.RoleOrganizer,
	})
	if err != nil {
		log.Fatalf("create user failed: %v", err)
	}
	defer func() {
		_, _ = pool.Exec(context.Background(), "DELETE FROM users WHERE id = $1::uuid", created.ID)
	}()

	fetched, err := repo.GetByEmail(ctx, testEmail)
	if err != nil {
		log.Fatalf("get by email failed: %v", err)
	}

	_, err = repo.Create(ctx, user.CreateUserInput{
		Email:        testEmail,
		PasswordHash: "another_hash",
		Role:         user.RoleVisitor,
	})
	if !errors.Is(err, user.ErrEmailTaken) {
		log.Fatalf("expected ErrEmailTaken, got: %v", err)
	}

	_, err = repo.Create(ctx, user.CreateUserInput{
		Email:        "invalid_role_" + testEmail,
		PasswordHash: "hash",
		Role:         user.Role("admin"),
	})
	if !errors.Is(err, user.ErrInvalidRole) {
		log.Fatalf("expected ErrInvalidRole, got: %v", err)
	}

	fmt.Printf("OK: user created & verified (id=%s, email=%s, role=%s, created_at=%s)\n",
		fetched.ID, fetched.Email, fetched.Role, fetched.CreatedAt.Format(time.RFC3339))
}
