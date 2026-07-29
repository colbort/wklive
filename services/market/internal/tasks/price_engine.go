package tasks

import (
	"context"
	"errors"
	"fmt"
	"time"

	"wklive/common/alert"
	"wklive/common/notify"
	"wklive/services/market/internal/priceengine"

	"github.com/zeromicro/go-zero/core/logx"
)

const (
	operationalAlertReminderInterval = 30 * time.Minute
	operationalAlertRetryInterval    = 30 * time.Second
)

func StartPriceEngine(
	ctx context.Context,
	engine *priceengine.Engine,
	alertNotifier alert.Notifier,
) {
	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		var lastUnavailableLog time.Time
		inputUnavailable := false
		var alertTracker alert.DeliveryTracker
		for {
			err := engine.RunOnce(ctx, time.Now().UnixMilli())
			if err != nil && ctx.Err() == nil {
				if errors.Is(err, priceengine.ErrInputUnavailable) {
					now := time.Now()
					if lastUnavailableLog.IsZero() || now.Sub(lastUnavailableLog) >= 30*time.Second {
						// Missing authoritative inputs must remain visible with
						// production logging, but the engine runs every second.
						// Keep the error actionable without restoring the
						// previous one-error-per-tick log storm.
						logx.Errorf("price engine waiting for formula input: %v", err)
						lastUnavailableLog = now
					}
					detail := err.Error()
					fingerprint := priceEngineInputAlertFingerprint(err)
					if alertTracker.ShouldPublishFiring(
						fingerprint,
						now,
						operationalAlertReminderInterval,
						operationalAlertRetryInterval,
					) {
						publishErr := publishPriceEngineInputAlert(
							ctx,
							alertNotifier,
							alert.StateFiring,
							detail,
							now.UnixMilli(),
						)
						alertTracker.MarkFiringAttempt(fingerprint, now, publishErr == nil)
						if publishErr != nil {
							logx.Errorf("publish price engine input alert failed: %v", publishErr)
						}
					}
					inputUnavailable = true
				} else {
					logx.Errorf("price engine evaluation failed: %v", err)
				}
			} else if err == nil {
				now := time.Now()
				if inputUnavailable {
					logx.Infof("price engine formula input recovered")
					inputUnavailable = false
					lastUnavailableLog = time.Time{}
				}
				if alertTracker.ShouldPublishResolved(now, operationalAlertRetryInterval) {
					publishErr := publishPriceEngineInputAlert(
						ctx,
						alertNotifier,
						alert.StateResolved,
						"all active formulas have authoritative inputs",
						now.UnixMilli(),
					)
					alertTracker.MarkResolvedAttempt(now, publishErr == nil)
					if publishErr != nil {
						logx.Errorf("publish price engine recovery alert failed: %v", publishErr)
					}
				}
			}
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
			}
		}
	}()
}

func priceEngineInputAlertFingerprint(err error) string {
	var unavailable *priceengine.InputUnavailableError
	if errors.As(err, &unavailable) {
		return unavailable.AlertFingerprint()
	}
	return err.Error()
}

func publishPriceEngineInputAlert(
	ctx context.Context,
	notifier alert.Notifier,
	state string,
	detail string,
	now int64,
) error {
	title := "Price Engine 输入缺失"
	if state == alert.StateResolved {
		title = "Price Engine 输入恢复"
	}
	at := alert.New(
		alert.TypePriceEngineInput,
		state,
		notify.EventLevelError,
		"market",
		"price-engine-input",
		title,
		fmt.Sprintf("state=%s detail=%s", state, detail),
		now,
	)
	at.Data = map[string]any{"detail": detail}
	return alert.Notify(ctx, notifier, at)
}
