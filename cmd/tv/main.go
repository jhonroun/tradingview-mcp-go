// Command tv is the TradingView CLI.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/jhonroun/tradingview-mcp-go/internal/cdp"
	"github.com/jhonroun/tradingview-mcp-go/internal/cli"
	"github.com/jhonroun/tradingview-mcp-go/internal/discovery"
	"github.com/jhonroun/tradingview-mcp-go/internal/tools/alerts"
	"github.com/jhonroun/tradingview-mcp-go/internal/tools/capture"
	charttools "github.com/jhonroun/tradingview-mcp-go/internal/tools/chart"
	"github.com/jhonroun/tradingview-mcp-go/internal/tools/data"
	"github.com/jhonroun/tradingview-mcp-go/internal/tools/drawing"
	"github.com/jhonroun/tradingview-mcp-go/internal/tools/health"
	"github.com/jhonroun/tradingview-mcp-go/internal/tools/indicators"
	"github.com/jhonroun/tradingview-mcp-go/internal/tools/pane"
	"github.com/jhonroun/tradingview-mcp-go/internal/tools/pine"
	"github.com/jhonroun/tradingview-mcp-go/internal/tools/replay"
	"github.com/jhonroun/tradingview-mcp-go/internal/tools/tab"
	uitools "github.com/jhonroun/tradingview-mcp-go/internal/tools/ui"
	"github.com/jhonroun/tradingview-mcp-go/internal/tools/batch"
	"github.com/jhonroun/tradingview-mcp-go/internal/stream"
)

func init() {
	// ── health ────────────────────────────────────────────────────────────────
	cli.Register(cli.Command{
		Name:        "status",
		Description: "Check CDP connection to TradingView",
		Handler: func(args []string, opts map[string]string) (interface{}, error) {
			return health.HealthCheck()
		},
	})

	cli.Register(cli.Command{
		Name:        "launch",
		Description: "Launch TradingView with CDP enabled",
		Handler: func(args []string, opts map[string]string) (interface{}, error) {
			a := health.LaunchArgs{}
			if p, ok := opts["port"]; ok {
				if port, err := strconv.Atoi(p); err == nil {
					a.Port = &port
				}
			}
			if _, ok := opts["no-kill"]; ok {
				b := false
				a.KillExisting = &b
			}
			if p, ok := opts["tv-path"]; ok {
				a.TvPath = &p
			}
			return health.Launch(a)
		},
	})

	cli.Register(cli.Command{
		Name:        "discover",
		Description: "Report which TradingView API paths are available",
		Handler: func(args []string, opts map[string]string) (interface{}, error) {
			return health.Discover()
		},
	})

	cli.Register(cli.Command{
		Name:        "ui-state",
		Description: "Get current TradingView UI panel state",
		Handler: func(args []string, opts map[string]string) (interface{}, error) {
			return health.UIState()
		},
	})

	cli.Register(cli.Command{
		Name:        "doctor",
		Description: "Diagnose TradingView installation and CDP connectivity",
		Handler: func(args []string, opts map[string]string) (interface{}, error) {
			result := map[string]interface{}{"success": true}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			targets, err := cdp.ListTargets(ctx, "localhost", 9222)
			if err != nil {
				result["cdp"] = map[string]interface{}{
					"ok":    false,
					"error": err.Error(),
					"hint":  "Start TradingView with --remote-debugging-port=9222, or run: tv launch",
				}
			} else {
				chartTarget, ferr := cdp.FindChartTarget(targets)
				if ferr != nil {
					result["cdp"] = map[string]interface{}{
						"ok":      true,
						"targets": len(targets),
						"chart":   false,
						"hint":    "CDP is available but no chart target found — open a chart in TradingView",
					}
				} else {
					result["cdp"] = map[string]interface{}{
						"ok":       true,
						"targets":  len(targets),
						"chart":    true,
						"targetId": chartTarget.ID,
						"url":      chartTarget.URL,
					}
				}
			}

			found, err := discovery.Find()
			if err != nil {
				result["install"] = map[string]interface{}{"ok": false, "error": err.Error()}
			} else {
				result["install"] = map[string]interface{}{
					"ok": true, "path": found.Path,
					"source": found.Source, "platform": found.Platform,
				}
			}
			return result, nil
		},
	})

	// ── chart ─────────────────────────────────────────────────────────────────
	cli.Register(cli.Command{
		Name:        "chart-state",
		Description: "Get current chart symbol, timeframe, and indicators",
		Handler: func(args []string, opts map[string]string) (interface{}, error) {
			return charttools.GetState()
		},
	})

	// ── data ──────────────────────────────────────────────────────────────────
	cli.Register(cli.Command{
		Name:        "quote",
		Description: "Get real-time quote (price, OHLC, volume) — tv quote [SYMBOL]",
		Handler: func(args []string, opts map[string]string) (interface{}, error) {
			sym := ""
			if len(args) > 0 {
				sym = args[0]
			}
			return data.GetQuote(sym)
		},
	})

	cli.Register(cli.Command{
		Name:        "ohlcv",
		Description: "Get OHLCV bars — tv ohlcv [--count N] [--summary]",
		Handler: func(args []string, opts map[string]string) (interface{}, error) {
			count := 100
			if v, ok := opts["count"]; ok {
				if n, err := strconv.Atoi(v); err == nil {
					count = n
				}
			}
			_, summary := opts["summary"]
			return data.GetOhlcv(count, summary)
		},
	})

	cli.Register(cli.Command{
		Name:        "screenshot",
		Description: `Take a screenshot — tv screenshot [--region full|chart|strategy_tester] [--filename NAME]`,
		Handler: func(args []string, opts map[string]string) (interface{}, error) {
			region := opts["region"]
			filename := opts["filename"]
			return capture.CaptureScreenshot(region, filename)
		},
	})

	// ── chart control (P6) ────────────────────────────────────────────────────
	cli.Register(cli.Command{
		Name:        "set-symbol",
		Description: "Change chart symbol — tv set-symbol SYMBOL",
		Handler: func(args []string, opts map[string]string) (interface{}, error) {
			if len(args) == 0 {
				return nil, fmt.Errorf("usage: tv set-symbol SYMBOL")
			}
			return charttools.SetSymbol(args[0])
		},
	})

	cli.Register(cli.Command{
		Name:        "set-timeframe",
		Description: "Change chart resolution — tv set-timeframe TF (e.g. 1 5 15 60 D W M)",
		Handler: func(args []string, opts map[string]string) (interface{}, error) {
			if len(args) == 0 {
				return nil, fmt.Errorf("usage: tv set-timeframe TIMEFRAME")
			}
			return charttools.SetTimeframe(args[0])
		},
	})

	cli.Register(cli.Command{
		Name:        "set-type",
		Description: "Change chart type — tv set-type TYPE (Candles, Bars, Line, Area, HeikinAshi, …)",
		Handler: func(args []string, opts map[string]string) (interface{}, error) {
			if len(args) == 0 {
				return nil, fmt.Errorf("usage: tv set-type TYPE")
			}
			return charttools.SetType(args[0])
		},
	})

	cli.Register(cli.Command{
		Name:        "symbol-info",
		Description: "Get extended info for the current chart symbol",
		Handler: func(args []string, opts map[string]string) (interface{}, error) {
			return charttools.SymbolInfo()
		},
	})

	cli.Register(cli.Command{
		Name:        "symbol-search",
		Description: "Search TradingView symbols — tv symbol-search QUERY [--type TYPE] [--exchange EXCHANGE]",
		Handler: func(args []string, opts map[string]string) (interface{}, error) {
			if len(args) == 0 {
				return nil, fmt.Errorf("usage: tv symbol-search QUERY [--type TYPE] [--exchange EXCHANGE]")
			}
			return charttools.SymbolSearch(args[0], opts["type"], opts["exchange"])
		},
	})

	// ── indicator control (P6) ────────────────────────────────────────────────
	cli.Register(cli.Command{
		Name:        "indicator-toggle",
		Description: "Show/hide an indicator — tv indicator-toggle ENTITY_ID --visible=true|false",
		Handler: func(args []string, opts map[string]string) (interface{}, error) {
			if len(args) == 0 {
				return nil, fmt.Errorf("usage: tv indicator-toggle ENTITY_ID --visible=true|false")
			}
			visible := opts["visible"] != "false"
			return indicators.ToggleVisibility(args[0], visible)
		},
	})

	// ── pine script (P7) ──────────────────────────────────────────────────────
	cli.Register(cli.Command{
		Name:        "pine",
		Description: "Pine Script operations — tv pine <get|set|compile|smart-compile|errors|console|save|new|open|list|analyze|check> [args]",
		Handler: func(args []string, opts map[string]string) (interface{}, error) {
			if len(args) == 0 {
				return nil, fmt.Errorf("usage: tv pine <get|set|compile|smart-compile|errors|console|save|new|open|list|analyze|check>")
			}
			switch args[0] {
			case "get":
				return pine.GetSource()
			case "set":
				if len(args) < 2 {
					return nil, fmt.Errorf("usage: tv pine set SOURCE_CODE")
				}
				return pine.SetSource(strings.Join(args[1:], " "))
			case "compile", "raw-compile":
				return pine.Compile()
			case "smart-compile":
				return pine.SmartCompile()
			case "errors":
				return pine.GetErrors()
			case "console":
				return pine.GetConsole()
			case "save":
				return pine.Save()
			case "new":
				t := opts["type"]
				if t == "" && len(args) > 1 {
					t = args[1]
				}
				if t == "" {
					t = "indicator"
				}
				return pine.NewScript(t)
			case "open":
				if len(args) < 2 {
					return nil, fmt.Errorf("usage: tv pine open SCRIPT_NAME")
				}
				return pine.OpenScript(strings.Join(args[1:], " "))
			case "list":
				return pine.ListScripts()
			case "analyze":
				if len(args) < 2 {
					return nil, fmt.Errorf("usage: tv pine analyze SOURCE_CODE")
				}
				return pine.Analyze(strings.Join(args[1:], " ")), nil
			case "check":
				if len(args) < 2 {
					return nil, fmt.Errorf("usage: tv pine check SOURCE_CODE")
				}
				return pine.Check(strings.Join(args[1:], " "))
			default:
				return nil, fmt.Errorf("unknown pine subcommand %q", args[0])
			}
		},
	})

	// ── drawing (P8) ──────────────────────────────────────────────────────────
	cli.Register(cli.Command{
		Name:        "draw",
		Description: "Drawing operations — tv draw <shape|list|get|remove|clear> [args]",
		Handler: func(args []string, opts map[string]string) (interface{}, error) {
			if len(args) == 0 {
				return nil, fmt.Errorf("usage: tv draw <shape|list|get|remove|clear>")
			}
			switch args[0] {
			case "list":
				return drawing.ListDrawings()
			case "get":
				if len(args) < 2 {
					return nil, fmt.Errorf("usage: tv draw get ENTITY_ID")
				}
				return drawing.GetProperties(args[1])
			case "remove":
				if len(args) < 2 {
					return nil, fmt.Errorf("usage: tv draw remove ENTITY_ID")
				}
				return drawing.RemoveOne(args[1])
			case "clear":
				return drawing.ClearAll()
			case "shape":
				shapeName := opts["shape"]
				if shapeName == "" && len(args) > 1 {
					shapeName = args[1]
				}
				if shapeName == "" {
					return nil, fmt.Errorf("usage: tv draw shape SHAPE_NAME --time=TS --price=PRICE [--time2=TS --price2=PRICE] [--text=TEXT]")
				}
				t1, _ := strconv.ParseFloat(opts["time"], 64)
				p1, _ := strconv.ParseFloat(opts["price"], 64)
				da := drawing.DrawShapeArgs{
					Shape: shapeName,
					Point: drawing.DrawPoint{Time: t1, Price: p1},
					Text:  opts["text"],
				}
				if opts["time2"] != "" {
					t2, _ := strconv.ParseFloat(opts["time2"], 64)
					p2, _ := strconv.ParseFloat(opts["price2"], 64)
					da.Point2 = &drawing.DrawPoint{Time: t2, Price: p2}
				}
				return drawing.DrawShape(da)
			default:
				return nil, fmt.Errorf("unknown draw subcommand %q", args[0])
			}
		},
	})

	// ── pane (P10) ────────────────────────────────────────────────────────────
	cli.Register(cli.Command{
		Name:        "pane",
		Description: "Pane operations — tv pane <list|set-layout|focus|set-symbol> [args]",
		Handler: func(args []string, opts map[string]string) (interface{}, error) {
			if len(args) == 0 {
				return nil, fmt.Errorf("usage: tv pane <list|set-layout|focus|set-symbol>")
			}
			switch args[0] {
			case "list":
				return pane.ListPanes()
			case "set-layout":
				if len(args) < 2 {
					return nil, fmt.Errorf("usage: tv pane set-layout LAYOUT")
				}
				return pane.SetLayout(args[1])
			case "focus":
				if len(args) < 2 {
					return nil, fmt.Errorf("usage: tv pane focus INDEX")
				}
				idx, _ := strconv.Atoi(args[1])
				return pane.FocusPane(idx)
			case "set-symbol":
				if len(args) < 3 {
					return nil, fmt.Errorf("usage: tv pane set-symbol INDEX SYMBOL")
				}
				idx, _ := strconv.Atoi(args[1])
				return pane.SetPaneSymbol(idx, args[2])
			default:
				return nil, fmt.Errorf("unknown pane subcommand %q", args[0])
			}
		},
	})

	// ── replay (P11) ─────────────────────────────────────────────────────────
	cli.Register(cli.Command{
		Name:        "replay",
		Description: "Replay mode — tv replay <start|step|stop|status|autoplay|trade> [args]",
		Handler: func(args []string, opts map[string]string) (interface{}, error) {
			if len(args) == 0 {
				return nil, fmt.Errorf("usage: tv replay <start|step|stop|status|autoplay|trade>")
			}
			switch args[0] {
			case "start":
				return replay.Start(opts["date"])
			case "step":
				return replay.Step()
			case "stop":
				return replay.Stop()
			case "status":
				return replay.Status()
			case "autoplay":
				speed := 0
				if v, ok := opts["speed"]; ok {
					speed, _ = strconv.Atoi(v)
				}
				return replay.Autoplay(speed)
			case "trade":
				if len(args) < 2 {
					return nil, fmt.Errorf("usage: tv replay trade buy|sell|close")
				}
				return replay.Trade(args[1])
			default:
				return nil, fmt.Errorf("unknown replay subcommand %q", args[0])
			}
		},
	})

	// ── tab (P10) ─────────────────────────────────────────────────────────────
	cli.Register(cli.Command{
		Name:        "tab",
		Description: "Tab operations — tv tab <list|new|close|switch ID>",
		Handler: func(args []string, opts map[string]string) (interface{}, error) {
			if len(args) == 0 {
				return nil, fmt.Errorf("usage: tv tab <list|new|close|switch ID>")
			}
			switch args[0] {
			case "list":
				return tab.ListTabs()
			case "new":
				return tab.NewTab()
			case "close":
				return tab.CloseTab()
			case "switch":
				if len(args) < 2 {
					return nil, fmt.Errorf("usage: tv tab switch TAB_ID")
				}
				return tab.SwitchTab(args[1])
			default:
				return nil, fmt.Errorf("unknown tab subcommand %q", args[0])
			}
		},
	})

	// ── alerts (P9) ───────────────────────────────────────────────────────────
	cli.Register(cli.Command{
		Name:        "alert",
		Description: "Alert operations — tv alert <list|create|delete> [args]",
		Handler: func(args []string, opts map[string]string) (interface{}, error) {
			if len(args) == 0 {
				return nil, fmt.Errorf("usage: tv alert <list|create|delete>")
			}
			switch args[0] {
			case "list":
				return alerts.ListAlerts()
			case "create":
				price, _ := strconv.ParseFloat(opts["price"], 64)
				return alerts.CreateAlert(opts["condition"], price, opts["message"])
			case "delete":
				_, deleteAll := opts["all"]
				return alerts.DeleteAlerts(deleteAll)
			default:
				return nil, fmt.Errorf("unknown alert subcommand %q", args[0])
			}
		},
	})

	// ── ui automation (P12) ──────────────────────────────────────────────────
	cli.Register(cli.Command{
		Name:        "ui",
		Description: "UI automation — tv ui <click|open-panel|fullscreen|keyboard|type|hover|scroll|mouse|find|eval> [args]",
		Handler: func(args []string, opts map[string]string) (interface{}, error) {
			if len(args) == 0 {
				return nil, fmt.Errorf("usage: tv ui <click|open-panel|fullscreen|keyboard|type|hover|scroll|mouse|find|eval>")
			}
			switch args[0] {
			case "click":
				return uitools.Click(opts["by"], opts["value"])
			case "open-panel":
				return uitools.OpenPanel(opts["panel"], opts["action"])
			case "fullscreen":
				return uitools.Fullscreen()
			case "keyboard":
				if len(args) < 2 {
					return nil, fmt.Errorf("usage: tv ui keyboard KEY [--modifiers ctrl,shift]")
				}
				var mods []string
				if m, ok := opts["modifiers"]; ok && m != "" {
					for _, mod := range strings.Split(m, ",") {
						mods = append(mods, strings.TrimSpace(mod))
					}
				}
				return uitools.Keyboard(args[1], mods)
			case "type":
				if len(args) < 2 {
					return nil, fmt.Errorf("usage: tv ui type TEXT")
				}
				return uitools.TypeText(strings.Join(args[1:], " "))
			case "hover":
				return uitools.Hover(opts["by"], opts["value"])
			case "scroll":
				if len(args) < 2 {
					return nil, fmt.Errorf("usage: tv ui scroll up|down|left|right [--amount N]")
				}
				amount, _ := strconv.Atoi(opts["amount"])
				return uitools.Scroll(args[1], amount)
			case "mouse":
				x, _ := strconv.ParseFloat(opts["x"], 64)
				y, _ := strconv.ParseFloat(opts["y"], 64)
				_, dbl := opts["double"]
				return uitools.MouseClick(x, y, opts["button"], dbl)
			case "find":
				if len(args) < 2 {
					return nil, fmt.Errorf("usage: tv ui find QUERY [--strategy text|aria-label|css]")
				}
				return uitools.FindElement(strings.Join(args[1:], " "), opts["strategy"])
			case "eval":
				if len(args) < 2 {
					return nil, fmt.Errorf("usage: tv ui eval JS_EXPRESSION")
				}
				return uitools.Evaluate(strings.Join(args[1:], " "))
			default:
				return nil, fmt.Errorf("unknown ui subcommand %q", args[0])
			}
		},
	})

	// ── batch (P14) ───────────────────────────────────────────────────────────
	cli.Register(cli.Command{
		Name:        "batch",
		Description: "Batch operations — tv batch --symbols SYM1,SYM2 --action screenshot|get_ohlcv|get_strategy_results [--timeframes TF1,TF2] [--delay MS] [--count N]",
		Handler: func(args []string, opts map[string]string) (interface{}, error) {
			symStr := opts["symbols"]
			if symStr == "" {
				return nil, fmt.Errorf("usage: tv batch --symbols SYM1,SYM2 --action ACTION")
			}
			symbols := strings.Split(symStr, ",")
			for i := range symbols {
				symbols[i] = strings.TrimSpace(symbols[i])
			}
			var timeframes []string
			if tfStr := opts["timeframes"]; tfStr != "" {
				for _, tf := range strings.Split(tfStr, ",") {
					timeframes = append(timeframes, strings.TrimSpace(tf))
				}
			}
			delayMs, _ := strconv.Atoi(opts["delay"])
			count, _ := strconv.Atoi(opts["count"])
			return batch.BatchRun(symbols, timeframes, opts["action"], delayMs, count)
		},
	})

	// ── layout (P12) ──────────────────────────────────────────────────────────
	cli.Register(cli.Command{
		Name:        "layout",
		Description: "Layout operations — tv layout <list|switch NAME>",
		Handler: func(args []string, opts map[string]string) (interface{}, error) {
			if len(args) == 0 {
				return nil, fmt.Errorf("usage: tv layout <list|switch NAME>")
			}
			switch args[0] {
			case "list":
				return uitools.LayoutList()
			case "switch":
				if len(args) < 2 {
					return nil, fmt.Errorf("usage: tv layout switch NAME")
				}
				return uitools.LayoutSwitch(strings.Join(args[1:], " "))
			default:
				return nil, fmt.Errorf("unknown layout subcommand %q", args[0])
			}
		},
	})

	// ── watchlist (P9) ────────────────────────────────────────────────────────
	cli.Register(cli.Command{
		Name:        "watchlist",
		Description: "Watchlist operations — tv watchlist <get|add SYMBOL>",
		Handler: func(args []string, opts map[string]string) (interface{}, error) {
			if len(args) == 0 {
				return nil, fmt.Errorf("usage: tv watchlist <get|add SYMBOL>")
			}
			switch args[0] {
			case "get":
				return alerts.GetWatchlist()
			case "add":
				if len(args) < 2 {
					return nil, fmt.Errorf("usage: tv watchlist add SYMBOL")
				}
				return alerts.AddToWatchlist(args[1])
			default:
				return nil, fmt.Errorf("unknown watchlist subcommand %q", args[0])
			}
		},
	})
}

func main() {
	// Stream commands are special: they write JSONL forever and never return a
	// value, so they bypass the normal cli.Dispatch JSON-result path.
	args := os.Args[1:]
	if len(args) >= 2 && args[0] == "stream" {
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()

		sub := args[1]
		rest := args[2:]

		opts := make(map[string]string)
		for i := 0; i < len(rest); i++ {
			a := rest[i]
			if strings.HasPrefix(a, "--") {
				kv := strings.TrimPrefix(a, "--")
				if idx := strings.IndexByte(kv, '='); idx >= 0 {
					opts[kv[:idx]] = kv[idx+1:]
				} else if i+1 < len(rest) && !strings.HasPrefix(rest[i+1], "--") {
					opts[kv] = rest[i+1]
					i++
				} else {
					opts[kv] = "true"
				}
			}
		}

		intervalMs, _ := strconv.Atoi(opts["interval"])
		filter := opts["filter"]

		fmt.Fprintln(os.Stderr, "⚠  tradingview-mcp-go  |  Unofficial tool. Not affiliated with TradingView Inc. or Anthropic.")
		fmt.Fprintln(os.Stderr, "   Streams from your locally running TradingView Desktop instance only.")
		fmt.Fprintln(os.Stderr, "   Does not connect to TradingView servers. Requires --remote-debugging-port=9222.")
		fmt.Fprintln(os.Stderr, "   Ensure your usage complies with TradingView's Terms of Use.")

		var err error
		switch sub {
		case "quote":
			err = stream.StreamQuote(ctx, os.Stdout, os.Stderr, intervalMs)
		case "bars":
			err = stream.StreamBars(ctx, os.Stdout, os.Stderr, intervalMs)
		case "values":
			err = stream.StreamValues(ctx, os.Stdout, os.Stderr, intervalMs)
		case "lines":
			err = stream.StreamLines(ctx, os.Stdout, os.Stderr, intervalMs, filter)
		case "labels":
			err = stream.StreamLabels(ctx, os.Stdout, os.Stderr, intervalMs, filter)
		case "tables":
			err = stream.StreamTables(ctx, os.Stdout, os.Stderr, intervalMs, filter)
		case "all":
			err = stream.StreamAllPanes(ctx, os.Stdout, os.Stderr, intervalMs)
		default:
			fmt.Fprintf(os.Stderr, "unknown stream subcommand %q; valid: quote bars values lines labels tables all\n", sub)
			os.Exit(1)
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "stream error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	cli.Dispatch(args)
}
