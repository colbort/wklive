ALTER TABLE `t_itick_kline_sync_progress`
  ADD COLUMN `contiguous_ts` bigint NOT NULL DEFAULT '0' COMMENT '最后连续完整已确认K线时间戳（毫秒）' AFTER `latest_ts`,
  ADD COLUMN `recent_check_ts` bigint NOT NULL DEFAULT '0' COMMENT '最近一次REST校准时间（毫秒）' AFTER `contiguous_ts`;

UPDATE `t_itick_kline_sync_progress`
SET `contiguous_ts` = `latest_ts`
WHERE `contiguous_ts` = 0
  AND `latest_ts` > 0;
