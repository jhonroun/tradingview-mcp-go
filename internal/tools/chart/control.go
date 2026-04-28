package chart

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/jhonroun/tradingview-mcp-go/internal/cdp"
	"github.com/jhonroun/tradingview-mcp-go/internal/mcp"
	tv "github.com/jhonroun/tradingview-mcp-go/internal/tradingview"
)

var chartTypeMap = map[string]int{
	"bars": 0, "candles": 1, "line": 2, "area": 3,
	"renko": 4, "kagi": 5, "pointandfigure": 6, "linebreak": 7,
	"heikinashi": 8, "hollowcandles": 9,
}

// SetSymbol sets the chart symbol and waits for the chart to reload.
func SetSymbol(symbol string) (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()

	setExpr := fmt.Sprintf(`new Promise(function(resolve) {
		%s.setSymbol(%s, function() { setTimeout(resolve, 500); });
	})`, tv.ChartAPI, tv.SafeString(symbol))

	err := cdp.WithSession(ctx, func(c *cdp.Client, _ *cdp.Target) error {
		if _, err := c.Evaluate(ctx, setExpr, true); err != nil {
			return err
		}
		waitForChartReady(ctx, c, symbol)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{"success": true, "symbol": symbol}, nil
}

// SetTimeframe changes the chart resolution (e.g. "60", "D", "W").
func SetTimeframe(timeframe string) (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	expr := fmt.Sprintf(`(function() {
		%s.setResolution(%s, {}); return true;
	})()`, tv.ChartAPI, tv.SafeString(timeframe))

	err := cdp.WithSession(ctx, func(c *cdp.Client, _ *cdp.Target) error {
		_, err := c.Evaluate(ctx, expr, false)
		return err
	})
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{"success": true, "timeframe": timeframe}, nil
}

// SetType changes the chart type by name (Candles, Bars, Line, etc.).
func SetType(chartType string) (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	key := strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(chartType, " ", ""), "_", ""))
	typeNum, ok := chartTypeMap[key]
	if !ok {
		return nil, fmt.Errorf("unknown chart type %q; valid: Bars, Candles, Line, Area, Renko, Kagi, PointAndFigure, LineBreak, HeikinAshi, HollowCandles", chartType)
	}

	expr := fmt.Sprintf(`(function() { %s.setChartType(%d); return true; })()`, tv.ChartAPI, typeNum)

	err := cdp.WithSession(ctx, func(c *cdp.Client, _ *cdp.Target) error {
		_, err := c.Evaluate(ctx, expr, false)
		return err
	})
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{"success": true, "chartType": chartType, "typeNum": typeNum}, nil
}

// ManageIndicatorArgs holds parameters for chart_manage_indicator.
type ManageIndicatorArgs struct {
	Action         string        `json:"action"`
	Name           string        `json:"name"`
	EntityID       string        `json:"entity_id"`
	Inputs         []interface{} `json:"inputs"`
	AllowRemoveAny bool          `json:"allow_remove_any"`
}

type addStudyEvaluation struct {
	CreateError string      `json:"createError"`
	LimitText   string      `json:"limitText"`
	Before      []StudyInfo `json:"before"`
	After       []StudyInfo `json:"after"`
	NewStudies  []StudyInfo `json:"newStudies"`
}

type removeStudyEvaluation struct {
	Removed bool        `json:"removed"`
	Error   string      `json:"error"`
	Before  []StudyInfo `json:"before"`
	After   []StudyInfo `json:"after"`
}

// ManageIndicator adds or removes an indicator on the chart.
func ManageIndicator(args ManageIndicatorArgs) (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()

	switch strings.ToLower(args.Action) {
	case "add":
		return addIndicator(ctx, args.Name, args.Inputs, args.AllowRemoveAny)
	case "remove":
		return removeIndicator(ctx, args.EntityID)
	default:
		return nil, fmt.Errorf("unknown action %q; use add or remove", args.Action)
	}
}

const chartStudyControlJS = `
function studyControlDelay(ms) {
	return new Promise(function(resolve) { setTimeout(resolve, ms); });
}
function studyControlStudyInfo(s) {
	var id = "";
	var name = "";
	try { id = s.id || s.entityId || s._id || ""; } catch(e) {}
	try { name = s.name || s.title || s.shortName || ""; } catch(e) {}
	if (!name) {
		try {
			var mi = typeof s.metaInfo === "function" ? s.metaInfo() : s.metaInfo;
			name = mi && (mi.description || mi.shortDescription || mi.id || "");
		} catch(e) {}
	}
	return { id: String(id || ""), name: String(name || "unknown") };
}
function studyControlCollectStudies(chart) {
	try {
		var all = chart.getAllStudies() || [];
		return Array.prototype.map.call(all, studyControlStudyInfo).filter(function(s) {
			return s.id || s.name;
		});
	} catch(e) {
		return [];
	}
}
function studyControlCollectLimitText() {
	var candidates = [];
	var seen = {};
	function add(text) {
		text = String(text || "").replace(/\s+/g, " ").trim();
		if (!text || text.length > 2000 || seen[text]) return;
		seen[text] = true;
		candidates.push(text);
	}
	try {
		var selectors = '[role="dialog"],[data-name*="dialog"],[class*="dialog"],[class*="toast"],[class*="notification"],[class*="popup"],[class*="modal"]';
		Array.prototype.forEach.call(document.querySelectorAll(selectors), function(node) {
			add(node.innerText || node.textContent || "");
		});
	} catch(e) {}
	if (candidates.length === 0) {
		try {
			var re = /indicator|study|subscription|maximum|limit|available|upgrade|plan/i;
			var nodes = document.querySelectorAll("div,span");
			for (var i = 0; i < nodes.length && candidates.length < 20; i++) {
				var text = nodes[i].innerText || nodes[i].textContent || "";
				if (re.test(text)) add(text);
			}
		} catch(e) {}
	}
	return candidates.slice(0, 20).join("\n");
}
`

func addIndicator(ctx context.Context, name string, inputs []interface{}, allowRemoveAny bool) (map[string]interface{}, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("name is required for add action")
	}
	inputsJSON, _ := json.Marshal(inputs)
	if inputs == nil {
		inputsJSON = []byte("[]")
	}

	var first addStudyEvaluation
	var retry addStudyEvaluation
	var removed StudyInfo
	var removeEval removeStudyEvaluation
	var removalLogPath string
	var removalLogErr error

	err := cdp.WithSession(ctx, func(c *cdp.Client, _ *cdp.Target) error {
		var err error
		first, err = evaluateAddStudy(ctx, c, name, inputsJSON)
		if err != nil {
			return err
		}
		if _, ok := addedStudy(first); ok {
			return nil
		}
		details, limitReached := detectStudyLimit(first.CreateError, first.LimitText, first.After)
		if !limitReached || !allowRemoveAny {
			return nil
		}
		var ok bool
		removed, ok = selectStudyForRemoval(first.After)
		if !ok {
			return nil
		}
		removeEval, err = evaluateRemoveStudy(ctx, c, removed.ID)
		if err != nil {
			return err
		}
		if removeEval.Error != "" || !removeEval.Removed {
			return nil
		}
		removalLogPath, removalLogErr = appendStudyRemovalLog(studyRemovalLogEntry{
			EntityID:        removed.ID,
			Name:            removed.Name,
			Reason:          "study_limit_reached_allow_remove_any",
			RequestedName:   name,
			Limit:           details.Limit,
			CurrentStudies:  len(first.After),
			AllowRemoveAny:  true,
			TradingViewPath: "chart.removeEntity",
		})
		retry, err = evaluateAddStudy(ctx, c, name, inputsJSON)
		return err
	})
	if err != nil {
		return nil, err
	}

	if study, ok := addedStudy(first); ok {
		return buildStudyAddSuccessResult(name, first, study), nil
	}
	details, limitReached := detectStudyLimit(first.CreateError, first.LimitText, first.After)
	if limitReached && !allowRemoveAny {
		return buildStudyLimitResult("add", name, first.After, details), nil
	}
	if limitReached && allowRemoveAny {
		if removed.ID == "" {
			result := buildStudyLimitResult("add", name, first.After, details)
			result["allowRemoveAny"] = true
			result["error"] = "TradingView study limit reached, but no removable study entity was found."
			return result, nil
		}
		if removeEval.Error != "" || !removeEval.Removed {
			result := buildStudyLimitResult("add", name, first.After, details)
			result["allowRemoveAny"] = true
			result["selectedStudy"] = removed
			result["removeError"] = removeEval.Error
			result["error"] = "TradingView study limit reached, but the selected study could not be removed."
			return result, nil
		}
		if study, ok := addedStudy(retry); ok {
			result := buildStudyAddSuccessResult(name, retry, study)
			result["limitWasReached"] = true
			result["allowRemoveAny"] = true
			result["removedStudy"] = removed
			result["currentStudiesBeforeRemoval"] = normalizeStudyInfos(first.After)
			result["removal_logged"] = removalLogErr == nil
			if removalLogPath != "" {
				result["removal_log_path"] = removalLogPath
			}
			if removalLogErr != nil {
				result["removal_log_error"] = removalLogErr.Error()
			}
			return result, nil
		}
		retryDetails, retryLimit := detectStudyLimit(retry.CreateError, retry.LimitText, retry.After)
		if retryLimit {
			result := buildStudyLimitResult("add", name, retry.After, retryDetails)
			result["allowRemoveAny"] = true
			result["removedStudy"] = removed
			result["removal_logged"] = removalLogErr == nil
			if removalLogPath != "" {
				result["removal_log_path"] = removalLogPath
			}
			if removalLogErr != nil {
				result["removal_log_error"] = removalLogErr.Error()
			}
			return result, nil
		}
		result := buildStudyAddFailedResult(name, retry)
		result["allowRemoveAny"] = true
		result["removedStudy"] = removed
		result["removal_logged"] = removalLogErr == nil
		if removalLogPath != "" {
			result["removal_log_path"] = removalLogPath
		}
		if removalLogErr != nil {
			result["removal_log_error"] = removalLogErr.Error()
		}
		return result, nil
	}
	return buildStudyAddFailedResult(name, first), nil
}

func removeIndicator(ctx context.Context, entityID string) (map[string]interface{}, error) {
	if entityID == "" {
		return nil, fmt.Errorf("entity_id is required for remove action")
	}
	expr := fmt.Sprintf(`(function() {
		%s.removeEntity(%s); return true;
	})()`, tv.ChartAPI, tv.SafeString(entityID))

	err := cdp.WithSession(ctx, func(c *cdp.Client, _ *cdp.Target) error {
		_, err := c.Evaluate(ctx, expr, false)
		return err
	})
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{"success": true, "status": "ok", "action": "remove", "entityId": entityID}, nil
}

func evaluateAddStudy(ctx context.Context, c *cdp.Client, name string, inputsJSON []byte) (addStudyEvaluation, error) {
	expr := fmt.Sprintf(`(async function() {
		%s
		var chart = %s;
		var before = studyControlCollectStudies(chart);
		var beforeIds = {};
		before.forEach(function(s) { if (s.id) beforeIds[s.id] = true; });
		var createError = "";
		try {
			await Promise.resolve(chart.createStudy(%s, false, false, %s));
		} catch(e) {
			createError = String((e && (e.message || e.description)) || e || "");
		}
		await studyControlDelay(1500);
		var after = studyControlCollectStudies(chart);
		var newStudies = after.filter(function(s) { return s.id && !beforeIds[s.id]; });
		return {
			createError: createError,
			limitText: studyControlCollectLimitText(),
			before: before,
			after: after,
			newStudies: newStudies
		};
	})()`, chartStudyControlJS, tv.ChartAPI, tv.SafeString(name), string(inputsJSON))

	raw, err := c.EvaluateWithOptions(ctx, expr, cdp.EvaluateOptions{
		AwaitPromise:  true,
		ReturnByValue: true,
		Timeout:       20 * time.Second,
	})
	if err != nil {
		return addStudyEvaluation{}, err
	}
	var result addStudyEvaluation
	if err := json.Unmarshal(raw, &result); err != nil {
		return addStudyEvaluation{}, fmt.Errorf("parse add indicator result: %w", err)
	}
	result.Before = normalizeStudyInfos(result.Before)
	result.After = normalizeStudyInfos(result.After)
	result.NewStudies = normalizeStudyInfos(result.NewStudies)
	return result, nil
}

func evaluateRemoveStudy(ctx context.Context, c *cdp.Client, entityID string) (removeStudyEvaluation, error) {
	expr := fmt.Sprintf(`(async function() {
		%s
		var chart = %s;
		var before = studyControlCollectStudies(chart);
		var removeError = "";
		try {
			chart.removeEntity(%s);
		} catch(e) {
			removeError = String((e && (e.message || e.description)) || e || "");
		}
		await studyControlDelay(1000);
		var after = studyControlCollectStudies(chart);
		var stillPresent = after.some(function(s) { return s.id === %s; });
		return {
			removed: !removeError && !stillPresent,
			error: removeError,
			before: before,
			after: after
		};
	})()`, chartStudyControlJS, tv.ChartAPI, tv.SafeString(entityID), tv.SafeString(entityID))

	raw, err := c.EvaluateWithOptions(ctx, expr, cdp.EvaluateOptions{
		AwaitPromise:  true,
		ReturnByValue: true,
		Timeout:       8 * time.Second,
	})
	if err != nil {
		return removeStudyEvaluation{}, err
	}
	var result removeStudyEvaluation
	if err := json.Unmarshal(raw, &result); err != nil {
		return removeStudyEvaluation{}, fmt.Errorf("parse remove indicator result: %w", err)
	}
	result.Before = normalizeStudyInfos(result.Before)
	result.After = normalizeStudyInfos(result.After)
	return result, nil
}

func addedStudy(result addStudyEvaluation) (StudyInfo, bool) {
	for _, study := range result.NewStudies {
		if strings.TrimSpace(study.ID) != "" {
			return study, true
		}
	}
	return StudyInfo{}, false
}

func buildStudyAddSuccessResult(requestedName string, result addStudyEvaluation, study StudyInfo) map[string]interface{} {
	return map[string]interface{}{
		"success":        true,
		"status":         "ok",
		"action":         "add",
		"entityId":       study.ID,
		"name":           requestedName,
		"study":          study,
		"currentStudies": normalizeStudyInfos(result.After),
	}
}

func buildStudyAddFailedResult(requestedName string, result addStudyEvaluation) map[string]interface{} {
	out := map[string]interface{}{
		"success":           false,
		"status":            "study_add_failed",
		"error":             "TradingView did not add a new study.",
		"action":            "add",
		"requestedName":     requestedName,
		"currentStudies":    normalizeStudyInfos(result.After),
		"currentStudyCount": len(result.After),
	}
	if result.CreateError != "" {
		out["create_error"] = result.CreateError
	}
	if result.LimitText != "" {
		out["ui_message"] = compactLimitMessage(result.LimitText)
	}
	return out
}

// SetVisibleRange zooms the chart to show bars between from and to (Unix seconds).
func SetVisibleRange(from, to int64) (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	expr := fmt.Sprintf(`(function() {
		var chart = %s;
		var bars = %s;
		var size = bars.size();
		var fromTs = %d * 1000, toTs = %d * 1000;
		var fromIdx = 0, toIdx = size - 1;
		for (var i = 0; i < size; i++) {
			var bar = bars.get(i);
			if (!bar) continue;
			var t = bar.time;
			if (t <= fromTs) fromIdx = i;
			if (t <= toTs) toIdx = i;
		}
		chart.zoomToBarsRange(fromIdx, toIdx);
		return { success: true, fromIdx: fromIdx, toIdx: toIdx };
	})()`, tv.ChartAPI, tv.BarsPath, from, to)

	var raw json.RawMessage
	err := cdp.WithSession(ctx, func(c *cdp.Client, _ *cdp.Target) error {
		time.Sleep(500 * time.Millisecond)
		var err error
		raw, err = c.Evaluate(ctx, expr, false)
		return err
	})
	if err != nil {
		return nil, err
	}
	var result map[string]interface{}
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("parse range result: %w", err)
	}
	result["from"] = from
	result["to"] = to
	return result, nil
}

// ScrollToDate scrolls the chart to center on the bar at the given Unix timestamp.
func ScrollToDate(ts int64) (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	expr := fmt.Sprintf(`(function() {
		var chart = %s;
		var bars = %s;
		var size = bars.size();
		var target = %d * 1000;
		var idx = 0;
		for (var i = 0; i < size; i++) {
			var bar = bars.get(i);
			if (!bar) continue;
			if (bar.time <= target) idx = i;
		}
		var from = Math.max(0, idx - 25);
		var to = Math.min(size - 1, idx + 25);
		chart.zoomToBarsRange(from, to);
		return { success: true, barIdx: idx };
	})()`, tv.ChartAPI, tv.BarsPath, ts)

	var raw json.RawMessage
	err := cdp.WithSession(ctx, func(c *cdp.Client, _ *cdp.Target) error {
		var err error
		raw, err = c.Evaluate(ctx, expr, false)
		return err
	})
	if err != nil {
		return nil, err
	}
	var result map[string]interface{}
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("parse scroll result: %w", err)
	}
	result["timestamp"] = ts
	return result, nil
}

func registerControlTools(reg *mcp.Registry) {
	reg.Register(mcp.ToolDef{
		Name:        "chart_set_symbol",
		Description: "Change the chart symbol and wait for it to reload with new data",
		Schema: mcp.InputSchema{
			Type: "object",
			Properties: map[string]mcp.PropertySchema{
				"symbol": {Type: "string", Description: "Symbol to set (e.g. BTCUSDT, NASDAQ:AAPL)"},
			},
			Required: []string{"symbol"},
		},
		Handler: func(args json.RawMessage) (interface{}, error) {
			var p struct {
				Symbol string `json:"symbol"`
			}
			if err := json.Unmarshal(args, &p); err != nil {
				return map[string]interface{}{"success": false, "error": err.Error()}, nil
			}
			result, err := SetSymbol(p.Symbol)
			if err != nil {
				return map[string]interface{}{"success": false, "error": err.Error()}, nil
			}
			return result, nil
		},
	})

	reg.Register(mcp.ToolDef{
		Name:        "chart_set_timeframe",
		Description: "Change the chart resolution/timeframe (e.g. 1, 5, 15, 60, 240, D, W, M)",
		Schema: mcp.InputSchema{
			Type: "object",
			Properties: map[string]mcp.PropertySchema{
				"timeframe": {Type: "string", Description: "Resolution string: 1, 5, 15, 30, 60, 240, D, W, M"},
			},
			Required: []string{"timeframe"},
		},
		Handler: func(args json.RawMessage) (interface{}, error) {
			var p struct {
				Timeframe string `json:"timeframe"`
			}
			if err := json.Unmarshal(args, &p); err != nil {
				return map[string]interface{}{"success": false, "error": err.Error()}, nil
			}
			result, err := SetTimeframe(p.Timeframe)
			if err != nil {
				return map[string]interface{}{"success": false, "error": err.Error()}, nil
			}
			return result, nil
		},
	})

	reg.Register(mcp.ToolDef{
		Name:        "chart_set_type",
		Description: "Change the chart type: Bars, Candles, Line, Area, Renko, Kagi, PointAndFigure, LineBreak, HeikinAshi, HollowCandles",
		Schema: mcp.InputSchema{
			Type: "object",
			Properties: map[string]mcp.PropertySchema{
				"chart_type": {Type: "string", Description: "Chart type name"},
			},
			Required: []string{"chart_type"},
		},
		Handler: func(args json.RawMessage) (interface{}, error) {
			var p struct {
				ChartType string `json:"chart_type"`
			}
			if err := json.Unmarshal(args, &p); err != nil {
				return map[string]interface{}{"success": false, "error": err.Error()}, nil
			}
			result, err := SetType(p.ChartType)
			if err != nil {
				return map[string]interface{}{"success": false, "error": err.Error()}, nil
			}
			return result, nil
		},
	})

	reg.Register(mcp.ToolDef{
		Name:        "chart_manage_indicator",
		Description: "Add or remove an indicator. action=add requires name; action=remove requires entity_id",
		Schema: mcp.InputSchema{
			Type: "object",
			Properties: map[string]mcp.PropertySchema{
				"action":    {Type: "string", Description: "add or remove"},
				"name":      {Type: "string", Description: "Indicator name for add (e.g. RSI, MACD, Bollinger Bands)"},
				"entity_id": {Type: "string", Description: "Entity ID for remove (from chart_get_state)"},
				"inputs":    {Type: "array", Description: "Optional input values array for add"},
				"allow_remove_any": {
					Type:        "boolean",
					Description: "For add only: if TradingView reports a study limit, explicitly allow removing the most recent existing study, logging it to research, and retrying once",
				},
			},
			Required: []string{"action"},
		},
		Handler: func(args json.RawMessage) (interface{}, error) {
			var p ManageIndicatorArgs
			if err := json.Unmarshal(args, &p); err != nil {
				return map[string]interface{}{"success": false, "error": err.Error()}, nil
			}
			result, err := ManageIndicator(p)
			if err != nil {
				return map[string]interface{}{"success": false, "error": err.Error()}, nil
			}
			return result, nil
		},
	})

	reg.Register(mcp.ToolDef{
		Name:        "chart_set_visible_range",
		Description: "Zoom the chart to show bars between from and to timestamps (Unix seconds)",
		Schema: mcp.InputSchema{
			Type: "object",
			Properties: map[string]mcp.PropertySchema{
				"from": {Type: "number", Description: "Start timestamp (Unix seconds)"},
				"to":   {Type: "number", Description: "End timestamp (Unix seconds)"},
			},
			Required: []string{"from", "to"},
		},
		Handler: func(args json.RawMessage) (interface{}, error) {
			var p struct {
				From int64 `json:"from"`
				To   int64 `json:"to"`
			}
			if err := json.Unmarshal(args, &p); err != nil {
				return map[string]interface{}{"success": false, "error": err.Error()}, nil
			}
			result, err := SetVisibleRange(p.From, p.To)
			if err != nil {
				return map[string]interface{}{"success": false, "error": err.Error()}, nil
			}
			return result, nil
		},
	})

	reg.Register(mcp.ToolDef{
		Name:        "chart_scroll_to_date",
		Description: "Scroll the chart to center on a specific date/timestamp (Unix seconds)",
		Schema: mcp.InputSchema{
			Type: "object",
			Properties: map[string]mcp.PropertySchema{
				"timestamp": {Type: "number", Description: "Unix timestamp in seconds"},
			},
			Required: []string{"timestamp"},
		},
		Handler: func(args json.RawMessage) (interface{}, error) {
			var p struct {
				Timestamp int64 `json:"timestamp"`
			}
			if err := json.Unmarshal(args, &p); err != nil {
				return map[string]interface{}{"success": false, "error": err.Error()}, nil
			}
			result, err := ScrollToDate(p.Timestamp)
			if err != nil {
				return map[string]interface{}{"success": false, "error": err.Error()}, nil
			}
			return result, nil
		},
	})
}
