package jquants

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const baseURL = "https://api.jquants.com/v2"

// Client は J-Quants API v2 クライアントです。
// 認証は x-api-key ヘッダーで行います。
type Client struct {
	apiKey string
	http   *http.Client
}

func New(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
		http:   &http.Client{Timeout: 30 * time.Second},
	}
}

// ── HTTP ──────────────────────────────────────────────────────────────────────

func (c *Client) get(ctx context.Context, path string, out any) error {
	req, _ := http.NewRequestWithContext(ctx, "GET", baseURL+path, nil)
	req.Header.Set("x-api-key", c.apiKey)

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("jquants GET %s: %w", path, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("jquants GET %s: status %d: %s", path, resp.StatusCode, b)
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

// ── API メソッド ───────────────────────────────────────────────────────────────

// DailyQuotes は指定日（YYYY-MM-DD）の全銘柄終値を返します。
// ページネーションを自動処理します。
func (c *Client) DailyQuotes(ctx context.Context, date string) ([]DailyQuote, error) {
	var all []DailyQuote
	path := "/equities/bars/daily?date=" + url.QueryEscape(date)

	for {
		var resp struct {
			Data          []DailyQuote `json:"data"`
			PaginationKey string       `json:"pagination_key"`
		}
		if err := c.get(ctx, path, &resp); err != nil {
			return nil, err
		}
		all = append(all, resp.Data...)
		if resp.PaginationKey == "" {
			break
		}
		path = "/equities/bars/daily?date=" + url.QueryEscape(date) + "&pagination_key=" + url.QueryEscape(resp.PaginationKey)
	}
	return all, nil
}

// DailyQuoteByCode は指定銘柄コードの指定日終値を返します。
func (c *Client) DailyQuoteByCode(ctx context.Context, code, date string) ([]DailyQuote, error) {
	var resp struct {
		Data []DailyQuote `json:"data"`
	}
	path := "/equities/bars/daily?code=" + url.QueryEscape(code) + "&date=" + url.QueryEscape(date)
	if err := c.get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

// EquitiesMaster は指定日基準の上場銘柄情報を返します。
func (c *Client) EquitiesMaster(ctx context.Context, date string) ([]EquityMaster, error) {
	var all []EquityMaster
	path := "/equities/master"
	if date != "" {
		path += "?date=" + url.QueryEscape(date)
	}

	for {
		var resp struct {
			Data          []EquityMaster `json:"data"`
			PaginationKey string         `json:"pagination_key"`
		}
		if err := c.get(ctx, path, &resp); err != nil {
			return nil, err
		}
		all = append(all, resp.Data...)
		if resp.PaginationKey == "" {
			break
		}
		if date != "" {
			path = "/equities/master?date=" + url.QueryEscape(date) + "&pagination_key=" + url.QueryEscape(resp.PaginationKey)
		} else {
			path = "/equities/master?pagination_key=" + url.QueryEscape(resp.PaginationKey)
		}
	}

	return all, nil
}

// FinancialStatements は指定銘柄の決算情報一覧を返します。
func (c *Client) FinancialStatements(ctx context.Context, code string) ([]FinancialStatement, error) {
	var all []FinancialStatement
	path := "/fins/statements?code=" + code

	for {
		var resp struct {
			Statements    []FinancialStatement `json:"statements"`
			PaginationKey string               `json:"pagination_key"`
		}
		if err := c.get(ctx, path, &resp); err != nil {
			return nil, err
		}
		all = append(all, resp.Statements...)
		if resp.PaginationKey == "" {
			break
		}
		path = "/fins/statements?code=" + code + "&pagination_key=" + resp.PaginationKey
	}
	return all, nil
}

// FinancialStatementsByDate は指定日（YYYY-MM-DD）に開示された全銘柄の決算情報を返します。
func (c *Client) FinancialStatementsByDate(ctx context.Context, date string) ([]FinancialStatement, error) {
	var all []FinancialStatement
	path := "/fins/statements?date=" + date

	for {
		var resp struct {
			Statements    []FinancialStatement `json:"statements"`
			PaginationKey string               `json:"pagination_key"`
		}
		if err := c.get(ctx, path, &resp); err != nil {
			return nil, err
		}
		all = append(all, resp.Statements...)
		if resp.PaginationKey == "" {
			break
		}
		path = "/fins/statements?date=" + date + "&pagination_key=" + resp.PaginationKey
	}
	return all, nil
}

// EarningsAnnouncements は直近の決算発表予定一覧を返します。
// v2 エンドポイント: /fins/earnings_announcement_dates_times
func (c *Client) EarningsAnnouncements(ctx context.Context) ([]EarningsAnnouncement, error) {
	var all []EarningsAnnouncement
	path := "/fins/earnings_announcement_dates_times"

	for {
		var resp struct {
			Announcements []EarningsAnnouncement `json:"earnings_announcement_dates_times"`
			PaginationKey string                 `json:"pagination_key"`
		}
		if err := c.get(ctx, path, &resp); err != nil {
			return nil, err
		}
		all = append(all, resp.Announcements...)
		if resp.PaginationKey == "" {
			break
		}
		path = "/fins/earnings_announcement_dates_times?pagination_key=" + resp.PaginationKey
	}
	return all, nil
}
