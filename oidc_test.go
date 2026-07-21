package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// owner: preferred_username manda; el sub es el respaldo si falta
func TestOwner(t *testing.T) {
	cases := []struct {
		claims idClaims
		want   string
	}{
		{idClaims{Subject: "uuid-1", PreferredUsername: "juanmier"}, "juanmier"},
		{idClaims{Subject: "uuid-2", PreferredUsername: ""}, "uuid-2"},
	}
	for _, c := range cases {
		if got := c.claims.owner(); got != c.want {
			t.Errorf("owner() = %q, quería %q", got, c.want)
		}
	}
}

// una ruta protegida sin cookie de sesión responde 401 sin llamar al handler.
// el provider no está inicializado, pero ensure() falla contra un emisor falso,
// así que este test verifica el camino "idp no disponible / no autenticado".
func TestRequireAuthRejectsWithoutSession(t *testing.T) {
	a := newAuth(oidcConfig{
		issuer:      "http://127.0.0.1:1", // inalcanzable a propósito
		clientID:    "cid",
		redirectURL: "http://localhost/auth/callback",
	})
	called := false
	handler := a.RequireAuth(func(w http.ResponseWriter, r *http.Request) { called = true })

	req := httptest.NewRequest(http.MethodGet, "/my", nil)
	rec := httptest.NewRecorder()
	handler(rec, req)

	if called {
		t.Fatal("el handler protegido no debía ejecutarse sin sesión")
	}
	if rec.Code == http.StatusOK {
		t.Fatalf("esperaba un código de error, no 200")
	}
}

// las cookies de sesión se marcan httpOnly (el js no puede leer el token)
func TestSessionCookieIsHttpOnly(t *testing.T) {
	a := newAuth(oidcConfig{redirectURL: "https://url.example/auth/callback"})
	rec := httptest.NewRecorder()
	a.setSessionCookie(rec, cookieIDToken, "un-token")

	cookies := rec.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("esperaba 1 cookie, hubo %d", len(cookies))
	}
	ck := cookies[0]
	if !ck.HttpOnly {
		t.Error("la cookie de sesión debe ser httpOnly")
	}
	if !ck.Secure {
		t.Error("con redirect https la cookie debe ser Secure")
	}
}
