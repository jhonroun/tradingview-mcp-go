package chart

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/jhonroun/tradingview-mcp-go/internal/cdp"
	"github.com/jhonroun/tradingview-mcp-go/internal/mcp"
	tv "github.com/jhonroun/tradingview-mcp-go/internal/tradingview"
)

var reEM = regexp.MustCompile(`</?em>`)

const symbolSearchEndpoint = "https://symbol-search.tradingview.com/symbol_search/v3/"

// SymbolInfo returns extended info for the current chart symbol via chart.symbolExt().
func SymbolInfo() (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	const expr = `(function() { return ` + tv.ChartAPI + `.symbolExt(); })()`

	var raw json.RawMessage
	err := cdp.WithSession(ctx, func(c *cdp.Client, _ *cdp.Target) error {
		var err error
		raw, err = c.Evaluate(ctx, expr, false)
		return err
	})
	if err != nil {
		return nil, err
	}
	var info map[string]interface{}
	if err := json.Unmarshal(raw, &info); err != nil {
		return nil, fmt.Errorf("parse symbol info: %w", err)
	}
	// Ensure required contract fields are always present (empty string sentinel).
	for _, key := range []string{"symbol", "exchange", "description", "type"} {
		if _, ok := info[key]; !ok {
			info[key] = ""
		}
	}
	info["success"] = true
	return info, nil
}

// SymbolSearchResult is one hit from the TradingView symbol search API.
type SymbolSearchResult struct {
	Symbol      string `json:"symbol"`
	Description string `json:"description"`
	Type        string `json:"type"`
	Exchange    string `json:"exchange"`
}

type symbolSearchAPIResult struct {
	Symbol      string `json:"symbol"`
	Description string `json:"description"`
	Type        string `json:"type"`
	Exchange    string `json:"exchange"`
}

// SymbolSearch queries the TradingView symbol search API (max 15 results).
func SymbolSearch(query, searchType, exchange string) ([]SymbolSearchResult, error) {
	query = strings.TrimSpace(query)
	searchType = strings.TrimSpace(searchType)
	exchange = strings.TrimSpace(exchange)
	if query == "" {
		return nil, fmt.Errorf("query is required")
	}

	searchURL := buildSymbolSearchURL(query, exchange)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, searchURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Origin", "https://www.tradingview.com")
	req.Header.Set("Referer", "https://www.tradingview.com/")
	req.Header.Set("User-Agent", "Mozilla/5.0 TradingView-MCP-Go")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("symbol search: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("symbol search: TradingView API returned %d: %s", resp.StatusCode, bodySnippet(body))
	}

	apiResults, err := parseSymbolSearchResults(body)
	if err != nil {
		return nil, fmt.Errorf("parse symbol search: %w", err)
	}

	return normalizeSymbolSearchResults(apiResults, searchType, exchange), nil
}

func SymbolSearchResponse(query, searchType, exchange string) (map[string]interface{}, error) {
	results, err := SymbolSearch(query, searchType, exchange)
	if err != nil {
		return nil, err
	}
	return buildSymbolSearchResponse(query, searchType, exchange, results), nil
}

func buildSymbolSearchResponse(query, searchType, exchange string, results []SymbolSearchResult) map[string]interface{} {
	if results == nil {
		results = []SymbolSearchResult{}
	}
	result := map[string]interface{}{
		"success": true,
		"status":  "ok",
		"source":  "tradingview_symbol_search_api",
		"query":   strings.TrimSpace(query),
		"results": results,
		"count":   len(results),
	}
	if strings.TrimSpace(searchType) != "" {
		result["type_filter"] = strings.TrimSpace(searchType)
	}
	if strings.TrimSpace(exchange) != "" {
		result["exchange_filter"] = strings.TrimSpace(exchange)
	}
	if len(results) == 0 {
		result["status"] = "no_results"
		result["reason"] = "TradingView symbol search API returned no results after applying requested filters."
	}
	return result
}

func buildSymbolSearchURL(query, exchange string) string {
	params := url.Values{
		"text":   {query},
		"hl":     {"1"},
		"lang":   {"en"},
		"domain": {"production"},
	}
	if exchange != "" {
		params.Set("exchange", exchange)
	}
	return symbolSearchEndpoint + "?" + params.Encode()
}

func parseSymbolSearchResults(body []byte) ([]symbolSearchAPIResult, error) {
	body = []byte(strings.TrimSpace(string(body)))
	if len(body) == 0 {
		return nil, nil
	}
	if body[0] == '[' {
		var legacy []symbolSearchAPIResult
		if err := json.Unmarshal(body, &legacy); err != nil {
			return nil, err
		}
		return legacy, nil
	}
	var data struct {
		Symbols []symbolSearchAPIResult `json:"symbols"`
	}
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, err
	}
	return data.Symbols, nil
}

func normalizeSymbolSearchResults(apiResults []symbolSearchAPIResult, searchType, exchange string) []SymbolSearchResult {
	strip := func(s string) string { return reEM.ReplaceAllString(s, "") }
	searchType = strings.TrimSpace(strings.ToLower(searchType))
	exchange = strings.TrimSpace(strings.ToLower(exchange))

	results := make([]SymbolSearchResult, 0, len(apiResults))
	for _, s := range apiResults {
		if searchType != "" && strings.ToLower(s.Type) != searchType {
			continue
		}
		if exchange != "" && strings.ToLower(s.Exchange) != exchange {
			continue
		}
		if len(results) >= 15 {
			break
		}
		results = append(results, SymbolSearchResult{
			Symbol:      strip(s.Symbol),
			Description: strip(s.Description),
			Type:        s.Type,
			Exchange:    s.Exchange,
		})
	}
	return results
}

func bodySnippet(body []byte) string {
	const max = 300
	text := strings.TrimSpace(string(body))
	if len(text) > max {
		return text[:max] + "...(truncated)"
	}
	return text
}

func registerSymbolTools(reg *mcp.Registry) {
	reg.Register(mcp.ToolDef{
		Name:        "symbol_info",
		Description: "Get extended info for the current chart symbol: full name, exchange, type, description",
		Schema:      mcp.InputSchema{Type: "object"},
		Handler: func(args json.RawMessage) (interface{}, error) {
			result, err := SymbolInfo()
			if err != nil {
				return map[string]interface{}{"success": false, "error": err.Error()}, nil
			}
			return result, nil
		},
	})

	reg.Register(mcp.ToolDef{
		Name:        "symbol_search",
		Description: "Search TradingView symbols by name or description. Returns up to 15 matches.",
		Schema: mcp.InputSchema{
			Type: "object",
			Properties: map[string]mcp.PropertySchema{
				"query":    {Type: "string", Description: "Search text (ticker or name)"},
				"type":     {Type: "string", Description: "Filter by type: stock, forex, crypto, futures, index (optional)"},
				"exchange": {Type: "string", Description: "Filter by exchange, e.g. NASDAQ, NYSE (optional)"},
			},
			Required: []string{"query"},
		},
		Handler: func(args json.RawMessage) (interface{}, error) {
			var p struct {
				Query    string `json:"query"`
				Type     string `json:"type"`
				Exchange string `json:"exchange"`
			}
			if err := json.Unmarshal(args, &p); err != nil {
				return map[string]interface{}{"success": false, "error": err.Error()}, nil
			}
			result, err := SymbolSearchResponse(p.Query, p.Type, p.Exchange)
			if err != nil {
				return map[string]interface{}{"success": false, "error": err.Error()}, nil
			}
			return result, nil
		},
	})
}
