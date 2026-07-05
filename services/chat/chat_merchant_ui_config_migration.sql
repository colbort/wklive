ALTER TABLE `t_chat_merchant_info`
  ADD COLUMN `title` varchar(128) NOT NULL DEFAULT '' COMMENT 'chat-ui标题' AFTER `merchant_id`,
  ADD COLUMN `ui_config` json DEFAULT NULL COMMENT 'chat-ui展示配置' AFTER `api_secret`,
  ADD COLUMN `feature_config` json DEFAULT NULL COMMENT 'chat-ui功能开关' AFTER `ui_config`;
