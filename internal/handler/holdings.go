package handler

import (
	"encoding/json"
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"github.com/kusakari/itsumo/internal/auth"
	"github.com/kusakari/itsumo/internal/domain"
	"github.com/kusakari/itsumo/internal/store/postgres"
	"github.com/kusakari/itsumo/web"
)

type HoldingsHandler struct {
	pg  *postgres.Repo
	mgr *auth.Manager
}

func NewHoldingsHandler(pg *postgres.Repo, mgr *auth.Manager) *HoldingsHandler {
	return &HoldingsHandler{pg: pg, mgr: mgr}
}

func (h *HoldingsHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /holdings", h.holdingsPage)
	mux.HandleFunc("POST /holdings", h.upsertHolding)
	mux.HandleFunc("POST /holdings/{code}/delete", h.deleteHolding)
	mux.HandleFunc("GET /api/stocks/search", h.searchStocks)
}

type holdingsPageData struct {
	Holdings []*domain.Holding
}

func (h *HoldingsHandler) holdingsPage(w http.ResponseWriter, r *http.Request) {
	userID := h.mgr.CurrentUserID(r)
	holdings, err := h.pg.ListHoldings(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl, err := template.New("").Funcs(funcMap()).ParseFS(web.Templates,
		"templates/base.html",
		"templates/holdings.html",
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	tmpl.ExecuteTemplate(w, "base", &holdingsPageData{Holdings: holdings})
}

func (h *HoldingsHandler) upsertHolding(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	code := strings.TrimSpace(r.FormValue("stock_code"))
	if code == "" {
		http.Error(w, "証券コードを選択してください", http.StatusBadRequest)
		return
	}
	shares, err := strconv.Atoi(r.FormValue("shares"))
	if err != nil || shares <= 0 {
		http.Error(w, "株数は1以上の整数を入力してください", http.StatusBadRequest)
		return
	}
	userID := h.mgr.CurrentUserID(r)
	if err := h.pg.UpsertHolding(r.Context(), userID, code, shares); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/holdings", http.StatusSeeOther)
}

func (h *HoldingsHandler) deleteHolding(w http.ResponseWriter, r *http.Request) {
	userID := h.mgr.CurrentUserID(r)
	code := r.PathValue("code")
	if err := h.pg.DeleteHolding(r.Context(), userID, code); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/holdings", http.StatusSeeOther)
}

func (h *HoldingsHandler) searchStocks(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	if q == "" {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("[]"))
		return
	}
	results, err := h.pg.SearchStocks(r.Context(), q)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if results == nil {
		results = []domain.StockSuggestion{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}
