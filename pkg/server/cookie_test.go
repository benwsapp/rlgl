package server_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/benwsapp/rlgl/pkg/server"
)

func TestSetSecureCookie(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()

	opts := server.SecureCookieOptions{
		Name:     "test-cookie",
		Value:    "test-value",
		Path:     "/test",
		MaxAge:   3600,
		HTTPOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}

	server.SetSecureCookie(rec, opts)

	cookies := rec.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("expected 1 cookie, got %d", len(cookies))
	}

	cookie := cookies[0]

	if cookie.Name != "test-cookie" {
		t.Errorf("expected name 'test-cookie', got %s", cookie.Name)
	}

	if cookie.Value != "test-value" {
		t.Errorf("expected value 'test-value', got %s", cookie.Value)
	}

	if cookie.Path != "/test" {
		t.Errorf("expected path '/test', got %s", cookie.Path)
	}

	if cookie.MaxAge != 3600 {
		t.Errorf("expected MaxAge 3600, got %d", cookie.MaxAge)
	}

	if !cookie.HttpOnly {
		t.Error("expected HttpOnly to be true")
	}

	if !cookie.Secure {
		t.Error("expected Secure to be true")
	}

	if cookie.SameSite != http.SameSiteLaxMode {
		t.Errorf("expected SameSite Lax, got %v", cookie.SameSite)
	}
}

func TestSetSecureCookieDefaults(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()

	opts := server.SecureCookieOptions{
		Name:  "test",
		Value: "value",
	}

	server.SetSecureCookie(rec, opts)

	cookies := rec.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("expected 1 cookie, got %d", len(cookies))
	}

	cookie := cookies[0]

	if cookie.Path != "/" {
		t.Errorf("expected default path '/', got %s", cookie.Path)
	}

	if cookie.SameSite != http.SameSiteLaxMode {
		t.Errorf("expected default SameSite Lax, got %v", cookie.SameSite)
	}
}

func TestSetStrictCookie(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()

	server.SetStrictCookie(rec, "strict-cookie", "strict-value", 7200)

	cookies := rec.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("expected 1 cookie, got %d", len(cookies))
	}

	cookie := cookies[0]

	if cookie.Name != "strict-cookie" {
		t.Errorf("expected name 'strict-cookie', got %s", cookie.Name)
	}

	if cookie.Value != "strict-value" {
		t.Errorf("expected value 'strict-value', got %s", cookie.Value)
	}

	if cookie.MaxAge != 7200 {
		t.Errorf("expected MaxAge 7200, got %d", cookie.MaxAge)
	}

	if !cookie.HttpOnly {
		t.Error("expected HttpOnly to be true")
	}

	if !cookie.Secure {
		t.Error("expected Secure to be true")
	}

	if cookie.SameSite != http.SameSiteStrictMode {
		t.Errorf("expected SameSite Strict, got %v", cookie.SameSite)
	}

	if cookie.Path != "/" {
		t.Errorf("expected path '/', got %s", cookie.Path)
	}
}

func TestDeleteCookie(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()

	server.DeleteCookie(rec, "cookie-to-delete")

	cookies := rec.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("expected 1 cookie, got %d", len(cookies))
	}

	cookie := cookies[0]

	if cookie.Name != "cookie-to-delete" {
		t.Errorf("expected name 'cookie-to-delete', got %s", cookie.Name)
	}

	if cookie.Value != "" {
		t.Errorf("expected empty value, got %s", cookie.Value)
	}

	if cookie.MaxAge != -1 {
		t.Errorf("expected MaxAge -1, got %d", cookie.MaxAge)
	}

	if !cookie.HttpOnly {
		t.Error("expected HttpOnly to be true")
	}

	if !cookie.Secure {
		t.Error("expected Secure to be true")
	}

	if cookie.SameSite != http.SameSiteLaxMode {
		t.Errorf("expected SameSite Lax, got %v", cookie.SameSite)
	}

	if !cookie.Expires.IsZero() && cookie.Expires.Unix() != 0 {
		t.Errorf("expected Expires to be Unix epoch, got %v", cookie.Expires)
	}
}

func TestSecureCookieOptionsStruct(t *testing.T) {
	t.Parallel()

	opts := server.SecureCookieOptions{
		Name:     "test",
		Value:    "value",
		Path:     "/path",
		MaxAge:   1000,
		HTTPOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	}

	if opts.Name != "test" {
		t.Errorf("expected name 'test', got %s", opts.Name)
	}

	if opts.Value != "value" {
		t.Errorf("expected value 'value', got %s", opts.Value)
	}

	if opts.Path != "/path" {
		t.Errorf("expected path '/path', got %s", opts.Path)
	}

	if opts.MaxAge != 1000 {
		t.Errorf("expected MaxAge 1000, got %d", opts.MaxAge)
	}

	if opts.HTTPOnly != true {
		t.Error("expected HTTPOnly to be true")
	}

	if opts.Secure != true {
		t.Error("expected Secure to be true")
	}

	if opts.SameSite != http.SameSiteStrictMode {
		t.Errorf("expected SameSite Strict, got %v", opts.SameSite)
	}
}
