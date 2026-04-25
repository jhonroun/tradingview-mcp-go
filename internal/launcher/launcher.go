// Package launcher starts TradingView Desktop with CDP remote debugging enabled.
package launcher

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/jhonroun/tradingview-mcp-go/internal/cdp"
	"github.com/jhonroun/tradingview-mcp-go/internal/discovery"
)

// Launch finds and starts TradingView with --remote-debugging-port=<port>.
// If tvPath is non-empty it is used directly, skipping auto-discovery.
// If killExisting is true, any running TradingView processes are killed first.
func Launch(port int, killExisting bool, tvPath string) (map[string]interface{}, error) {
	if killExisting {
		killRunning()
	}

	execPath := tvPath
	var source string
	if execPath == "" {
		found, err := discovery.Find()
		if err != nil {
			return map[string]interface{}{"success": false, "error": err.Error()}, nil
		}
		execPath = found.Path
		source = found.Source
	} else {
		source = "cli-flag"
		if _, err := os.Stat(execPath); err != nil {
			return map[string]interface{}{
				"success": false,
				"error":   fmt.Sprintf("--tv-path not found: %v", err),
			}, nil
		}
	}

	cmd := exec.Command(execPath, fmt.Sprintf("--remote-debugging-port=%d", port))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   fmt.Sprintf("failed to start TradingView: %v", err),
		}, nil
	}

	// Wait up to 15 s for CDP to become available.
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	base := map[string]interface{}{
		"success":  true,
		"pid":      cmd.Process.Pid,
		"port":     port,
		"platform": runtime.GOOS,
		"path":     execPath,
		"source":   source,
	}

	for {
		targets, err := cdp.ListTargets(ctx, "localhost", port)
		if err == nil {
			if _, err := cdp.FindChartTarget(targets); err == nil {
				return base, nil
			}
		}
		select {
		case <-ctx.Done():
			base["hint"] = "TradingView started but chart not yet detected; open a chart"
			return base, nil
		case <-time.After(500 * time.Millisecond):
		}
	}
}

func killRunning() {
	switch runtime.GOOS {
	case "windows":
		_ = exec.Command("taskkill", "/F", "/IM", "TradingView.exe").Run()
	case "darwin":
		_ = exec.Command("pkill", "-f", "TradingView").Run()
	default:
		_ = exec.Command("pkill", "-f", "tradingview").Run()
	}
}
