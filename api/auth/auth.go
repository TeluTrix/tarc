package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/TeluTrix/tarc/api/role"
	"github.com/TeluTrix/tarc/api/user"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const (
	sessionCookie = "tarc_session"
	sessionTTL    = 24 * time.Hour
)

type Session struct {
	TokenHash string    `gorm:"type:char(64);primaryKey"`
	UserID    uuid.UUID `gorm:"type:char(36);not null;index"`
	ExpiresAt time.Time `gorm:"not null;index"`
	RevokedAt *time.Time
	User      user.SchemaUser `gorm:"foreignKey:UserID"`
}

type contextKey string

const userContextKey contextKey = "authenticated_user"

type Service struct {
	db           *gorm.DB
	secureCookie bool
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db, secureCookie: os.Getenv("COOKIE_SECURE") == "true"}
}

func (s *Service) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/auth/register", s.register)
	mux.HandleFunc("POST /api/auth/login", s.login)
	mux.HandleFunc("POST /api/auth/logout", s.logout)
	mux.Handle("GET /api/auth/me", s.RequireAuth(http.HandlerFunc(s.aboutMe)))
}

func (s *Service) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u, err := s.currentUser(r)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "authentication required")
			return
		}

		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), userContextKey, u)))
	})
}

func CurrentUser(ctx context.Context) (user.SchemaUser, bool) {
	u, ok := ctx.Value(userContextKey).(user.SchemaUser)
	return u, ok
}

func (s *Service) register(w http.ResponseWriter, r *http.Request) {
	var input credentials
	if !decodeJSON(w, r, &input) {
		return
	}
	if err := input.validate(); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not create account")
		return
	}

	newUser := user.SchemaUser{
		ID:           uuid.New(),
		Email:        strings.ToLower(strings.TrimSpace(input.Email)),
		PasswordHash: string(passwordHash),
		RoleName:     role.UserRole,
	}
	if err := s.db.Create(&newUser).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			writeError(w, http.StatusConflict, "email is already registered")
			return
		}
		writeError(w, http.StatusInternalServerError, "could not create account")
		return
	}

	s.signIn(w, newUser)
}

func (s *Service) login(w http.ResponseWriter, r *http.Request) {
	var input credentials
	if !decodeJSON(w, r, &input) {
		return
	}

	var existing user.SchemaUser
	email := strings.ToLower(strings.TrimSpace(input.Email))
	if err := s.db.Where("email = ?", email).First(&existing).Error; err != nil || bcrypt.CompareHashAndPassword([]byte(existing.PasswordHash), []byte(input.Password)) != nil {
		writeError(w, http.StatusUnauthorized, "invalid email or password")
		return
	}

	s.signIn(w, existing)
}

func (s *Service) logout(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie(sessionCookie); err == nil {
		now := time.Now().UTC()
		s.db.Model(&Session{}).Where("token_hash = ?", hashToken(cookie.Value)).Updates(map[string]any{"revoked_at": &now})
	}

	http.SetCookie(w, clearCookie())
	w.WriteHeader(http.StatusNoContent)
}

func (s *Service) aboutMe(w http.ResponseWriter, r *http.Request) {
	u, ok := CurrentUser(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"id":    u.ID,
		"email": u.Email,
		"role":  u.RoleName,
	})
}

func (s *Service) signIn(w http.ResponseWriter, u user.SchemaUser) {
	rawToken := make([]byte, 32)
	if _, err := rand.Read(rawToken); err != nil {
		writeError(w, http.StatusInternalServerError, "could not create session")
		return
	}

	token := hex.EncodeToString(rawToken)
	session := Session{
		TokenHash: hashToken(token),
		UserID:    u.ID,
		ExpiresAt: time.Now().UTC().Add(sessionTTL),
	}
	if err := s.db.Create(&session).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "could not create session")
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookie,
		Value:    token,
		Path:     "/",
		Expires:  session.ExpiresAt,
		MaxAge:   int(sessionTTL.Seconds()),
		HttpOnly: true,
		Secure:   s.secureCookie,
		SameSite: http.SameSiteLaxMode,
	})
	writeJSON(w, http.StatusOK, map[string]any{
		"token": token,
		"id":    u.ID,
		"email": u.Email,
		"role":  u.RoleName,
	})
}

func (s *Service) currentUser(r *http.Request) (user.SchemaUser, error) {
	cookie, err := r.Cookie(sessionCookie)
	if err != nil {
		return user.SchemaUser{}, err
	}

	var session Session
	if err := s.db.Preload("User").Where("token_hash = ? AND revoked_at IS NULL AND expires_at > ?", hashToken(cookie.Value), time.Now().UTC()).First(&session).Error; err != nil {
		return user.SchemaUser{}, err
	}
	return session.User, nil
}

type credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (c credentials) validate() error {
	if !strings.Contains(c.Email, "@") {
		return errors.New("a valid email is required")
	}
	if len(c.Password) < 8 {
		return errors.New("password must be at least 8 characters")
	}
	return nil
}

func decodeJSON(w http.ResponseWriter, r *http.Request, destination any) bool {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return false
	}
	return true
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func clearCookie() *http.Cookie {
	return &http.Cookie{Name: sessionCookie, Value: "", Path: "/", MaxAge: -1, HttpOnly: true, SameSite: http.SameSiteLaxMode}
}
