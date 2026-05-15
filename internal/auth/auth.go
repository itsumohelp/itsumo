package auth

import (
	"context"
	"crypto/ecdsa"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	jwtv5 "github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2"
)

const (
	authCookieName  = "itsumo_session"
	stateCookieName = "itsumo_oauth_state"
	sessionMaxAge   = 30 * 24 * 60 * 60
	stateMaxAge     = 10 * 60
)

type Manager struct {
	secret []byte
	secure bool
	Google *OIDCProvider
	Apple  *AppleProvider
}

type OIDCProvider struct {
	config   *oauth2.Config
	verifier *oidc.IDTokenVerifier
}

type AppleProvider struct {
	config     *oauth2.Config
	verifier   *oidc.IDTokenVerifier
	teamID     string
	keyID      string
	privateKey *ecdsa.PrivateKey
}

func NewManager(secret []byte, secure bool) *Manager {
	return &Manager{secret: secret, secure: secure}
}

func (m *Manager) SetupGoogle(ctx context.Context, clientID, clientSecret, redirectURL string) error {
	provider, err := oidc.NewProvider(ctx, "https://accounts.google.com")
	if err != nil {
		return fmt.Errorf("google oidc: %w", err)
	}
	m.Google = &OIDCProvider{
		verifier: provider.Verifier(&oidc.Config{ClientID: clientID}),
		config: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			Endpoint:     provider.Endpoint(),
			RedirectURL:  redirectURL,
			Scopes:       []string{oidc.ScopeOpenID},
		},
	}
	return nil
}

func (m *Manager) SetupApple(ctx context.Context, clientID, teamID, keyID, privateKeyPEM, redirectURL string) error {
	keyPEM := strings.ReplaceAll(privateKeyPEM, `\n`, "\n")
	block, _ := pem.Decode([]byte(keyPEM))
	if block == nil {
		return errors.New("apple: invalid PEM")
	}
	raw, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return fmt.Errorf("apple: parse key: %w", err)
	}
	key, ok := raw.(*ecdsa.PrivateKey)
	if !ok {
		return errors.New("apple: expected ECDSA key")
	}

	provider, err := oidc.NewProvider(ctx, "https://appleid.apple.com")
	if err != nil {
		return fmt.Errorf("apple oidc: %w", err)
	}
	clientSecret, err := appleClientSecret(teamID, clientID, keyID, key)
	if err != nil {
		return fmt.Errorf("apple client secret: %w", err)
	}
	m.Apple = &AppleProvider{
		teamID:     teamID,
		keyID:      keyID,
		privateKey: key,
		verifier:   provider.Verifier(&oidc.Config{ClientID: clientID}),
		config: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			Endpoint:     provider.Endpoint(),
			RedirectURL:  redirectURL,
			Scopes:       []string{oidc.ScopeOpenID},
		},
	}
	return nil
}

// AppleRefreshClientSecret は Apple の client_secret JWT を再生成します。
// 最大有効期限は 180 日なので定期的な呼び出しが必要です。
func (m *Manager) AppleRefreshClientSecret() error {
	if m.Apple == nil {
		return nil
	}
	s, err := appleClientSecret(m.Apple.teamID, m.Apple.config.ClientID, m.Apple.keyID, m.Apple.privateKey)
	if err != nil {
		return err
	}
	m.Apple.config.ClientSecret = s
	return nil
}

func appleClientSecret(teamID, clientID, keyID string, key *ecdsa.PrivateKey) (string, error) {
	now := time.Now()
	token := jwtv5.NewWithClaims(jwtv5.SigningMethodES256, jwtv5.MapClaims{
		"iss": teamID,
		"iat": now.Unix(),
		"exp": now.Add(180 * 24 * time.Hour).Unix(),
		"aud": "https://appleid.apple.com",
		"sub": clientID,
	})
	token.Header["kid"] = keyID
	return token.SignedString(key)
}

// VerifyIDToken は OIDC の id_token を検証して issuer:sub 形式のユーザーIDを返します。
func (p *OIDCProvider) VerifyIDToken(ctx context.Context, rawToken string) (string, error) {
	tok, err := p.verifier.Verify(ctx, rawToken)
	if err != nil {
		return "", err
	}
	return tok.Issuer + ":" + tok.Subject, nil
}

func (p *AppleProvider) VerifyIDToken(ctx context.Context, rawToken string) (string, error) {
	tok, err := p.verifier.Verify(ctx, rawToken)
	if err != nil {
		return "", err
	}
	return tok.Issuer + ":" + tok.Subject, nil
}

func (p *OIDCProvider) AuthCodeURL(state string) string {
	return p.config.AuthCodeURL(state)
}

func (p *AppleProvider) AuthCodeURL(state string) string {
	return p.config.AuthCodeURL(state, oauth2.SetAuthURLParam("response_mode", "form_post"))
}

func (p *OIDCProvider) Exchange(ctx context.Context, code string) (*oauth2.Token, error) {
	return p.config.Exchange(ctx, code)
}

func (p *AppleProvider) Exchange(ctx context.Context, code string) (*oauth2.Token, error) {
	return p.config.Exchange(ctx, code)
}

// ── Session ───────────────────────────────────────────────────────────────────

type authClaim struct {
	UserID    string `json:"u"`
	ExpiresAt int64  `json:"e"`
}

func (m *Manager) SetAuthCookie(w http.ResponseWriter, userID string) error {
	data, err := json.Marshal(authClaim{
		UserID:    userID,
		ExpiresAt: time.Now().Add(30 * 24 * time.Hour).Unix(),
	})
	if err != nil {
		return err
	}
	http.SetCookie(w, m.signedCookie(authCookieName, data, sessionMaxAge, http.SameSiteLaxMode))
	return nil
}

func (m *Manager) IsAuthenticated(r *http.Request) bool {
	_, ok := m.getAuthClaim(r)
	return ok
}

func (m *Manager) CurrentUserID(r *http.Request) string {
	claim, ok := m.getAuthClaim(r)
	if !ok {
		return ""
	}
	return claim.UserID
}

func (m *Manager) getAuthClaim(r *http.Request) (*authClaim, bool) {
	data, err := m.verifyCookie(r, authCookieName)
	if err != nil {
		return nil, false
	}
	var claim authClaim
	if err := json.Unmarshal(data, &claim); err != nil {
		return nil, false
	}
	if time.Now().Unix() >= claim.ExpiresAt {
		return nil, false
	}
	return &claim, true
}

func (m *Manager) ClearAuthCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{Name: authCookieName, Value: "", Path: "/", MaxAge: -1})
}

// ── State cookie (OAuth CSRF 対策) ────────────────────────────────────────────

func (m *Manager) SetStateCookie(w http.ResponseWriter, crossSite bool) (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	state := base64.URLEncoding.EncodeToString(b)
	sameSite := http.SameSiteLaxMode
	if crossSite {
		// Apple の form_post コールバックはクロスサイト POST になるため SameSite=None が必要
		// SameSite=None は Secure=true (HTTPS) 環境のみ有効
		sameSite = http.SameSiteNoneMode
	}
	http.SetCookie(w, m.signedCookie(stateCookieName, []byte(state), stateMaxAge, sameSite))
	return state, nil
}

func (m *Manager) VerifyStateCookie(r *http.Request, state string) bool {
	data, err := m.verifyCookie(r, stateCookieName)
	if err != nil {
		return false
	}
	return hmac.Equal(data, []byte(state))
}

func (m *Manager) ClearStateCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{Name: stateCookieName, Value: "", Path: "/", MaxAge: -1})
}

// ── Cookie helpers ────────────────────────────────────────────────────────────

func (m *Manager) signedCookie(name string, data []byte, maxAge int, sameSite http.SameSite) *http.Cookie {
	mac := hmac.New(sha256.New, m.secret)
	mac.Write(data)
	val := base64.URLEncoding.EncodeToString(data) + "." + base64.URLEncoding.EncodeToString(mac.Sum(nil))
	return &http.Cookie{
		Name:     name,
		Value:    val,
		Path:     "/",
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   m.secure,
		SameSite: sameSite,
	}
}

func (m *Manager) verifyCookie(r *http.Request, name string) ([]byte, error) {
	cookie, err := r.Cookie(name)
	if err != nil {
		return nil, err
	}
	idx := strings.LastIndex(cookie.Value, ".")
	if idx < 0 {
		return nil, errors.New("invalid cookie format")
	}
	data, err := base64.URLEncoding.DecodeString(cookie.Value[:idx])
	if err != nil {
		return nil, err
	}
	sig, err := base64.URLEncoding.DecodeString(cookie.Value[idx+1:])
	if err != nil {
		return nil, err
	}
	mac := hmac.New(sha256.New, m.secret)
	mac.Write(data)
	if !hmac.Equal(sig, mac.Sum(nil)) {
		return nil, errors.New("invalid signature")
	}
	return data, nil
}
