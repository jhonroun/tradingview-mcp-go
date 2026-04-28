// Package data implements all read-only chart data tools.
package data

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/jhonroun/tradingview-mcp-go/internal/cdp"
	"github.com/jhonroun/tradingview-mcp-go/internal/mcp"
	tv "github.com/jhonroun/tradingview-mcp-go/internal/tradingview"
)

const (
	maxOHLCVBars = 500
	maxTrades    = 20
	maxOrders    = 50
)

// ---------- helpers ----------

func withSession(ctx context.Context, fn func(*cdp.Client) (json.RawMessage, error)) (json.RawMessage, error) {
	var result json.RawMessage
	err := cdp.WithSession(ctx, func(c *cdp.Client, _ *cdp.Target) error {
		var err error
		result, err = fn(c)
		return err
	})
	return result, err
}

// buildGraphicsJS mirrors the JS helper from data.js.
func buildGraphicsJS(collectionName, mapKey, filter string) string {
	return `(function() {
		var chart = window.TradingViewApi._activeChartWidgetWV.value()._chartWidget;
		var model = chart.model();
		var sources = model.model().dataSources();
		var results = [];
		var filter = ` + tv.SafeString(filter) + `;
		for (var si = 0; si < sources.length; si++) {
			var s = sources[si];
			if (!s.metaInfo) continue;
			try {
				var meta = s.metaInfo();
				var name = meta.description || meta.shortDescription || '';
				if (!name) continue;
				if (filter && name.indexOf(filter) === -1) continue;
				var g = s._graphics;
				if (!g || !g._primitivesCollection) continue;
				var pc = g._primitivesCollection;
				var items = [];
				try {
					var outer = pc.` + collectionName + `;
					if (outer) {
						var inner = outer.get('` + mapKey + `');
						if (inner) {
							var coll = inner.get(false);
							if (coll && coll._primitivesDataById && coll._primitivesDataById.size > 0) {
								coll._primitivesDataById.forEach(function(v, id) { items.push({id: id, raw: v}); });
							}
						}
					}
				} catch(e) {}
				if (items.length === 0 && '` + collectionName + `' === 'dwgtablecells') {
					try {
						var tcOuter = pc.dwgtablecells;
						if (tcOuter) {
							var tcColl = tcOuter.get('tableCells');
							if (tcColl && tcColl._primitivesDataById && tcColl._primitivesDataById.size > 0) {
								tcColl._primitivesDataById.forEach(function(v, id) { items.push({id: id, raw: v}); });
							}
						}
					} catch(e) {}
				}
				if (items.length > 0) results.push({name: name, count: items.length, items: items});
			} catch(e) {}
		}
		return results;
	})()`
}

// ---------- OHLCV ----------

type Bar struct {
	Time   int64   `json:"time"`
	Open   float64 `json:"open"`
	High   float64 `json:"high"`
	Low    float64 `json:"low"`
	Close  float64 `json:"close"`
	Volume float64 `json:"volume"`
}

func GetOhlcv(count int, summary bool) (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	limit := count
	if limit <= 0 {
		limit = 100
	}
	if limit > maxOHLCVBars {
		limit = maxOHLCVBars
	}

	expr := fmt.Sprintf(`(function() {
		var bars = `+tv.BarsPath+`;
		if (!bars || typeof bars.lastIndex !== 'function') return null;
		var result = [];
		var end = bars.lastIndex();
		var start = Math.max(bars.firstIndex(), end - %d + 1);
		for (var i = start; i <= end; i++) {
			var v = bars.valueAt(i);
			if (v) result.push({time: v[0], open: v[1], high: v[2], low: v[3], close: v[4], volume: v[5] || 0});
		}
		return {bars: result, total_bars: bars.size(), source: 'direct_bars'};
	})()`, limit)

	raw, err := withSession(ctx, func(c *cdp.Client) (json.RawMessage, error) {
		return c.Evaluate(ctx, expr, false)
	})
	if err != nil {
		return nil, err
	}
	if string(raw) == "null" {
		return nil, fmt.Errorf("could not extract OHLCV data; the chart may still be loading")
	}

	var data struct {
		Bars      []Bar  `json:"bars"`
		TotalBars int    `json:"total_bars"`
		Source    string `json:"source"`
	}
	if err := json.Unmarshal(raw, &data); err != nil {
		return nil, fmt.Errorf("parse ohlcv: %w", err)
	}
	if len(data.Bars) == 0 {
		return nil, fmt.Errorf("could not extract OHLCV data; the chart may still be loading")
	}

	if summary {
		bars := data.Bars
		first, last := bars[0], bars[len(bars)-1]
		hi, lo, volSum := -math.MaxFloat64, math.MaxFloat64, 0.0
		for _, b := range bars {
			if b.High > hi {
				hi = b.High
			}
			if b.Low < lo {
				lo = b.Low
			}
			volSum += b.Volume
		}
		n := float64(len(bars))
		chg := round2(last.Close - first.Open)
		chgPct := fmt.Sprintf("%.2f%%", round2((last.Close-first.Open)/first.Open*100))
		tail := bars
		if len(tail) > 5 {
			tail = tail[len(tail)-5:]
		}
		return map[string]interface{}{
			"success":     true,
			"bar_count":   len(bars),
			"period":      map[string]interface{}{"from": first.Time, "to": last.Time},
			"open":        first.Open,
			"close":       last.Close,
			"high":        round2(hi),
			"low":         round2(lo),
			"range":       round2(hi - lo),
			"change":      chg,
			"change_pct":  chgPct,
			"avg_volume":  math.Round(volSum / n),
			"last_5_bars": tail,
		}, nil
	}

	return map[string]interface{}{
		"success":         true,
		"bar_count":       len(data.Bars),
		"total_available": data.TotalBars,
		"source":          data.Source,
		"bars":            data.Bars,
	}, nil
}

func round2(v float64) float64 { return math.Round(v*100) / 100 }

// ---------- Quote ----------

func GetQuote(symbol string) (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	expr := fmt.Sprintf(`(function() {
		var api = `+tv.ChartAPI+`;
		var sym = %s;
		if (!sym) { try { sym = api.symbol(); } catch(e) {} }
		if (!sym) { try { sym = api.symbolExt().symbol; } catch(e) {} }
		var ext = {};
		try { ext = api.symbolExt() || {}; } catch(e) {}
		var bars = `+tv.BarsPath+`;
		var quote = { symbol: sym, bid: 0, ask: 0, change: 0, change_pct: 0 };
		if (bars && typeof bars.lastIndex === 'function') {
			var lastIdx = bars.lastIndex();
			var last = bars.valueAt(lastIdx);
			if (last) {
				quote.time   = last[0];
				quote.open   = last[1];
				quote.high   = last[2];
				quote.low    = last[3];
				quote.close  = last[4];
				quote.last   = last[4];
				quote.volume = last[5] || 0;
				// change vs previous bar
				var prev = bars.valueAt(lastIdx - 1);
				if (prev && prev[4]) {
					quote.change     = +((last[4] - prev[4]).toFixed(8));
					quote.change_pct = prev[4] !== 0
						? +((last[4] - prev[4]) / prev[4] * 100).toFixed(4)
						: 0;
				}
			}
		}
		try {
			var bidEl = document.querySelector('[class*="bid"] [class*="price"], [class*="dom-"] [class*="bid"]');
			var askEl = document.querySelector('[class*="ask"] [class*="price"], [class*="dom-"] [class*="ask"]');
			if (bidEl) quote.bid_display_value = bidEl.textContent.trim();
			if (askEl) quote.ask_display_value = askEl.textContent.trim();
		} catch(e) {}
		try {
			var hdr = document.querySelector('[class*="headerRow"] [class*="last-"]');
			if (hdr) quote.header_price_display_value = hdr.textContent.trim();
		} catch(e) {}
		if (ext.description) quote.description = ext.description;
		if (ext.exchange)    quote.exchange    = ext.exchange;
		if (ext.type)        quote.type        = ext.type;
		return quote;
	})()`, tv.SafeString(symbol))

	raw, err := withSession(ctx, func(c *cdp.Client) (json.RawMessage, error) {
		return c.Evaluate(ctx, expr, false)
	})
	if err != nil {
		return nil, err
	}

	var q map[string]interface{}
	if err := json.Unmarshal(raw, &q); err != nil {
		return nil, fmt.Errorf("parse quote: %w", err)
	}
	if q["last"] == nil && q["close"] == nil {
		return nil, fmt.Errorf("could not retrieve quote; the chart may still be loading")
	}
	if s, ok := q["bid_display_value"].(string); ok {
		if v, ok := ParseDisplayNumber(s); ok {
			q["bid"] = v
			q["bid_source"] = SourceTradingViewUIDOM
			q["bid_reliability"] = ReliabilityDisplayValueLocalizedUIString
		}
	}
	if s, ok := q["ask_display_value"].(string); ok {
		if v, ok := ParseDisplayNumber(s); ok {
			q["ask"] = v
			q["ask_source"] = SourceTradingViewUIDOM
			q["ask_reliability"] = ReliabilityDisplayValueLocalizedUIString
		}
	}
	if s, ok := q["header_price_display_value"].(string); ok {
		if v, ok := ParseDisplayNumber(s); ok {
			q["header_price"] = v
			q["header_price_source"] = SourceTradingViewUIDOM
			q["header_price_reliability"] = ReliabilityDisplayValueLocalizedUIString
		}
	}
	// Sentinel guarantees: bid/ask/change/change_pct are always numeric.
	for _, key := range []string{"bid", "ask", "change", "change_pct"} {
		if q[key] == nil {
			q[key] = float64(0)
		}
	}
	markQuoteBidAskAvailability(q)
	q["success"] = true
	return q, nil
}

func markQuoteBidAskAvailability(q map[string]interface{}) {
	bid, bidNumber := numberValue(q["bid"])
	ask, askNumber := numberValue(q["ask"])
	bidAvailable := bidNumber && bid > 0
	askAvailable := askNumber && ask > 0

	q["bidAvailable"] = bidAvailable
	q["askAvailable"] = askAvailable
	q["bidAskAvailable"] = bidAvailable && askAvailable

	if bidAvailable && askAvailable {
		if q["bid_source"] == nil {
			q["bid_source"] = "tradingview_quote"
		}
		if q["ask_source"] == nil {
			q["ask_source"] = "tradingview_quote"
		}
		return
	}

	q["sourceLimitation"] = "tradingview_bid_ask_unavailable"
	q["warning"] = "TradingView did not expose usable bid/ask for this symbol; bid and ask are zero sentinels and must not be treated as executable quotes."
	if isMOEXFuturesQuote(q) {
		q["sourceLimitation"] = "tradingview_moex_futures_bid_ask_unavailable"
		q["warning"] = "TradingView did not expose usable bid/ask for this MOEX futures symbol; bid and ask are zero sentinels and must not be treated as executable quotes."
	}
	if !bidAvailable {
		q["bid_source"] = "tradingview_unavailable"
	}
	if !askAvailable {
		q["ask_source"] = "tradingview_unavailable"
	}
}

func isMOEXFuturesQuote(q map[string]interface{}) bool {
	exchange := lowerString(q["exchange"])
	symbol := lowerString(q["symbol"])
	typ := lowerString(q["type"])
	return exchange == "rus" && (typ == "futures" || strings.Contains(symbol, "!"))
}

func lowerString(v interface{}) string {
	s, _ := v.(string)
	return strings.ToLower(strings.TrimSpace(s))
}

func numberValue(v interface{}) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, !math.IsNaN(n) && !math.IsInf(n, 0)
	case float32:
		f := float64(n)
		return f, !math.IsNaN(f) && !math.IsInf(f, 0)
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	case json.Number:
		f, err := n.Float64()
		return f, err == nil && !math.IsNaN(f) && !math.IsInf(f, 0)
	default:
		return 0, false
	}
}

// ---------- Study values ----------

// StudyPlot is one output line of an indicator (e.g. "RSI", "Signal", "Histogram").
// values[0] is the current bar; the array holds only what dataWindowView exposes.
type StudyPlot struct {
	Name                    string    `json:"name"`
	PlotID                  string    `json:"plot_id,omitempty"`
	Type                    string    `json:"type,omitempty"`
	Current                 *float64  `json:"current"`
	Values                  []float64 `json:"values"`
	ValueIndex              int       `json:"value_index,omitempty"`
	IsHidden                bool      `json:"is_hidden,omitempty"`
	DisplayValue            string    `json:"display_value,omitempty"`
	Source                  string    `json:"source,omitempty"`
	Reliability             string    `json:"reliability,omitempty"`
	ReliableForTradingLogic bool      `json:"reliableForTradingLogic"`
}

// StudyResult is one indicator entry in data_get_study_values.
type StudyResult struct {
	Name                    string      `json:"name"`
	EntityID                string      `json:"entity_id"`
	PlotCount               int         `json:"plot_count"`
	Plots                   []StudyPlot `json:"plots"`
	CurrentBarIndex         *int        `json:"current_bar_index,omitempty"`
	Time                    *int64      `json:"time,omitempty"`
	TotalBars               int         `json:"total_bars,omitempty"`
	Coverage                string      `json:"coverage,omitempty"`
	Source                  string      `json:"source,omitempty"`
	Reliability             string      `json:"reliability,omitempty"`
	ReliableForTradingLogic bool        `json:"reliableForTradingLogic"`
}

type rawStudyPlot struct {
	Name         string `json:"name"`
	DisplayValue string `json:"display_value"`
}

type rawStudyResult struct {
	Name      string         `json:"name"`
	EntityID  string         `json:"entity_id"`
	PlotCount int            `json:"plot_count"`
	Plots     []rawStudyPlot `json:"plots"`
}

func GetStudyValues() (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	raw, err := withSession(ctx, func(c *cdp.Client) (json.RawMessage, error) {
		return c.Evaluate(ctx, buildStudyModelJS(studyModelQuery{}, 0, false, true), false)
	})
	if err != nil {
		return nil, err
	}
	var result struct {
		Studies []StudyResult `json:"studies"`
	}
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("parse study values: %w", err)
	}
	if result.Studies == nil {
		result.Studies = []StudyResult{}
	}
	return map[string]interface{}{
		"success":                 true,
		"study_count":             len(result.Studies),
		"source":                  SourceTradingViewStudyModel,
		"reliability":             ReliabilityPineRuntimeUnstableInternal,
		"reliableForTradingLogic": reliableForTradingLogicFromStudyModel,
		"coverage":                "loaded_chart_bars",
		"studies":                 result.Studies,
	}, nil
}

func normalizeRawStudyResults(rawStudies []rawStudyResult) []StudyResult {
	if rawStudies == nil {
		return []StudyResult{}
	}
	studies := make([]StudyResult, 0, len(rawStudies))
	for _, rawStudy := range rawStudies {
		plots := normalizeRawStudyPlots(rawStudy.Plots)
		studies = append(studies, StudyResult{
			Name:                    rawStudy.Name,
			EntityID:                rawStudy.EntityID,
			PlotCount:               len(plots),
			Plots:                   plots,
			Source:                  SourceTradingViewUIDataWindow,
			Reliability:             ReliabilityDisplayValueLocalizedUIString,
			ReliableForTradingLogic: reliableForTradingLogicFromDisplayStrings,
		})
	}
	return studies
}

func normalizeRawStudyPlots(rawPlots []rawStudyPlot) []StudyPlot {
	if rawPlots == nil {
		return []StudyPlot{}
	}
	plots := make([]StudyPlot, 0, len(rawPlots))
	for _, rawPlot := range rawPlots {
		plot := StudyPlot{
			Name:                    rawPlot.Name,
			DisplayValue:            rawPlot.DisplayValue,
			Source:                  SourceTradingViewUIDataWindow,
			Reliability:             ReliabilityDisplayValueLocalizedUIString,
			ReliableForTradingLogic: reliableForTradingLogicFromDisplayStrings,
			Values:                  []float64{},
		}
		if current, ok := ParseDisplayNumber(rawPlot.DisplayValue); ok {
			plot.Current = &current
			plot.Values = []float64{current}
		}
		plots = append(plots, plot)
	}
	return plots
}

// ---------- Pine graphics ----------

type graphicsItem struct {
	ID  interface{}     `json:"id"`
	Raw json.RawMessage `json:"raw"`
}
type graphicsStudy struct {
	Name  string         `json:"name"`
	Count int            `json:"count"`
	Items []graphicsItem `json:"items"`
}

func evalGraphics(ctx context.Context, c *cdp.Client, collection, mapKey, filter string) ([]graphicsStudy, error) {
	raw, err := c.Evaluate(ctx, buildGraphicsJS(collection, mapKey, filter), false)
	if err != nil {
		return nil, err
	}
	var studies []graphicsStudy
	if err := json.Unmarshal(raw, &studies); err != nil {
		return nil, fmt.Errorf("parse graphics: %w", err)
	}
	return studies, nil
}

// GetPineLines returns deduplicated horizontal price levels per Pine study.
func GetPineLines(studyFilter string, verbose bool) (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	var raw []graphicsStudy
	err := cdp.WithSession(ctx, func(c *cdp.Client, _ *cdp.Target) error {
		var err error
		raw, err = evalGraphics(ctx, c, "dwglines", "lines", studyFilter)
		return err
	})
	if err != nil {
		return nil, err
	}
	if len(raw) == 0 {
		return map[string]interface{}{"success": true, "study_count": 0, "studies": []interface{}{}}, nil
	}

	type lineRaw struct {
		Y1 *float64    `json:"y1"`
		Y2 *float64    `json:"y2"`
		X1 interface{} `json:"x1"`
		X2 interface{} `json:"x2"`
		St interface{} `json:"st"`
		W  interface{} `json:"w"`
		Ci interface{} `json:"ci"`
	}
	type studyOut struct {
		Name             string      `json:"name"`
		TotalLines       int         `json:"total_lines"`
		HorizontalLevels []float64   `json:"horizontal_levels"`
		AllLines         interface{} `json:"all_lines,omitempty"`
	}

	studies := make([]studyOut, 0, len(raw))
	for _, s := range raw {
		hLevels := []float64{}
		seen := map[float64]bool{}
		var allLines []interface{}
		for _, item := range s.Items {
			var v lineRaw
			_ = json.Unmarshal(item.Raw, &v)
			if v.Y1 != nil && v.Y2 != nil && *v.Y1 == *v.Y2 {
				y := round2(*v.Y1)
				if !seen[y] {
					hLevels = append(hLevels, y)
					seen[y] = true
				}
			}
			if verbose {
				var y1, y2 *float64
				if v.Y1 != nil {
					r2 := round2(*v.Y1)
					y1 = &r2
				}
				if v.Y2 != nil {
					r2 := round2(*v.Y2)
					y2 = &r2
				}
				allLines = append(allLines, map[string]interface{}{
					"id": item.ID, "y1": y1, "y2": y2, "x1": v.X1, "x2": v.X2,
					"horizontal": v.Y1 != nil && v.Y2 != nil && *v.Y1 == *v.Y2,
					"style":      v.St, "width": v.W, "color": v.Ci,
				})
			}
		}
		// sort descending
		for i := 0; i < len(hLevels); i++ {
			for j := i + 1; j < len(hLevels); j++ {
				if hLevels[j] > hLevels[i] {
					hLevels[i], hLevels[j] = hLevels[j], hLevels[i]
				}
			}
		}
		out := studyOut{Name: s.Name, TotalLines: s.Count, HorizontalLevels: hLevels}
		if verbose {
			out.AllLines = allLines
		}
		studies = append(studies, out)
	}
	return map[string]interface{}{"success": true, "study_count": len(studies), "studies": studies}, nil
}

// GetPineLabels returns text/price label pairs per Pine study.
func GetPineLabels(studyFilter string, maxLabels int, verbose bool) (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if maxLabels <= 0 {
		maxLabels = 50
	}

	var raw []graphicsStudy
	err := cdp.WithSession(ctx, func(c *cdp.Client, _ *cdp.Target) error {
		var err error
		raw, err = evalGraphics(ctx, c, "dwglabels", "labels", studyFilter)
		return err
	})
	if err != nil {
		return nil, err
	}
	if len(raw) == 0 {
		return map[string]interface{}{"success": true, "study_count": 0, "studies": []interface{}{}}, nil
	}

	type labelRaw struct {
		T   string      `json:"t"`
		Y   *float64    `json:"y"`
		X   interface{} `json:"x"`
		Yl  interface{} `json:"yl"`
		Sz  interface{} `json:"sz"`
		Tci interface{} `json:"tci"`
		Ci  interface{} `json:"ci"`
	}

	type studyOut struct {
		Name        string        `json:"name"`
		TotalLabels int           `json:"total_labels"`
		Showing     int           `json:"showing"`
		Labels      []interface{} `json:"labels"`
	}

	studies := make([]studyOut, 0, len(raw))
	for _, s := range raw {
		var labels []interface{}
		for _, item := range s.Items {
			var v labelRaw
			_ = json.Unmarshal(item.Raw, &v)
			var price *float64
			if v.Y != nil {
				r2 := round2(*v.Y)
				price = &r2
			}
			if v.T == "" && price == nil {
				continue
			}
			if verbose {
				labels = append(labels, map[string]interface{}{
					"id": item.ID, "text": v.T, "price": price,
					"x": v.X, "yloc": v.Yl, "size": v.Sz,
					"textColor": v.Tci, "color": v.Ci,
				})
			} else {
				labels = append(labels, map[string]interface{}{"text": v.T, "price": price})
			}
		}
		if len(labels) > maxLabels {
			labels = labels[len(labels)-maxLabels:]
		}
		studies = append(studies, studyOut{
			Name: s.Name, TotalLabels: s.Count,
			Showing: len(labels), Labels: labels,
		})
	}
	return map[string]interface{}{"success": true, "study_count": len(studies), "studies": studies}, nil
}

// GetPineTables returns formatted table rows per Pine study.
func GetPineTables(studyFilter string) (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	var raw []graphicsStudy
	err := cdp.WithSession(ctx, func(c *cdp.Client, _ *cdp.Target) error {
		var err error
		raw, err = evalGraphics(ctx, c, "dwgtablecells", "tableCells", studyFilter)
		return err
	})
	if err != nil {
		return nil, err
	}
	if len(raw) == 0 {
		return map[string]interface{}{"success": true, "study_count": 0, "studies": []interface{}{}}, nil
	}

	type cellRaw struct {
		Tid interface{} `json:"tid"`
		Row int         `json:"row"`
		Col int         `json:"col"`
		T   string      `json:"t"`
	}

	type studyOut struct {
		Name   string        `json:"name"`
		Tables []interface{} `json:"tables"`
	}

	studies := make([]studyOut, 0, len(raw))
	for _, s := range raw {
		// tid → row → col → text
		tables := map[interface{}]map[int]map[int]string{}
		for _, item := range s.Items {
			var v cellRaw
			_ = json.Unmarshal(item.Raw, &v)
			if tables[v.Tid] == nil {
				tables[v.Tid] = map[int]map[int]string{}
			}
			if tables[v.Tid][v.Row] == nil {
				tables[v.Tid][v.Row] = map[int]string{}
			}
			tables[v.Tid][v.Row][v.Col] = v.T
		}
		var tableList []interface{}
		for _, rows := range tables {
			rowNums := sortedKeys(rows)
			var formatted []string
			for _, rn := range rowNums {
				colNums := sortedIntKeys(rows[rn])
				var cells []string
				for _, cn := range colNums {
					if t := rows[rn][cn]; t != "" {
						cells = append(cells, t)
					}
				}
				if len(cells) > 0 {
					formatted = append(formatted, joinStr(cells, " | "))
				}
			}
			tableList = append(tableList, map[string]interface{}{"rows": formatted})
		}
		studies = append(studies, studyOut{Name: s.Name, Tables: tableList})
	}
	return map[string]interface{}{"success": true, "study_count": len(studies), "studies": studies}, nil
}

// GetPineBoxes returns deduplicated {high, low} price zones per Pine study.
func GetPineBoxes(studyFilter string, verbose bool) (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	var raw []graphicsStudy
	err := cdp.WithSession(ctx, func(c *cdp.Client, _ *cdp.Target) error {
		var err error
		raw, err = evalGraphics(ctx, c, "dwgboxes", "boxes", studyFilter)
		return err
	})
	if err != nil {
		return nil, err
	}
	if len(raw) == 0 {
		return map[string]interface{}{"success": true, "study_count": 0, "studies": []interface{}{}}, nil
	}

	type boxRaw struct {
		Y1 *float64    `json:"y1"`
		Y2 *float64    `json:"y2"`
		X1 interface{} `json:"x1"`
		X2 interface{} `json:"x2"`
		C  interface{} `json:"c"`
		Bc interface{} `json:"bc"`
	}

	type zone struct {
		High float64 `json:"high"`
		Low  float64 `json:"low"`
	}
	type studyOut struct {
		Name       string      `json:"name"`
		TotalBoxes int         `json:"total_boxes"`
		Zones      []zone      `json:"zones"`
		AllBoxes   interface{} `json:"all_boxes,omitempty"`
	}

	studies := make([]studyOut, 0, len(raw))
	for _, s := range raw {
		var zones []zone
		seen := map[string]bool{}
		var allBoxes []interface{}
		for _, item := range s.Items {
			var v boxRaw
			_ = json.Unmarshal(item.Raw, &v)
			if v.Y1 != nil && v.Y2 != nil {
				hi := round2(math.Max(*v.Y1, *v.Y2))
				lo := round2(math.Min(*v.Y1, *v.Y2))
				key := fmt.Sprintf("%v:%v", hi, lo)
				if !seen[key] {
					zones = append(zones, zone{High: hi, Low: lo})
					seen[key] = true
				}
				if verbose {
					allBoxes = append(allBoxes, map[string]interface{}{
						"id": item.ID, "high": hi, "low": lo,
						"x1": v.X1, "x2": v.X2, "borderColor": v.C, "bgColor": v.Bc,
					})
				}
			}
		}
		// sort zones descending by high
		for i := 0; i < len(zones); i++ {
			for j := i + 1; j < len(zones); j++ {
				if zones[j].High > zones[i].High {
					zones[i], zones[j] = zones[j], zones[i]
				}
			}
		}
		out := studyOut{Name: s.Name, TotalBoxes: s.Count, Zones: zones}
		if verbose {
			out.AllBoxes = allBoxes
		}
		studies = append(studies, out)
	}
	return map[string]interface{}{"success": true, "study_count": len(studies), "studies": studies}, nil
}

// ---------- Indicator ----------

func GetIndicator(entityID string) (map[string]interface{}, error) {
	return GetIndicatorByQuery(entityID, "")
}

// ---------- Strategy ----------

func GetStrategyResults() (map[string]interface{}, error) {
	return evaluateStrategyReport(strategyReportModeSummary, 0)
}

func GetTrades(maxTradeCount int) (map[string]interface{}, error) {
	return evaluateStrategyReport(strategyReportModeTrades, clampStrategyLimit(maxTradeCount, maxTrades, maxTrades))
}

func GetOrders(maxOrderCount int) (map[string]interface{}, error) {
	return evaluateStrategyReport(strategyReportModeOrders, clampStrategyLimit(maxOrderCount, maxOrders, maxOrders))
}

func GetEquity() (map[string]interface{}, error) {
	return evaluateStrategyReport(strategyReportModeEquity, maxOHLCVBars)
}

// ---------- DOM / Depth ----------

type rawDepthLevel struct {
	PriceDisplayValue string `json:"price_display_value"`
	SizeDisplayValue  string `json:"size_display_value,omitempty"`
}

type depthLevel struct {
	Price                   float64 `json:"price"`
	Size                    float64 `json:"size"`
	PriceDisplayValue       string  `json:"price_display_value,omitempty"`
	SizeDisplayValue        string  `json:"size_display_value,omitempty"`
	Source                  string  `json:"source"`
	Reliability             string  `json:"reliability"`
	ReliableForTradingLogic bool    `json:"reliableForTradingLogic"`
}

func GetDepth() (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	const expr = `(function() {
		var domPanel = document.querySelector('[class*="depth"]')
			|| document.querySelector('[class*="orderBook"]')
			|| document.querySelector('[class*="dom-"]')
			|| document.querySelector('[class*="DOM"]')
			|| document.querySelector('[data-name="dom"]');
		if (!domPanel) return { found: false, error: 'DOM / Depth of Market panel not found.' };
		var bids = [], asks = [];
		var rows = domPanel.querySelectorAll('[class*="row"], tr');
		for (var i = 0; i < rows.length; i++) {
			var row = rows[i];
			var priceEl = row.querySelector('[class*="price"]');
			var sizeEl  = row.querySelector('[class*="size"], [class*="volume"], [class*="qty"]');
			if (!priceEl) continue;
			var level = {
				price_display_value: priceEl.textContent.trim(),
				size_display_value: sizeEl ? sizeEl.textContent.trim() : ''
			};
			var rowClass = row.className || '';
			var rowHTML  = row.innerHTML  || '';
			if (/bid|buy/i.test(rowClass)  || /bid|buy/i.test(rowHTML))  bids.push(level);
			else if (/ask|sell/i.test(rowClass) || /ask|sell/i.test(rowHTML)) asks.push(level);
			else if (i < rows.length / 2) asks.push(level);
			else bids.push(level);
		}
		if (bids.length === 0 && asks.length === 0) {
			var cells = domPanel.querySelectorAll('[class*="cell"], td');
			var prices = [];
			cells.forEach(function(c) { var val = c.textContent.trim(); if (val) prices.push(val); });
			if (prices.length > 0) return { found: true, raw_values: prices.slice(0, 50), bids: [], asks: [], note: 'Could not classify bid/ask levels.' };
		}
		return { found: true, bids: bids, asks: asks };
	})()`

	raw, err := withSession(ctx, func(c *cdp.Client) (json.RawMessage, error) {
		return c.Evaluate(ctx, expr, false)
	})
	if err != nil {
		return nil, err
	}
	var result struct {
		Found     bool            `json:"found"`
		Bids      []rawDepthLevel `json:"bids"`
		Asks      []rawDepthLevel `json:"asks"`
		RawValues []string        `json:"raw_values,omitempty"`
		Note      string          `json:"note,omitempty"`
		Error     string          `json:"error,omitempty"`
	}
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("parse depth: %w", err)
	}
	if !result.Found {
		return nil, fmt.Errorf("%s", coalesce(result.Error, "DOM panel not found"))
	}
	bids := normalizeDepthLevels(result.Bids)
	asks := normalizeDepthLevels(result.Asks)
	sort.Slice(bids, func(i, j int) bool { return bids[i].Price > bids[j].Price })
	sort.Slice(asks, func(i, j int) bool { return asks[i].Price < asks[j].Price })
	var spread *float64
	if len(asks) > 0 && len(bids) > 0 {
		v := round6(asks[0].Price - bids[0].Price)
		spread = &v
	}
	out := map[string]interface{}{
		"success":                 true,
		"bid_levels":              len(bids),
		"ask_levels":              len(asks),
		"spread":                  spread,
		"source":                  SourceTradingViewUIDOM,
		"reliability":             ReliabilityDisplayValueLocalizedUIString,
		"reliableForTradingLogic": reliableForTradingLogicFromDisplayStrings,
		"bids":                    bids,
		"asks":                    asks,
	}
	if result.RawValues != nil {
		out["raw_values"] = parseDisplayNumberStrings(result.RawValues)
		out["raw_display_values"] = result.RawValues
	}
	if result.Note != "" {
		out["note"] = result.Note
	}
	return out, nil
}

func normalizeDepthLevels(raw []rawDepthLevel) []depthLevel {
	if raw == nil {
		return []depthLevel{}
	}
	levels := make([]depthLevel, 0, len(raw))
	for _, r := range raw {
		price, ok := ParseDisplayNumber(r.PriceDisplayValue)
		if !ok {
			continue
		}
		size := 0.0
		if parsedSize, ok := ParseDisplayNumber(r.SizeDisplayValue); ok {
			size = parsedSize
		}
		levels = append(levels, depthLevel{
			Price:                   price,
			Size:                    size,
			PriceDisplayValue:       r.PriceDisplayValue,
			SizeDisplayValue:        r.SizeDisplayValue,
			Source:                  SourceTradingViewUIDOM,
			Reliability:             ReliabilityDisplayValueLocalizedUIString,
			ReliableForTradingLogic: reliableForTradingLogicFromDisplayStrings,
		})
	}
	return levels
}

func parseDisplayNumberStrings(values []string) []float64 {
	if values == nil {
		return []float64{}
	}
	out := make([]float64, 0, len(values))
	for _, v := range values {
		if parsed, ok := ParseDisplayNumber(v); ok && parsed > 0 {
			out = append(out, parsed)
		}
	}
	return out
}

// ---------- small helpers ----------

func coalesce(a, b string) string {
	if a != "" {
		return a
	}
	return b
}

func sortedKeys(m map[int]map[int]string) []int {
	keys := make([]int, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	for i := 0; i < len(keys); i++ {
		for j := i + 1; j < len(keys); j++ {
			if keys[j] < keys[i] {
				keys[i], keys[j] = keys[j], keys[i]
			}
		}
	}
	return keys
}

func sortedIntKeys(m map[int]string) []int {
	keys := make([]int, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	for i := 0; i < len(keys); i++ {
		for j := i + 1; j < len(keys); j++ {
			if keys[j] < keys[i] {
				keys[i], keys[j] = keys[j], keys[i]
			}
		}
	}
	return keys
}

func joinStr(parts []string, sep string) string {
	if len(parts) == 0 {
		return ""
	}
	r := parts[0]
	for _, p := range parts[1:] {
		r += sep + p
	}
	return r
}

func round6(v float64) float64 { return math.Round(v*1_000_000) / 1_000_000 }

// ---------- RegisterTools ----------

func RegisterTools(reg *mcp.Registry) {
	reg.Register(mcp.ToolDef{
		Name:        "data_get_ohlcv",
		Description: "Get OHLCV bar data from the chart. Use summary=true for compact stats instead of all bars (saves context).",
		Schema: mcp.InputSchema{
			Type: "object",
			Properties: map[string]mcp.PropertySchema{
				"count":   {Type: "number", Description: "Number of bars to retrieve (max 500, default 100)"},
				"summary": {Type: "boolean", Description: "Return summary stats instead of all bars — much smaller output"},
			},
		},
		Handler: func(args json.RawMessage) (interface{}, error) {
			var p struct {
				Count   int  `json:"count"`
				Summary bool `json:"summary"`
			}
			_ = json.Unmarshal(args, &p)
			result, err := GetOhlcv(p.Count, p.Summary)
			if err != nil {
				return map[string]interface{}{"success": false, "error": err.Error()}, nil
			}
			return result, nil
		},
	})

	reg.Register(mcp.ToolDef{
		Name:        "quote_get",
		Description: "Get real-time quote data for a symbol (price, OHLC, volume)",
		Schema: mcp.InputSchema{
			Type: "object",
			Properties: map[string]mcp.PropertySchema{
				"symbol": {Type: "string", Description: "Symbol to quote (blank = current chart symbol)"},
			},
		},
		Handler: func(args json.RawMessage) (interface{}, error) {
			var p struct {
				Symbol string `json:"symbol"`
			}
			_ = json.Unmarshal(args, &p)
			result, err := GetQuote(p.Symbol)
			if err != nil {
				return map[string]interface{}{"success": false, "error": err.Error()}, nil
			}
			return result, nil
		},
	})

	reg.Register(mcp.ToolDef{
		Name:        "data_get_study_values",
		Description: "Get current numeric values from all visible indicators (RSI, MACD, BB, EMA, etc.)",
		Schema:      mcp.InputSchema{Type: "object"},
		Handler: func(args json.RawMessage) (interface{}, error) {
			result, err := GetStudyValues()
			if err != nil {
				return map[string]interface{}{"success": false, "error": err.Error()}, nil
			}
			return result, nil
		},
	})

	reg.Register(mcp.ToolDef{
		Name:        "data_get_pine_lines",
		Description: "Read horizontal price levels drawn by Pine Script indicators (line.new). Returns deduplicated price levels per study.",
		Schema: mcp.InputSchema{
			Type: "object",
			Properties: map[string]mcp.PropertySchema{
				"study_filter": {Type: "string", Description: "Substring to match study name. Omit for all."},
				"verbose":      {Type: "boolean", Description: "Return raw line data with IDs, coordinates, colors"},
			},
		},
		Handler: func(args json.RawMessage) (interface{}, error) {
			var p struct {
				StudyFilter string `json:"study_filter"`
				Verbose     bool   `json:"verbose"`
			}
			_ = json.Unmarshal(args, &p)
			result, err := GetPineLines(p.StudyFilter, p.Verbose)
			if err != nil {
				return map[string]interface{}{"success": false, "error": err.Error()}, nil
			}
			return result, nil
		},
	})

	reg.Register(mcp.ToolDef{
		Name:        "data_get_pine_labels",
		Description: "Read text labels drawn by Pine Script indicators (label.new). Returns text and price pairs.",
		Schema: mcp.InputSchema{
			Type: "object",
			Properties: map[string]mcp.PropertySchema{
				"study_filter": {Type: "string", Description: "Substring to match study name. Omit for all."},
				"max_labels":   {Type: "number", Description: "Max labels per study (default 50)"},
				"verbose":      {Type: "boolean", Description: "Return raw label data with IDs, colors, positions"},
			},
		},
		Handler: func(args json.RawMessage) (interface{}, error) {
			var p struct {
				StudyFilter string `json:"study_filter"`
				MaxLabels   int    `json:"max_labels"`
				Verbose     bool   `json:"verbose"`
			}
			_ = json.Unmarshal(args, &p)
			result, err := GetPineLabels(p.StudyFilter, p.MaxLabels, p.Verbose)
			if err != nil {
				return map[string]interface{}{"success": false, "error": err.Error()}, nil
			}
			return result, nil
		},
	})

	reg.Register(mcp.ToolDef{
		Name:        "data_get_pine_tables",
		Description: "Read table data drawn by Pine Script indicators (table.new). Returns formatted text rows per table.",
		Schema: mcp.InputSchema{
			Type: "object",
			Properties: map[string]mcp.PropertySchema{
				"study_filter": {Type: "string", Description: "Substring to match study name. Omit for all."},
			},
		},
		Handler: func(args json.RawMessage) (interface{}, error) {
			var p struct {
				StudyFilter string `json:"study_filter"`
			}
			_ = json.Unmarshal(args, &p)
			result, err := GetPineTables(p.StudyFilter)
			if err != nil {
				return map[string]interface{}{"success": false, "error": err.Error()}, nil
			}
			return result, nil
		},
	})

	reg.Register(mcp.ToolDef{
		Name:        "data_get_pine_boxes",
		Description: "Read box/zone boundaries drawn by Pine Script indicators (box.new). Returns deduplicated {high, low} price zones.",
		Schema: mcp.InputSchema{
			Type: "object",
			Properties: map[string]mcp.PropertySchema{
				"study_filter": {Type: "string", Description: "Substring to match study name. Omit for all."},
				"verbose":      {Type: "boolean", Description: "Return all boxes with IDs and coordinates"},
			},
		},
		Handler: func(args json.RawMessage) (interface{}, error) {
			var p struct {
				StudyFilter string `json:"study_filter"`
				Verbose     bool   `json:"verbose"`
			}
			_ = json.Unmarshal(args, &p)
			result, err := GetPineBoxes(p.StudyFilter, p.Verbose)
			if err != nil {
				return map[string]interface{}{"success": false, "error": err.Error()}, nil
			}
			return result, nil
		},
	})

	reg.Register(mcp.ToolDef{
		Name:        "data_get_indicator",
		Description: "Get real indicator/study values from the TradingView study model by entity ID or name",
		Schema: mcp.InputSchema{
			Type: "object",
			Properties: map[string]mcp.PropertySchema{
				"entity_id": {Type: "string", Description: "Study entity ID (from chart_get_state)"},
				"name":      {Type: "string", Description: "Study name or plot title substring to match when entity_id is omitted"},
			},
		},
		Handler: func(args json.RawMessage) (interface{}, error) {
			var p struct {
				EntityID string `json:"entity_id"`
				Name     string `json:"name"`
			}
			if err := json.Unmarshal(args, &p); err != nil {
				return map[string]interface{}{"success": false, "error": err.Error()}, nil
			}
			if p.EntityID == "" && p.Name == "" {
				return map[string]interface{}{"success": false, "error": "entity_id or name is required"}, nil
			}
			result, err := GetIndicatorByQuery(p.EntityID, p.Name)
			if err != nil {
				return map[string]interface{}{"success": false, "error": err.Error()}, nil
			}
			return result, nil
		},
	})

	reg.Register(mcp.ToolDef{
		Name:        "data_get_indicator_history",
		Description: "Get historical indicator/study values from the TradingView study model for loaded chart bars",
		Schema: mcp.InputSchema{
			Type: "object",
			Properties: map[string]mcp.PropertySchema{
				"entity_id": {Type: "string", Description: "Study entity ID (from chart_get_state)"},
				"name":      {Type: "string", Description: "Study name or plot title substring to match when entity_id is omitted"},
				"max_bars":  {Type: "number", Description: "Maximum loaded chart bars to return (default 500, max 500)"},
			},
		},
		Handler: func(args json.RawMessage) (interface{}, error) {
			var p struct {
				EntityID string `json:"entity_id"`
				Name     string `json:"name"`
				MaxBars  int    `json:"max_bars"`
			}
			if err := json.Unmarshal(args, &p); err != nil {
				return map[string]interface{}{"success": false, "error": err.Error()}, nil
			}
			if p.EntityID == "" && p.Name == "" {
				return map[string]interface{}{"success": false, "error": "entity_id or name is required"}, nil
			}
			result, err := GetIndicatorHistory(p.EntityID, p.Name, p.MaxBars)
			if err != nil {
				return map[string]interface{}{"success": false, "error": err.Error()}, nil
			}
			return result, nil
		},
	})

	reg.Register(mcp.ToolDef{
		Name:        "data_get_strategy_results",
		Description: "Get strategy performance metrics from Strategy Tester",
		Schema:      mcp.InputSchema{Type: "object"},
		Handler: func(args json.RawMessage) (interface{}, error) {
			result, err := GetStrategyResults()
			if err != nil {
				return map[string]interface{}{"success": false, "error": err.Error()}, nil
			}
			return result, nil
		},
	})

	reg.Register(mcp.ToolDef{
		Name:        "data_get_trades",
		Description: "Get trade list from Strategy Tester",
		Schema: mcp.InputSchema{
			Type: "object",
			Properties: map[string]mcp.PropertySchema{
				"max_trades": {Type: "number", Description: "Maximum trades to return (max 20)"},
			},
		},
		Handler: func(args json.RawMessage) (interface{}, error) {
			var p struct {
				MaxTrades int `json:"max_trades"`
			}
			_ = json.Unmarshal(args, &p)
			result, err := GetTrades(p.MaxTrades)
			if err != nil {
				return map[string]interface{}{"success": false, "error": err.Error()}, nil
			}
			return result, nil
		},
	})

	reg.Register(mcp.ToolDef{
		Name:        "data_get_orders",
		Description: "Get filled order list from Strategy Tester",
		Schema: mcp.InputSchema{
			Type: "object",
			Properties: map[string]mcp.PropertySchema{
				"max_orders": {Type: "number", Description: "Maximum orders to return (max 50)"},
			},
		},
		Handler: func(args json.RawMessage) (interface{}, error) {
			var p struct {
				MaxOrders int `json:"max_orders"`
			}
			_ = json.Unmarshal(args, &p)
			result, err := GetOrders(p.MaxOrders)
			if err != nil {
				return map[string]interface{}{"success": false, "error": err.Error()}, nil
			}
			return result, nil
		},
	})

	reg.Register(mcp.ToolDef{
		Name:        "data_get_equity",
		Description: "Get equity curve data from Strategy Tester",
		Schema:      mcp.InputSchema{Type: "object"},
		Handler: func(args json.RawMessage) (interface{}, error) {
			result, err := GetEquity()
			if err != nil {
				return map[string]interface{}{"success": false, "error": err.Error()}, nil
			}
			return result, nil
		},
	})

	reg.Register(mcp.ToolDef{
		Name:        "depth_get",
		Description: "Get order book / DOM (Depth of Market) data from the chart",
		Schema:      mcp.InputSchema{Type: "object"},
		Handler: func(args json.RawMessage) (interface{}, error) {
			result, err := GetDepth()
			if err != nil {
				return map[string]interface{}{
					"success": false, "error": err.Error(),
					"hint": "Open the DOM panel in TradingView before using this tool.",
				}, nil
			}
			return result, nil
		},
	})
}
