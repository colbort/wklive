ALTER TABLE `t_user`
  ADD COLUMN `source_origin` varchar(255) NOT NULL DEFAULT ''
    COMMENT '游客首次登录的规范化域名Origin'
    AFTER `source`,
  ADD KEY `idx_tenant_guest_origin_activity`
    (`tenant_id`, `is_guest`, `source_origin`, `deleted`, `last_login_time`);
