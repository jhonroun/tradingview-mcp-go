// Package health implements tv_health_check, tv_discover, tv_ui_state, and tv_launch.
package health

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jhonroun/tradingview-mcp-go/internal/cdp"
	"github.com/jhonroun/tradingview-mcp-go/internal/launcher"
	"github.com/jhonroun/tradingview-mcp-go/internal/mcp"
)

// HealthCheckResult mirrors the Node.js healthCheck() response shape.
type HealthCheckResult struct {
	Success   bool   `json:"success"`
	Connected bool   `json:"connected"`
	TargetURL string `json:"targetUrl,omitempty"`
	TargetID  string `json:"targetId,omitempty"`
	Error     string `json:"error,omitempty"`
	Hint      string `json:"hint,omitempty"`
}

// LaunchArgs are the optional arguments for tv_launch / Launch().
type LaunchArgs struct {
	Port         *int    `json:"port,omitempty"`
	KillExisting *bool   `json:"kill_existing,omitempty"`
	TvPath       *string `json:"tv_path,omitempty"`
}

type compatibilityProbe struct {
	Name        string                 `json:"name"`
	Path        string                 `json:"path"`
	Purpose     string                 `json:"purpose"`
	Status      string                 `json:"status"`
	Compatible  bool                   `json:"compatible"`
	Available   bool                   `json:"available"`
	Stability   string                 `json:"stability"`
	Reliability string                 `json:"reliability,omitempty"`
	Details     map[string]interface{} `json:"details,omitempty"`
	Error       string                 `json:"error,omitempty"`
}

// HealthCheck verifies CDP connectivity and returns chart target info.
func HealthCheck() (*HealthCheckResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	targets, err := cdp.ListTargets(ctx, "localhost", 9222)
	if err != nil {
		return &HealthCheckResult{
			Success:   false,
			Connected: false,
			Error:     err.Error(),
			Hint:      "TradingView is not running with CDP enabled. Use the tv_launch tool to start it automatically.",
		}, nil
	}
	target, err := cdp.FindChartTarget(targets)
	if err != nil {
		return &HealthCheckResult{
			Success:   false,
			Connected: true,
			Error:     err.Error(),
			Hint:      "TradingView is running but no chart is open. Open a chart to enable chart tools.",
		}, nil
	}
	return &HealthCheckResult{
		Success:   true,
		Connected: true,
		TargetURL: target.URL,
		TargetID:  target.ID,
	}, nil
}

// Discover reports which known TradingView API paths are reachable.
func Discover() (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	targets, err := cdp.ListTargets(ctx, "localhost", 9222)
	if err != nil {
		return nil, fmt.Errorf("CDP not available: %w", err)
	}
	target, err := cdp.FindChartTarget(targets)
	if err != nil {
		return nil, err
	}
	client, err := cdp.Connect(ctx, target)
	if err != nil {
		return nil, err
	}
	defer client.Close()
	if err := client.EnableDomains(ctx); err != nil {
		return nil, err
	}

	knownPaths := map[string]string{
		"chartApi":              "window.TradingViewApi._activeChartWidgetWV.value()",
		"chartWidgetCollection": "window.TradingViewApi._chartWidgetCollection",
		"replayApi":             "window.TradingViewApi._replayApi",
		"alertService":          "window.TradingViewApi._alertService",
	}
	available := make(map[string]bool, len(knownPaths))
	for name, expr := range knownPaths {
		check := fmt.Sprintf("typeof (%s) !== 'undefined' && (%s) !== null", expr, expr)
		val, err := client.Evaluate(ctx, check, false)
		if err == nil {
			var b bool
			_ = json.Unmarshal(val, &b)
			available[name] = b
		} else {
			available[name] = false
		}
	}
	probes, probeErr := runCompatibilityProbes(ctx, client)
	result := map[string]interface{}{
		"success":                        true,
		"paths":                          available,
		"unstable_internal_paths":        true,
		"compatibility_probe_count":      len(probes),
		"compatibility_probes":           probes,
		"compatibility_probe_contract":   "compatible=true means the internal path/method exists; available=true means useful data is present in the current chart state",
		"compatibility_probe_limitation": "TradingView internal paths are undocumented and must be re-probed after TradingView updates.",
	}
	if probeErr != nil {
		result["compatibility_probe_error"] = probeErr.Error()
	}
	return result, nil
}

func runCompatibilityProbes(ctx context.Context, client *cdp.Client) ([]compatibilityProbe, error) {
	raw, err := client.EvaluateWithOptions(ctx, buildCompatibilityProbeJS(), cdp.EvaluateOptions{
		AwaitPromise:  true,
		ReturnByValue: true,
		Timeout:       12 * time.Second,
	})
	if err != nil {
		return []compatibilityProbe{{
			Name:       "compatibility_probe_runtime",
			Path:       "Runtime.evaluate",
			Purpose:    "Run non-mutating compatibility probes for undocumented TradingView internals.",
			Status:     "error",
			Compatible: false,
			Available:  false,
			Stability:  "unstable_internal_path",
			Error:      err.Error(),
		}}, err
	}
	var probes []compatibilityProbe
	if err := json.Unmarshal(raw, &probes); err != nil {
		return []compatibilityProbe{{
			Name:       "compatibility_probe_parse",
			Path:       "Runtime.evaluate",
			Purpose:    "Parse TradingView internal compatibility probe result.",
			Status:     "error",
			Compatible: false,
			Available:  false,
			Stability:  "unstable_internal_path",
			Error:      err.Error(),
		}}, err
	}
	return probes, nil
}

func buildCompatibilityProbeJS() string {
	return `(async function() {
		var STABILITY = "unstable_internal_path";
		var RELIABILITY_STUDY = "reliable_pine_runtime_value_unstable_internal_path";
		var RELIABILITY_BACKTEST = "reliable_backtesting_report_unstable_internal_path";
		var probes = [];

		function unwrap(value) {
			var current = value;
			for (var i = 0; i < 5; i++) {
				if (current == null) return current;
				if (typeof current === "function") {
					current = current();
					continue;
				}
				if (typeof current === "object" && typeof current.value === "function") {
					current = current.value();
					continue;
				}
				if (typeof current === "object" && Object.prototype.hasOwnProperty.call(current, "_value")) {
					current = current._value;
					continue;
				}
				return current;
			}
			return current;
		}
		function arrayify(value) {
			value = unwrap(value);
			if (Array.isArray(value)) return value;
			if (!value || typeof value === "string") return [];
			if (typeof value.length === "number") {
				try { return Array.prototype.slice.call(value); } catch(e) {}
			}
			if (Array.isArray(value.values)) return value.values;
			return [];
		}
		function add(name, path, purpose, status, compatible, available, reliability, details, error) {
			var probe = {
				name: name,
				path: path,
				purpose: purpose,
				status: status,
				compatible: !!compatible,
				available: !!available,
				stability: STABILITY
			};
			if (reliability) probe.reliability = reliability;
			if (details) probe.details = details;
			if (error) probe.error = String(error);
			probes.push(probe);
		}
		function sourceId(source) {
			try {
				var id = source && typeof source.id === "function" ? source.id() : unwrap(source && source.id);
				if (id != null && id !== "") return String(id);
			} catch(e) {}
			try {
				var privateId = unwrap(source && source._id);
				if (privateId != null && privateId !== "") return String(privateId);
			} catch(e) {}
			return "";
		}
		function sourceName(source) {
			try {
				var meta = source && typeof source.metaInfo === "function" ? source.metaInfo() : unwrap(source && source.metaInfo);
				if (meta) return meta.description || meta.shortDescription || meta.name || "";
			} catch(e) {}
			try {
				var title = source && typeof source.title === "function" ? source.title() : unwrap(source && source.title);
				if (title) return String(title);
			} catch(e) {}
			return "";
		}
		function plotInfos(meta) {
			var plots = meta && Array.isArray(meta.plots) ? meta.plots : [];
			var styles = meta && meta.styles ? meta.styles : {};
			var infos = [];
			for (var i = 0; i < plots.length; i++) {
				var plot = plots[i] || {};
				var id = String(plot.id || ("plot_" + i));
				var style = styles[id] || {};
				var title = style.title || plot.title || plot.name || id;
				infos.push({ plot_id: id, name: String(title || id), value_index: i + 1 });
			}
			return infos;
		}
		function hasEquityPlot(source) {
			var meta = null;
			try { meta = source && typeof source.metaInfo === "function" ? source.metaInfo() : unwrap(source && source.metaInfo); } catch(e) {}
			var infos = plotInfos(meta);
			for (var i = 0; i < infos.length; i++) {
				var name = String(infos[i].name || "").toLowerCase();
				var id = String(infos[i].plot_id || "").toLowerCase();
				if (name.indexOf("strategy equity") >= 0 || id.indexOf("strategy equity") >= 0) {
					return { found: true, plot: infos[i], available_plots: infos };
				}
			}
			return { found: false, available_plots: infos };
		}

		var api = window.TradingViewApi;
		add("tradingview_api", "window.TradingViewApi", "Root TradingView Desktop API object.", api ? "ok" : "unavailable", !!api, !!api, "", null, "");
		var activeWidget = null;
		try { activeWidget = api ? unwrap(api._activeChartWidgetWV) : null; } catch(e) {}
		add("active_chart_widget", "window.TradingViewApi._activeChartWidgetWV.value()", "Active chart widget used by chart/data tools.", activeWidget ? "ok" : "unavailable", !!activeWidget, !!activeWidget, "", null, "");

		var chartWidget = activeWidget && (activeWidget._chartWidget || activeWidget);
		var chartModel = null;
		var model = null;
		try { chartModel = chartWidget && typeof chartWidget.model === "function" ? chartWidget.model() : null; } catch(e) {}
		try { model = chartModel && typeof chartModel.model === "function" ? chartModel.model() : chartModel; } catch(e) {}
		add("chart_model", "chart.model().model()", "Internal chart model used to reach studies and strategies.", model ? "ok" : "unavailable", !!model, !!model, "", null, "");

		var dataSources = [];
		var hasDataSources = !!(model && typeof model.dataSources === "function");
		try { if (hasDataSources) dataSources = arrayify(model.dataSources()); } catch(e) {}
		add("data_sources", "model.dataSources()", "Internal study/source collection.", hasDataSources ? "ok" : "unavailable", hasDataSources, dataSources.length > 0, RELIABILITY_STUDY, { source_count: dataSources.length }, "");

		var studyModelOK = false;
		for (var dsI = 0; dsI < dataSources.length; dsI++) {
			try {
				var data = dataSources[dsI] && typeof dataSources[dsI].data === "function" ? dataSources[dsI].data() : unwrap(dataSources[dsI] && dataSources[dsI].data);
				if (data && typeof data.valueAt === "function" && typeof data.fullRangeIterator === "function") {
					studyModelOK = true;
					break;
				}
			} catch(e) {}
		}
		add("study_model_data", "study.data().valueAt()/fullRangeIterator()", "Numeric Pine runtime values for indicators/studies.", studyModelOK ? "ok" : "unavailable", studyModelOK, studyModelOK, RELIABILITY_STUDY, { source_count: dataSources.length }, "");

		var strategySources = [];
		var hasStrategySources = !!(model && typeof model.strategySources === "function");
		try { if (hasStrategySources) strategySources = arrayify(model.strategySources()); } catch(e) {}
		add("strategy_sources", "model.strategySources()", "Internal strategy source collection.", hasStrategySources ? (strategySources.length > 0 ? "ok" : "no_strategy_loaded") : "unavailable", hasStrategySources, strategySources.length > 0, RELIABILITY_BACKTEST, { strategy_source_count: strategySources.length }, "");

		var activeStrategy = null;
		var hasActiveStrategySource = !!(model && typeof model.activeStrategySource === "function");
		try { if (hasActiveStrategySource) activeStrategy = unwrap(model.activeStrategySource()); } catch(e) {}
		if (!activeStrategy && strategySources.length > 0) activeStrategy = strategySources[0];
		add("active_strategy_source", "model.activeStrategySource()", "Active TradingView strategy source.", hasActiveStrategySource || strategySources.length > 0 ? (activeStrategy ? "ok" : "no_strategy_loaded") : "unavailable", hasActiveStrategySource || strategySources.length > 0, !!activeStrategy, RELIABILITY_BACKTEST, { active_strategy_id: sourceId(activeStrategy), active_strategy_name: sourceName(activeStrategy) }, "");

		var hasBacktestingAPI = !!(api && typeof api.backtestingStrategyApi === "function");
		if (!hasBacktestingAPI) {
			add("backtesting_strategy_api", "window.TradingViewApi.backtestingStrategyApi()", "Async Strategy Tester report API.", "unavailable", false, false, RELIABILITY_BACKTEST, null, "");
		} else {
			try {
				var bt = await api.backtestingStrategyApi();
				var report = bt && unwrap(bt.activeStrategyReportData || bt._activeStrategyReportData || bt._reportData);
				var reportKeys = report && typeof report === "object" ? Object.keys(report) : [];
				var reportStatus = report ? "ok" : "strategy_report_unavailable";
				var reportAvailable = !!report;
				if (!activeStrategy && strategySources.length === 0) {
					reportStatus = "no_strategy_loaded";
					reportAvailable = false;
				}
				add("backtesting_strategy_api", "await window.TradingViewApi.backtestingStrategyApi()", "Async Strategy Tester report API.", reportStatus, true, reportAvailable, RELIABILITY_BACKTEST, { report_keys: reportKeys }, "");
			} catch(e) {
				add("backtesting_strategy_api", "await window.TradingViewApi.backtestingStrategyApi()", "Async Strategy Tester report API.", "error", true, false, RELIABILITY_BACKTEST, null, e && e.message ? e.message : e);
			}
		}

		if (!activeStrategy) {
			add("strategy_equity_plot", "model.strategySources()[0].data().fullRangeIterator()", "Explicit Pine Strategy Equity plot for loaded-bar equity.", "no_strategy_loaded", hasStrategySources, false, RELIABILITY_STUDY, null, "");
		} else {
			var eq = hasEquityPlot(activeStrategy);
			add("strategy_equity_plot", "strategySource.metaInfo().plots/styles + data().fullRangeIterator()", "Explicit Pine Strategy Equity plot for loaded-bar equity.", eq.found ? "ok" : "needs_equity_plot", true, eq.found, RELIABILITY_STUDY, { available_plots: eq.available_plots, plot: eq.plot || null, coverage: "loaded_chart_bars" }, "");
		}

		return probes;
	})()`
}

// UIState returns currently visible UI panels via DOM inspection.
func UIState() (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	targets, err := cdp.ListTargets(ctx, "localhost", 9222)
	if err != nil {
		return nil, fmt.Errorf("CDP not available: %w", err)
	}
	target, err := cdp.FindChartTarget(targets)
	if err != nil {
		return nil, err
	}
	client, err := cdp.Connect(ctx, target)
	if err != nil {
		return nil, err
	}
	defer client.Close()
	if err := client.EnableDomains(ctx); err != nil {
		return nil, err
	}

	const expr = `(function() {
		var panels = [];
		document.querySelectorAll('[data-name]').forEach(function(el) {
			if (el.offsetParent !== null) panels.push(el.getAttribute('data-name'));
		});
		return JSON.stringify({success: true, visiblePanels: panels.filter(function(v,i,a){return a.indexOf(v)===i;})});
	})()`

	raw, err := client.Evaluate(ctx, expr, false)
	if err != nil {
		return nil, err
	}
	// raw is a JSON string containing the stringified result.
	var jsonStr string
	if err := json.Unmarshal(raw, &jsonStr); err != nil {
		return nil, fmt.Errorf("parse ui state: %w", err)
	}
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, fmt.Errorf("parse ui state JSON: %w", err)
	}
	return result, nil
}

// Launch starts TradingView with CDP enabled.
func Launch(args LaunchArgs) (map[string]interface{}, error) {
	port := 9222
	if args.Port != nil {
		port = *args.Port
	}
	killExisting := true
	if args.KillExisting != nil {
		killExisting = *args.KillExisting
	}
	tvPath := ""
	if args.TvPath != nil {
		tvPath = *args.TvPath
	}
	return launcher.Launch(port, killExisting, tvPath)
}

// RegisterTools registers all health-group tools into the MCP registry.
func RegisterTools(reg *mcp.Registry) {
	reg.Register(mcp.ToolDef{
		Name:        "tv_health_check",
		Description: "Check CDP connection to TradingView and return current chart state",
		Schema:      mcp.InputSchema{Type: "object"},
		Handler: func(args json.RawMessage) (interface{}, error) {
			result, err := HealthCheck()
			if err != nil {
				return map[string]interface{}{
					"success": false,
					"error":   err.Error(),
					"hint":    "TradingView is not running with CDP enabled. Use the tv_launch tool to start it automatically.",
				}, nil
			}
			return result, nil
		},
	})

	reg.Register(mcp.ToolDef{
		Name:        "tv_discover",
		Description: "Report which known TradingView API paths are available and their methods",
		Schema:      mcp.InputSchema{Type: "object"},
		Handler: func(args json.RawMessage) (interface{}, error) {
			result, err := Discover()
			if err != nil {
				return map[string]interface{}{"success": false, "error": err.Error()}, nil
			}
			return result, nil
		},
	})

	reg.Register(mcp.ToolDef{
		Name:        "tv_ui_state",
		Description: "Get current UI state: which panels are open, what buttons are visible/enabled/disabled",
		Schema:      mcp.InputSchema{Type: "object"},
		Handler: func(args json.RawMessage) (interface{}, error) {
			result, err := UIState()
			if err != nil {
				return map[string]interface{}{"success": false, "error": err.Error()}, nil
			}
			return result, nil
		},
	})

	reg.Register(mcp.ToolDef{
		Name:        "tv_launch",
		Description: "Launch TradingView Desktop with Chrome DevTools Protocol (remote debugging) enabled. Auto-detects install location on Mac, Windows, and Linux.",
		Schema: mcp.InputSchema{
			Type: "object",
			Properties: map[string]mcp.PropertySchema{
				"port":          {Type: "number", Description: "CDP port (default 9222)"},
				"kill_existing": {Type: "boolean", Description: "Kill existing TradingView instances first (default true)"},
				"tv_path":       {Type: "string", Description: "Explicit path to TradingView executable (overrides auto-discovery)"},
			},
		},
		Handler: func(args json.RawMessage) (interface{}, error) {
			var a LaunchArgs
			if len(args) > 0 {
				if err := json.Unmarshal(args, &a); err != nil {
					return nil, fmt.Errorf("invalid arguments: %w", err)
				}
			}
			return Launch(a)
		},
	})
}
