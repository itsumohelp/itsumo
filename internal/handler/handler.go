package handler

import (
	"context"
	"fmt"
	"hash/fnv"
	"html/template"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/kusakari/itsumo/internal/domain"
	"github.com/kusakari/itsumo/internal/store"
	"github.com/kusakari/itsumo/web"
)

type Handler struct {
	repo store.Repository
}

var urlPattern = regexp.MustCompile(`https?://[^\s<>"']+`)

func New(repo store.Repository) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /{$}", h.listTrades)
	mux.HandleFunc("GET /trades/new", h.newTradeForm)
	mux.HandleFunc("POST /trades", h.createTrade)
	mux.HandleFunc("GET /prices/latest", h.latestPrice)
	mux.HandleFunc("GET /prices/check", h.checkPrices)
	mux.HandleFunc("GET /companies/options", h.companyOptions)
	mux.HandleFunc("GET /trades/{id}", h.getTrade)
	mux.HandleFunc("POST /trades/{id}/events", h.createEvent)
	mux.HandleFunc("GET /trades/{id}/close", h.closeTradeForm)
	mux.HandleFunc("POST /trades/{id}/close", h.closeTrade)
	mux.HandleFunc("POST /votes/trade/{id}/{target}", h.voteTrade)
	mux.HandleFunc("POST /votes/event/{id}", h.voteEvent)
}

// ── Template helpers ──────────────────────────────────────────────────────────

func funcMap() template.FuncMap {
	return template.FuncMap{
		"not": func(b bool) bool { return !b },
		"formatPrice": func(f float64) string {
			return fmt.Sprintf("¥%.0f", f)
		},
		"formatVolume": func(f float64) string {
			switch {
			case f >= 1_000_000:
				return fmt.Sprintf("%.1fM", f/1_000_000)
			case f >= 1_000:
				return fmt.Sprintf("%.1fK", f/1_000)
			default:
				return fmt.Sprintf("%.0f", f)
			}
		},
		"formatDate": func(t time.Time) string {
			if t.IsZero() {
				// ── Template helpers ──────────────────────────────────────────────────────────
			}
			return t.Format("2006-01-02")
		},
		"today": func() string {
			return time.Now().Format("2006-01-02")
		},
		"timeAgo": func(t time.Time) string {
			d := time.Since(t)
			days := int(d.Hours() / 24)
			switch {
			case days == 0:
				return "今日"
			case days < 7:
				return fmt.Sprintf("%d日前", days)
			case days < 30:
				return fmt.Sprintf("%d週間前", days/7)
			default:
				return fmt.Sprintf("%dヶ月前", days/30)
			}
		},
		"kindLabel": func(k domain.EventKind) string {
			switch k {
			case domain.KindEarnings:
				return "決算"
			case domain.KindNews:
				return "ニュース"
			default:
				return "メモ"
			}
		},
		"pnlClass": func(p float64) string {
			if p >= 0 {
				return "pnl-pos"
			}
			return "pnl-neg"
		},
		"dict": func(pairs ...any) map[string]any {
			m := make(map[string]any, len(pairs)/2)
			for i := 0; i+1 < len(pairs); i += 2 {
				if k, ok := pairs[i].(string); ok {
					m[k] = pairs[i+1]
				}
			}
			return m
		},
		"focusLabel": func(f domain.EntryFocus) string {
			switch f {
			case domain.FocusValue:
				return "割安感（PERにフォーカス）"
			case domain.FocusMarket:
				return "地合（TOPIXにフォーカス）"
			default:
				return "将来的な価格上昇（株価にフォーカス）"
			}
		},
		"focusParams": func(f domain.EntryFocus) []string {
			switch f {
			case domain.FocusValue:
				return []string{"PER", "PBR", "EPS見通し", "同業比較"}
			case domain.FocusMarket:
				return []string{"TOPIX騰落", "出来高", "セクター相対強弱", "金利・為替"}
			default:
				return []string{"株価推移", "高値/安値更新", "出来高変化", "需給イベント"}
			}
		},
		"focusTimelineHint": func(f domain.EntryFocus) string {
			switch f {
			case domain.FocusValue:
				return "PER/PBR の変化、EPS修正、同業比較の変化を中心に記録"
			case domain.FocusMarket:
				return "TOPIXやセクター地合、金利/為替など外部環境の変化を中心に記録"
			default:
				return "株価・出来高・需給イベントなど、価格アクション中心に記録"
			}
		},
		"commentHTML": func(body string) template.HTML {
			return linkifyComment(body)
		},
		"isBotComment": func(body string) bool {
			return strings.HasPrefix(body, "[BOT]")
		},
		"stripBotPrefix": func(body string) string {
			return strings.TrimSpace(strings.TrimPrefix(body, "[BOT]"))
		},
	}
}

func linkifyComment(body string) template.HTML {
	if body == "" {
		return ""
	}
	matches := urlPattern.FindAllStringIndex(body, -1)
	if len(matches) == 0 {
		escaped := template.HTMLEscapeString(body)
		escaped = strings.ReplaceAll(escaped, "\n", "<br>")
		return template.HTML(escaped)
	}

	var b strings.Builder
	last := 0
	for _, m := range matches {
		if m[0] > last {
			b.WriteString(template.HTMLEscapeString(body[last:m[0]]))
		}
		u := body[m[0]:m[1]]
		escURL := template.HTMLEscapeString(u)
		b.WriteString(`<a href="`)
		b.WriteString(escURL)
		b.WriteString(`" target="_blank" rel="noopener">`)
		b.WriteString(escURL)
		b.WriteString(`</a>`)
		last = m[1]
	}
	if last < len(body) {
		b.WriteString(template.HTMLEscapeString(body[last:]))
	}
	out := strings.ReplaceAll(b.String(), "\n", "<br>")
	return template.HTML(out)
}

var partials = []string{
	"templates/partials/vote_buttons.html",
	"templates/partials/company_picker.html",
	"templates/partials/company_list.html",
}

func (h *Handler) render(w http.ResponseWriter, page string, data any) {
	files := append([]string{"templates/base.html", "templates/" + page + ".html"}, partials...)
	tmpl, err := template.New("").Funcs(funcMap()).ParseFS(web.Templates, files...)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.ExecuteTemplate(w, "base", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *Handler) renderPartial(w http.ResponseWriter, name string, data any) {
	tmpl, err := template.New(name).Funcs(funcMap()).ParseFS(web.Templates, "templates/partials/"+name+".html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.ExecuteTemplate(w, name, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// ── Trades ────────────────────────────────────────────────────────────────────

type listData struct {
	Open   []*domain.Trade
	Closed []*domain.Trade
}

func (h *Handler) listTrades(w http.ResponseWriter, r *http.Request) {
	trades, err := h.repo.ListTrades(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	data := &listData{}
	for _, t := range trades {
		if t.Status == domain.StatusOpen {
			data.Open = append(data.Open, t)
		} else {
			data.Closed = append(data.Closed, t)
		}
	}
	h.render(w, "trade_list", data)
}

type companyItem struct {
	Code     string
	Name     string
	Industry string
	Icon     string
}

type industryItem struct {
	Name string
	Icon string
}

type companyPickerData struct {
	Industries []industryItem
	Industry   string
	Keyword    string
	Companies  []companyItem
	HasQuery   bool
	EntryTags  []string
}

func (h *Handler) newTradeForm(w http.ResponseWriter, r *http.Request) {
	data, err := h.loadCompanyPickerData(r.Context(), "", "")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	h.render(w, "trade_new", data)
}

func (h *Handler) companyOptions(w http.ResponseWriter, r *http.Request) {
	industry := strings.TrimSpace(r.URL.Query().Get("industry"))
	keyword := strings.TrimSpace(r.URL.Query().Get("q"))
	data, err := h.loadCompanyPickerData(r.Context(), industry, keyword)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	h.renderPartial(w, "company_list", data)
}

func (h *Handler) loadCompanyPickerData(ctx context.Context, industry, keyword string) (*companyPickerData, error) {
	industries, err := h.repo.ListCompanyIndustries(ctx)
	if err != nil {
		return nil, err
	}
	companies, err := h.repo.ListCompanies(ctx, industry, keyword, 200)
	if err != nil {
		return nil, err
	}
	items := make([]companyItem, 0, len(companies))
	for _, c := range companies {
		items = append(items, companyItem{
			Code:     c.Code,
			Name:     c.Name,
			Industry: c.Industry,
			Icon:     companyIcon(c.Name, c.Industry),
		})
	}
	sort.Strings(industries)
	industryItems := make([]industryItem, 0, len(industries))
	for _, ind := range industries {
		industryItems = append(industryItems, industryItem{
			Name: ind,
			Icon: companyIcon("", ind),
		})
	}
	return &companyPickerData{
		Industries: industryItems,
		Industry:   industry,
		Keyword:    keyword,
		Companies:  items,
		HasQuery:   strings.TrimSpace(industry) != "" || strings.TrimSpace(keyword) != "",
		EntryTags:  domain.EntryTagOptions,
	}, nil
}

func companyIcon(name, industry string) string {
	if icon, ok := iconByIndustry(industry); ok {
		return icon
	}

	x := name + " " + industry
	switch {
	case hasAny(x, "銀行", "フィナンシャル", "証券"):
		return "🏦"
	case hasAny(x, "自動車", "モーター"):
		return "🚗"
	case hasAny(x, "電力", "ガス", "エネルギー"):
		return "⚡"
	case hasAny(x, "医薬", "ヘルス", "薬"):
		return "💊"
	case hasAny(x, "通信", "ソフト", "情報"):
		return "💻"
	case hasAny(x, "食品", "飲料"):
		return "🍽️"
	case strings.Contains(x, "不動産"):
		return "🏢"
	case hasAny(x, "運輸", "海運", "空運"):
		return "🚚"
	default:
		return iconFromHash(name + "|" + industry)
	}
}

func iconByIndustry(industry string) (string, bool) {
	s := strings.TrimSpace(industry)
	switch {
	case hasAny(s, "水産", "農林"):
		return "🐟", true
	case hasAny(s, "鉱業"):
		return "⛏️", true
	case hasAny(s, "建設"):
		return "🏗️", true
	case hasAny(s, "食料品"):
		return "🍽️", true
	case hasAny(s, "繊維"):
		return "🧵", true
	case hasAny(s, "パルプ", "紙"):
		return "📄", true
	case hasAny(s, "化学"):
		return "🧪", true
	case hasAny(s, "医薬"):
		return "💊", true
	case hasAny(s, "石油", "石炭"):
		return "🛢️", true
	case hasAny(s, "ゴム"):
		return "🛞", true
	case hasAny(s, "ガラス", "土石"):
		return "🧱", true
	case hasAny(s, "鉄鋼"):
		return "🔩", true
	case hasAny(s, "非鉄金属"):
		return "🪙", true
	case hasAny(s, "金属製品"):
		return "🛠️", true
	case hasAny(s, "機械"):
		return "⚙️", true
	case hasAny(s, "電気機器"):
		return "💡", true
	case hasAny(s, "輸送用機器"):
		return "🚗", true
	case hasAny(s, "精密機器"):
		return "🔬", true
	case hasAny(s, "その他製品"):
		return "🎁", true
	case hasAny(s, "電気・ガス"):
		return "⚡", true
	case hasAny(s, "陸運"):
		return "🚆", true
	case hasAny(s, "海運"):
		return "🚢", true
	case hasAny(s, "空運"):
		return "✈️", true
	case hasAny(s, "倉庫"):
		return "📦", true
	case hasAny(s, "情報・通信"):
		return "💻", true
	case hasAny(s, "卸売"):
		return "📚", true
	case hasAny(s, "小売"):
		return "🛒", true
	case hasAny(s, "銀行"):
		return "🏦", true
	case hasAny(s, "証券", "商品先物"):
		return "📈", true
	case hasAny(s, "保険"):
		return "🛡️", true
	case hasAny(s, "その他金融"):
		return "💳", true
	case hasAny(s, "不動産"):
		return "🏢", true
	case hasAny(s, "サービス"):
		return "🧰", true
	default:
		return "", false
	}
}

func hasAny(s string, keywords ...string) bool {
	for _, k := range keywords {
		if strings.Contains(s, k) {
			return true
		}
	}
	return false
}

func iconFromHash(seed string) string {
	icons := []string{"🌟", "🧩", "🛰️", "🧭", "🎯", "🔷", "🟢", "🟠", "🔶", "🧠", "🧿", "🔋", "🪄", "🧱", "🧷", "🗂️", "🧮", "🧰", "🪙", "📌"}
	h := fnv.New32a()
	_, _ = h.Write([]byte(seed))
	return icons[int(h.Sum32())%len(icons)]
}

func (h *Handler) createTrade(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	entryPrice, err := strconv.ParseFloat(r.FormValue("entry_price"), 64)
	if err != nil {
		http.Error(w, "invalid entry price", http.StatusBadRequest)
		return
	}
	entryDate, err := time.Parse("2006-01-02", r.FormValue("entry_date"))
	if err != nil {
		entryDate = time.Now()
	}
	entryTags := r.Form["entry_tags"]
	t := &domain.Trade{
		Ticker:         r.FormValue("ticker"),
		Name:           r.FormValue("name"),
		Status:         domain.StatusOpen,
		EntryFocus:     parseEntryFocus(r.FormValue("entry_focus")),
		EntryTags:      entryTags,
		EntryPrice:     entryPrice,
		EntryDate:      entryDate,
		EntryRationale: r.FormValue("entry_rationale"),
	}
	if err := h.repo.CreateTrade(r.Context(), t); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/trades/"+t.ID, http.StatusSeeOther)
}

func parseEntryFocus(v string) domain.EntryFocus {
	switch domain.EntryFocus(strings.TrimSpace(v)) {
	case domain.FocusValue:
		return domain.FocusValue
	case domain.FocusMarket:
		return domain.FocusMarket
	default:
		return domain.FocusPriceUp
	}
}

type tradeDetailData struct {
	Trade  *domain.Trade
	Events []*domain.Event
}

func (h *Handler) getTrade(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	t, err := h.repo.GetTrade(r.Context(), id)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	events, err := h.repo.ListEvents(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	events, err = h.ensureBotEarningsEvent(r.Context(), t, events)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	h.render(w, "trade_detail", &tradeDetailData{Trade: t, Events: events})
}

func (h *Handler) ensureBotEarningsEvent(ctx context.Context, t *domain.Trade, events []*domain.Event) ([]*domain.Event, error) {
	var ann *domain.EarningsAnnouncement
	var err error
	for _, code := range tickerCandidates(normalizeTicker(t.Ticker)) {
		ann, err = h.repo.GetNextEarningsAnnouncement(ctx, code)
		if err != nil {
			return events, err
		}
		if ann != nil {
			break
		}
	}
	if ann == nil {
		return events, nil
	}

	body := fmt.Sprintf("[BOT] 次回決算予定: %s", ann.Date.Format("2006-01-02"))
	if ann.FiscalQuarter != "" {
		body += " / " + ann.FiscalQuarter
	}
	for _, e := range events {
		if e.Kind == domain.KindEarnings && e.Body == body {
			return events, nil
		}
	}

	e := &domain.Event{
		TradeID:   t.ID,
		Kind:      domain.KindEarnings,
		Body:      body,
		EventDate: ann.Date,
	}
	if err := h.repo.CreateEvent(ctx, e); err != nil {
		return events, err
	}
	events = append(events, e)
	sort.Slice(events, func(i, j int) bool {
		return events[i].EventDate.Before(events[j].EventDate)
	})
	return events, nil
}

func (h *Handler) closeTradeForm(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	t, err := h.repo.GetTrade(r.Context(), id)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	h.render(w, "trade_close", t)
}

func (h *Handler) closeTrade(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	id := r.PathValue("id")
	exitPrice, err := strconv.ParseFloat(r.FormValue("exit_price"), 64)
	if err != nil {
		http.Error(w, "invalid exit price", http.StatusBadRequest)
		return
	}
	exitDate, err := time.Parse("2006-01-02", r.FormValue("exit_date"))
	if err != nil {
		exitDate = time.Now()
	}
	if err := h.repo.CloseTrade(r.Context(), id, exitPrice, exitDate, r.FormValue("exit_rationale")); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/trades/"+id, http.StatusSeeOther)
}

type latestPriceData struct {
	Ticker string
	Code   string
	Price  *domain.DailyPrice
	Found  bool
	Error  string
}

const checkPricesLimit = 30

type checkPricesData struct {
	Code   string
	Prices []*domain.DailyPrice
	Found  bool
	Error  string
}

func (h *Handler) checkPrices(w http.ResponseWriter, r *http.Request) {
	code := strings.ToUpper(strings.TrimSpace(r.URL.Query().Get("code")))
	if code == "" {
		h.render(w, "prices_check", &checkPricesData{})
		return
	}

	var prices []*domain.DailyPrice
	var lastErr string
	for _, c := range tickerCandidates(code) {
		pp, err := h.repo.ListPricesByCode(r.Context(), c, checkPricesLimit)
		if err != nil {
			lastErr = "データ取得に失敗しました"
			continue
		}
		if len(pp) > 0 {
			code = c
			prices = pp
			break
		}
	}

	if len(prices) == 0 && lastErr == "" {
		lastErr = "該当する株価データが見つかりません"
	}
	h.render(w, "prices_check", &checkPricesData{
		Code:   code,
		Prices: prices,
		Found:  len(prices) > 0,
		Error:  lastErr,
	})
}

func (h *Handler) latestPrice(w http.ResponseWriter, r *http.Request) {
	ticker := normalizeTicker(r.URL.Query().Get("ticker"))
	if ticker == "" {
		h.renderPartial(w, "price_preview", &latestPriceData{})
		return
	}

	for _, code := range tickerCandidates(ticker) {
		p, err := h.repo.GetLatestPrice(r.Context(), code)
		if err != nil {
			h.renderPartial(w, "price_preview", &latestPriceData{Ticker: ticker, Error: "株価の取得に失敗しました"})
			return
		}
		if p != nil {
			h.renderPartial(w, "price_preview", &latestPriceData{
				Ticker: ticker,
				Code:   code,
				Price:  p,
				Found:  true,
			})
			return
		}
	}

	h.renderPartial(w, "price_preview", &latestPriceData{Ticker: ticker, Error: "該当する株価データが見つかりません"})
}

func normalizeTicker(s string) string {
	s = strings.TrimSpace(s)
	return strings.ToUpper(strings.ReplaceAll(s, " ", ""))
}

func tickerCandidates(ticker string) []string {
	seen := map[string]struct{}{}
	add := func(c string, out *[]string) {
		if c == "" {
			return
		}
		if _, ok := seen[c]; ok {
			return
		}
		seen[c] = struct{}{}
		*out = append(*out, c)
	}

	out := make([]string, 0, 3)
	add(ticker, &out)
	if len(ticker) == 4 {
		add(ticker+"0", &out)
	}
	if len(ticker) == 5 {
		add(ticker[:4], &out)
	}
	return out
}

// ── Events ────────────────────────────────────────────────────────────────────

func (h *Handler) createEvent(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	tradeID := r.PathValue("id")
	body := strings.TrimSpace(r.FormValue("body"))
	if body == "" {
		http.Error(w, "comment is required", http.StatusBadRequest)
		return
	}
	url := extractFirstURL(body)
	e := &domain.Event{
		TradeID:   tradeID,
		Kind:      domain.KindNote,
		Body:      body,
		URL:       url,
		EventDate: time.Now(),
	}
	if err := h.repo.CreateEvent(r.Context(), e); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/trades/"+tradeID, http.StatusSeeOther)
}

func extractFirstURL(body string) string {
	return urlPattern.FindString(body)
}

// ── Votes ─────────────────────────────────────────────────────────────────────

type voteData struct {
	ID   string
	URL  string
	Up   int
	Down int
}

func (h *Handler) voteTrade(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	id := r.PathValue("id")
	target := r.PathValue("target") // "entry" or "exit"
	dir := r.FormValue("dir")       // "up" or "down"

	votes, err := h.repo.VoteTrade(r.Context(), id, target, dir)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	h.renderPartial(w, "vote_buttons", &voteData{
		ID:   target + "-" + id,
		URL:  "/votes/trade/" + id + "/" + target,
		Up:   votes.Up,
		Down: votes.Down,
	})
}

func (h *Handler) voteEvent(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	id := r.PathValue("id")
	dir := r.FormValue("dir")

	votes, err := h.repo.VoteEvent(r.Context(), id, dir)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	h.renderPartial(w, "vote_buttons", &voteData{
		ID:   "event-" + id,
		URL:  "/votes/event/" + id,
		Up:   votes.Up,
		Down: votes.Down,
	})
}
