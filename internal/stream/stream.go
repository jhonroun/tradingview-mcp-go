// Package stream implements JSONL streaming of TradingView real-time data.
// All streams write to an io.Writer (stdout in CLI usage), run until ctx is
// cancelled (SIGINT/SIGTERM), and only emit a line when data changes (dedup).
package stream

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/jhonroun/tradingview-mcp-go/internal/cdp"
)

const (
	chartAPI = `window.TradingViewApi._activeChartWidgetWV.value()`
	model    = chartAPI + `._chartWidget.model()`
	cwc      = `window.TradingViewApi._chartWidgetCollection`
)

// fetcher is a function that pulls data from TradingView via a live CDP client.
// Returns nil to signal "no data yet, skip this tick".
type fetcher func(ctx context.Context, c *cdp.Client) (map[string]interface{}, error)

// pollLoop is the core streaming engine — mirrors pollLoop() in core/stream.js.
// It connects once, then polls fetcher at intervalMs, emitting JSONL on change.
// On CDP connection errors it reconnects. Stops when ctx is cancelled.
func pollLoop(ctx context.Context, w io.Writer, errW io.Writer, label string, intervalMs int, dedupe bool, fn fetcher) error {
	start := time.Now()
	fmt.Fprintf(errW, "[stream:%s] started, interval=%dms, Ctrl+C to stop\n", label, intervalMs)

	interval := time.Duration(intervalMs) * time.Millisecond
	var lastHash string
	var client *cdp.Client

	connect := func() {
		if client != nil {
			client.Close()
			client = nil
		}
		connCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		targets, err := cdp.ListTargets(connCtx, "localhost", 9222)
		if err != nil {
			return
		}
		target, err := cdp.FindChartTarget(targets)
		if err != nil {
			return
		}
		c, err := cdp.Connect(connCtx, target)
		if err != nil {
			return
		}
		if err := c.EnableDomains(connCtx); err != nil {
			c.Close()
			return
		}
		client = c
	}

	connect()

	for {
		select {
		case <-ctx.Done():
			if client != nil {
				client.Close()
			}
			elapsed := time.Since(start).Seconds()
			fmt.Fprintf(errW, "[stream:%s] stopped after %.1fs\n", label, elapsed)
			return nil
		default:
		}

		if client == nil {
			time.Sleep(2 * time.Second)
			connect()
			continue
		}

		evalCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		data, err := fn(evalCtx, client)
		cancel()

		if err != nil {
			// Connection-level errors — reconnect silently.
			if isCDPError(err) {
				client.Close()
				client = nil
				time.Sleep(2 * time.Second)
				continue
			}
			fmt.Fprintf(errW, "[stream:%s] error: %v\n", label, err)
		} else if data != nil {
			data["_ts"] = time.Now().UnixMilli()
			data["_stream"] = label

			raw, _ := json.Marshal(data)
			hash := string(raw)

			if !dedupe || hash != lastHash {
				lastHash = hash
				fmt.Fprintf(w, "%s\n", raw)
			}
		}

		select {
		case <-ctx.Done():
		case <-time.After(interval):
		}
	}
}

func isCDPError(err error) bool {
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "cdp") ||
		strings.Contains(msg, "econnrefused") ||
		strings.Contains(msg, "websocket") ||
		strings.Contains(msg, "connection") ||
		strings.Contains(msg, "closed")
}

func evalData(ctx context.Context, c *cdp.Client, expr string) (map[string]interface{}, error) {
	raw, err := c.Evaluate(ctx, expr, false)
	if err != nil {
		return nil, err
	}
	var v interface{}
	if err := json.Unmarshal(raw, &v); err != nil {
		return nil, err
	}
	if v == nil {
		return nil, nil
	}
	if m, ok := v.(map[string]interface{}); ok {
		return m, nil
	}
	// Scalar result — shouldn't happen for our expressions.
	return nil, nil
}

// ── Quote ────────────────────────────────────────────────────────────────────

const fetchQuoteExpr = `(function() {
	var chart = ` + chartAPI + `;
	var m = ` + model + `;
	var bars = m.mainSeries().bars();
	var last = bars.lastIndex();
	var v = bars.valueAt(last);
	if (!v) return null;
	return { symbol: chart.symbol(), time: v[0], open: v[1], high: v[2], low: v[3], close: v[4], volume: v[5] || 0 };
})()`

// StreamQuote streams real-time price ticks (OHLCV of the last bar).
// Default interval: 300 ms.
func StreamQuote(ctx context.Context, w io.Writer, errW io.Writer, intervalMs int) error {
	if intervalMs <= 0 {
		intervalMs = 300
	}
	return pollLoop(ctx, w, errW, "quote", intervalMs, true, func(ctx context.Context, c *cdp.Client) (map[string]interface{}, error) {
		return evalData(ctx, c, fetchQuoteExpr)
	})
}

// ── Bars ─────────────────────────────────────────────────────────────────────

const fetchLastBarExpr = `(function() {
	var chart = ` + chartAPI + `;
	var m = ` + model + `;
	var bars = m.mainSeries().bars();
	var last = bars.lastIndex();
	var v = bars.valueAt(last);
	if (!v) return null;
	return { symbol: chart.symbol(), resolution: chart.resolution(), bar_time: v[0], open: v[1], high: v[2], low: v[3], close: v[4], volume: v[5] || 0, bar_index: last };
})()`

// StreamBars streams last-bar updates (emits on new bar or price change).
// Default interval: 500 ms.
func StreamBars(ctx context.Context, w io.Writer, errW io.Writer, intervalMs int) error {
	if intervalMs <= 0 {
		intervalMs = 500
	}
	return pollLoop(ctx, w, errW, "bars", intervalMs, true, func(ctx context.Context, c *cdp.Client) (map[string]interface{}, error) {
		return evalData(ctx, c, fetchLastBarExpr)
	})
}

// ── Values ───────────────────────────────────────────────────────────────────

const fetchValuesExpr = `(function() {
	var chart = ` + chartAPI + `;
	var m = ` + model + `;
	var studies = chart.getAllStudies();
	var results = [];
	for (var i = 0; i < studies.length; i++) {
		try {
			var study = chart.getStudyById(studies[i].id);
			if (!study || !study.isVisible()) continue;
			var src = study._study || study;
			var data = src._lastBarValues || src._data;
			if (!data) continue;
			var vals = {};
			if (typeof data === 'object') {
				for (var k in data) {
					if (typeof data[k] === 'number' && !isNaN(data[k])) vals[k] = data[k];
				}
			}
			if (Object.keys(vals).length > 0) results.push({ name: studies[i].name, values: vals });
		} catch(e) {}
	}
	return { symbol: chart.symbol(), study_count: results.length, studies: results };
})()`

// StreamValues streams indicator values (RSI, MACD, etc.).
// Default interval: 500 ms.
func StreamValues(ctx context.Context, w io.Writer, errW io.Writer, intervalMs int) error {
	if intervalMs <= 0 {
		intervalMs = 500
	}
	return pollLoop(ctx, w, errW, "values", intervalMs, true, func(ctx context.Context, c *cdp.Client) (map[string]interface{}, error) {
		return evalData(ctx, c, fetchValuesExpr)
	})
}

// ── Lines ────────────────────────────────────────────────────────────────────

func buildLinesExpr(studyFilter string) string {
	filter := "null"
	if studyFilter != "" {
		b, _ := json.Marshal(studyFilter)
		filter = string(b)
	}
	return `(function() {
	var filter = ` + filter + `;
	var chart = ` + chartAPI + `;
	var studies = chart.getAllStudies();
	var results = [];
	for (var i = 0; i < studies.length; i++) {
		var s = studies[i];
		if (filter && (s.name || '').toLowerCase().indexOf(filter.toLowerCase()) === -1) continue;
		try {
			var study = chart.getStudyById(s.id);
			if (!study) continue;
			var src = study._study || study;
			var g = src._graphics || (src._source && src._source._graphics);
			if (!g) continue;
			var pc = g._primitivesCollection;
			if (!pc || !pc.dwglines) continue;
			var linesMap = pc.dwglines.get('lines');
			if (!linesMap) continue;
			var data = linesMap.get(false);
			if (!data || !data._primitivesDataById) continue;
			var levels = [];
			var seen = {};
			data._primitivesDataById.forEach(function(line) {
				var p1 = line.points && line.points[0] ? line.points[0].price : null;
				var p2 = line.points && line.points[1] ? line.points[1].price : null;
				var price = (p1 !== null && p1 === p2) ? p1 : (p1 || p2);
				if (price !== null && !seen[price]) { seen[price] = true; levels.push(price); }
			});
			levels.sort(function(a, b) { return b - a; });
			if (levels.length > 0) results.push({ study: s.name, levels: levels });
		} catch(e) {}
	}
	return { symbol: chart.symbol(), study_count: results.length, studies: results };
})()`
}

// StreamLines streams Pine Script line.new() price levels.
// Default interval: 1000 ms.
func StreamLines(ctx context.Context, w io.Writer, errW io.Writer, intervalMs int, filter string) error {
	if intervalMs <= 0 {
		intervalMs = 1000
	}
	expr := buildLinesExpr(filter)
	return pollLoop(ctx, w, errW, "lines", intervalMs, true, func(ctx context.Context, c *cdp.Client) (map[string]interface{}, error) {
		return evalData(ctx, c, expr)
	})
}

// ── Labels ───────────────────────────────────────────────────────────────────

func buildLabelsExpr(studyFilter string) string {
	filter := "null"
	if studyFilter != "" {
		b, _ := json.Marshal(studyFilter)
		filter = string(b)
	}
	return `(function() {
	var filter = ` + filter + `;
	var chart = ` + chartAPI + `;
	var studies = chart.getAllStudies();
	var results = [];
	for (var i = 0; i < studies.length; i++) {
		var s = studies[i];
		if (filter && (s.name || '').toLowerCase().indexOf(filter.toLowerCase()) === -1) continue;
		try {
			var study = chart.getStudyById(s.id);
			if (!study) continue;
			var src = study._study || study;
			var g = src._graphics || (src._source && src._source._graphics);
			if (!g) continue;
			var pc = g._primitivesCollection;
			if (!pc || !pc.dwglabels) continue;
			var labelsMap = pc.dwglabels.get('labels');
			if (!labelsMap) continue;
			var data = labelsMap.get(false);
			if (!data || !data._primitivesDataById) continue;
			var labels = [];
			data._primitivesDataById.forEach(function(lbl) {
				var text = lbl.text || '';
				var price = lbl.points && lbl.points[0] ? lbl.points[0].price : null;
				if (text) labels.push({ text: text, price: price });
			});
			if (labels.length > 0) results.push({ study: s.name, labels: labels.slice(0, 50) });
		} catch(e) {}
	}
	return { symbol: chart.symbol(), study_count: results.length, studies: results };
})()`
}

// StreamLabels streams Pine Script label.new() annotations.
// Default interval: 1000 ms.
func StreamLabels(ctx context.Context, w io.Writer, errW io.Writer, intervalMs int, filter string) error {
	if intervalMs <= 0 {
		intervalMs = 1000
	}
	expr := buildLabelsExpr(filter)
	return pollLoop(ctx, w, errW, "labels", intervalMs, true, func(ctx context.Context, c *cdp.Client) (map[string]interface{}, error) {
		return evalData(ctx, c, expr)
	})
}

// ── Tables ───────────────────────────────────────────────────────────────────

func buildTablesExpr(studyFilter string) string {
	filter := "null"
	if studyFilter != "" {
		b, _ := json.Marshal(studyFilter)
		filter = string(b)
	}
	return `(function() {
	var filter = ` + filter + `;
	var chart = ` + chartAPI + `;
	var studies = chart.getAllStudies();
	var results = [];
	for (var i = 0; i < studies.length; i++) {
		var s = studies[i];
		if (filter && (s.name || '').toLowerCase().indexOf(filter.toLowerCase()) === -1) continue;
		try {
			var study = chart.getStudyById(s.id);
			if (!study) continue;
			var src = study._study || study;
			var g = src._graphics || (src._source && src._source._graphics);
			if (!g) continue;
			var pc = g._primitivesCollection;
			if (!pc || !pc.ownFirstValue) continue;
			var tableMap = pc.ownFirstValue();
			if (!tableMap) continue;
			var tables = [];
			if (typeof tableMap.forEach === 'function') {
				tableMap.forEach(function(table) {
					if (!table || !table.data) return;
					var rows = [];
					for (var r = 0; r < table.data.length; r++) {
						var row = [];
						for (var c = 0; c < table.data[r].length; c++) { row.push(table.data[r][c].text || ''); }
						rows.push(row);
					}
					tables.push({ rows: rows });
				});
			}
			if (tables.length > 0) results.push({ study: s.name, tables: tables });
		} catch(e) {}
	}
	return { symbol: chart.symbol(), study_count: results.length, studies: results };
})()`
}

// StreamTables streams Pine Script table.new() data.
// Default interval: 2000 ms.
func StreamTables(ctx context.Context, w io.Writer, errW io.Writer, intervalMs int, filter string) error {
	if intervalMs <= 0 {
		intervalMs = 2000
	}
	expr := buildTablesExpr(filter)
	return pollLoop(ctx, w, errW, "tables", intervalMs, true, func(ctx context.Context, c *cdp.Client) (map[string]interface{}, error) {
		return evalData(ctx, c, expr)
	})
}

// ── All panes ─────────────────────────────────────────────────────────────────

const fetchAllPanesExpr = `(function() {
	var cwcObj = ` + cwc + `;
	var all = cwcObj.getAll();
	var layoutType = cwcObj._layoutType;
	if (typeof layoutType === 'object' && layoutType && typeof layoutType.value === 'function') layoutType = layoutType.value();
	var count = cwcObj.inlineChartsCount;
	if (typeof count === 'object' && count && typeof count.value === 'function') count = count.value();
	var panes = [];
	for (var i = 0; i < Math.min(all.length, count || all.length); i++) {
		try {
			var c = all[i];
			var m = c.model();
			var ms = m.mainSeries();
			var bars = ms.bars();
			var last = bars.lastIndex();
			var v = bars.valueAt(last);
			if (!v) { panes.push({ index: i, symbol: ms.symbol(), error: 'no bars' }); continue; }
			panes.push({ index: i, symbol: ms.symbol(), resolution: ms.interval(), time: v[0], open: v[1], high: v[2], low: v[3], close: v[4], volume: v[5] || 0 });
		} catch(e) { panes.push({ index: i, error: e.message }); }
	}
	return { layout: layoutType, pane_count: panes.length, panes: panes };
})()`

// StreamAllPanes streams all chart panes simultaneously (multi-symbol monitoring).
// Default interval: 500 ms.
func StreamAllPanes(ctx context.Context, w io.Writer, errW io.Writer, intervalMs int) error {
	if intervalMs <= 0 {
		intervalMs = 500
	}
	return pollLoop(ctx, w, errW, "all-panes", intervalMs, true, func(ctx context.Context, c *cdp.Client) (map[string]interface{}, error) {
		return evalData(ctx, c, fetchAllPanesExpr)
	})
}
