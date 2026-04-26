package chart

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"time"

	"github.com/jhonroun/tradingview-mcp-go/internal/cdp"
	"github.com/jhonroun/tradingview-mcp-go/internal/mcp"
	tv "github.com/jhonroun/tradingview-mcp-go/internal/tradingview"
)

var reEM = regexp.MustCompile(`</?em>`)

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

// SymbolSearch queries the TradingView symbol search API (max 15 results).
func SymbolSearch(query, searchType, exchange string) ([]SymbolSearchResult, error) {
	params := url.Values{
		"text":        {query},
		"type":        {searchType},
		"exchange":    {exchange},
		"hl":          {"1"},
		"lang":        {"en"},
		"domain":      {"production"},
		"search_type": {"undefined"},
	}
	searchURL := "https://symbol-search.tradingview.com/symbol_search/v3/?" + params.Encode()

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

	var data struct {
		Symbols []struct {
			Symbol      string `json:"symbol"`
			Description string `json:"description"`
			Type        string `json:"type"`
			Exchange    string `json:"exchange"`
		} `json:"symbols"`
	}
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, fmt.Errorf("parse symbol search: %w", err)
	}

	strip := func(s string) string { return reEM.ReplaceAllString(s, "") }

	results := make([]SymbolSearchResult, 0, len(data.Symbols))
	for _, s := range data.Symbols {
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
	return results, nil
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
			results, err := SymbolSearch(p.Query, p.Type, p.Exchange)
			if err != nil {
				return map[string]interface{}{"success": false, "error": err.Error()}, nil
			}
			return map[string]interface{}{"success": true, "results": results, "count": len(results)}, nil
		},
	})
}
