package middleware

import (
	"net/http"

	"github.com/VladimirVereshchagin/scheduler/internal/auth"
	"github.com/VladimirVereshchagin/scheduler/internal/config"
)

// Auth - authentication check
func Auth(next http.HandlerFunc, cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pass := cfg.Password
		if pass == "" {
			// Password not set, skipping
			next(w, r)
			return
		}

		// Token from cookie
		cookie, err := r.Cookie("token")
		if err != nil {
			// No token, return 401
			http.Error(w, "Authentication required", http.StatusUnauthorized)
			return
		}

		// Validate token
		_, err = auth.ParseToken(cookie.Value, pass)
		if err != nil {
			// Token is invalid
			http.Error(w, "Authentication required", http.StatusUnauthorized)
			return
		}

		// Token is valid, proceed
		next(w, r)
	}
}
