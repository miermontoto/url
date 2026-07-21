package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

// cliente oidc de mier id en modo backend-for-frontend: el navegador nunca ve un
// token. el backend hace el intercambio de code (confidential client + pkce) y
// guarda el id_token / refresh_token en cookies httpOnly. las rutas protegidas
// verifican el id_token contra el jwks del emisor (stateless, sin sesiones en db).

const (
	cookieIDToken = "url_id"      // id_token: aserción de identidad de la sesión
	cookieRefresh = "url_rt"      // refresh_token para renovar sin re-login
	cookieState   = "url_state"   // state del round-trip (anti-csrf de login)
	cookiePKCE    = "url_pkce"    // verifier pkce
	cookieNonce   = "url_nonce"   // nonce del id_token
	flowCookieTTL = 10 * time.Minute
	stateBytes    = 32
)

// claims que consume la app: el owner de las urls es el preferred_username
type idClaims struct {
	Subject           string `json:"sub"`
	PreferredUsername string `json:"preferred_username"`
	Email             string `json:"email"`
}

// el owner efectivo: preferred_username y, si falta, el sub como respaldo
func (c idClaims) owner() string {
	if c.PreferredUsername != "" {
		return c.PreferredUsername
	}
	return c.Subject
}

// config del cliente oidc, leída del entorno
type oidcConfig struct {
	issuer       string
	clientID     string
	clientSecret string
	redirectURL  string
	postLogout   string // a dónde vuelve el navegador tras el end_session del idp
}

// auth agrupa el proveedor y el oauth2.Config. el proveedor se inicializa de forma
// perezosa (una llamada de red al emisor): así una caída del idp no tumba el boot
// ni las redirecciones públicas, solo el login.
type auth struct {
	cfg      oidcConfig
	mu       sync.Mutex
	provider *oidc.Provider
	verifier *oidc.IDTokenVerifier
	oauth    *oauth2.Config
	secure   bool // cookies Secure: activo cuando el redirect es https
}

func newAuth(cfg oidcConfig) *auth {
	return &auth{cfg: cfg, secure: strings.HasPrefix(cfg.redirectURL, "https://")}
}

// asegura provider/verifier/oauth inicializados; reintenta si un intento previo falló
func (a *auth) ensure(ctx context.Context) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.provider != nil {
		return nil
	}
	provider, err := oidc.NewProvider(ctx, a.cfg.issuer)
	if err != nil {
		return err
	}
	a.provider = provider
	a.verifier = provider.Verifier(&oidc.Config{ClientID: a.cfg.clientID})
	a.oauth = &oauth2.Config{
		ClientID:     a.cfg.clientID,
		ClientSecret: a.cfg.clientSecret,
		RedirectURL:  a.cfg.redirectURL,
		Endpoint:     provider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email", oidc.ScopeOfflineAccess},
	}
	return nil
}

func randToken() string {
	b := make([]byte, stateBytes)
	rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

// cookie httpOnly del flujo (state/pkce/nonce): corta y acotada a /auth
func (a *auth) setFlowCookie(w http.ResponseWriter, name, value string) {
	http.SetCookie(w, &http.Cookie{
		Name: name, Value: value, Path: "/auth", HttpOnly: true,
		Secure: a.secure, SameSite: http.SameSiteLaxMode, MaxAge: int(flowCookieTTL.Seconds()),
	})
}

// cookie de sesión (id_token/refresh): httpOnly, disponible en toda la app
func (a *auth) setSessionCookie(w http.ResponseWriter, name, value string) {
	http.SetCookie(w, &http.Cookie{
		Name: name, Value: value, Path: "/", HttpOnly: true,
		Secure: a.secure, SameSite: http.SameSiteLaxMode,
	})
}

func (a *auth) clearCookie(w http.ResponseWriter, name, path string) {
	http.SetCookie(w, &http.Cookie{
		Name: name, Value: "", Path: path, HttpOnly: true,
		Secure: a.secure, SameSite: http.SameSiteLaxMode, MaxAge: -1,
	})
}

// GET /auth/login: arranca el flujo authorization code + pkce y redirige al emisor
func (a *auth) LoginHandler(w http.ResponseWriter, r *http.Request) {
	if err := a.ensure(r.Context()); err != nil {
		http.Error(w, "identity provider unavailable", http.StatusServiceUnavailable)
		return
	}
	state, verifier, nonce := randToken(), oauth2.GenerateVerifier(), randToken()
	a.setFlowCookie(w, cookieState, state)
	a.setFlowCookie(w, cookiePKCE, verifier)
	a.setFlowCookie(w, cookieNonce, nonce)
	url := a.oauth.AuthCodeURL(state, oidc.Nonce(nonce), oauth2.S256ChallengeOption(verifier))
	http.Redirect(w, r, url, http.StatusFound)
}

// GET /auth/callback: valida el state, intercambia el code y persiste la sesión
func (a *auth) CallbackHandler(w http.ResponseWriter, r *http.Request) {
	if err := a.ensure(r.Context()); err != nil {
		http.Error(w, "identity provider unavailable", http.StatusServiceUnavailable)
		return
	}
	// binding de navegador: el state de la url debe coincidir con el de la cookie
	stateCookie, err := r.Cookie(cookieState)
	if err != nil || stateCookie.Value != r.URL.Query().Get("state") {
		http.Error(w, "invalid state", http.StatusBadRequest)
		return
	}
	pkce, err := r.Cookie(cookiePKCE)
	if err != nil {
		http.Error(w, "missing pkce verifier", http.StatusBadRequest)
		return
	}
	nonce, err := r.Cookie(cookieNonce)
	if err != nil {
		http.Error(w, "missing nonce", http.StatusBadRequest)
		return
	}
	// cookies del flujo consumidas: se limpian pase lo que pase
	for _, n := range []string{cookieState, cookiePKCE, cookieNonce} {
		a.clearCookie(w, n, "/auth")
	}

	oauth2Token, err := a.oauth.Exchange(r.Context(), r.URL.Query().Get("code"), oauth2.VerifierOption(pkce.Value))
	if err != nil {
		http.Error(w, "code exchange failed", http.StatusBadGateway)
		return
	}
	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		http.Error(w, "no id_token in response", http.StatusBadGateway)
		return
	}
	idToken, err := a.verifier.Verify(r.Context(), rawIDToken)
	if err != nil {
		http.Error(w, "id_token verification failed", http.StatusBadGateway)
		return
	}
	if idToken.Nonce != nonce.Value {
		http.Error(w, "nonce mismatch", http.StatusBadRequest)
		return
	}

	a.setSessionCookie(w, cookieIDToken, rawIDToken)
	if oauth2Token.RefreshToken != "" {
		a.setSessionCookie(w, cookieRefresh, oauth2Token.RefreshToken)
	}
	http.Redirect(w, r, "/", http.StatusFound)
}

// GET /auth/logout: por defecto cierra SOLO la sesión local (las cookies de url) y
// deja viva la sesión sso de mier id, así volver a entrar es silencioso (sin pedir
// credenciales otra vez). con ?all=1 propaga al idp (end_session) para "salir de todo".
func (a *auth) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	idToken, _ := r.Cookie(cookieIDToken)
	a.clearCookie(w, cookieIDToken, "/")
	a.clearCookie(w, cookieRefresh, "/")

	if r.URL.Query().Get("all") == "1" && a.cfg.postLogout != "" && a.ensure(r.Context()) == nil {
		if endpoint, ok := a.endSessionEndpoint(); ok {
			q := "post_logout_redirect_uri=" + a.cfg.postLogout
			if idToken != nil {
				q += "&id_token_hint=" + idToken.Value
			}
			http.Redirect(w, r, endpoint+"?"+q, http.StatusFound)
			return
		}
	}
	http.Redirect(w, r, "/", http.StatusFound)
}

// end_session_endpoint del discovery (no lo expone la api tipada de go-oidc)
func (a *auth) endSessionEndpoint() (string, bool) {
	var claims struct {
		EndSession string `json:"end_session_endpoint"`
	}
	if a.provider == nil || a.provider.Claims(&claims) != nil || claims.EndSession == "" {
		return "", false
	}
	return claims.EndSession, true
}

// GET /auth/me: identidad actual para que el frontend decida qué vista mostrar
func (a *auth) MeHandler(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(claimsKey).(idClaims)
	if !ok {
		failureResponse(w, "not authenticated", http.StatusUnauthorized)
		return
	}
	successResponse(w, map[string]string{"username": claims.owner(), "email": claims.Email})
}

type contextKey string

const claimsKey contextKey = "claims"

// RequireAuth verifica el id_token de la cookie; si expiró, lo renueva con el
// refresh_token. deja las claims en el contexto y el owner en la cabecera X-User.
func (a *auth) RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := a.ensure(r.Context()); err != nil {
			failureResponse(w, "identity provider unavailable", http.StatusServiceUnavailable)
			return
		}
		claims, ok := a.authenticate(w, r)
		if !ok {
			failureResponse(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		r.Header.Set("X-User", claims.owner())
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), claimsKey, claims)))
	}
}

// verifica la cookie de sesión; renueva con el refresh si el id_token no vale
func (a *auth) authenticate(w http.ResponseWriter, r *http.Request) (idClaims, bool) {
	if c, err := r.Cookie(cookieIDToken); err == nil {
		if idToken, err := a.verifier.Verify(r.Context(), c.Value); err == nil {
			var claims idClaims
			if idToken.Claims(&claims) == nil {
				return claims, true
			}
		}
	}
	// id_token ausente o caducado: intenta renovar con el refresh_token
	rt, err := r.Cookie(cookieRefresh)
	if err != nil {
		return idClaims{}, false
	}
	newToken, err := a.oauth.TokenSource(r.Context(), &oauth2.Token{RefreshToken: rt.Value}).Token()
	if err != nil {
		a.clearCookie(w, cookieRefresh, "/")
		return idClaims{}, false
	}
	rawIDToken, ok := newToken.Extra("id_token").(string)
	if !ok {
		return idClaims{}, false
	}
	idToken, err := a.verifier.Verify(r.Context(), rawIDToken)
	if err != nil {
		return idClaims{}, false
	}
	var claims idClaims
	if idToken.Claims(&claims) != nil {
		return idClaims{}, false
	}
	a.setSessionCookie(w, cookieIDToken, rawIDToken)
	if newToken.RefreshToken != "" && newToken.RefreshToken != rt.Value {
		a.setSessionCookie(w, cookieRefresh, newToken.RefreshToken)
	}
	return claims, true
}

// ownerOf extrae el owner de las claims del contexto (rutas ya autenticadas)
func ownerOf(r *http.Request) string {
	if claims, ok := r.Context().Value(claimsKey).(idClaims); ok {
		return claims.owner()
	}
	return r.Header.Get("X-User")
}
