package handler

import (
	"html/template"
	"net/http"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/kusakari/itsumo/internal/auth"
	"github.com/kusakari/itsumo/web"
	"golang.org/x/oauth2"
)

type AuthHandler struct {
	mgr *auth.Manager
}

func NewAuthHandler(mgr *auth.Manager) *AuthHandler {
	return &AuthHandler{mgr: mgr}
}

func (h *AuthHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /", h.home)
	mux.HandleFunc("GET /auth/login", h.loginPage)
	mux.HandleFunc("GET /auth/google", h.googleStart)
	mux.HandleFunc("GET /auth/google/callback", h.googleCallback)
	mux.HandleFunc("GET /auth/apple", h.appleStart)
	mux.HandleFunc("POST /auth/apple/callback", h.appleCallback)
	mux.HandleFunc("POST /auth/logout", h.logout)
}

func (h *AuthHandler) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/auth/") {
			next.ServeHTTP(w, r)
			return
		}
		if !h.mgr.IsAuthenticated(r) {
			http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (h *AuthHandler) renderAuth(w http.ResponseWriter, name string, data any) {
	tmpl, err := template.New("").Funcs(funcMap()).ParseFS(web.Templates,
		"templates/base.html",
		"templates/"+name+".html",
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.ExecuteTemplate(w, "base", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *AuthHandler) home(w http.ResponseWriter, r *http.Request) {
	h.renderAuth(w, "home", nil)
}

func (h *AuthHandler) loginPage(w http.ResponseWriter, r *http.Request) {
	if h.mgr.IsAuthenticated(r) {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	h.renderAuth(w, "auth_login", &loginData{
		HasGoogle: h.mgr.Google != nil,
		HasApple:  h.mgr.Apple != nil,
	})
}

// ── Google ────────────────────────────────────────────────────────────────────

func (h *AuthHandler) googleStart(w http.ResponseWriter, r *http.Request) {
	if h.mgr.Google == nil {
		http.Error(w, "Google login not configured", http.StatusNotFound)
		return
	}
	state, err := h.mgr.SetStateCookie(w, false)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, h.mgr.Google.AuthCodeURL(state), http.StatusFound)
}

func (h *AuthHandler) googleCallback(w http.ResponseWriter, r *http.Request) {
	if !h.mgr.VerifyStateCookie(r, r.URL.Query().Get("state")) {
		http.Error(w, "invalid state", http.StatusBadRequest)
		return
	}
	h.mgr.ClearStateCookie(w)

	token, err := h.mgr.Google.Exchange(r.Context(), r.URL.Query().Get("code"))
	if err != nil {
		http.Error(w, "token exchange failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		http.Error(w, "missing id_token", http.StatusInternalServerError)
		return
	}
	userID, err := h.mgr.Google.VerifyIDToken(r.Context(), rawIDToken)
	if err != nil {
		http.Error(w, "id_token verification failed: "+err.Error(), http.StatusUnauthorized)
		return
	}
	if err := h.mgr.SetAuthCookie(w, userID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// ── Apple ─────────────────────────────────────────────────────────────────────

func (h *AuthHandler) appleStart(w http.ResponseWriter, r *http.Request) {
	if h.mgr.Apple == nil {
		http.Error(w, "Apple login not configured", http.StatusNotFound)
		return
	}
	// Apple の form_post コールバックはクロスサイト POST のため SameSite=None が必要
	state, err := h.mgr.SetStateCookie(w, true)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, h.mgr.Apple.AuthCodeURL(state), http.StatusFound)
}

func (h *AuthHandler) appleCallback(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if !h.mgr.VerifyStateCookie(r, r.FormValue("state")) {
		http.Error(w, "invalid state", http.StatusBadRequest)
		return
	}
	h.mgr.ClearStateCookie(w)

	// Apple は id_token をコールバック時に直接送ってくる場合と code フロー両方に対応
	rawIDToken := r.FormValue("id_token")
	if rawIDToken == "" {
		token, err := h.mgr.Apple.Exchange(r.Context(), r.FormValue("code"))
		if err != nil {
			http.Error(w, "token exchange failed: "+err.Error(), http.StatusInternalServerError)
			return
		}
		rawIDToken, _ = token.Extra("id_token").(string)
	}
	if rawIDToken == "" {
		http.Error(w, "missing id_token", http.StatusInternalServerError)
		return
	}
	userID, err := h.mgr.Apple.VerifyIDToken(r.Context(), rawIDToken)
	if err != nil {
		http.Error(w, "id_token verification failed: "+err.Error(), http.StatusUnauthorized)
		return
	}
	if err := h.mgr.SetAuthCookie(w, userID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *AuthHandler) logout(w http.ResponseWriter, r *http.Request) {
	h.mgr.ClearAuthCookie(w)
	http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
}

// ── Template data ─────────────────────────────────────────────────────────────

type loginData struct {
	HasGoogle bool
	HasApple  bool
}

// ── OIDC extra token field helper ─────────────────────────────────────────────
// go-oidc の IDToken から issuer+sub を取るため再エクスポートしておく
var _ = oidc.ScopeOpenID
var _ = oauth2.Token{}
