package middleware

import (
	"net/http"

	"git.tecblic.com/sanyog-tecblic/ecom/controller/config"
	"github.com/alexedwards/scs/v2"
	"github.com/justinas/nosurf"
)

var app config.AppConfig
var session *scs.SessionManager

func NoSurf(next http.Handler) http.Handler {
	csrfHandler := nosurf.New(next)

	csrfHandler.SetBaseCookie(http.Cookie{
		HttpOnly: true,
		Path:     "/",
		Secure:   app.InProduction,
		SameSite: http.SameSiteLaxMode,
	})
	return csrfHandler
}

// SessionLoad loads and saves session data for current request
func SessionLoad(next http.Handler) http.Handler {
	return session.LoadAndSave(next)
}
