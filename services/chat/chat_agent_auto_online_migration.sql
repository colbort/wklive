ALTER TABLE `t_chat_agent`
  ADD COLUMN `auto_online` tinyint NOT NULL DEFAULT '2' COMMENT '登录是否自动上线:1是 2否'
  AFTER `status`;
