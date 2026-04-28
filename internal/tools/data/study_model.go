package data

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jhonroun/tradingview-mcp-go/internal/cdp"
	tv "github.com/jhonroun/tradingview-mcp-go/internal/tradingview"
)

const defaultIndicatorHistoryLimit = 500

type studyModelQuery struct {
	EntityID string `json:"entity_id,omitempty"`
	Name     string `json:"name,omitempty"`
	Query    string `json:"query,omitempty"`
}

// StudyHistoryValue is one plot value inside a historical study row.
type StudyHistoryValue struct {
	PlotID string  `json:"plot_id"`
	Name   string  `json:"name"`
	Value  float64 `json:"value"`
}

// StudyHistoryPoint is one bar of historical study output.
type StudyHistoryPoint struct {
	Index      int                 `json:"index"`
	Time       int64               `json:"time"`
	Values     map[string]float64  `json:"values"`
	PlotValues []StudyHistoryValue `json:"plot_values"`
}

func GetIndicatorByQuery(entityID, name string) (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	query := studyModelQuery{EntityID: entityID, Name: name}
	raw, err := withSession(ctx, func(c *cdp.Client) (json.RawMessage, error) {
		return c.Evaluate(ctx, buildStudyModelJS(query, 0, false, false), false)
	})
	if err != nil {
		return nil, err
	}
	var result map[string]interface{}
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("parse indicator: %w", err)
	}
	if success, _ := result["success"].(bool); !success {
		if errMsg, _ := result["error"].(string); errMsg != "" {
			return nil, fmt.Errorf("%s", errMsg)
		}
		return nil, fmt.Errorf("study not found")
	}
	return result, nil
}

func GetIndicatorHistory(entityID, name string, maxBars int) (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	limit := maxBars
	if limit <= 0 {
		limit = defaultIndicatorHistoryLimit
	}
	if limit > maxOHLCVBars {
		limit = maxOHLCVBars
	}

	query := studyModelQuery{EntityID: entityID, Name: name}
	raw, err := withSession(ctx, func(c *cdp.Client) (json.RawMessage, error) {
		return c.Evaluate(ctx, buildStudyModelJS(query, limit, true, false), false)
	})
	if err != nil {
		return nil, err
	}
	var result map[string]interface{}
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("parse indicator history: %w", err)
	}
	if success, _ := result["success"].(bool); !success {
		if errMsg, _ := result["error"].(string); errMsg != "" {
			return nil, fmt.Errorf("%s", errMsg)
		}
		return nil, fmt.Errorf("study not found")
	}
	return result, nil
}

func buildStudyModelJS(query studyModelQuery, historyLimit int, includeHistory, allStudies bool) string {
	q, _ := json.Marshal(query)
	return fmt.Sprintf(`(function() {
		var SOURCE = %s;
		var RELIABILITY = %s;
		var query = %s;
		var includeHistory = %v;
		var allStudies = %v;
		var historyLimit = %d;

		function sourceId(s) {
			try { return String(typeof s.id === 'function' ? s.id() : (s.id || '')); } catch(e) { return ''; }
		}
		function lower(v) { return String(v || '').toLowerCase(); }
		function metaName(meta) { return meta ? (meta.description || meta.shortDescription || meta.name || '') : ''; }
		function finiteNumber(v) { return typeof v === 'number' && isFinite(v); }
		function rowValue(row) {
			if (!row) return null;
			if (Array.isArray(row)) return { index: null, value: row };
			if (Array.isArray(row.value)) return { index: row.index == null ? null : row.index, value: row.value };
			return null;
		}
		function usefulType(t) {
			t = lower(t);
			return t !== 'colorer' && t !== 'bg_colorer' && t !== 'bar_colorer' &&
				t !== 'wick_colorer' && t !== 'border_colorer' && t !== 'alertcondition';
		}
		function plotInfos(meta) {
			var plots = (meta && Array.isArray(meta.plots)) ? meta.plots : [];
			var styles = (meta && meta.styles) ? meta.styles : {};
			var infos = [];
			for (var i = 0; i < plots.length; i++) {
				var p = plots[i] || {};
				var id = String(p.id || ('plot_' + i));
				var st = styles[id] || {};
				var title = st.title || p.title || p.name || id;
				infos.push({
					plot_id: id,
					name: title,
					type: String(p.type || ''),
					value_index: i + 1,
					is_hidden: !!st.isHidden
				});
			}
			return infos;
		}
		function mappedPlotValues(row, infos) {
			var rv = rowValue(row);
			if (!rv || !rv.value || rv.value.length === 0) return { time: null, index: null, plots: [], values: {} };
			var arr = rv.value;
			var values = {};
			var plots = [];
			for (var i = 0; i < infos.length; i++) {
				var info = infos[i];
				if (!usefulType(info.type) || info.is_hidden) continue;
				var value = arr[info.value_index];
				if (!finiteNumber(value)) continue;
				values[info.name] = value;
				plots.push({
					name: info.name,
					plot_id: info.plot_id,
					type: info.type,
					current: value,
					values: [value],
					value_index: info.value_index,
					is_hidden: info.is_hidden,
					source: SOURCE,
					reliability: RELIABILITY,
					reliableForTradingLogic: true
				});
			}
			return { time: arr[0] || null, index: rv.index, plots: plots, values: values };
		}
		function currentRow(data) {
			if (!data) return null;
			try {
				if (typeof data.lastIndex === 'function' && typeof data.valueAt === 'function') {
					var idx = data.lastIndex();
					var row = data.valueAt(idx);
					if (Array.isArray(row)) return { index: idx, value: row };
					return row;
				}
			} catch(e) {}
			try {
				var it = data.fullRangeIterator && data.fullRangeIterator();
				var last = null;
				for (var guard = 0; it && guard < 200000; guard++) {
					var n = it.next();
					if (!n || n.done) break;
					last = n.value;
				}
				return last;
			} catch(e) {}
			return null;
		}
		function historyRows(data, infos, limit) {
			var rows = [];
			try {
				var it = data && data.fullRangeIterator && data.fullRangeIterator();
				for (var guard = 0; it && guard < 200000; guard++) {
					var n = it.next();
					if (!n || n.done) break;
					var mapped = mappedPlotValues(n.value, infos);
					if (!mapped.time || mapped.plots.length === 0) continue;
					var values = {};
					var plotValues = [];
					for (var i = 0; i < mapped.plots.length; i++) {
						var p = mapped.plots[i];
						values[p.name] = p.current;
						plotValues.push({ plot_id: p.plot_id, name: p.name, value: p.current });
					}
					rows.push({ index: n.value.index, time: mapped.time, values: values, plot_values: plotValues });
				}
			} catch(e) {}
			if (limit > 0 && rows.length > limit) rows = rows.slice(rows.length - limit);
			return rows;
		}
		function userStudyIds(api) {
			var ids = {};
			try {
				var studies = api.getAllStudies ? api.getAllStudies() : [];
				if (Array.isArray(studies)) {
					for (var i = 0; i < studies.length; i++) ids[String(studies[i].id || '')] = true;
				}
			} catch(e) {}
			return ids;
		}
		function matches(s, meta, infos, userIds) {
			var id = sourceId(s);
			var name = metaName(meta);
			var hay = lower([id, name, meta && meta.shortDescription, meta && meta.description].concat(infos.map(function(p) { return p.name; })).join(' '));
			var eid = lower(query.entity_id);
			var qname = lower(query.name || query.query);
			if (eid) return lower(id) === eid;
			if (qname) return hay.indexOf(qname) !== -1;
			return !!userIds[id];
		}
		function sourceToStudy(s, includeHistory, historyLimit) {
			var id = sourceId(s);
			var meta = s.metaInfo ? s.metaInfo() : null;
			var name = metaName(meta);
			var data = s.data ? s.data() : null;
			var infos = plotInfos(meta);
			var mapped = mappedPlotValues(currentRow(data), infos);
			var study = {
				name: name,
				entity_id: id,
				plot_count: mapped.plots.length,
				plots: mapped.plots,
				current_bar_index: mapped.index,
				time: mapped.time,
				total_bars: (data && typeof data.size === 'function') ? data.size() : 0,
				coverage: 'loaded_chart_bars',
				source: SOURCE,
				reliability: RELIABILITY,
				reliableForTradingLogic: true
			};
			try {
				var publicStudy = api.getStudyById ? api.getStudyById(id) : null;
				if (publicStudy && typeof publicStudy.isVisible === 'function') study.visible = publicStudy.isVisible();
				if (publicStudy && typeof publicStudy.getInputValues === 'function') {
					var rawInputs = publicStudy.getInputValues();
					var inputs = {};
					if (rawInputs && rawInputs.length) {
						for (var ii = 0; ii < rawInputs.length; ii++) {
							var inp = rawInputs[ii];
							if (!inp || !inp.id || inp.value === undefined) continue;
							var val = inp.value;
							if (typeof val === 'string' && val.length > 500) continue;
							if (typeof val === 'string' && inp.id === 'text' && val.length > 200) continue;
							if (typeof val === 'string' && val.length > 200) val = val.substring(0, 200) + '...(truncated)';
							inputs[inp.id] = val;
						}
					}
					study.inputs = inputs;
				}
			} catch(e) {}
			if (includeHistory) {
				study.history = historyRows(data, infos, historyLimit);
				study.history_count = study.history.length;
			}
			return study;
		}

		try {
			var api = `+tv.ChartAPI+`;
			var chart = api._chartWidget;
			var sources = chart.model().model().dataSources();
			var userIds = userStudyIds(api);
			var studies = [];
			for (var si = 0; si < sources.length; si++) {
				var s = sources[si];
				if (!s || !s.metaInfo || !s.data) continue;
				var meta = s.metaInfo();
				var infos = plotInfos(meta);
				if (!matches(s, meta, infos, userIds)) continue;
				var id = sourceId(s);
				if (!allStudies && !query.entity_id && !query.name && !query.query && !userIds[id]) continue;
				var study = sourceToStudy(s, includeHistory, historyLimit);
				if (study.plot_count === 0 && !includeHistory) continue;
				studies.push(study);
				if (!allStudies) break;
			}
			if (allStudies) {
				return {
					success: true,
					study_count: studies.length,
					studies: studies,
					source: SOURCE,
					reliability: RELIABILITY,
					reliableForTradingLogic: true,
					coverage: 'loaded_chart_bars'
				};
			}
			if (studies.length === 0) {
				return {
					success: false,
					error: 'Study not found',
					source: SOURCE,
					reliability: RELIABILITY,
					reliableForTradingLogic: true
				};
			}
			var one = studies[0];
			one.success = true;
			return one;
		} catch(e) {
			return {
				success: false,
				error: e.message,
				source: SOURCE,
				reliability: RELIABILITY,
				reliableForTradingLogic: true
			};
		}
	})()`,
		tv.SafeString(SourceTradingViewStudyModel),
		tv.SafeString(ReliabilityPineRuntimeUnstableInternal),
		string(q),
		includeHistory,
		allStudies,
		historyLimit,
	)
}
