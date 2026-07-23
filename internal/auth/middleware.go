package auth

import (
	"context"
	"net/http"
	"ticDesk/internal/models"
)

type contextKey string

const UserContextKey contextKey = "user"

// RequireAuth middleware ensures the user is logged in.
func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := SessionManager.GetString(r.Context(), "user_id")
		if userID == "" {
			if r.Header.Get("HX-Request") == "true" {
				w.Header().Set("HX-Redirect", "/login")
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		userRole := SessionManager.GetString(r.Context(), "user_role")
		userName := SessionManager.GetString(r.Context(), "user_name")
		userEmail := SessionManager.GetString(r.Context(), "user_email")

		user := &models.User{
			ID:    userID,
			Name:  userName,
			Email: userEmail,
			Role:  models.UserRole(userRole),
		}

		ctx := context.WithValue(r.Context(), UserContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireRole middleware restricts route access to specific roles.
func RequireRole(allowedRoles ...models.UserRole) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, ok := r.Context().Value(UserContextKey).(*models.User)
			if !ok || user == nil {
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}

			hasRole := false
			for _, role := range allowedRoles {
				if user.Role == role {
					hasRole = true
					break
				}
			}

			if !hasRole {
				http.Error(w, "Forbidden: insufficient permissions", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// GetUserFromContext retrieves the User struct stored in context by RequireAuth.
func GetUserFromContext(ctx context.Context) *models.User {
	user, ok := ctx.Value(UserContextKey).(*models.User)
	if !ok {
		return nil
	}
	return user
}
