ALTER TABLE `t_itick_kline_sync_progress`
  ADD COLUMN `contiguous_ts` bigint NOT NULL DEFAULT '0' COMMENT '最后连续完整已确认K线时间戳（毫秒）' AFTER `latest_ts`,
  ADD COLUMN `recent_check_ts` bigint NOT NULL DEFAULT '0' COMMENT '最近一次REST校准时间（毫秒）' AFTER `contiguous_ts`;

UPDATE `t_itick_kline_sync_progress`
SET `contiguous_ts` = `latest_ts`
WHERE `contiguous_ts` = 0
  AND `latest_ts` > 0;


ALTER TABLE `t_itick_category`
  ADD COLUMN `sync_priority` tinyint NOT NULL DEFAULT '2' COMMENT 'K线同步优先级: 1-高 2-普通 3-低' AFTER `app_visible`;


ALTER TABLE `t_itick_product`
  ADD COLUMN `sync_priority` tinyint NOT NULL DEFAULT '2' COMMENT 'K线同步优先级: 1-高 2-普通 3-低' AFTER `app_visible`;
