package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"field_service_management_api/internal/auth"
	"field_service_management_api/internal/config"
	"field_service_management_api/internal/models"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

// Register handles POST /api/auth/register.
// It creates a new user in fsm.users and returns a signed JWT.
func Register(db *pgxpool.Pool, cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req models.RegisterRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		req.Username = strings.TrimSpace(req.Username)
		req.Email = strings.TrimSpace(strings.ToLower(req.Email))

		if req.Username == "" || req.Email == "" || req.Password == "" {
			writeError(w, http.StatusBadRequest, "username, email, and password are required")
			return
		}

		hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to hash password")
			return
		}

		var userID int64
		err = db.QueryRow(
			context.Background(),
			`INSERT INTO fsm.users (username, email, password)
			 VALUES ($1, $2, $3)
			 RETURNING id`,
			req.Username, req.Email, string(hash),
		).Scan(&userID)
		if err != nil {
			// Surface a user-friendly message for duplicate email/username
			if strings.Contains(err.Error(), "unique") || strings.Contains(err.Error(), "duplicate") {
				writeError(w, http.StatusConflict, "email or username already in use")
				return
			}
			writeError(w, http.StatusInternalServerError, "failed to create user")
			return
		}

		token, err := auth.GenerateToken(userID, req.Email, req.Username, cfg.JWTSecret)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to generate token")
			return
		}

		writeJSON(w, http.StatusCreated, models.AuthResponse{
			Token: token,
			User: models.UserInfo{
				ID:       userID,
				Username: req.Username,
				Email:    req.Email,
			},
		})
	}
}

// Login handles POST /api/auth/login.
// It looks up the user by email, validates the bcrypt password, and returns a signed JWT.
func Login(db *pgxpool.Pool, cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req models.LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		req.Email = strings.TrimSpace(strings.ToLower(req.Email))

		if req.Email == "" || req.Password == "" {
			writeError(w, http.StatusBadRequest, "email and password are required")
			return
		}

		var user models.User
		err := db.QueryRow(
			context.Background(),
			`SELECT id, username, email, password FROM fsm.users WHERE email = $1`,
			req.Email,
		).Scan(&user.ID, &user.Username, &user.Email, &user.Password)
		if err != nil {
			// Return generic message to avoid user enumeration
			writeError(w, http.StatusUnauthorized, "invalid email or password")
			return
		}

		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
			writeError(w, http.StatusUnauthorized, "invalid email or password")
			return
		}

		token, err := auth.GenerateToken(user.ID, user.Email, user.Username, cfg.JWTSecret)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to generate token")
			return
		}

		writeJSON(w, http.StatusOK, models.AuthResponse{
			Token: token,
			User: models.UserInfo{
				ID:       user.ID,
				Username: user.Username,
				Email:    user.Email,
			},
		})
	}
}
