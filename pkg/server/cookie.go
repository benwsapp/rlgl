package server

import (
	"net/http"
	"time"
)

type SecureCookieOptions struct {
	Name     string
	Value    string
	Path     string
	MaxAge   int
	HTTPOnly bool
	Secure   bool
	SameSite http.SameSite
}

// SetSecureCookie sets a secure HTTP cookie with configurable options.
// Uses SameSite=Lax for CSRF protection while allowing top-level navigation.
func SetSecureCookie(responseWriter http.ResponseWriter, opts SecureCookieOptions) {
	if opts.Path == "" {
		opts.Path = "/"
	}

	if opts.SameSite == 0 {
		opts.SameSite = http.SameSiteLaxMode
	}

	cookie := &http.Cookie{
		Name:     opts.Name,
		Value:    opts.Value,
		Path:     opts.Path,
		MaxAge:   opts.MaxAge,
		HttpOnly: opts.HTTPOnly,
		Secure:   opts.Secure,
		SameSite: opts.SameSite,
	}

	http.SetCookie(responseWriter, cookie)
}

func SetStrictCookie(responseWriter http.ResponseWriter, name, value string, maxAge int) {
	cookie := &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	}

	http.SetCookie(responseWriter, cookie)
}

func DeleteCookie(responseWriter http.ResponseWriter, name string) {
	cookie := &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Unix(0, 0),
	}

	http.SetCookie(responseWriter, cookie)
}
