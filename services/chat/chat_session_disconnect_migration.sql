ALTER TABLE `t_chat_session`
  ADD COLUMN `disconnect_time` bigint NOT NULL DEFAULT '0' COMMENT '网络异常断开时间戳(毫秒)' AFTER `close_reason`,
  ADD COLUMN `before_disconnect_status` tinyint NOT NULL DEFAULT '0' COMMENT '网络异常断开前状态' AFTER `disconnect_time`,
  ADD KEY `idx_status_disconnect_time` (`status`, `disconnect_time`);
