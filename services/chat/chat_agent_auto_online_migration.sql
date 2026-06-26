ALTER TABLE `t_chat_agent`
  ADD COLUMN `auto_online` tinyint NOT NULL DEFAULT '2' COMMENT '登录是否自动上线:1是 2否'
  AFTER `status`;


ALTER TABLE `t_chat_category`
  ADD COLUMN `group_id` bigint NOT NULL DEFAULT '0' COMMENT '客服分组ID'
  AFTER `category_name`;


ALTER TABLE `t_chat_session`
  ADD COLUMN `agent_user_id` bigint NOT NULL DEFAULT '0' COMMENT '当前坐席的用户ID'
  AFTER `agent_id`;