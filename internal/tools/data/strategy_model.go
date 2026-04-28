package data

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/jhonroun/tradingview-mcp-go/internal/cdp"
)

const (
	SourceTradingViewBacktestingAPI        = "tradingview_backtesting_api"
	SourceTradingViewStrategyPlot          = "tradingview_strategy_plot"
	SourceDerivedFromOHLCVAndTrades        = "derived_from_ohlcv_and_trades"
	SourceDerivedFromBacktestingTrades     = "derived_from_backtesting_report_trades"
	ReliabilityBacktestingUnstableInternal = "reliable_backtesting_report_unstable_internal_path"
	ReliabilityDerivedTradeExitPoints      = "derived_from_backtesting_report_trades_exit_points"
	SuggestedStrategyEquityPineLine        = `plot(strategy.equity, "Strategy Equity", display=display.data_window)`
)

type strategyReportMode string

const (
	strategyReportModeSummary strategyReportMode = "summary"
	strategyReportModeTrades  strategyReportMode = "trades"
	strategyReportModeOrders  strategyReportMode = "orders"
	strategyReportModeEquity  strategyReportMode = "equity"
)

func clampStrategyLimit(value, fallback, max int) int {
	if value <= 0 {
		value = fallback
	}
	if value > max {
		value = max
	}
	return value
}

func evaluateStrategyReport(mode strategyReportMode, limit int) (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	raw, err := withSession(ctx, func(c *cdp.Client) (json.RawMessage, error) {
		return c.EvaluateWithOptions(ctx, buildStrategyReportJS(mode, limit), cdp.EvaluateOptions{
			AwaitPromise:  true,
			ReturnByValue: true,
			Timeout:       18 * time.Second,
		})
	})
	if err != nil {
		return nil, err
	}
	var result map[string]interface{}
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("parse strategy report: %w", err)
	}
	if _, ok := result["success"]; !ok {
		result["success"] = false
		result["status"] = "strategy_report_shape_unverified"
		result["source"] = SourceTradingViewBacktestingAPI
		result["error"] = "TradingView backtesting API returned an unrecognized payload."
	}
	return result, nil
}

func buildStrategyReportJS(mode strategyReportMode, limit int) string {
	modeJSON, _ := json.Marshal(string(mode))
	return `(async function() {
		var MODE = ` + string(modeJSON) + `;
		var LIMIT = ` + strconv.Itoa(limit) + `;
		var SOURCE = "` + SourceTradingViewBacktestingAPI + `";
		var STRATEGY_PLOT_SOURCE = "` + SourceTradingViewStrategyPlot + `";
		var STRATEGY_PLOT_RELIABILITY = "` + ReliabilityPineRuntimeUnstableInternal + `";
		var SUGGESTED_EQUITY_PINE_LINE = ` + strconv.Quote(SuggestedStrategyEquityPineLine) + `;
		var DERIVED_OHLCV_TRADES_SOURCE = "` + SourceDerivedFromOHLCVAndTrades + `";
		var DERIVED_SOURCE = "` + SourceDerivedFromBacktestingTrades + `";
		var RELIABILITY = "` + ReliabilityBacktestingUnstableInternal + `";
		var DERIVED_RELIABILITY = "` + ReliabilityDerivedTradeExitPoints + `";

		function assign(target, extra) {
			if (!extra || typeof extra !== "object") return target;
			Object.keys(extra).forEach(function(key) { target[key] = extra[key]; });
			return target;
		}
		function fail(status, error, extra) {
			return assign({
				success: false,
				status: status,
				source: SOURCE,
				reliability: RELIABILITY,
				reliableForTradingLogic: false,
				error: error
			}, extra);
		}
		function ok(extra) {
			return assign({
				success: true,
				status: "ok",
				source: SOURCE,
				reliability: RELIABILITY,
				reliableForTradingLogic: true
			}, extra);
		}
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
		function finiteNumber(value) {
			value = unwrap(value);
			if (typeof value === "number" && isFinite(value)) return value;
			if (value && typeof value === "object" && typeof value.value === "number" && isFinite(value.value)) return value.value;
			if (typeof value === "string" && value.trim() !== "") {
				var parsed = Number(value);
				if (isFinite(parsed)) return parsed;
			}
			return null;
		}
		function clone(value, depth, maxArray) {
			value = unwrap(value);
			if (value == null) return value;
			if (typeof value === "number" || typeof value === "string" || typeof value === "boolean") return value;
			if (typeof value === "bigint") return value.toString();
			if (typeof value === "function" || typeof value === "symbol") return undefined;
			if (depth <= 0) return "[MaxDepth]";
			if (Array.isArray(value)) {
				var out = [];
				var len = Math.min(value.length, maxArray || 100);
				for (var i = 0; i < len; i++) {
					var item = clone(value[i], depth - 1, maxArray);
					if (item !== undefined) out.push(item);
				}
				return out;
			}
			var result = {};
			Object.keys(value).forEach(function(key) {
				if (key.indexOf("_") === 0 && key !== "_value") return;
				var item = clone(value[key], depth - 1, maxArray);
				if (item !== undefined) result[key] = item;
			});
			return result;
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
		function chartModel() {
			var api = window.TradingViewApi || {};
			var activeWidget = unwrap(api._activeChartWidgetWV);
			var chartWidget = activeWidget && (activeWidget._chartWidget || activeWidget);
			var chartModel = chartWidget && typeof chartWidget.model === "function" ? chartWidget.model() : null;
			return chartModel && typeof chartModel.model === "function" ? chartModel.model() : chartModel;
		}
		function strategyState(model) {
			var sources = [];
			var active = null;
			if (model && typeof model.strategySources === "function") {
				sources = arrayify(model.strategySources());
			}
			if (model && typeof model.activeStrategySource === "function") {
				active = unwrap(model.activeStrategySource());
			}
			if (!active && sources.length > 0) active = sources[0];
			return {
				active_source: active,
				strategy_sources_raw: sources,
				info: {
					strategy_loaded: !!active || sources.length > 0,
					strategy_source_count: sources.length,
					active_strategy_id: active ? sourceId(active) : "",
					active_strategy_name: active ? sourceName(active) : "",
					strategy_sources: sources.map(function(source) {
						return { id: sourceId(source), name: sourceName(source) };
					})
				}
			};
		}
		function reportFromBacktestingAPI(bt) {
			if (!bt) return null;
			return unwrap(bt.activeStrategyReportData || bt._activeStrategyReportData || bt._reportData);
		}
		function metricMap(performance) {
			performance = unwrap(performance);
			if (!performance || typeof performance !== "object") return {};
			var all = unwrap(performance.all);
			if (all && typeof all === "object" && !Array.isArray(all)) return clone(all, 4, 100) || {};
			return clone(performance, 4, 100) || {};
		}
		function reportShape(report) {
			var performance = unwrap(report && report.performance);
			var trades = arrayify(report && report.trades);
			var orders = arrayify(report && report.filledOrders);
			return {
				report_keys: report && typeof report === "object" ? Object.keys(report) : [],
				has_performance: !!performance,
				has_trades: trades.length > 0,
				has_filled_orders: orders.length > 0,
				total_trade_count: trades.length,
				total_order_count: orders.length
			};
		}
		function firstArray() {
			for (var i = 0; i < arguments.length; i++) {
				var arr = arrayify(arguments[i]);
				if (arr.length > 0) return arr;
			}
			return [];
		}
		function tradeExitTime(trade) {
			var exit = unwrap(trade && trade.exit);
			return finiteNumber(exit && (exit.time || exit.timestamp || exit.barTime)) ||
				finiteNumber(trade && (trade.exitTime || trade.time || trade.barTime));
		}
		function tradeCumulativeProfit(trade) {
			return finiteNumber(trade && (trade.cumulativeProfit || trade.cumProfit || trade.netProfit || trade.profitCumulative));
		}
		function tradeDrawdown(trade) {
			return finiteNumber(trade && (trade.drawdown || trade.maxDrawdown || trade.runupDrawdown));
		}
		function initialCapital(performance) {
			performance = unwrap(performance);
			return finiteNumber(performance && performance.initialCapital) ||
				finiteNumber(performance && performance.all && performance.all.initialCapital) ||
				0;
		}
		function lower(value) {
			return String(value == null ? "" : value).toLowerCase();
		}
		function rowValue(row) {
			if (!row) return null;
			if (Array.isArray(row)) return { index: null, value: row };
			if (Array.isArray(row.value)) return { index: row.index == null ? null : row.index, value: row.value };
			return null;
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
				infos.push({
					plot_id: id,
					name: String(title || id),
					type: String(plot.type || ""),
					value_index: i + 1,
					is_hidden: !!style.isHidden
				});
			}
			return infos;
		}
		function isStrategyEquityPlot(info) {
			var title = lower(info && info.name).replace(/[_-]+/g, " ").trim();
			var plotId = lower(info && info.plot_id).replace(/[_-]+/g, " ").trim();
			return title === "strategy equity" || title.indexOf("strategy equity") !== -1 ||
				plotId === "strategy equity" || plotId.indexOf("strategy equity") !== -1;
		}
		function timeToMs(value) {
			var time = finiteNumber(value);
			if (time == null) return null;
			return time < 1000000000000 ? Math.round(time * 1000) : Math.round(time);
		}
		function strategyEquityFromPlot(source, limit) {
			var meta = null;
			var data = null;
			try { meta = source && typeof source.metaInfo === "function" ? source.metaInfo() : unwrap(source && source.metaInfo); } catch(e) {}
			try { data = source && typeof source.data === "function" ? source.data() : unwrap(source && source.data); } catch(e) {}
			var infos = plotInfos(meta);
			var equityInfo = null;
			for (var i = 0; i < infos.length; i++) {
				if (isStrategyEquityPlot(infos[i])) {
					equityInfo = infos[i];
					break;
				}
			}
			if (!equityInfo) {
				return {
					found: false,
					available_plots: infos.map(function(info) {
						return { plot_id: info.plot_id, name: info.name, value_index: info.value_index, type: info.type, is_hidden: info.is_hidden };
					})
				};
			}
			var points = [];
			var totalRows = 0;
			try {
				var it = data && data.fullRangeIterator && data.fullRangeIterator();
				for (var guard = 0; it && guard < 200000; guard++) {
					var next = it.next();
					if (!next || next.done) break;
					totalRows++;
					var row = rowValue(next.value);
					if (!row || !row.value) continue;
					var equity = finiteNumber(row.value[equityInfo.value_index]);
					var time = timeToMs(row.value[0]);
					if (equity == null || time == null) continue;
					points.push({
						index: row.index,
						time: time,
						equity: equity
					});
				}
			} catch(e) {
				return {
					found: true,
					error: e && e.message ? e.message : String(e),
					plot: equityInfo,
					points: [],
					loaded_bar_count: totalRows
				};
			}
			var allPointCount = points.length;
			if (limit > 0 && points.length > limit) points = points.slice(points.length - limit);
			return {
				found: true,
				plot: equityInfo,
				points: points,
				data_points: points.length,
				total_data_points: allPointCount,
				loaded_bar_count: totalRows
			};
		}
		function derivedFallbackDescriptor(report, limit) {
			var fallback = {
				available: false,
				source: DERIVED_OHLCV_TRADES_SOURCE,
				report_source: SOURCE,
				reliability: "derived_conditional_on_complete_ohlcv_trades_settings",
				reliableForTradingLogic: false,
				coverage: "not_computed",
				requirements: [
					"complete OHLCV coverage for the requested range",
					"report.trades",
					"initial capital",
					"symbol point value",
					"commission and slippage settings",
					"TradingView order fill timing"
				],
				limitations: [
					"Derived reconstruction is not native TradingView Pine runtime output.",
					"Loaded chart bars can be shorter than the full backtest history.",
					"Trade-exit cumulative profit alone is not a bar-by-bar equity curve."
				]
			};
			if (!report) return fallback;
			var performance = unwrap(report.performance);
			var trades = arrayify(report.trades);
			var base = initialCapital(performance);
			var points = [];
			for (var i = 0; i < trades.length && points.length < limit; i++) {
				var cumulative = tradeCumulativeProfit(trades[i]);
				if (cumulative == null) continue;
				var point = {
					source_trade_number: i + 1,
					cumulative_pnl: cumulative,
					equity: base + cumulative
				};
				var time = tradeExitTime(trades[i]);
				if (time != null) point.time = timeToMs(time);
				var drawdown = tradeDrawdown(trades[i]);
				if (drawdown != null) point.drawdown = drawdown;
				points.push(point);
			}
			if (points.length > 0) {
				fallback.available = true;
				fallback.coverage = "trade_exit_points_only";
				fallback.data = points;
				fallback.data_points = points.length;
				fallback.total_trade_count = trades.length;
				fallback.warning = "This fallback is not full bar-by-bar equity; it is a derived trade-exit series from the backtesting report.";
			}
			return fallback;
		}
		function needsEquityPlotResult(info, plotResult, report, extra) {
			return withStrategyInfo(assign({
				success: false,
				status: "needs_equity_plot",
				source: STRATEGY_PLOT_SOURCE,
				reliability: STRATEGY_PLOT_RELIABILITY,
				reliableForTradingLogic: false,
				coverage: "unavailable",
				error: "Strategy Equity plot was not found in the active strategy source.",
				suggested_pine_line: SUGGESTED_EQUITY_PINE_LINE,
				available_plots: plotResult && plotResult.available_plots ? plotResult.available_plots : [],
				derived_fallback: derivedFallbackDescriptor(report, limitForFallback(LIMIT))
			}, extra), info);
		}
		function limitForFallback(value) {
			return value > 0 ? value : 500;
		}
		function withStrategyInfo(result, info) {
			result.strategy_loaded = info.strategy_loaded;
			result.strategy_source_count = info.strategy_source_count;
			result.active_strategy_id = info.active_strategy_id;
			result.active_strategy_name = info.active_strategy_name;
			result.strategy_sources = info.strategy_sources;
			return result;
		}

		try {
			var model = chartModel();
			var state = strategyState(model);
			var info = state.info;
			if (!info.strategy_loaded) {
				return fail("no_strategy_loaded", "No TradingView strategy is loaded on the active chart.", info);
			}

			if (MODE === "equity") {
				var plotResult = strategyEquityFromPlot(state.active_source, LIMIT);
				if (plotResult.found && plotResult.points && plotResult.points.length > 0) {
					return withStrategyInfo({
						success: true,
						status: "ok",
						source: STRATEGY_PLOT_SOURCE,
						reliability: STRATEGY_PLOT_RELIABILITY,
						reliableForTradingLogic: true,
						coverage: "loaded_chart_bars",
						data: plotResult.points,
						data_points: plotResult.data_points,
						total_data_points: plotResult.total_data_points,
						loaded_bar_count: plotResult.loaded_bar_count,
						plot: plotResult.plot,
						limit: LIMIT,
						warning: "Strategy Equity plot data covers loaded chart bars only, not necessarily the full backtest history."
					}, info);
				}
				if (plotResult.found && plotResult.error) {
					return withStrategyInfo({
						success: false,
						status: "strategy_report_unavailable",
						source: STRATEGY_PLOT_SOURCE,
						reliability: STRATEGY_PLOT_RELIABILITY,
						reliableForTradingLogic: false,
						coverage: "unavailable",
						error: plotResult.error,
						plot: plotResult.plot,
						suggested_pine_line: SUGGESTED_EQUITY_PINE_LINE
					}, info);
				}
				if (plotResult.found) {
					return withStrategyInfo({
						success: false,
						status: "strategy_report_unavailable",
						source: STRATEGY_PLOT_SOURCE,
						reliability: STRATEGY_PLOT_RELIABILITY,
						reliableForTradingLogic: false,
						coverage: "loaded_chart_bars",
						error: "Strategy Equity plot was found, but no numeric equity rows were available in the loaded chart data.",
						plot: plotResult.plot,
						data_points: 0,
						total_data_points: plotResult.total_data_points || 0,
						loaded_bar_count: plotResult.loaded_bar_count || 0
					}, info);
				}
				var fallbackReport = null;
				var fallbackExtra = {};
				try {
					if (window.TradingViewApi && typeof window.TradingViewApi.backtestingStrategyApi === "function") {
						var fallbackBT = await window.TradingViewApi.backtestingStrategyApi();
						fallbackReport = reportFromBacktestingAPI(fallbackBT);
						if (!fallbackReport) fallbackExtra.fallback_report_status = "strategy_report_unavailable";
					} else {
						fallbackExtra.fallback_report_status = "tradingview_backtesting_api_unavailable";
					}
				} catch(e) {
					fallbackExtra.fallback_report_status = "strategy_report_unavailable";
					fallbackExtra.fallback_report_error = e && e.message ? e.message : String(e);
				}
				return needsEquityPlotResult(info, plotResult, fallbackReport, fallbackExtra);
			}

			if (!window.TradingViewApi || typeof window.TradingViewApi.backtestingStrategyApi !== "function") {
				return fail("tradingview_backtesting_api_unavailable", "window.TradingViewApi.backtestingStrategyApi() is unavailable.", info);
			}

			var bt = await window.TradingViewApi.backtestingStrategyApi();
			if (!bt) {
				return fail("tradingview_backtesting_api_unavailable", "window.TradingViewApi.backtestingStrategyApi() returned null.", info);
			}
			var report = reportFromBacktestingAPI(bt);
			if (!report) {
				return fail("strategy_report_unavailable", "Strategy is loaded, but activeStrategyReportData is unavailable.", info);
			}

			var shape = reportShape(report);
			if (!shape.has_performance && !shape.has_trades && !shape.has_filled_orders) {
				return fail("strategy_report_shape_unverified", "Strategy report does not expose performance, trades, or filledOrders.", assign(info, shape));
			}

			var performance = unwrap(report.performance);
			var settings = unwrap(report.settings);
			var currency = clone(report.currency, 2, 20);
			var trades = arrayify(report.trades);
			var orders = arrayify(report.filledOrders);

			if (MODE === "summary") {
				var metrics = metricMap(performance);
				return withStrategyInfo(ok(assign(shape, {
					currency: currency,
					settings: clone(settings, 5, 100),
					performance: clone(performance, 6, 200),
					metrics: metrics,
					metric_count: Object.keys(metrics).length
				})), info);
			}

			if (MODE === "trades") {
				var tradeRows = trades.slice(0, LIMIT).map(function(trade) { return clone(trade, 6, 100); });
				return withStrategyInfo(ok({
					currency: currency,
					trades: tradeRows,
					trade_count: tradeRows.length,
					total_trade_count: trades.length,
					limit: LIMIT
				}), info);
			}

			if (MODE === "orders") {
				var orderRows = orders.slice(0, LIMIT).map(function(order) { return clone(order, 5, 100); });
				return withStrategyInfo(ok({
					currency: currency,
					orders: orderRows,
					order_count: orderRows.length,
					total_order_count: orders.length,
					limit: LIMIT
				}), info);
			}

			return fail("strategy_report_shape_unverified", "Unknown strategy report mode: " + MODE, info);
		} catch (e) {
			return fail("strategy_report_unavailable", e && e.message ? e.message : String(e), {});
		}
	})()`
}
