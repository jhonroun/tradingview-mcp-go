package health

import (
	"strings"
	"testing"
)

func TestBuildCompatibilityProbeJSContainsRequiredPaths(t *testing.T) {
	js := buildCompatibilityProbeJS()
	for _, want := range []string{
		"window.TradingViewApi",
		"_activeChartWidgetWV",
		"chart.model().model()",
		"model.dataSources",
		"valueAt",
		"fullRangeIterator",
		"model.strategySources",
		"model.activeStrategySource",
		"await api.backtestingStrategyApi()",
		"strategy_equity_plot",
		"loaded_chart_bars",
		"unstable_internal_path",
	} {
		if !strings.Contains(js, want) {
			t.Errorf("compatibility probe JS missing %q", want)
		}
	}
}

func TestCompatibilityProbeShapeSupportsUnavailableStates(t *testing.T) {
	js := buildCompatibilityProbeJS()
	for _, want := range []string{
		"no_strategy_loaded",
		"strategy_report_unavailable",
		"needs_equity_plot",
		"compatible",
		"available",
		"reliability",
	} {
		if !strings.Contains(js, want) {
			t.Errorf("compatibility probe JS missing status/field %q", want)
		}
	}
}
