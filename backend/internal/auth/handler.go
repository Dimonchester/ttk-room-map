package auth

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/mail"
	"strings"
	"time"

	"github.com/Dimonchester/ttk-room-map/backend/internal/user"
)

const refreshCookieName = "refresh_token"

type contextKey string

const userContextKey contextKey = "authenticated_user"

type FieldError struct {
	Field   string `json:"field"`
	Code    string `json:"code"`
	Message string `json:"message,omitempty"`
}

type Problem struct {
	Type    string       `json:"type"`
	Title   string       `json:"title"`
	Status  int          `json:"status"`
	Code    string       `json:"code"`
	Detail  string       `json:"detail,omitempty"`
	Message string       `json:"message,omitempty"`
	Fields  []FieldError `json:"fields,omitempty"`
}

type CredentialsRequest struct {
	Email    string    `json:"email"`
	Password string    `json:"password"`
	Role     user.Role `json:"role,omitempty"`
}

type AuthResponse struct {
	AccessToken string     `json:"accessToken"`
	ExpiresIn   int        `json:"expiresIn"`
	User        *user.User `json:"user"`
}

type Handler struct {
	users  *user.Repository
	tokens *TokenManager
}

func NewHandler(users *user.Repository, tokens *TokenManager) *Handler {
	return &Handler{
		users:  users,
		tokens: tokens,
	}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/auth/register", h.Register)
	mux.HandleFunc("POST /api/auth/login", h.Login)
	mux.HandleFunc("POST /api/auth/refresh", h.Refresh)
	mux.HandleFunc("POST /api/auth/logout", h.Logout)
	mux.Handle("GET /api/auth/me", h.RequireAuth(http.HandlerFunc(h.Me)))
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req CredentialsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeProblem(w, http.StatusBadRequest, "invalid_json", "Некорректный формат JSON", nil)
		return
	}

	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	if fields := validateCredentials(req.Email, req.Password); len(fields) > 0 {
		writeProblem(w, http.StatusUnprocessableEntity, "validation_error", "Ошибка валидации входных данных", fields)
		return
	}

	role := req.Role
	if role == "" {
		role = user.RoleOrganizer
	}

	createdUser, err := h.users.Create(r.Context(), user.CreateUserInput{
		Email:        req.Email,
		PasswordHash: req.Password,
		Role:         role,
	})
	if err != nil {
		if errors.Is(err, user.ErrEmailTaken) {
			writeProblem(w, http.StatusConflict, "email_already_exists", "Пользователь с таким email уже существует", []FieldError{
				{Field: "email", Code: "email_already_exists", Message: "Этот email уже занят"},
			})
			return
		}
		if errors.Is(err, user.ErrInvalidRole) {
			writeProblem(w, http.StatusUnprocessableEntity, "invalid_role", "Недопустимая роль пользователя", []FieldError{
				{Field: "role", Code: "invalid_role", Message: "Допустимые значения: organizer, visitor"},
			})
			return
		}
		writeProblem(w, http.StatusInternalServerError, "internal_error", "Не удалось создать пользователя", nil)
		return
	}

	h.issueSession(w, http.StatusCreated, createdUser)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req CredentialsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeProblem(w, http.StatusBadRequest, "invalid_json", "Некорректный формат JSON", nil)
		return
	}

	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	if fields := validateCredentials(req.Email, req.Password); len(fields) > 0 {
		writeProblem(w, http.StatusUnprocessableEntity, "validation_error", "Ошибка валидации входных данных", fields)
		return
	}

	u, err := h.users.GetByEmail(r.Context(), req.Email)
	if err != nil {
		if errors.Is(err, user.ErrNotFound) {
			writeProblem(w, http.StatusUnauthorized, "invalid_credentials", "Неверный email или пароль", nil)
			return
		}
		writeProblem(w, http.StatusInternalServerError, "internal_error", "Ошибка при проверке учетных данных", nil)
		return
	}

	if u.PasswordHash != req.Password {
		writeProblem(w, http.StatusUnauthorized, "invalid_credentials", "Неверный email или пароль", nil)
		return
	}

	h.issueSession(w, http.StatusOK, u)
}

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(refreshCookieName)
	if err != nil || cookie.Value == "" {
		writeProblem(w, http.StatusUnauthorized, "unauthorized", "Отсутствует refresh-токен", nil)
		return
	}

	claims, err := h.tokens.Verify(cookie.Value, TokenTypeRefresh)
	if err != nil {
		h.clearRefreshCookie(w)
		writeProblem(w, http.StatusUnauthorized, "invalid_token", "Недействительный или истекший refresh-токен", nil)
		return
	}

	u, err := h.users.GetByID(r.Context(), claims.Sub)
	if err != nil {
		h.clearRefreshCookie(w)
		writeProblem(w, http.StatusUnauthorized, "user_not_found", "Пользователь не найден", nil)
		return
	}

	h.issueSession(w, http.StatusOK, u)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	h.clearRefreshCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	u, ok := UserFromContext(r.Context())
	if !ok {
		writeProblem(w, http.StatusUnauthorized, "unauthorized", "Необходима авторизация", nil)
		return
	}
	writeJSON(w, http.StatusOK, u)
}

func (h *Handler) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			writeProblem(w, http.StatusUnauthorized, "unauthorized", "Отсутствует токен авторизации", nil)
			return
		}

		tokenStr := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
		claims, err := h.tokens.Verify(tokenStr, TokenTypeAccess)
		if err != nil {
			writeProblem(w, http.StatusUnauthorized, "invalid_token", "Недействительный или истекший access-токен", nil)
			return
		}

		u, err := h.users.GetByID(r.Context(), claims.Sub)
		if err != nil {
			writeProblem(w, http.StatusUnauthorized, "unauthorized", "Пользователь не найден", nil)
			return
		}

		ctx := context.WithValue(r.Context(), userContextKey, u)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func UserFromContext(ctx context.Context) (*user.User, bool) {
	u, ok := ctx.Value(userContextKey).(*user.User)
	return u, ok
}

func (h *Handler) issueSession(w http.ResponseWriter, status int, u *user.User) {
	accessToken, err := h.tokens.GenerateAccessToken(u)
	if err != nil {
		writeProblem(w, http.StatusInternalServerError, "token_error", "Не удалось сгенерировать токен", nil)
		return
	}

	refreshToken, err := h.tokens.GenerateRefreshToken(u)
	if err != nil {
		writeProblem(w, http.StatusInternalServerError, "token_error", "Не удалось сгенерировать refresh-токен", nil)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     refreshCookieName,
		Value:    refreshToken,
		Path:     "/api/auth",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(h.tokens.RefreshTTL().Seconds()),
	})

	writeJSON(w, status, AuthResponse{
		AccessToken: accessToken,
		ExpiresIn:   h.tokens.AccessTTLSeconds(),
		User:        u,
	})
}

func (h *Handler) clearRefreshCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     refreshCookieName,
		Value:    "",
		Path:     "/api/auth",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
	})
}

func validateCredentials(email, password string) []FieldError {
	var fields []FieldError

	if email == "" {
		fields = append(fields, FieldError{
			Field:   "email",
			Code:    "required",
			Message: "Email обязателен",
		})
	} else if _, err := mail.ParseAddress(email); err != nil {
		fields = append(fields, FieldError{
			Field:   "email",
			Code:    "invalid_format",
			Message: "Некорректный формат email",
		})
	}

	if password == "" {
		fields = append(fields, FieldError{
			Field:   "password",
			Code:    "required",
			Message: "Пароль обязателен",
		})
	} else if len(password) < 6 {
		fields = append(fields, FieldError{
			Field:   "password",
			Code:    "min_length",
			Message: "Пароль должен содержать минимум 6 символов",
		})
	}

	return fields
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeProblem(w http.ResponseWriter, status int, code, detail string, fields []FieldError) {
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(Problem{
		Type:    "about:blank",
		Title:   http.StatusText(status),
		Status:  status,
		Code:    code,
		Detail:  detail,
		Message: detail,
		Fields:  fields,
	})
}
