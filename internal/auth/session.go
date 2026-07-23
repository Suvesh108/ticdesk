package auth

import (
	"net/http"
	"time"

	"github.com/alexedwards/scs/v2"
)

var SessionManager *scs.SessionManager

// InitSessionManager initializes the SCS session manager.
func InitSessionManager(secret string) *scs.SessionManager {
	sm := scs.New()
	sm.Lifetime = 24 * time.Hour
	sm.Cookie.Name = "ticdesk_session"
	sm.Cookie.HttpOnly = true
	sm.Cookie.SameSite = http.SameSiteLaxMode
	sm.Cookie.Secure = false // Set true in production with HTTPS

	SessionManager = sm
	return sm
}
