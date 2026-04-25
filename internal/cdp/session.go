package cdp

import "context"

// WithSession connects to the TradingView chart target, enables required domains,
// calls fn, and closes the connection. This eliminates the connect/enable/close
// boilerplate in every tool implementation.
func WithSession(ctx context.Context, fn func(*Client, *Target) error) error {
	targets, err := ListTargets(ctx, defaultHost, defaultPort)
	if err != nil {
		return err
	}
	target, err := FindChartTarget(targets)
	if err != nil {
		return err
	}
	client, err := Connect(ctx, target)
	if err != nil {
		return err
	}
	defer client.Close()
	if err := client.EnableDomains(ctx); err != nil {
		return err
	}
	return fn(client, target)
}
