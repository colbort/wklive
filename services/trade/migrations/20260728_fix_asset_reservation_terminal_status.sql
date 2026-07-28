-- Repair reservations written by the former single-statement settlement
-- update. MySQL evaluates UPDATE assignments left-to-right, so that statement
-- added the current amount twice while deciding the terminal status.
UPDATE t_trade_asset_reservation
SET status = CASE
        WHEN consumed_amount = reserved_amount AND released_amount = 0 THEN 4
        ELSE 6
    END,
    next_retry_at = 0,
    last_error_msg = '',
    version = version + 1,
    update_times = CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000 AS UNSIGNED)
WHERE
    (
        consumed_amount = reserved_amount
        AND released_amount = 0
        AND status <> 4
    )
    OR (
        consumed_amount + released_amount = reserved_amount
        AND released_amount > 0
        AND status <> 6
    );
