ALTER TABLE `t_user`
  MODIFY COLUMN `register_type` TINYINT NOT NULL DEFAULT 1
  COMMENT '注册方式：1用户名 2手机号 3邮箱 4游客 5做市账户';

UPDATE `t_user`
SET
  `register_type` = 5,
  `update_times` = GREATEST(`update_times`, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000)
WHERE `account_type` = 2
  AND `register_type` <> 5;
