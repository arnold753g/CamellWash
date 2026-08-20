package middleware

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"starter-backend/internal/config"
	"starter-backend/internal/models"
	"starter-backend/internal/respond"
	"starter-backend/internal/security"

	"gorm.io/gorm"
)

const activityWriteInterval = time.Minute

type Authenticator struct {
	db      *gorm.DB
	cfg     config.Config
	origins map[string]struct{}
}

func NewAuthenticator(db *gorm.DB, cfg config.Config) *Authenticator {
	origins := make(map[string]struct{}, len(cfg.AllowedOrigins))
	for _, origin := range cfg.AllowedOrigins {
		origins[origin] = struct{}{}
	}
	return &Authenticator{db: db, cfg: cfg, origins: origins}
}

func (a *Authenticator) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(security.CookieName)
		if err != nil || strings.TrimSpace(cookie.Value) == "" {
			respond.Error(w, r, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication is required", nil)
			return
		}
		if unsafeMethod(r.Method) {
			if origin := strings.TrimRight(strings.TrimSpace(r.Header.Get("Origin")), "/"); origin != "" {
				if _, ok := a.origins[origin]; !ok {
					respond.Error(w, r, http.StatusForbidden, "ORIGIN_FORBIDDEN", "Request origin is not allowed", nil)
					return
				}
			}
			if !security.ValidCSRFToken(a.cfg.SessionSecret, cookie.Value, r.Header.Get("X-CSRF-Token")) {
				respond.Error(w, r, http.StatusForbidden, "CSRF_INVALID", "CSRF protection failed", nil)
				return
			}
		}

		var session models.Session
		err = a.db.Preload("User").Where("token_hash = ? AND revoked_at IS NULL", security.HashToken(cookie.Value)).First(&session).Error
		if err != nil {
			respond.Error(w, r, http.StatusUnauthorized, "INVALID_SESSION", "Session is invalid or expired", nil)
			return
		}
		now := time.Now().UTC()
		deadline := session.LastActivityAt.Add(a.cfg.SessionIdleTimeout)
		if session.ExpiresAt.Before(deadline) {
			deadline = session.ExpiresAt
		}
		if !now.Before(deadline) || session.User.Status != models.StatusActive {
			_ = a.db.Model(&models.Session{}).Where("id = ? AND revoked_at IS NULL", session.ID).Update("revoked_at", now).Error
			respond.Error(w, r, http.StatusUnauthorized, "SESSION_EXPIRED", "Session is invalid or expired", nil)
			return
		}
		if !models.ValidRole(session.User.Role) {
			respond.Error(w, r, http.StatusForbidden, "FORBIDDEN", "Account role is not allowed", nil)
			return
		}
		if now.Sub(session.LastActivityAt) >= activityWriteInterval {
			result := a.db.Model(&models.Session{}).
				Where("id = ? AND revoked_at IS NULL AND expires_at > ?", session.ID, now).
				Update("last_activity_at", now)
			if result.Error != nil || result.RowsAffected != 1 {
				respond.Error(w, r, http.StatusUnauthorized, "INVALID_SESSION", "Session is invalid or expired", nil)
				return
			}
		}
		next.ServeHTTP(w, WithPrincipal(r, Principal{User: session.User, SessionID: session.ID}))
	})
}

func (a *Authenticator) OriginGuard(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := strings.TrimRight(strings.TrimSpace(r.Header.Get("Origin")), "/")
		if origin != "" {
			if _, ok := a.origins[origin]; !ok {
				respond.Error(w, r, http.StatusForbidden, "ORIGIN_FORBIDDEN", "Request origin is not allowed", nil)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

func RequireRoles(roles ...string) func(http.Handler) http.Handler {
	allowed := make(map[string]struct{}, len(roles))
	for _, role := range roles {
		allowed[role] = struct{}{}
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			principal, ok := CurrentPrincipal(r)
			if !ok {
				respond.Error(w, r, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication is required", nil)
				return
			}
			if _, ok := allowed[principal.User.Role]; !ok {
				respond.Error(w, r, http.StatusForbidden, "FORBIDDEN", "You do not have permission for this action", nil)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func RevokeUserSessions(db *gorm.DB, userID uint64) error {
	now := time.Now().UTC()
	return db.Model(&models.Session{}).Where("user_id = ? AND revoked_at IS NULL", userID).Update("revoked_at", now).Error
}

func unsafeMethod(method string) bool {
	return method != http.MethodGet && method != http.MethodHead && method != http.MethodOptions
}

func IsNotFound(err error) bool { return errors.Is(err, gorm.ErrRecordNotFound) }
