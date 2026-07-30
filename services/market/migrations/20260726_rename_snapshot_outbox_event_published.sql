ALTER TABLE `t_itick_snapshot_outbox`
  CHANGE COLUMN `option_published_at` `event_published_at`
  BIGINT NOT NULL DEFAULT 0 COMMENT '权威行情Kafka事件发布完成时间；无需发布时同样置完成';
