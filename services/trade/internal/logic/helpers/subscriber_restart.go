package helpers

import (
	"context"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

const (
	subscriberRestartBaseDelay = time.Second
	subscriberRestartMaxDelay  = 30 * time.Second
)

// RunSubscriberWithRestart keeps a long-running subscriber alive across
// broker outages. Subscribe must block while connected and return when the
// connection stops. Context cancellation terminates the loop immediately.
func RunSubscriberWithRestart(ctx context.Context, name string, subscribe func() error) {
	runSubscriberWithRestart(ctx, name, subscriberRestartBaseDelay, subscriberRestartMaxDelay, subscribe)
}

func runSubscriberWithRestart(ctx context.Context, name string, baseDelay, maxDelay time.Duration, subscribe func() error) {
	if subscribe == nil || baseDelay <= 0 || maxDelay < baseDelay {
		return
	}
	backoff := baseDelay
	for ctx.Err() == nil {
		connectedAt := time.Now()
		err := subscribe()
		if ctx.Err() != nil {
			return
		}
		if time.Since(connectedAt) >= maxDelay {
			backoff = baseDelay
		}
		logx.Errorf("%s stopped, restarting in %s: %v", name, backoff, err)
		timer := time.NewTimer(backoff)
		select {
		case <-ctx.Done():
			if !timer.Stop() {
				<-timer.C
			}
			return
		case <-timer.C:
		}
		if backoff < maxDelay {
			backoff *= 2
			if backoff > maxDelay {
				backoff = maxDelay
			}
		}
	}
}
